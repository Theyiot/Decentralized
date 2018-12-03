package gossiper

import (
	"encoding/hex"
	"net"
)

/*
	receiveDataReplyPacket handles the packets of DataReply type.
 */
func (gossiper *Gossiper) receiveDataReplyPacket(gossipPacket GossipPacket, addr *net.UDPAddr) {
	fileChannel, exist := gossiper.ReceivingFile.Load(hex.EncodeToString(gossipPacket.DataReply.HashValue))
	if !exist {
		return
	}

	fileChannel.(chan[]byte) <- gossipPacket.DataReply.Data
}
