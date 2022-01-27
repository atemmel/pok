package pok

import (
	"github.com/atemmel/pok/pkg/constants"
	"github.com/atemmel/pok/pkg/textures"
	"github.com/atemmel/pok/pkg/debug"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	//"math"
	"math/rand"
)

const (
	RainVelocityX = -5
	RainVelocityY = 20
)

func CreateRainWeather(r *Renderer) *RainWeather {
	var err error
	rain := &RainWeather{}
	rain.textures[0], err = textures.LoadWithError(constants.ImagesDir + "weather/rain_1.png")
	debug.Assert(err)
	rain.textures[1], err = textures.LoadWithError(constants.ImagesDir + "weather/rain_2.png")
	debug.Assert(err)
	rain.textures[2], err = textures.LoadWithError(constants.ImagesDir + "weather/rain_3.png")
	debug.Assert(err)

	rain.renderer = r

	for i := 0; i < 32; i++ {
		rain.spawnOriginalSetOfParticles()
	}

	return rain
}

type RainWeather struct {
	renderer *Renderer
	particles []rainParticle
	textures [3]*ebiten.Image
}

type rainParticle struct {
	x, y float64
	textureIndex int
}

func (r *RainWeather) Update() {
	for i := range r.particles {
		r.particles[i].update()
	}
	r.cullParticles()
}

func (r *rainParticle) update() {
	r.x += RainVelocityX
	r.y += RainVelocityY
}

func (r *RainWeather) cullParticles() {
	view := r.renderer.Cam.AsRect()
	boxes := [3]image.Rectangle{
		r.textures[0].Bounds(),
		r.textures[1].Bounds(),
		r.textures[2].Bounds(),
	}

	for i := range r.particles {
		box := boxes[r.particles[i].textureIndex]
		top := int(r.particles[i].y) - box.Max.Y
		bot := int(r.particles[i].y) + box.Max.Y
		left := int(r.particles[i].x) + box.Max.X
		right := int(r.particles[i].x) - box.Max.X

		if top > view.Max.Y {
			r.particles[i].y -= float64(view.Dy() + box.Max.Y)
		} else if bot < view.Min.X {
			r.particles[i].y += float64(view.Dy() + box.Max.Y)
		}
		if left < view.Min.X {
			r.particles[i].x += float64(view.Dx() + box.Max.X)
		} else if right > view.Max.X {
			r.particles[i].x -= float64(view.Dx() + box.Max.X)
		}
	}
}

func (r *RainWeather) spawnOriginalSetOfParticles() {
	h := r.renderer.Cam.H
	w := r.renderer.Cam.W
	y := r.renderer.Cam.Y - h / 2
	x := r.renderer.Cam.X
	left := x - w / 2 + 200

	r.particles = append(r.particles, rainParticle{
		x: (rand.Float64() * w) + left,
		y: (rand.Float64() * h) + y,
		textureIndex: rand.Intn(len(r.textures)),
	})
}

func (r *RainWeather) Draw(rend* Renderer) {
	for i := range r.particles {
		r := &RenderTarget{
			&ebiten.DrawImageOptions{},
			r.textures[r.particles[i].textureIndex],
			nil,
			r.particles[i].x,
			r.particles[i].y,
			downPourZ,
		}
		rend.Draw(r)
	}
}
