package editor

import(
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"image"
	"image/color"
)

type ContextMenuInfo struct {
	x int
	y int
	active bool
	font font.Face
	img *ebiten.Image
	items []ContextMenuItem
	boundingBoxes []image.Rectangle
}

var( 
	ContextMenu ContextMenuInfo
)

const(
	ContextMenuVerticalSpaceBetweenItems = 2
	ContextMenuInnerBorderPadding = 2
	ContextMenuPadding = 4
)

type ContextMenuItem struct {
	String string
	OnClick func()
}

func (menu *ContextMenuInfo) Init(font font.Face) {
	menu.font = font
	menu.img = nil
}

func (menu *ContextMenuInfo) Open(cx, cy int, items []ContextMenuItem) {
	menu.x = cx
	menu.y = cy
	menu.active = true
	if menu.img != nil {
		menu.img.Dispose()
	}

	menu.build(cx, cy, items)
}

func (menu *ContextMenuInfo) build(cx, cy int, items []ContextMenuItem) {
	menu.items = items
	menu.boundingBoxes = make([]image.Rectangle, len(menu.items))

	// calculate image size
	widestWidth := 0
	heightSum := 0

	for i, item := range items {
		r := text.BoundString(menu.font, item.String)
		w := r.Dx()
		h := r.Dy()

		// find widest item
		if widestWidth < w {
			widestWidth = w
		}

		// get total height
		heightSum += h
		heightSum += ContextMenuVerticalSpaceBetweenItems
		heightSum += ContextMenuInnerBorderPadding * 2

		// save bounding box to check for clicks later
		menu.boundingBoxes[i] = r.Add(image.Pt(cx, cy))
	}
	heightSum -= ContextMenuVerticalSpaceBetweenItems
	heightSum += ContextMenuInnerBorderPadding

	// add padding to the equation
	cWidth := widestWidth + (ContextMenuPadding * 2) + 1
	cHeight := heightSum + (ContextMenuPadding * 2)


	// build image in regular ram first
	img := image.NewRGBA(image.Rect(0, 0, cWidth, cHeight))
	
		// fill
	for x := 0; x < cWidth; x++ {
		for y := 0; y < cHeight; y++ {
			img.Set(x, y, Bg)
		}
	}

	// border
	for x := 0; x < cWidth; x++ {
		img.Set(x, 0, Border)
		img.Set(x, cHeight - 1, Border)
	}

	for y := 0; y < cHeight; y++ {
		img.Set(0, y, Border)
		img.Set(cWidth - 1, y, Border)
	}

	img.Set(0, 0, color.Transparent)
	img.Set(0, cHeight - 1, color.Transparent)
	img.Set(cWidth - 1, 0, color.Transparent)
	img.Set(cWidth - 1, cHeight - 1, color.Transparent)

	img.Set(1, 0, color.Transparent)
	img.Set(1, cHeight - 1, color.Transparent)
	img.Set(cWidth - 2, 0, color.Transparent)
	img.Set(cWidth - 2, cHeight - 1, color.Transparent)

	img.Set(0, 1, color.Transparent)
	img.Set(0, cHeight - 2, color.Transparent)
	img.Set(cWidth - 1, 1, color.Transparent)
	img.Set(cWidth - 1, cHeight - 2, color.Transparent)

	img.Set(1, 1, Border)
	img.Set(1, cHeight - 2, Border)
	img.Set(cWidth - 2, 1, Border)
	img.Set(cWidth - 2, cHeight - 2, Border)


	// inner borders
	baseY := ContextMenuPadding
	for _, r := range menu.boundingBoxes {
		baseX := ContextMenuPadding - 2
		w := cWidth - ContextMenuPadding * 2 + 3
		h := r.Dy() + ContextMenuInnerBorderPadding * 2

		for x := 1; x < w; x++ {
			img.Set(baseX + x, baseY, Border)
			img.Set(baseX + x, baseY + h, Border)
		}

		for y := 1; y < h; y++ {
			img.Set(baseX, baseY + y, Border)
			img.Set(baseX + w, baseY + y, Border)
		}

		baseY += h
		baseY += ContextMenuVerticalSpaceBetweenItems 
		baseY += ContextMenuInnerBorderPadding
	}

	// then make it in vram
	menu.img = ebiten.NewImageFromImage(img)
	y := ContextMenuPadding
	for i, item := range items {
		// advance y
		y += menu.boundingBoxes[i].Dy()
		y += ContextMenuInnerBorderPadding

		// draw shadow first and offset
		text.Draw(menu.img, item.String, menu.font, 
			ContextMenuPadding + 1,
			y + 1,
			FgShadow,
		)

		// then draw actual color on top
		text.Draw(menu.img, item.String, menu.font, 
			ContextMenuPadding,
			y,
			Fg,
		)

		y += ContextMenuVerticalSpaceBetweenItems
		y += ContextMenuInnerBorderPadding * 2
	}
}

func (menu *ContextMenuInfo) Close() {
	menu.active = false
}

func (menu *ContextMenuInfo) IsOpen() bool {
	return menu.active
}

func (menu *ContextMenuInfo) Draw(target *ebiten.Image) {
	if menu.IsOpen() {
		opt := &ebiten.DrawImageOptions{}
		opt.GeoM.Translate(float64(menu.x), float64(menu.y))
		target.DrawImage(menu.img, opt)
	}
}
