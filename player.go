package main

type Player struct {
	Gx float64
	Gy float64
	X int
	Y int
	AnimationState int
	Frames int
	Tx int
	Ty int
	Dir Direction
	IsWalking bool
}

type Direction int

const(
	Static Direction = 0
	Down Direction = 1
	Left Direction = 2
	Right Direction = 3
	Up Direction = 4
)
