package main

import (
	"fmt"
	"net"
	"time"
)

func ping(host string, port string) error {
	address := net.JoinHostPort(host, port)
	conn, err := net.DialTimeout("udp", address, 1*time.Second)
	if conn != nil {
		fmt.Println(conn.LocalAddr())
		defer conn.Close()
	}
	return err
}
