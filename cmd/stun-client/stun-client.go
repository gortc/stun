package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/gortc/stun"
)

var (
	network = flag.String("network", stun.DefaultNet, "Stun network type")
	server  = flag.String("server", stun.DefaultSTUNServer, "Stun server address")
	local   = flag.String("local", "", "Local network address")
)

func main() {
	flag.Parse()

	c, err := stun.Dial(*network, *local, *server)
	if err != nil {
		log.Fatal("dial:", err)
	}
	c.HandleTransactions()

	defer func() {
		if err := c.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	deadline := time.Now().Add(time.Second * 5)
	message, err := c.Do(stun.MustBuild(stun.TransactionID, stun.BindingRequest), deadline)
	if err != nil {
		log.Fatal("do:", err)
	}
	var xorAddr stun.XORMappedAddress
	if err := xorAddr.GetFrom(message); err != nil {
		log.Fatalln(err)
	}
	fmt.Println(xorAddr)
}
