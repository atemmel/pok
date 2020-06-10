package main

import (
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"fmt"
	"io/ioutil"
	"image"
	"encoding/json"
	"log"
)

var tileset *ebiten.Image
var p1img []*ebiten.Image

const (
	tileSize = 32
	nTilesX = 8
)

func init() {
	img, _, err := ebitenutil.NewImageFromFile("./resources/images/geng.png", ebiten.FilterDefault)
	if err != nil {
		log.Fatal(err)
	}
	tileset, _, err = ebitenutil.NewImageFromFile("./resources/images/tileset1.png", ebiten.FilterDefault)
	if err != nil {
		log.Fatal(err)
	}
	p1img = append(p1img, img)
}

type Game struct{
	tileMap TileMap
}

type TileMap struct {
	Width int
	Height int
	Tiles []int
}

func (g *Game) Update(screen *ebiten.Image) error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.DrawTileset(screen)
	for _, img := range p1img {
		screen.DrawImage(img, nil)
	}
}

func (g *Game) Load(str string) {
	data, err := ioutil.ReadFile(str)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, &g.tileMap)
	if err != nil {
		panic(err)
	}
}

func (g *Game) DrawTileset(screen *ebiten.Image) {
	for i, n := range g.tileMap.Tiles {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(i%g.tileMap.Width)*tileSize, float64(i/g.tileMap.Width)*tileSize)

		tx := (n % nTilesX) * tileSize
		ty := (n / nTilesX) * tileSize

		screen.DrawImage(tileset.SubImage(image.Rect(tx, ty, tx + tileSize, ty + tileSize)).(*ebiten.Image), op)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 240
}

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Title")
	ebiten.SetWindowResizable(true)

	game := &Game{}
	game.Load("./resources/tilemaps/tilemap.json")
	fmt.Println(game.tileMap)
	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
