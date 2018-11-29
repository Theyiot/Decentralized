package gossiper

import (
	"github.com/Theyiot/Peerster/util"
	"github.com/dedis/protobuf"
)

func (gossiper *Gossiper) handleClient() {
	for {
		buf := make([]byte, 4096)
		n, _, err := gossiper.UIServer.ReadFromUDP(buf)
		if util.CheckAndPrintError(err) {
			continue
		}

		packet := ClientGossipPacket{}
		err = protobuf.Decode(buf[:n], &packet)
		if util.CheckAndPrintError(err) {
			continue
		}

		go gossiper.sendClientMessage(packet)
	}
}

func (gossiper *Gossiper) sendClientMessage(packet ClientGossipPacket) {
	if !checkExactlyOnePacketTypeClient(packet) {
		println("CLIENT SIDE : More than one field of the packet was not <nil>, dropping this packet")
		return
	}

	if packet.Simple != nil {
		content := packet.Simple.Contents
		if gossiper.Simple { //SIMPLE PACKET
			gossiper.sendSimplePacket(content)
		} else { //RUMOR PACKET
			gossiper.sendRumorPacket(content)
		}
	} else if packet.Private != nil { //PRIVATE PACKET
		gossiper.sendPrivatePacket(packet.Private.Text, packet.Private.Destination)
	} else if packet.File != nil {
		if packet.File.Destination == "" {
			gossiper.indexFile(packet.File.FileName)
		} else {
			gossiper.requestFile(packet.File.FileName, packet.File.Destination, packet.File.Request)
		}
	} else {
		println("ERROR : client did not send any know kind of packets.")
	}
}

func checkExactlyOnePacketTypeClient(gossipPacket ClientGossipPacket) bool {
	count := 0
	if gossipPacket.Simple != nil { count++ }
	if gossipPacket.Private != nil { count++ }
	if gossipPacket.File != nil { count++ }
	if count == 0 {
		println("Found 0 matching type of packet")
	}
	return count == 1
}