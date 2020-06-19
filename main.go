package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"io/ioutil"
	"image"
	"image/color"
	"encoding/json"
	"log"
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

var isServing = false

func init() {
	flag.BoolVar(&isServing, "serve", false, "Run as game server")
	flag.Parse()
	if isServing {
		server := NewServer()
		server.Serve()
	} else {
		initGame()
	}
}

func initGame() {
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
	client Client
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

func (player *Player) TryStep(dir Direction, g *Game) {
	if !player.IsWalking && dir == Static {
		if player.AnimationState != 0 {
			player.Animate()
		} else {
			player.EndAnim()
		}
		return
	}

	if !player.IsWalking {
		player.Dir = dir
		ox, oy := player.X, player.Y
		player.UpdatePosition()
		index := player.Y * g.tileMap.Width + player.X
		if g.tileMap.Collision[index] {
			player.X, player.Y = ox, oy	// Restore position
			// Thud noise
			player.Dir = dir
			player.ChangeAnim()
			player.Animate()
			player.IsWalking = false
		} else {
			player.IsWalking = true
		}
	} else {
		player.Animate()
		player.Step(dir, g)
	}
}

func (player *Player) Step(dir Direction, g *Game) {
	player.Frames++
	if player.Dir == Up {
		player.Ty = 34
		player.Gy += -playerVelocity
	} else if player.Dir == Down {
		player.Ty = 0
		player.Gy += playerVelocity
	} else if player.Dir == Left {
		player.Ty = 34 * 2
		player.Gx += -playerVelocity
	} else if player.Dir == Right {
		player.Ty = 34 * 3
		player.Gx += playerVelocity
	}

	if player.Frames == tileSize / 2 {
		player.IsWalking = false
		player.Frames = 0
	}
}

func (player *Player) Animate() {
	if player.AnimationState % 8 == 0 {
		player.NextAnim()
	}
	player.AnimationState++
	if player.AnimationState == playerMaxCycle {
		player.AnimationState = 0
	}
}

func (player *Player) NextAnim() {
	player.Tx += 34
	if player.Tx >= 34 * 4 {
		player.Tx = 0
	}
}

func (player *Player) ChangeAnim() {
	if player.Dir == Up {
		player.Ty = 34
	} else if player.Dir == Down {
		player.Ty = 0
	} else if player.Dir == Left {
		player.Ty = 34 * 2
	} else if player.Dir == Right {
		player.Ty = 34 * 3
	}
}

func (player *Player) EndAnim() {
	player.AnimationState = 0
	player.Tx = 0
}

func (player *Player) UpdatePosition() {
	if player.Dir == Up {
		player.Y--
	} else if player.Dir == Down {
		player.Y++
	} else if player.Dir == Left {
		player.X--
	} else if player.Dir == Right {
		player.X++
	}
}

type Camera struct {
	x float64
	y float64
}

func (cam *Camera) LookAt(player *Player) {
	cam.x = player.Gx * 2 - 320 / 2 + tileSize + tileSize / 2
	cam.y = player.Gy * 2 - 240 / 2 + tileSize + tileSize / 2
}

func (cam *Camera) TransformThenRender(world *ebiten.Image, target *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-cam.x, -cam.y)
	target.DrawImage(world, op)
}

var selectedTile = 0

var ticks = 0

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
		return errors.New("")	//TODO Gotta be a better way to do this
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

	ticks++

	if ticks % 60 == 0 {
		if g.client.active {
			g.client.WritePlayer(&g.player)
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.DrawTileset(g.world)
	g.DrawPlayer(g.player)

	if g.client.active {
		g.client.playerMap.mutex.Lock()
		for _, player := range g.client.playerMap.players {
			g.DrawPlayer(player)
		}
		g.client.playerMap.mutex.Unlock()
	}

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
		g.camera.x, g.camera.y, g.player.X, g.player.Y, g.player.IsWalking) )
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

func (g *Game) DrawPlayer(player Player) {
	playerOpt := &ebiten.DrawImageOptions{}
	playerOpt.GeoM.Translate(player.Gx + playerOffsetX, player.Gy + playerOffsetY)
	playerOpt.GeoM.Scale(2,2)
	g.world.DrawImage(playerImg.SubImage(image.Rect(player.Tx, player.Ty, player.Tx + tileSize, player.Ty + tileSize)).(*ebiten.Image), playerOpt)
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
	if isServing {
		return
	}

	var err error

	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Title")
	ebiten.SetWindowResizable(true)

	game := &Game{}
	game.world, _ = ebiten.NewImage(640, 480, ebiten.FilterDefault)
	game.Load("./resources/tilemaps/tilemap.json")
	game.player.X = 1
	game.player.Y = 1
	game.client = CreateClient()

	game.player.id = game.client.Connect()
	if game.client.active {
		go game.client.ReadPlayer()
	}

	defer game.client.Disconnect()
	defer game.Save()

	if err = ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
