package main

import (
	"fmt"
	"log"
	"net"
)

type Logger struct {
	addr *net.UDPAddr
}

func (logger *Logger) InitUdpSender(server string) {
	var err error
	logger.addr, err = net.ResolveUDPAddr("udp", server)
	if err != nil {
		fmt.Println("InitUdpSender -> net.ResolveUDPAddr: ", err.Error())
		log.Fatal(err)
	}
}

func (logger *Logger) UdpSend(logtext string) {

	conn, err := net.DialUDP("udp", nil, logger.addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	msg := []byte(logtext)
	_, err = conn.Write(msg)
	if err != nil {
		log.Fatal(err)
	}
}

var LOG Logger
