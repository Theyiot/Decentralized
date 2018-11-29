package util

import (
	"math/rand"
	"net"
	"strings"
	"sync"
)

type AddrSet struct {
	Addresses	[]*net.UDPAddr
	Size     	int
	lock	 	sync.RWMutex
}

func (set *AddrSet) Add(address string) {
	if IsValidAddress(address) && !set.Contains(address) {
		peerAddr, err := net.ResolveUDPAddr("udp4", address)
		if CheckAndPrintError(err) {
			return
		}

		set.lock.Lock()
		set.Addresses = append(set.Addresses, peerAddr)
		set.Size++
		set.lock.Unlock()
	}
}

func (set *AddrSet) Contains(addr string) bool {
	set.lock.RLock()
	defer set.lock.RUnlock()
	for _, address := range set.Addresses {
		if strings.EqualFold(addr, address.String()) {
			return true
		}
	}
	return false
}

func (set *AddrSet) String() string {
	str := "\nPEERS "
	set.lock.RLock()
	for i, address := range set.Addresses {
		if i > 0 { str += "," }
		str += address.String()
	}
	set.lock.RUnlock()
	return str
}

func (set *AddrSet) GetSize() int {
	set.lock.RLock()
	defer set.lock.RUnlock()
	return set.Size
}

func (set *AddrSet) IsEmpty() bool {
	set.lock.RLock()
	defer set.lock.RUnlock()
	return set.Size <= 0
}

func (set *AddrSet) ChooseRandomPeerExcept(except string) *net.UDPAddr {
	var address *net.UDPAddr
	set.lock.RLock()
	defer set.lock.RUnlock()
	for address = set.Addresses[rand.Intn(set.Size)] ;
		strings.EqualFold(address.String(), except) ;
		address = set.Addresses[rand.Intn(set.Size)] {}

	return address
}

func (set *AddrSet) ChooseRandomPeer() *net.UDPAddr {
	set.lock.RLock()
	defer set.lock.RUnlock()
	return set.Addresses[rand.Intn(set.Size)]
}

func (set *AddrSet) GetAddressesAsStringArray() []string {
	addresses := make([]string, 0)
	set.lock.RLock()
	defer set.lock.RUnlock()
	for _, addr := range set.Addresses {
		addresses = append(addresses, addr.String())
	}
	return addresses
}

func CreateSet(str string) AddrSet {
	peers := AddrSet{ Addresses: nil, Size: 0 }
	if strings.EqualFold(str, "") {
		return peers
	} else {
		split := strings.Split(str, ",")
		for _, peer := range split {
			peers.Add(peer)
		}
	}
	return peers
}
