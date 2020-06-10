package main

import (
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"fmt"
	"io/ioutil"
	"image"
	"image/color"
	"encoding/json"
	"log"
)

var tileset *ebiten.Image
var p1img []*ebiten.Image
var dot *ebiten.Image
var dotX float64
var dotY float64

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

	dot, err = ebiten.NewImage(32, 32, ebiten.FilterDefault)
	if err != nil {
		panic(err)
	}

	dotClr := color.RGBA{255, 0, 0, 255}

	for p := 0; p < dot.Bounds().Max.X; p++ {
		dot.Set(p, 0, dotClr)
		dot.Set(p, dot.Bounds().Max.Y - 1, dotClr)
	}

	for p := 1; p < dot.Bounds().Max.Y - 1; p++ {
		dot.Set(0, p, dotClr)
		dot.Set(dot.Bounds().Max.Y - 1, p, dotClr)
	}
}

type Game struct{
	tileMap TileMap
}

type TileMap struct {
	Width int
	Height int
	Tiles []int
}

var scroll = 0.
var selectedTile = 0

func (g *Game) Update(screen *ebiten.Image) error {
	_, dy := ebiten.Wheel()
	if dy != 0. && len(g.tileMap.Tiles) > selectedTile {
		if dy < 0 {
			g.tileMap.Tiles[selectedTile]--
		} else {
			g.tileMap.Tiles[selectedTile]++
		}
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButton(0)) {
		cx, cy := ebiten.CursorPosition();
		cx -= cx % tileSize
		cy -= cy % tileSize
		dotX = float64(cx)
		dotY = float64(cy)
		selectedTile =  cx / tileSize + cy / tileSize * g.tileMap.Width
		fmt.Println("selectedTile:", selectedTile)
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.DrawTileset(screen)
	for _, img := range p1img {
		screen.DrawImage(img, nil)
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(dotX, dotY);
	screen.DrawImage(dot, op)
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
