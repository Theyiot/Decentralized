package gossiper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/Theyiot/Peerster/constants"
	"github.com/Theyiot/Peerster/util"
	"math/rand"
	"net"
	"os"
	"sort"
	"strconv"
)

/*
	indexFile takes care of indexing a file that is in the "_SharedFiles" folder, the user simply has to
	provide the name of the file to index (if the file is placed in the right folder)
 */
func (gossiper *Gossiper) indexFile(fileName string) {
	file, err := os.Open(constants.PATH_SHARED_FILES + fileName)
	defer file.Close()
	if util.CheckAndPrintError(err) {
		return
	}
	fileStat, err := file.Stat()
	if util.CheckAndPrintError(err) {
		return
	} else if fileStat.Size() > constants.CHUNK_SIZE * sha256.Size {
		println("Cannot index file " + fileName + " because it is too big : " + fmt.Sprint(fileStat.Size()) +
			" instead of at most " + fmt.Sprint(constants.CHUNK_SIZE * sha256.Size))
		return
	}
	totalByte := int64(0)
	metaFile := make([]byte, 0)
	for totalByte < fileStat.Size() {
		chunk := make([]byte, constants.CHUNK_SIZE)
		n, err := file.Read(chunk)
		if util.CheckAndPrintError(err) {
			return
		}
		hash := sha256.Sum256(chunk[:n])
		hashString := hex.EncodeToString(hash[:])
		metaFile = append(metaFile, hash[:]...)
		err = writeChunk(hashString, chunk[:n])
		if util.CheckAndPrintError(err) {
			return
		}
		totalByte += int64(n)
	}
	metaHash := sha256.Sum256(metaFile)
	metaHashHex := hex.EncodeToString(metaHash[:])
	indexedFile := IndexedFile{FileName: fileName, FileSize: fileStat.Size(), MetaFile: metaFile}
	gossiper.IndexedFiles.Store(metaHashHex, indexedFile)

	fileTransaction := File{ Name: fileName, Size:totalByte, MetafileHash:metaFile }
	transaction := TxPublish{ HopLimit:constants.HOP_LIMIT_SMALL, File: fileTransaction}
	_, exist := gossiper.NameToMetaHash.Load(transaction.File.Name)
	if exist {
		return
	}
	gossiper.Transactions.Add(&transaction)
	gossiper.broadcastGossipPacket(GossipPacket{ TxPublish: &transaction }, gossiper.Peers.GetAddresses())

	util.CheckAndPrintError(writeChunk(metaHashHex, metaFile))
	return
}

/*
	requestFile allows the user to download and store a file from multiple peers. Hence, the user only needs to
	provide the name under which the file needs to be stored and the metaHash of that file
 */
func (gossiper *Gossiper) requestFile(fileName string, metaHashHex string) {
	searchedFile, found := gossiper.SearchedFiles.Load(metaHashHex)
	if !found {
		println("ERROR : Requesting an unknown file from multiple peers")
		return
	}
	searchedFileChunks := searchedFile.([]SearchedFileChunk)
	if len(searchedFileChunks) < 1 || searchedFileChunks[0].ChunkCount != uint64(len(searchedFileChunks)) {
		println("ERROR : Trying to request for which we don't know where to find all the chunks")
		return
	}
	sort.Slice(searchedFileChunks, func(i, j int) bool {
		return searchedFileChunks[i].ChunkID < searchedFileChunks[j].ChunkID
	})

	destination := searchedFileChunks[0].owningPeers[0]
	metaFile, success := gossiper.requestMetaFile(fileName, destination, metaHashHex)
	if !success || !checkAndPrintSameHash(metaHashHex, metaFile){
		return
	}

	hashesCopy := gossiper.getHashesAsList(metaFile)
	indexedFile := IndexedFile{MetaFile: metaFile, FileName: fileName}

	file, err := os.Create(constants.PATH_DOWNOADS + fileName)
	if util.CheckAndPrintError(err) {
		return
	}
	defer file.Close()

	fileSize := 0
	for i, request := range hashesCopy {
		destination = searchedFileChunks[i].owningPeers[rand.Intn(len(searchedFileChunks[i].owningPeers))]
		n := gossiper.requestFileChunk(fileName, destination, request, i, file)
		fileSize += n
	}
	gossiper.ToPrint <- "RECONSTRUCTED file " + fileName

	indexedFile.FileSize = int64(fileSize)
	gossiper.IndexedFiles.Store(metaHashHex, indexedFile)
}

/*
	requestFileFrom allows the user to download and store a file from a given peer. This method assumes that
	the user is sure that the peer from who it requests that file has the entirety of it
 */
func (gossiper *Gossiper) requestFileFrom(fileName string, destination string, hashHex string) {
	metaFile, success := gossiper.requestMetaFile(fileName, destination, hashHex)
	if !success {
		return
	}

	var indexedFile IndexedFile
	hashesCopy := gossiper.getHashesAsList(metaFile)

	file, err := os.Create(constants.PATH_DOWNOADS + fileName)
	if util.CheckAndPrintError(err) {
		return
	}
	defer file.Close()

	fileSize := 0
	for i, request := range hashesCopy {
		n := gossiper.requestFileChunk(fileName, destination, request, i, file)
		fileSize += n
	}
	gossiper.ToPrint <- "RECONSTRUCTED file " + fileName

	indexedFile.FileSize = int64(fileSize)
	gossiper.IndexedFiles.Store(hashHex, indexedFile)
}

/*
	requestFileChunk allows the user to download and store a chunk of a file from a given peer
 */
func (gossiper *Gossiper) requestFileChunk(fileName string, destination string, request []byte, i int, file *os.File) int {
	str := "DOWNLOADING " + fileName + " chunk " + strconv.Itoa(i + 1) + " from " + destination
	gossiper.ToPrint <- str

	hashHex := hex.EncodeToString(request)
	fileChannel := make(chan[]byte)
	addr, exist := gossiper.DSDV.Load(destination)
	if !exist {
		return 0
	}

	gossiper.ReceivingFile.Store(hashHex, fileChannel)
	defer gossiper.ReceivingFile.Delete(hex.EncodeToString(request))

	chunk, err := gossiper.sendDataRequest(request, destination, addr.(*net.UDPAddr), fileChannel)
	if util.CheckAndPrintError(err) || !checkAndPrintSameHash(hex.EncodeToString(request), chunk) {
		return 0
	}
	n, err := file.Write(chunk)
	if n != len(chunk) {
		println("ERROR : The method write did not write the entire buffer")
		return 0
	}
	if util.CheckAndPrintError(err) {
		os.Remove(constants.PATH_SHARED_FILES + fileName)
		return 0
	}
	return n
}

/*
	requestMetaFile allows the user to download and store a metaFile from a given peer
 */
func (gossiper *Gossiper) requestMetaFile(fileName, destination, hashHex string) ([]byte, bool) {
	addrNotCasted, exist := gossiper.DSDV.Load(destination)
	if !exist {
		println("trying to request metafile from an unknown peer for peer name : " + destination)
		return nil, false
	}
	addr := addrNotCasted.(*net.UDPAddr)

	fileChannel := make(chan[]byte)
	defer close(fileChannel)
	gossiper.ReceivingFile.Store(hashHex, fileChannel)
	defer gossiper.ReceivingFile.Delete(hashHex)

	str := "DOWNLOADING metafile of " + fileName + " from " + destination
	gossiper.ToPrint <- str

	metaFile, err := gossiper.sendDataRequest(stringToHash(hashHex), destination, addr, fileChannel)
	if util.CheckAndPrintError(err) || !checkAndPrintSameHash(hashHex, metaFile) {
		return nil, false
	}
	return metaFile, true
}

/*
	getHashesAsList takes as input a metaFile, which it converts to a list of hash, which is more convenient
	for the downloading process
 */
func (gossiper *Gossiper) getHashesAsList(metaFile []byte) [][]byte {
	hashes := make([][]byte, 0)

	for i := 0 ; i < len(metaFile) / sha256.Size ; i++ {
		index := i * sha256.Size
		hashes = append(hashes, metaFile[index:index + sha256.Size])
	}

	hashesCopy := make([][]byte, 0)
	for _, request := range hashes {
		tmpArray := make([]byte, sha256.Size)
		copy(tmpArray[:], request[:])
		hashesCopy = append(hashesCopy, tmpArray)
	}

	return hashesCopy
}

/*
	writeChunk obviously takes care of writing a chunk, given the hash of this chunk and the data. It writes
	it in the hidden "._FileChunks" folder, to allow the user to keep his indexed files over multiple usage
	of the program
 */
func writeChunk(hashHex string, data []byte) error {
	fileTmp, err := os.Create(constants.PATH_FILE_CHUNKS + hashHex)
	if err != nil {
		return err
	}
	n := len(data)
	byteWrote := 0
	for byteWrote < n {
		byteTmp, err := fileTmp.Write(data[:n])
		if err != nil {
			fileTmp.Close()
			return err
		}
		byteWrote += byteTmp
	}
	fileTmp.Close()

	return nil
}

/*
	readChunk obviously takes care of reading the file corresponding to the given hash, from the "._FileChunks" folder
 */
func readChunk(hashHex string) ([]byte, error) {
	var stat os.FileInfo
	var err error
	if stat, err = os.Stat(constants.PATH_FILE_CHUNKS + hashHex); os.IsNotExist(err) {
		return []byte{}, err
	}
	fileTmp, err := os.Open(constants.PATH_FILE_CHUNKS + hashHex)
	if err != nil {
		return []byte{}, err
	}
	buffer := make([]byte, constants.CHUNK_SIZE)
	byteRead := int64(0)
	for byteRead < stat.Size() {
		byteTmp, err := fileTmp.Read(buffer[byteRead:])
		if err != nil {
			fileTmp.Close()
			return []byte{}, err
		}
		byteRead += int64(byteTmp)
	}
	fileTmp.Close()

	return buffer[:byteRead], nil
}






