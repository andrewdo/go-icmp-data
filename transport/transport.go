package transport

import (
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"log"
	"math/rand"
	"net"
	"os"
)

const(
	readBufferSize         = 1500
	icmpCodePreKeyRequest  = 15
	icmpCodePreKeyResponse = 16
	icmpCodeMessage        = 8
)

func getConnection() *icmp.PacketConn {
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	return conn
}

func Send(dest net.Addr, code int, msg []byte) {
	conn := getConnection()
	defer conn.Close()

	// TODO: message chunks
	id := rand.Int()
	m := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: code,
		Body: &icmp.Echo{
			ID: id,
			Seq: 1,
			Data: msg,
		},
	}
	wb, err := m.Marshal(nil)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := conn.WriteTo(wb, dest); err != nil {
		panic(err)
	}
}

func Receive() (*icmp.Message, net.Addr) {
	conn := getConnection()
	defer conn.Close()

	for {
		rb := make([]byte, readBufferSize)
		n, peer, err := conn.ReadFrom(rb)
		if err != nil {
			log.Fatal(err)
		}
		rm, err := icmp.ParseMessage(1, rb[:n])
		if err != nil {
			log.Fatal(err)
		}

		return rm, peer
	}
}
