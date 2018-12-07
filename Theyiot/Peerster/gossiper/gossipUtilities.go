package gossiper

import (
	"encoding/hex"
	"fmt"
	"github.com/Theyiot/Peerster/util"
	"github.com/dedis/protobuf"
	"strings"
	"sync"
	"time"
)

/*
	printMessages continuously waits for message to print in the terminal
 */
func (gossiper *Gossiper) printMessages() {
	for str := range gossiper.ToPrint {
		fmt.Println(str)
	}
}

/*
	sendPacket continuously waits for packets to send to other peers
 */
func (gossiper *Gossiper) sendPacket() {
	for packet := range gossiper.ToSend {
		bytes, err := protobuf.Encode(packet.GossipPacket)
		if util.CheckAndPrintError(err) {
			continue
		}
		gossiper.GossipServer.WriteToUDP(bytes, packet.Address)
	}
}

/*
	antiEntropy sends a status packet to a random peer every second, in order to make sure that the entire
	network is up-to-date
 */
func (gossiper *Gossiper) antiEntropy() {
	for {
		ticker := time.NewTicker(time.Second)

		select {
		case <- ticker.C:
			if !gossiper.Peers.IsEmpty() {
				randomPeer := gossiper.Peers.ChooseRandomPeer()
				gossiper.ToSend <- PacketToSend{ Address: randomPeer, GossipPacket: &GossipPacket{ Status: gossiper.constructStatuses() }}
				ticker.Stop()
			}
		}
	}
}

/*
	routeRumor continuously sends empty rumor to a random peer in order to make sure that we know
	every other peer in the network. The frequence at which we send these messages is choosable by
	the user and by default is disabled
 */
func (gossiper *Gossiper) routeRumor(rtimer uint) {
	gossiper.sendRouteRumor()
	for {
		ticker := time.NewTicker(time.Duration(rtimer) * time.Second)

		select {
		case <- ticker.C:
			gossiper.sendRouteRumor()
		}
	}
}

/*
	sendRouteRumor takes care of sending an empty rumor to a random peer
 */
func (gossiper *Gossiper) sendRouteRumor() {
	if !gossiper.Peers.IsEmpty() { //WE DO NOTHING WHILE WE DON'T KNOW ONE PEER AT LEAST
		id, _ := gossiper.VectorClock.LoadOrStore(gossiper.Name, uint32(1))
		gossipPacket := GossipPacket{Rumor: &RumorMessage{Text: "", Origin: gossiper.Name, ID: id.(uint32)}}

		//WE STORE THE RUMOR PACKET
		gossiper.Rumors.Store(fmt.Sprint(id.(uint32)) + "@" + gossiper.Name,
			GossipPacketTimed{GossipPacket: gossipPacket, Timestamp: time.Now()})

		//WE STORE THE RIGHT VECTOR CLOCK VALUE
		gossiper.VectorClock.Store(gossiper.Name, id.(uint32) + uint32(1))

		//SENDING THE ROUTE RUMOR
		peerAddr := gossiper.Peers.ChooseRandomPeer()
		gossiper.ToSend <- PacketToSend{GossipPacket: &gossipPacket, Address: peerAddr}
	}
}

/*
	sendPacket continuously waits for packets to send to other peers
 */
func (gossiper *Gossiper) addBlockToBlockchain() {
	for block := range gossiper.ToAddToBlockchain {
		newHash := block.Hash()
		newHashHex := hex.EncodeToString(newHash[:])
		prevHashHex := hex.EncodeToString(block.PrevHash[:])
		//CHECKING IF WE ARE ON LONGEST CHAIN
		if prevHashHex == gossiper.CurrentBlock.GetCurrentHash() {
			str := gossiper.printChain(block)
			gossiper.ToPrint <- str
		}
		gossiper.Blockchain.Store(newHashHex, block)

		if gossiper.CurrentBlock.GetCurrentHash() == prevHashHex {
			gossiper.CurrentBlock.IncrementDepth()
			gossiper.CurrentBlock.SetCurrentHash(newHashHex)
			for _, transaction := range block.Transactions {
				gossiper.NameToMetaHash.Store(transaction.File.Name, transaction.File.MetafileHash)
			}
			gossiper.BlockMined <- Signal{}
			gossiper.mineBlock(newHash)
		} else {
			newDepth := gossiper.computeDepth(newHashHex)
			if newDepth > gossiper.CurrentBlock.GetDepth() {
				gossiper.ToPrint <- "FORK-LONGER rewind " + " blocks"
				gossiper.switchBranch(prevHashHex)
				gossiper.CurrentBlock.SetDepth(newDepth)
			} else {
				gossiper.ToPrint <- "FORK-SHORTER " + newHashHex
			}
		}
	}
}

func (gossiper *Gossiper) switchBranch(prevHashHex string) {
	gossiper.NameToMetaHash = sync.Map{}
	hashHex := prevHashHex
	for {
		block, exist := gossiper.Blockchain.Load(hashHex)
		if !exist {
			return
		}
		for _, transaction := range block.(Block).Transactions {
			gossiper.NameToMetaHash.Store(transaction.File.Name, transaction.File.MetafileHash)
		}
		gossiper.Transactions.flushFromBlock(block.(Block))
		hashHex = hex.EncodeToString(block.(Block).PrevHash[:])
	}
}

func (gossiper *Gossiper) printChain(block Block) string {
	str := "CHAIN"
	for {
		str += " " + printBlock(block)
		newBlock, exist := gossiper.Blockchain.Load(hex.EncodeToString(block.PrevHash[:]))
		if !exist {
			break
		}
		block = newBlock.(Block)
	}
	return str
}

func printBlock(block Block) string {
	transactions := make([]string, 0)
	for _, transaction := range block.Transactions {
		transactions = append(transactions, transaction.File.Name)
	}
	hashHex := hex.EncodeToString(block.Hash()[:])
	prevHashHex := hex.EncodeToString(block.PrevHash[:])
	return "[" + hashHex + ":" + prevHashHex + ":" + strings.Join(transactions, ",")
}
