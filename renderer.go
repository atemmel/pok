package main

import (
	"github.com/hajimehoshi/ebiten"
	"image"
	"sort"
)

type RenderTarget struct {
	Op *ebiten.DrawImageOptions
	Src *ebiten.Image
	SubImage *image.Rectangle	// nil if drawing entire Src
	X float64
	Y float64
	Z uint32
}

type DrawOrder []RenderTarget

func (do DrawOrder) Len() int {
	return len(do)
}

func (do DrawOrder) Swap(i, j int) {
	do[i], do[j] = do[j], do[i]
}

func (do DrawOrder) Less(i, j int) bool {
	if do[i].Z != do[j].Z {
		return do[i].Z < do[j].Z
	}
	return do[i].Y < do[j].Y
}

type Camera struct {
	X float64
	Y float64
}

type Renderer struct {
	dest *ebiten.Image
	targets []RenderTarget
	Cam Camera
}

func NewRenderer(width int, height int) Renderer {
	img, _ := ebiten.NewImage(width, height, ebiten.FilterDefault)
	return Renderer {
		img,
		make([]RenderTarget, 0),
		Camera{},
	}
}

func (r *Renderer) LookAt(x float64, y float64) {
	r.Cam.X = x
	r.Cam.Y = y
}

func (r *Renderer) Draw(target *RenderTarget) {
	r.targets = append(r.targets, *target)
}

func (r *Renderer) Display(screen *ebiten.Image) {
	sort.Sort(DrawOrder(r.targets) )
	for _, t := range r.targets {
		t.Op.GeoM.Translate(t.X, t.Y)
		if t.SubImage != nil {
			r.dest.DrawImage(t.Src.SubImage(*t.SubImage).(*ebiten.Image), t.Op) 
		} else {
			r.dest.DrawImage(t.Src, t.Op)
		}
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-r.Cam.X, -r.Cam.Y)
	screen.DrawImage(r.dest, op)
	r.targets = r.targets[:0]
}

func (r *Renderer) prepareRenderTargets() {
	
}
