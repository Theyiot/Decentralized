package gossiper

import (
	"github.com/Theyiot/Peerster/util"
	"sort"
	"strconv"
	"strings"
)

const DEFAULT_HOP_LIMIT = 10
const CHUNK_SIZE = 8192
const DOWNLOAD_PATH = "_Downloads/"
const SHARED_FILES_PATH = "_SharedFiles/"
const FILE_CHUNKS_PATH = "._FileChunks/"

func (gossiper *Gossiper) GetRumorsAsList() []RumorMessageTimed {
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

func (gossiper *Gossiper) GetPeersNameAsMap() []string {
	names := make([]string, 0)
	gossiper.DSDV.Range(func(origin, address interface{}) bool {
		names = append(names, origin.(string))
		return true
	})
	sort.Strings(names)
	return names
}

func (gossiper *Gossiper) GetPrivateMessagesAsMap() map[string][]PrivateMessageTimed {
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

func (gossiper *Gossiper) GetIndexedFilesAsMap() map[string]string {
	indexedFiles := make(map[string]string, 0)
	gossiper.IndexedFiles.Range(func(metaHash, file interface{}) bool {
		indexedFiles[metaHash.(string)] = file.(IndexedFile).FileName
		return true
	})
	return indexedFiles
}

func splitKey(key string) (uint32, string, error) {
	idOrigin := strings.SplitN(key, "@", 2)
	idString, origin := idOrigin[0], idOrigin[1]
	id, err := strconv.ParseUint(idString, 10, 32)
	return uint32(id), origin, err
}

func (gossiper *Gossiper) constructStatuses() *StatusPacket {
	var statuses []PeerStatus
	gossiper.VectorClock.Range(func(origin, nextID interface{}) bool {
		statuses = append(statuses, PeerStatus{ Identifier: origin.(string), NextID: nextID.(uint32) })
		return true
	})
	return &StatusPacket{ Want: statuses }
}






