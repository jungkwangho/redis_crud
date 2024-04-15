package main

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	CrudServers    []string `json:"crud_servers"`
	RedisServers   []string `json:"redis_servers"`
	CrudServersLB  string   `json:"crud_servers_lb"`
	RedisServersLB string   `json:"redis_servers_lb"`
	LogServer      string   `json:"log_server"`
	RetryMilliSecs int64    `json:"retry_millisecs"`
	GetRetry       int64    `json:"retry_get"`
}

func (config *Config) LoadConfig() {

	bytes, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatal("Failed to read config:", err.Error())
	}
	err = json.Unmarshal(bytes, config)
	if err != nil {
		log.Fatal("Failed to unmarshal config: ", err.Error())
	}

}

func (config *Config) SaveConfig() {

	bytes, err := json.Marshal(config)
	if err != nil {
		log.Fatal("Failed to marshal config: ", err.Error())
	}
	err = os.WriteFile("config.json", bytes, 0660)
	if err != nil {
		log.Fatal("Failed to write config: ", err.Error())
	}

}

var CONFIG Config
