package pok

import (
	"github.com/hajimehoshi/ebiten"
	"image"
	"image/color"
)

type ScrollDirection int

const (
	nItemsPerRow = 6
	columnLen = 10

	xGridPos = WindowSizeX / 2 - (nItemsPerRow * TileSize) - 10
	yGridPos = 10

	ScrollUp = 0
	ScrollDown = 1
)


type Grid struct {
	tileSet *ebiten.Image
	selection *ebiten.Image
	selectionX float64
	selectionY float64
	currentIndex int
	currentCol int
	maxCol int
	rect image.Rectangle
}

func NewGrid(tileSet *ebiten.Image) Grid {
	selection, _ := ebiten.NewImage(TileSize, TileSize, ebiten.FilterDefault)
	selectionClr := color.RGBA{255, 0, 0, 255}
	for p := 0; p < selection.Bounds().Max.X; p++ {
		selection.Set(p, 0, selectionClr)
		selection.Set(p, selection.Bounds().Max.Y - 1, selectionClr)
	}
	for p := 1; p < selection.Bounds().Max.Y - 1; p++ {
		selection.Set(0, p, selectionClr)
		selection.Set(selection.Bounds().Max.Y - 1, p, selectionClr)
	}

	totalTilesetItems := tileSet.Bounds().Max.X * tileSet.Bounds().Max.Y / TileSize
	return Grid{
		tileSet,
		selection,
		0,
		0,
		0,
		0,
		totalTilesetItems / nItemsPerRow,
		image.Rect(xGridPos, yGridPos, xGridPos + TileSize * nItemsPerRow, yGridPos + TileSize * columnLen),
	}
}

func (g *Grid) Draw(target *ebiten.Image) {
	w, _ := g.tileSet.Size()
	nTilesX := w / TileSize
	for i := 0; i < columnLen; i++ {
		for j := 0; j < nItemsPerRow; j++ {
			n := (i + g.currentCol) * nItemsPerRow + j
			tx := n % nTilesX * TileSize
			ty := n / nTilesX * TileSize
			gx := float64(j) * TileSize
			gy := float64(i) * TileSize
			opt := &ebiten.DrawImageOptions{}
			opt.GeoM.Translate(gx + xGridPos, gy + yGridPos)
			rect := image.Rect(tx, ty, tx + TileSize, ty + TileSize)
			target.DrawImage(g.tileSet.SubImage(rect).(*ebiten.Image), opt)
		}
	}

	if g.selectionY >= 0 && g.selectionY < float64(TileSize * columnLen) {
		opt := &ebiten.DrawImageOptions{}
		opt.GeoM.Translate(xGridPos + g.selectionX, yGridPos + g.selectionY)
		target.DrawImage(g.selection, opt)
	}
}

func (g *Grid) Scroll(dir ScrollDirection) {
	if dir == ScrollUp && g.currentCol < g.maxCol {
		g.currentCol++
		g.selectionY -= TileSize
	} else if dir == ScrollDown && g.currentCol > 0 {
		g.currentCol--
		g.selectionY += TileSize
	}
}

func (g *Grid) Select(cx, cy int) {
	// Translate
	cx -= xGridPos
	cy -= yGridPos
	ix := cx / TileSize
	iy := cy / TileSize
	g.selectionX = float64(ix) * TileSize
	g.selectionY = float64(iy) * TileSize
	g.currentIndex = ix + (g.currentCol + iy) * nItemsPerRow
}

func (g *Grid) Contains(p image.Point) bool {
	return p.In(g.rect)
}

func (g *Grid) GetIndex() int {
	return g.currentIndex
}
