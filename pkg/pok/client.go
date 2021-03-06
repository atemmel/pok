package pok

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

	Active bool
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
		log.Println(err)
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
	data = data[:len(data) - 1]	// Discard newline byte
	id, err := strconv.Atoi(string(data))
	if err != nil {
		log.Println("Id given (", string(data), ") was not valid")
		return -1
	}
	c.Active = true
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
			//TODO Identify which errors should be ignored and which 
			//     errors should abort the connection
			log.Println("Could not read message:", err)
			c.Active = false
			return
		} else {
			player := Player{}
			err := json.Unmarshal(bytes, &player)
			if err == nil {
				c.updatePlayer(&player)
			} else {
				log.Println(err)
				log.Println("Recieved:", string(bytes))
			}
		}
	}
}

func (c *Client) updatePlayer(player *Player) {
	c.playerMap.mutex.Lock()
	if !player.Connected {	// He disconnected
		delete(c.playerMap.players, player.Id)
		log.Println("Player", player.Id, "disconnected")
	} else {
		c.playerMap.players[player.Id] = *player
	}
	c.playerMap.mutex.Unlock()
}

func (c *Client) Disconnect() {
	if !c.Active {
		return
	}
	log.Println("Disconnecting...")
	c.Active = false
	c.rw.WriteByte(0)
	c.conn.Close()
}
