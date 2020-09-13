package pok

import (
	"github.com/hajimehoshi/ebiten"
	"image"
)

const (
	nItemsPerRow = 4
	columnLen = 10

	xGridPos = WindowSizeX / 2 - (nItemsPerRow * TileSize) - 10
	yGridPos = 10
)

type Grid struct {
	tileSet *ebiten.Image
	currentCol int
	maxCol int
}

func NewGrid(tileSet *ebiten.Image) Grid {
	totalTilesetItems := tileSet.Bounds().Max.X * tileSet.Bounds().Max.Y / TileSize
	return Grid{
		tileSet,
		0,
		totalTilesetItems / nItemsPerRow,
	}
}

func (g *Grid) Draw(target *ebiten.Image) {
	for i := 0; i < columnLen; i++ {
		for j := 0; j < nItemsPerRow; j++ {
			n := i * nItemsPerRow + j
			tx := n % NTilesX * TileSize
			ty := n / NTilesX * TileSize
			gx := float64(j) * TileSize
			gy := float64(i) * TileSize
			opt := &ebiten.DrawImageOptions{}
			opt.GeoM.Translate(gx + xGridPos, gy + yGridPos)
			rect := image.Rect(tx, ty, tx + TileSize, ty + TileSize)
			target.DrawImage(g.tileSet.SubImage(rect).(*ebiten.Image), opt)
		}
	}
}
