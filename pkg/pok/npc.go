package pok

import(
	"errors"
	"github.com/atemmel/pok/pkg/constants"
	"github.com/atemmel/pok/pkg/debug"
	"github.com/atemmel/pok/pkg/dialog"
	"github.com/atemmel/pok/pkg/textures"
	"image"
	"math/rand"
)

type Npc struct {
	Char Character
	Dialog *dialog.DialogTree
	NpcTextureIndex int
	MovementInfo NpcMovementInfo
	TalkedTo bool
}

type NpcInfo struct {
	Texture string
	DialogPath string
	X, Y, Z int
	MovementInfo NpcMovementInfo
}

type NpcMovementStrategy int

const(
	Stay NpcMovementStrategy = iota
	Loop
	Rewind
	Zone
)

type NpcMovementInfo struct {
	Strategy NpcMovementStrategy
	Commands []int
	currentIndex int
	zoneFramesUntilNextStep int
	rewindDirection bool
}

func SelectFramesUntilNextStep() int {
	const Min = 60
	const Max = 60 * 8
	return rand.Intn(Max - Min) + Min
}

const(
	NpcOffsetX = -8
	NpcOffsetY = -14
)

func BuildNpcFromNpcInfo(t *TileMap, info *NpcInfo) Npc {
	tree, err := dialog.ReadDialogTreeFromFile(constants.DialogDir + info.DialogPath)
	debug.Assert(err)

	if info.MovementInfo.Strategy == Zone {
		info.MovementInfo.zoneFramesUntilNextStep = SelectFramesUntilNextStep()
	}

	npc := Npc{
		Character{},
		tree,
		-1,
		info.MovementInfo,
		false,
	}

	npc.Char.Gx = float64(info.X) * constants.TileSize
	npc.Char.Gy = float64(info.Y) * constants.TileSize

	npc.Char.X = info.X
	npc.Char.Y = info.Y

	_, npc.NpcTextureIndex = textures.Load(constants.CharacterImagesDir + info.Texture)

	return npc
}

func (npc* Npc) Update(g *Game) {
	if npc.TalkedTo {
		npc.TalkedTo = !g.Dialog.Hidden
		npc.Char.SetDirection(Static)
		return
	}

	switch npc.MovementInfo.Strategy {
		case Stay:
			return
		case Loop:
			npc.doLoopStrategy(g)
		case Rewind:
			npc.doRewindStrategy(g)
		case Zone:
			npc.doZoneStrategy(g)
	}
}

func (npc *Npc) doLoopStrategy(g *Game) {
	currentIndex := &npc.MovementInfo.currentIndex
	currentDir := Direction(npc.MovementInfo.Commands[*currentIndex])
	npc.Char.TryStep(currentDir, g)
	result := npc.Char.Update(g)
	if result {
		npc.Char.isWalking = false
		*currentIndex++
		if *currentIndex >= len(npc.MovementInfo.Commands) {
			*currentIndex = 0
		}
	}
}

func (npc *Npc) doRewindStrategy(g *Game) {
	currentIndex := &npc.MovementInfo.currentIndex
	currentDir := Direction(npc.MovementInfo.Commands[*currentIndex])
	rewDir := &npc.MovementInfo.rewindDirection
	if *rewDir {
		currentDir = currentDir.Inverse()
	}
	npc.Char.TryStep(currentDir, g)
	result := npc.Char.Update(g)
	if result {
		npc.Char.isWalking = false
		// If not rewinding
		if !*rewDir {
			*currentIndex++
			if *currentIndex >= len(npc.MovementInfo.Commands) {
				*currentIndex--
				*rewDir = true
			}
		} else {
			*currentIndex--
			if *currentIndex < 0 {
				*currentIndex++
				*rewDir = false
			}
		}
	}
}

func (npc *Npc) doZoneStrategy(g *Game) {
	npc.Char.TryStep(npc.Char.dir, g)
	result := npc.Char.Update(g)

	if result {
		npc.Char.isWalking = false
		npc.Char.dir = Static
	}

	frames := &npc.MovementInfo.zoneFramesUntilNextStep
	*frames--
	if *frames <= 0 {
		if len(npc.MovementInfo.Commands) < 4 {
			debug.Assert(errors.New("Could not form rectangle from MovementInfo.Commands"))
		}

		x1 := npc.MovementInfo.Commands[0]
		y1 := npc.MovementInfo.Commands[1]
		x2 := npc.MovementInfo.Commands[2]
		y2 := npc.MovementInfo.Commands[3]

		ox, oy := npc.Char.X, npc.Char.Y
		rect := image.Rect(x1, y1, x2 + 1, y2 + 1)

		availableDirs := make([]Direction, 0, 4)

		for _, dir := range []Direction{Up, Down, Left, Right} {
			npc.Char.dir = dir
			npc.Char.UpdatePosition()
			nx, ny := npc.Char.X, npc.Char.Y
			pt := image.Point{nx, ny}
			npc.Char.X, npc.Char.Y = ox, oy

			if pt.In(rect) && !g.TileIsOccupied(nx, ny, npc.Char.Z) {
				availableDirs = append(availableDirs, dir)
			}
		}

		if len(availableDirs) > 0 {
			npc.Char.dir = availableDirs[rand.Intn(len(availableDirs))]
		} else {
			npc.Char.dir = Static
		}

		*frames = SelectFramesUntilNextStep()
	}
}
