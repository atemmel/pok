package main

import (
	"fmt"
	"github.com/lonng/nano"
	"github.com/lonng/nano/component"
	"github.com/lonng/nano/serialize/json"
	"github.com/lonng/nano/session"
	"net/http"
)

type Server struct {
	component.Base
	group *nano.Group
}

func NewServer() *Server {
	return &Server{
		group: nano.NewGroup("server"),
	}
}

func (s *Server) Init() {
	fmt.Println("Init was run")
}

func (s *Server) AfterInit() {
	fmt.Println("AfterInit was run")
}

func (s *Server) BeforeShutdown() {
	fmt.Println("BeforeShutdown was run")
}

func (s *Server) Shutdown() {
	fmt.Println("Shutdown was run")
}

func (s *Server) JoinHandler(session *session.Session, player *Player) error {
	fmt.Println("Something joined...")
	return nil
}

func (s *Server) PlayerHandler(session *session.Session, player *Player) error {
	fmt.Println("Player did something...")
	return nil
}

func serve() {
	nano.Register(NewServer())
	nano.SetSerializer(json.NewSerializer())
	nano.EnableDebug()
	nano.SetCheckOriginFunc(func(_ *http.Request) bool { return true } )
	nano.Listen(":3250")
}

