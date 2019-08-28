package main

import (
"bufio"
"github.com/andrewdo/go-icmp-data/transport"
"log"
"net"
"os"
)

func main() {
	addr, err := net.LookupIP("client")
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

		transport.Send(a, 15, []byte(cmd), true)
	}
}