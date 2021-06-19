package fonts

import(
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"io/ioutil"
)

func LoadFont(path string, size float64) (font.Face, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	tt, err := opentype.Parse(bytes)
	if err != nil {
		return nil, err
	}

	const dpi = 72
	font, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    size,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	return font, nil
}
