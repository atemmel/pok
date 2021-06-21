package pok

import (
	"github.com/atemmel/pok/pkg/constants"
	"github.com/atemmel/pok/pkg/debug"
	"github.com/atemmel/pok/pkg/dialog"
	"github.com/atemmel/pok/pkg/textures"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"io/ioutil"
	"image/color"
)

const (
	textXDelta = 12
	//textYDelta = 17
	textYDelta = 21
)

const (
	TextSlow = iota
	TextNormal
	TextFast
	TextInstant
)

var (
	fgClr = color.RGBA{80, 80, 88, 255}
	bgClr = color.RGBA{160, 160, 168, 255}
)

type DialogBox struct {
	Hidden bool
	fullStr string
	dispStr string
	font font.Face
	box *ebiten.Image
	speed int
	ticks int
}

func NewDialogBox() DialogBox {
	data, err := ioutil.ReadFile(constants.FontsDir + "pokemon_pixel_font.ttf")
	debug.Assert(err)

	tt, err := truetype.Parse(data)
	debug.Assert(err)

	db := DialogBox{}

	const dpi = 72
	db.font = truetype.NewFace(tt, &truetype.Options{
		Size: 16,
		DPI: dpi,
		Hinting: font.HintingFull,
	})

	db.box, err = textures.LoadWithError(constants.ImagesDir + "dialog0.png")
	debug.Assert(err)

	db.Hidden = true
	db.speed = TextNormal

	return db
}

func (d *DialogBox) SetString(str string) {
	result := str
	hasBreak := false
	for i := range str {
		if str[i] == '\n' {
			hasBreak = true
		}
	}
	if !hasBreak && len(str) > dialog.MaxLetters {
		index := dialog.MaxLetters
		for i := dialog.MaxLetters; i > 0; i-- {
			if result[i] == ' ' {
				index = i
				break
			}
		}
		result = result[:index] + "\n" + result[index + 1:];
	}

	d.ticks = 0
	if d.speed == TextInstant {
		d.dispStr = result
	} else {
		d.dispStr = ""
	}
	d.fullStr = result
}

func (d *DialogBox) IsDone() bool {
	return len(d.dispStr) >= len(d.fullStr)
}

func (d *DialogBox) Update() {
	if !d.Hidden && !d.IsDone() {
		switch d.speed {
			case TextSlow:
				if d.ticks >= 3 {
					d.nextChar()
				}
			case TextNormal:
				if d.ticks >= 2 {
					d.nextChar()
				}
			case TextFast:
				if d.ticks >= 1 {
					d.nextChar()
				}
		}
		d.ticks++
	}
}

func (d *DialogBox) nextChar() {
	d.dispStr = d.fullStr[:len(d.dispStr)+1]
	d.ticks = 0
}

func (d *DialogBox) Draw(target *ebiten.Image) {
	if d.Hidden {
		return
	}
	opt := &ebiten.DrawImageOptions{}
	dx := constants.DisplaySizeX / 2 - d.box.Bounds().Dx() / 2
	dy := constants.DisplaySizeY - d.box.Bounds().Dy() - 4
	opt.GeoM.Translate(float64(dx), float64(dy))
	target.DrawImage(d.box, opt)
	text.Draw(target, d.dispStr, d.font, dx + textXDelta + 1, dy + textYDelta, bgClr)
	text.Draw(target, d.dispStr, d.font, dx + textXDelta, dy + textYDelta + 1, bgClr)
	text.Draw(target, d.dispStr, d.font, dx + textXDelta + 1, dy + textYDelta + 1, bgClr)
	text.Draw(target, d.dispStr, d.font, dx + textXDelta, dy + textYDelta, fgClr)
}
