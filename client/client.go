package main

import (
	"github.com/andrewdo/go-icmp-data/transport"
	"log"
	"net"
)

func main() {
	transport.Send(&net.IPAddr{IP: net.ParseIP("127.0.0.1")}, 15, []byte("FFFF"))
	x, y := transport.Receive()

	log.Println(x, y)
}