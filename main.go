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
	"os"
)

var tileset *ebiten.Image
var p1img []*ebiten.Image
var playerImg *ebiten.Image
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
	p1img = append(p1img, img)

	playerImg, _, err = ebitenutil.NewImageFromFile("./resources/images/lcuas.png", ebiten.FilterDefault)
	if err != nil {
		log.Fatal(err)
	}

	tileset, _, err = ebitenutil.NewImageFromFile("./resources/images/tileset1.png", ebiten.FilterDefault)
	if err != nil {
		log.Fatal(err)
	}

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
	path string
	player Player
	world *ebiten.Image
	camera Camera
}

type TileMap struct {
	Width int
	Height int
	Tiles []int
}

const framesPerState = 3
const playerMaxCycle = 4
const playerVelocity = (tileSize / 2) / playerMaxCycle
const playerOffsetX = 7
const playerOffsetY = 1

type Player struct {
	gx float64
	gy float64
	x int
	y int
	state int
	frames int
	tx int
	ty int
	dir Direction
}

type Direction int

const(
	Static Direction = 0
	Down Direction = 1
	Left Direction = 2
	Right Direction = 3
	Up Direction = 4
)

func (player *Player) Step(dir Direction) {
	if player.dir == Static {
		player.dir = dir
	} else {
		player.frames++

		if player.frames == framesPerState {
			player.frames = 0
			if player.dir == Up {
				player.ty = 34
				player.gy += -playerVelocity
			} else if player.dir == Down {
				player.ty = 0
				player.gy += playerVelocity
			} else if player.dir == Left {
				player.ty = 34 * 2
				player.gx += -playerVelocity
			} else if player.dir == Right {
				player.ty = 34 * 3
				player.gx += playerVelocity
			}

			if player.state % 2 == 0 {
				player.NextAnim()
			}
			player.state++
			if player.state == playerMaxCycle {
				player.state = 0
				player.dir = Static
				player.tx = 0
			}
		}
	}

}

func (player *Player) NextAnim() {
	player.tx += 34
}

type Camera struct {
	x *float64
	y *float64
}

func (cam *Camera) TransformThenRender(world *ebiten.Image, target *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	if cam.x != nil && cam.y != nil {
		op.GeoM.Translate(-*cam.x, -*cam.y)
	}
	target.DrawImage(world, op)
}

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

	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		g.Save()
		os.Exit(0)
	}

	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		g.player.Step(Up)
	} else if ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.player.Step(Down)
	} else if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.player.Step(Right)
	} else if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.player.Step(Left)
	} else {
		g.player.Step(Static)
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	//g.DrawTileset(screen)
	g.DrawTileset(g.world)

	playerOpt := &ebiten.DrawImageOptions{}
	playerOpt.GeoM.Translate(g.player.gx + playerOffsetX, g.player.gy + playerOffsetY)
	playerOpt.GeoM.Scale(2,2)
	//screen.DrawImage(playerImg.SubImage(image.Rect(g.player.tx, g.player.ty, g.player.tx + tileSize, g.player.ty + tileSize)).(*ebiten.Image), playerOpt)
	g.world.DrawImage(playerImg.SubImage(image.Rect(g.player.tx, g.player.ty, g.player.tx + tileSize, g.player.ty + tileSize)).(*ebiten.Image), playerOpt)

	/*
	for _, img := range p1img {
		screen.DrawImage(img, nil)
	}
	*/

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(dotX, dotY);
	g.world.DrawImage(dot, op)

	//screen.DrawImage(g.world, op)
	g.camera.TransformThenRender(g.world, screen)
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
	g.path = str
}

func (g *Game) Save() {
	bytes, err := json.Marshal(g.tileMap)
	if err != nil {
		fmt.Println(err)
	}
	ioutil.WriteFile(g.path, bytes, 0644)
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
	game.world, _ = ebiten.NewImage(640, 480, ebiten.FilterDefault)
	game.camera.x = &game.player.gx
	game.camera.y = &game.player.gy
	game.Load("./resources/tilemaps/tilemap.json")
	fmt.Println(game.tileMap)
	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
