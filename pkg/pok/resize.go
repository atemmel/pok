package pok

import (
	"fmt"
	"github.com/hajimehoshi/ebiten"
	"image/color"
	"math"
)

type Corner struct {
	x float64
	y float64
}

type Resize struct {
	circle *ebiten.Image
	tileMap *TileMap
}

func NewResize(tileMap *TileMap) Resize {
	if tileMap == nil {
		panic("TileMap was nil :/")
	}
	img, _ := ebiten.NewImage(32, 32, ebiten.FilterDefault)
	fillCircle(img, 16, 16, 8, color.RGBA{255,0,255,255})
	r := Resize{
		img,
		tileMap,
	}
	return r
}

func fillCircle(img *ebiten.Image, x0, y0, r int, clr color.Color) {
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

func (r *Resize) HasCorners() bool {
	return !(r.tileMap.Height == 0 || r.tileMap.Width == 0)
}

func (r *Resize) GetCorners() [4]Corner {
	return [...]Corner{
		{
			-float64(r.circle.Bounds().Max.X / 2),
			-float64(r.circle.Bounds().Max.Y / 2),
		},
		{
			float64(r.tileMap.Width * TileSize -r.circle.Bounds().Max.X / 2),
			float64(r.tileMap.Height * TileSize -r.circle.Bounds().Max.Y / 2),
		},
		{
			-float64(r.circle.Bounds().Max.X / 2),
			float64(r.tileMap.Height * TileSize -r.circle.Bounds().Max.Y / 2),
		},
		{
			float64(r.tileMap.Width * TileSize -r.circle.Bounds().Max.X / 2),
			-float64(r.circle.Bounds().Max.Y / 2),
		},
	}
}

func (r *Resize) Draw(rend *Renderer) {
	if !r.HasCorners() {
		return
	}

	corners := r.GetCorners()

	for i := range corners {
		rend.Draw(&RenderTarget{
			&ebiten.DrawImageOptions{},
			r.circle,
			nil,
			corners[i].x,
			corners[i].y,
			100,
		})
	}
	/*
	if  r.tileMap.Height == 0 || r.tileMap.Width == 0 {
		return
	}

	rend.Draw(&RenderTarget{
		&ebiten.DrawImageOptions{},
		r.circle,
		nil,
		-float64(r.circle.Bounds().Max.X / 2),
		-float64(r.circle.Bounds().Max.Y / 2),
		100,
	})

	rend.Draw(&RenderTarget{
		&ebiten.DrawImageOptions{},
		r.circle,
		nil,
		float64(r.tileMap.Width * TileSize -r.circle.Bounds().Max.X / 2),
		float64(r.tileMap.Height * TileSize -r.circle.Bounds().Max.Y / 2),
		100,
	})

	rend.Draw(&RenderTarget{
		&ebiten.DrawImageOptions{},
		r.circle,
		nil,
		-float64(r.circle.Bounds().Max.X / 2),
		float64(r.tileMap.Height * TileSize -r.circle.Bounds().Max.Y / 2),
		100,
	})

	rend.Draw(&RenderTarget{
		&ebiten.DrawImageOptions{},
		r.circle,
		nil,
		float64(r.tileMap.Width * TileSize -r.circle.Bounds().Max.X / 2),
		-float64(r.circle.Bounds().Max.Y / 2),
		100,
	})
	*/
}

func circleIntersect(x1, x2, y1, y2, r float64) bool {
	s := math.Sqrt(x1*x1 + x2*x2)
	return s <= r
}

func (r *Resize) tryClick(cx, cy int, cam *Camera) bool {
	if !r.HasCorners() {
		return false
	}

	/*
	corners := r.GetCorners()
	cx += int(cam.X)
	cy += int(cam.Y)

	for i := range corners {
		
	}
	*/
	fmt.Println(cx,cy)

	return false
}
