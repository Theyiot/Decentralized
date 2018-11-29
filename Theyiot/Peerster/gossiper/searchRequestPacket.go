package gossiper

import (
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"net"
	"os"
	"strings"
)

func (gossiper *Gossiper) receiveSearchRequest(gossipPacket GossipPacket, addr *net.UDPAddr) {
	results := make([]*SearchResult, 0)
	for _, keyword := range gossipPacket.SearchRequest.Keywords {
		gossiper.IndexedFiles.Range(func(hash, indexedFile interface{}) bool {
			chunkMap := make([]uint64, 0)
			file := indexedFile.(IndexedFile)
			metaHash := hash.([]byte)
			if strings.Contains(file.FileName, keyword) {
				for i := 0 ; i < len(file.MetaFile) / sha256.Size ; i++ {
					index := i * sha256.Size
					hashHex := hex.EncodeToString(file.MetaFile[index:index + sha256.Size])
					if _, err := os.Stat(FILE_CHUNKS_PATH + hashHex); !os.IsNotExist(err) {
						chunkMap = append(chunkMap, uint64(i))
					}
				}
				result := SearchResult{FileName:file.FileName, MetafileHash:metaHash,
					ChunkCount:uint64(len(file.MetaFile) / sha256.Size), ChunkMap:chunkMap}
				results = append(results, &result)
			}
			return true
		})
	}

	searchReply := SearchReply{Origin:gossiper.Name, Destination:gossipPacket.SearchRequest.Origin, HopLimit:DEFAULT_HOP_LIMIT,
		Results:results}
	packetToSend := PacketToSend{GossipPacket:&GossipPacket{SearchReply:&searchReply}, Address:addr}
	gossiper.ToSend <- packetToSend

	//if budget is 1, then decreasing it will be 0 and we forward to 0 peer
	if gossipPacket.SearchRequest.Budget > 1 {
		gossiper.forwardSearchPacket(gossipPacket, addr)
	}
}

func (gossiper *Gossiper) forwardSearchPacket(gossipPacket GossipPacket, addr *net.UDPAddr) {
	request := gossipPacket.SearchRequest
	budget := request.Budget - 1
	//TODO : Modify that
	//We don't want to send back to origin
	addresses := gossiper.Peers.GetAddressesExcept(addr.String())
	size := len(addresses)

	var index int
	indexes := make([]int, 0)
	for i := 0 ; i < min(int(budget), size) ; i++ {
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
		rest := int(budget) % size
		//We increase the budget of some peers in order to have the right total budget
		if rest < i {
			dividend++
		}
		requestPacket := SearchRequest{Budget:dividend, Origin:request.Origin, Keywords:request.Keywords}
		packetToSend := PacketToSend{GossipPacket:&GossipPacket{SearchRequest:&requestPacket}, Address:addresses[index]}
		gossiper.ToSend <- packetToSend
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}