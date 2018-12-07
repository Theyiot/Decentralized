package gossiper

/*
	receiveSimplePacket handles the packets of simple type
 */
func (gossiper *Gossiper) receiveSimplePacket(gossipPacket GossipPacket) {
	origin := gossipPacket.Simple.OriginalName
	relayAddr := gossipPacket.Simple.RelayPeerAddr
	content := gossipPacket.Simple.Contents
	str := "SIMPLE MESSAGE origin " + origin + " from " + relayAddr + " contents " + content +
		gossiper.Peers.String()
	gossiper.ToPrint <- str

	gossipPacket.Simple.RelayPeerAddr = gossiper.GossipAddr

	// TRANSMITTING PACKET TO ALL KNOWN PEERS EXCEPT RELAY PEER
	gossiper.broadcastGossipPacket(gossipPacket, gossiper.Peers.GetAddressesExcept(relayAddr))
}

/*
	sendSimplePacket takes care of sending a simple message to all known peers
 */
func (gossiper *Gossiper) sendSimplePacket(content string) {
	str := "CLIENT MESSAGE " + content + gossiper.Peers.String()
	gossiper.ToPrint <- str
	simpleMessage := SimpleMessage{Contents: content, RelayPeerAddr: gossiper.GossipAddr, OriginalName: gossiper.Name}
	gossipPacket := GossipPacket{Simple: &simpleMessage}
	gossipPacket.Simple.OriginalName = gossiper.Name
	gossipPacket.Simple.RelayPeerAddr = gossiper.GossipAddr

	gossiper.broadcastGossipPacket(gossipPacket, gossiper.Peers.GetAddresses())
}