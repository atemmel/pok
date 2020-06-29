package main

import (
	"github.com/hajimehoshi/ebiten"
	"image"
	"image/color"
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
	W int
	H int
}

type Renderer struct {
	dest *ebiten.Image
	targets []RenderTarget
	Cam Camera
}

func NewRenderer(destWidth int, destHeight int, screenWidth int, screenHeight int) Renderer {
	img, _ := ebiten.NewImage(destWidth, destHeight, ebiten.FilterDefault)
	return Renderer {
		img,
		make([]RenderTarget, 0),
		Camera{0, 0, screenWidth, screenHeight},
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
	r.clear()
	r.cullRenderTargets()
	r.prepareRenderTargets()

	for _, t := range r.targets {
		//t.Op.GeoM.Translate(t.X, t.Y)
		t.Op.GeoM.Translate(t.X -r.Cam.X, t.Y -r.Cam.Y)
		if t.SubImage != nil {
			r.dest.DrawImage(t.Src.SubImage(*t.SubImage).(*ebiten.Image), t.Op) 
		} else {
			r.dest.DrawImage(t.Src, t.Op)
		}
	}

	op := &ebiten.DrawImageOptions{}

	//op.GeoM.Translate(-r.Cam.X, -r.Cam.Y)
	screen.DrawImage(r.dest, op)
	r.targets = r.targets[:0]
	//return r.dest
}

func (r *Renderer) prepareRenderTargets() {
	sort.Sort(DrawOrder(r.targets) )
}

func (r *Renderer) cullRenderTargets() {
	rect := image.Rect(int(r.Cam.X), int(r.Cam.Y), int(r.Cam.X) + r.Cam.W, int(r.Cam.Y) + r.Cam.H)
	for i := 0; i < len(r.targets); i++ {
		prospect := image.Rect(
			int(r.targets[i].X),
			int(r.targets[i].Y),
			int(r.targets[i].X),
			int(r.targets[i].Y),
		)

		if r.targets[i].SubImage != nil {
			prospect.Max.X += r.targets[i].SubImage.Max.X - r.targets[i].SubImage.Min.X
			prospect.Max.Y += r.targets[i].SubImage.Max.Y - r.targets[i].SubImage.Min.Y
		} else {
			prospect.Max.X += r.targets[i].Src.Bounds().Max.X
			prospect.Max.Y += r.targets[i].Src.Bounds().Max.Y
		}

		if !rect.Overlaps(prospect) {
			r.targets[i] = r.targets[len(r.targets) - 1] // Copy last element
			r.targets = r.targets[:len(r.targets) - 1]	// Pop back
		}
	}
}

func (r *Renderer) clear() {
	r.dest.Fill(color.Black)
}
