package gossiper

import "strings"

func (gossiper *Gossiper) receiveSimplePacket(gossipPacket GossipPacket) {
	origin := gossipPacket.Simple.OriginalName
	relayAddr := gossipPacket.Simple.RelayPeerAddr
	content := gossipPacket.Simple.Contents
	str := "SIMPLE MESSAGE origin " + origin + " from " + relayAddr + " contents " + content +
		gossiper.Peers.String()
	gossiper.ToPrint <- str

	gossipPacket.Simple.RelayPeerAddr = gossiper.GossipAddr

	// TRANSMITTING PACKET TO ALL KNOWN PEERS EXCEPT RELAY PEER
	for _, address := range gossiper.Peers.GetAddresses() {
		if strings.EqualFold(address.String(), relayAddr) {
			continue
		}

		gossiper.ToSend <- PacketToSend{Address: address, GossipPacket: &gossipPacket}
	}
}

func (gossiper *Gossiper) sendSimplePacket(content string) {
	str := "CLIENT MESSAGE " + content + gossiper.Peers.String()
	gossiper.ToPrint <- str
	simpleMsg := SimpleMessage{Contents: content, RelayPeerAddr: gossiper.GossipAddr, OriginalName: gossiper.Name}
	simplePacket := GossipPacket{Simple: &simpleMsg}
	simplePacket.Simple.OriginalName = gossiper.Name
	simplePacket.Simple.RelayPeerAddr = gossiper.GossipAddr

	for _, address := range gossiper.Peers.GetAddresses() {
		gossiper.ToSend <- PacketToSend{Address: address, GossipPacket: &simplePacket}
	}
}