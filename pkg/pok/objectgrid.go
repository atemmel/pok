package pok

import (
	"github.com/atemmel/pok/pkg/constants"
	"github.com/atemmel/pok/pkg/textures"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"image/color"
)

type ObjectGrid struct {
	objs []EditorObject
	rects []image.Rectangle
	tileMap *TileMap
	selection *ebiten.Image
	selectionX, selectionY float64
	currentIndex int
	scrollDepth int
	maxScrollDepth int
	rect image.Rectangle
}

func NewObjectGrid(tileMap *TileMap, objs []EditorObject) ObjectGrid {
	maxDepth := 0
	for i := range objs {
		maxDepth += objs[i].H
	}

	rects := make([]image.Rectangle, len(objs))

	x := 0
	y := 0
	oldh := 0

	for i := range objs {
		w := objs[i].W * constants.TileSize
		h := objs[i].H * constants.TileSize

		if x + w >= 8 * constants.TileSize {
			x = 0
			y += oldh
		}

		r := image.Rect(x, y, x + w, y + h)
		rects[i] = r
		x += w
		oldh = h
	}

	return ObjectGrid{
		objs,
		rects,
		tileMap,
		nil,
		0, 0,
		0,
		0,
		maxDepth,
		image.Rect(xGridPos, yGridPos, xGridPos + constants.TileSize * 8, yGridPos + constants.TileSize * columnLen),
	}
}

func (og *ObjectGrid) Draw(target *ebiten.Image) {
	y := 0.0
	for i, ob := range og.objs {
		r := image.Rect(ob.X * constants.TileSize, ob.Y * constants.TileSize, (ob.X + ob.W) * constants.TileSize, (ob.Y + ob.H) * constants.TileSize)
		//img := og.tileMap.images[ob.textureIndex]
		img := textures.Access(og.tileMap.textureMapping[ob.textureIndex])
		opt := &ebiten.DrawImageOptions{}
		dx := xGridPos + float64(og.rects[i].Min.X)
		dy := yGridPos + float64(og.rects[i].Min.Y)
		opt.GeoM.Translate(dx, dy)
		y += float64(ob.H) * constants.TileSize
		target.DrawImage(img.SubImage(r).(*ebiten.Image), opt)
	}

	if og.selection != nil && og.currentIndex >= 0 && og.currentIndex < len(og.objs) {
		opt := &ebiten.DrawImageOptions{}
		opt.GeoM.Translate(xGridPos + og.selectionX, yGridPos + og.selectionY)
		target.DrawImage(og.selection, opt)
	}
}

func (og *ObjectGrid) Scroll(dir ScrollDirection) {

}

func (og *ObjectGrid) Select(cx, cy int) int {
	cy -= yGridPos
	cx -= xGridPos

	p := image.Point{cx, cy}
	for i := range og.objs {
		if !p.In(og.rects[i]) {
			continue
		}

		currentIndex := i
		ob := &og.rects[currentIndex]
		selection := ebiten.NewImage(ob.Bounds().Dx(), ob.Bounds().Dy())

		selectionClr := color.RGBA{255, 0, 0, 255}
		for p := 0; p < selection.Bounds().Max.X; p++ {
			selection.Set(p, 0, selectionClr)
			selection.Set(p, selection.Bounds().Max.Y - 1, selectionClr)
		}
		for p := 1; p < selection.Bounds().Max.Y - 1; p++ {
			selection.Set(0, p, selectionClr)
			selection.Set(selection.Bounds().Max.X - 1, p, selectionClr)
		}
		og.selection = selection
		og.selectionX = float64(og.rects[i].Min.X)
		og.selectionY = float64(og.rects[i].Min.Y)

		return i
	}

	return -1
}

func (og *ObjectGrid) Contains(p image.Point) bool {
	return p.In(og.rect)
}
