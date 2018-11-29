package gossiper

import (
	"fmt"
	"net"
	"time"
)

func (gossiper *Gossiper) receivePrivatePacket(gossipPacket GossipPacket, senderAddr string) {
	dest, origin := gossipPacket.Private.Destination, gossipPacket.Private.Origin
	if dest == gossiper.Name {
		str := "PRIVATE origin " + origin + " hop-limit " + fmt.Sprint(gossipPacket.Private.HopLimit) +
			" contents " + gossipPacket.Private.Text + gossiper.Peers.String()
		gossiper.ToPrint <- str

		gossiper.updatePrivates(gossipPacket, origin)
		return
	}

	gossiper.forwardPrivatePacket(gossipPacket, dest, senderAddr)
}

func (gossiper *Gossiper) sendPrivatePacket(content string, dest string) {
	addr, exist := gossiper.DSDV.Load(dest)
	if !exist {
		println("ERROR : Trying to send a private message to an unknown peer, discarding this message")
		return
	}
	str := "CLIENT MESSAGE " + content + gossiper.Peers.String()
	gossiper.ToPrint <- str

	privateMsg := PrivateMessage{Origin: gossiper.Name, Text: content, ID: 0, Destination: dest, HopLimit: DEFAULT_HOP_LIMIT}
	gossipPacket := GossipPacket{Private: &privateMsg}
	gossiper.ToSend <- PacketToSend{GossipPacket: &gossipPacket, Address: addr.(*net.UDPAddr)}

	gossiper.updatePrivates(gossipPacket, dest)
}

func (gossiper *Gossiper) updatePrivates(gossipPacket GossipPacket, peerName string) {
	gossipPacketTimed := GossipPacketTimed{GossipPacket: gossipPacket, Timestamp: time.Now()}
	destMessages, exist := gossiper.Privates.Load(peerName)
	var messages []GossipPacketTimed
	if !exist { //WE BEGIN OUR PRIVATE DISCUSSION WITH "DEST"
		messages = []GossipPacketTimed{gossipPacketTimed}
	} else { //WE ALREADY HAVE A DISCUSSION WITH "DEST"
		messages = append(destMessages.([]GossipPacketTimed), gossipPacketTimed)
	}
	gossiper.Privates.Store(peerName, messages)
}

func (gossiper *Gossiper) forwardPrivatePacket(gossipPacket GossipPacket, dest string, senderAddr string) {
	//WE DECREASE AND DISCARD INVALID PACKET
	gossipPacket.Private.HopLimit--
	if gossipPacket.Private.HopLimit == 0 {
		return
	}

	nextHopAddr, exist := gossiper.DSDV.Load(dest)
	if !exist {
		println("ERROR : don't know how to forward to " + dest + " but was in DSDV for address " + senderAddr)
		return
	}

	packetToSend := PacketToSend{Address: nextHopAddr.(*net.UDPAddr), GossipPacket: &gossipPacket}
	gossiper.ToSend <- packetToSend
}