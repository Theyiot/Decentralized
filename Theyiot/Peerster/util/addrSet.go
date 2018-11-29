package util

import (
	"math/rand"
	"net"
	"strings"
	"sync"
)

type AddrSet struct {
	addresses []*net.UDPAddr
	size      int
	lock      sync.RWMutex
}

func (set *AddrSet) Add(address string) {
	if IsValidAddress(address) && !set.Contains(address) {
		peerAddr, err := net.ResolveUDPAddr("udp4", address)
		if CheckAndPrintError(err) {
			return
		}

		set.lock.Lock()
		set.addresses = append(set.addresses, peerAddr)
		set.size++
		set.lock.Unlock()
	}
}

func (set *AddrSet) Contains(addr string) bool {
	set.lock.RLock()
	defer set.lock.RUnlock()
	for _, address := range set.addresses {
		if strings.EqualFold(addr, address.String()) {
			return true
		}
	}
	return false
}

func (set *AddrSet) String() string {
	str := "\nPEERS "
	set.lock.RLock()
	for i, address := range set.addresses {
		if i > 0 { str += "," }
		str += address.String()
	}
	set.lock.RUnlock()
	return str
}

func (set *AddrSet) GetSize() int {
	set.lock.RLock()
	defer set.lock.RUnlock()
	return set.size
}

func (set *AddrSet) IsEmpty() bool {
	set.lock.RLock()
	defer set.lock.RUnlock()
	return set.size <= 0
}

func (set *AddrSet) ChooseRandomPeerExcept(except string) *net.UDPAddr {
	var address *net.UDPAddr
	set.lock.RLock()
	defer set.lock.RUnlock()
	for address = set.addresses[rand.Intn(set.size)] ;
		strings.EqualFold(address.String(), except) ;
		address = set.addresses[rand.Intn(set.size)] {}

	return address
}

func (set *AddrSet) GetAddressesExcept(except string) []*net.UDPAddr {
	addressesCopy := make([]*net.UDPAddr, 0)
	set.lock.RLock()
	defer set.lock.RUnlock()
	for _, addr := range set.addresses {
		if !strings.EqualFold(addr.String(), except) {
			addressesCopy = append(addressesCopy, addr)
		}
	}
	return addressesCopy
}

func (set *AddrSet) GetAddresses() []*net.UDPAddr {
	addressesCopy := make([]*net.UDPAddr, 0)
	set.lock.RLock()
	defer set.lock.RUnlock()
	for _, addr := range set.addresses {
		addressesCopy = append(addressesCopy, addr)
	}
	return addressesCopy
}

func (set *AddrSet) ChooseRandomPeer() *net.UDPAddr {
	set.lock.RLock()
	defer set.lock.RUnlock()
	return set.addresses[rand.Intn(set.size)]
}

func (set *AddrSet) GetAddressesAsStringArray() []string {
	addresses := make([]string, 0)
	set.lock.RLock()
	defer set.lock.RUnlock()
	for _, addr := range set.addresses {
		addresses = append(addresses, addr.String())
	}
	return addresses
}

func CreateSet(str string) AddrSet {
	peers := AddrSet{ addresses: nil, size: 0 }
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
