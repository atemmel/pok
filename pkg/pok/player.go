package pok

import (
	"github.com/hajimehoshi/ebiten"
)

type Player struct {
	Id int
	Char Character
	Connected bool
	Location string
}

const(
	playerMaxCycle = 8
	playerOffsetX = NpcOffsetX
	playerOffsetY = NpcOffsetY
)

var turnCheck = 0

func (player *Player) TryStep(dir Direction, g *Game) {
	if !player.Char.isWalking && dir == Static {
		if turnCheck > 0 && turnCheck < TurnCheckLimit &&
			player.Char.animationState == 0 {
			player.Animate()
		}
		turnCheck = 0
		if player.Char.animationState != 0 {
			player.Animate()
		} else {
			player.EndAnim()
		}
		return
	}

	if !player.Char.isWalking {
		if player.Char.dir == dir {
			turnCheck++
		}
		player.Char.dir = dir
		player.ChangeAnim()
		if turnCheck >= TurnCheckLimit {
			ox, oy := player.Char.X, player.Char.Y
			player.UpdatePosition()
			if g.TileIsOccupied(player.Char.X, player.Char.Y, player.Char.Z) {
				player.Char.X, player.Char.Y = ox, oy	// Restore position
				// Thud noise
				if player.Char.animationState == playerMaxCycle -1 {
					g.Audio.PlayThud()
				}
				player.Char.dir = dir
				player.Animate()
				player.Char.isWalking = false
			} else {
				if player.Char.isRunning {
					player.Char.velocity = RunVelocity
				} else {
					player.Char.velocity = WalkVelocity
				}
				player.Char.isWalking = true
			}
		}
	}
}

func (player *Player) Update(g *Game) {
	if !player.Char.isWalking {
		return
	}

	player.Animate()
	player.Step(g)
}

func (player *Player) Step(g *Game) {
	player.Char.frames++
	if player.Char.dir == Up {
		//player.Char.Ty = 34
		player.Char.Ty = 32 * 3
		player.Char.Gy += -player.Char.velocity
	} else if player.Char.dir == Down {
		player.Char.Ty = 0
		player.Char.Gy += player.Char.velocity
	} else if player.Char.dir == Left {
		//player.Char.Ty = 34 * 2
		player.Char.Ty = 32
		player.Char.Gx += -player.Char.velocity
	} else if player.Char.dir == Right {
		//player.Char.Ty = 34 * 3
		player.Char.Ty = 32 * 2
		player.Char.Gx += player.Char.velocity
	}

	if player.Char.frames * int(player.Char.velocity) >= TileSize {
		player.Char.isWalking = false
		player.Char.frames = 0
		if i := g.Ows.tileMap.HasExitAt(player.Char.X, player.Char.Y, player.Char.Z); i > -1 {
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
	if player.Char.animationState % 8 == 0 {
		player.NextAnim()
	}
	player.Char.animationState++

	if player.Char.animationState == playerMaxCycle {
		player.Char.animationState = 0
	}
}

func (player *Player) NextAnim() {
	player.Char.Tx += 32
	if (player.Char.velocity <= WalkVelocity || !player.Char.isWalking) && player.Char.Tx >= 32 * 4 {
		player.Char.Tx = 0
	} else if player.Char.velocity > WalkVelocity && player.Char.isWalking {
		if player.Char.Tx < 32 {
			player.Char.Tx += 32
		}
		if player.Char.Tx >= 32 * 4 {
			player.Char.Tx = 0
		}
	}
}

func (player *Player) ChangeAnim() {
	if player.Char.isRunning {
		activePlayerImg = playerRunningImg
	} else {
		activePlayerImg = playerImg
	}

	if player.Char.dir == Up {
		player.Char.Ty = 32 * 3
	} else if player.Char.dir == Down {
		player.Char.Ty = 0
	} else if player.Char.dir == Left {
		player.Char.Ty = 32
	} else if player.Char.dir == Right {
		player.Char.Ty = 32 * 2
	}
}

func (player *Player) EndAnim() {
	player.Char.animationState = 0
	player.Char.Tx = 0
}

func (player *Player) UpdatePosition() {
	if player.Char.dir == Up {
		player.Char.Y--
	} else if player.Char.dir == Down {
		player.Char.Y++
	} else if player.Char.dir == Left {
		player.Char.X--
	} else if player.Char.dir == Right {
		player.Char.X++
	}
}
