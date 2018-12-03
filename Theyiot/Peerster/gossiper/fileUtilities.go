package gossiper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/Theyiot/Peerster/util"
	"math/rand"
	"net"
	"os"
	"sort"
	"strconv"
)

func (gossiper *Gossiper) indexFile(fileName string) {
	file, err := os.Open(PATH_SHARED_FILES + fileName)
	defer file.Close()
	if util.CheckAndPrintError(err) {
		return
	}
	fileStat, err := file.Stat()
	if util.CheckAndPrintError(err) {
		return
	} else if fileStat.Size() > CHUNK_SIZE * sha256.Size {
		println("Cannot index file " + fileName + " because it is too big : " + fmt.Sprint(fileStat.Size()) +
			" instead of at most " + fmt.Sprint(CHUNK_SIZE * sha256.Size))
		return
	}
	totalByte := int64(0)
	metaFile := make([]byte, 0)
	for totalByte < fileStat.Size() {
		chunk := make([]byte, CHUNK_SIZE)
		n, err := file.Read(chunk)
		if util.CheckAndPrintError(err) {
			return
		}
		hash := sha256.Sum256(chunk[:n])
		hashString := hex.EncodeToString(hash[:])
		metaFile = append(metaFile, hash[:]...)
		err = writeFile(hashString, chunk[:n])
		if util.CheckAndPrintError(err) {
			return
		}
		totalByte += int64(n)
	}
	metaHash := sha256.Sum256(metaFile)
	metaHashHex := hex.EncodeToString(metaHash[:])
	indexedFile := IndexedFile{FileName: fileName, FileSize: fileStat.Size(), MetaFile: metaFile}
	gossiper.IndexedFiles.Store(metaHashHex, indexedFile)

	util.CheckAndPrintError(writeFile(metaHashHex, metaFile))
	return
}

/*
This method allows the user to download and store, with a given file name, a file corresponding to the provided metahash
 */
func (gossiper *Gossiper) requestFile(fileName string, hashHex string) {
	searchedFile, found := gossiper.SearchedFiles.Load(hashHex)
	if !found {
		println("Requesting an unknown file from multiple peers for : " + hashHex)
		return
	}
	searchedFileChunks := searchedFile.([]SearchedFileChunk)
	if len(searchedFileChunks) < 1 || searchedFileChunks[0].ChunkCount != uint64(len(searchedFileChunks)) {
		println("Trying to request for which we don't know where to find all the chunks")
		return
	}
	sort.Slice(searchedFileChunks, func(i, j int) bool {
		return searchedFileChunks[i].ChunkID < searchedFileChunks[j].ChunkID
	})

	destination := searchedFileChunks[0].owningPeers[0]
	metaFile, success := gossiper.requestMetaFile(fileName, destination, hashHex)
	if !success || !checkAndPrintSameHash(hashHex, metaFile){
		return
	}

	hashesCopy := gossiper.GetHashesCopy(metaFile)
	indexedFile := IndexedFile{MetaFile: metaFile, FileName: fileName}

	file, err := os.Create(PATH_DOWNOADS + fileName)
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
	gossiper.IndexedFiles.Store(hashHex, indexedFile)
}

func (gossiper *Gossiper) requestFileFrom(fileName string, destination string, hashHex string) {
	metaFile, success := gossiper.requestMetaFile(fileName, destination, hashHex)
	if !success {
		return
	}

	var indexedFile IndexedFile
	hashesCopy := gossiper.GetHashesCopy(metaFile)

	file, err := os.Create(PATH_DOWNOADS + fileName)
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
		os.Remove(PATH_SHARED_FILES + fileName)
		return 0
	}
	return n
}

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

func (gossiper *Gossiper) GetHashesCopy(metaFile []byte) [][]byte {
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

func writeFile(fileName string, data []byte) error {
	fileTmp, err := os.Create(PATH_FILE_CHUNKS + fileName)
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

func readFile(hash string) ([]byte, error) {
	var stat os.FileInfo
	var err error
	if stat, err = os.Stat(PATH_FILE_CHUNKS + hash); os.IsNotExist(err) {
		return []byte{}, err
	}
	fileTmp, err := os.Open(PATH_FILE_CHUNKS + hash)
	if err != nil {
		return []byte{}, err
	}
	buffer := make([]byte, CHUNK_SIZE)
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






