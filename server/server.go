package main

import (
	"github.com/andrewdo/go-icmp-data/transport"
	"golang.org/x/net/icmp"
	"log"
	"sync"
)

func handleMessages() {
	for {
		msg, _ := transport.Receive()

		switch msg.Code {
		case transport.IcmpCodeCommandMsg:
			handleCommandMessage(msg)
			break
		default:
			log.Println("Unhandled ICMP code", msg.Code)
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