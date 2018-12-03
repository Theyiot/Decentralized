package gossiper

import (
	"encoding/json"
	"github.com/Theyiot/Peerster/util"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

/*
	sendPublicMessage allows to send from the UI some public or rumor messages
 */
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

/*
	receivePublicMessage allows the user to receive simple or rumor messages to display them on the UI
 */
func receivePublicMessage(gossiper *Gossiper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rumors := gossiper.getRumorsAsList()
		json.NewEncoder(w).Encode(rumors)
	}
}

/*
	addPeerFromWeb allows the user to add new peers from the UI
 */
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

/*
	refreshPeersAddress allows the user to display the latest update of the new peers' addresses on the UI
 */
func refreshPeersAddress(gossiper *Gossiper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		peers := gossiper.Peers.GetAddressesAsStringArray()
		json.NewEncoder(w).Encode(peers)
	}
}


/*
	getAndSetPersonalID allows the user to provide his personal identifiers to display them in the UI
 */
func getAndSetPersonalID(gossiper *Gossiper) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		id := WebServerID{ Name: gossiper.Name, Address: gossiper.GossipAddr }
		json.NewEncoder(w).Encode(id)
	}
}

/*
	refreshPeersName allows the to display the latest update of the new peers' names on the UI
 */
func refreshPeersName(gossiper *Gossiper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		names := gossiper.getPeersNameAsList()
		json.NewEncoder(w).Encode(names)
	}
}

/*
	sendPrivateMessage allows to send from the UI some private messages
 */
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

/*
	receivePrivateMessage allows the user to receive private messages to display them on the UI
 */
func receivePrivateMessage(gossiper *Gossiper) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		mapMessages := gossiper.getPrivateMessagesAsMap()
		json.NewEncoder(w).Encode(mapMessages)
	}
}

/*
	indexFile allows the user to index a file he has chosen from the UI
 */
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
		mapIndexedFiles := gossiper.getIndexedFilesAsMap()
		json.NewEncoder(w).Encode(mapIndexedFiles)
	}
}

/*
	listIndexedFiles allows the user to obtain a list of all indexed files, in order to display them in the UI
 */
func listIndexedFiles(gossiper *Gossiper) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		mapIndexedFiles := gossiper.getIndexedFilesAsMap()
		json.NewEncoder(w).Encode(mapIndexedFiles)
	}
}

/*
	requestFile allows the user to request a given file from the UI
 */
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
		mapIndexedFiles := gossiper.getIndexedFilesAsMap()
		json.NewEncoder(w).Encode(mapIndexedFiles)
	}
}

/*
	This function takes care of starting the web server and to link all the function to the right path
 */
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