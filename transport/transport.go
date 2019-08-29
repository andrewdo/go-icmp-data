package transport

import (
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"log"
	"math/rand"
	"net"
	"os"
	"time"
)

const(
	readBufferSize         	= 1500
	timeoutSeconds			= 5
	numRetries				= 5
	IcmpCodeAck				= 0
	IcmpCodeCommandMsg		= 15
	IcmpCodeCommandReply	= 16
)

type Packet struct {
	From *net.Addr
	Message *icmp.Message
}

func getConnection() *icmp.PacketConn {
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	return conn
}

func Send(dest net.Addr, msg []byte, code int) *Packet {
	conn := getConnection()
	defer conn.Close()

	// TODO: message chunks and use ID for concurrency
	id := rand.Int()
	m := icmp.Message{
		Type: ipv4.ICMPTypeEchoReply,
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

	// keep sending the message until we get a response
	for retries := numRetries + 1; retries > 0; retries-- {
		if _, err := conn.WriteTo(wb, dest); err != nil {
			panic(err)
		}

		if code == IcmpCodeCommandReply {
			return nil
		}

		if r := waitForReply(conn, dest, msg); r != nil {
			return r
		}

		rand.Seed(time.Now().UnixNano())
		time.Sleep(time.Duration(rand.Intn(5) + 1) * time.Second)

		log.Println("Retrying message", dest, code, msg)
	}

	log.Fatal("Permanent failure sending message", dest, code, msg)

	return nil
}

func waitForReply(conn *icmp.PacketConn, dest net.Addr, msg []byte) *Packet {
	ch := make(chan *Packet, 1)
	go func() {
		rb := make([]byte, readBufferSize)
		n, peer, err := conn.ReadFrom(rb)
		if err != nil {
			log.Println(err)
			ch <- nil
		}
		rm, err := icmp.ParseMessage(1, rb[:n])
		if err != nil {
			log.Println(err)
			ch <- nil
		}

		if rb, ok := rm.Body.(*icmp.Echo); ok {
			if peer.String() == dest.String() && rm.Code == IcmpCodeCommandReply  {
				log.Println("Received reply", string(rb.Data))
				ch <- &Packet{
					From:    &peer,
					Message: rm,
				}
			} else {
				log.Println("Received message was not the response")
			}
		} else {
			log.Println("Failed to parse message as Echo body")
		}

		ch <- nil
	}()

	go func() {
		time.Sleep(timeoutSeconds * time.Second)
		ch <- nil
	}()

	select{
	case p := <-ch:
		return p
	}
}

func Receive(ch chan Packet) {
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

		ch <- Packet{
			From:    &peer,
			Message: rm,
		}
	}
}
