package main

import (
	"encoding/json"	//TODO Change to protobuf implementation
	"log"
	"net"
)

const DataLength = 4096

type Client struct {
	conn net.Conn
	data []byte
	active bool
}

func StartClient() (Client, error) {
	log.Println("Attempting to start tcp client")
	c := Client{}
	err := error(nil)
	c.conn, err = net.Dial("tcp", "127.0.0.1:3200")
	if err != nil {
		log.Println("tcp client failed to start")
		c.active = false
		return c, err
	}
	c.data = make([]byte, 4096)
	c.active = true
	return c, err
}

func (c *Client) GetPlayer() {
	length, err := c.conn.Read(c.data)
	checkError(err)

	player := Player{}
	err = json.Unmarshal(c.data[:length], &player)
	checkError(err)


	log.Println("player.x was:", player.X)
}

func (c *Client) Close() {
	c.conn.Close()
}
