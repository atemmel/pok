package editor

import (
	"image"
	"image/color"
)

var ( 
	Fg = color.White
	FgShadow = color.Black
	Bg = color.RGBA{163, 164, 165, 255}
	Border = color.Black
	Hovering = color.RGBA{103, 104, 105, 255}
)

func CreateNeatImageWithBorder(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// fill
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			img.Set(x, y, Bg)
		}
	}

	// border
	for x := 0; x < w; x++ {
		img.Set(x, 0, Border)
		img.Set(x, h - 1, Border)
	}

	for y := 0; y < h; y++ {
		img.Set(0, y, Border)
		img.Set(w - 1, y, Border)
	}

	img.Set(0, 0, color.Transparent)
	img.Set(0, h - 1, color.Transparent)
	img.Set(w - 1, 0, color.Transparent)
	img.Set(w - 1, h - 1, color.Transparent)

	img.Set(1, 0, color.Transparent)
	img.Set(1, h - 1, color.Transparent)
	img.Set(w - 2, 0, color.Transparent)
	img.Set(w - 2, h - 1, color.Transparent)

	img.Set(0, 1, color.Transparent)
	img.Set(0, h - 2, color.Transparent)
	img.Set(w - 1, 1, color.Transparent)
	img.Set(w - 1, h - 2, color.Transparent)

	img.Set(1, 1, Border)
	img.Set(1, h - 2, Border)
	img.Set(w - 2, 1, Border)
	img.Set(w - 2, h - 2, Border)

	return img
}
