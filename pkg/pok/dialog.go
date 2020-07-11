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
	box *ebiten.Image
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

	boxColor := color.RGBA{
		255, 255, 255, 255,
	}

	db.box, _ = ebiten.NewImage(DisplaySizeX, 16 * 2 + 10, ebiten.FilterDefault)
	db.box.Fill(boxColor)

	return db
}

func (d *DialogBox) Draw(target *ebiten.Image) {
	clr := color.RGBA{20, 20, 20, 255}
	opt := &ebiten.DrawImageOptions{}
	opt.GeoM.Translate(0, DisplaySizeY - (16 * 2 + 10))
	target.DrawImage(d.box, opt)
	text.Draw(target, "Tjena moss", d.font, dialogX, dialogY, clr)
}
