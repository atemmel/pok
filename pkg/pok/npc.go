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
}

type NpcInfo struct {
	Texture string
	DialogPath string
	X, Y, Z int
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
