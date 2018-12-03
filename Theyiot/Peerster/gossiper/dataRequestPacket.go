package gossiper

import (
	"encoding/hex"
	"errors"
	"github.com/Theyiot/Peerster/util"
	"net"
	"os"
	"time"
)

func (gossiper *Gossiper) receiveDataRequestPacket(gossipPacket GossipPacket, addr *net.UDPAddr) {
	dest, hash := gossipPacket.DataRequest.Destination, gossipPacket.DataRequest.HashValue
	hashHex := hex.EncodeToString(hash)

	if !IsHexHash(hashHex) {
		println("ERROR : trying to request a data chunk from something that is not a hash : " + string(hash))
		return
	}

	//FILE REQUEST IS NOT DESTINED TO US
	if dest != gossiper.Name {
		gossiper.forwardDataRequestPacket(gossipPacket, addr.String())
	}

	//REQUESTED HASH CORRESPONDS TO A METAFILE
	if _, err := os.Stat(PATH_FILE_CHUNKS + hashHex); os.IsNotExist(err) {
		println("ERROR : cannot find file for hash : " + hashHex)
		return
	}
	data, err := readFile(hashHex)
	if util.CheckAndPrintError(err) {
		return
	}
	dataReply := DataReply{HashValue: hash, HopLimit: DEFAULT_HOP_LIMIT, Destination: dest,
		Origin: gossiper.Name, Data: data}
	gossiper.ToSend <- PacketToSend{GossipPacket: &GossipPacket{DataReply: &dataReply}, Address: addr}
}

func (gossiper *Gossiper) forwardDataRequestPacket(gossipPacket GossipPacket, senderAddr string) {
	//WE DECREASE AND DISCARD INVALID PACKET
	gossipPacket.DataRequest.HopLimit--
	if gossipPacket.DataRequest.HopLimit == 0 {
		return
	}

	nextHopAddr, exist := gossiper.DSDV.Load(gossipPacket.DataRequest.Destination)
	if !exist {
		println("ERROR : don't know how to forward to " + gossipPacket.DataRequest.Destination)
		return
	}

	packetToSend := PacketToSend{Address: nextHopAddr.(*net.UDPAddr), GossipPacket: &gossipPacket}
	gossiper.ToSend <- packetToSend
}

func (gossiper *Gossiper) sendDataRequest(hash []byte, dest string, addr *net.UDPAddr, fileChannel chan[]byte) ([]byte, error) {
	dataRequest := DataRequest{HashValue: hash, HopLimit: DEFAULT_HOP_LIMIT, Destination: dest,
		Origin: gossiper.Name}
	gossiper.ToSend <- PacketToSend{GossipPacket: &GossipPacket{DataRequest: &dataRequest}, Address: addr}

	// WE TRY TO SEND THE DATA REQUEST 10 TIMES, IT PROBABLY WON'T SUCCESS IF IT FAILED THIS NUMBER OF TIME
	i, numberOfTries := 0, 10
	for i < numberOfTries {
		timer := time.NewTimer(5 * time.Second)

		select {
		case fileChunk := <- fileChannel:
			util.CheckAndPrintError(writeFile(hex.EncodeToString(hash), fileChunk))
			return fileChunk, nil

		case <- timer.C:
			gossiper.ToSend <- PacketToSend{GossipPacket: &GossipPacket{DataRequest: &dataRequest}, Address: addr}
		}
		i++
	}
	return []byte{}, errors.New("Peer " + dest + " did not answer for request " + hex.EncodeToString(hash) + " too much time")
}