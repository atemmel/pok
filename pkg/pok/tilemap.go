package pok

import(
	"encoding/json"
	"github.com/hajimehoshi/ebiten"
	"io/ioutil"
	"image"
)

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

func (t *TileMap) Draw(rend *Renderer) {
	for j := range t.Tiles {
		if drawOnlyCurrentLayer && j != currentLayer {
			continue
		}
		for i, n := range t.Tiles[j] {
			x := float64(i % t.Width) * TileSize
			y := float64(i / t.Width) * TileSize

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
		}
	}
}

func (t *TileMap) OpenFile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, t)
	if err != nil {
		return err
	}
	return nil
}
