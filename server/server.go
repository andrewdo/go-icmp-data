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

const messageTypeCommandRequest = 0x01
const messageTypeCommandOutput = 0x02

func serveCommands(cmdCh chan string, outCh chan string) {
	cmds := make([]string, 0)
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

func handlePacket(p *transport.Packet, cmds []string, outCh chan string) []string {
	switch p.Payload.Type {
	case messageTypeCommandRequest:
		// reply with next command, if any
		if len(cmds) > 0 {
			p.Respond(&transport.Payload{
				Type: messageTypeCommandRequest,
				Data: []byte(cmds[0]),
			})
			cmds = cmds[1:]
		} else {
			p.Respond(&transport.Payload{
				Type: messageTypeCommandRequest,
				Data: []byte{},
			})
		}
		break
	case messageTypeCommandOutput:
		// pass the command out thru the channel
		log.Println(p.Payload.Data)
		outCh <- string(p.Payload.Data)
		p.Respond(&transport.Payload{
			Type: messageTypeCommandOutput,
			Data: []byte{},
		})
		break
	}

	return cmds
}

func main() {
	cmdCh := make(chan string, 0)
	outCh := make(chan string, 0)
	defer close(cmdCh)
	defer close(outCh)

	var wg sync.WaitGroup
	wg.Add(2)

	go serveCommands(cmdCh, outCh)
	go func() {
		for {
			fmt.Print("> ")
			reader := bufio.NewReader(os.Stdin)
			cmd, err := reader.ReadString('\n')
			if err != nil {
				log.Fatal(err)
			}

			cmdCh <- strings.TrimSpace(cmd)
			o := <-outCh
			fmt.Println(o)
		}
	}()

	wg.Wait()
}