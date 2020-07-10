package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/atemmel/pok/pkg/pok"
	"github.com/hajimehoshi/ebiten"
	"io/ioutil"
)

var buildPath = ""
var buildW = 0
var buildH = 0

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

func init() {
	flag.StringVar(&buildPath, "build", "", "Generates a blank JSON map of dimension NxM")
	flag.IntVar(&buildW, "width", 10, "Desired width of JSON build")
	flag.IntVar(&buildH, "height", 10, "Desired height of JSON build")
	flag.Parse()
	if len(buildPath) > 0 {
		return
	} else {
		pok.InitGame()
	}
}

func main() {
	if len(buildPath) > 0 {
		build()
		return
	}

	ed := pok.NewEditor()

	ebiten.SetWindowSize(pok.WindowSizeX, pok.WindowSizeY)
	ebiten.SetWindowTitle("Title")
	ebiten.SetWindowResizable(true)

	if err := ebiten.RunGame(ed); err != nil {
		panic(err)
	}
}
