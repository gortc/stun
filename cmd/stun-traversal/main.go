package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/gortc/stun"
)

var (
	network = flag.String("network", stun.DefaultNet, "Stun network type")
	server  = flag.String("server", stun.DefaultSTUNServer, "Stun server address")
	local   = flag.String("local", "", "Local network address")
)

const (
	pingMsg         = "ping"
	pongMsg         = "pong"
	keepaliveMillis = 500
)

func main() {
	flag.Parse()

	c, err := stun.Dial(*network, *local, *server)
	if err != nil {
		log.Fatalln("dial:", err)
	}

	defer c.Close()

	log.Printf("Listening on %s\n", c.LocalAddr())

	// Start listening to start transaction handling
	messageChan := readUntilClosed(c)

	err = getPubAddr(c)
	if err != nil {
		log.Fatalln("get pub addr:", err)
	}

	var peerAddr net.Addr
	peerAddrChan := getPeerAddr()

	keepalive := time.Tick(keepaliveMillis * time.Millisecond)
	keepaliveMsg := pingMsg

	var quit <-chan time.Time

	gotPong := false
	sentPong := false

	for {
		select {
		case message, ok := <-messageChan:
			if !ok {
				return
			}

			switch {
			case string(message) == pingMsg:
				keepaliveMsg = pongMsg

			case string(message) == pongMsg:
				if !gotPong {
					log.Println("Received pong message.")
				}

				// One client may skip sending ping if it receives
				// a ping message before knowning the peer address.
				keepaliveMsg = pongMsg

				gotPong = true

			default:
				log.Fatalln("unknown message", message)
			}

		case addr := <-peerAddrChan:
			peerAddr = addr

		case <-keepalive:
			// Keep NAT binding alive using STUN server or the peer once it's known
			if peerAddr == nil {
				err = c.Indicate(stun.MustBuild(stun.TransactionID, stun.BindingRequest))
			} else {
				_, err = c.WriteTo([]byte(keepaliveMsg), peerAddr)
				if keepaliveMsg == pongMsg {
					sentPong = true
				}
			}

			if err != nil {
				log.Fatalln("keepalive:", err)
			}

		case <-quit:
			c.Close()
		}

		if quit == nil && gotPong && sentPong {
			log.Println("Success! Quitting in two seconds.")
			quit = time.After(2 * time.Second)
		}
	}
}

func getPubAddr(c *stun.Client) error {
	deadline := time.Now().Add(time.Second * 5)
	message, err := c.Do(stun.MustBuild(stun.TransactionID, stun.BindingRequest), deadline)
	if err != nil {
		return fmt.Errorf("do: %v", err)
	}
	var publicAddr stun.XORMappedAddress
	if err := publicAddr.GetFrom(message); err != nil {
		return fmt.Errorf("get from: %v", err)
	}

	log.Printf("My public address: %s\n", publicAddr)

	return nil
}

func getPeerAddr() <-chan net.Addr {
	result := make(chan net.Addr)

	go func() {
		reader := bufio.NewReader(os.Stdin)
		log.Println("Enter remote peer address:")
		for {
			peer, _ := reader.ReadString('\n')
			addr, err := stun.ResolveAddr(*network, strings.Trim(peer, " \r\n"))
			if err != nil {
				log.Println("Invalid address:", err)
				continue
			}
			result <- addr
			return
		}
	}()

	return result
}

func readUntilClosed(conn stun.PacketConn) <-chan []byte {
	messages := make(chan []byte)
	go func() {
		for {
			buf := make([]byte, 1024)

			n, _, err := conn.ReadFrom(buf)
			if err != nil {
				close(messages)
				return
			}

			messages <- buf[:n]
		}
	}()
	return messages
}
