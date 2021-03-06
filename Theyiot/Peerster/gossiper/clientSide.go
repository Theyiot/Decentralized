package gossiper

import (
	"github.com/Theyiot/Peerster/util"
	"github.com/dedis/protobuf"
)

/*
	handleClient takes care of receiving and processing all the request from the client. It loops indefinitely
	and waits for UDP packets from the client
 */
func (gossiper *Gossiper) handleClient() {
	for {
		buf := make([]byte, 4096)
		n, _, err := gossiper.UIServer.ReadFromUDP(buf)
		if util.CheckAndPrintError(err) {
			continue
		}

		packet := ClientPacket{}
		err = protobuf.Decode(buf[:n], &packet)
		if util.CheckAndPrintError(err) {
			continue
		}

		go gossiper.sendClientMessage(packet)
	}
}

/*
	sendClientMessage takes care of processing the packets received from the client. It makes sure the packets
	are valid and forwards the request to the right function
 */
func (gossiper *Gossiper) sendClientMessage(packet ClientPacket) {
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
	} else if packet.FileIndex != nil {
		gossiper.indexFile(packet.FileIndex.FileName)
	} else if packet.FileRequest != nil {
		if packet.FileRequest.Destination == "" {
			gossiper.requestFile(packet.FileRequest.FileName, packet.FileRequest.Request)
		} else {
			gossiper.requestFileFrom(packet.FileRequest.FileName, packet.FileRequest.Destination, packet.FileRequest.Request)
		}
	} else if packet.FileSearchRequest != nil {
		gossiper.sendSearchRequest(packet.FileSearchRequest.Keywords, packet.FileSearchRequest.Budget)
	} else {
		println("ERROR : client did not send any know kind of packets.")
	}
}

/*
	checkExactlyOnePacketTypeClient checks that there is one and only one type of packet that is not nil
 */
func checkExactlyOnePacketTypeClient(gossipPacket ClientPacket) bool {
	count := 0
	if gossipPacket.Simple != nil { count++ }
	if gossipPacket.Private != nil { count++ }
	if gossipPacket.FileIndex != nil { count++ }
	if gossipPacket.FileRequest != nil { count++ }
	if gossipPacket.FileSearchRequest != nil { count++ }
	if count == 0 {
		println("Found 0 matching type of packet")
	}
	return count == 1
}