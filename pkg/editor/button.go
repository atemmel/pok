package editor

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"image"
)

const (
	paddingX = 4
	paddingY = 4
)

var (
	buttons []button
	buttonFont font.Face

	minH = 4
	minW = 4
)

type ButtonInfo struct {
	Content string
	OnClick func()
	VisibilityCondition func() bool
	X, Y int
}

type button struct {
	img *ebiten.Image
	rect image.Rectangle
	onClick func()
	condition func() bool
	x, y float64
}

func AddButton(button *ButtonInfo) {
	buttons = append(buttons, buttonFromButtonInfo(button))
}

func pollButtons(cx, cy int) bool {
	pt := image.Pt(cx, cy)

	for i := range buttons {
		if (buttons[i].condition == nil || buttons[i].condition()) && pt.In(buttons[i].rect) {
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton(0)) {
				buttons[i].onClick()
			}
			return true
		}
	}

	return false
}

func drawButtons(target *ebiten.Image) {
	for i := range buttons {
		if (buttons[i].condition == nil || buttons[i].condition()) {
			opt := &ebiten.DrawImageOptions{}
			opt.GeoM.Translate(buttons[i].x, buttons[i].y)
			target.DrawImage(buttons[i].img, opt)
		}
	}
}

func initButtons(font font.Face) {
	buttonFont = font
	buttons = make([]button, 0, 4)
}

func buttonFromButtonInfo(buttonInfo *ButtonInfo) button {
	r := text.BoundString(buttonFont, buttonInfo.Content)
	w := r.Dx()
	h := r.Dy()

	if w < minW {
		w = minW
	}

	if h < minH {
		h = minH
	} else if h > minH {
		minH = h
	}

	w += paddingX * 2
	h += paddingY * 2

	boundingBox := image.Rect(buttonInfo.X, buttonInfo.Y, buttonInfo.X + w, buttonInfo.Y + h)

	src := CreateNeatImageWithBorder(w, h)
	img := ebiten.NewImageFromImage(src)

	const extraOffset = 9
	text.Draw(img, buttonInfo.Content, buttonFont, paddingX + 1, paddingY + extraOffset + 1, FgShadow)
	text.Draw(img, buttonInfo.Content, buttonFont, paddingX, paddingY + extraOffset, Fg)

	return button{
		img: img,
		rect: boundingBox,
		onClick: buttonInfo.OnClick,
		condition: buttonInfo.VisibilityCondition,
		x: float64(buttonInfo.X),
		y: float64(buttonInfo.Y),
	}
}
