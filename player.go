package main

type Player struct {
	gx float64
	gy float64
	x int
	y int
	animationState int
	frames int
	tx int
	ty int
	dir Direction
	isWalking bool
}

type Direction int

const(
	Static Direction = 0
	Down Direction = 1
	Left Direction = 2
	Right Direction = 3
	Up Direction = 4
)
