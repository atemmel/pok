package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/atemmel/pok/pkg/pok"
	"github.com/hajimehoshi/ebiten"
	"io/ioutil"
)

var isServing = false
var buildPath = ""
var buildW = 0
var buildH = 0

func init() {
	flag.BoolVar(&isServing, "serve", false, "Run as game server")
	flag.StringVar(&buildPath, "build", "", "Generates a blank JSON map of dimension NxM")
	flag.IntVar(&buildW, "width", 10, "Desired width of JSON build")
	flag.IntVar(&buildH, "height", 10, "Desired height of JSON build")
	flag.Parse()
	if len(buildPath) > 0 {
		return
	}
	if isServing {
		server := pok.NewServer()
		server.Serve()
	} else {
		pok.InitGame()
	}
}

func build() {
	tex := make([][]int, 1)
	tex[0] = make([]int, buildW * buildH)

	col := make([][]bool, 1)
	col[0] = make([]bool, buildW * buildH)

	tiles := pok.TileMap{
		tex,
		col,
		tex,
		make([]string, 0),
		make([]pok.Exit, 0),
		make([]pok.Entry, 0),
		buildW,
		buildH,
	}

	fmt.Println("Wrote", buildW, "*", buildH, "=", buildW * buildH, "tileset")

	bytes, _ := json.Marshal(tiles)
	ioutil.WriteFile(buildPath, bytes, 0644)
}

func main() {
	if isServing {
		return
	}

	if len(buildPath) > 0 {
		build()
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
