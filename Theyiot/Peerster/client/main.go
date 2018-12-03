package main

import (
	"errors"
	"flag"
	"github.com/Theyiot/Peerster/constants"
	"github.com/Theyiot/Peerster/gossiper"
	"github.com/Theyiot/Peerster/util"
	"github.com/dedis/protobuf"
	"net"
	"strings"
)

func main() {
	uiPort := flag.String("UIPort", constants.DEFAULT_PORT, "Port for the UI client")
	destination := flag.String("dest", "", "Node to send the private message")
	file := flag.String("file", "", "Node to send the private message")
	request := flag.String("request", "", "The hash of the file to download")
	msg := flag.String("msg", constants.DEFAULT_MESSAGE, "message to be sent")
	keywords := flag.String("keywords", "", "keywords to search")
	budget := flag.Int("budget", constants.DEFAULT_BUDGET, "budget of search request")

	flag.Parse()

	serverAddr,err := net.ResolveUDPAddr(constants.UDP_VERSION, constants.LOCALHOST + ":" + *uiPort)
	util.FailOnError(err)

	conn, err := net.DialUDP(constants.UDP_VERSION, nil, serverAddr)
	defer conn.Close()

	clientGossipPacket, err := createClientPacket(*destination, *file, *request, *keywords, *msg, *budget)

	if util.CheckAndPrintError(err) {
		return
	}

	bytesToSend, err := protobuf.Encode(&clientGossipPacket)
	util.FailOnError(err)

	conn.Write(bytesToSend)
}

/*
	createClientPacket is used to create the right type of client packet to be sent to the gossiper.
	It uses the provided parameters in order to determine the type of packet to return and may
	return an error if the provided parameters are inconsistent (if the type can not be inferred
	from the provided inputs of the user)
 */
func createClientPacket(destination, fileName, request, keywords, msg string, budget int) (gossiper.ClientPacket, error) {
	if isSimpleMessage(destination, fileName, request, keywords, msg) {
		return gossiper.ClientPacket{Simple: &gossiper.SimpleMessage{Contents: msg}}, nil
	} else if isPrivateMessage(destination, fileName, request, keywords, msg) {
		return gossiper.ClientPacket{Private: &gossiper.PrivateMessage{Text: msg, Destination: destination}}, nil
	} else if isFileIndexingRequest(destination, fileName, request, keywords) {
		return gossiper.ClientPacket{FileIndex: &gossiper.FileIndexMessage{FileName: fileName}}, nil
	} else if isFileRequest(fileName, request, keywords) {
		// DESTINATION IS EITHER A PEER, EITHER THE EMPTY STRING
		fileRequest := gossiper.FileRequestMessage{FileName: fileName, Destination: destination, Request: request}
		return gossiper.ClientPacket{FileRequest: &fileRequest}, nil
	} else if isSearchRequest(destination, fileName, request, keywords) {
		keywordsArray := strings.Split(strings.Replace(keywords, " ", "", -1), ",")
		searchRequestMessage := gossiper.SearchRequestMessage{Keywords: keywordsArray, Budget:uint64(budget)}
		return gossiper.ClientPacket{FileSearchRequest: &searchRequestMessage}, nil
	} else {
		return gossiper.ClientPacket{}, errors.New("unknown type of action for the values provided")
	}
}

/*
	isSimpleMessage determines and return a boolean value that tells whether the provided parameters
	correspond to a simple message or not
 */
func isSimpleMessage(destination, fileName, request, keywords, msg string) bool {
	return !isNull(msg) && isNull(destination) && isNull(fileName) && isNull(request) && isNull(keywords)
}

/*
	isPrivateMessage determines and return a boolean value that tells whether the provided parameters
	correspond to a private message or not
 */
func isPrivateMessage(destination, fileName, request, keywords, msg string) bool {
	return !isNull(msg) && !isNull(destination) && isNull(fileName) && isNull(request) && isNull(keywords)
}

/*
	isFileIndexingRequest determines and return a boolean value that tells whether the provided parameters
	correspond to a request for file indexing or not
 */
func isFileIndexingRequest(destination, fileName, request, keywords string) bool {
	return isNull(destination) && !isNull(fileName) && isNull(request) && isNull(keywords)
}

/*
	isFileRequest determines and return a boolean value that tells whether the provided parameters
	correspond to a file request or not. A file request may or not have a destination, so in both
	case it is a valid file request
 */
func isFileRequest(fileName, request, keywords string) bool {
	return !isNull(fileName) && !isNull(request) && isNull(keywords)
}

/*
	isSearchRequest determines and return a boolean value that tells whether the provided parameters
	correspond to a search request or not
 */
func isSearchRequest(destination, fileName, request, keywords string) bool {
	return isNull(destination) && isNull(fileName) && isNull(request) && !isNull(keywords)
}

/*
	isNull checks whether the provided string is empty or not
 */
func isNull(str string) bool {
	return str == ""
}