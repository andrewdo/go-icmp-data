package main

import (
	"bufio"
	"github.com/andrewdo/go-icmp-data/transport"
	"log"
	"net"
	"os"
)

func main() {
	addr, err := net.LookupIP("server")
	if err != nil {
		panic(err)
	}
	a := &net.IPAddr{IP: addr[0]}

	for {
		reader := bufio.NewReader(os.Stdin)
		cmd, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Sending command", cmd)

		transport.Send(a, transport.IcmpCodeCommandMsg, []byte(cmd), true)
	}
}