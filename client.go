package stun

import (
	"errors"
	"fmt"
	"log"
	"net"
	"runtime"
	"sync"
	"time"
)

const (
	defaultTimeoutRate = time.Millisecond * 100
	netUDP             = "udp"
	netUDP4            = "udp4"
	netUDP6            = "udp6"
	DefaultNet         = "udp"
	DefaultSTUNServer  = "gortc.io:3478"
)

var (
	// ErrNoConnection means that ClientOptions.Connection is nil.
	ErrNoConnection = errors.New("no connection provided")
	// ErrConnection means that the client already has a connection set
	ErrConnection = errors.New("connection already provided")
	// ErrNet means the network type is not supported
	ErrNet = errors.New("network type not supported")
	// ErrClientClosed indicates that client is closed.
	ErrClientClosed = errors.New("client is closed")
)

// Dial creates a stun connection to a STUN server
// using the supplied options.
func Dial(network, localaddress, stunserveraddress string, options ...func(*Client) error) (*Client, error) {
	if stunserveraddress == "" {
		stunserveraddress = DefaultSTUNServer
	}
	var laddr net.Addr
	var err error
	if localaddress != "" {
		laddr, err = ResolveAddr(network, localaddress)
		if err != nil {
			return nil, fmt.Errorf("localaddr: %v", err)
		}
	}
	raddr, err := ResolveAddr(network, stunserveraddress)
	if err != nil {
		return nil, fmt.Errorf("stunserveraddress: %v", err)
	}
	conn, err := listen(network, laddr)
	if err != nil {
		return nil, fmt.Errorf("listen: %v", err)
	}

	options = append(options, WithPacketConn(conn))
	options = append(options, WithSTUNServer(raddr))

	return NewClient(options...)
}

// ResolveAddr returns an address.
func ResolveAddr(network, address string) (net.Addr, error) {
	if network == "" {
		network = DefaultNet
	}

	switch network {
	case netUDP, netUDP4, netUDP6:
		return net.ResolveUDPAddr(network, address)
	default:
		return nil, ErrNet
	}
}

func listen(network string, laddr net.Addr) (PacketConn, error) {
	switch network {
	case netUDP, netUDP4, netUDP6:
		var addr *net.UDPAddr
		if laddr != nil {
			addr = laddr.(*net.UDPAddr)
		}
		return net.ListenUDP(network, addr)
	default:
		return nil, ErrNet
	}
}

// PacketConn represents a subset of net.PacketConn.
type PacketConn interface {
	ReadFrom(b []byte) (n int, addr net.Addr, err error)
	WriteTo(b []byte, addr net.Addr) (n int, err error)
	Close() error
	LocalAddr() net.Addr
}

// Client simulates "connection" to STUN server.
// The caller should either continuously call ReadFrom or
// use ReadUntilClosed to keep transaction processing active.
type Client struct {
	a          ClientAgent
	c          PacketConn
	serveraddr net.Addr
	close      chan struct{}
	gcRate     time.Duration
	closed     bool
	closedMux  sync.RWMutex
	wg         sync.WaitGroup
}

// Client itself implements the PacketConn interface
var _ PacketConn = (*Client)(nil)

// NewClient initializes new Client manually from provided options.
// Usage of Dial is preffered for most applications.
func NewClient(options ...func(*Client) error) (*Client, error) {
	c := &Client{
		close:  make(chan struct{}),
		gcRate: defaultTimeoutRate,
	}

	for _, option := range options {
		option(c)
	}

	if c.c == nil {
		return nil, ErrNoConnection
	}
	if c.a == nil {
		c.a = NewAgent(AgentOptions{})
	}

	runtime.SetFinalizer(c, clientFinalizer)
	return c, nil
}

// WithTimeoutRate allows the default timeout rate of 100ms to be overwritten.
func WithTimeoutRate(d time.Duration) func(*Client) error {
	return func(c *Client) error {
		c.gcRate = d
		return nil
	}
}

// WithAgent allows overwriting the default stun.ClientAgent.
func WithAgent(a ClientAgent) func(*Client) error {
	return func(c *Client) error {
		c.a = a
		return nil
	}
}

// WithPacketConn
func WithPacketConn(conn PacketConn) func(*Client) error {
	return func(c *Client) error {
		if c.c != nil {
			return ErrConnection
		}
		c.c = conn
		return nil
	}
}

// WithSTUNServer
func WithSTUNServer(addr net.Addr) func(*Client) error {
	return func(c *Client) error {
		c.serveraddr = addr
		return nil
	}
}

func clientFinalizer(c *Client) {
	if c == nil {
		return
	}
	err := c.Close()
	if err == ErrClientClosed {
		return
	}
	if err == nil {
		log.Println("client: called finalizer on non-closed client")
		return
	}
	log.Println("client: called finalizer on non-closed client:", err)
}

// ClientAgent is Agent implementation that is used by Client to
// process transactions.
type ClientAgent interface {
	Process(*Message) error
	Close() error
	Start(id [TransactionIDSize]byte, deadline time.Time, f Handler) error
	Stop(id [TransactionIDSize]byte) error
	Collect(time.Time) error
}

// StopErr occurs when Client fails to stop transaction while
// processing error.
type StopErr struct {
	Err   error // value returned by Stop()
	Cause error // error that caused Stop() call
}

func (e StopErr) Error() string {
	return fmt.Sprintf("error while stopping due to %s: %s",
		sprintErr(e.Cause), sprintErr(e.Err),
	)
}

// CloseErr indicates client close failure.
type CloseErr struct {
	AgentErr      error
	ConnectionErr error
}

func sprintErr(err error) string {
	if err == nil {
		return "<nil>"
	}
	return err.Error()
}

func (c CloseErr) Error() string {
	return fmt.Sprintf("failed to close: %s (connection), %s (agent)",
		sprintErr(c.ConnectionErr), sprintErr(c.AgentErr),
	)
}

// HandleTransactions is a convenience method which
// starts ReadUntilClosed and CollectUntilClosed
// and is used to automatically process and garbage collect transactions.
// Non-stun messages are dropped. Alternatively, use ReadFrom to
// manually process transactions and handle non-stun messages.
func (c *Client) HandleTransactions() {
	c.ReadUntilClosed()
	c.CollectUntilClosed()
}

// ReadUntilClosed is used to automatically process transactions.
// Non-stun messages are dropped. Alternatively, use ReadFrom to
// manually process transactions and handle non-stun messages.
func (c *Client) ReadUntilClosed() {
	c.wg.Add(1)
	go c.readUntilClosed()
}

func (c *Client) readUntilClosed() {
	defer c.wg.Done()
	buf := make([]byte, 1024)
	for {
		select {
		case <-c.close:
			return
		default:
		}
		_, _, err := c.ReadFrom(buf)
		if err == ErrAgentClosed {
			return
		}
	}
}

func closedOrPanic(err error) {
	if err == nil || err == ErrAgentClosed {
		return
	}
	panic(err)
}

// CollectUntilClosed is used to atimatically trigger transaction garbage collection.
// Alternatively, use Collect for manual collection.
func (c *Client) CollectUntilClosed() {
	c.wg.Add(1)
	go c.collectUntilClosed()
}

func (c *Client) collectUntilClosed() {
	defer c.wg.Done()
	t := time.NewTicker(c.gcRate)
	for {
		select {
		case <-c.close:
			t.Stop()
			return
		case gcTime := <-t.C:
			closedOrPanic(c.Collect(gcTime))
		}
	}
}

// Collect is used to manually trigger transaction collection.
// Alternatively, use CollectUntilClosed for automated collection.
func (c *Client) Collect(gcTime time.Time) error {
	return c.a.Collect(gcTime)
}

// Close stops internal connection and agent, returning CloseErr on error.
func (c *Client) Close() error {
	if err := c.checkInit(); err != nil {
		return err
	}
	c.closedMux.Lock()
	if c.closed {
		c.closedMux.Unlock()
		return ErrClientClosed
	}
	c.closed = true
	c.closedMux.Unlock()
	agentErr, connErr := c.a.Close(), c.c.Close()
	close(c.close)
	c.wg.Wait()
	if agentErr == nil && connErr == nil {
		return nil
	}
	return CloseErr{
		AgentErr:      agentErr,
		ConnectionErr: connErr,
	}
}

// Indicate sends indication m to server. Shorthand to Start call
// with zero deadline and callback.
func (c *Client) Indicate(m *Message) error {
	return c.Start(m, time.Time{}, nil)
}

// callbackWaitHandler blocks on wait() call until callback is called.
type callbackWaitHandler struct {
	callback  func(event Event)
	cond      *sync.Cond
	processed bool
}

func (s *callbackWaitHandler) HandleEvent(e Event) {
	if s.callback == nil {
		panic("s.callback is nil")
	}
	s.callback(e)
	s.cond.L.Lock()
	s.processed = true
	s.cond.Broadcast()
	s.cond.L.Unlock()
}

func (s *callbackWaitHandler) wait() {
	s.cond.L.Lock()
	for !s.processed {
		s.cond.Wait()
	}
	s.cond.L.Unlock()
}

func (s *callbackWaitHandler) setCallback(f func(event Event)) {
	if f == nil {
		panic("f is nil")
	}
	s.callback = f
}

func (s *callbackWaitHandler) reset() {
	s.processed = false
	s.callback = nil
}

var callbackWaitHandlerPool = sync.Pool{
	New: func() interface{} {
		return &callbackWaitHandler{
			cond: sync.NewCond(new(sync.Mutex)),
		}
	},
}

// ErrClientNotInitialized means that client connection or agent is nil.
var ErrClientNotInitialized = errors.New("client not initialized")

func (c *Client) checkInit() error {
	if c == nil || c.c == nil || c.a == nil || c.close == nil {
		return ErrClientNotInitialized
	}
	return nil
}

// Start starts transaction (if h set) and writes message to server, handler
// is called asynchronously.
func (c *Client) Start(m *Message, d time.Time, h Handler) error {
	return c.StartTo(m, c.serveraddr, d, h)
}

// StartTo starts transaction (if h set) and writes message to a specific peer, handler
// is called asynchronously.
func (c *Client) StartTo(m *Message, raddr net.Addr, d time.Time, h Handler) error {
	if err := c.checkInit(); err != nil {
		return err
	}
	c.closedMux.RLock()
	closed := c.closed
	c.closedMux.RUnlock()
	if closed {
		return ErrClientClosed
	}
	if h != nil {
		// Starting transaction only if h is set. Useful for indications.
		if err := c.a.Start(m.TransactionID, d, h); err != nil {
			return err
		}
	}
	_, err := c.c.WriteTo(m.Raw, raddr)
	if err != nil && h != nil {
		// Stopping transaction instead of waiting until deadline.
		if stopErr := c.a.Stop(m.TransactionID); stopErr != nil {
			return StopErr{
				Err:   stopErr,
				Cause: err,
			}
		}
	}
	return err
}

// Do is Start wrapper that waits until callback is called. If no callback
// provided, Indicate is called instead.
//
// Do has cpu overhead due to blocking, see BenchmarkClient_Do.
// Use Start method for less overhead.
func (c *Client) Do(m *Message, d time.Time) (*Message, error) {
	return c.DoTo(m, c.serveraddr, d)
}

// DoTo is StartTo wrapper that waits until callback is called. If no callback
// provided, Indicate is called instead.
//
// Do has cpu overhead due to blocking, see BenchmarkClient_Do.
// Use Start method for less overhead.
func (c *Client) DoTo(m *Message, raddr net.Addr, d time.Time) (*Message, error) {
	if err := c.checkInit(); err != nil {
		return nil, err
	}
	h := callbackWaitHandlerPool.Get().(*callbackWaitHandler)
	var eventErr error
	var message *Message
	h.setCallback(func(event Event) {
		eventErr = event.Error
		message = event.Message
	})
	defer func() {
		h.reset()
		callbackWaitHandlerPool.Put(h)
	}()
	if err := c.StartTo(m, raddr, d, h); err != nil {
		return nil, err
	}
	h.wait()
	return message, eventErr
}

// ReadFrom is used to keep transaction processing aliv and
// receive non-stun messages over the connection. Alternatively,
// See ReadUntilClosed for automated transaction processing.
func (c *Client) ReadFrom(b []byte) (n int, addr net.Addr, err error) {
	for {
		n, addr, err = c.c.ReadFrom(b)
		if err != nil {
			return
		}
		if !IsMessage(b[:n]) {
			return
		}
		m := new(Message)
		m.Raw = b[:n]
		if m.Decode() != nil {
			return // The caller may be able to handle the packet
		}
		err = c.a.Process(m)
		if err != nil {
			return
		}
	}
}

// WriteTo is used to write a message over the connection to the remote peer
func (c *Client) WriteTo(b []byte, addr net.Addr) (int, error) {
	return c.c.WriteTo(b, addr)
}

// LocalAddr returns the local network address.
func (c *Client) LocalAddr() net.Addr {
	return c.c.LocalAddr()
}
