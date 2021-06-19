package pok

import(
	"encoding/json"
	"github.com/atemmel/pok/pkg/constants"
	"github.com/atemmel/pok/pkg/textures"
	"github.com/hajimehoshi/ebiten/v2"
	"io/ioutil"
	"image"
	"strings"
)

var debugLoadingEnabled bool = true

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
	NpcInfo []NpcInfo

	textureMapping []int

	npcs []Npc
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

func (t *TileMap) HasTexture(index int) bool {
	for i := range t.textureMapping {
		if t.textureMapping[i] == index {
			return true
		}
	}
	return false
}

func (t *TileMap) AppendTexture(index int, str string) {
	t.textureMapping = append(t.textureMapping, index)
	t.Textures = append(t.Textures, str)
}

func (t *TileMap) MapReverse(j int) int {
	for i := range t.textureMapping {
		if t.textureMapping[i] == j {
			return i
		}
	}
	return -1
}

func (t *TileMap) Draw(rend *Renderer) {
	t.DrawWithOffset(rend, 0, 0)
}

func (t *TileMap) UpdateNpcs(g *Game) {
	for i := range t.npcs {
		t.npcs[i].Update(g)
	}
}

func (t *TileMap) drawNpcs(rend *Renderer, offsetX, offsetY float64) {
	for i := range t.npcs {
		index := t.npcs[i].NpcTextureIndex
		t.npcs[i].Char.Draw(textures.Access(index), rend, offsetX, offsetY)
	}
}

func (t *TileMap) DrawWithOffset(rend *Renderer, offsetX, offsetY float64) {
	for j := range t.Tiles {
		if drawOnlyCurrentLayer && j != currentLayer {
			continue
		}
		for i, n := range t.Tiles[j] {
			// Do not "draw" invisible sprites
			if t.Tiles[j][i] < 0 {
				continue;
			}

			ix, iy := t.Coords(i)
			x := float64(ix) * constants.TileSize
			y := float64(iy) * constants.TileSize

			img := textures.Access(t.textureMapping[t.TextureIndicies[j][i]])
			nTilesX := img.Bounds().Dx() / constants.TileSize

			tx := (n % nTilesX) * constants.TileSize
			ty := (n / nTilesX) * constants.TileSize

			if tx < 0 || ty < 0 {
				continue
			}

			opt := &ebiten.DrawImageOptions{}

			rect := image.Rect(tx, ty, tx + constants.TileSize, ty + constants.TileSize)
			rend.Draw(&RenderTarget{
				opt,
				img,
				&rect,
				x + offsetX,
				y + offsetY,
				j, //j * 2,
			})
		}
	}

	t.drawNpcs(rend, offsetX, offsetY)
}

func (t *TileMap) OpenFile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		if debugLoadingEnabled {
			// We might wish to develop, and if so, we do this
			if i := strings.LastIndex(path, "/"); i != -1 {
				dbgPath := path[i+1:]
				// Go assignment is acting up
				var secondErr error
				data, secondErr = ioutil.ReadFile(dbgPath)
				if secondErr != nil {
					return secondErr
				}
			} else {
				return err
			}
		} else {
			return err
		}
	}
	err = json.Unmarshal(data, t)
	if err != nil {
		return err
	}

	indicies := make([]int, len(t.Textures))

	for i := range indicies {
		_, index := textures.Load(constants.TileMapImagesDir + t.Textures[i])
		indicies[i] = index
	}

	t.textureMapping = indicies

	t.npcs = t.npcs[:0]
	err = t.createNpcs()

	return err
}

func (t *TileMap) Index(x, y int) int {
	return y * t.Width + x
}

func (t *TileMap) Coords(i int) (int, int) {
	return i % t.Width, i / t.Width
}

func (t *TileMap) SaveToFile(path string) error {
	data, err := json.Marshal(t)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0644)
}

func (t *TileMap) NTilesX(textureIndex int) int {
	img := textures.Access(textureIndex)
	w, _ := img.Size()
	return w / constants.TileSize
}

func (t *TileMap) InsertObject(obj *EditorObject, objIndex, i, z int, placedObjects *[]PlacedEditorObject) {
	col, row := t.Coords(i)

	existingObjectIndex := HasPlacedObjectAt(*placedObjects, col, row)
	if existingObjectIndex != -1 {
		t.EraseObject((*placedObjects)[existingObjectIndex], obj)

		// Erase from placedObjects
		(*placedObjects)[existingObjectIndex] = (*placedObjects)[len(*placedObjects) - 1]
		*placedObjects = (*placedObjects)[:len(*placedObjects) - 1]
	}

	// Get max depth
	maxZ := 0
	for _, z := range obj.Z {
		if z > maxZ {
			maxZ = z
		}
	}
	maxZ++

	// Append layers as necessary
	for maxZ > len(t.Tiles) {
		t.AppendLayer()
	}

	zIndex := 0

	for y := 0; y != obj.H; y++ {
		gy := row + y

		ty := (obj.Y + y) * t.NTilesX(obj.textureIndex)

		for x := 0; x != obj.W; x++ {
			gx := col + x

			if (gx < 0 || gx >= t.Width) || (gy < 0 || gy >= t.Height) {
				zIndex++
				continue
			}

			tx := obj.X + x

			tile := ty + tx
			index := t.Index(gx, gy)
			depth := z + obj.Z[zIndex]

			t.Tiles[depth][index] = tile
			t.TextureIndicies[depth][index] = obj.textureIndex

			if (y > 0 || obj.H == 1) && (x > 0 || obj.W == 1) {
				t.Collision[z][index] = true
			}

			zIndex++
		}
	}

	p := PlacedEditorObject{
		col, row, z,
		objIndex,
	}

	*placedObjects = append(*placedObjects, p)
}

func (t *TileMap) EraseObject(pob PlacedEditorObject, obj *EditorObject) {
	zIndex := 0

	for y := 0; y < obj.H; y++ {
		gy := pob.Y + y

		for x := 0; x < obj.W; x++ {
			gx := pob.X + x

			if (gx < 0 || gx >= t.Width) || (gy < 0 || gy >= t.Height) {
				zIndex++
				continue
			}

			index := t.Index(gx, gy)
			depth := pob.Z + obj.Z[zIndex]

			t.Tiles[depth][index] = -1
			t.TextureIndicies[depth][index] = 0

			if (y > 0 || obj.H == 1) && (x > 0 || obj.W == 1) {
				t.Collision[pob.Z][index] = false
			}

			zIndex++
		}
	}
}

func (t *TileMap) AppendLayer() {
	t.Tiles = append(t.Tiles, make([]int, len(t.Tiles[0])))
	for i := range t.Tiles[len(t.Tiles) - 1] {
		t.Tiles[len(t.Tiles)-1][i] = -1
	}
	t.Collision = append(t.Collision, make([]bool, len(t.Collision[0])))
	t.TextureIndicies = append(t.TextureIndicies, make([]int, len(t.TextureIndicies[0])))
	for i := range t.TextureIndicies[len(t.TextureIndicies) - 1] {
		t.TextureIndicies[len(t.TextureIndicies)-1][i] = 0
	}
}

func (t *TileMap) RemoveLayer(index int) {
	if len(t.Collision) == 1 {
		return
	} else if len(t.Collision) == index + 1 {
		t.Collision = t.Collision[:index]
		t.TextureIndicies = t.TextureIndicies[:index]
		t.Tiles = t.Tiles[:index]
		return
	}

	t.Collision = append(t.Collision[:index], t.Collision[index + 1:]...)
	t.TextureIndicies = append(t.TextureIndicies[:index], t.TextureIndicies[index + 1:]...)
	t.Tiles = append(t.Tiles[:index], t.Tiles[index + 1:]...)
}

func (t *TileMap) Resize(dx, dy, origin int) {
	if (t.Width == 1 && dx < 0) || (t.Height == 1 && dy < 0) || origin == -1 {
		return
	}

	ndx, ndy := dx, dy

	insertCol := func(x int) {
		if x == t.Width - 1 {
			x++
		}
		t.Width++
		for i := range t.Tiles {
			for j := 0; j < t.Height; j++ {
				index := j * t.Width + x
				t.Tiles[i] = append(t.Tiles[i], 0)
				t.Collision[i] = append(t.Collision[i], false)
				t.TextureIndicies[i] = append(t.TextureIndicies[i], 0)
				copy(t.Tiles[i][index + 1:], t.Tiles[i][index:])
				copy(t.Collision[i][index + 1:], t.Collision[i][index:])
				copy(t.TextureIndicies[i][index + 1:], t.TextureIndicies[i][index:])

				t.Tiles[i][index] = 0
				t.Collision[i][index] = false
				t.TextureIndicies[i][index] = 0
			}
		}
	}

	insertRow := func(y int) {
		if y == t.Height - 1 {
			y++
		}
		t.Height++
		for i := range t.Tiles {
			for j := 0; j < t.Width; j++ {
				index := y * t.Width + j
				t.Tiles[i] = append(t.Tiles[i], 0)
				t.Collision[i] = append(t.Collision[i], false)
				t.TextureIndicies[i] = append(t.TextureIndicies[i], 0)
				copy(t.Tiles[i][index + 1:], t.Tiles[i][index:])
				copy(t.Collision[i][index + 1:], t.Collision[i][index:])
				copy(t.TextureIndicies[i][index + 1:], t.TextureIndicies[i][index:])
				t.Tiles[i][index] = 0
				t.Collision[i][index] = false
				t.TextureIndicies[i][index] = 0
			}
		}
	}

	eraseCol := func(x int) {
		if t.Width <= 1 {
			return;
		}
		t.Width--
		for i := range t.Tiles {
			for j := 0; j < t.Height; j++ {
				index := j * t.Width + x
				copy(t.Tiles[i][index:], t.Tiles[i][index + 1:])
				t.Tiles[i] = t.Tiles[i][:len(t.Tiles[i]) - 1]
				copy(t.Collision[i][index:], t.Collision[i][index + 1:])
				t.Collision[i] = t.Collision[i][:len(t.Collision[i]) - 1]
				copy(t.TextureIndicies[i][index:], t.TextureIndicies[i][index + 1:])
				t.TextureIndicies[i] = t.TextureIndicies[i][:len(t.TextureIndicies[i]) - 1]
			}
		}
	}

	eraseRow := func(y int) {
		if t.Height <= 1 {
			return
		}
		t.Height--
		for i := range t.Tiles {
			start := y * t.Width
			end := y * t.Width + t.Width
			t.Tiles[i] = append(t.Tiles[i][:start], t.Tiles[i][end:]...)
			t.Collision[i] = append(t.Collision[i][:start], t.Collision[i][end:]...)
			t.TextureIndicies[i] = append(t.TextureIndicies[i][:start], t.TextureIndicies[i][end:]...)
		}
	}

	if origin == TopLeftCorner || origin == BotLeftCorner {
		if dx < 0 {	// Crop left side
			for ; dx < 0; dx++ {
				eraseCol(0)
			}
		} else if dx > 0 { // Grow left side
			for ; dx > 0; dx-- {
				insertCol(0)
			}
		}
	} else if origin == TopRightCorner || origin == BotRightCorner {
		if dx < 0 { // Crop right side 
			for ; dx < 0; dx++ {
				eraseCol(t.Width - 1)
			}
		} else if dx > 0 { // Grow right side 
			for ; dx > 0; dx-- {
				insertCol(t.Width - 1)
			}
		}
	}

	if origin == TopLeftCorner || origin == TopRightCorner {
		if dy < 0 {	// Crop top side
			for ; dy < 0; dy++ {
				eraseRow(0)
			}
		} else if dy > 0 { // Grow top side
			for ; dy > 0; dy-- {
				insertRow(0)
			}
		}
	} else if origin == BotLeftCorner || origin == BotRightCorner {
		if dy < 0 { // Crop bot side 
			for ; dy < 0; dy++ {
				eraseRow(t.Height - 1)
			}
		} else if dy > 0 { // Grow bot side 
			for ; dy > 0; dy-- {
				insertRow(t.Height - 1)
			}
		}
	}

	if origin == BotLeftCorner || origin == BotRightCorner {
		ndy = 0
	}

	if origin == BotRightCorner || origin == TopRightCorner {
		ndx = 0
	}

	t.moveNpcs(ndx, ndy)
}

func (t *TileMap) moveNpcs(dx, dy int) {
	for i := range t.npcs {
		nx, ny := t.npcs[i].Char.X + dx, t.npcs[i].Char.Y + dy

		if nx < 0 {
			dx += nx
		} else if nx >= t.Width {
			dx += t.Width - nx - 1
		}

		if ny < 0 {
			dy += ny
		} else if ny >= t.Height {
			dy += t.Height - ny - 1
		}

		t.npcs[i].Char.Gx += float64(dx * constants.TileSize)
		t.npcs[i].Char.Gy += float64(dy * constants.TileSize)

		t.npcs[i].Char.X += dx
		t.npcs[i].Char.Y += dy

		t.NpcInfo[i].X += dx
		t.NpcInfo[i].Y += dy
	}
}

func (t *TileMap) PlaceEntry(entry Entry) {
	t.Entries = append(t.Entries, entry)
}

func (t *TileMap) PlaceExit(exit Exit) {
	t.Exits = append(t.Exits, exit)
}

func (t *TileMap) PlaceNpc(ni *NpcInfo) {
	t.NpcInfo = append(t.NpcInfo, *ni)
	t.npcs = append(t.npcs, BuildNpcFromNpcInfo(t, ni))
}

func (t *TileMap) RemoveNpc(index int) {
	t.NpcInfo[index] = t.NpcInfo[len(t.NpcInfo)-1]
	t.npcs[index] = t.npcs[len(t.npcs)-1]
	t.NpcInfo = t.NpcInfo[:len(t.NpcInfo) - 1]
	t.npcs = t.npcs[:len(t.npcs) - 1]
}

func (t *TileMap) Contains(x, y int) bool {
	return x < t.Width && x >= 0 && y < t.Height && y >= 0
}

func (t *TileMap) FillCollision(x, y, w, h, z int) {
	for j:= y; j < y + h; j++ {
		for i := x; i < x + w; i++ {
			if t.Contains(i, j) {
				index := i + t.Width * j
				t.Collision[z][index] = true
			}
		}
	}
}

func CreateTileMap(width int, height int, texture []string) *TileMap {
	textureMapping := make([]int, len(texture))

	for i := range texture {
		_, index := textures.Load(texture[i])
		textureMapping[i] = index
	}

	tex := make([][]int, 1)
	tex[0] = make([]int, width * height)

	col := make([][]bool, 1)
	col[0] = make([]bool, width * height)

	ind := make([][]int, 1)
	ind[0] = make([]int, width * height)

	tiles := &TileMap{
		tex,
		col,
		ind,
		texture,
		make([]Exit, 0),
		make([]Entry, 0),
		width,
		height,
		make([]NpcInfo, 0),
		textureMapping,
		make([]Npc, 0),
	}
	return tiles
}

func (t *TileMap) createNpcs() error {

	for i := range t.NpcInfo {
		npc := BuildNpcFromNpcInfo(t, &t.NpcInfo[i])
		t.npcs = append(t.npcs, npc)
	}

	return nil
}
