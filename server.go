package main

import (
	"encoding/json"	//TODO Change to protobuf implementation
	"log"
	"net"
	"time"
)

const MaxConnections = 16

type Server struct {
	listener net.Listener
	conns []net.Conn
}

func (s *Server) Serve() {
	log.Println("starting tcp server")
	var err error
	s.listener, err = net.Listen("tcp", "127.0.0.1:3200")
	checkError(err)

	go s.distribute()

	for {
		if conn, err := s.listener.Accept(); err == nil {
			s.handleConn(conn)
		}
	}
}

func (s *Server) handleConn(conn net.Conn) {
	if len(s.conns) >= MaxConnections {
		log.Println("Maximum number of connections reached")
		return
	}

	log.Println("New connection")
	s.conns = append(s.conns, conn)
}

func checkError(err error) {
	if err != nil {
		log.Println(err.Error())
	}
}

func (s *Server) distribute() {
	player := Player{}
	player.X = 4
	data, err := json.Marshal(player)
	checkError(err)

	for {
		time.Sleep(1000 * time.Millisecond)
		for i, c := range s.conns {
			_, err = c.Write(data)

			if err != nil {	// Close connection on error
				s.conns[i].Close()
				s.conns = append(s.conns[:i], s.conns[i + 1:]...)
			}
		}
	}
}

/*
func handleConn(conn net.Conn) {
	log.Println("Client connected!")
	defer conn.Close()

	player := Player{}
	log.Println(player)
	data, err := json.Marshal(player)
	checkError(err)

	length, err := conn.Write(data)
	checkError(err)

	log.Println("Sent", length, "bytes of data")
}
*/
