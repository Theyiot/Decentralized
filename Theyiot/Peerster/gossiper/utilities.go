package gossiper

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/Theyiot/Peerster/constants"
	"github.com/Theyiot/Peerster/util"
	"math/rand"
	"net"
	"sort"
	"strconv"
	"strings"
)

/*
	getRumorAsList returns a list of all the non-empty rumors we have received so far
 */
func (gossiper *Gossiper) getRumorsAsList() []RumorMessageTimed {
	rumors := make([]RumorMessageTimed, 0)
	gossiper.Rumors.Range(func(key, rumorPacket interface{}) bool {
		id, origin, err := splitKey(key.(string))
		packet := rumorPacket.(GossipPacketTimed)
		msg, timestamp := packet.GossipPacket.Rumor.Text, packet.Timestamp
		if util.CheckAndPrintError(err) {
			return true
		}
		if msg != "" {
			rumor := RumorMessage{ Text: msg, ID: id, Origin: origin }
			rumors = append(rumors, RumorMessageTimed{ Rumor: rumor, Timestamp: timestamp })
		}
		return true
	})
	sort.SliceStable(rumors, func(i int, j int) bool {
		return rumors[i].Timestamp.Before(rumors[j].Timestamp)
	})
	return rumors
}

/*
	getPeersNameAsList returns a list of all the known peers name
 */
func (gossiper *Gossiper) getPeersNameAsList() []string {
	names := make([]string, 0)
	gossiper.DSDV.Range(func(origin, address interface{}) bool {
		names = append(names, origin.(string))
		return true
	})
	sort.Strings(names)
	return names
}

/*
	getPrivateMessagesAsMap returns a map of the form origin -> []PrivateMessageTimed for all our known peers
 */
func (gossiper *Gossiper) getPrivateMessagesAsMap() map[string][]PrivateMessageTimed {
	privates := make(map[string][]PrivateMessageTimed, 0)
	gossiper.DSDV.Range(func(origin, _ interface{}) bool {
		privates[origin.(string)] = make([]PrivateMessageTimed, 0)
		return true
	})
	gossiper.Privates.Range(func(origin, packets interface{}) bool {
		msgListForOrigin := privates[origin.(string)]
		for _, packet := range packets.([]GossipPacketTimed) {
			msgListForOrigin = append(msgListForOrigin, PrivateMessageTimed{Private: *packet.GossipPacket.Private,
				Timestamp: packet.Timestamp})
		}
		privates[origin.(string)] = msgListForOrigin
		return true
	})
	return privates
}

/*
	getIndexedFilesAsMap returns a map of the form metaHash -> fileName for all our indexed files
 */
func (gossiper *Gossiper) getIndexedFilesAsMap() map[string]string {
	indexedFiles := make(map[string]string, 0)
	gossiper.IndexedFiles.Range(func(metaHash, file interface{}) bool {
		indexedFiles[metaHash.(string)] = file.(IndexedFile).FileName
		return true
	})
	return indexedFiles
}

/*
	splitKey tries to split a key that have the form id@origin. It returns an error if the splitting process
	fails at some point
 */
func splitKey(key string) (uint32, string, error) {
	idOrigin := strings.SplitN(key, "@", 2)
	idString, origin := idOrigin[0], idOrigin[1]
	id, err := strconv.ParseUint(idString, 10, 32)
	return uint32(id), origin, err
}

/*
	constructStatuses iterate through our vector clock and create a status packet from this information
 */
func (gossiper *Gossiper) constructStatuses() *StatusPacket {
	var statuses []PeerStatus
	gossiper.VectorClock.Range(func(origin, nextID interface{}) bool {
		statuses = append(statuses, PeerStatus{ Identifier: origin.(string), NextID: nextID.(uint32) })
		return true
	})
	return &StatusPacket{ Want: statuses }
}

func (gossiper *Gossiper) broadcastGossipPacket(gossipPacket GossipPacket, addresses []*net.UDPAddr) {
	for _, peer := range addresses {
		gossiper.ToSend <- PacketToSend{ GossipPacket:&gossipPacket, Address:peer }
	}
}

func (gossiper *Gossiper) mineBlock(prevHash [32]byte) {
	newBlock := Block{ PrevHash:prevHash }
	var hash [sha256.Size]byte

	for {
		select {
		case <- gossiper.BlockMined:
			return
		default:
			newBlock.Nonce = generateRandomNounce()
			hash = newBlock.Hash()
			if hash[0] == 0 && hash[1] == 0 {
				break
			}
		}
	}

	newHashHex := hex.EncodeToString(hash[:])
	gossiper.ToPrint <- "FOUND-BLOCK " + newHashHex
	gossipPacket := GossipPacket{ BlockPublish:&BlockPublish{ Block:newBlock, HopLimit:constants.HOP_LIMIT_BIG } }
	gossiper.ToAddToBlockchain <- newBlock
	gossiper.broadcastGossipPacket(gossipPacket, gossiper.Peers.GetAddresses())

	gossiper.mineBlock(hash)
}

func generateRandomNounce() (nounce [32]byte) {
	random := make([]byte, constants.NOUNCE_SIZE)
	rand.Read(random)
	copy(nounce[:], random)
	return
}




