package pok

import (
	"github.com/hajimehoshi/ebiten"
	"image/color"
	"math"
)

type Corner struct {
	x float64
	y float64
}

type Resize struct {
	outlined *ebiten.Image
	filled *ebiten.Image
	tileMap *TileMap
	holding [4]bool
	holdStart Corner
	holdEnd Corner
}

const (
	DragRadius = 8
)

func NewResize(tileMap *TileMap) Resize {
	if tileMap == nil {
		panic("TileMap was nil :/")
	}
	img, _ := ebiten.NewImage(32, 32, ebiten.FilterDefault)
	img2, _ := ebiten.NewImage(32, 32, ebiten.FilterDefault)
	outlineCircle(img, 16, 16, DragRadius, color.RGBA{255,0,255,255})
	fillCircle(img2, 16, 16, DragRadius, color.RGBA{255,0,255,255})
	r := Resize{
		img,
		img2,
		tileMap,
		[4]bool{
			false, false, false, false,
		},
		Corner{0,0},
		Corner{0,0},
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
	return [...]Corner{
		{
			-float64(r.outlined.Bounds().Max.X / 2),
			-float64(r.outlined.Bounds().Max.Y / 2),
		},
		{
			float64(r.tileMap.Width * TileSize -r.outlined.Bounds().Max.X / 2),
			float64(r.tileMap.Height * TileSize -r.outlined.Bounds().Max.Y / 2),
		},
		{
			-float64(r.outlined.Bounds().Max.X / 2),
			float64(r.tileMap.Height * TileSize -r.outlined.Bounds().Max.Y / 2),
		},
		{
			float64(r.tileMap.Width * TileSize -r.outlined.Bounds().Max.X / 2),
			-float64(r.outlined.Bounds().Max.Y / 2),
		},
	}
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
			corners[i].x,
			corners[i].y,
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

	for i := range  r.holding {
		r.holding[i] = false
	}

	corners := r.GetCorners()
	px += int(cam.X)
	py += int(cam.Y)

	for i := range corners {
		cx := corners[i].x + float64(r.outlined.Bounds().Max.X / 2)
		cy := corners[i].y + float64(r.outlined.Bounds().Max.X / 2)

		b := circleIntersect(float64(px), float64(py), cx, cy, DragRadius)
		r.holding[i] = b
		if b {
			return true
		}
	}

	return false
}

func (r *Resize) Hold() {
	
}

func (r *Resize) Release() {
	for i := range r.holding {
		r.holding[i] = false
	}
}
