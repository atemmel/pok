package pok

import (
	"errors"
	"fmt"
	"github.com/atemmel/pok/pkg/constants"
	"github.com/atemmel/pok/pkg/dialog"
	"github.com/atemmel/pok/pkg/jobs"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

var playerImg *ebiten.Image
var playerRunningImg *ebiten.Image
var playerBikingImg *ebiten.Image
var playerSurfingImg *ebiten.Image
var playerUsingHMImg *ebiten.Image
var sharpedoImg *ebiten.Image
var beachSplashImg *ebiten.Image

var activePlayerImg *ebiten.Image

var selectedHm int = None

func aboutToUseHM() bool {
	return selectedHm != None
}

const (
	None = iota
	Surf
)

const waterSplashOffsetY = 13
const waterSplashOffsetX = 4

var waterSplashFrame int = 0
const nWaterSplashFrames = 3

func WaterSplashAnim() {
	waterSplashFrame++
	if waterSplashFrame >= nWaterSplashFrames {
		waterSplashFrame = 0
	}
}

type GameState interface {
	GetInputs(g *Game) error
	Update(g *Game) error
	Draw(g *Game, screen *ebiten.Image)
}

type OverworldState struct {
	tileMap TileMap
	collector dialog.DialogTreeCollector
	weather Weather
	//hailWeather HailWeather
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

func pressedItem() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyX)
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
	for i := range o.tileMap.Npcs {
		npc := &(o.tileMap.Npcs[i].Char)
		if npc.X == x && npc.Y == y {
			o.talkWith(g, i)
			break
		}
	}

	// check water
	if g.Player.Char.CoordinateContainsWater(x, y, g) {
		if !g.Player.Char.isSurfing {
			o.collector = dialog.MakeDialogTreeCollector(&dialog.DialogTree{
				&dialog.DialogNode{
					Dialog: "Sharpedo used Surf!",
					Next: dialog.Link(1),
				},
				&dialog.EffectDialogNode{
					Effect: "surf",
					Next: nil,
				},
			})

			result := o.collector.Peek()
			g.Dialog.SetString(result.Dialog)
			g.Dialog.Hidden = false
		}
	}

	// check rocks to smash
	rockIndex := o.tileMap.GetUnsmashedRockIndexAt(x, y, g.Player.Char.Z)
	if rockIndex != -1 {
		o.collector = dialog.MakeDialogTreeCollector(&dialog.DialogTree{
			&dialog.DialogNode{
				Dialog: "Bidoof used Rock Smash!",
				Next: dialog.Link(1),
			},
			&dialog.EffectDialogNode{
				Effect: "rocksmash",
				Next: nil,
			},
		})

		result := o.collector.Peek()
		g.Dialog.SetString(result.Dialog)
		g.Dialog.Hidden = false
	}

	// check trees to cut
	treeIndex := o.tileMap.GetUncutTreeIndexAt(x, y, g.Player.Char.Z)
	if treeIndex != -1 {
		o.tileMap.CuttableTrees[treeIndex].cut = true
	}
}

func (o *OverworldState) talkWith(g *Game, npcIndex int) {
	char := &(o.tileMap.Npcs[npcIndex].Char)
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
	o.tileMap.Npcs[npcIndex].TalkedTo = true
	tree := o.tileMap.Npcs[npcIndex].Dialog
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

	if pressedItem() {
		g.Player.Char.isBiking = !g.Player.Char.isBiking
	}
}

func (o *OverworldState) CheckDialogInputs(g *Game) {
	g.Player.Char.TryStep(Static, g)
	if g.Dialog.IsDone() {

		COLLECT_AGAIN:

		result := o.collector.Peek()
		if result == nil {
			g.Dialog.Hidden = true
			return
		}

		switch result.NodeId {
			case dialog.DialogNodeId:
				if pressedInteract() {
					_ = o.collector.CollectOnce()
					goto COLLECT_AGAIN
				}
				break
			case dialog.EffectDialogNodeId:
				if result.Opt == "surf" {
					beginSurf(g)
				} else if result.Opt == "rocksmash" {
					beginRockSmash(g)
				}
				_ = o.collector.CollectOnce();
				goto COLLECT_AGAIN
		}
	}
}

func beginSurf(g *Game) {
	nx, ny := g.Player.Char.X, g.Player.Char.Y

	switch g.Player.Char.dir {
	case Down:
		ny++
	case Right:
		nx++
	case Left:
		nx--
	case Up:
		ny--
	}

	g.Player.Char.X, g.Player.Char.Y = nx, ny

	g.Audio.PlayPlayerJump()
	g.Player.Char.isBiking = false
	g.Player.Char.isJumping = true
	g.Player.Char.isWalking = true
	g.Player.Char.velocity = WalkVelocity
	g.Player.Char.currentJumpTarget = constants.TileSize

	selectedHm = Surf
}

func beginRockSmash(g *Game) {
	x, y := g.Player.Char.X, g.Player.Char.Y

	switch g.Player.Char.dir {
		case Down:
			y++
		case Right:
			x++
		case Left:
			x--
		case Up:
			y--
	}

	z := g.Player.Char.Z

	rockIndex := g.Ows.tileMap.GetUnsmashedRockIndexAt(x, y, z)
	if rockIndex == -1 {
		return
	}

	g.Ows.tileMap.Rocks[rockIndex].smashed = true
}

func (o *OverworldState) Update(g *Game) error {
	g.Player.Update(g)
	jobs.TickAllOneFrame()
	o.tileMap.UpdateNpcs(g)
	if o.weather != nil {
		o.weather.Update()
	}

	if g.Client.Active {
		g.Client.WritePlayer(&g.Player)
	}

	g.Dialog.Update()

	return nil
}

func (o *OverworldState) Draw(g *Game, screen *ebiten.Image) {
	o.tileMap.Draw(&g.Rend, false, 0)
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

	if o.weather != nil {
		o.weather.Draw(&g.Rend)
	}

	g.CenterRendererOnPlayer()
	g.Rend.Display(screen)

	if DrawDebugInfo {
		x, y, z := g.Player.Char.X, g.Player.Char.Y, g.Player.Char.Z
		ebitenutil.DebugPrint(screen, fmt.Sprintf(
`player.x: %f
player.y: %f
player.z: %d
player.id: %d
isStaircaseRightNow: %t
cam.x: %f
cam.y: %f`,
			g.Player.Char.Gx, g.Player.Char.Gy, g.Player.Char.Z, g.Player.Id, g.Player.Char.isStairCase(x, y, z, g),
			g.Rend.Cam.X, g.Rend.Cam.Y) )
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
