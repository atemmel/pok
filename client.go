package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"strconv"
	"sync"
)

const DataLength = 4096

type Client struct {
	conf ClientConfig
	rw *bufio.ReadWriter
	conn net.Conn
	playerMap PlayerMap
	active bool
}

type PlayerMap struct {
	players map[int]Player
	mutex sync.Mutex
}

func CreateClient() Client {
	return Client{
		ClientConfig{},
		nil,
		nil,
		PlayerMap{
			make(map[int]Player),
			sync.Mutex{},
		},
		false,
	}
}

func (c *Client) Connect() int {
	log.Println("Attempting to connect to server...")
	var err error

	c.conf, err = ReadClientConfig()
	if err != nil {
		log.Println("Could not read client config")
		return -1
	}

	c.conn, err = net.Dial("tcp", c.conf.ServerUrl + ":" + c.conf.ServerPort)
	if err != nil {
		log.Println("Connection failed")
		return -1
	}

	log.Println("Connection succeeded!")
	c.rw = bufio.NewReadWriter(
		bufio.NewReader(c.conn),
		bufio.NewWriter(c.conn),
	)
	data, err := c.rw.ReadBytes('\n')
	if err != nil {
		log.Println("Could not be given an id")
		return -1
	}
	data = data[:len(data) - 1]	// remove newline byte
	id, err := strconv.Atoi(string(data))
	if err != nil {
		log.Println("Id given (", string(data), ") was not valid")
		return -1
	}
	c.active = true
	return id
}

func (c *Client) WritePlayer(player *Player) {
	data, _ := json.Marshal(player)
	data = append(data, '\n')
	c.rw.Write(data)
	c.rw.Flush()
}

func (c *Client) ReadPlayer() {
	for {
		bytes, err := c.rw.ReadBytes('\n')
		if err != nil {
			log.Println("Could not read message...")
		} else {
			log.Println("Read:", string(bytes))
			player := Player{}
			err := json.Unmarshal(bytes, &player)
			if err == nil {
				c.updatePlayer(player)
			} else {
				log.Println(err)
			}
		}
	}
}

func (c *Client) updatePlayer(player Player) {
	c.playerMap.mutex.Lock()
	c.playerMap.players[player.id] = player
	c.playerMap.mutex.Unlock()
}

func (c *Client) Disconnect() {
	if !c.active {
		return
	}
	log.Println("Disconnecting...")
	c.active = false
	c.rw.WriteByte(0)
	c.conn.Close()
}
