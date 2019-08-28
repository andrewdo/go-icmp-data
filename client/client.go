package main

import (
	"github.com/andrewdo/go-icmp-data/transport"
	"log"
	"net"
	"time"
)

func main() {
	addr, err := net.LookupIP("server")
	if err != nil {
		panic(err)
	}

	for {
		transport.Send(&net.IPAddr{IP: addr[0]}, 15, []byte("FFFF"))
		x, y := transport.Receive()

		log.Println(x, y)

		time.Sleep(10 * time.Second)
	}
}