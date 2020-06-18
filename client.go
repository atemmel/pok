package main

import (
	"log"
	"net"
)

const DataLength = 4096

type Client struct {
	conn net.Conn
	data []byte
	active bool
}

func CreateClient() Client {
	return Client{
		nil,
		make([]byte, DataLength),
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
		c.active = true
	 }
}

func (c *Client) Disconnect() {

}
