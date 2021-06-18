package pok

import(
	"github.com/atemmel/pok/pkg/constants"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
)

type AutoTileGrid struct {
	grid Grid
}

func NewAutoTileGrid(tileSet *ebiten.Image, nTilesX int, atis []AutoTileInfo) AutoTileGrid {
	w := len(atis) * constants.TileSize
	img := ebiten.NewImage(w, constants.TileSize)

	for i := range atis {
		tx := (atis[i].Center % nTilesX) * constants.TileSize
		ty := (atis[i].Center / nTilesX) * constants.TileSize

		rect := image.Rect(tx, ty, tx + constants.TileSize, ty + constants.TileSize)
		opt := &ebiten.DrawImageOptions{}

		opt.GeoM.Translate(constants.TileSize * float64(i), 0)
		img.DrawImage(tileSet.SubImage(rect).(*ebiten.Image), opt)
	}

	grid := NewGrid(img, constants.TileSize)
	return AutoTileGrid{
		grid,
	}
}

func (atg *AutoTileGrid) Draw(target *ebiten.Image) {
	atg.grid.Draw(target)
}

func (atg *AutoTileGrid) Scroll(dir ScrollDirection) {
	atg.grid.Scroll(dir)
}

func (atg *AutoTileGrid) Select(cx, cy int) {
	atg.grid.Select(cx, cy)
}

func (atg *AutoTileGrid) Contains(p image.Point) bool {
	return atg.grid.Contains(p)
}

func (atg *AutoTileGrid) GetIndex() int {
	return atg.grid.GetIndex()
}
