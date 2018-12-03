package gossiper

import (
	"flag"
	"github.com/Theyiot/Peerster/constants"
	"github.com/Theyiot/Peerster/util"
	"net"
	"strings"
	"sync"
)


/*
	StartGossip takes care of launching the different functionalities of our peerster and to parse the
	parameters provided by the client
 */
func StartGossip() {
	//FLAGS DEFINITION
	uiPort := flag.String("UIPort", constants.DEFAULT_PORT, "Port for the UI client")
	gossipAddr := flag.String("gossipAddr", constants.DEFAULT_GOSSIP_ADDR, "ip:port for the receiver")
	name := flag.String("name", constants.DEFAULT_NAME, "name of the gossiper")
	peersToSplit := flag.String("peers", "", "comma-separated list of peers of the form ip:port")
	simple := flag.Bool("simple", false, "run gossiper in simple broadcast mode")
	rtimer := flag.Uint("rtimer", 0, "Time between each route rumor")
	flag.Parse()

	peers := util.CreateAddrSet(strings.Replace(*peersToSplit, " ", "", -1))

	uiServerAddr, err := net.ResolveUDPAddr(constants.UDP_VERSION, constants.LOCALHOST + ":" + *uiPort)
	util.FailOnError(err)
	gossipServerAddr, err := net.ResolveUDPAddr(constants.UDP_VERSION, *gossipAddr)
	util.FailOnError(err)

	uiServer, err := net.ListenUDP(constants.UDP_VERSION, uiServerAddr)
	util.FailOnError(err)
	defer uiServer.Close()
	gossipServer, err := net.ListenUDP(constants.UDP_VERSION, gossipServerAddr)
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
		Privates:      sync.Map{},
		DSDV:          sync.Map{},
		ReceivingFile: sync.Map{},
		IndexedFiles:  sync.Map{},
		SearchedFiles: sync.Map{},
		SearchRequests:sync.Map{},
		Acks:          sync.Map{},
		ActiveSearches:	util.CreateFullMatchesSet(),
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