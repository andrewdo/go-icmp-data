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
)

type Packet struct {
	From *net.Addr
	Message *icmp.Message
	Payload *Payload
}

type Payload struct {
	Type byte
	Data []byte
}

func (p *Packet) Respond(pl *Payload) *Packet {
	return send(*p.From, pl, true, false)
}

func (p *Payload) getBytes() []byte {
	if p.Data == nil {
		return []byte{p.Type}
	}

	return append([]byte{p.Type}, p.Data...)
}

func getConnection() *icmp.PacketConn {
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	return conn
}

func Send(dest net.Addr, p *Payload) *Packet {
	// TODO: message chunks and use ID for concurrency
	return send(dest, p, false, true)
}

func send(d net.Addr, p *Payload, isReply bool, wait bool) *Packet {
	conn := getConnection()
	defer conn.Close()

	m := &icmp.Message{
		Code: 0,
		Body: &icmp.Echo{
			ID: rand.Int(),
			Seq: 1,
			Data: p.getBytes(),
		},
	}

	if isReply {
		m.Type = ipv4.ICMPTypeEchoReply
	} else {
		m.Type = ipv4.ICMPTypeEcho
	}

	wb, err := m.Marshal(nil)
	if err != nil {
		log.Fatal(err)
	}

	for retries := numRetries + 1; retries > 0; retries-- {
		if _, err := conn.WriteTo(wb, d); err != nil {
			panic(err)
		}

		log.Println("Sent message", m.Checksum, m.Body)

		if !wait {
			return nil
		}

		if r := waitForReply(conn, d, p); r != nil {
			return r
		}

		time.Sleep(time.Duration(rand.Intn(timeoutSeconds) + 1) * time.Second)

		log.Println("Retrying message", d, m.Code, m.Body)
	}

	log.Fatal("Failed after max retries", numRetries)
	return nil
}

func waitForReply(conn *icmp.PacketConn, dest net.Addr, p *Payload) *Packet {
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
				if len(rb.Data) == 0 {
					log.Println("Skipping empty payload", rb)
					continue
				}
				if peer.String() == dest.String() && rb.Data[0] == p.Type  {
					log.Println("Received reply", string(rb.Data))

					ch <- &Packet{
						From:    	&peer,
						Message:	rm,
						Payload:	&Payload{
							Type:	rb.Data[0],
							Data:	rb.Data[1:],
						},
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
			if len(b.Data) == 0 {
				log.Println("Skipping empty payload", rm)
				continue
			}
			ch <- &Packet{
				From:    	&peer,
				Message:	rm,
				Payload:	&Payload{
					Type:	b.Data[0],
					Data:	b.Data[:1],
				},
			}
		}
	}
}
