package gossiper

import (
	"github.com/Theyiot/Peerster/util"
	"github.com/dedis/protobuf"
)

func (gossiper *Gossiper) handleGossip() {
	// DataReplyPacket, Data : 8192, Hash : 32, Hop-Limit : 32, Origin : 512, dest : 512
	biggestPacketSize := CHUNK_SIZE + 1088
	buf := make([]byte, biggestPacketSize)

	for {
		n, addr, err := gossiper.GossipServer.ReadFromUDP(buf)
		if util.CheckAndPrintError(err) {
			continue
		}
		senderAddr := addr.String()
		//Peers will only be added in case it is not already in the set of peers
		gossiper.Peers.Add(senderAddr)

		gossipPacket := GossipPacket{}
		err = protobuf.Decode(buf[:n], &gossipPacket)
		if util.CheckAndPrintError(err) {
			continue
		}

		if !checkExactlyOnePacketType(gossipPacket) {
			println("GOSSIP SIDE : More than one field of the packet was not <nil>, dropping this packet")
			continue
		}

		go func() {
			if gossipPacket.Simple != nil { //SIMPLE PACKET
				gossiper.receiveSimplePacket(gossipPacket)
			} else if gossipPacket.Rumor != nil { //RUMOR PACKET
				gossiper.receiveRumorPacket(gossipPacket, addr)
			} else if gossipPacket.Status != nil { //STATUS PACKET
				gossiper.receiveStatusPacket(gossipPacket, addr)
			} else if gossipPacket.Private != nil { //PRIVATE PAQUET
				gossiper.receivePrivatePacket(gossipPacket, senderAddr)
			} else if gossipPacket.DataRequest != nil { //DATA REQUEST PACKET
				gossiper.receiveDataRequestPacket(gossipPacket, addr)
			} else if gossipPacket.DataReply != nil { //DATA REPLY PACKET
				gossiper.receiveDataReplyPacket(gossipPacket, addr)
			} else if gossipPacket.SearchRequest != nil { //DATA REPLY PACKET
				gossiper.receiveSearchRequest(gossipPacket, addr)
			} else if gossipPacket.SearchReply != nil { //DATA REPLY PACKET

			} else {
				println("Received packet type that should not be sent to other peer")
			}
		}()
	}
}

func checkExactlyOnePacketType(gossipPacket GossipPacket) bool {
	count := 0
	if gossipPacket.Simple != nil { count++ }
	if gossipPacket.Rumor != nil { count++ }
	if gossipPacket.Status != nil { count++ }
	if gossipPacket.Private != nil { count++ }
	if gossipPacket.DataRequest != nil { count++ }
	if gossipPacket.DataReply != nil { count++ }
	if gossipPacket.SearchRequest != nil { count++ }
	if gossipPacket.SearchReply != nil { count++ }
	if count == 0 {
		println("Found 0 matching type of packet")
	}
	return count == 1
}
