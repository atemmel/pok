package pok

import (
	"github.com/hajimehoshi/ebiten"
)

type Direction int

type Player struct {
	Id int
	Gx float64
	Gy float64
	X int
	Y int
	Z int
	Tx int
	Ty int
	Connected bool
	Location string

	dir Direction
	isWalking bool
	isRunning bool
	frames int
	animationState int
	velocity float64
}

const(
	Static Direction = 0
	Down Direction = 1
	Left Direction = 2
	Right Direction = 3
	Up Direction = 4

	TurnCheckLimit = 5	// in frames

	playerMaxCycle = 8
	playerWalkVelocity = 1
	playerRunVelocity = 2
	playerOffsetX = 13 - TileSize
	playerOffsetY = 0 - TileSize
)

var turnCheck = 0

func (player *Player) TryStep(dir Direction, g *Game) {
	if !player.isWalking && dir == Static {
		if turnCheck > 0 && turnCheck < TurnCheckLimit &&
			player.animationState == 0 {
			player.Animate()
		}
		turnCheck = 0
		if player.animationState != 0 {
			player.Animate()
		} else {
			player.EndAnim()
		}
		return
	}

	if !player.isWalking {
		if player.dir == dir {
			turnCheck++
		}
		player.dir = dir
		player.ChangeAnim()
		if turnCheck >= TurnCheckLimit {
			ox, oy := player.X, player.Y
			player.UpdatePosition()
			if g.TileIsOccupied(player.X, player.Y, player.Z) {
				player.X, player.Y = ox, oy	// Restore position
				// Thud noise
				if player.animationState == playerMaxCycle -1 {
					g.Audio.PlayThud()
				}
				player.dir = dir
				player.Animate()
				player.isWalking = false
			} else {
				if player.isRunning {
					player.velocity = playerRunVelocity
				} else {
					player.velocity = playerWalkVelocity
				}
				player.isWalking = true
			}
		}
	}
}

func (player *Player) Update(g *Game) {
	if !player.isWalking {
		return
	}

	player.Animate()
	player.Step(g)
}

func (player *Player) Step(g *Game) {
	player.frames++
	if player.dir == Up {
		player.Ty = 34
		player.Gy += -player.velocity
	} else if player.dir == Down {
		player.Ty = 0
		player.Gy += player.velocity
	} else if player.dir == Left {
		player.Ty = 34 * 2
		player.Gx += -player.velocity
	} else if player.dir == Right {
		player.Ty = 34 * 3
		player.Gx += player.velocity
	}

	if player.frames * int(player.velocity) >= TileSize {
		player.isWalking = false
		player.frames = 0
		if i := g.Ows.tileMap.HasExitAt(player.X, player.Y, player.Z); i > -1 {
			if g.Ows.tileMap.Exits[i].Target != "" {
				img, _ := ebiten.NewImage(DisplaySizeX, DisplaySizeY, ebiten.FilterDefault);
				g.As.Draw(g, img)
				g.As = NewTransitionState(img, TileMapDir + g.Ows.tileMap.Exits[i].Target, g.Ows.tileMap.Exits[i].Id)
				g.Audio.PlayDoor()
			}
		}
	}
}

func (player *Player) Animate() {
	if player.animationState % 8 == 0 {
		player.NextAnim()
	}
	player.animationState++

	if player.animationState == playerMaxCycle {
		player.animationState = 0
	}
}

func (player *Player) NextAnim() {
	player.Tx += 34
	if (player.velocity <= playerWalkVelocity || !player.isWalking) && player.Tx >= 34 * 4 {
		player.Tx = 0
	} else if player.velocity > playerWalkVelocity && player.isWalking {
		if player.Tx < 170 {
			player.Tx += 170
		}
		if player.Tx >= 170 + 34 * 4 {
			player.Tx = 170
		}
	}
}

func (player *Player) ChangeAnim() {
	if player.dir == Up {
		player.Ty = 34
	} else if player.dir == Down {
		player.Ty = 0
	} else if player.dir == Left {
		player.Ty = 34 * 2
	} else if player.dir == Right {
		player.Ty = 34 * 3
	}
}

func (player *Player) EndAnim() {
	player.animationState = 0
	player.Tx = 0
}

func (player *Player) UpdatePosition() {
	if player.dir == Up {
		player.Y--
	} else if player.dir == Down {
		player.Y++
	} else if player.dir == Left {
		player.X--
	} else if player.dir == Right {
		player.X++
	}
}
