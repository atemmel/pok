package main

import(
	"image"
	"github.com/atemmel/pok/pkg/editor"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
)

var(
	buttonFont font.Face
	buttons map [int]buttonData
)

type Button struct {
	X, Y int
	OnClick func()
	Title string
}

func InitButtons(font font.Face) {
	buttonFont = font
	buttons = make(map [int]buttonData)
}

func AddButton(button *Button) {
	i := len(buttons)
	buttons[i] = buttonDataFromButton(button)
}

func DeleteButton(i int) {
	delete(buttons, i)
}

func DrawButtons(target *ebiten.Image) {
	for i := range buttons {
		opt := &ebiten.DrawImageOptions{}
		opt.GeoM.Translate(buttons[i].x, buttons[i].y)
		target.DrawImage(buttons[i].img, opt)
	}
}

func PollButtons(cx, cy int) bool {
	pt := image.Pt(cx, cy)
	for i, b := range buttons {
		bounds := b.img.Bounds()
		bounds = bounds.Add(image.Pt(int(b.x), int(b.y)))
		if pt.In(bounds) {
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton(0)) {
				buttons[i].onClick()
				return true
			}
		}
	}

	return false
}

func makeButtonTitle(title string) *ebiten.Image {
	const minDim = 4
	const paddingX = 4
	const paddingY = 4

	r := text.BoundString(buttonFont, title)
	w := r.Dx()
	h := r.Dy()

	if w < minDim {
		w = minDim
	}

	if h < minDim {
		h = minDim
	}

	w += paddingX * 2
	h += paddingX * 2

	src := editor.CreateNeatImageWithBorder(w, h)
	img := ebiten.NewImageFromImage(src)

	const extraOffset = 9
	text.Draw(img, title, buttonFont, paddingX + 1, paddingY + extraOffset + 1, editor.FgShadow)
	text.Draw(img, title, buttonFont, paddingX, paddingY + extraOffset, editor.Fg)
	return img
}

func buttonDataFromButton(button *Button) buttonData {
	return buttonData{
		x: float64(button.X),
		y: float64(button.Y),
		onClick: button.OnClick,
		img: makeButtonTitle(button.Title),
	}
}

type buttonData struct {
	x, y float64
	onClick func()
	img *ebiten.Image
}
