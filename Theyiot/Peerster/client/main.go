package main

import (
	"flag"
	"github.com/Theyiot/Peerster/gossiper"
	"github.com/Theyiot/Peerster/util"
	"github.com/dedis/protobuf"
	"net"
	"strconv"
)

func main() {
	uiPort := flag.Int("UIPort", 8080, "Port for the UI client")
	dest := flag.String("dest", "", "Node to send the private message")
	file := flag.String("file", "", "Node to send the private message")
	request := flag.String("request", "", "The hash of the file to download")
	msg := flag.String("msg", "default message", "message to be sent")
	keywords := flag.String("keywords", "", "keywords to search")
	budget := flag.Int("budget", 2, "budget of search request")

	flag.Parse()

	if *msg == "" && *file == "" {
		println("You must either send a message or choose a file (to download or to load)")
		return
	}
	serverAddr,err := net.ResolveUDPAddr("udp4", "127.0.0.1:" + strconv.Itoa(*uiPort))
	util.FailOnError(err)

	conn, err := net.DialUDP("udp4", nil, serverAddr)
	defer conn.Close()

	var clientGossipPacket gossiper.ClientGossipPacket
	if *file != "" { //FILE MESSAGE
		if *request == "" {
			if *dest != "" {
				println("cannot request a chunk from peer " + *dest + " if you don't provide the chunk's hash")
			}
			clientGossipPacket = gossiper.ClientGossipPacket{File: &gossiper.FileMessage{FileName: *file}}
		} else {
			if !gossiper.IsHexHash(*request) {
				println("request is not a hash, should be a 64 hexadecimal " + "characters string, " +
					"but was " + *request)
			}
			fileRequest := gossiper.FileMessage{FileName: *file, Destination: *dest, Request: *request}
			clientGossipPacket = gossiper.ClientGossipPacket{File: &fileRequest}
		}
	} else if *dest != "" { //PRIVATE MESSAGE
		privateMessage := gossiper.PrivateMessage{Text: *msg, Destination: *dest}
		clientGossipPacket = gossiper.ClientGossipPacket{Private: &privateMessage}
	} else { //SIMPLE MESSAGE
		simpleMessage := gossiper.SimpleMessage{Contents: *msg}
		clientGossipPacket = gossiper.ClientGossipPacket{Simple: &simpleMessage}
	}

	bytesToSend, err := protobuf.Encode(&clientGossipPacket)
	util.FailOnError(err)

	conn.Write(bytesToSend)
}