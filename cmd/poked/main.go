package main

import (
	"github.com/atemmel/pok/pkg/pok"
	"github.com/hajimehoshi/ebiten"
)

func main() {
	ed := pok.NewEditor()

	ebiten.SetWindowSize(pok.WindowSizeX, pok.WindowSizeY)
	ebiten.SetWindowTitle("Title")
	ebiten.SetWindowResizable(true)

	if err := ebiten.RunGame(ed); err != nil {
		panic(err)
	}
}
