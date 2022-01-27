package pok

import (
	"github.com/atemmel/pok/pkg/constants"
	"github.com/atemmel/pok/pkg/textures"
	"github.com/atemmel/pok/pkg/debug"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"math"
	"math/rand"
)

const downPourZ = 9001;

var loadedHail = false;

func CreateHailWeather(r* Renderer) HailWeather {
	var err error
	hail := HailWeather{}

	hail.textures[0], err = textures.LoadWithError(constants.ImagesDir + "weather/hail_1.png")
	debug.Assert(err)
	hail.textures[1], err = textures.LoadWithError(constants.ImagesDir + "weather/hail_2.png")
	debug.Assert(err)
	hail.textures[2], err = textures.LoadWithError(constants.ImagesDir + "weather/hail_3.png")
	debug.Assert(err)

	hail.renderer = r

	for i := 0; i < 32; i++ {
		hail.spawnOriginalSetOfParticles()
	}

	return hail
}

type HailWeather struct {
	renderer *Renderer
	particles []hailParticle
	textures [3]*ebiten.Image
	step float64
}

type hailParticle struct {
	x, y float64
	xOffset, xScale float64
	textureIndex int
}

func (h *hailParticle) update(step float64) {
	h.x += (math.Cos((step + h.xOffset) * 0.1) * h.xScale) * 2.0 - 0.4
	h.y += 0.5
}

func (h *HailWeather) Update() {
	for i := range h.particles {
		h.particles[i].update(h.step)
	}

	h.step += 0.16667
}

func (hail *HailWeather) spawnOriginalSetOfParticles() {
	h := hail.renderer.Cam.H
	y := hail.renderer.Cam.Y - h / 2

	top := rand.Float64() * h + y
	hail.createParticleWithY(top)
}

func (hail *HailWeather) createParticle() {
	rect := hail.renderer.Cam.AsRect()
	top := rect.Min.Y - rect.Dy() / 2
	hail.createParticleWithY(float64(top))
}

func (hail *HailWeather) createParticleWithY(y float64) {
	x := hail.renderer.Cam.X
	w := hail.renderer.Cam.W
	left := x - w / 2 + 200
	hail.particles = append(hail.particles, hailParticle{
		(rand.Float64() * w) + left,
		y,
		rand.Float64() * math.Pi * 2,
		(rand.Float64() * 0.1) + 0.1,
		rand.Intn(len(hail.textures)),
	})
}

func (h *HailWeather) cullParticles() {
	view := h.renderer.Cam.AsRect()
	boxes := [3]image.Rectangle{
		h.textures[0].Bounds(),
		h.textures[1].Bounds(),
		h.textures[2].Bounds(),
	}

	for i := 0; i < len(h.particles); i++ {
		box := boxes[h.particles[i].textureIndex]
		top := int(h.particles[i].y) - box.Max.Y
		bot := int(h.particles[i].y) + box.Max.Y
		left := int(h.particles[i].x) + box.Max.X
		right := int(h.particles[i].x) - box.Max.X

		if top > view.Max.Y {
			h.particles[i].y -= float64(view.Dy() + box.Max.Y)
		} else if bot < view.Min.Y {
			h.particles[i].y += float64(view.Dy() + box.Max.Y)
		}
		if left < view.Min.X {
			h.particles[i].x += float64(view.Dx() + box.Max.X)
		} else if right > view.Max.X {
			h.particles[i].x -= float64(view.Dx() + box.Max.X)
		}

	}
}

func (h *HailWeather) Draw(rend *Renderer) {
	for i := range h.particles {
		r := &RenderTarget{
			&ebiten.DrawImageOptions{},
			h.textures[h.particles[i].textureIndex],
			nil,
			h.particles[i].x,
			h.particles[i].y,
			downPourZ,
		}
		rend.Draw(r)
	}
}
