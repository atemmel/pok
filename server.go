package main

import (
	//"encoding/json"	//TODO Change to protobuf implementation
	"bufio"
	"encoding/json"
	"log"
	"net"
	//"time"
)

const MaxConnections = 16

type Message struct {
	author net.Conn
	contents []byte
}

type Server struct {
	listener net.Listener
	conns map[net.Conn] int
	newConn chan net.Conn
	deadConn chan net.Conn
	messageChan chan Message
	idGen int
}

func NewServer() Server {
	return Server {
		nil,
		make(map[net.Conn]int),
		make(chan net.Conn),
		make(chan net.Conn),
		make(chan Message),
		0,
	}
}

func (s *Server) Serve() {
	var err error
	log.Println("Starting server...")
	s.listener, err = net.Listen("tcp", ":3200")
	if err != nil {
		panic(err)
	}

	log.Println("Server is now running")

	go s.acceptConnections()

	for {
		select {
			case conn := <-s.newConn:
				s.conns[conn] = s.idGen
				log.Println("New connection with id", s.idGen)
				go s.readClient(conn, s.idGen)
				s.idGen++
			case conn := <-s.deadConn:
				log.Println("Connection with id", s.conns[conn], "died")
				delete(s.conns, conn)
			case message := <-s.messageChan:
				s.broadcast(message)
		}
	}
}

func (s *Server) acceptConnections() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Println(err)
		}

		if len(s.conns) >= MaxConnections {
			log.Println("Maximum number of active connections reached, connection dismissed")
			conn.Close()
		} else {
			s.newConn <- conn
		}
	}
}

func (s *Server) readClient(conn net.Conn, id int) {
	reader := bufio.NewReader(conn)
	for {
		data, err := reader.ReadBytes('\n')
		if err != nil {
			break
		}
		log.Println("Recieved message from", id)
		if s.isValidMessage(data) {
			data = append(data, '\n')
			s.messageChan <- Message{conn, data}
		}
	}

	s.deadConn <- conn
}

func (s *Server) isValidMessage(bytes []byte) bool {
	player := Player{}
	err := json.Unmarshal(bytes, &player)
	if err != nil {
		log.Println("Ill-formed message recieved:", err)
	}
	return err == nil
}

func (s *Server) broadcast(message Message) {
	for c := range s.conns {
		if c != message.author {
			c.Write(message.contents)
		}
	}
}
