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

// TODO: This function should only take one argument
func (t *TileMap) Draw(rend *Renderer, tileset *ebiten.Image) {
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

func (t *TileMap) SaveToFile(path string) error {
	data, err := json.Marshal(t)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0644)
}

func (t *TileMap) Resize(dx, dy int) {
	if (t.Width == 1 && dx < 0) || (t.Height == 1 && dy < 1) {
		return
	}

	invalidate := func(x, y int) {
		j := y * t.Width + x
		for i := range t.Tiles {
			copy(t.Tiles[i][j:], t.Tiles[i][j + 1:])
			t.Tiles[i][len(t.Tiles[i]) - 1] = 0
			t.Tiles[i] = t.Tiles[i][:len(t.Tiles[i]) - 1]

			copy(t.Collision[i][j:], t.Collision[i][j + 1:])
			t.Collision[i][len(t.Collision[i]) - 1] = false
			t.Collision[i] = t.Collision[i][:len(t.Collision[i]) - 1]

			copy(t.TextureIndicies[i][j:], t.TextureIndicies[i][j + 1:])
			t.TextureIndicies[i][len(t.TextureIndicies[i]) - 1] = 0
			t.TextureIndicies[i] = t.TextureIndicies[i][:len(t.TextureIndicies[i]) - 1]
		}
	}

	newWidth := t.Width + dx
	newHeight := t.Height + dy

	maxx := newWidth
	maxy := newHeight
	minx := t.Width
	miny := t.Height

	if maxx < minx {
		maxx, minx = minx, maxx
	}

	if maxy < miny {
		maxy, miny = miny, maxy
	}

	for j := maxy - 1; j >= miny; j-- {
		for i := maxx - 1; i >= minx; i-- {
			invalidate(i, j)
		}
	}

	t.Width = newWidth
	t.Height = newHeight

	for i, exit := range t.Exits {
		if exit.X >= t.Width || exit.Y >= t.Height {
			t.Exits[i] = t.Exits[len(t.Exits) - 1]
			t.Exits = t.Exits[:len(t.Exits) - 1]
		}
	}

	for i, entry := range t.Entries {
		if entry.X >= t.Width || entry.Y >= t.Height {
			t.Entries[i] = t.Entries[len(t.Entries) - 1]
			t.Entries = t.Entries[:len(t.Entries) - 1]
		}
	}
}

func CreateTileMap(width int, height int) TileMap {
	tex := make([][]int, 1)
	tex[0] = make([]int, width * height)

	col := make([][]bool, 1)
	col[0] = make([]bool, width * height)

	tiles := TileMap{
		tex,
		col,
		tex,
		make([]string, 0),
		make([]Exit, 0),
		make([]Entry, 0),
		width,
		height,
	}
	return tiles
}
