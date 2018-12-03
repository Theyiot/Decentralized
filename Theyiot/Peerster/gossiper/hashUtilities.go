package gossiper

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"github.com/Theyiot/Peerster/util"
	"regexp"
)

/*
	Hash is the unique way to hash a Block for this project
 */
func (b *Block) Hash() (out [32]byte) {
	h := sha256.New()
	h.Write(b.PrevHash[:])
	h.Write(b.Nonce[:])
	binary.Write(h,binary.LittleEndian,
		uint32(len(b.Transactions)))
	for _, t := range b.Transactions {
		th := t.Hash()
		h.Write(th[:])
	}
	copy(out[:], h.Sum(nil))
	return
}

/*
	Hash is the unique way to hash a TxPublish for this project
 */
func (t *TxPublish) Hash() (out [32]byte) {
	h := sha256.New()
	binary.Write(h,binary.LittleEndian,
		uint32(len(t.File.Name)))
	h.Write([]byte(t.File.Name))
	h.Write(t.File.MetafileHash)
	copy(out[:], h.Sum(nil))
	return
}

/*
	IsHexHash check whether a given string corresponds to a correct SHA256 hexadecimal hash
 */
func IsHexHash(hashHex string) bool {
	return regexp.MustCompile("^[[:xdigit:]]{64}?$").MatchString(hashHex)
}

/*
	checkAndPrintSameHash makes sure that the provided SHA256 hash in hexadecimal corresponds to the SHA256
	hash of the provided bytes. If that is not the case, the method returns false and print a message that
	tells that the hashes were different
 */
func checkAndPrintSameHash(hashHex string, data []byte) bool {
	dataHash := sha256.Sum256(data)
	if hex.EncodeToString(dataHash[:]) != hashHex {
		println("ERROR : Hashes " + hex.EncodeToString(dataHash[:]) + " and " + hashHex + " were not equal. File may be dangerous")
		return false
	}
	return true
}

/*
	stringToHash takes a SHA256 hash in hexdecimal and converts it to an array of bytes. It the conversion
	fails or if the input is not valid, an error is displayed and an empty array of bytes is returned
 */
func stringToHash(hashHex string) []byte {
	hash, err := hex.DecodeString(hashHex)
	if util.CheckAndPrintError(err) {
		return []byte{}
	}
	return hash
}
