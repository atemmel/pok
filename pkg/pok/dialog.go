package pok

import (
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/text"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"io/ioutil"
	"image/color"
)

const (
	textXDelta = 12
	textYDelta = 17
	maxLetters = 44
)

var (
	fgClr = color.RGBA{80, 80, 88, 255}
	bgClr = color.RGBA{160, 160, 168, 255}
)

type DialogBox struct {
	Hidden bool
	str string
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

	db.box, _, err = ebitenutil.NewImageFromFile("./resources/images/dialog1.png", ebiten.FilterDefault);
	if err != nil {
		panic(err);
	}

	//db.SetString("Lorem ipsum dolor sit amet, lorem ipsum dolor sit amet")
	db.Hidden = true;

	return db
}

func (d *DialogBox) SetString(str string) {
	result := str
	if(len(str) > maxLetters) {
		index := maxLetters
		for i := maxLetters; i > 0; i-- {
			if result[i] == ' ' {
				index = i
				break
			}
		}
		result = result[:index] + "\n" + result[index + 1:];
	}
	d.str = result
}

func (d *DialogBox) Draw(target *ebiten.Image) {
	if d.Hidden {
		return
	}
	opt := &ebiten.DrawImageOptions{}
	dx := DisplaySizeX / 2 - d.box.Bounds().Dx() / 2
	dy := DisplaySizeY - d.box.Bounds().Dy() - 4
	opt.GeoM.Translate(float64(dx), float64(dy))
	target.DrawImage(d.box, opt)
	text.Draw(target, d.str, d.font, dx + textXDelta + 1, dy + textYDelta, bgClr)
	text.Draw(target, d.str, d.font, dx + textXDelta, dy + textYDelta + 1, bgClr)
	text.Draw(target, d.str, d.font, dx + textXDelta + 1, dy + textYDelta + 1, bgClr)
	text.Draw(target, d.str, d.font, dx + textXDelta, dy + textYDelta, fgClr)
}
