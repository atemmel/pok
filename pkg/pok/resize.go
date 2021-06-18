package pok

import (
	"github.com/atemmel/pok/pkg/constants"
	"github.com/atemmel/pok/pkg/debug"
	"github.com/hajimehoshi/ebiten/v2"
	"image/color"
	"math"
	"errors"
)

type Corner struct {
	x float64
	y float64
}

type Resize struct {
	outlined *ebiten.Image
	filled *ebiten.Image
	tileMap *TileMap
	offset *Vec2
	holding [4]bool
	holdStart Corner
	holdEnd Corner
	holdIndex int
	clickStartX int
	clickStartY int
	dx int
	dy int
}

const (
	DragRadius = 8

	TopLeftCorner  = 0
	TopRightCorner = 1
	BotLeftCorner  = 2
	BotRightCorner = 3
)

func NewResize(tileMap *TileMap, offset *Vec2) Resize {
	if tileMap == nil {
		debug.Assert(errors.New("tileMap was nil :/"))
	}
	img := ebiten.NewImage(32, 32)
	img2 := ebiten.NewImage(32, 32)
	outlineCircle(img, 16, 16, DragRadius, color.RGBA{255,0,255,255})
	fillCircle(img2, 16, 16, DragRadius, color.RGBA{255,0,255,255})
	r := Resize{
		img,
		img2,
		tileMap,
		offset,
		[4]bool{
			false, false, false, false,
		},
		Corner{0,0},
		Corner{0,0},
		-1,
		0,
		0,
		0,
		0,
	}
	return r
}

func outlineCircle(img *ebiten.Image, x0, y0, r int, clr color.Color) {
	// https://en.wikipedia.org/wiki/Midpoint_circle_algorithm
	// https://stackoverflow.com/questions/51626905/drawing-circles-with-two-radius-in-golang
	x, y, dx, dy := r-1, 0, 1, 1
    err := dx - (r * 2)

    for x > y {
        img.Set(x0+x, y0+y, clr)
        img.Set(x0+y, y0+x, clr)
        img.Set(x0-y, y0+x, clr)
        img.Set(x0-x, y0+y, clr)
        img.Set(x0-x, y0-y, clr)
        img.Set(x0-y, y0-x, clr)
        img.Set(x0+y, y0-x, clr)
        img.Set(x0+x, y0-y, clr)

        if err <= 0 {
            y++
            err += dy
            dy += 2
        }
        if err > 0 {
            x--
            dx += 2
            err += dx - (r * 2)
        }
    }
}

func fillCircle(img *ebiten.Image, x0, y0, r int, clr color.Color) {
	for w := 0; w < r * 2; w++ {
		for h := 0; h < r * 2; h++ {
			dx := r - w
			dy := r - h
			if dx*dx + dy*dy < r*r {
				img.Set(x0 + dx, y0 + dy, clr)
			}
		}
	}
}

func (r *Resize) HasCorners() bool {
	return !(r.tileMap.Height == 0 || r.tileMap.Width == 0)
}

func (r *Resize) GetCorners() [4]Corner {
	corners := [...]Corner{
		{
			-float64(r.outlined.Bounds().Max.X / 2),
			-float64(r.outlined.Bounds().Max.Y / 2),
		},
		{
			float64(r.tileMap.Width * constants.TileSize -r.outlined.Bounds().Max.X / 2),
			-float64(r.outlined.Bounds().Max.Y / 2),
		},
		{
			-float64(r.outlined.Bounds().Max.X / 2),
			float64(r.tileMap.Height * constants.TileSize -r.outlined.Bounds().Max.Y / 2),
		},
		{
			float64(r.tileMap.Width * constants.TileSize -r.outlined.Bounds().Max.X / 2),
			float64(r.tileMap.Height * constants.TileSize -r.outlined.Bounds().Max.Y / 2),
		},
	}

	if r.holdIndex == -1 {
		return corners
	}

	dx := float64(r.dx / constants.TileSize) * constants.TileSize
	dy := float64(r.dy / constants.TileSize) * constants.TileSize

	corners = r.moveCorners(corners, dx, dy)

	return corners
}

func (r *Resize) Draw(rend *Renderer) {
	if !r.HasCorners() {
		return
	}

	corners := r.GetCorners()

	var target *ebiten.Image
	for i := range corners {
		if r.holding[i] {
			target = r.filled
		} else {
			target = r.outlined
		}
		rend.Draw(&RenderTarget{
			&ebiten.DrawImageOptions{},
			target,
			nil,
			corners[i].x + r.offset.X,
			corners[i].y + r.offset.Y,
			100,
		})
	}
}

func circleIntersect(x1, y1, x2, y2, r float64) bool {
	dx := (x1 - x2)
	dy := (y1 - y2)
	return math.Abs(dx*dx + dy*dy) < r*r
}

func (r *Resize) tryClick(px, py int, cam *Camera) bool {
	if !r.HasCorners() {
		return false
	}

	for i := range r.holding {
		r.holding[i] = false
	}

	corners := r.GetCorners()
	px += int(cam.X)
	py += int(cam.Y)

	for i := range corners {
		cx := (corners[i].x + float64(r.outlined.Bounds().Max.X / 2) + r.offset.X) * cam.Scale
		cy := (corners[i].y + float64(r.outlined.Bounds().Max.X / 2) + r.offset.Y) * cam.Scale

		b := circleIntersect(float64(px), float64(py), cx, cy, DragRadius)
		r.holding[i] = b
		if b {
			if !r.IsHolding() {
				r.holdIndex = i
				r.clickStartX, r.clickStartY = ebiten.CursorPosition()
			}
			return true
		}
	}

	return false
}

func (r *Resize) IsHolding() bool {
	return r.holdIndex != -1
}

func (r *Resize) Hold() {
	cx, cy := ebiten.CursorPosition()
	r.dx = cx - r.clickStartX
	r.dy = cy - r.clickStartY
}

func (r *Resize) Release() (int, int, int) {
	if !r.IsHolding() {
		return 0, 0, -1
	}

	x := r.dx / constants.TileSize
	y := r.dy / constants.TileSize
	i := r.holdIndex
	r.holding[r.holdIndex] = false
	r.holdIndex = -1

	switch i {
		case TopLeftCorner:
			return -x, -y, i
		case TopRightCorner:
			return x, -y, i
		case BotLeftCorner:
			return -x, y, i
		case BotRightCorner:
			return x, y, i
	}

	return 0, 0, -1
}

func (r *Resize) moveCorners(corners [4]Corner, dx, dy float64) [4]Corner {
	switch r.holdIndex {
		case TopLeftCorner:
			if !(dx < 0 && dy > 0) && !(dx > 0 && dy < 0) {
				corners[TopLeftCorner].x += dx
				corners[TopLeftCorner].y += dy
				if dy != 0 {
					corners[TopRightCorner].y += dy
				}
				if dx != 0 {
					corners[BotLeftCorner].x += dx
				}
			}
		case TopRightCorner:
			if !(dx > 0 && dy > 0) && !(dx < 0 && dy < 0) {
				corners[TopRightCorner].x += dx
				corners[TopRightCorner].y += dy
				if dy != 0 {
					corners[TopLeftCorner].y += dy
				}
				if dx != 0 {
					corners[BotRightCorner].x += dx
				}
			}
		case BotLeftCorner:
			if !(dx < 0 && dy < 0) && !(dx > 0 && dy > 0) {
				corners[BotLeftCorner].x += dx
				corners[BotLeftCorner].y += dy
				if dy != 0 {
					corners[BotRightCorner].y += dy
				}
				if dx != 0 {
					corners[TopLeftCorner].x += dx
				}
			}
		case BotRightCorner:
			if !(dx > 0 && dy < 0) && !(dx < 0 && dy > 0) {
				corners[BotRightCorner].x += dx
				corners[BotRightCorner].y += dy
				if dy != 0 {
					corners[BotLeftCorner].y += dy
				}
				if dx != 0 {
					corners[TopRightCorner].x += dx
				}
			}
	}

	return corners
}
