package pok

import (
	"github.com/hajimehoshi/ebiten"
	"image/color"
)

type TransitionState struct {
	Ticks int

	file string
	exitId int
	magnitude int
	fadeFrom *ebiten.Image
	currentFade *ebiten.Image
}

const nTransitionTicks = 10

func NewTransitionState(src *ebiten.Image, file string, exitId int) *TransitionState {
	img, _ := ebiten.NewImageFromImage(src, ebiten.FilterDefault)
	fade, _ := ebiten.NewImage(img.Bounds().Max.X, img.Bounds().Max.Y, ebiten.FilterDefault)
	fade.Fill(color.RGBA{0, 0, 0, 0})
	return &TransitionState{
		0,
		file,
		exitId,
		1,
		img,
		fade,
	}
}

func (t *TransitionState) GetInputs(g *Game) error {
	return nil
}

func (t *TransitionState) Update(g *Game) error {
	t.Ticks += t.magnitude
	if t.Ticks > nTransitionTicks {
		g.Load(t.file, t.exitId)
		g.Ows.Update(g)
		g.Ows.Draw(g, t.fadeFrom)
		t.magnitude = -1;
		return nil
	} else if t.Ticks == 0 {
		g.As = &g.Ows
		return nil
	}
	scale := float64(t.Ticks) / float64(nTransitionTicks)
	t.currentFade.Fill(color.RGBA{0, 0, 0, uint8(255.0 * scale)})
	return nil
}

func (t *TransitionState) Draw(g *Game, screen *ebiten.Image) {
	screen.DrawImage(t.fadeFrom, &ebiten.DrawImageOptions{})
	screen.DrawImage(t.currentFade, &ebiten.DrawImageOptions{})
}
