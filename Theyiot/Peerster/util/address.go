package util

import (
	"net"
	"strconv"
	"strings"
)

/*
	IsValidAddress checks whether the provided address, in the form ip:port has a valid IP address and a valid port
 */
func IsValidAddress(peer string) bool {
	ipPort := strings.Split(peer, ":")
	if len(ipPort) != 2 {
		println("ERROR : Invalid Address, should be of the form ip:port, but was : " + peer)
		return false //No need to parse IP and Port if even the form of the Address is wrong
	}
	ip, portStr := ipPort[0], ipPort[1]

	//Check for errors in IP and port
	return CheckValidIP(ip) && CheckValidPort(portStr)
}

/*
	CheckValidIP checks whether a given IP address is actually one
 */
func CheckValidIP(ip string) bool {
	if net.ParseIP(ip) == nil {
		println("ERROR : Invalid IP Address for : " + ip)
		return false
	}
	return true
}

/*
	checkValidPort check that the provided port is valid and not reserved
 */
func CheckValidPort(portStr string) bool {
	port, err := strconv.ParseInt(portStr, 10, 32)
	if CheckAndPrintError(err) {
		return false
	}
	if port <= 1024 || port >= (1 << 16) {
		println("ERROR : Invalid port, should be between 1024 and 65536 (2^16) both excluded, but was : " + portStr)
		return false
	}
	return true
}
