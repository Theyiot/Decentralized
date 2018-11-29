package gossiper

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/Theyiot/Peerster/util"
	"regexp"
)

func IsHexHash(hash string) bool {
	return regexp.MustCompile("^[[:xdigit:]]{64}?$").MatchString(hash)
}

func checkAndPrintSameHash(hash string, data []byte) bool {
	dataHash := sha256.Sum256(data)
	if hex.EncodeToString(dataHash[:]) != hash {
		println("Hashes " + hex.EncodeToString(dataHash[:]) + " and " + hash + " were not equal. File may be dangerous")
		return false
	}
	return true
}

func stringToHash(str string) []byte {
	hash, err := hex.DecodeString(str)
	if util.CheckAndPrintError(err) {
		return []byte{}
	}
	return hash
}
