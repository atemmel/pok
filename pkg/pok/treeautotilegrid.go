package pok

import(
	"github.com/atemmel/pok/pkg/constants"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
)

type TreeAutoTileGrid struct {
	grid Grid
}

func NewTreeAutoTileGrid(tileSet *ebiten.Image, tatis []TreeAutoTileInfo) TreeAutoTileGrid {
	w := len(tatis) * constants.TileSize * SingleTreeWidth
	h := SingleTreeHeight * constants.TileSize

	img := ebiten.NewImage(w, h)

	for i := range tatis {
		tx := tatis[i].SingleStart.X * constants.TileSize
		ty := tatis[i].SingleStart.Y * constants.TileSize

		rect := image.Rect(tx, ty, tx + (SingleTreeWidth * constants.TileSize), ty + (SingleTreeHeight * constants.TileSize))
		opt := &ebiten.DrawImageOptions{}

		opt.GeoM.Translate(constants.TileSize * float64(i * SingleTreeWidth), 0)
		img.DrawImage(tileSet.SubImage(rect).(*ebiten.Image), opt)
	}

	grid := NewGrid(img, constants.TileSize * SingleTreeWidth)
	return TreeAutoTileGrid{
		grid,
	}
}

func (tatg *TreeAutoTileGrid) Draw(target *ebiten.Image) {
	tatg.grid.Draw(target)
}

func (tatg *TreeAutoTileGrid) Scroll(dir ScrollDirection) {
	tatg.grid.Scroll(dir)
}

func (tatg *TreeAutoTileGrid) Select(cx, cy int) {
	tatg.grid.Select(cx, cy)
}

func (tatg *TreeAutoTileGrid) Contains(p image.Point) bool {
	return tatg.grid.Contains(p)
}

func (tatg *TreeAutoTileGrid) GetIndex() int {
	return tatg.grid.GetIndex()
}
