package main

import (
	"github.com/andrewdo/go-icmp-data/transport"
	"golang.org/x/net/icmp"
	"log"
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
					handleCommandMessage(p.Message)
					break
				default:
					log.Println("Unhandled ICMP code", p.Message.Code)
				}

				break
			}
		}
	}
}

func handleCommandMessage(msg *icmp.Message) {
	if b, ok := msg.Body.(*icmp.Echo); ok {
		cmd := string(b.Data)
		log.Println("Got command", cmd)
		return
	}

	log.Println("Ignoring ICMP message", msg)
}

func main() {
	var wg sync.WaitGroup
	wg.Add(1)

	go handleMessages()

	wg.Wait()
}