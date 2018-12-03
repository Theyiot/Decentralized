package gossiper

import (
	"encoding/hex"
	"fmt"
	"github.com/Theyiot/Peerster/constants"
	"github.com/Theyiot/Peerster/util"
	"net"
	"strings"
)

/*
	receiveSearchReply handles the packets of search reply type
 */
func (gossiper *Gossiper) receiveSearchReply(gossipPacket GossipPacket, addr *net.UDPAddr) {
	_, exist := gossiper.DSDV.LoadOrStore(gossipPacket.SearchReply.Origin, addr)
	if !exist {
		gossiper.ToPrint <- "DSDV " + gossipPacket.SearchReply.Origin + " " + addr.String()
	}

	if gossipPacket.SearchReply.Destination != gossiper.Name {
		gossiper.forwardSearchReplyPacket(gossipPacket)
		return
	}

	for _, result := range gossipPacket.SearchReply.Results {
		peerName, hashHex := gossipPacket.SearchReply.Origin, hex.EncodeToString(result.MetafileHash)
		str := "FOUND match " + result.FileName + " at " + peerName+ " metafile=" + hashHex + " chunks="
		for i := 0 ; i < len(result.ChunkMap) ; i++ {
			str += fmt.Sprint(result.ChunkMap[i])
			if i < len(result.ChunkMap) - 1 {
				str += ","
			}
		}
		gossiper.ToPrint <- str

		searchFileChunksNotCasted, _ := gossiper.SearchedFiles.LoadOrStore(hashHex, make([]SearchedFileChunk, 0))
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
			gossiper.SearchedFiles.Store(hashHex, searchFileChunks)
		}
	}
}

/*
	forwardSearchReplyPacket takes care of forwarding a point-to-point search reply message to the right peer
 */
func (gossiper *Gossiper) forwardSearchReplyPacket(gossipPacket GossipPacket) {
	//WE DECREASE AND DISCARD INVALID PACKET
	gossipPacket.SearchReply.HopLimit--
	if gossipPacket.SearchReply.HopLimit == 0 {
		return
	}

	nextHopAddr, exist := gossiper.DSDV.Load(gossipPacket.SearchReply.Destination)
	if !exist {
		println("ERROR : don't know how to forward to " + gossipPacket.SearchReply.Destination)
		return
	}

	packetToSend := PacketToSend{Address: nextHopAddr.(*net.UDPAddr), GossipPacket: &gossipPacket}
	gossiper.ToSend <- packetToSend
}

/*
	findFullMatches is called whenever the file with the given name is complete. The method takes care
	to increase the fullMatches counter of all the searches that have a keyword that is a substring of
	the filename
 */
func (gossiper *Gossiper) findFullMatches(fileName string) {
	keywords := gossiper.ActiveSearches.GetSetCopy()
	size := gossiper.ActiveSearches.Size()
	for i := size - 1 ; i >= 0 ; i-- {
		for _, keyword := range keywords[i] {
			if strings.Contains(fileName, keyword) {
				fullMatches, err := gossiper.ActiveSearches.IncrementFullMatchIndex(i)
				if util.CheckAndPrintError(err) {
					return
				} else if fullMatches == constants.DEFAULT_FULL_MATCHES {
					gossiper.ToPrint <- "SEARCH FINISHED"
					gossiper.ActiveSearches.Remove(i)
					channel, exist := gossiper.FinishedSearches.Load(strings.Join(keywords[i], ","))
					if exist {
						channel.(chan Signal) <- Signal{}
					}
				}
			}
		}
	}
}