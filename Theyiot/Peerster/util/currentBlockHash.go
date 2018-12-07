package util

import (
	"github.com/Theyiot/Peerster/constants"
	"sync"
)

type CurrentBlockHash struct {
	currentHash		string
	depth			uint64
	lock			sync.RWMutex
}

func (current *CurrentBlockHash) GetCurrentHash() string {
	current.lock.RLock()
	defer current.lock.RUnlock()
	return current.currentHash
}
func (current *CurrentBlockHash) SetCurrentHash(newHash string) {
	current.lock.RLock()
	defer current.lock.RUnlock()
	current.currentHash = newHash
}

func (current *CurrentBlockHash) GetDepth() uint64 {
	current.lock.RLock()
	defer current.lock.RUnlock()
	return current.depth
}

func (current *CurrentBlockHash) IncrementDepth() {
	current.lock.RLock()
	defer current.lock.RUnlock()
	current.depth++
}

func (current *CurrentBlockHash) SetDepth(newDepth uint64) {
	current.lock.RLock()
	defer current.lock.RUnlock()
	current.depth = newDepth
}

func CreateCurrentBlockHash() *CurrentBlockHash {
	return &CurrentBlockHash{ currentHash:constants.DEFAULT_BLOCK_HASH }
}