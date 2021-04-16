package pok

import(
	"github.com/hajimehoshi/ebiten"
	"image"
)

type Direction int

type Character struct {
	Gx float64
	Gy float64
	X int
	Y int
	Z int
	Tx int
	Ty int

	dir Direction
	isWalking bool
	isRunning bool
	frames int
	animationState int
	velocity float64
}

const (
	Static Direction = iota
	Down
	Left
	Right
	Up
)

const(
	TurnCheckLimit = 5	// in frames
	WalkVelocity = 1
	RunVelocity = 2
)

func (c *Character) Draw(img *ebiten.Image, rend *Renderer, offsetX, offsetY float64) {
	charOpt := &ebiten.DrawImageOptions{}

	x := c.Gx + NpcOffsetX + offsetX
	y := c.Gy + NpcOffsetY + offsetY

	playerRect := image.Rect(
		c.Tx,
		c.Ty,
		c.Tx + (TileSize * 2),
		c.Ty + (TileSize * 2),
	)

	rend.Draw(&RenderTarget{
		charOpt,
		img,
		&playerRect,
		x,
		y,
		3,
	})
}

func (c *Character) SetDirection(dir Direction) {
	switch dir {
		case Down:
			c.Ty = 0 * TileSize
		case Left:
			c.Ty = 2 * TileSize
		case Right:
			c.Ty = 4 * TileSize
		case Up:
			c.Ty = 6 * TileSize
	}

	c.dir = dir
}

//TODO: Extend later, leave Game param in for now
func (c *Character) Update(g *Game) {
	if !c.isWalking {
		return
	}

	c.Animate()
	c.Step()
}

func (c *Character) Step() {
	c.frames++
	if c.dir == Up {
		//player.Char.Ty = 34
		c.Ty = 32 * 3
		c.Gy += -c.velocity
	} else if c.dir == Down {
		c.Ty = 0
		c.Gy += c.velocity
	} else if c.dir == Left {
		//player.Char.Ty = 34 * 2
		c.Ty = 32
		c.Gx += -c.velocity
	} else if c.dir == Right {
		//player.Char.Ty = 34 * 3
		c.Ty = 32 * 2
		c.Gx += c.velocity
	}
}

func (c *Character) Animate() {
	if c.animationState % 8 == 0 {
		c.NextAnim()
	}

	c.animationState++

	if c.animationState == playerMaxCycle {
		c.animationState = 0
	}
}

func (c *Character) NextAnim() {
	c.Tx += 32
	if (c.velocity <= WalkVelocity || !c.isWalking) && c.Tx >= 32 * 4 {
		c.Tx = 0
	} else if c.velocity > WalkVelocity && c.isWalking {
		if c.Tx < 32 {
			c.Tx += 32
		}
		if c.Tx >= 32 * 4 {
			c.Tx = 0
		}
	}
}

func (c *Character) UpdatePosition() {
	if c.dir == Up {
		c.Y--
	} else if c.dir == Down {
		c.Y++
	} else if c.dir == Left {
		c.X--
	} else if c.dir == Right {
		c.X++
	}
}
