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
	turnCheck int
	velocity float64
}

const (
	Static Direction = iota
	Down
	Left
	Right
	Up
)

func (dir *Direction) Inverse() Direction {
	switch *dir {
		case Down:
			return Up
		case Up:
			return Down
		case Left:
			return Right
		case Right:
			return Left
	}
	return Static
}

const(
	WalkVelocity = 1
	RunVelocity = 2
	characterMaxCycle = 8
	turnCheckLimit = 5 // in frames
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
	c.dir = dir
	c.ChangeAnim()
}

func (c *Character) ChangeAnim() {
	switch c.dir {
		case Up:
			c.Ty = 32 * 3
		case Down:
			c.Ty = 0
		case Left:
			c.Ty = 32
		case Right:
			c.Ty = 32 * 2
	}
}

//TODO: Extend later, leave Game param in for now
// Returns true if a step was just completed
func (c *Character) Update(g *Game) bool {
	if !c.isWalking {
		return false
	}

	c.Animate()
	c.Step()

	if c.frames * int(c.velocity) >= TileSize {
		c.frames = 0
		return true
	}

	return false
}

func (c *Character) Step() {
	c.frames++
	switch c.dir {
		case Up:
			c.Ty = 32 * 3
			c.Gy += -c.velocity
		case Down:
			c.Ty = 0
			c.Gy += c.velocity
		case Left:
			c.Ty = 32
			c.Gx += -c.velocity
		case Right:
			c.Ty = 32 * 2
			c.Gx += c.velocity
	}
}

func (c *Character) Animate() {
	if c.animationState % 8 == 0 {
		c.NextAnim()
	}

	c.animationState++

	if c.animationState == characterMaxCycle {
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
	switch c.dir {
		case Up:
			c.Y--
		case Down:
			c.Y++
		case Left:
			c.X--
		case Right:
			c.X++
	}
}

func (c *Character) TryStep(dir Direction, g *Game) {
	if !c.isWalking && dir == Static {
		if c.turnCheck > 0 && c.turnCheck < turnCheckLimit && c.animationState == 0 {
			c.Animate()
		}
		c.turnCheck = 0
		if c.animationState != 0 {
			c.Animate()
		} else {
			c.EndAnim()
		}
		return
	}

	if !c.isWalking {
		if c.dir == dir {
			c.turnCheck++
		}
		c.dir = dir
		c.ChangeAnim()
		if c.turnCheck >= turnCheckLimit {
			// Save old position
			ox, oy := c.X, c.Y
			c.UpdatePosition()
			// Save new position
			nx, ny := c.X, c.Y
			// Restore old position
			c.X, c.Y = ox, oy
			if g.TileIsOccupied(nx, ny, c.Z) {
				// Thud noise
				if c.animationState == characterMaxCycle -1 {
					g.Audio.PlayThud()
				}
				c.dir = dir
				c.Animate()
				c.isWalking = false
			} else {
				// Accept new position
				c.X, c.Y = nx, ny
				if c.isRunning {
					c.velocity = RunVelocity
				} else {
					c.velocity = WalkVelocity
				}
				c.isWalking = true
			}
		}
	}
}

func (c *Character) EndAnim() {
	c.animationState = 0
	c.Tx = 0
}
