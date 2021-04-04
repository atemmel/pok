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
