package textures

import(
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/atemmel/pok/pkg/debug"
	"github.com/atemmel/pok/pkg/constants"

	_ "image/png"
)

var(
	aliases map[string]int
	textures []*ebiten.Image

	baseTextureIndex = InvalidIndex
	waterTextureIndex = InvalidIndex
	stairTextureIndex = InvalidIndex
)

const(
	InvalidIndex = -1

	preAlloc = 8
	baseTextureStr = constants.TileMapImagesDir + "base.png"
	waterTextureStr = constants.TileMapImagesDir + "water.png"
	stairTextureStr = constants.TileMapImagesDir + "stairs.png"
)

func Init() {
	aliases = make(map[string]int, preAlloc)
	textures = make([]*ebiten.Image, 0, preAlloc)
}

func Load(path string) (*ebiten.Image, int) {
	index, ok := aliases[path]
	if !ok {
		return insertNewTexture(path);
	}
	return Access(index), index
}

func LoadWithError(path string) (*ebiten.Image, error) {
	img, _, err := ebitenutil.NewImageFromFile(path)
	return img, err
}

func Access(index int) *ebiten.Image {
	return textures[index];
}

func IsWater(index int) bool {
	return index == waterTextureIndex
}

func IsBase(index int) bool {
	return index == baseTextureIndex
}

func IsStair(index int) bool {
	return index == stairTextureIndex
}

func insertNewTexture(path string) (*ebiten.Image, int) {
	img, _, err := ebitenutil.NewImageFromFile(path)
	debug.Assert(err)

	for i, ptr := range textures {
		if ptr == nil {
			aliases[path] = i
			textures[i] = img
			checkForSpecialTextures(path, i)
			return img, i
		}
	}

	i := len(textures)
	aliases[path] = i
	textures = append(textures, img)
	checkForSpecialTextures(path, i)
	return img, i
}

func checkForSpecialTextures(path string, index int) {
	if baseTextureIndex == InvalidIndex && baseTextureStr == path {
		baseTextureIndex = index
	} else if waterTextureIndex == InvalidIndex && waterTextureStr == path {
		waterTextureIndex = index
	} else if stairTextureIndex == InvalidIndex && stairTextureStr == path {
		stairTextureIndex = index
	}
}
