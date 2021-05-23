package pok

import (
	"github.com/hajimehoshi/ebiten"
	"image"
	"image/color"
)

type ScrollDirection int

const (
	maxGridWidth = 8 * TileSize
	maxGridHeight = 20 * TileSize
	columnLen = 20

	xGridPos = WindowSizeX / 2 - maxGridWidth - 10
	yGridPos = 10

	ScrollUp = 0
	ScrollDown = 1
)


type Grid struct {
	tileSet *ebiten.Image
	selection *ebiten.Image
	selectionX float64
	selectionY float64
	innerWidth int
	currentIndex int
	currentCol int
	maxRow int
	nItemsPerRow int
	rect image.Rectangle
}

func NewGrid(tileSet *ebiten.Image, innerWidth int) Grid {
	selection, _ := ebiten.NewImage(innerWidth, innerWidth, ebiten.FilterDefault)
	selectionClr := color.RGBA{255, 0, 0, 255}
	for p := 0; p < selection.Bounds().Max.X; p++ {
		selection.Set(p, 0, selectionClr)
		selection.Set(p, selection.Bounds().Max.Y - 1, selectionClr)
	}
	for p := 1; p < selection.Bounds().Max.Y - 1; p++ {
		selection.Set(0, p, selectionClr)
		selection.Set(selection.Bounds().Max.X - 1, p, selectionClr)
	}

	totalTilesetItems := tileSet.Bounds().Max.X * tileSet.Bounds().Max.Y / innerWidth
	nItemsPerRow := maxGridWidth / innerWidth
	return Grid{
		tileSet: tileSet,
		selection: selection,
		selectionX: 0,
		selectionY: 0,
		innerWidth: innerWidth,
		currentIndex: 0,
		currentCol: 0,
		maxRow: totalTilesetItems / nItemsPerRow,
		nItemsPerRow: nItemsPerRow,
		rect: image.Rect(xGridPos, yGridPos, xGridPos + innerWidth * nItemsPerRow, yGridPos + innerWidth * columnLen),
	}
}

func (g *Grid) Draw(target *ebiten.Image) {
	w, _ := g.tileSet.Size()
	if w < 1 {
		return
	}
	//target.DrawImage(g.tileSet, &ebiten.DrawImageOptions{})
	nTilesX := w / g.innerWidth
	for i := 0; i < columnLen; i++ {
		for j := 0; j < g.nItemsPerRow; j++ {
			n := (i + g.currentCol) * g.nItemsPerRow + j
			tx := n % nTilesX * g.innerWidth
			ty := n / nTilesX * g.innerWidth
			gx := float64(j * g.innerWidth)
			gy := float64(i * g.innerWidth)
			opt := &ebiten.DrawImageOptions{}
			opt.GeoM.Translate(gx + xGridPos, gy + yGridPos)
			rect := image.Rect(tx, ty, tx + g.innerWidth, ty + g.innerWidth)
			target.DrawImage(g.tileSet.SubImage(rect).(*ebiten.Image), opt)
		}
	}

	if g.selectionY >= 0 && g.selectionY < float64(g.innerWidth * columnLen) {
		opt := &ebiten.DrawImageOptions{}
		opt.GeoM.Translate(xGridPos + g.selectionX, yGridPos + g.selectionY)
		target.DrawImage(g.selection, opt)
	}
}

func (g *Grid) Scroll(dir ScrollDirection) {
	if dir == ScrollUp && g.currentCol < g.maxRow {
		g.currentCol++
		g.selectionY -= float64(g.innerWidth)
	} else if dir == ScrollDown && g.currentCol > 0 {
		g.currentCol--
		g.selectionY += float64(g.innerWidth)
	}
}

func (g *Grid) Select(cx, cy int) {
	// Translate
	cx -= xGridPos
	cy -= yGridPos
	ix := cx / g.innerWidth
	iy := cy / g.innerWidth
	g.selectionX = float64(ix * g.innerWidth)
	g.selectionY = float64(iy * g.innerWidth)
	g.currentIndex = ix + (g.currentCol + iy) * g.nItemsPerRow
}

func (g *Grid) Contains(p image.Point) bool {
	return p.In(g.rect)
}

func (g *Grid) GetIndex() int {
	return g.currentIndex
}
