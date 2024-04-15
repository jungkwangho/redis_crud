package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

func run_redis_checker(myaddr string) {

	for {
		setter(myaddr)
		time.Sleep(time.Duration(CONFIG.RetryMilliSecs) * time.Millisecond)
	}
}

func handleUDPConnection(conn *net.UDPConn) {

	buffer := make([]byte, 2048)

	n, addr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		fmt.Println("handleUDPConnection -> conn.ReadFromUDP: ", err.Error())
		log.Fatal(err)
	}

	var checkReq CheckRequest
	var checkRes CheckResult
	err = json.Unmarshal(buffer[:n], &checkReq)
	if err != nil {
		fmt.Println("handleUDPConnection -> conn.Unmarshal: ", err.Error())
		log.Fatal(err.Error())
	}

	result, err := getter(checkReq.Key, checkReq.Value)
	if err != nil {
		checkRes.RemoteError = err.Error()
		checkRes.RemoteResult = result
	} else {
		checkRes.RemoteError = ""
		checkRes.RemoteResult = result
	}

	jsonBytes, err := json.Marshal(checkRes)
	if err != nil {
		LOG.UdpSend(err.Error())
	}

	_, err = conn.WriteToUDP(jsonBytes, addr)
	if err != nil {
		LOG.UdpSend(err.Error())
	}

}

func main() {

	// 인자로 listen할 주소를 받는다.
	var myaddr string

	if len(os.Args) > 1 {
		myaddr = os.Args[1]
	} else {
		log.Fatal("Listen address should be given like 'crud 127.0.0.1:9999'")
		os.Exit(-1)
	}

	// 설정을 불러온다.
	CONFIG.LoadConfig()

	// 로그를 활성화 한다.
	LOG.InitUdpSender(CONFIG.LogServer)

	// redis 체크 로직을 실행한다.
	go run_redis_checker(myaddr)

	// remote_getter 를 처리할 서버를 기동한다.
	udpAddr, err := net.ResolveUDPAddr("udp4", myaddr)
	if err != nil {
		fmt.Println("main -> net.ResolveUDPAddr: ", err.Error())
		log.Fatal(err)
	}

	listen, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("main -> net.ListenUDP: ", err.Error())
		log.Fatal(err)
	}
	defer listen.Close()

	fmt.Println("CRUD server up : using address ", myaddr)

	for {
		handleUDPConnection(listen)
	}

}
