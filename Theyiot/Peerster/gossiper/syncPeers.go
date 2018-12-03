package gossiper

import (
	"fmt"
	"net"
)

/*
	syncPeers takes care of syncing two peers (ourselves and the peer with addr). It first synchronize
	the peer with ourselves, then ourselves and him, after which it returns that we are synced. The method
	does not calls itself recursively explicitly, but calls two method that starts again the process
	of rumormongering, which will in the end call this method again, until the two peers are synced
 */
func (gossiper *Gossiper) syncPeers(statusPacket GossipPacket, addr *net.UDPAddr) bool {
	if gossiper.syncPeerWithMe(statusPacket, addr) {
		if gossiper.syncMeWithPeer(statusPacket, addr) {
			return true
		}
	}
	return false
}

/*
	syncPeerWithMe checks that we do not own a rumor that the other peer does not know. If we own one
	such packet, then we send it to the other peer by rumormongering with him
 */
func (gossiper *Gossiper) syncPeerWithMe(statusPacket GossipPacket, peerAddr *net.UDPAddr) bool {
	//LOOK FOR NOT UP-TO-DATE PACKETS
	for _, status := range statusPacket.Status.Want {
		identifier, statusNextID := status.Identifier, status.NextID
		myNextID, peerKnown := gossiper.VectorClock.Load(identifier)
		if peerKnown && statusNextID < myNextID.(uint32) {
			key := fmt.Sprint(statusNextID) + "@" + identifier
			rumor, exist := gossiper.Rumors.Load(key)
			if !exist {
				println("ERROR : Tried to access " + key + " which should be known but is not")
				return false
			}

			gossiper.rumormonger(rumor.(GossipPacketTimed).GossipPacket, peerAddr)
			return false
		}
	}
	//LOOK FOR UNKNOWN PACKETS
	upToDate := true
	gossiper.VectorClock.Range(func(identifier, _ interface{}) bool {
		isKnown := false
		for _, status := range statusPacket.Status.Want {
			if status.Identifier == identifier.(string) {
				isKnown = true
			}
		}
		if !isKnown {
			upToDate = false
			key := "1@" + identifier.(string)
			rumor, exist := gossiper.Rumors.Load(key)
			if !exist {
				println("ERROR : Tried to access " + key + " which should be known but is not")
				return true
			}

			gossiper.rumormonger(rumor.(GossipPacketTimed).GossipPacket, peerAddr)
			return false
		}
		return true
	})

	return upToDate
}

/*
	syncMeWithPeer checks that we own all the rumors that the other peer knows. If there is one
	rumor we don't know, we send him our status so that the other peer know it has to send it
	this packet
 */
func (gossiper *Gossiper) syncMeWithPeer(statusPacket GossipPacket, peerAddr *net.UDPAddr) bool {
	for _, status := range statusPacket.Status.Want {
		statusNextID := status.NextID
		myNextID, peerKnown := gossiper.VectorClock.Load(status.Identifier)
		if !peerKnown || statusNextID > myNextID.(uint32) {
			gossiper.ToSend <- PacketToSend{ Address: peerAddr, GossipPacket: &GossipPacket{ Status: gossiper.constructStatuses() } }
			return false
		}
	}

	return true
}
