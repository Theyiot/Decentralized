package util

import (
	"github.com/Theyiot/Peerster/constants"
	"math/rand"
	"net"
	"strings"
	"sync"
)

type AddrSet struct {
	addresses []*net.UDPAddr
	lock      sync.RWMutex
}

/*
	Add allows the user to add a new address to the set of peers. The method also makes sure that the provided
	address is actually valid and ignores it if it is not the case
 */
func (set *AddrSet) Add(address string) {
	if IsValidAddress(address) && !set.contains(address) {
		peerAddr, err := net.ResolveUDPAddr(constants.UDP_VERSION, address)
		if CheckAndPrintError(err) {
			return
		}

		set.lock.Lock()
		set.addresses = append(set.addresses, peerAddr)
		set.lock.Unlock()
	}
}

/*
	contains checks whether the provided address is already present in our set or not
 */
func (set *AddrSet) contains(addr string) bool {
	set.lock.RLock()
	defer set.lock.RUnlock()
	for _, address := range set.addresses {
		if strings.EqualFold(addr, address.String()) {
			return true
		}
	}
	return false
}

/*
	String returns the string representation of our set. This is that methods that takes care of adding am
	end of line before adding the known peers
 */
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

/*
	Size returns the size of the set
 */
func (set *AddrSet) Size() int {
	set.lock.RLock()
	defer set.lock.RUnlock()
	return len(set.addresses)
}

/*
	returns whether our set is empty or not
 */
func (set *AddrSet) IsEmpty() bool {
	set.lock.RLock()
	defer set.lock.RUnlock()
	return len(set.addresses) <= 0
}

/*
	ChooseRandomPeerExcept chooses a random peer in the set except the one that is given as a
	parameter. This function assumes that the user has checked that there is more than one peer
	in the set. If that is not the case however, an error message is printed and the function
	returns a nil (which will probably break the program anyway)
 */
func (set *AddrSet) ChooseRandomPeerExcept(except string) *net.UDPAddr {
	var address *net.UDPAddr
	set.lock.RLock()
	defer set.lock.RUnlock()
	if set.Size() <= 1 {
		println("ERROR : requesting a random peer except one but he is the only known peer")
		return nil
	}
	for address = set.addresses[rand.Intn(len(set.addresses))] ;
		strings.EqualFold(address.String(), except) ;
		address = set.addresses[rand.Intn(len(set.addresses))] {}

	return address
}

/*
	GetAddressesExcept returns the list of addresses except the one that is given as parameter
 */
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

/*
	GetAddresses returns a copy of the addresses from the set
 */
func (set *AddrSet) GetAddresses() []*net.UDPAddr {
	addressesCopy := make([]*net.UDPAddr, 0)
	set.lock.RLock()
	defer set.lock.RUnlock()
	for _, addr := range set.addresses {
		addressesCopy = append(addressesCopy, addr)
	}
	return addressesCopy
}

/*
	ChooseRandomPeer return a peer chosen at random from the list. This method assumes that the user checks
	that the set is not empty before calling it. However, if the user do call with an empty set, the method
	will display an error message and return a nil address
 */
func (set *AddrSet) ChooseRandomPeer() *net.UDPAddr {
	set.lock.RLock()
	defer set.lock.RUnlock()
	if set.Size() == 0 {
		println("Cannot return random peer from an empty list")
		return nil
	}
	return set.addresses[rand.Intn(len(set.addresses))]
}

/*
	GetAddressesAsStringArray returns a list of all the addresses in the set, in the form of a string array
 */
func (set *AddrSet) GetAddressesAsStringArray() []string {
	addresses := make([]string, 0)
	set.lock.RLock()
	defer set.lock.RUnlock()
	for _, addr := range set.addresses {
		addresses = append(addresses, addr.String())
	}
	return addresses
}

/*
	CreateAddrSet creates an AddrSet from the given string, that may be empty or not
 */
func CreateAddrSet(str string) *AddrSet {
	peers := AddrSet{ addresses: make([]*net.UDPAddr, 0) }
	if strings.EqualFold(str, "") {
		return &peers
	} else {
		split := strings.Split(str, ",")
		for _, peer := range split {
			peers.Add(peer)
		}
	}
	return &peers
}
