package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"io/ioutil"
	"image"
	"image/color"
	"log"
)

const (
	tileSize = 32
	nTilesX = 8
	TileMapDir =  "./resources/tilemaps/"
)

var isServing = false
var buildPath = ""
var buildW = 0
var buildH = 0

type Game struct {
	ows OverworldState
	as GameState
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
	var err error
	playerImg, _, err = ebitenutil.NewImageFromFile("./resources/images/lucas.png", ebiten.FilterDefault)
	if err != nil {
		log.Fatal(err)
	}

	tileset, _, err = ebitenutil.NewImageFromFile("./resources/images/tileset1.png", ebiten.FilterDefault)
	if err != nil {
		log.Fatal(err)
	}

	selection, err = ebiten.NewImage(tileSize, tileSize, ebiten.FilterDefault)
	if err != nil {
		panic(err)
	}

	collisionMarker, err = ebiten.NewImage(tileSize, tileSize, ebiten.FilterDefault)
	if err != nil {
		panic(err)
	}

	exitMarker, err = ebiten.NewImage(tileSize, tileSize, ebiten.FilterDefault)
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

func (g *Game) TileIsOccupied(x int, y int, z int) bool {
	if x < 0 || x >= g.ows.tileMap.Width || y < 0 ||  y >= g.ows.tileMap.Height {
		return true
	}

	index := y * g.ows.tileMap.Width + x

	// Out of bounds check
	if z < 0 || z >= len(g.ows.tileMap.Tiles) {
		return true
	}

	if index >= len(g.ows.tileMap.Tiles[z]) || index < 0 {
		return true
	}

	if g.ows.tileMap.Collision[z][index] {
		return true
	}

	for _, p := range g.client.playerMap.players {
		if p.X == x && p.Y == y {
			return true
		}
	}

	return false
}

func (g *Game) Update(screen *ebiten.Image) error {
	err := g.as.GetInputs(g)
	if err != nil {
		return err
	}
	err = g.as.Update(g)
	if err != nil {
		return err
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.as.Draw(g, screen)
}

func (g *Game) Load(str string, entrypoint int) {
	data, err := ioutil.ReadFile(str)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, &g.ows.tileMap)
	if err != nil {
		panic(err)
	}
	g.player.Location = str
	g.player.X = g.ows.tileMap.Entries[entrypoint].X
	g.player.Y = g.ows.tileMap.Entries[entrypoint].Y
	g.player.Gx = float64(g.player.X * tileSize)
	g.player.Gy = float64(g.player.Y * tileSize)
	g.rend = NewRenderer(320,
		240,
		320,
		240,
	)
}

func (g *Game) Save() {
	bytes, err := json.Marshal(g.ows.tileMap)
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
		3,
	})
}

func (g *Game) DrawTileset() {
	for j := range g.ows.tileMap.Tiles {
		for i, n := range g.ows.tileMap.Tiles[j] {
			x := float64(i % g.ows.tileMap.Width) * tileSize
			y := float64(i / g.ows.tileMap.Width) * tileSize

			tx := (n % nTilesX) * tileSize
			ty := (n / nTilesX) * tileSize

			if tx < 0 || ty < 0 {
				continue
			}

			rect := image.Rect(tx, ty, tx + tileSize, ty + tileSize)
			g.rend.Draw(&RenderTarget{
				&ebiten.DrawImageOptions{},
				tileset,
				&rect,
				x,
				y,
				uint32(j * 2),
			})

			if g.ows.tileMap.Collision[j][i] {
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

	for i := range g.ows.tileMap.Exits {
		g.rend.Draw(&RenderTarget{
			&ebiten.DrawImageOptions{},
			exitMarker,
			nil,
			float64(g.ows.tileMap.Exits[i].X * tileSize),
			float64(g.ows.tileMap.Exits[i].Y * tileSize),
			100,
		})
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 240
	//return 640, 480
}

func build() {
	tex := make([][]int, 1)
	tex[0] = make([]int, buildW * buildH)

	col := make([][]bool, 1)
	col[0] = make([]bool, buildW * buildH)

	tiles := TileMap{
		tex,
		col,
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
	game.as = &game.ows

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
