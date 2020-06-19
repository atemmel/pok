package main

import (
	"encoding/json"
	"io/ioutil"
)

const ConfigDir = "./"
const ServerConfigFile = "config_server.json"
const ClientConfigFile = "config_client.json"

type ServerConfig struct {
	Url string
	Port string
}

type ClientConfig struct {
	ServerUrl string
	ServerPort string
}

//TODO A lot of code dupe here

func ReadServerConfig() (ServerConfig, error) {
	var conf ServerConfig
	data, err := ioutil.ReadFile(ConfigDir + ServerConfigFile)
	if err != nil {
		return conf, err
	}

	err = json.Unmarshal(data, &conf)
	if err != nil {
		return conf, err
	}

	return conf, nil
}

func ReadClientConfig() (ClientConfig, error) {

	var conf ClientConfig
	data, err := ioutil.ReadFile(ConfigDir + ClientConfigFile)
	if err != nil {
		return conf, err
	}

	err = json.Unmarshal(data, &conf)
	if err != nil {
		return conf, err
	}

	return conf, nil
}
