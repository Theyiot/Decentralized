package gossiper

import (
	"fmt"
	"net"
)

func (gossiper *Gossiper) receiveStatusPacket(statusPacket GossipPacket, addr *net.UDPAddr) {
	str := "STATUS from " + addr.String()
	for _, status := range statusPacket.Status.Want {
		identifier, statusNextID := status.Identifier, status.NextID

		if status.NextID > 1 {
			str += " peer " + identifier + " nextID " + fmt.Sprint(statusNextID)

			if channel, exist := gossiper.Acks.Load(identifier + fmt.Sprint(statusNextID - 1) + addr.String()); exist {
				select { // NON-BLOCKING SEND
				case channel.(chan GossipPacket) <- statusPacket:
				default:
				}
				return
			}
		}
	}

	if gossiper.syncPeers(statusPacket, addr) {
		str += gossiper.Peers.String() + "\nIN SYNC WITH " + addr.String()
		gossiper.ToPrint <- str
	}
}