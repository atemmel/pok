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
}

func StartClient() (*Client, error) {
	log.Println("Starting tcp client")
	c := Client{}
	err := error(nil)
	c.conn, err = net.Dial("tcp", "127.0.0.1:3200")
	if err != nil {
		return nil, err
	}
	c.data = make([]byte, 4096)
	return &c, err
}

func (c *Client) GetPlayer() {
	length, err := c.conn.Read(c.data)
	checkError(err)

	player := Player{}
	err = json.Unmarshal(c.data[:length], &player)
	checkError(err)


	log.Println("player.x was:", player.X)
}
