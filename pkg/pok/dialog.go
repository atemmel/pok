package pok

import (
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/text"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"io/ioutil"
	"image/color"
)

const (
	dialogX = 120
	dialogY = 120
)

type DialogBox struct {
	str *string
	font font.Face
}

func NewDialogBox() DialogBox {
	data, err := ioutil.ReadFile("./resources/pokemon_pixel_font.ttf")
	if err != nil {
		panic(err)
	}

	tt, err := truetype.Parse(data)
	if err != nil {
		panic(err)
	}

	db := DialogBox{}

	const dpi = 72
	db.font = truetype.NewFace(tt, &truetype.Options{
		Size: 16,
		DPI: dpi,
		Hinting: font.HintingFull,
	})

	return db
}

func (d *DialogBox) Draw(target *ebiten.Image) {
	clr := color.RGBA{20, 20, 20, 255}
	text.Draw(target, "Tjena moss", d.font, dialogX, dialogY, clr)
}
