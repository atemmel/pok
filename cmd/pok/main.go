package main

import (
	"flag"
	"fmt"
	"github.com/atemmel/pok/pkg/constants"
	"github.com/atemmel/pok/pkg/debug"
	"github.com/atemmel/pok/pkg/pok"
	"github.com/atemmel/pok/pkg/textures"
	"github.com/hajimehoshi/ebiten/v2"
)

var LogFileName string = "error.log"

var onlineEnabled = false
var isServing = false
var disableAudio = false
var fileToOpen string

func init() {
	debug.InitAssert(&LogFileName, false)
	flag.BoolVar(&isServing, "serve", false, "Run as game server")
	flag.BoolVar(&disableAudio, "disable-audio", false, "Toggle audio")
	flag.BoolVar(&pok.DrawDebugInfo, "draw-debug-info", false, "Draw debug info")
	flag.Parse()

	if onlineEnabled {
		if isServing {
			server := pok.NewServer()
			server.Serve()
		}
	}
	fileToOpen = flag.Arg(0)
}

func main() {
	if isServing {
		return
	}

	if fileToOpen == "" {
		fmt.Println("File to open not specified, lacks command line argument")
		return
	}

	ebiten.SetWindowSize(constants.WindowSizeX, constants.WindowSizeY)
	ebiten.SetWindowTitle("pok")
	ebiten.SetWindowResizable(true)

	textures.Init()
	game := pok.CreateGame()

	game.Load(fileToOpen, 0)
	defer game.Save()
	game.Audio = pok.NewAudio()
	if !disableAudio {
		game.PlayAudio()
	}

	if onlineEnabled {
		game.Client = pok.CreateClient()
		connect := func() {
			game.Player.Id = game.Client.Connect()
			if game.Client.Active {
				game.Player.Connected = true
				go game.Client.ReadPlayer()
			}
		}

		go connect()
		defer game.Client.Disconnect()
	}

	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
