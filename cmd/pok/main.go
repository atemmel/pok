package main

import (
	"flag"
	"github.com/atemmel/pok/pkg/pok"
	"github.com/hajimehoshi/ebiten"
)

var isServing = false

func init() {
	flag.BoolVar(&isServing, "serve", false, "Run as game server")
	flag.Parse()
	if isServing {
		server := pok.NewServer()
		server.Serve()
	} else {
		pok.InitGame()
	}
}

func main() {
	if isServing {
		return
	}

	var err error

	ebiten.SetWindowSize(pok.WindowSizeX, pok.WindowSizeY)
	ebiten.SetWindowTitle("Title")
	ebiten.SetWindowResizable(true)

	game := &pok.Game{}
	game.As = &game.Ows

	game.Load(pok.TileMapDir + "old.json", 0)
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
