package pok

import (
	"github.com/hajimehoshi/ebiten/inpututil"
	"github.com/hajimehoshi/ebiten"
)

type Typewriter struct {
	Active bool
	Input string
	Query string
	callback func(string)
}

func (tw *Typewriter) Start(query string, fn func(string)) {
	tw.Active = true
	tw.Input = ""
	tw.Query = query
	tw.callback = fn
}

func (tw *Typewriter) HandleInputs() {
	tw.Input += string(ebiten.InputChars())
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		if len(tw.Input) > 0 {
			tw.Input = tw.Input[:len(tw.Input)-1]
		}
	} else if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		tw.Active = false
		tw.callback(tw.Input)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		tw.Active = false
		tw.callback("")
	}
}

func (tw *Typewriter) GetDisplayString() string {
	return tw.Query + tw.Input
}

