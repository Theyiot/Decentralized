package gossiper

import (
	"net"
)

func (gossiper *Gossiper) receiveDataReplyPacket(gossipPacket GossipPacket, addr *net.UDPAddr) {
	fileChannel, exist := gossiper.ReceivingFile.Load(addr.String())
	if !exist {
		return
	}

	fileChannel.(chan[]byte) <- gossipPacket.DataReply.Data
}
