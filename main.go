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
var collisionMarker *ebiten.Image
var selectionX float64
var selectionY float64
var m2Pressed = false

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

	collisionMarker, err = ebiten.NewImage(32, 32, ebiten.FilterDefault)
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

	collisionClr := color.RGBA{255, 0, 255, 255}

	for p := 0; p < 4; p++ {
		for q := 0; q < 4; q++ {
			collisionMarker.Set(p, q, collisionClr)
		}
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
	Collision []bool
}

const playerMaxCycle = 8
const playerVelocity = 1
const playerOffsetX = 7
const playerOffsetY = 1

type Player struct {
	gx float64
	gy float64
	x int
	y int
	animationState int
	frames int
	tx int
	ty int
	dir Direction
	isWalking bool
}

type Direction int

const(
	Static Direction = 0
	Down Direction = 1
	Left Direction = 2
	Right Direction = 3
	Up Direction = 4
)

func (player *Player) TryStep(dir Direction, g *Game) {
	if !player.isWalking && dir == Static {
		player.EndAnim()
		return
	}

	if !player.isWalking {
		player.dir = dir
		ox, oy := player.x, player.y
		player.UpdatePosition()
		index := player.y * g.tileMap.Width + player.x
		if g.tileMap.Collision[index] {
			player.x, player.y = ox, oy	// Restore position
			// Thud noise
			player.dir = dir
			player.ChangeAnim()
			player.Animate()
			player.isWalking = false
		} else {
			player.isWalking = true
		}
	} else {
		player.Animate()
		player.Step(dir, g)
	}
}

func (player *Player) Step(dir Direction, g *Game) {
	player.frames++
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

	if player.frames == tileSize / 2 {
		player.isWalking = false
		player.frames = 0
	}
}

func (player *Player) Animate() {
	if player.animationState % 8 == 0 {
		player.NextAnim()
	}
	player.animationState++
	if player.animationState == playerMaxCycle {
		player.animationState = 0
	}
}

func (player *Player) NextAnim() {
	player.tx += 34
	if player.tx >= 34 * 4 {
		player.tx = 0
	}
}

func (player *Player) ChangeAnim() {
	if player.dir == Up {
		player.ty = 34
	} else if player.dir == Down {
		player.ty = 0
	} else if player.dir == Left {
		player.ty = 34 * 2
	} else if player.dir == Right {
		player.ty = 34 * 3
	}
}

func (player *Player) EndAnim() {
	player.animationState = 0
	player.tx = 0
}

func (player *Player) UpdatePosition() {
	if player.dir == Up {
		player.y--
	} else if player.dir == Down {
		player.y++
	} else if player.dir == Left {
		player.x--
	} else if player.dir == Right {
		player.x++
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

	//TODO Remove code dupe
	if ebiten.IsMouseButtonPressed(ebiten.MouseButton(0)) {
		cx, cy := ebiten.CursorPosition();
		cx += int(g.camera.x)
		cy += int(g.camera.y)
		cx -= cx % tileSize
		cy -= cy % tileSize
		selectionX = float64(cx)
		selectionY = float64(cy)
		selectedTile =  cx / tileSize + cy / tileSize * g.tileMap.Width
		fmt.Println("selectedTile:", selectedTile)
	} 

	if !m2Pressed && ebiten.IsMouseButtonPressed(ebiten.MouseButton(1)) {
		m2Pressed = true
		cx, cy := ebiten.CursorPosition();
		cx += int(g.camera.x)
		cy += int(g.camera.y)
		cx -= cx % tileSize
		cy -= cy % tileSize
		selectionX = float64(cx)
		selectionY = float64(cy)
		selectedTile =  cx / tileSize + cy / tileSize * g.tileMap.Width
		g.tileMap.Collision[selectedTile] = !g.tileMap.Collision[selectedTile]
	} else if !ebiten.IsMouseButtonPressed(ebiten.MouseButton(1)) {
		m2Pressed = false
	}

	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		g.Save()
		os.Exit(0)
	}

	if ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyK) || ebiten.IsKeyPressed(ebiten.KeyW) {
		g.player.TryStep(Up, g)
	} else if ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyJ) || ebiten.IsKeyPressed(ebiten.KeyS) {
		g.player.TryStep(Down, g)
	} else if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyL) || ebiten.IsKeyPressed(ebiten.KeyD) {
		g.player.TryStep(Right, g)
	} else if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyH) || ebiten.IsKeyPressed(ebiten.KeyA) {
		g.player.TryStep(Left, g)
	} else {
		g.player.TryStep(Static, g)
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
	ebitenutil.DebugPrint(screen, fmt.Sprintf(
		`camera.x: %f
camera.y: %f
player.x: %d
player.y: %d
player.isWalking: %t`,
		g.camera.x, g.camera.y, g.player.x, g.player.y, g.player.isWalking) )
}

func (g *Game) Load(str string) {
	data, err := ioutil.ReadFile(str)
	if err != nil {
		fmt.Println("Web build assumed, dumping default file data...")
		data = []byte(`
{"Width":20,"Height":10,"Tiles":[3,0,0,0,5,0,0,0,16,18,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,24,26,0,0,0,0,0,0,0,0,0,0,0,0,0,345,0,3,0,4,32,34,0,0,0,0,0,0,0,0,0,0,6,0,0,0,4,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,630,655,0,4,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,9,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}`)
		//panic(err)
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
		if g.tileMap.Collision[i] {
			screen.DrawImage(collisionMarker, op)
		}
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
	game.player.x = 1
	game.player.y = 1
	game.player.isWalking = false
	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
