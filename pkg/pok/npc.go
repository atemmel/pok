package pok

import(
	"github.com/atemmel/pok/pkg/dialog"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

//TODO: Expand functionality later
type Npc struct {
	Char Character
	Dialog *dialog.DialogTree
	NpcTextureIndex int
	MovementInfo NpcMovementInfo
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
	rewindDirection bool
}

const(
	NpcOffsetX = -8
	NpcOffsetY = -14
)

func BuildNpcFromNpcInfo(t *TileMap, info *NpcInfo) Npc {
	tree, err := dialog.ReadDialogTreeFromFile(DialogDir + info.DialogPath)
	Assert(err)
	npc := Npc{
		Character{},
		tree,
		-1,
		info.MovementInfo,
	}

	npc.Char.Gx = float64(info.X) * TileSize
	npc.Char.Gy = float64(info.Y) * TileSize

	npc.Char.X = info.X
	npc.Char.Y = info.Y

	for i, s := range t.npcImagesStrings {
		if info.Texture == s {
			npc.NpcTextureIndex = i
			break
		}
	}

	if npc.NpcTextureIndex == -1 {
		texture, _, err := ebitenutil.NewImageFromFile(CharacterImagesDir + info.Texture, ebiten.FilterDefault)

		Assert(err)

		npc.NpcTextureIndex = len(t.npcImages)
		t.npcImages = append(t.npcImages, texture)
		t.npcImagesStrings = append(t.npcImagesStrings, info.Texture)
	}

	return npc
}

func (npc* Npc) Update(g *Game) {
	switch npc.MovementInfo.Strategy {
		case Stay:
			return
		case Loop:
			npc.doLoopStrategy(g)
		case Rewind:
			npc.doRewindStrategy(g)
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
