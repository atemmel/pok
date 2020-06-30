package main

import (
	"errors"
	"fmt"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
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
var selectedTile = 0
var currentLayer = 0

var plusPressed = false
var minusPressed = false
var pPressed = false

type Exit struct {
	Target string
	Id int
	X int
	Y int
	Z int
}

type Entry struct {
	Id int
	X int
	Y int
	Z int
}

type TileMap struct {
	Tiles [][]int
	Collision [][]bool
	Exits []Exit
	Entries []Entry
	Width int
	Height int
}


type GameState interface {
	GetInputs(g *Game) error
	Update(g *Game) error
	Draw(g *Game, screen *ebiten.Image)
}

type OverworldState struct {
	tileMap TileMap
}

func (o *OverworldState) GetInputs(g *Game) error {
	_, dy := ebiten.Wheel()
	if dy != 0. && len(g.ows.tileMap.Tiles[currentLayer]) > selectedTile && selectedTile >= 0 {
		if dy < 0 {
			g.ows.tileMap.Tiles[currentLayer][selectedTile]--
		} else {
			g.ows.tileMap.Tiles[currentLayer][selectedTile]++
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
		if 0 <= selectedTile && selectedTile < len(g.ows.tileMap.Tiles[currentLayer]) {
			g.ows.tileMap.Collision[currentLayer][selectedTile] = !g.ows.tileMap.Collision[currentLayer][selectedTile]
		}
	} else if !ebiten.IsMouseButtonPressed(ebiten.MouseButton(1)) {
		m2Pressed = false
	}

	if !m3Pressed && ebiten.IsMouseButtonPressed(ebiten.MouseButton(2)) {
		m3Pressed = true
		cx, cy := ebiten.CursorPosition();
		g.SelectTileFromMouse(cx, cy)
		if 0 <= selectedTile && selectedTile < len(g.ows.tileMap.Tiles[currentLayer]) {
			if i := g.ows.tileMap.HasExitAt(selectionX, selectionY, currentLayer); i != -1 {
				g.ows.tileMap.Exits[i] = g.ows.tileMap.Exits[len(g.ows.tileMap.Exits) - 1]
				g.ows.tileMap.Exits = g.ows.tileMap.Exits[:len(g.ows.tileMap.Exits) - 1]
			} else {
				g.ows.tileMap.Exits = append(g.ows.tileMap.Exits, Exit{
					"",
					0,
					selectionX,
					selectionY,
					currentLayer,
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
		if 0 <= selectedTile && selectedTile < len(g.ows.tileMap.Tiles[currentLayer]) {
			copyBuffer = g.ows.tileMap.Tiles[currentLayer][selectedTile]
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyV) {
		if 0 <= selectedTile && selectedTile < len(g.ows.tileMap.Tiles[currentLayer]) {
			g.ows.tileMap.Tiles[currentLayer][selectedTile] = copyBuffer
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyMinus) && !plusPressed {	// Plus
		if currentLayer + 1 < len(o.tileMap.Tiles) {
			currentLayer++
		}
		plusPressed = true
	} else if !ebiten.IsKeyPressed(ebiten.KeyMinus) {
		plusPressed = false
	}

	if ebiten.IsKeyPressed(ebiten.KeySlash) && !minusPressed {	// Minus
		if currentLayer > 0 {
			currentLayer--
		}
		minusPressed = true
	} else if !ebiten.IsKeyPressed(ebiten.KeySlash) {
		minusPressed = false
	}

	if ebiten.IsKeyPressed(ebiten.KeyP) && !pPressed {	// Minus
		o.tileMap.Tiles = append(o.tileMap.Tiles, make([]int, len(o.tileMap.Tiles[0])))
		o.tileMap.Collision = append(o.tileMap.Collision, make([]bool, len(o.tileMap.Collision[0])))
		pPressed = true
	} else if !ebiten.IsKeyPressed(ebiten.KeyP) {
		pPressed = false
	}
	return nil
}

func (o *OverworldState) Update(g *Game) error {
	g.player.Update(g)

	if g.client.active {
		g.client.WritePlayer(&g.player)
	}

	return nil
}

func (o *OverworldState) Draw(g *Game, screen *ebiten.Image) {
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
`player.x: %d
player.y: %d
player.id: %d
currentLayer: %d`,
		g.player.X, g.player.Y, g.player.Id, currentLayer) )
}

func (g *Game) CenterRendererOnPlayer() {
	g.rend.LookAt(
		g.player.Gx - DisplaySizeX / 2 + tileSize / 2,
		g.player.Gy - DisplaySizeY / 2 + tileSize / 2,
	)
}

func (g *Game) SelectTileFromMouse(cx, cy int) {
	cx += int(g.rend.Cam.X)
	cy += int(g.rend.Cam.Y)
	cx -= cx % tileSize
	cy -= cy % tileSize
	selectionX = cx / tileSize
	selectionY = cy / tileSize
	selectedTile =  selectionX + selectionY * g.ows.tileMap.Width
}
