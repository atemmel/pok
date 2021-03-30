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

const(
	Static Direction = 0
	Down Direction = 1
	Left Direction = 2
	Right Direction = 3
	Up Direction = 4

	TurnCheckLimit = 5	// in frames
	WalkVelocity = 1
	RunVelocity = 2
)

func (c *Character) Draw(img *ebiten.Image, rend *Renderer) {
	charOpt := &ebiten.DrawImageOptions{}

	x := c.Gx + NpcOffsetX + 0.5
	y := c.Gy + NpcOffsetY

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
