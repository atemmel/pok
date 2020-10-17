package pok

import(
	"encoding/json"
	"fmt"
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

func (t *TileMap) Resize(dx, dy, origin int) {
	if (t.Width == 1 && dx < 0) || (t.Height == 1 && dy < 0) || origin == -1 {
		return
	}

	/*
	invalidate := func(x, y int) {
		j := y * t.Width + x
		for i := range t.Tiles {
			copy(t.Tiles[i][j:], t.Tiles[i][j + 1:])
			t.Tiles[i][len(t.Tiles[i]) - 1] = 10
			t.Tiles[i] = t.Tiles[i][:len(t.Tiles[i]) - 1]

			copy(t.Collision[i][j:], t.Collision[i][j + 1:])
			t.Collision[i][len(t.Collision[i]) - 1] = false
			t.Collision[i] = t.Collision[i][:len(t.Collision[i]) - 1]

			copy(t.TextureIndicies[i][j:], t.TextureIndicies[i][j + 1:])
			t.TextureIndicies[i][len(t.TextureIndicies[i]) - 1] = 0
			t.TextureIndicies[i] = t.TextureIndicies[i][:len(t.TextureIndicies[i]) - 1]
		}
	}
	*/

	insertCol := func(x int) {
		for i := range t.Tiles {
			for j := 0; j < t.Height; j++ {
				index := j * t.Width + x
				t.Tiles[i] = append(t.Tiles[i][:index], 0)
				t.Collision[i] = append(t.Collision[i][:index], false)
				t.TextureIndicies[i] = append(t.TextureIndicies[i][:index], 0)
				fmt.Println("Appending", x, j)
			}
		}
	}

	insertRow := func(y int) {
		for i := range t.Tiles {
			for j := 0; j < t.Width; j++ {
				index := y * t.Width + j
				t.Tiles[i] = append(t.Tiles[i][:index], 0)
				t.Collision[i] = append(t.Collision[i][:index], false)
				t.TextureIndicies[i] = append(t.TextureIndicies[i][:index], 0)
				fmt.Println("Appending", j, y)
			}
		}
	}

	/*
	eraseCol := func(x int) {

	}
	*/

	newWidth := t.Width + dx
	newHeight := t.Height + dy

	if origin == TopLeftCorner || origin == BotLeftCorner {
		if dx < 0 {	// Crop left side

		} else if dx > 0 { // Grow left side
			insertCol(0)
		}
	} else if origin == TopRightCorner || origin == TopLeftCorner {
		if dx < 0 { // Crop right side 

		} else if dx > 0 { // Grow right side 
			insertCol(t.Width - 1)
		}
	}

	t.Width = newWidth

	if origin == TopLeftCorner || origin == TopRightCorner {
		if dy < 0 {	// Crop top side

		} else if dy > 0 { // Grow top side
			insertRow(0)
		}
	} else if origin == BotLeftCorner || origin == BotRightCorner {
		if dy < 0 { // Crop bot side 

		} else if dy > 0 { // Grow bot side 
			insertRow(t.Height - 1)
		}
	}

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
