package gossiper

import (
	"github.com/Theyiot/Peerster/util"
	"net"
	"sync"
	"time"
)

//DEFAULT PACKETS
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

type PeerStatus struct {
	Identifier	string
	NextID		uint32
}

type StatusPacket struct {
	Want	[]PeerStatus
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

type GossipPacket struct {
	Simple        *SimpleMessage
	Rumor         *RumorMessage
	Status        *StatusPacket
	Private       *PrivateMessage
	DataRequest   *DataRequest
	DataReply     *DataReply
	SearchRequest *SearchRequest
	SearchReply   *SearchReply
}

type ClientGossipPacket struct {
	Simple				*SimpleMessage
	Private 			*PrivateMessage
	FileRequest			*FileRequestMessage
	FileIndex			*FileIndexMessage
	FileSearchRequest	*SearchRequestMessage
}

type PacketToSend struct {
	GossipPacket *GossipPacket
	Address      *net.UDPAddr
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
	FileName		string
	ChunkID			uint64
	ChunkCount		uint64
	owningPeers		[]string
	lock			sync.RWMutex
}

type ActiveSearches struct {
	Keywords		*util.FullMatchesSet
}

//OTHERS
type Gossiper struct {
	UIServer       *net.UDPConn
	GossipAddr     string
	GossipServer   *net.UDPConn
	Name           string
	Simple         bool
	Peers          *util.AddrSet
	VectorClock    sync.Map //Map[origin]id
	Rumors         sync.Map //Map[id@origin]GossipPacket	(only rumors)
	Privates       sync.Map //Map[origin]GossipPacket		(only privates)
	DSDV           sync.Map //Map[origin]*net.UDPAddr
	IndexedFiles   sync.Map //Map[metaHash(string)]IndexedFile
	ReceivingFile  sync.Map //Map[net.UDPAddr]chan []byte
	SearchedFiles  sync.Map //Map[metahash]SearchedFileChunk
	Acks           sync.Map //Map[origin + id + address]channel
	SearchRequests sync.Map //Map[origin + keyword]SearchRequests
	ActiveSearches	*util.FullMatchesSet
	ToPrint        chan string
	ToSend         chan PacketToSend
}

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