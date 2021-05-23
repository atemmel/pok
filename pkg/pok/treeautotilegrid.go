package pok

import(
	"github.com/hajimehoshi/ebiten"
	"image"
)

type TreeAutoTileGrid struct {
	grid Grid
}

func NewTreeAutoTileGrid(tileSet *ebiten.Image, tatis []TreeAutoTileInfo) TreeAutoTileGrid {
	w := len(tatis) * TileSize * SingleTreeWidth
	h := SingleTreeHeight * TileSize

	img, _ := ebiten.NewImage(w, h, ebiten.FilterDefault)

	for i := range tatis {
		tx := tatis[i].SingleStart.X * TileSize
		ty := tatis[i].SingleStart.Y * TileSize

		rect := image.Rect(tx, ty, tx + (SingleTreeWidth * TileSize), ty + (SingleTreeHeight * TileSize))
		opt := &ebiten.DrawImageOptions{}

		opt.GeoM.Translate(TileSize * float64(i * SingleTreeWidth), 0)
		img.DrawImage(tileSet.SubImage(rect).(*ebiten.Image), opt)
	}

	grid := NewGrid(img, TileSize * SingleTreeWidth)
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
