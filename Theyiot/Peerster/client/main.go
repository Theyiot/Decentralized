package main

import (
	"errors"
	"flag"
	"github.com/Theyiot/Peerster/gossiper"
	"github.com/Theyiot/Peerster/util"
	"github.com/dedis/protobuf"
	"net"
	"strconv"
	"strings"
)

func main() {
	uiPort := flag.Int("UIPort", gossiper.DEFAULT_PORT, "Port for the UI client")
	destination := flag.String("dest", "", "Node to send the private message")
	file := flag.String("file", "", "Node to send the private message")
	request := flag.String("request", "", "The hash of the file to download")
	msg := flag.String("msg", "default message", "message to be sent")
	keywords := flag.String("keywords", "", "keywords to search")
	budget := flag.Int("budget", gossiper.DEFAULT_BUDGET, "budget of search request")

	flag.Parse()

	serverAddr,err := net.ResolveUDPAddr("udp4", "127.0.0.1:" + strconv.Itoa(*uiPort))
	util.FailOnError(err)

	conn, err := net.DialUDP("udp4", nil, serverAddr)
	defer conn.Close()

	clientGossipPacket, err := createGossipPacket(*destination, *file, *request, *keywords, *msg, *budget)

	if util.CheckAndPrintError(err) {
		return
	}

	bytesToSend, err := protobuf.Encode(&clientGossipPacket)
	util.FailOnError(err)

	conn.Write(bytesToSend)
}

func createGossipPacket(destination, fileName, request, keywords, msg string, budget int) (gossiper.ClientGossipPacket, error) {
	if isSimpleMessage(destination, fileName, request, keywords, msg) {
		return gossiper.ClientGossipPacket{Simple: &gossiper.SimpleMessage{Contents: msg}}, nil
	} else if isPrivateMessage(destination, fileName, request, keywords, msg) {
		return gossiper.ClientGossipPacket{Private: &gossiper.PrivateMessage{Text: msg, Destination: destination}}, nil
	} else if isFileIndexingRequest(destination, fileName, request, keywords) {
		return gossiper.ClientGossipPacket{FileIndex: &gossiper.FileIndexMessage{FileName: fileName}}, nil
	} else if isFileRequest(fileName, request, keywords) {
		// DESTINATION IS EITHER A PEER; EITHER THE EMPTY STRING
		fileRequest := gossiper.FileRequestMessage{FileName: fileName, Destination: destination, Request: request}
		return gossiper.ClientGossipPacket{FileRequest: &fileRequest}, nil
	} else if isSearchRequest(destination, fileName, request, keywords) {
		keywordsArray := strings.Split(strings.Replace(keywords, " ", "", -1), ",")
		searchRequestMessage := gossiper.SearchRequestMessage{Keywords: keywordsArray, Budget:uint64(budget)}
		return gossiper.ClientGossipPacket{FileSearchRequest: &searchRequestMessage}, nil
	} else {
		return gossiper.ClientGossipPacket{}, errors.New("unknown type of action for the values provided")
	}
}

func isSimpleMessage(destination, fileName, request, keywords, msg string) bool {
	return !isNull(msg) && isNull(destination) && isNull(fileName) && isNull(request) && isNull(keywords)
}

func isPrivateMessage(destination, fileName, request, keywords, msg string) bool {
	return !isNull(msg) && !isNull(destination) && isNull(fileName) && isNull(request) && isNull(keywords)
}

func isFileIndexingRequest(destination, fileName, request, keywords string) bool {
	return isNull(destination) && !isNull(fileName) && isNull(request) && isNull(keywords)
}

func isFileRequest(fileName, request, keywords string) bool {
	return !isNull(fileName) && !isNull(request) && isNull(keywords)
}

func isSearchRequest(destination, fileName, request, keywords string) bool {
	return isNull(destination) && isNull(fileName) && isNull(request) && !isNull(keywords)
}

func isNull(str string) bool {
	return str == ""
}