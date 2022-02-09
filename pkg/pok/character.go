package pok

import(
	"github.com/atemmel/pok/pkg/constants"
	"github.com/atemmel/pok/pkg/textures"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
)

type Direction int

type Character struct {
	Gx float64
	Gy float64
	OffsetY float64
	X int
	Y int
	Z int
	Tx int
	Ty int

	dir Direction
	isWalking bool
	isRunning bool
	isBiking bool
	isJumping bool
	isSurfing bool
	hasUsedStrength bool
	isTraversingStaircaseDown bool
	isTraversingStaircaseUp bool
	frames int
	animationState int
	turnCheck int
	currentJumpTarget int
	velocity float64
}

const (
	DoJump = iota
	DoCollision
	DoNone
)

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
	BikeVelocity = 4
	JumpVelocity = 1
	characterMaxCycle = 8
	turnCheckLimit = 5 // in frames
)

func (c *Character) Draw(img *ebiten.Image, rend *Renderer, offsetX, offsetY float64) {
	charOpt := &ebiten.DrawImageOptions{}

	x := c.Gx + NpcOffsetX + offsetX
	y := c.Gy + NpcOffsetY + offsetY + c.OffsetY

	playerRect := image.Rect(
		c.Tx,
		c.Ty,
		c.Tx + (constants.TileSize * 2),
		c.Ty + (constants.TileSize * 2),
	)

	rend.Draw(&RenderTarget{
		charOpt,
		img,
		&playerRect,
		x,
		y,
		2,
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

	if c.isJumping {
		if c.frames * int(c.velocity) >= c.currentJumpTarget {
			c.frames = 0
			c.OffsetY = 0
			c.isJumping = false
			return true
		}
	} else if c.frames * int(c.velocity) >= constants.TileSize {
		c.frames = 0
		return true
	}

	return false
}

func (c *Character) Step() {
	c.frames++

	if c.isJumping {
		x := float64(c.frames) / float64(c.currentJumpTarget)
		c.OffsetY = (-4.0 * ((x - 0.5) * (x - 0.5)) + 1) * -8
	}

	if c.isTraversingStaircaseUp {
		if c.dir == Up || c.dir == Down {
			c.Gy -= 0.25 * c.velocity
		} else {
			c.Gy -= 0.5 * c.velocity
		}
	} else if c.isTraversingStaircaseDown {
		if c.dir == Up || c.dir == Down {
			c.Gy += 0.25 * c.velocity
		} else {
			c.Gy += 0.5 * c.velocity
		}
	}

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
			
			c.handleBoulder(nx, ny, c.Z, c.dir, g)

			if g.TileIsOccupied(nx, ny, c.Z) {
				// Thud noise
				if c.animationState == characterMaxCycle -1 {
					g.Audio.PlayThud()
				}
				c.dir = dir
				c.Animate()
				c.isWalking = false
			} else {

				containsWater := c.CoordinateContainsWater(nx, ny, g)
				// Accept new position
				if res := c.TryJumpLedge(nx, ny, g); res == DoJump {
					g.Audio.PlayPlayerJump()
					c.isJumping = true
					c.currentJumpTarget = constants.TileSize * 2
					switch c.dir {
					case Down:
						ny++
					case Right:
						nx++
					case Left:
						nx--
					}
				} else if res == DoCollision || (c.CoordinateContainsWater(nx, ny, g) && !c.isSurfing) {

					if c.animationState == characterMaxCycle -1 {
						g.Audio.PlayThud()
					}
					c.dir = dir
					c.Animate()
					c.isWalking = false
					return
				}

				c.handleStairCase(g, nx, ny)

				c.X, c.Y = nx, ny

				if !containsWater && c.isSurfing {
					g.Audio.PlayPlayerJump()
					c.isJumping = true
					c.isWalking = true
					c.velocity = WalkVelocity
					c.currentJumpTarget = constants.TileSize
					c.isSurfing = false
				}

				if c.isJumping {
					c.velocity = JumpVelocity
				} else if c.isRunning {
					c.velocity = RunVelocity
				} else if c.isBiking {
					c.velocity = BikeVelocity
				} else {
					c.velocity = WalkVelocity
				}
				c.isWalking = true
			}
		}
	}
}

func (c *Character) TryJumpLedge(nx, ny int, g *Game) int {
	if c.Z + 1 >= len(g.Ows.tileMap.Tiles) {
		return DoNone
	}

	//TODO: Check texture index as well
	isDownLedge := func(i int) bool {
		return textures.IsBase(g.Ows.tileMap.TextureIndicies[c.Z + 1][i]) && (g.Ows.tileMap.Tiles[c.Z + 1][i] == 213 || g.Ows.tileMap.Tiles[c.Z + 1][i] == 214 || g.Ows.tileMap.Tiles[c.Z + 1][i] == 215)
	}

	isRightLedge := func(i int) bool {
		return textures.IsBase(g.Ows.tileMap.TextureIndicies[c.Z + 1][i]) && (g.Ows.tileMap.Tiles[c.Z + 1][i] == 233 || g.Ows.tileMap.Tiles[c.Z + 1][i] == 241 || g.Ows.tileMap.Tiles[c.Z + 1][i] == 249)
	}

	isLeftLedge := func(i int) bool {
		return textures.IsBase(g.Ows.tileMap.TextureIndicies[c.Z + 1][i]) && (g.Ows.tileMap.Tiles[c.Z + 1][i] == 232 || g.Ows.tileMap.Tiles[c.Z + 1][i] == 240 || g.Ows.tileMap.Tiles[c.Z + 1][i] == 248)
	}

	index := g.Ows.tileMap.Index(nx, ny)
	if c.dir == Down && isDownLedge(index) {
		if g.TileIsOccupied(nx, ny + 1, c.Z) {
			return DoCollision
		}
		return DoJump
	} else if c.dir != Down && isDownLedge(index) {
		return DoCollision
	}

	if c.dir == Right && isRightLedge(index) {
		if g.TileIsOccupied(nx + 1, ny, c.Z) {
			return DoCollision
		}
		return DoJump
	} else if c.dir != Right && isRightLedge(index) {
		return DoCollision
	}

	if c.dir == Left && isLeftLedge(index) {
		if g.TileIsOccupied(nx - 1, ny, c.Z) {
			return DoCollision
		}
		return DoJump
	} else if c.dir != Left && isLeftLedge(index) {
		return DoCollision
	}

	return DoNone
}

func (c *Character) CoordinateContainsWater(x, y int, g *Game) bool {
	const innerWaterTile = 67
	index := y * g.Ows.tileMap.Width + x
	textureIndex := g.Ows.tileMap.TextureMapping[g.Ows.tileMap.TextureIndicies[c.Z][index]]

	return textures.IsWater(textureIndex) && g.Ows.tileMap.Tiles[c.Z][index] == innerWaterTile
}

func (c *Character) EndAnim() {
	c.animationState = 0
	c.Tx = 0
	c.isJumping = false
}

func (c *Character) isStairCase(x, y, z int, g *Game) bool {
	if len(g.Ows.tileMap.TextureIndicies) <= z + 1 {
		return false
	}

	index := g.Ows.tileMap.Index(x, y)
	if textures.IsStair(g.Ows.tileMap.TextureIndicies[z + 1][index]) {
		return true
	}
	return false
}

//TODO: This should not be like this
var stairRight = []int{
	171,
	193,
	215,
}

var stairLeft = []int {
	175,
	197,
	219,
}

var stairUp = []int {
	194,
	195,
}

func (c *Character) handleStairCase(g *Game, nx, ny int) {

	index := g.Ows.tileMap.Index(c.X, c.Y)
	nextIndex := g.Ows.tileMap.Index(nx, ny)

	indicies := [2]int{index, nextIndex}

	c.isTraversingStaircaseUp = false
	c.isTraversingStaircaseDown = false

	c.handleStairCaseWithOffset(g, indicies[:], 0)
	if len(g.Ows.tileMap.Tiles) > c.Z + 1 {
		c.handleStairCaseWithOffset(g, indicies[:], 1)
	} else if c.Z - 1 >= 0 {
		c.handleStairCaseWithOffset(g, indicies[:], -1)
	}

	if c.isTraversingStaircaseUp {
		c.Z += 1
	} else if c.isTraversingStaircaseDown {
		c.Z -= 1
	}
}

func (c *Character) handleStairCaseWithOffset(g *Game, indiciesToCheck []int, zOffset int) {
	for _, ind := range indiciesToCheck {
		for _, i := range stairRight {
			if g.Ows.tileMap.Tiles[c.Z + zOffset][ind] == i {
				switch c.dir {
					case Right:
						c.isTraversingStaircaseUp = true
					case Left:
						c.isTraversingStaircaseDown = true
				}
				return
			}
		}

		for _, i := range stairLeft {
			if g.Ows.tileMap.Tiles[c.Z + zOffset][ind] == i {
				switch c.dir {
					case Right:
						c.isTraversingStaircaseDown = true
					case Left:
						c.isTraversingStaircaseUp = true
				}
				return
			}
		}

		for _, i := range stairUp {
			if g.Ows.tileMap.Tiles[c.Z + zOffset][ind] == i {
				switch c.dir {
					case Up:
						c.isTraversingStaircaseUp = true
					case Down:
						c.isTraversingStaircaseDown = true
				}
			}
		}
	}
}

func (c *Character) handleBoulder(x, y, z int, dir Direction, g *Game) {
	if !c.hasUsedStrength {
		return
	}

	boulderIndex := g.Ows.tileMap.GetBoulderIndexAt(x, y, z)
	if boulderIndex == -1 {
		return
	}

	boulder := &g.Ows.tileMap.Boulders[boulderIndex]
	boulder.dir = dir
}
