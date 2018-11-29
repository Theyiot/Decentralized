package gossiper

import (
	"fmt"
	"github.com/Theyiot/Peerster/util"
	"github.com/dedis/protobuf"
	"time"
)

func (gossiper *Gossiper) printMessages() {
	for str := range gossiper.ToPrint {
		fmt.Println(str)
	}
}

func (gossiper *Gossiper) sendPacket() {
	for packet := range gossiper.ToSend {
		bytes, err := protobuf.Encode(packet.GossipPacket)
		if util.CheckAndPrintError(err) {
			continue
		}
		gossiper.GossipServer.WriteToUDP(bytes, packet.Address)
	}
}

func (gossiper *Gossiper) antiEntropy() {
	for {
		ticker := time.NewTicker(time.Second)

		select {
		case <- ticker.C:
			if gossiper.Peers.GetSize() > 0 {
				randomPeer := gossiper.Peers.ChooseRandomPeer()
				gossiper.ToSend <- PacketToSend{ Address: randomPeer, GossipPacket: &GossipPacket{ Status: gossiper.constructStatuses() }}
				ticker.Stop()
			}
		}
	}
}

func (gossiper *Gossiper) routeRumor(rtimer uint) {
	gossiper.sendRouteRumor()
	for {
		ticker := time.NewTicker(time.Duration(rtimer) * time.Second)

		select {
		case <- ticker.C:
			gossiper.sendRouteRumor()
		}
	}
}

func (gossiper *Gossiper) sendRouteRumor() {
	if !gossiper.Peers.IsEmpty() { //WE DO NOTHING WHILE WE DON'T KNOW ONE PEER AT LEAST
		id, _ := gossiper.VectorClock.LoadOrStore(gossiper.Name, uint32(1))
		gossipPacket := GossipPacket{Rumor: &RumorMessage{Text: "", Origin: gossiper.Name, ID: id.(uint32)}}

		//WE STORE THE RUMOR PACKET
		gossiper.Rumors.Store(fmt.Sprint(id.(uint32)) + "@" + gossiper.Name,
			GossipPacketTimed{GossipPacket: gossipPacket, Timestamp: time.Now()})

		//WE STORE THE RIGHT VECTOR CLOCK VALUE
		gossiper.VectorClock.Store(gossiper.Name, id.(uint32) + uint32(1))

		//SENDING THE ROUTE RUMOR
		peerAddr := gossiper.Peers.ChooseRandomPeer()
		gossiper.ToSend <- PacketToSend{GossipPacket: &gossipPacket, Address: peerAddr}
	}
}
