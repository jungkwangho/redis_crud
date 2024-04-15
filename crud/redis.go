package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func remove(slice []string, s int) []string {
	return append(slice[:s], slice[s+1:]...)
}

func get_partial_list(items []string, exclude string) []string {

	if len(items) <= 0 {
		return []string{}
	}

	var i int
	for i = 0; i < len(items); i++ {
		if items[i] == exclude {
			break
		}
	}

	return remove(items, i)
}

func get_random_item(items []string, exclude string) string {

	var allthesame bool
	allthesame = true

	for i := 0; i < len(items); i++ {
		if items[i] != exclude {
			allthesame = false
			break
		}
	}

	if allthesame {
		return ""
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for {
		i := r.Intn(len(items))
		if items[i] != exclude {
			return items[i]
		}
	}
}

func get_my_key() string {
	// TEST_KEY_<PID>_<TIMESTAMP> 키 이름을 구한다.
	pid := strconv.Itoa(os.Getpid())
	now := fmt.Sprint(time.Now().UnixMicro())

	return "KEY_" + pid + "_" + now
}

func get_random_string(length int) string {
	// 랜덤한 문자열울 구한다. (64)
	chars := "1234567890abcdefzhijklmlopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	str := ""
	for i := 0; i < length; i++ {
		str = str + string(chars[rand.Intn(len(chars))])
	}
	return str
}

func setter(myaddr string) {

	var rdb interface{}

	if CONFIG.RedisServersLB != "" {
		rdb = redis.NewClient(&redis.Options{
			Addr:     CONFIG.RedisServersLB,
			Password: "",
			DB:       0,
		})
	} else {
		redis_servers := CONFIG.RedisServers
		rdb = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    redis_servers,
			Password: "",
		})
	}

	key := get_my_key()
	value := get_random_string(64)

RETRY:
	var err error
	if reflect.TypeOf(rdb).Kind() == 22 {
		err = rdb.(*redis.Client).Set(ctx, key, value, 0).Err()
	} else {
		err = rdb.(*redis.ClusterClient).Set(ctx, key, value, 0).Err()
	}

	if err != nil {
		if err.Error() == "OOM command not allowed when used memory > 'maxmemory'." {
			fmt.Println("WARN: setter -> Set: ", key, " memory exceeded")
			time.Sleep(time.Second)
			goto RETRY
		} else {
			fmt.Println("ERROR: setter -> rdb.Set: ", err.Error())
			log.Fatal(err.Error())
		}
	}

	local_getter(key, value)
	fmt.Println("local getter end")

	remote_getter(myaddr, key, value)
	fmt.Println("remote getter end")

	if reflect.TypeOf(rdb).Kind() == 22 {
		rdb.(*redis.Client).Del(ctx, key)
	} else {
		rdb.(*redis.ClusterClient).Del(ctx, key)
	}
}

func getter(key, value string) (string, error) {

	var rdb interface{}

	if CONFIG.RedisServersLB != "" {
		rdb = redis.NewClient(&redis.Options{
			Addr:     CONFIG.RedisServersLB,
			Password: "",
			DB:       0,
		})
	} else {
		redis_servers := CONFIG.RedisServers
		rdb = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    redis_servers,
			Password: "",
		})
	}

	i := 0
RETRY:

	var get string
	var err error
	if reflect.TypeOf(rdb).Kind() == 22 {

		get, err = rdb.(*redis.Client).Get(ctx, key).Result()
	} else {
		get, err = rdb.(*redis.ClusterClient).Get(ctx, key).Result()

	}
	if err != nil {
		i++
		if true {
			time.Sleep(time.Second)
			if err == redis.Nil {
				fmt.Println("WARN: getter -> Get: ", key, " not exists.")
				time.Sleep(time.Second)
				goto RETRY
			} else if err.Error() == "i/o timeout" {
				fmt.Println("WARN: getter -> Get: ", key, " i/o timeout.")
				time.Sleep(time.Second)
				goto RETRY
			} else {

				log.Fatal("Can't get value from redis: ", err.Error(), "\n")
			}

		} else {
			log.Fatal("Can't get value from redis\n")
		}
	}
	//fmt.Println("Get from server: ", os.Getpid(), "key: ", key, " get: ", get)

	if get != value {
		return "<FAIL> 1:" + get + " 2:" + value + " ", errors.New("not the same")
	} else {
		return "<OK>", nil
	}
}

func local_getter(key, value string) {

	result, err := getter(key, value)
	if err != nil {
		LOG.UdpSend(result + err.Error())
	}
}

func strlist2map(strlist []string, exclude string) map[string]bool {
	resultmap := make(map[string]bool)
	for i := 0; i < len(strlist); i++ {
		if strlist[i] != exclude {
			resultmap[strlist[i]] = false
		}
	}
	return resultmap
}

func allchecked(checkmap *map[string]bool) bool {
	for _, v := range *checkmap {
		if !v {
			return false
		}
	}
	return true
}

func check(checkmap *map[string]bool, key string) {
	(*checkmap)[key] = true
}

func remote_getter(myaddr string, key, value string) {

	checkmap := strlist2map(CONFIG.CrudServers, myaddr)

START_CHECK:

	if allchecked(&checkmap) {
		fmt.Println("All checked")
		return
	}

	// 현재 서버주소를 제외한 crud_server 주소 구함
	crud_server := get_random_item(CONFIG.CrudServers, myaddr)
	if crud_server == "" {
		fmt.Println("Can't get another crud server")
		return
	}

	// 다른 crud 서버의 check 결과를 가져온다.
	var checkRes CheckResult
	var checkReq CheckRequest
	checkReq.Key = key
	checkReq.Value = value

	jsonBytes, err := json.Marshal(checkReq)
	if err != nil {
		log.Fatal("Failed to Marshal CheckRequest")
		return
	}

	remoteAddr, err := net.ResolveUDPAddr("udp", crud_server)
	if err != nil {
		log.Fatal("Failed to resolve udp addr")
		return
	}

	conn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		log.Fatal("Failed to dial udp")
		return
	}
	defer conn.Close()

	_, err = conn.Write(jsonBytes)
	if err != nil {
		log.Fatal("Failed to write udp")
		return
	}

	buffer := make([]byte, 2048)
	n, _, err := conn.ReadFromUDP(buffer)

	check(&checkmap, crud_server)

	if err != nil {
		goto START_CHECK
	}

	err = json.Unmarshal(buffer[:n], &checkRes)
	if err != nil {
		log.Fatal("Failed to marshal CheckResult")
		return
	}

	if checkRes.RemoteError != "" {
		fmt.Println("RemoteResult: ", checkRes.RemoteResult)
		LOG.UdpSend(checkRes.RemoteResult)
	} else {
		return
	}

}
