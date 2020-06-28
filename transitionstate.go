package main

func NewTransitionState() TransitionState {
	return TransitionState{

	}
}

type TransitionState struct {
	Ticks int
}

func (t *TransitionState) GetInputs(g *Game) error {
	return nil
}

func (t *TransitionState) Update(g *Game) error {
	return nil
}

func (t *TransitionState) Draw(g *Game) {
}

