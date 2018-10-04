package main

import (
	"flag"
	"fmt"
	"net"
	"time"
	"strconv"
)

type SimpleMessage struct {
	OriginalName string
	RelayPeerAddr string
	Contents string
}

type GossipPacket struct {
	Simple *SimpleMessage
}

func CheckError(err error) {
	if err  != nil {
		fmt.Println("Error: " , err)
	}
}

func main() {
	uiPort := flag.Int("UIPort", 8080, "Port for the UI client")
	msg := flag.String("msg", "default message", "message to be sent")

	flag.Parse()

	ServerAddr,err := net.ResolveUDPAddr("udp", *gossipAddr)
	CheckError(err)

	LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	CheckError(err)

	Conn, err := net.DialUDP("udp", LocalAddr, ServerAddr)
	CheckError(err)

	defer Conn.Close()
	i := 0
	for {
		msg := strconv.Itoa(i)
		i++
		buf := []byte(msg)
		_,err := Conn.Write(buf)
		if err != nil {
			fmt.Println(msg, err)
		}
		time.Sleep(time.Second * 1)
	}
}