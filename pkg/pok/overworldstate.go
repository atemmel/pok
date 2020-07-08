package pok

import (
	"errors"
	"fmt"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"image"
)

var tileset *ebiten.Image
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
var uPressed = false
var iPressed = false
var drawOnlyCurrentLayer = false
var drawUi = false

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
	TextureIndicies [][]int
	Textures []string
	Exits []Exit
	Entries []Entry
	Width int
	Height int
}

func (t *TileMap) HasExitAt(x, y, z int) int {
	for i := range t.Exits {
		if t.Exits[i].X == x && t.Exits[i].Y == y && t.Exits[i].Z == z {
			return i
		}
	}
	return -1
}

func (t *TileMap) GetEntryWithId(id int) int {
	for i := range t.Entries {
		if t.Entries[i].Id == id {
			return i
		}
	}
	return -1
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
	if dy != 0. && len(g.Ows.tileMap.Tiles[currentLayer]) > selectedTile && selectedTile >= 0 {
		if dy < 0 {
			g.Ows.tileMap.Tiles[currentLayer][selectedTile]--
		} else {
			g.Ows.tileMap.Tiles[currentLayer][selectedTile]++
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
		if 0 <= selectedTile && selectedTile < len(g.Ows.tileMap.Tiles[currentLayer]) {
			g.Ows.tileMap.Collision[currentLayer][selectedTile] = !g.Ows.tileMap.Collision[currentLayer][selectedTile]
		}
	} else if !ebiten.IsMouseButtonPressed(ebiten.MouseButton(1)) {
		m2Pressed = false
	}

	if !m3Pressed && ebiten.IsMouseButtonPressed(ebiten.MouseButton(2)) {
		m3Pressed = true
		cx, cy := ebiten.CursorPosition();
		g.SelectTileFromMouse(cx, cy)
		if 0 <= selectedTile && selectedTile < len(g.Ows.tileMap.Tiles[currentLayer]) {
			if i := g.Ows.tileMap.HasExitAt(selectionX, selectionY, currentLayer); i != -1 {
				g.Ows.tileMap.Exits[i] = g.Ows.tileMap.Exits[len(g.Ows.tileMap.Exits) - 1]
				g.Ows.tileMap.Exits = g.Ows.tileMap.Exits[:len(g.Ows.tileMap.Exits) - 1]
			} else {
				g.Ows.tileMap.Exits = append(g.Ows.tileMap.Exits, Exit{
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

	if !g.Player.isWalking && ebiten.IsKeyPressed(ebiten.KeyShift) {
		g.Player.isRunning = true
	} else {
		g.Player.isRunning = false
	}

	if ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyK) ||
		ebiten.IsKeyPressed(ebiten.KeyW) {
		g.Player.TryStep(Up, g)
	} else if ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyJ) ||
		ebiten.IsKeyPressed(ebiten.KeyS) {
		g.Player.TryStep(Down, g)
	} else if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyL) ||
		ebiten.IsKeyPressed(ebiten.KeyD) {
		g.Player.TryStep(Right, g)
	} else if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyH) ||
		ebiten.IsKeyPressed(ebiten.KeyA) {
		g.Player.TryStep(Left, g)
	} else {
		g.Player.TryStep(Static, g)
	}

	if ebiten.IsKeyPressed(ebiten.KeyC) {
		if 0 <= selectedTile && selectedTile < len(g.Ows.tileMap.Tiles[currentLayer]) {
			copyBuffer = g.Ows.tileMap.Tiles[currentLayer][selectedTile]
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyV) {
		if 0 <= selectedTile && selectedTile < len(g.Ows.tileMap.Tiles[currentLayer]) {
			g.Ows.tileMap.Tiles[currentLayer][selectedTile] = copyBuffer
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

	if ebiten.IsKeyPressed(ebiten.KeyU) && !uPressed {
		drawOnlyCurrentLayer = !drawOnlyCurrentLayer
		uPressed = true
	} else if !ebiten.IsKeyPressed(ebiten.KeyU) {
		uPressed = false
	}

	if ebiten.IsKeyPressed(ebiten.KeyI) && !iPressed {
		drawUi = !drawUi
		iPressed = true
	} else if !ebiten.IsKeyPressed(ebiten.KeyI) {
		iPressed = false
	}

	return nil
}

func (o *OverworldState) Update(g *Game) error {
	g.Player.Update(g)

	if g.Client.Active {
		g.Client.WritePlayer(&g.Player)
	}

	return nil
}

func (o *OverworldState) Draw(g *Game, screen *ebiten.Image) {
	o.DrawTileset(&g.Rend)
	g.DrawPlayer(&g.Player)

	if g.Client.Active {
		g.Client.playerMap.mutex.Lock()
		for _, player := range g.Client.playerMap.players {
			if player.Location == g.Player.Location {
				g.DrawPlayer(&player)
			}
		}
		g.Client.playerMap.mutex.Unlock()
	}

	if drawUi {
		g.Rend.Draw(&RenderTarget{
			&ebiten.DrawImageOptions{},
			selection,
			nil,
			float64(selectionX * TileSize),
			float64(selectionY * TileSize),
			100,
		})
	}

	g.CenterRendererOnPlayer()
	g.Rend.Display(screen)

	if drawUi {
		ebitenutil.DebugPrint(screen, fmt.Sprintf(
`player.x: %d
player.y: %d
player.id: %d
currentLayer: %d
drawOnlyCurrentLayer: %t
selectedTexture: %d`,
			g.Player.X, g.Player.Y, g.Player.Id, currentLayer,
			drawOnlyCurrentLayer, o.tileMap.Tiles[currentLayer][selectedTile]) )
	}
}

func (o *OverworldState) DrawTileset(rend *Renderer) {
	for j := range o.tileMap.Tiles {
		if drawOnlyCurrentLayer && j != currentLayer {
			continue
		}
		for i, n := range o.tileMap.Tiles[j] {
			x := float64(i % o.tileMap.Width) * TileSize
			y := float64(i / o.tileMap.Width) * TileSize

			tx := (n % NTilesX) * TileSize
			ty := (n / NTilesX) * TileSize

			if tx < 0 || ty < 0 {
				continue
			}

			rect := image.Rect(tx, ty, tx + TileSize, ty + TileSize)
			rend.Draw(&RenderTarget{
				&ebiten.DrawImageOptions{},
				tileset,
				&rect,
				x,
				y,
				uint32(j * 2),
			})

			if drawUi && currentLayer == j && o.tileMap.Collision[j][i] {
				rend.Draw(&RenderTarget{
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

	if drawUi {
		for i := range o.tileMap.Exits {
			rend.Draw(&RenderTarget{
				&ebiten.DrawImageOptions{},
				exitMarker,
				nil,
				float64(o.tileMap.Exits[i].X * TileSize),
				float64(o.tileMap.Exits[i].Y * TileSize),
				100,
			})
		}
	}
}

func (g *Game) CenterRendererOnPlayer() {
	g.Rend.LookAt(
		g.Player.Gx - DisplaySizeX / 2 + TileSize / 2,
		g.Player.Gy - DisplaySizeY / 2 + TileSize / 2,
	)
}

func (g *Game) SelectTileFromMouse(cx, cy int) {
	cx += int(g.Rend.Cam.X)
	cy += int(g.Rend.Cam.Y)
	cx -= cx % TileSize
	cy -= cy % TileSize
	selectionX = cx / TileSize
	selectionY = cy / TileSize
	selectedTile =  selectionX + selectionY * g.Ows.tileMap.Width
}
