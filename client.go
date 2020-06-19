package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
)

const DataLength = 4096

type Client struct {
	rw *bufio.ReadWriter
	conn net.Conn
	active bool
}

func CreateClient() Client {
	return Client{
		nil,
		nil,
		false,
	}
}

func (c *Client) Connect() {
	log.Println("Attempting to connect to server...")
	 var err error
	 c.conn, err = net.Dial("tcp", ":3200")
	 if err != nil {
		log.Println("Connection failed")
	 } else {
		log.Println("Connection succeeded!")
		c.rw = bufio.NewReadWriter(
			bufio.NewReader(c.conn),
			bufio.NewWriter(c.conn),
		)
		c.active = true
	 }
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
		}
	}
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
