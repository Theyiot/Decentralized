package gossiper

import (
	"encoding/json"
	"github.com/Theyiot/Peerster/util"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

//MESSAGE POST
func sendPublicMessage(gossiper *Gossiper) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		var msg SingleStringJSON
		if r.Body == nil {
			http.Error(w, "The request should not be empty", 400)
			return
		}
		err := json.NewDecoder(r.Body).Decode(&msg)
		if util.CheckAndPrintError(err) {
			http.Error(w, err.Error(), 400)
			return
		}
		if gossiper.Simple {
			gossiper.sendSimplePacket(msg.Text)
		} else {
			gossiper.sendRumorPacket(msg.Text)
		}
	}
}

//MESSAGE GET
func receivePublicMessage(gossiper *Gossiper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rumors := gossiper.GetRumorsAsList()
		json.NewEncoder(w).Encode(rumors)
	}
}

//PEERS POST
func addPeerFromWeb(gossiper *Gossiper) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		var address SingleStringJSON
		if r.Body == nil {
			http.Error(w, "The request should not be empty", 400)
			return
		}
		err := json.NewDecoder(r.Body).Decode(&address)
		if util.CheckAndPrintError(err) {
			http.Error(w, err.Error(), 400)
			return
		}
		gossiper.Peers.Add(address.Text)
	}
}

//PEERS GET
func refreshPeersAddress(gossiper *Gossiper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		peers := gossiper.Peers.GetAddressesAsStringArray()
		json.NewEncoder(w).Encode(peers)
	}
}

//ID GET
func getAndSetPersonalID(gossiper *Gossiper) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		id := WebServerID{ Name: gossiper.Name, Address: gossiper.GossipAddr }
		json.NewEncoder(w).Encode(id)
	}
}

//PEERS NAME GET
func refreshPeersName(gossiper *Gossiper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		names := gossiper.GetPeersNameAsMap()
		json.NewEncoder(w).Encode(names)
	}
}

func sendPrivateMessage(gossiper *Gossiper) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		var msg StringAndPeerJSON
		if r.Body == nil {
			http.Error(w, "The request should not be empty", 400)
			return
		}
		err := json.NewDecoder(r.Body).Decode(&msg)
		if util.CheckAndPrintError(err) {
			http.Error(w, err.Error(), 400)
			return
		}
		gossiper.sendPrivatePacket(msg.Text, msg.Peer)
	}
}

func receivePrivateMessage(gossiper *Gossiper) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		mapMessages := gossiper.GetPrivateMessagesAsMap()
		json.NewEncoder(w).Encode(mapMessages)
	}
}

func indexFile(gossiper *Gossiper) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		var fileName SingleStringJSON
		if r.Body == nil {
			http.Error(w, "The request should not be empty", 400)
			return
		}
		err := json.NewDecoder(r.Body).Decode(&fileName)
		if util.CheckAndPrintError(err) {
			http.Error(w, err.Error(), 400)
			return
		}
		gossiper.indexFile(fileName.Text)
		mapIndexedFiles := gossiper.GetIndexedFilesAsMap()
		json.NewEncoder(w).Encode(mapIndexedFiles)
	}
}

func listIndexedFiles(gossiper *Gossiper) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		mapIndexedFiles := gossiper.GetIndexedFilesAsMap()
		json.NewEncoder(w).Encode(mapIndexedFiles)
	}
}

func requestFile(gossiper *Gossiper) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		var fileRequest FileRequestJSON
		if r.Body == nil {
			http.Error(w, "The request should not be empty", 400)
			return
		}
		err := json.NewDecoder(r.Body).Decode(&fileRequest)
		if util.CheckAndPrintError(err) {
			http.Error(w, err.Error(), 400)
			return
		}

		gossiper.requestFileFrom(fileRequest.FileName, fileRequest.Dest, fileRequest.Request)
		mapIndexedFiles := gossiper.GetIndexedFilesAsMap()
		json.NewEncoder(w).Encode(mapIndexedFiles)
	}
}

func (gossiper *Gossiper) StartWebServer(port string) {
	if !util.CheckValidPort(port) {
		os.Exit(0)
	}
	r := mux.NewRouter()

	// MESSAGES
	r.HandleFunc("/message", receivePublicMessage(gossiper)).Methods("GET")
	r.HandleFunc("/message", sendPublicMessage(gossiper)).Methods("POST")

	// PEERS
	r.HandleFunc("/node", refreshPeersAddress(gossiper)).Methods("GET")
	r.HandleFunc("/node", addPeerFromWeb(gossiper)).Methods("POST")

	// ID
	r.HandleFunc("/id", getAndSetPersonalID(gossiper)).Methods("GET")

	// PEER NAMES
	r.HandleFunc("/name", refreshPeersName(gossiper)).Methods("GET")

	// PRIVATE MESSAGES
	r.HandleFunc("/private", receivePrivateMessage(gossiper)).Methods("GET")
	r.HandleFunc("/private", sendPrivateMessage(gossiper)).Methods("POST")

	// INDEXED FILES
	r.HandleFunc("/fileIndexing", indexFile(gossiper)).Methods("POST")
	r.HandleFunc("/fileIndexing", listIndexedFiles(gossiper)).Methods("GET")
	r.HandleFunc("/fileRequesting", requestFile(gossiper)).Methods("POST")

	//LINK FRONTEND AND BACKEND
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("webserver")))

	//LAUNCHING SERVER
	log.Fatal(http.ListenAndServe("localhost:" + port, r))
}