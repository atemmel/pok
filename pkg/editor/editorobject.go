package editor

import(
	"github.com/atemmel/pok/pkg/pok"
	"github.com/atemmel/pok/pkg/textures"
	"encoding/json"
	"io/ioutil"
	"image"
	"strings"
)

var ObjectIdIncrementer = 0

type EditorObject struct {
	Texture string
	X, Y int
	W, H int
	Z []int

	HasCollision bool
	CollidesWithTop bool
	CollidesWithBottom bool
	CollidesWithLeft bool
	CollidesWithRight bool

	textureIndex int
}

func (e *EditorObject) FindAndSetCorrectTexture(allTextures []string) {
	e.textureIndex = 0
	for i := range allTextures {
		if allTextures[i] == e.Texture {
			e.textureIndex = i
			return
		}
	}
	_, e.textureIndex = textures.Load(e.Texture)
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
	for i, instance := range pobs {
		obj := edobjs[instance.Index]

		r := image.Rect(0, 0, obj.W, obj.H).Add(image.Pt(instance.X, instance.Y))
		if prospect.In(r) {
			return i
		}
	}
	return -1
}

func HasPlacedObjectExactlyAt(pobs []PlacedEditorObject, x, y int) int {
	for i, obj := range pobs {
		if obj.X == x && obj.Y == y {
			return i
		}
	}
	return -1
}

type ObjectInsertionParameters struct {
	TileMap *pok.TileMap
	ObjectInstances *[]PlacedEditorObject
	ObjectTypes []EditorObject
	ObjectTypeIndex int
	xyIndex int
	zIndex int
}

func InsertObjectIntoTileMap(params *ObjectInsertionParameters) {
	t := params.TileMap
	col, row := t.Coords(params.xyIndex)
	edobj := &params.ObjectTypes[params.ObjectTypeIndex]
	objectInstances := params.ObjectInstances
	existingObjectIndex := HasPlacedObjectAt(*objectInstances, params.ObjectTypes, col ,row)
	// if exists
	if existingObjectIndex != -1 {
		edobj.EraseObject(t, (*objectInstances)[existingObjectIndex])

		// fast pop? does order matter? it might
		(*objectInstances)[existingObjectIndex] = (*objectInstances)[len(*objectInstances) - 1]
		*objectInstances = (*objectInstances)[:len(*objectInstances) - 1]
	}

	// get max depth
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

	for y := 0; y < edobj.H; y++ {
		gy := row + y

		ty := (edobj.Y + y) * t.NTilesX(edobj.textureIndex)

		for x := 0; x < edobj.W; x++ {
			gx := col + x

			// if outside of tilemap
			if (gx < 0 || gx >= t.Width) || (gy < 0 || gy >= t.Height) {
				zIndex++
				continue
			}

			tx := edobj.X + x

			tile := ty + tx
			index := t.Index(gx, gy)
			depth := params.zIndex + edobj.Z[zIndex]

			t.Tiles[depth][index] = tile
			t.TextureIndicies[depth][index] = texIndex

			if !edobj.HasCollision {
				goto SKIP
			} else if !edobj.CollidesWithTop && y == 0 {
				goto SKIP
			} else if !edobj.CollidesWithBottom && y == edobj.H - 1 {
				goto SKIP
			} else if !edobj.CollidesWithLeft && x == 0 {
				goto SKIP
			} else if !edobj.CollidesWithRight && x == edobj.W - 1 {
				goto SKIP
			}

			t.Collision[params.zIndex][index] = true

SKIP:
			zIndex++
		}
	}

	p := PlacedEditorObject{
		col, row, params.zIndex,
		params.ObjectTypeIndex,
	}

	*objectInstances = append(*objectInstances, p)
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

			if !edobj.HasCollision {
				goto SKIP
			} else if !edobj.CollidesWithTop && y == 0 {
				goto SKIP
			} else if !edobj.CollidesWithBottom && y == edobj.H - 1 {
				goto SKIP
			} else if !edobj.CollidesWithLeft && x == 0 {
				goto SKIP
			} else if !edobj.CollidesWithRight && x == edobj.W - 1 {
				goto SKIP
			}

			t.Collision[pob.Z][index] = false

SKIP:

			zIndex++
		}
	}
}
