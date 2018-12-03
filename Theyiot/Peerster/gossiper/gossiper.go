package gossiper

import (
	"flag"
	"github.com/Theyiot/Peerster/util"
	"net"
	"strings"
	"sync"
)

func StartGossip() {
	//FLAGS DEFINITION
	uiPort := flag.String("UIPort", "8080", "Port for the UI client")
	gossipAddr := flag.String("gossipAddr", "127.0.0.1:5000", "ip:port for the receiver")
	name := flag.String("name", "nodeA", "name of the gossiper")
	peersToSplit := flag.String("peers", "", "comma-separated list of peers of the form ip:port")
	simple := flag.Bool("simple", false, "run gossiper in simple broadcast mode")
	rtimer := flag.Uint("rtimer", 0, "Time between each route rumor")
	flag.Parse()

	peers := util.CreateAddrSet(strings.Replace(*peersToSplit, " ", "", -1))

	uiServerAddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:" + *uiPort)
	util.FailOnError(err)
	gossipServerAddr, err := net.ResolveUDPAddr("udp4", *gossipAddr)
	util.FailOnError(err)

	uiServer, err := net.ListenUDP("udp4", uiServerAddr)
	util.FailOnError(err)
	defer uiServer.Close()
	gossipServer, err := net.ListenUDP("udp4", gossipServerAddr)
	util.FailOnError(err)
	defer gossipServer.Close()

	gossiper := Gossiper{
		UIServer:      uiServer,
		GossipAddr:    *gossipAddr,
		GossipServer:  gossipServer,
		Name:          *name,
		Simple:        *simple,
		Peers:         peers,
		VectorClock:   sync.Map{},
		Rumors:        sync.Map{},
		DSDV:          sync.Map{},
		Privates:      sync.Map{},
		ReceivingFile: sync.Map{},
		IndexedFiles:  sync.Map{},
		SearchedFiles: sync.Map{},
		SearchRequests:sync.Map{},
		Acks:          sync.Map{},
		ActiveSearches:	util.CreateStringSet(),
		ToPrint:       make(chan string),
		ToSend:        make(chan PacketToSend),
	}

	//UI COMMUNICATION
	go gossiper.handleClient()

	//GOSSIP COMMUNICATION
	go gossiper.handleGossip()

	//ROUTE RUMOR
	if *rtimer > 0 {
		go gossiper.routeRumor(*rtimer)
	}

	//ANTI-ENTROPY
	if !gossiper.Simple {
		go gossiper.antiEntropy()
	}

	//SENDING PACKETS
	go gossiper.sendPacket()

	//OPENING WEB SERVER
	go gossiper.StartWebServer(*uiPort)

	//PRINTING CONTENTS
	gossiper.printMessages()
}