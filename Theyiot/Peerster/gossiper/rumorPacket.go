package gossiper

import (
	"fmt"
	"net"
	"time"
)

func (gossiper *Gossiper) sendRumorPacket(content string) {
	str := "CLIENT MESSAGE " + content + gossiper.Peers.String()
	gossiper.ToPrint <- str
	id, _ := gossiper.VectorClock.LoadOrStore(gossiper.Name, uint32(1))

	rumorMessage := RumorMessage{Text: content, ID: id.(uint32), Origin: gossiper.Name}
	rumorPacket := GossipPacket{Rumor: &rumorMessage}
	gossipPacketTimed := GossipPacketTimed{GossipPacket: rumorPacket, Timestamp: time.Now()}
	gossiper.Rumors.Store(fmt.Sprint(id)+"@"+gossiper.Name, gossipPacketTimed)
	gossiper.VectorClock.Store(gossiper.Name, id.(uint32)+uint32(1))


	for _, address := range gossiper.Peers.Addresses {
		gossiper.ToSend <- PacketToSend{Address: address, GossipPacket: &rumorPacket}
	}
}

func (gossiper *Gossiper) receiveRumorPacket(gossipPacket GossipPacket, addr *net.UDPAddr) {
	id, origin, msg := gossipPacket.Rumor.ID, gossipPacket.Rumor.Origin, gossipPacket.Rumor.Text
	senderAddr := addr.String()

	_, alreadyReceived := gossiper.Rumors.Load(fmt.Sprint(id) + "@" + origin)
	nextID, known := gossiper.VectorClock.LoadOrStore(origin, uint32(1))
	if alreadyReceived || (!known && id != uint32(1)) || (known && nextID.(uint32) != id) {
		return
	}

	//UPDATING RUMORS LIST
	gossiper.Rumors.Store(fmt.Sprint(id) + "@" + origin, GossipPacketTimed{ GossipPacket: gossipPacket, Timestamp: time.Now() })

	// UPDATING VECTOR CLOCK
	gossiper.VectorClock.Store(origin, nextID.(uint32) + uint32(1))

	str := "RUMOR origin " + origin + " from " + senderAddr + " ID " +
		fmt.Sprint(id) + " contents " + msg + gossiper.Peers.String()
	gossiper.ToPrint <- str

	// UPDATING DESTINATION-SEQUENCED DISTANCE VECTOR
	knownAddr, exist := gossiper.DSDV.Load(origin)
	if !exist || knownAddr.(*net.UDPAddr).String() != senderAddr && origin != gossiper.Name {
		gossiper.DSDV.Store(origin, addr)
		gossiper.ToPrint <- "DSDV " + origin + " " + senderAddr
	}

	// SENDING STATUS
	gossiper.ToSend <- PacketToSend{Address: addr,
		GossipPacket: &GossipPacket{Status: gossiper.constructStatuses()}}

	// WE RUMORMONGER ONLY IF WE KNOW ONE OTHER PEER THAT SENT THE RUMOR
	if gossiper.Peers.GetSize() > 1 {
		gossiper.rumormonger(gossipPacket, gossiper.Peers.ChooseRandomPeerExcept(senderAddr))
	}
}