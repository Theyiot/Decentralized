package gossiper

import (
	"encoding/hex"
	"github.com/Theyiot/Peerster/constants"
	"net"
)

func (gossiper *Gossiper) receiveTxPublish(gossipPacket GossipPacket, addr *net.UDPAddr) {
	_, exist := gossiper.NameToMetaHash.Load(gossipPacket.TxPublish.File.Name)
	if exist {
		return
	}
	gossiper.Transactions.Add(gossipPacket.TxPublish)
}

func (gossiper *Gossiper) forwardTxPublish(gossipPacket GossipPacket, peerAddr string) {
	//WE DECREASE AND DISCARD INVALID PACKET
	gossipPacket.TxPublish.HopLimit--
	if gossipPacket.TxPublish.HopLimit == 0 {
		return
	}

	gossiper.broadcastGossipPacket(gossipPacket, gossiper.Peers.GetAddressesExcept(peerAddr))
}

func (gossiper *Gossiper) receiveBlockPublish(gossipPacket GossipPacket, addr *net.UDPAddr) {
	block := gossipPacket.BlockPublish.Block
	newHash, prevHashHex := block.Hash(), hex.EncodeToString(block.PrevHash[:])
	_, exist := gossiper.Blockchain.Load(prevHashHex)

	// CHECK IF THE INITIAL ARRAY IS STILL MINED
	if !exist || newHash[0] != 0 || newHash[1] != 0 {
		return
	}

	gossiper.ToAddToBlockchain <- block
	gossiper.forwardBlockPublish(gossipPacket, addr.String())
}

func (gossiper *Gossiper) forwardBlockPublish(gossipPacket GossipPacket, peerAddr string) {
	//WE DECREASE AND DISCARD INVALID PACKET
	gossipPacket.BlockPublish.HopLimit--
	if gossipPacket.BlockPublish.HopLimit == 0 {
		return
	}

	gossiper.broadcastGossipPacket(gossipPacket, gossiper.Peers.GetAddressesExcept(peerAddr))
}

func (gossiper *Gossiper) checkNameInLongestChain(newTransaction TxPublish) bool {
	lastBlock := false
	blockHashHex := gossiper.CurrentBlock.GetCurrentHash()

	for {
		block, _ := gossiper.Blockchain.Load(blockHashHex)
		blockHashHex = hex.EncodeToString(block.(Block).PrevHash[:])
		for _, transaction := range block.(Block).Transactions {
			if transaction.File.Name == newTransaction.File.Name {
				return true
			}
		}
		if blockHashHex == constants.DEFAULT_BLOCK_HASH {
			if lastBlock {
				return false
			}
			lastBlock = true
		}
	}
}

func (gossiper *Gossiper) computeDepth(prevHashHex string) uint64 {
	counter := uint64(1)
	for {
		block, _ := gossiper.Blockchain.Load(prevHashHex)
		if hex.EncodeToString(block.(Block).PrevHash[:]) == constants.DEFAULT_BLOCK_HASH {
			return counter
		}
		prevHashHex = hex.EncodeToString(block.(Block).PrevHash[:])
	}
}
