package gossiper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/Theyiot/Peerster/util"
	"net"
	"os"
	"strconv"
)

func (gossiper *Gossiper) indexFile(fileName string) {
	file, err := os.Open(SHARED_FILES_PATH + fileName)
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

	err = writeFile(metaHashHex, metaFile)
	if util.CheckAndPrintError(err) {
		return
	}

	return
}

func (gossiper *Gossiper) requestFile(fileName string, dest string, hashHex string) {
	addrNotCasted, exist := gossiper.DSDV.Load(dest)
	if !exist {
		println("ERROR : trying to request a file from an unknown peer")
		return
	}
	addr := addrNotCasted.(*net.UDPAddr)

	fileChannel := make(chan[]byte)
	defer close(fileChannel)
	gossiper.ReceivingFile.Store(addr.String(), fileChannel)

	str := "DOWNLOADING metafile of " + fileName + " from " + dest
	gossiper.ToPrint <- str

	metaFile, err := gossiper.sendDataRequest(stringToHash(hashHex), dest, addr, fileChannel)
	if util.CheckAndPrintError(err) {
		gossiper.ReceivingFile.Delete(addr.String())
		return
	}
	if !checkAndPrintSameHash(hashHex, metaFile) {
		gossiper.ReceivingFile.Delete(addr.String())
		return
	}

	var indexedFile IndexedFile
	hashes := make([][]byte, 0)

	indexedFile = IndexedFile{MetaFile: metaFile, FileName: fileName}
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

	fileSize := 0
	file, err := os.Create(DOWNLOAD_PATH + fileName)
	if util.CheckAndPrintError(err) {
		return
	}
	defer file.Close()

	for i, request := range hashesCopy {
		str := "DOWNLOADING " + fileName + " chunk " + strconv.Itoa(i + 1) + " from " + dest
		gossiper.ToPrint <- str

		chunk, err := gossiper.sendDataRequest(request, dest, addr, fileChannel)
		if util.CheckAndPrintError(err) || !checkAndPrintSameHash(hex.EncodeToString(request), chunk) {
			gossiper.ReceivingFile.Delete(addr.String())
			os.Remove(SHARED_FILES_PATH + fileName)
			return
		}
		n, err := file.Write(chunk)
		if n != len(chunk) {
			println("ERROR : The method write did not write the entire buffer")
			gossiper.ReceivingFile.Delete(addr.String())
			os.Remove(SHARED_FILES_PATH + fileName)
			return
		}
		if util.CheckAndPrintError(err) {
			gossiper.ReceivingFile.Delete(addr.String())
			os.Remove(SHARED_FILES_PATH + fileName)
			return
		}
		fileSize += n
	}
	str = "RECONSTRUCTED file " + fileName
	gossiper.ToPrint <- str

	indexedFile.FileSize = int64(fileSize)
	gossiper.IndexedFiles.Store(hashHex, indexedFile)
}

func writeFile(fileName string, data []byte) error {
	fileTmp, err := os.Create(FILE_CHUNKS_PATH + fileName)
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
	if stat, err = os.Stat(FILE_CHUNKS_PATH + hash); os.IsNotExist(err) {
		return []byte{}, err
	}
	fileTmp, err := os.Open(FILE_CHUNKS_PATH + hash)
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






