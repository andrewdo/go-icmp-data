package transport

import (
	"crypto/md5"
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
	icmpCodeAck				= 0
	icmpCodePreKeyRequest  	= 15
	icmpCodePreKeyResponse 	= 16
	icmpCodeMessage        	= 8
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

	// TODO: message chunks and use ID for concurrency
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

	// keep sending message until we get an ack
	for retries := numRetries + 1; retries > 0; retries-- {
		if _, err := conn.WriteTo(wb, dest); err != nil {
			panic(err)
		}

		if waitForAck(conn, dest, msg) {
			return
		}

		log.Println("Retrying message", dest, code, msg)
	}

	log.Fatal("Permanent failure sending message", dest, code, msg)
}

func waitForAck(conn *icmp.PacketConn, dest net.Addr, msg []byte) bool {
	ch := make(chan bool, 1)
	go func() {
		rb := make([]byte, readBufferSize)
		n, peer, err := conn.ReadFrom(rb)
		if err != nil {
			log.Println(err)
			ch <- false
		}
		rm, err := icmp.ParseMessage(1, rb[:n])
		if err != nil {
			log.Println(err)
			ch <- false
		}

		if rb, ok := rm.Body.(*icmp.Echo); ok {
			var sig [16]byte
			_ = copy(sig[:], rb.Data)
			// && nb == 16 && md5.Sum(msg) == sig
			if peer == dest && rm.Code == icmpCodeAck  {
				ch <- true
			}
		}

		ch <- false
	}()

	go func() {
		time.Sleep(timeoutSeconds * time.Second)
		ch <- false
	}()

	select{
	case s := <-ch:
		if s {
			return true
		}
	}

	return false
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

		// send an ack
		if rb, ok := rm.Body.(*icmp.Echo); ok {
			sig := md5.Sum(rb.Data)
			go Send(peer, icmpCodeAck, sig[:])
		}

		return rm, peer
	}
}
