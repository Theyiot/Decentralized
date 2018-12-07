package gossiper

import (
	"crypto/sha256"
	"github.com/Theyiot/Peerster/constants"
	"github.com/Theyiot/Peerster/util"
	"net"
	"sync"
	"time"
)

// MESSAGES
type SimpleMessage struct {
	OriginalName  string
	RelayPeerAddr string
	Contents      string
}

type RumorMessage struct {
	Origin	string
	ID		uint32
	Text	string
}

type PrivateMessage struct {
	Origin      string
	ID          uint32
	Text        string
	Destination string
	HopLimit    uint32
}

type FileIndexMessage struct {
	FileName	string
}

type FileRequestMessage struct {
	FileName    string
	Destination string
	Request     string
}

type SearchRequestMessage struct {
	Keywords 	[]string
	Budget		uint64
}

// PACKETS
type StatusPacket struct {
	Want	[]PeerStatus
}

type PeerStatus struct {
	Identifier	string
	NextID		uint32
}

type DataRequest struct {
	Origin			string
	Destination		string
	HopLimit		uint32
	HashValue		[]byte
}

type DataReply struct {
	Origin			string
	Destination		string
	HopLimit		uint32
	HashValue		[]byte
	Data			[]byte
}

type SearchRequest struct {
	Origin		string
	Budget		uint64
	Keywords 	[]string
}

type SearchReply struct {
	Origin 			string
	Destination 	string
	HopLimit 		uint32
	Results			[]*SearchResult
}

type SearchResult struct {
	FileName		string
	MetafileHash	[]byte
	ChunkMap		[]uint64
	ChunkCount		uint64
}

//TIMED PACKETS
type RumorMessageTimed struct {
	Rumor		RumorMessage
	Timestamp	time.Time
}

type PrivateMessageTimed struct {
	Private			PrivateMessage
	Timestamp		time.Time
}

type GossipPacketTimed struct {
	GossipPacket	GossipPacket
	Timestamp		time.Time
}

//CLIENT AND GOSSIP PACKET
type GossipPacket struct {
	Simple			*SimpleMessage
	Rumor			*RumorMessage
	Status			*StatusPacket
	Private			*PrivateMessage
	DataRequest		*DataRequest
	DataReply		*DataReply
	SearchRequest	*SearchRequest
	SearchReply		*SearchReply
	TxPublish		*TxPublish
	BlockPublish	*BlockPublish
}

type ClientPacket struct {
	Simple            *SimpleMessage
	Private           *PrivateMessage
	FileRequest       *FileRequestMessage
	FileIndex         *FileIndexMessage
	FileSearchRequest *SearchRequestMessage
}

//FILES
type FileToIndex struct {
	FileName		string
}

type IndexedFile struct {
	FileName		string
	FileSize		int64
	MetaFile		[]byte
}

type SearchedFileChunk struct {
	FileName    string
	ChunkID     uint64
	ChunkCount  uint64
	owningPeers []string
	lock        sync.RWMutex
}

//BLOCKCHAIN
type TxPublish struct {
	File		File
	HopLimit	uint32
}

type File struct {
	Name			string
	Size			int64
	MetafileHash	[]byte
}

type BlockPublish struct {
	Block		Block
	HopLimit	uint32
}

type Block struct {
	PrevHash		[sha256.Size]byte
	Nonce			[constants.NOUNCE_SIZE]byte
	Transactions	[]TxPublish
}

//OTHERS
type Gossiper struct {
	UIServer       		*net.UDPConn
	GossipAddr     		string
	GossipServer   		*net.UDPConn
	Name           		string
	Simple         		bool
	CurrentBlock		*util.CurrentBlockHash
	Transactions		*TransactionsSet
	Peers          		*util.AddrSet
	ActiveSearches		*util.FullMatchesSet
	NameToMetaHash		sync.Map //Map[name]MetaHash
	VectorClock    		sync.Map //Map[origin]id
	Rumors         		sync.Map //Map[id@origin]GossipPacket	(only rumors)
	Privates       		sync.Map //Map[origin]GossipPacket		(only privates)
	DSDV           		sync.Map //Map[origin]*net.UDPAddr
	IndexedFiles      	sync.Map //Map[metaHash(string)]IndexedFile
	ReceivingFile     	sync.Map //Map[net.UDPAddr]chan([]byte)
	SearchedFiles     	sync.Map //Map[metahash]SearchedFileChunk
	Acks              	sync.Map //Map[origin + id + address]chan(statusPacket)
	SearchRequests    	sync.Map //Map[origin + keyword]SearchRequests
	FinishedSearches  	sync.Map //Map[keywords]chan(Signal)
	Blockchain        	sync.Map //Map[blockHash]block
	ToPrint          	chan string
	ToSend            	chan PacketToSend
	ToAddToBlockchain 	chan Block
	BlockMined        	chan Signal
}

// WEB STRUCTS
type SingleStringJSON struct {
	Text	string
}

type StringAndPeerJSON struct {
	Text	string
	Peer	string
}

type FileRequestJSON struct {
	FileName	string
	Request		string
	Dest		string
}

type WebServerID  struct {
	Name		string
	Address		string
}

// UTILITIES STRUCTS
type PacketToSend struct {
	GossipPacket *GossipPacket
	Address      *net.UDPAddr
}

type CurrentBlock struct {
	HashHex		string
	lock		sync.RWMutex
}

type Signal struct {}