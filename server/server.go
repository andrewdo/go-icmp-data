package main

import (
	"fmt"
	"github.com/andrewdo/go-icmp-data/transport"
	"github.com/miguelsandro/curve25519-go/axlsign"
	"golang.org/x/net/icmp"
	"log"
	"math/rand"
	"sync"
)

const(
	readBufferSize         		= 1500
	codeNextCommandRequest		= 15
	codeNextCommandResponse     = 16
)

func handleMessages() {
	for {
		msg, peer := transport.Receive()

		switch msg.Code {
		case codeNextCommandRequest:
			transport.Send(peer, codeNextCommandResponse, []byte("tttt"))
			break
		default:
			log.Println("Unhandled ICMP code", msg.Code)
		}
	}
}

func handlePreKeyRequest(m *icmp.Message) {
	// respond with a signed Pre Key
}

func main() {
	p := make([]byte, 32)
	_, err := rand.Read(p)
	if err != nil {
		log.Fatal(err)
	}

	k := axlsign.GenerateKeyPair(p)
	fmt.Println(k.PrivateKey)

	var wg sync.WaitGroup
	wg.Add(1)

	go handleMessages()

	wg.Wait()
}