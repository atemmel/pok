package main

import (
	"flag"
	"github.com/atemmel/pok/pkg/constants"
	"github.com/atemmel/pok/pkg/debug"
	"github.com/atemmel/pok/pkg/pok"
	"github.com/atemmel/pok/pkg/textures"
	"github.com/hajimehoshi/ebiten/v2"
)

func init() {
	flag.Parse()
}

func main() {
	log := "editorerror.log"
	debug.InitAssert(&log, true)
	textures.Init()
	ed := pok.NewEditor(flag.Args())

	ebiten.SetWindowSize(constants.WindowSizeX, constants.WindowSizeY)
	ebiten.SetWindowTitle("poked")
	ebiten.SetWindowResizable(true)

	if err := ebiten.RunGame(ed); err != nil {
		panic(err)
	}
}
