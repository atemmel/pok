package pok

import (
	"errors"
	"fmt"
	"github.com/atemmel/pok/pkg/constants"
	"github.com/atemmel/pok/pkg/dialog"
	"github.com/atemmel/pok/pkg/textures"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

var playerImg *ebiten.Image
var playerRunningImg *ebiten.Image
var activePlayerImg *ebiten.Image
var beachSplashImg *ebiten.Image

const waterSplashOffsetY = 13
const waterSplashOffsetX = 4

const nWaterSplashFrames = 3

type GameState interface {
	GetInputs(g *Game) error
	Update(g *Game) error
	Draw(g *Game, screen *ebiten.Image)
}

type OverworldState struct {
	tileMap TileMap
	collector dialog.DialogTreeCollector
}

func gamepadUp() bool {
	return ebiten.GamepadAxis(0, 1) < -0.1 || ebiten.IsGamepadButtonPressed(0, ebiten.GamepadButton11)
}

func gamepadDown() bool {
	return ebiten.GamepadAxis(0, 1) > 0.1 || ebiten.IsGamepadButtonPressed(0, ebiten.GamepadButton13)
}

func gamepadLeft() bool {
	return ebiten.GamepadAxis(0, 0) < -0.1 || ebiten.IsGamepadButtonPressed(0, ebiten.GamepadButton14)
}

func gamepadRight() bool {
	return ebiten.GamepadAxis(0, 0) > 0.1 || ebiten.IsGamepadButtonPressed(0, ebiten.GamepadButton12)
}

func movingUp() bool {
	return ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyK) || ebiten.IsKeyPressed(ebiten.KeyW) || gamepadUp()
}

func movingDown() bool {
	return ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyJ) || ebiten.IsKeyPressed(ebiten.KeyS) || gamepadDown()
}

func movingLeft() bool {
	return ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyH) || ebiten.IsKeyPressed(ebiten.KeyA) || gamepadLeft()
}

func movingRight() bool {
	return ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyL) || ebiten.IsKeyPressed(ebiten.KeyD) || gamepadRight()
}

func holdingSprint() bool {
	return ebiten.IsKeyPressed(ebiten.KeyShift) || ebiten.IsGamepadButtonPressed(0, ebiten.GamepadButton1)
}

func pressedInteract() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyZ) || inpututil.IsKeyJustPressed(ebiten.KeyE)
}

func (o *OverworldState) tryInteract(g *Game) {
	if g.Player.Char.isWalking || g.Player.Char.isRunning {
		return
	}

	x, y := g.Player.Char.X, g.Player.Char.Y

	switch g.Player.Char.dir {
		case Up:
			y--
		case Down:
			y++
		case Left:
			x--
		case Right:
			x++
	}

	// check npcs
	for i := range o.tileMap.npcs {
		npc := &(o.tileMap.npcs[i].Char)
		if npc.X == x && npc.Y == y {
			o.talkWith(g, i)
			break
		}
	}

	// check water
	n := y * o.tileMap.Width + x
	index := o.tileMap.textureMapping[o.tileMap.TextureIndicies[g.Player.Char.Z][n]]
	if textures.IsWater(index) {
		g.Dialog.SetString("This is water :)))")
		g.Dialog.Hidden = false
	}
}

func (o *OverworldState) talkWith(g *Game, npcIndex int) {
	char := &(o.tileMap.npcs[npcIndex].Char)
	dx, dy := g.Player.Char.X - char.X, g.Player.Char.Y - char.Y
	dir := Static

	if dx == 1 {
		dir = Right
	} else if dx == -1 {
		dir = Left
	} else if dy == 1 {
		dir = Down
	} else if dy == -1 {
		dir = Up
	}

	char.SetDirection(dir)
	tree := o.tileMap.npcs[npcIndex].Dialog
	o.collector = dialog.MakeDialogTreeCollector(tree)
	result := o.collector.CollectOnce()
	if result != nil {
		g.Dialog.SetString(result.Dialog)
	} else {
		g.Dialog.SetString("Result was nil and shouldn't be >:(")
	}
	g.Dialog.Hidden = false
}

func (o *OverworldState) GetInputs(g *Game) error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return errors.New("")	//TODO Gotta be a better way to do this
	}

	if g.Dialog.Hidden {
		o.CheckMovementInputs(g)
	} else {
		o.CheckDialogInputs(g)
	}

	if ebiten.IsKeyPressed(ebiten.Key1) {
		g.Rend.Cam.Scale += 0.1
	} else if ebiten.IsKeyPressed(ebiten.Key2) {
		g.Rend.Cam.Scale -= 0.1
	}

	return nil
}

func (o *OverworldState) CheckMovementInputs(g *Game) {
	if !g.Player.Char.isWalking && holdingSprint() {
		g.Player.Char.isRunning = true
	} else {
		g.Player.Char.isRunning = false
	}

	if movingUp() {
		g.Player.Char.TryStep(Up, g)
	} else if movingDown() {
		g.Player.Char.TryStep(Down, g)
	} else if movingRight() {
		g.Player.Char.TryStep(Right, g)
	} else if movingLeft() {
		g.Player.Char.TryStep(Left, g)
	} else {
		g.Player.Char.TryStep(Static, g)
	}

	if pressedInteract() {
		o.tryInteract(g)
	}

}

func (o *OverworldState) CheckDialogInputs(g *Game) {
	g.Player.Char.TryStep(Static, g)
	if g.Dialog.IsDone() {
		if pressedInteract() {
			result := o.collector.CollectOnce()
			if result == nil {
				g.Dialog.Hidden = true
			} else {
				g.Dialog.SetString(result.Dialog)
			}
		}
	}
}

func (o *OverworldState) Update(g *Game) error {
	g.Player.Update(g)
	o.tileMap.Update()
	o.tileMap.UpdateNpcs(g)

	if g.Client.Active {
		g.Client.WritePlayer(&g.Player)
	}

	g.Dialog.Update()

	return nil
}

func (o *OverworldState) Draw(g *Game, screen *ebiten.Image) {
	o.tileMap.Draw(&g.Rend)
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
`player.x: %f
player.y: %f
player.id: %d
currentLayer: %d
drawOnlyCurrentLayer: %t
selectedTexture: %d`,
			g.Player.Char.Gx, g.Player.Char.Gy, g.Player.Id, currentLayer,
			drawOnlyCurrentLayer, o.tileMap.Tiles[currentLayer][selectedTile]) )
	}

	g.Dialog.Draw(screen)
}

//TODO: Remove usage of DisplaySizex, DisplaySizeY
func (g *Game) CenterRendererOnPlayer() {
	g.Rend.LookAt(
		g.Player.Char.Gx - constants.DisplaySizeX / 4 + constants.TileSize / 2,
		g.Player.Char.Gy - constants.DisplaySizeY / 4 + constants.TileSize / 2,
	)
}
