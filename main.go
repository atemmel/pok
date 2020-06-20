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
var turnCheck = 0

const (
	tileSize = 32
	nTilesX = 8
	TurnCheckLimit = 5	// Frames
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
	client Client
	rend Renderer
}

type TileMap struct {
	Width int
	Height int
	Tiles []int
	Collision []bool
}

const playerMaxCycle = 8
const playerVelocity = 2
const playerOffsetX = 13
const playerOffsetY = 0

func (player *Player) TryStep(dir Direction, g *Game) {
	if !player.isWalking && dir == Static {
		if turnCheck > 0 && turnCheck < TurnCheckLimit && 
			player.animationState == 0 {
			player.Animate()
		}
		turnCheck = 0
		if player.animationState != 0 {
			player.Animate()
		} else {
			player.EndAnim()
		}
		return
	}

	if !player.isWalking {
		if player.dir == dir {
			turnCheck++
		}
		player.dir = dir
		player.ChangeAnim()
		if turnCheck >= TurnCheckLimit {
			ox, oy := player.X, player.Y
			player.UpdatePosition()
			if g.TileIsOccupied(player.X, player.Y) {
				player.X, player.Y = ox, oy	// Restore position
				// Thud noise
				player.dir = dir
				player.Animate()
				player.isWalking = false
			} else {
				player.isWalking = true
			}
		}
	}
}

func (player *Player) Update() {
	if !player.isWalking {
		return
	}

	player.Animate()
	player.Step()
}

func (g *Game) TileIsOccupied(x int, y int) bool {
	index := y * g.tileMap.Width + x

	// Out of bounds check
	if index >= len(g.tileMap.Tiles) || index < 0 {
		return true
	}

	if g.tileMap.Collision[index] {
		return true
	}

	for _, p := range g.client.playerMap.players {
		if p.X == x && p.Y == y {
			return true
		}
	}

	return false
}

func (player *Player) Step() {
	player.frames++
	if player.dir == Up {
		player.Ty = 34
		player.Gy += -playerVelocity
	} else if player.dir == Down {
		player.Ty = 0
		player.Gy += playerVelocity
	} else if player.dir == Left {
		player.Ty = 34 * 2
		player.Gx += -playerVelocity
	} else if player.dir == Right {
		player.Ty = 34 * 3
		player.Gx += playerVelocity
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
	player.Tx += 34
	if player.Tx >= 34 * 4 {
		player.Tx = 0
	}
}

func (player *Player) ChangeAnim() {
	if player.dir == Up {
		player.Ty = 34
	} else if player.dir == Down {
		player.Ty = 0
	} else if player.dir == Left {
		player.Ty = 34 * 2
	} else if player.dir == Right {
		player.Ty = 34 * 3
	}
}

func (player *Player) EndAnim() {
	player.animationState = 0
	player.Tx = 0
}

func (player *Player) UpdatePosition() {
	if player.dir == Up {
		player.Y--
	} else if player.dir == Down {
		player.Y++
	} else if player.dir == Left {
		player.X--
	} else if player.dir == Right {
		player.X++
	}
}

func (g *Game) CenterRendererOnPlayer() {
	g.rend.LookAt(
		//g.player.Gx * 2 - 320 / 2 + tileSize + tileSize / 2,
		//g.player.Gy * 2 - 240 / 2 + tileSize + tileSize / 2,
		g.player.Gx - 320 / 2 + tileSize + tileSize / 2,
		g.player.Gy - 240 / 2 + tileSize + tileSize/ 2,
	)
}

var selectedTile = 0

var ticks = 0

func (g *Game) Update(screen *ebiten.Image) error {
	_, dy := ebiten.Wheel()
	if dy != 0. && len(g.tileMap.Tiles) > selectedTile && selectedTile >= 0{
		if dy < 0 {
			g.tileMap.Tiles[selectedTile]--
		} else {
			g.tileMap.Tiles[selectedTile]++
		}
	}

	//TODO Remove code dupe
	if ebiten.IsMouseButtonPressed(ebiten.MouseButton(0)) {
		cx, cy := ebiten.CursorPosition();
		cx += int(g.rend.Cam.X)
		cy += int(g.rend.Cam.Y)
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
		cx += int(g.rend.Cam.X)
		cy += int(g.rend.Cam.Y)
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

	if ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyK) ||
		ebiten.IsKeyPressed(ebiten.KeyW) {
		g.player.TryStep(Up, g)
	} else if ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyJ) ||
		ebiten.IsKeyPressed(ebiten.KeyS) {
		g.player.TryStep(Down, g)
	} else if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyL) ||
		ebiten.IsKeyPressed(ebiten.KeyD) {
		g.player.TryStep(Right, g)
	} else if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyH) ||
		ebiten.IsKeyPressed(ebiten.KeyA) {
		g.player.TryStep(Left, g)
	} else {
		g.player.TryStep(Static, g)
	}

	g.player.Update()

	ticks++

	//if ticks % 1 == 0 {	// Maybe unnecessary?
	if g.client.active {
		g.client.WritePlayer(&g.player)
	}
	//}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.DrawTileset()
	g.DrawPlayer(&g.player)

	if g.client.active {
		g.client.playerMap.mutex.Lock()
		for _, player := range g.client.playerMap.players {
			g.DrawPlayer(&player)
		}
		g.client.playerMap.mutex.Unlock()
	}

	g.rend.Draw(&RenderTarget{
		&ebiten.DrawImageOptions{},
		selection,
		nil,
		selectionX,
		selectionY,
		100,
	})

	g.CenterRendererOnPlayer()
	g.rend.Display(screen)
	ebitenutil.DebugPrint(screen, fmt.Sprintf(
`camera.x: %f
camera.y: %f
player.x: %d
player.y: %d
player.Gx: %f
player.Gy: %f
player.isWalking: %t
player.id: %d`,
		g.rend.Cam.X, g.rend.Cam.Y, g.player.X, g.player.Y,
		g.player.Gx, g.player.Gy, g.player.isWalking, g.player.Id) )
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

func (g *Game) DrawPlayer(player *Player) {
	playerOpt := &ebiten.DrawImageOptions{}
	playerOpt.GeoM.Scale(2,2)

	x := player.Gx + playerOffsetX
	y := player.Gy + playerOffsetY

	playerRect := image.Rect(
		player.Tx,
		player.Ty,
		player.Tx + tileSize,
		player.Ty + tileSize,
	)

	g.rend.Draw(&RenderTarget{
		playerOpt,
		playerImg,
		&playerRect,
		x,
		y,
		1,
	})
}

func (g *Game) DrawTileset() {
	for i, n := range g.tileMap.Tiles {
		x := float64(i % g.tileMap.Width) * tileSize
		y := float64(i / g.tileMap.Width) * tileSize

		tx := (n % nTilesX) * tileSize
		ty := (n / nTilesX) * tileSize

		rect := image.Rect(tx, ty, tx + tileSize, ty + tileSize)
		g.rend.Draw(&RenderTarget{
			&ebiten.DrawImageOptions{},
			tileset,
			&rect,
			x,
			y,
			0,
		})

		if g.tileMap.Collision[i] {
			g.rend.Draw(&RenderTarget{
				&ebiten.DrawImageOptions{},
				collisionMarker,
				nil,
				x,
				y,
				100,
			})
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
	game.rend = NewRenderer(640, 480)
	game.Load("./resources/tilemaps/tilemap.json")
	game.player.X = 1
	game.player.Y = 1
	game.client = CreateClient()

	game.player.Id = game.client.Connect()
	if game.client.active {
		game.player.Connected = true
		go game.client.ReadPlayer()
	}

	defer game.client.Disconnect()
	defer game.Save()

	if err = ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
