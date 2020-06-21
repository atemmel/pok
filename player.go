package main

type Player struct {
	Id int
	Gx float64
	Gy float64
	X int
	Y int
	Tx int
	Ty int
	Connected bool

	dir Direction
	isWalking bool
	isRunning bool
	frames int
	animationState int
	velocity float64
}

type Direction int

const(
	Static Direction = 0
	Down Direction = 1
	Left Direction = 2
	Right Direction = 3
	Up Direction = 4
)
