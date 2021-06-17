package pok

import (
	"github.com/atemmel/pok/pkg/constants"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
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
	Z int
}

type DebugLine struct {
	X1, Y1, X2, Y2 float64
	Clr color.Color
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
	W float64
	H float64
	Scale float64
}

func (c *Camera) AsRect() image.Rectangle {
	return image.Rect(
		int(c.X) - constants.TileSize,
		int(c.Y) - constants.TileSize,
		int(c.X + (c.W / c.Scale)),
		int(c.Y + (c.H / c.Scale)),
	)
}

type Renderer struct {
	dest *ebiten.Image
	targets []RenderTarget
	debugLines []DebugLine
	Cam Camera
}

func NewRenderer(screenWidth, screenHeight int, scale float64) Renderer {
	img, _ := ebiten.NewImage(screenWidth, screenHeight, ebiten.FilterDefault)
	return Renderer {
		img,
		make([]RenderTarget, 0),
		make([]DebugLine, 0),
		Camera{0, 0, float64(screenWidth), float64(screenHeight), scale},
	}
}

func (r *Renderer) LookAt(x float64, y float64) {
	r.Cam.X = x
	r.Cam.Y = y
}

func (r *Renderer) Draw(target *RenderTarget) {
	r.targets = append(r.targets, *target)
}

func (r *Renderer) DrawLine(line DebugLine) {
	r.debugLines = append(r.debugLines, line)
}

func (r *Renderer) Display(screen *ebiten.Image) {
	r.clear()
	r.cullRenderTargets()
	r.prepareRenderTargets()

	for _, t := range r.targets {
		t.Op.GeoM.Translate(t.X - r.Cam.X, t.Y - r.Cam.Y)
		t.Op.GeoM.Scale(r.Cam.Scale, r.Cam.Scale)
		if t.SubImage != nil {
			r.dest.DrawImage(t.Src.SubImage(*t.SubImage).(*ebiten.Image), t.Op)
		} else {
			r.dest.DrawImage(t.Src, t.Op)
		}
	}

	for _, d := range r.debugLines {
		x1 := (float64(d.X1) * r.Cam.Scale - r.Cam.X)
		y1 := (float64(d.Y1) * r.Cam.Scale - r.Cam.Y)
		x2 := (float64(d.X2) * r.Cam.Scale - r.Cam.X)
		y2 := (float64(d.Y2) * r.Cam.Scale - r.Cam.Y)
		ebitenutil.DrawLine(r.dest, x1, y1, x2, y2, d.Clr)
	}

	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(r.dest, op)
	r.targets = r.targets[:0]
	r.debugLines = r.debugLines[:0]
}

func (r *Renderer) prepareRenderTargets() {
	sort.Sort(DrawOrder(r.targets) )
}

func (r *Renderer) cullRenderTargets() {
	// interpret camera as a bounding box
	rect := r.Cam.AsRect()

	// iterate through targets
	for i := 0; i < len(r.targets); i++ {

		// create bounding box from target
		prospect := image.Rect(
			int(r.targets[i].X),
			int(r.targets[i].Y),
			int(r.targets[i].X),
			int(r.targets[i].Y),
		)

		if r.targets[i].SubImage != nil {
			prospect.Max.X += r.targets[i].SubImage.Dx()
			prospect.Max.Y += r.targets[i].SubImage.Dy()
		} else {
			prospect.Max.X += r.targets[i].Src.Bounds().Max.X
			prospect.Max.Y += r.targets[i].Src.Bounds().Max.Y
		}

		// if camera does not overlap target bounding box
		if !rect.Overlaps(prospect) {
			r.targets[i] = r.targets[len(r.targets) - 1] // Copy last element
			r.targets = r.targets[:len(r.targets) - 1]	// Pop back
			i--
		}
	}

	for i := 0; i < len(r.debugLines); i++ {
		prospectA := image.Point{int(r.debugLines[i].X1), int(r.debugLines[i].Y1)}
		prospectB := image.Point{int(r.debugLines[i].X2), int(r.debugLines[i].Y2)}

		if !prospectA.In(rect) && !prospectB.In(rect) {
			r.debugLines[i] = r.debugLines[len(r.debugLines) - 1]
			r.debugLines = r.debugLines[:len(r.debugLines) - 1]
			i--
		}
	}
}

func (r *Renderer) clear() {
	r.dest.Fill(color.RGBA{48, 64, 80, 255})
}

func (r *Renderer) ZoomToPoint(scale, x, y float64) {
	// undo prior scale (this assumes that the previous call contains the same x and y)
	factor := 1 - 1/r.Cam.Scale
	r.Cam.X -= (factor - 1) * x
	r.Cam.Y -= (factor - 1) * y

	// apply new scale
	r.Cam.Scale = scale
	factor = 1 - 1/r.Cam.Scale
	r.Cam.X += (factor - 1) * x
	r.Cam.Y += (factor - 1) * y
}

func (r *Renderer) ZoomToCenter(scale float64) {
	x := constants.DisplaySizeX / 2.0
	y := constants.DisplaySizeY / 2.0
	r.ZoomToPoint(scale, x, y)
}
