package textures

import(
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/atemmel/pok/pkg/debug"
)

var(
	aliases map[string]int
	textures []*ebiten.Image
)

const(
	preAlloc = 8
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

func Access(index int) *ebiten.Image {
	return textures[index];
}

func insertNewTexture(path string) (*ebiten.Image, int) {
	img, _, err := ebitenutil.NewImageFromFile(path, ebiten.FilterDefault)
	debug.Assert(err)

	for i, ptr := range textures {
		if ptr == nil {
			aliases[path] = i
			textures[i] = img
			return img, i
		}
	}

	i := len(textures)
	aliases[path] = i
	textures = append(textures, img)
	return img, i
}
