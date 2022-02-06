package editor

import (
	"github.com/atemmel/pok/pkg/constants"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"image"
	"image/color"
)

type ScrollDirection int

const (
	maxGridWidth = 8 * constants.TileSize
	maxGridHeight = 20 * constants.TileSize
	columnLen = 20

	xGridPos = constants.WindowSizeX / 2 - maxGridWidth - 10
	//yGridPos = 10
	yGridPos = 10 + 10

	xScrollBarPos = xGridPos + maxGridWidth + 2
	yScrollBarPos = yGridPos

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
	maxIndex int
	currentRow int
	maxRow int
	nItemsPerRow int
	rect image.Rectangle
	bar *Scrollbar
}

type Scrollbar struct {
	img *ebiten.Image
	x int
	y int
	min int
	max int
	holdOffsetY int
	isHolding bool
}

func NewScrollBar(x, y, rows int) *Scrollbar {
	const scrollBarWidth = 8
	scale := columnLen / float64(rows)

	if scale >= 1 {
		// no need
		return nil
	} else if scale < 0.05 {
		// minimum
		scale = 0.05
	}

	scrollBarHeight := int((maxGridHeight + columnLen) * scale)
	img := CreateNeatImageWithBorder(scrollBarWidth, scrollBarHeight)
	ebitenImg := ebiten.NewImageFromImage(img)

	return &Scrollbar{
		ebitenImg,
		x,
		y,
		y,
		y + maxGridHeight - scrollBarHeight,
		0,
		false,
	}
}

func (s *Scrollbar) Draw(target *ebiten.Image) {
	opt := &ebiten.DrawImageOptions{}
	opt.GeoM.Translate(float64(s.x), float64(s.y))
	if s.isHolding {
		opt.ColorM.Scale(0.5, 0.5, 0.5, 1.0)
	}
	target.DrawImage(s.img, opt)
}

func NewGrid(tileSet *ebiten.Image, innerWidth int) Grid {
	selection := ebiten.NewImage(innerWidth, innerWidth)
	selectionClr := color.RGBA{255, 0, 0, 255}
	for p := 0; p < selection.Bounds().Max.X; p++ {
		selection.Set(p, 0, selectionClr)
		selection.Set(p, selection.Bounds().Max.Y - 1, selectionClr)
	}
	for p := 1; p < selection.Bounds().Max.Y - 1; p++ {
		selection.Set(0, p, selectionClr)
		selection.Set(selection.Bounds().Max.X - 1, p, selectionClr)
	}

	totalTilesetItems := tileSet.Bounds().Max.X * tileSet.Bounds().Max.Y / (innerWidth * innerWidth)
	nItemsPerRow := maxGridWidth / innerWidth
	maxRow := totalTilesetItems / nItemsPerRow
	return Grid{
		tileSet: tileSet,
		selection: selection,
		selectionX: 0,
		selectionY: 0,
		innerWidth: innerWidth,
		currentIndex: 0,
		maxIndex: totalTilesetItems,
		currentRow: 0,
		maxRow: maxRow,
		nItemsPerRow: nItemsPerRow,
		rect: image.Rect(xGridPos, yGridPos, xGridPos + innerWidth * nItemsPerRow, yGridPos + innerWidth * columnLen),
		bar: NewScrollBar(xScrollBarPos, yScrollBarPos, maxRow),
	}
}

func (g *Grid) Draw(target *ebiten.Image) {
	w, _ := g.tileSet.Size()
	if w < 1 {
		return
	}
	nTilesX := w / g.innerWidth
	for i := 0; i < columnLen; i++ {
		for j := 0; j < g.nItemsPerRow; j++ {
			n := (i + g.currentRow) * g.nItemsPerRow + j
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

	if g.bar != nil {
		g.bar.Draw(target)
	}
}

func (g *Grid) Scroll(dir ScrollDirection) {

	// check bounds
	if dir == ScrollUp && g.currentRow < g.maxRow - columnLen + 1 {
		g.currentRow++
		g.selectionY -= float64(g.innerWidth)
	} else if dir == ScrollDown && g.currentRow > 0 {
		g.currentRow--
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
	g.currentIndex = ix + (g.currentRow + iy) * g.nItemsPerRow
}

func (g *Grid) Contains(p image.Point) bool {
	return p.In(g.rect)
}

func (g *Grid) PollScrollBar(cx, cy int) bool {
	if g.bar == nil {
		return false
	}

	holdResult := g.bar.handleHold(cx, cy)
	scale := g.bar.GetAmountScrolled()

	g.currentRow = int(scale * float64(g.maxRow))
	
	return holdResult
}

func (b *Scrollbar) handleHold(cx, cy int) bool {
	prospect := image.Pt(cx, cy)
	r := b.img.Bounds().Add(image.Pt(b.x, b.y))

	if b.isHolding {
		b.updatePos(cx, cy)
		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButton(0)) {
			b.isHolding = false
			justDidSomethingOfInterest = false
		}
		return false
	}

	if !prospect.In(r) {
		return false
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton(0)) {
		b.isHolding = true
		b.holdOffsetY = cy - b.y
		b.updatePos(cx, cy)
		justDidSomethingOfInterest = true
		justDidSomethingOfInterestLock = true
		return true
	}
	return false
}

func (b *Scrollbar) updatePos(cx, cy int) {
	// update the position
	center := cy - b.holdOffsetY
	if center < b.min {
		b.y = b.min
	} else if center > b.max {
		b.y = b.max
	} else {
		b.y = center
	}
}

func (b *Scrollbar) GetAmountScrolled() float64 {

	y0 := float64(b.y - b.min)
	yMax := float64(b.max - b.min)

	scale := y0 / yMax
	if scale > 1 {
		scale = 1
	}
	return scale
}

func (g *Grid) GetIndex() int {
	if g.currentIndex < g.maxIndex {
		return g.currentIndex
	}
	return g.maxIndex - 1
}
