package pok

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"image"
	"image/color"
)

const (
	paddingX = 4
	paddingY = 2
)

var (
	fg = color.White
	bg = color.RGBA{33, 34, 35, 255}
	border = color.Black

	buttons []button
	buttonFont font.Face
)

type ButtonInfo struct {
	Content string
	OnClick func()
	X, Y int
}

type button struct {
	img *ebiten.Image
	rect image.Rectangle
	onClick func()
	x, y float64
}

func AddButton(button *ButtonInfo) {
	buttons = append(buttons, buttonFromButtonInfo(button))
}

func pollButtons(cx, cy int) {
	pt := image.Pt(cx, cy)

	for i := range buttons {
		if pt.In(buttons[i].rect) {
			buttons[i].onClick()
			break
		}
	}
}

func drawButtons(target *ebiten.Image) {
	for i := range buttons {
		opt := &ebiten.DrawImageOptions{}
		opt.GeoM.Translate(buttons[i].x, buttons[i].y)
		target.DrawImage(buttons[i].img, opt)
	}
}

func initButtons(font font.Face) {
	buttonFont = font
	buttons = make([]button, 0, 4)
}

func buildBox(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// fill
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			img.Set(x, y, bg)
		}
	}

	// border
	for x := 0; x < w; x++ {
		img.Set(x, 0, border)
		img.Set(x, h - 1, border)
	}

	for y := 0; y < h; y++ {
		img.Set(0, y, border)
		img.Set(w - 1, y, border)
	}

	return img
}

func buttonFromButtonInfo(buttonInfo *ButtonInfo) button {
	r := text.BoundString(buttonFont, buttonInfo.Content)
	w := r.Max.X + paddingX * 2
	h := r.Max.Y + paddingY * 2

	boundingBox := image.Rect(buttonInfo.X, buttonInfo.Y, buttonInfo.X + w, buttonInfo.Y + h)

	src := buildBox(w, h)
	img := ebiten.NewImageFromImage(src)
	text.Draw(img, buttonInfo.Content, buttonFont, paddingX, paddingY, fg)

	return button{
		img: img,
		rect: boundingBox,
		onClick: buttonInfo.OnClick,
		x: float64(buttonInfo.X),
		y: float64(buttonInfo.Y),
	}
}
