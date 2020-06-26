package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/audio"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"io/ioutil"
	"image"
	"image/color"
	"log"
)

var tileset *ebiten.Image
var p1img []*ebiten.Image
var playerImg *ebiten.Image
var selection *ebiten.Image
var collisionMarker *ebiten.Image
var exitMarker *ebiten.Image
var selectionX int
var selectionY int
var m2Pressed = false
var m3Pressed = false
var copyBuffer = 0
var audioContext *audio.Context

const (
	tileSize = 32
	nTilesX = 8
	TileMapDir =  "./resources/tilemaps/"
)

var isServing = false
var buildPath = ""
var buildW = 0
var buildH = 0
var selectedTile = 0
var ticks = 0

type Exit struct {
	Target string
	Id int
	X int
	Y int
}

type Entry struct {
	Id int
	X int
	Y int
}

type TileMap struct {
	Tiles []int
	Collision []bool
	Exits []Exit
	Entries []Entry
	Width int
	Height int
}

type Game struct{
	tileMap TileMap
	player Player
	client Client
	rend Renderer
	audio Audio
}

func init() {
	flag.BoolVar(&isServing, "serve", false, "Run as game server")
	flag.StringVar(&buildPath, "build", "", "Generates a blank JSON map of dimension NxM")
	flag.IntVar(&buildW, "width", 10, "Desired width of JSON build")
	flag.IntVar(&buildH, "height", 10, "Desired height of JSON build")
	flag.Parse()
	if len(buildPath) > 0 {
		return
	}
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

	exitMarker, err = ebiten.NewImage(32, 32, ebiten.FilterDefault)
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

	exitClr := color.RGBA{0, 0, 255, 255}

	for p:= 0; p < 4; p++ {
		for q := 0; q < 4; q++ {
			exitMarker.Set(p + 14, q, exitClr)
		}
	}
}

func (t *TileMap) HasExitAt(x, y int) int {
	for i := range t.Exits {
		if t.Exits[i].X == x && t.Exits[i].Y == y {
			return i
		}
	}
	return -1
}

func (g *Game) TileIsOccupied(x int, y int) bool {
	if x < 0 || x >= g.tileMap.Width || y < 0 ||  y >= g.tileMap.Height {
		return true
	}

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

func (g *Game) CenterRendererOnPlayer() {
	g.rend.LookAt(
		g.player.Gx - 320 / 2 + tileSize / 2,
		g.player.Gy - 240 / 2 + tileSize / 2,
	)
}

func (g *Game) SelectTileFromMouse(cx, cy int) {
	cx += int(g.rend.Cam.X)
	cy += int(g.rend.Cam.Y)
	cx -= cx % tileSize
	cy -= cy % tileSize
	selectionX = cx / tileSize
	selectionY = cy / tileSize
	selectedTile =  selectionX + selectionY * g.tileMap.Width
}

func (g *Game) Update(screen *ebiten.Image) error {
	_, dy := ebiten.Wheel()
	if dy != 0. && len(g.tileMap.Tiles) > selectedTile && selectedTile >= 0{
		if dy < 0 {
			g.tileMap.Tiles[selectedTile]--
		} else {
			g.tileMap.Tiles[selectedTile]++
		}
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButton(0)) {
		cx, cy := ebiten.CursorPosition();
		g.SelectTileFromMouse(cx, cy)
	}

	if !m2Pressed && ebiten.IsMouseButtonPressed(ebiten.MouseButton(1)) {
		m2Pressed = true
		cx, cy := ebiten.CursorPosition();
		g.SelectTileFromMouse(cx, cy)
		if 0 <= selectedTile && selectedTile < len(g.tileMap.Tiles) {
			g.tileMap.Collision[selectedTile] = !g.tileMap.Collision[selectedTile]
		}
	} else if !ebiten.IsMouseButtonPressed(ebiten.MouseButton(1)) {
		m2Pressed = false
	}

	if !m3Pressed && ebiten.IsMouseButtonPressed(ebiten.MouseButton(2)) {
		m3Pressed = true
		cx, cy := ebiten.CursorPosition();
		g.SelectTileFromMouse(cx, cy)
		if 0 <= selectedTile && selectedTile < len(g.tileMap.Tiles) {
			if i := g.tileMap.HasExitAt(selectionX, selectionY); i != -1 {
				g.tileMap.Exits[i] = g.tileMap.Exits[len(g.tileMap.Exits) - 1]
				g.tileMap.Exits = g.tileMap.Exits[:len(g.tileMap.Exits) - 1]
			} else {
				g.tileMap.Exits = append(g.tileMap.Exits, Exit{
					"",
					0,
					selectionX,
					selectionY,
				})
			}
		}
	} else if !ebiten.IsMouseButtonPressed(ebiten.MouseButton(2)) {
		m3Pressed = false
	}

	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return errors.New("")	//TODO Gotta be a better way to do this
	}

	if !g.player.isWalking && ebiten.IsKeyPressed(ebiten.KeyShift) {
		g.player.isRunning = true
	} else {
		g.player.isRunning = false
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

	if ebiten.IsKeyPressed(ebiten.KeyC) {
		if 0 <= selectedTile && selectedTile < len(g.tileMap.Tiles) {
			copyBuffer = g.tileMap.Tiles[selectedTile]
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyV) {
		if 0 <= selectedTile && selectedTile < len(g.tileMap.Tiles) {
			g.tileMap.Tiles[selectedTile] = copyBuffer
		}
	}

	g.player.Update(g)

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
			if player.Location == g.player.Location {
				g.DrawPlayer(&player)
			}
		}
		g.client.playerMap.mutex.Unlock()
	}

	g.rend.Draw(&RenderTarget{
		&ebiten.DrawImageOptions{},
		selection,
		nil,
		float64(selectionX * tileSize),
		float64(selectionY * tileSize),
		100,
	})

	g.CenterRendererOnPlayer()
	g.rend.Display(screen)
	ebitenutil.DebugPrint(screen, fmt.Sprintf(
`camera.x: %f
camera.y: %f
player.x: %d
player.y: %d
player.id: %d
player.isRunning: %t`,
		g.rend.Cam.X, g.rend.Cam.Y, g.player.X, g.player.Y,
		g.player.Id, g.player.isRunning) )
}

func (g *Game) Load(str string, entrypoint int) {
	data, err := ioutil.ReadFile(str)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, &g.tileMap)
	if err != nil {
		panic(err)
	}
	g.player.Location = str
	g.player.X = g.tileMap.Entries[entrypoint].X
	g.player.Y = g.tileMap.Entries[entrypoint].Y
	g.player.Gx = float64(g.player.X * tileSize)
	g.player.Gy = float64(g.player.Y * tileSize)
	g.rend = NewRenderer(g.tileMap.Width * tileSize,
		g.tileMap.Height * tileSize,
		320,
		240,
	)
}

func (g *Game) Save() {
	bytes, err := json.Marshal(g.tileMap)
	if err != nil {
		fmt.Println(err)
	}
	ioutil.WriteFile(g.player.Location, bytes, 0644)
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

	for i := range g.tileMap.Exits {
		g.rend.Draw(&RenderTarget{
			&ebiten.DrawImageOptions{},
			exitMarker,
			nil,
			float64(g.tileMap.Exits[i].X * tileSize),
			float64(g.tileMap.Exits[i].Y * tileSize),
			100,
		})
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 240
	//return 640, 480
}

func build() {
	tiles := TileMap{
		make([]int, buildW * buildH),
		make([]bool, buildW * buildH),
		make([]Exit, 0),
		make([]Entry, 0),
		buildW,
		buildH,
	}

	fmt.Println("Wrote", buildW, "*", buildH, "=", buildW * buildH, "tileset")

	bytes, _ := json.Marshal(tiles)
	ioutil.WriteFile(buildPath, bytes, 0644)
}

func main() {
	if isServing {
		return
	}

	if len(buildPath) > 0 {
		build()
		return
	}

	var err error

	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Title")
	ebiten.SetWindowResizable(true)

	game := &Game{}

	game.Load(TileMapDir + "old.json", 0)
	game.client = CreateClient()
	game.audio = NewAudio()

	game.audio.audioPlayer.Play()

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
