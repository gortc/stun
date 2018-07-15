package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/gortc/stun"
)

func main() {
	var (
		err error
	)

	serverAddr := fmt.Sprintf("stun-server:%d", stun.DefaultPort)

	fmt.Println("START")
	for i := 0; i < 10; i++ {
		_, err = net.ResolveUDPAddr("udp", serverAddr)
		if err == nil {
			break
		}
		time.Sleep(time.Millisecond * 300 * time.Duration(i))
	}
	if err != nil {
		log.Fatalln("too many attempts to resolve:", err)
	}

	fmt.Println("DIALING", serverAddr)
	client, err := stun.Dial("", "", serverAddr)
	if err != nil {
		log.Fatalln("failed to dial:", err)
	}

	client.HandleTransactions()

	laddr := client.LocalAddr()
	fmt.Println("LISTEN ON", laddr)

	request, err := stun.Build(stun.BindingRequest, stun.TransactionID)
	if err != nil {
		log.Fatalln("failed to build:", err)
	}
	timeout := time.Second
	deadline := time.Now().Add(timeout)

	response, err := client.Do(request, deadline)
	if err != nil {
		log.Fatalln("failed to Do:", err)
	}
	if response.Type != stun.BindingSuccess {
		log.Fatalln("bad message", response)
	}
	var xorMapped stun.XORMappedAddress
	if err = response.Parse(&xorMapped); err != nil {
		log.Fatalln("failed to parse xor mapped address:", err)
	}
	lport, err := getPort(laddr)
	if err != nil {
		log.Fatalln("get port:", laddr)
	}
	if lport != xorMapped.Port {
		log.Fatalln(laddr, "!=", xorMapped)
	}
	fmt.Println("OK", response, "GOT", xorMapped)

	if err := client.Close(); err != nil {
		log.Fatalln("failed to close client:", err)
	}
}

// Get the port of a net.Addr
func getPort(addr net.Addr) (int, error) {
	switch a := addr.(type) {
	case *net.UDPAddr:
		return a.Port, nil
	case *net.TCPAddr:
		return a.Port, nil
	default:
		return -1, stun.ErrNet
	}
}
