package main

import (
	"fmt"
	"log"
	"net"
)

// TODO: 설정

func handleUDPConnection(conn *net.UDPConn) {

	buffer := make([]byte, 2048)

	n, addr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(addr, ": ", string(buffer[:n]))
}

func main() {

	udpAddr, err := net.ResolveUDPAddr("udp4", "localhost:6000")
	if err != nil {
		log.Fatal(err)
	}

	listen, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer listen.Close()

	fmt.Println("Log Viewer up : using port 6000")

	for {
		handleUDPConnection(listen)
	}

}
