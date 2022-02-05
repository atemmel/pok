package editor

import(
	"github.com/atemmel/pok/pkg/pok"
	"encoding/json"
	"io/ioutil"
	"image"
	"strings"
)

type EditorObject struct {
	Texture string
	X, Y int
	W, H int
	Z []int

	textureIndex int
}

func (e *EditorObject) FindAndSetCorrectTexture(textures []string) {
	e.textureIndex = 0
	for i := range textures {
		if textures[i] == e.Texture {
			e.textureIndex = i
		}
	}
}

type PlacedEditorObject struct {
	X, Y, Z int
	Index int
}

func ReadAllObjects(directory string) ([]EditorObject, error) {
	dirs, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	objs := make([]EditorObject, 0)
	for i := range dirs {
		if dirs[i].IsDir() || !strings.HasSuffix(dirs[i].Name(), ".edobj") {
			continue
		}

		bytes, err := ioutil.ReadFile(directory + dirs[i].Name())
		if err != nil {
			return nil, err
		}

		obj := EditorObject{}
		err = json.Unmarshal(bytes, &obj)
		if err != nil {
			return nil, err
		}

		objs = append(objs, obj)
	}

	return objs, nil
}

func HasPlacedObjectAt(pobs []PlacedEditorObject, edobjs []EditorObject, x, y int) int {
	prospect := image.Pt(x, y)
	for i := range pobs {
		obj := edobjs[pobs[i].Index]

		r := image.Rect(0, 0, obj.W, obj.H).Add(image.Pt(pobs[i].X, pobs[i].Y))
		if prospect.In(r) {
			return i
		}
	}
	return -1
}

//TODO: This looks gross, what is even going on inside here?
func (edobj* EditorObject) InsertObject(t *pok.TileMap, objIndex, i, z int, placedObjects *[]PlacedEditorObject, edobjs []EditorObject) {
	col, row := t.Coords(i)

	existingObjectIndex := HasPlacedObjectAt(*placedObjects, edobjs, col, row)
	if existingObjectIndex != -1 {
		//t.EraseObject((*placedObjects)[existingObjectIndex], edobj)
		edobj.EraseObject(t, (*placedObjects)[existingObjectIndex])

		// Erase from placedObjects
		(*placedObjects)[existingObjectIndex] = (*placedObjects)[len(*placedObjects) - 1]
		*placedObjects = (*placedObjects)[:len(*placedObjects) - 1]
	}

	// Get max depth
	maxZ := 0
	for _, z := range edobj.Z {
		if z > maxZ {
			maxZ = z
		}
	}
	maxZ++

	// Append layers as necessary
	for maxZ > len(t.Tiles) {
		t.AppendLayer()
	}

	t.MaybeAddTextureMapping(edobj.textureIndex, edobj.Texture)
	texIndex := t.GetTextureMapping(edobj.textureIndex)

	zIndex := 0

	for y := 0; y != edobj.H; y++ {
		gy := row + y

		ty := (edobj.Y + y) * t.NTilesX(edobj.textureIndex)

		for x := 0; x != edobj.W; x++ {
			gx := col + x

			if (gx < 0 || gx >= t.Width) || (gy < 0 || gy >= t.Height) {
				zIndex++
				continue
			}

			tx := edobj.X + x

			tile := ty + tx
			index := t.Index(gx, gy)
			depth := z + edobj.Z[zIndex]

			t.Tiles[depth][index] = tile
			t.TextureIndicies[depth][index] = texIndex

			if (y > 0 || edobj.H == 1) && (x > 0 || edobj.W == 1) {
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

func (edobj *EditorObject) EraseObject(t *pok.TileMap, pob PlacedEditorObject) {
	zIndex := 0

	for y := 0; y < edobj.H; y++ {
		gy := pob.Y + y

		for x := 0; x < edobj.W; x++ {
			gx := pob.X + x

			if (gx < 0 || gx >= t.Width) || (gy < 0 || gy >= t.Height) {
				zIndex++
				continue
			}

			index := t.Index(gx, gy)
			depth := pob.Z + edobj.Z[zIndex]

			t.Tiles[depth][index] = -1
			t.TextureIndicies[depth][index] = 0

			if (y > 0 || edobj.H == 1) && (x > 0 || edobj.W == 1) {
				t.Collision[pob.Z][index] = false
			}

			zIndex++
		}
	}
}
