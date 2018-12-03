package util

import (
	"errors"
	"sync"
	"sync/atomic"
)

type FullMatchesSet struct {
	keywords 		[][]string
	fullMatches		[]uint32
	size			int
	lock     		sync.RWMutex
}

func (set *FullMatchesSet) Contains(str []string) bool {
	set.lock.Lock()
	defer set.lock.Unlock()
	for _, list := range set.keywords {
		if equals(list, str) {
			return true
		}
	}
	return false
}

func (set *FullMatchesSet) Add(str []string) {
	if !set.Contains(str) {
		set.lock.Lock()
		set.keywords = append(set.keywords, str)
		set.fullMatches = append(set.fullMatches, 0)
		set.size++
		set.lock.Unlock()
	}
}

func (set *FullMatchesSet) Size() int {
	set.lock.RLock()
	defer set.lock.RUnlock()
	return set.size
}

func (set *FullMatchesSet) IncrementFullMatchIndex(i int) (uint32, error) {
	set.lock.Lock()
	defer set.lock.Unlock()
	if i < 0 || i >= set.size {
		return 0, errors.New("Index out of bounds")
	}
	return atomic.AddUint32(&set.fullMatches[i], 1), nil
}

func (set *FullMatchesSet) Remove(i int) {
	set.lock.Lock()
	defer set.lock.Unlock()
	if i < 0 || i >= set.size {
		return
	}
	newSet := make([][]string, 0)
	newFullMatches := make([]uint32, 0)
	for j := 0 ; j < set.size ; j++ {
		if i == j {
			continue
		}
		newSet = append(newSet, set.keywords[j])
		newFullMatches = append(newFullMatches, set.fullMatches[j])
	}
	set.keywords = newSet
	set.fullMatches = newFullMatches
	set.size--
}

func (set *FullMatchesSet) GetSetCopy() [][]string {
	set.lock.RLock()
	defer set.lock.RUnlock()
	setCopy := make([][]string, 0)
	for i := 0 ; i < set.size ; i++ {
		setCopy = append(setCopy, set.keywords[i])
	}
	return setCopy
}

func CreateStringSet() *FullMatchesSet {
	return &FullMatchesSet{ keywords: make([][]string, 0), fullMatches:make([]uint32, 0), size:0 }
}

func equals(l1, l2 []string) bool {
	if len(l1) != len(l2) {
		return false
	}
	for i := 0 ; i < len(l1) ; i++ {
		if l1[i] != l2[i] {
			return false
		}
	}
	return true
}