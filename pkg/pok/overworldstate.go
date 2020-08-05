package pok

import (
	"errors"
	"fmt"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

var playerImg *ebiten.Image

type GameState interface {
	GetInputs(g *Game) error
	Update(g *Game) error
	Draw(g *Game, screen *ebiten.Image)
}

type OverworldState struct {
	tileMap TileMap
	tileset *ebiten.Image
}

func (o *OverworldState) GetInputs(g *Game) error {
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
	o.tileMap.Draw(&g.Rend, o.tileset)
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

func (g *Game) CenterRendererOnPlayer() {
	g.Rend.LookAt(
		g.Player.Gx - DisplaySizeX / 2 + TileSize / 2,
		g.Player.Gy - DisplaySizeY / 2 + TileSize / 2,
	)
}
