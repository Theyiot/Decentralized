package gossiper

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

func (gossiper *Gossiper) rumormonger(gossipPacket GossipPacket, address *net.UDPAddr) {
	gossiper.ToPrint <- "MONGERING with " + address.String()

	gossiper.ToSend <- PacketToSend{ Address: address, GossipPacket: &gossipPacket }
	key := gossipPacket.Rumor.Origin + fmt.Sprint(gossipPacket.Rumor.ID) + address.String()
	_, exist := gossiper.Acks.Load(key)
	if exist {
		return
	}
	channel := make(chan GossipPacket)
	gossiper.Acks.Store(key, channel)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	select {
	case statusPacket := <-channel:
		gossiper.Acks.Delete(key)
		if gossiper.syncPeers(statusPacket, address) {
			gossiper.flipACoin(gossipPacket, address)
		}

	case <-ticker.C:
		gossiper.Acks.Delete(key)
		gossiper.flipACoin(gossipPacket, address)
	}
}

func (gossiper *Gossiper) flipACoin(gossipPacket GossipPacket, oldAddr *net.UDPAddr){
	if rand.Int() % 2 == 0 {
		newAddr := gossiper.Peers.ChooseRandomPeerExcept(oldAddr.String())

		str := "FLIPPED COIN sending rumor to " + newAddr.String()
		gossiper.ToPrint <- str
		gossiper.rumormonger(gossipPacket, newAddr)
	}
}
