package pok

import(
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

//TODO: Expand functionality later
type Npc struct {
	Char Character
	Dialogue string
	NpcTextureIndex int
}

type NpcInfo struct {
	Texture string
	Dialogue string
	X, Y int

}

func BuildNpcFromNpcInfo(t *TileMap, info *NpcInfo) Npc {
	npc := Npc{
		Character{},
		info.Dialogue,
		-1,
	}

	for i, s := range t.npcImagesStrings {
		if info.Texture == s {
			npc.NpcTextureIndex = i
			break
		}
	}

	if npc.NpcTextureIndex == -1 {
		texture, _, err := ebitenutil.NewImageFromFile(info.Texture, ebiten.FilterDefault)

		if err != nil {
			panic(err)
		}

		npc.NpcTextureIndex = len(t.npcImages)
		t.npcImages = append(t.npcImages, texture)
		t.npcImagesStrings = append(t.npcImagesStrings, info.Texture)
	}

	return npc
}
