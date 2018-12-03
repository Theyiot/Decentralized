package gossiper

import (
	"encoding/hex"
	"fmt"
	"github.com/Theyiot/Peerster/util"
	"net"
	"strings"
)

func (gossiper *Gossiper) receiveSearchReply(gossipPacket GossipPacket) {
	if gossipPacket.SearchReply.Destination != gossiper.Name {
		gossiper.forwardSearchReplyPacket(gossipPacket)
		return
	}

	for _, result := range gossipPacket.SearchReply.Results {
		peerName, hashHex := gossipPacket.SearchReply.Origin, hex.EncodeToString(result.MetafileHash)
		str := "FOUND match " + result.FileName + " at " + peerName+ " metafile=" + hashHex + " chunks="
		for i := 0 ; i < len(result.ChunkMap) ; i++ {
			//gossiper.SearchedFiles.LoadOrStore()
			str += fmt.Sprint(result.ChunkMap[i])
			if i < len(result.ChunkMap) - 1 {
				str += ","
			}
		}
		gossiper.ToPrint <- str



		searchFileChunksNotCasted, _ := gossiper.SearchedFiles.LoadOrStore(result.MetafileHash, make([]SearchedFileChunk, 0))
		searchFileChunks := searchFileChunksNotCasted.([]SearchedFileChunk)
		for _, id := range result.ChunkMap {
			exist := false
			for _, searchFileChunk := range searchFileChunks {
				if id == searchFileChunk.ChunkID {
					searchFileChunk.lock.Lock()
					searchFileChunk.owningPeers = append(searchFileChunk.owningPeers, peerName)
					searchFileChunk.lock.Unlock()
					exist = true
				}
			}
			if !exist {
				searchFileChunk := SearchedFileChunk{owningPeers:[]string {peerName}, ChunkID:id,
					ChunkCount:result.ChunkCount, FileName:result.FileName}
				searchFileChunks = append(searchFileChunks, searchFileChunk)
				if uint64(len(searchFileChunks)) == result.ChunkCount {
					gossiper.findFullMatches(result.FileName)
				}
			}
		}
	}
}

func (gossiper *Gossiper) forwardSearchReplyPacket(gossipPacket GossipPacket) {
	//WE DECREASE AND DISCARD INVALID PACKET
	gossipPacket.DataRequest.HopLimit--
	if gossipPacket.DataRequest.HopLimit == 0 {
		return
	}

	nextHopAddr, exist := gossiper.DSDV.Load(gossipPacket.DataReply.Destination)
	if !exist {
		println("ERROR : don't know how to forward to " + gossipPacket.DataReply.Destination)
		return
	}

	packetToSend := PacketToSend{Address: nextHopAddr.(*net.UDPAddr), GossipPacket: &gossipPacket}
	gossiper.ToSend <- packetToSend
}

func (gossiper *Gossiper) findFullMatches(fileName string) {
	keywords := gossiper.ActiveSearches.GetSetCopy()
	size := gossiper.ActiveSearches.Size()
	for i := size - 1 ; i >= 0 ; i-- {
		for _, keyword := range keywords[i] {
			if strings.Contains(fileName, keyword) {
				fullMatches, err := gossiper.ActiveSearches.IncrementFullMatchIndex(i)
				if util.CheckAndPrintError(err){
					return
				} else if fullMatches == DEFAULT_FULL_MATCHES {
					gossiper.ToPrint <- "SEARCH FINISHED"
					gossiper.ActiveSearches.Remove(i)
				}
			}
		}
	}
}