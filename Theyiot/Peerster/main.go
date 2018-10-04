package main

import (
	"fmt"
	"flag"
	"net"
	"os"
	"strconv"
)

/* A Simple function to verify error */
func CheckError(err error) {
	if err  != nil {
		fmt.Println("Error: " , err)
		os.Exit(0)
	}
}

func main() {
	uiPort := flag.Int("UIPort", 8080, "Port for the UI client")
	gossipAddr := flag.String("gossipAddr", "127.0.0.1:5000", "ip:port for the receiver")
	//name := flag.String("name", "nodeA", "name of the gossiper")
	//peers := flag.String("peers", "", "comma-separated list of peers of the form ip:port")
	simple := flag.Bool("simple", false, "run gossiper in simple broadcast mode")

	flag.Parse()

	fmt.Print(strconv.Itoa(*uiPort) + "\n")
	fmt.Print(*gossipAddr + "\n")
	fmt.Printf("%t\n", *simple)


	/* Lets prepare a address at any address at port 10001*/
	ServerAddr,err := net.ResolveUDPAddr("udp",":10001")
	CheckError(err)

	/* Now listen at selected port */
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	defer ServerConn.Close()

	buf := make([]byte, 1024)

	for {
		n,addr,err := ServerConn.ReadFromUDP(buf)
		fmt.Println("Received ",string(buf[0:n]), " from ",addr)

		if err != nil {
			fmt.Println("Error: ",err)
		}
	}
}