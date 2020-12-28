package main

import (
	"flag"
	"fmt"
	"github.com/atemmel/pok/pkg/pok"
	"github.com/hajimehoshi/ebiten"
)

var isServing = false
var fileToOpen string

func init() {
	flag.BoolVar(&isServing, "serve", false, "Run as game server")
	flag.StringVar(&fileToOpen, "path", "", "Path of file to load")
	flag.Parse()
	if isServing {
		server := pok.NewServer()
		server.Serve()
	}
}

func main() {
	if isServing {
		return
	}

	if fileToOpen == "" {
		fmt.Println("File to open not specified :/")
		return
	}

	var err error

	ebiten.SetWindowSize(pok.WindowSizeX, pok.WindowSizeY)
	ebiten.SetWindowTitle("Title")
	ebiten.SetWindowResizable(true)

	game := pok.CreateGame()

	//game.Load(pok.TileMapDir + "old.json", 0)
	game.Load(fileToOpen, 0)
	game.Client = pok.CreateClient()
	game.Audio = pok.NewAudio()

	game.PlayAudio()

	game.Player.Id = game.Client.Connect()
	if game.Client.Active {
		game.Player.Connected = true
		go game.Client.ReadPlayer()
	}

	defer game.Client.Disconnect()
	defer game.Save()

	if err = ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
