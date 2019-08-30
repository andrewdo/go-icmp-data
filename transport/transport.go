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
	readBufferSize			= 1500
	timeoutSeconds			= 5
	numRetries				= 5
	IcmpCodeCommandRequest	= 15
	IcmpCodeCommandReply	= 16
	IcmpCodeCommandOutput	= 8
	IcmpCodeAck				= 0
)

type Packet struct {
	From *net.Addr
	Message *icmp.Message
	Body *icmp.Echo
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
	// TODO: message chunks and use ID for concurrency
	var t ipv4.ICMPType
	if code == IcmpCodeCommandRequest || code == IcmpCodeCommandOutput {
		t = ipv4.ICMPTypeEcho
	} else {
		t = ipv4.ICMPTypeEchoReply
	}
	t = ipv4.ICMPTypeEchoReply
	id := rand.Int()

	m := &icmp.Message{
		Type: t,
		Code: code,
		Body: &icmp.Echo{
			ID: id,
			Seq: 1,
			Data: msg,
		},
	}

	return send(dest, m, code != IcmpCodeCommandReply && code != IcmpCodeAck)
}

func send(d net.Addr, m *icmp.Message, wait bool) *Packet {
	conn := getConnection()
	defer conn.Close()

	wb, err := m.Marshal(nil)
	if err != nil {
		log.Fatal(err)
	}

	for retries := numRetries + 1; retries > 0; retries-- {
		if _, err := conn.WriteTo(wb, d); err != nil {
			panic(err)
		}

		log.Println("Sent message", m.Body)

		if !wait {
			return nil
		}

		if r := waitForReply(conn, d); r != nil {
			return r
		}

		time.Sleep(time.Duration(rand.Intn(timeoutSeconds) + 1) * time.Second)

		log.Println("Retrying message", d, m.Code, m.Body)
	}

	log.Fatal("Failed after max retries", numRetries)
	return nil
}

func waitForReply(conn *icmp.PacketConn, dest net.Addr) *Packet {
	ch := make(chan *Packet, 1)
	go func() {
		rb := make([]byte, readBufferSize)
		for {
			n, peer, err := conn.ReadFrom(rb)
			if err != nil {
				log.Println(err)
				ch <- nil
				return
			}
			rm, err := icmp.ParseMessage(1, rb[:n])
			if err != nil {
				log.Println(err)
				ch <- nil
				return
			}

			if rb, ok := rm.Body.(*icmp.Echo); ok {
				if peer.String() == dest.String() && (rm.Code == IcmpCodeCommandReply || rm.Code == IcmpCodeAck)  {
					log.Println("Received reply", string(rb.Data))
					ch <- &Packet{
						From:    	&peer,
						Message:	rm,
						Body:		rb,
					}
					return
				} else {
					log.Println("Received message was not the response", peer.String(), dest.String(), rm.Code)
				}
			} else {
				log.Println("Failed to parse message as Echo body")
			}

			log.Println("Skipping message", rm)
		}
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

func Receive(ch chan *Packet) {
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

		if b, ok := rm.Body.(*icmp.Echo); ok {
			ch <- &Packet{
				From:    	&peer,
				Message:	rm,
				Body:		b,
			}
		}
	}
}
