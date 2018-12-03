package gossiper

import (
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

func (gossiper *Gossiper) sendSearchRequest(keywords []string, budget uint64) {
	gossiper.ActiveSearches.Add(keywords)

	if budget != DEFAULT_BUDGET {
		gossiper.sendSearchPacket(budget, gossiper.Name, keywords, gossiper.Peers.GetAddresses())
		return
	}
	for budget <= DEFAULT_MAX_BUDGET{
		gossiper.sendSearchPacket(budget, gossiper.Name, keywords, gossiper.Peers.GetAddresses())
		timer := time.NewTicker(time.Second)
		select {
		case <- timer.C:
			budget *= 2
		}
	}
}

func (gossiper *Gossiper) receiveSearchRequest(gossipPacket GossipPacket, addr *net.UDPAddr) {
	request := gossipPacket.SearchRequest
	key := strings.Join(request.Keywords, ",") + "@" + request.Origin
	_, exist := gossiper.SearchRequests.LoadOrStore(key, request)
	if exist { // DUPLICATE
		return
	}

	go func() {
		timer := time.NewTicker(time.Second / 2)
		select {
		case <- timer.C:
			gossiper.SearchRequests.Delete(key)
		}
	}()

	results := make([]*SearchResult, 0)
	for _, keyword := range request.Keywords {
		gossiper.IndexedFiles.Range(func(hash, indexedFile interface{}) bool {
			chunkMap := make([]uint64, 0)
			file := indexedFile.(IndexedFile)
			metaHash := hash.([]byte)
			if strings.Contains(file.FileName, keyword) {
				for i := 0 ; i < len(file.MetaFile) / sha256.Size ; i++ {
					index := i * sha256.Size
					hashHex := hex.EncodeToString(file.MetaFile[index:index + sha256.Size])
					if _, err := os.Stat(PATH_FILE_CHUNKS + hashHex); !os.IsNotExist(err) {
						chunkMap = append(chunkMap, uint64(i + 1))
					}
				}
				result := SearchResult{FileName:file.FileName, MetafileHash:metaHash,
					ChunkCount:uint64(len(file.MetaFile) / sha256.Size), ChunkMap:chunkMap}
				results = append(results, &result)
			}
			return true
		})
	}

	searchReply := SearchReply{Origin:gossiper.Name, Destination:request.Origin, HopLimit:DEFAULT_HOP_LIMIT,
		Results:results}
	packetToSend := PacketToSend{GossipPacket:&GossipPacket{SearchReply:&searchReply}, Address:addr}
	gossiper.ToSend <- packetToSend

	//if budget is 1, then decreasing it will be 0 and we forward to 0 peer
	if request.Budget > 1 {
		gossiper.sendSearchPacket(request.Budget - 1, request.Origin, request.Keywords, gossiper.Peers.GetAddressesExcept(addr.String()))
	}
}

func (gossiper *Gossiper) sendSearchPacket(budget uint64, origin string, keywords []string, addresses []*net.UDPAddr) {
	size := len(addresses)

	var index int
	indexes := make([]int, 0)
	for i := uint64(0) ; i < min(budget, uint64(size)) ; i++ {
		for {
			newIndex := true
			index = rand.Intn(size)
			for _, elem := range indexes {
				if elem == index {
					newIndex = false
				}
			}
			if newIndex {
				indexes = append(indexes, index)
				break
			}
		}

		dividend := budget / uint64(size)
		rest := budget % uint64(size)
		//We increase the budget of some peers in order to have the right total budget
		if rest < i {
			dividend++
		}
		requestPacket := SearchRequest{Budget:dividend, Origin:origin, Keywords:keywords}
		packetToSend := PacketToSend{GossipPacket:&GossipPacket{SearchRequest:&requestPacket}, Address:addresses[index]}
		gossiper.ToSend <- packetToSend
	}
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}