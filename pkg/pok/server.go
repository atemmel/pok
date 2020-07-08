package pok

import (
	"bufio"
	"encoding/json"	//TODO Change to protobuf implementation
	"log"
	"net"
	"strconv"
	"sync"
)

const MaxConnections = 16

type Message struct {
	author net.Conn
	contents []byte
}

type Server struct {
	conf ServerConfig
	listener net.Listener
	conns map[net.Conn] int
	connsMutex sync.Mutex
	newConn chan net.Conn
	deadConn chan net.Conn
	messageChan chan Message
	idGen int
}

func NewServer() Server {
	return Server {
		ServerConfig{},
		nil,
		make(map[net.Conn]int),
		sync.Mutex{},
		make(chan net.Conn),
		make(chan net.Conn),
		make(chan Message),
		0,
	}
}

func (s *Server) Serve() {
	var err error
	log.Println("Starting server...")
	s.conf, err = ReadServerConfig()
	if err != nil {
		log.Println("Could not read server config!")
		panic(err)
	}
	s.listener, err = net.Listen("tcp", s.conf.Url + ":" + s.conf.Port)
	if err != nil {
		panic(err)
	}

	log.Println("Server is now running")

	go s.acceptConnections()

	for {
		select {
			case conn := <-s.newConn:
				log.Println("New connection with id", s.idGen)
				s.designate(conn, s.idGen)
				go s.readClient(conn, s.idGen)
				s.idGen++
			case conn := <-s.deadConn:
				log.Println("Connection with id", s.conns[conn], "died")
				s.disconnect(conn)
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

		if s.isValidMessage(data) {
			s.messageChan <- Message{conn, data}
		} else {
			log.Println("Ill-formed message recieved:")
		}
	}

	s.deadConn <- conn
}

func (s *Server) designate(conn net.Conn, id int) {
	s.connsMutex.Lock()
	s.conns[conn] = id
	s.connsMutex.Unlock()

	msg := strconv.Itoa(id) + "\n"
	conn.Write([]byte(msg) )
}

func (s *Server) isValidMessage(bytes []byte) bool {
	player := Player{}
	err := json.Unmarshal(bytes, &player)
	return err == nil
}

func (s *Server) broadcast(message Message) {
	s.connsMutex.Lock()
	for c := range s.conns {
		if c != message.author {
			c.Write(message.contents)
		}
	}
	s.connsMutex.Unlock()
}

func (s *Server) disconnect(conn net.Conn) {
	player := Player{}

	s.connsMutex.Lock()
	player.Id = s.conns[conn]
	player.Connected = false
	bytes, _ := json.Marshal(player)
	bytes = append(bytes, '\n')

	delete(s.conns, conn)

	for c, id := range s.conns {
		log.Println("Sending kill message from", player.Id, "to", id)
		c.Write(bytes)
	}
	s.connsMutex.Unlock()
	log.Println("Kill message sent")
}
