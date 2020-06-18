package main

import (
	//"encoding/json"	//TODO Change to protobuf implementation
	"log"
	"net"
	//"time"
)

const MaxConnections = 16

type Server struct {
	listener net.Listener
	conns map[net.Conn] int
	newConn chan net.Conn
	deadConn chan net.Conn
	idGen int
}

func NewServer() Server {
	return Server {
		nil,
		make(map[net.Conn]int),
		make(chan net.Conn),
		make(chan net.Conn),
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

	go func() {
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				log.Println(err)
			}
			s.newConn <- conn
		}
	}()

	for {
		select {
			case conn := <-s.newConn:
				s.conns[conn] = s.idGen
				log.Println("New connection with id", s.idGen)
				s.idGen++
			case conn := <-s.deadConn:
				log.Println("Connection with id", s.conns[conn], "died")
				delete(s.conns, conn)
		}
	}
}
