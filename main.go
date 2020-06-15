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
var selection *ebiten.Image
var selectionX float64
var selectionY float64

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

	playerImg, _, err = ebitenutil.NewImageFromFile("./resources/images/lucas.png", ebiten.FilterDefault)
	if err != nil {
		log.Fatal(err)
	}

	tileset, _, err = ebitenutil.NewImageFromFile("./resources/images/tileset1.png", ebiten.FilterDefault)
	if err != nil {
		log.Fatal(err)
	}

	selection, err = ebiten.NewImage(32, 32, ebiten.FilterDefault)
	if err != nil {
		panic(err)
	}

	selectionClr := color.RGBA{255, 0, 0, 255}

	for p := 0; p < selection.Bounds().Max.X; p++ {
		selection.Set(p, 0, selectionClr)
		selection.Set(p, selection.Bounds().Max.Y - 1, selectionClr)
	}

	for p := 1; p < selection.Bounds().Max.Y - 1; p++ {
		selection.Set(0, p, selectionClr)
		selection.Set(selection.Bounds().Max.Y - 1, p, selectionClr)
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

const framesPerState = 2
//const playerMaxCycle = 7
const playerMaxCycle = 8
//const playerVelocity = float64(tileSize) / float64(playerMaxCycle * framesPerState) // = 2.285714
const playerVelocity = 2
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

			if player.state % 4 == 0 {
				player.NextAnim()
			}

			player.state++
			if player.state == playerMaxCycle {
				player.state = 0
				if dir == Static || dir != player.dir {
					player.dir = Static
				}
			}
		}
	}
}

func (player *Player) NextAnim() {
	player.tx += 34
	if player.tx >= 34 * 4 {
		player.tx = 0
	}
}

type Camera struct {
	x float64
	y float64
}

func (cam *Camera) LookAt(player *Player) {
	cam.x = player.gx * 2 - 320 / 2 + tileSize + tileSize / 2
	cam.y = player.gy * 2 - 240 / 2 + tileSize + tileSize / 2
}

func (cam *Camera) TransformThenRender(world *ebiten.Image, target *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-cam.x, -cam.y)
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
		selectionX = float64(cx)
		selectionY = float64(cy)
		selectedTile =  cx / tileSize + cy / tileSize * g.tileMap.Width
		fmt.Println("selectedTile:", selectedTile)
	}

	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		g.Save()
		os.Exit(0)
	}

	if ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyK) {
		g.player.Step(Up)
	} else if ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyJ) {
		g.player.Step(Down)
	} else if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyL) {
		g.player.Step(Right)
	} else if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyH) {
		g.player.Step(Left)
	} else {
		g.player.Step(Static)
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.DrawTileset(g.world)

	playerOpt := &ebiten.DrawImageOptions{}
	playerOpt.GeoM.Translate(g.player.gx + playerOffsetX, g.player.gy + playerOffsetY)
	playerOpt.GeoM.Scale(2,2)
	g.world.DrawImage(playerImg.SubImage(image.Rect(g.player.tx, g.player.ty, g.player.tx + tileSize, g.player.ty + tileSize)).(*ebiten.Image), playerOpt)

	/*
	for _, img := range p1img {
		screen.DrawImage(img, nil)
	}
	*/

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(selectionX, selectionY);
	g.world.DrawImage(selection, op)

	g.camera.LookAt(&g.player)
	g.camera.TransformThenRender(g.world, screen)
	ebitenutil.DebugPrint(screen, fmt.Sprintf("camera.x: %f\ncamera.y: %f\nplayer vel: %d", g.camera.x, g.camera.y, playerVelocity) )
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
	game.Load("./resources/tilemaps/tilemap.json")
	fmt.Println(game.tileMap)
	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
