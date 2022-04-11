package main

import(
	"errors"
	"image"
	"image/color"
	_ "image/png"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/atemmel/pok/pkg/constants"
	"github.com/atemmel/pok/pkg/pok"
	"math"
)

var( 
	MarkColor = color.RGBA{255, 0, 255, 255}
)

type Cropper struct {
	renderer pok.Renderer
	image *ebiten.Image
	markImg *ebiten.Image
	marks []Mark
	isMovingCamera bool
	isHolding bool
	isMovingMarks bool
	clickedIndex int
	cx, cy int
	sx, sy int
}

type Mark struct {
	X, Y float64
}

func NewCropper() *Cropper {
	image, _, err := ebitenutil.NewImageFromFile("resources/images/overworld/buildings.png")
	if err != nil {
		panic(err)
	}

	const markDim = 8
	mark := ebiten.NewImage(markDim, markDim)
	mark.Fill(MarkColor)

	marks := make([]Mark, 4)

	marks[0].X = 0
	marks[0].Y = 0

	marks[1].X = 64
	marks[1].Y = 0

	marks[2].X = 0
	marks[2].Y = 64

	marks[3].X = 64
	marks[3].Y = 64

	return &Cropper{
		renderer: pok.NewRenderer(
			constants.WindowSizeX,
			constants.WindowSizeY,
			1.0),
		image: image,
		markImg: mark,
		marks: marks,
		isMovingCamera: false,
		cx: 0,
		cy: 0,
		sx: 0,
		sy: 0,
		clickedIndex: -1,
	}
}

func (c *Cropper) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return errors.New("Clean exit")
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonMiddle) {
		// start hold
		c.isMovingCamera = true
		c.cx, c.cy = ebiten.CursorPosition()
	} else if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonMiddle) {
		// stop hold
		c.isMovingCamera = false
	}

	if c.isMovingCamera {
		x, y := ebiten.CursorPosition()
		dx, dy := x - c.cx, y - c.cy
		c.renderer.Cam.X -= float64(dx)
		c.renderer.Cam.Y -= float64(dy)
		c.cx = x
		c.cy = y
		return nil
	} 

	cx, cy := ebiten.CursorPosition()
	tx := int(float64(cx) / c.renderer.Cam.Scale)
	ty := int(float64(cy) / c.renderer.Cam.Scale)
	tx += int(math.Round(c.renderer.Cam.X))
	ty += int(math.Round(c.renderer.Cam.Y))

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		c.isMovingMarks = true
		c.sx, c.sy = tx, ty
	} else if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonRight) {
		c.isMovingMarks = false
	}

	if c.isMovingMarks {
		dx, dy := tx - c.sx, ty - c.sy

		for i := range c.marks {
			c.marks[i].X += float64(dx)
			c.marks[i].Y += float64(dy)
		}

		c.sx = tx
		c.sy = ty
		return nil
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		c.clickedIndex = c.indexOfCornerClicked(tx, ty)
		if c.clickedIndex != -1 {
			c.isHolding = true
		}
	} else if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		c.isHolding = false
		c.clickedIndex = -1
	}

	if c.isHolding {
		c.marks[c.clickedIndex].X = float64(tx)
		c.marks[c.clickedIndex].Y = float64(ty)

		switch c.clickedIndex {
		case 0:	// top left
			c.marks[2].X = float64(tx)
			c.marks[1].Y = float64(ty)
		case 1:	// top right
			c.marks[3].X = float64(tx)
			c.marks[0].Y = float64(ty)
		case 2: // bottom left
			c.marks[0].X = float64(tx)
			c.marks[3].Y = float64(ty)
		case 3: // bottom right
			c.marks[1].X = float64(tx)
			c.marks[2].Y = float64(ty)
		}
	}

	_, dy := ebiten.Wheel()
	if dy != 0. {
		if dy < 0 {
			if c.renderer.Cam.Scale > 0.50000001 {
				c.renderer.ZoomToCenter(c.renderer.Cam.Scale - 0.1)
			}
		} else {
			if c.renderer.Cam.Scale < 2.0 {
				c.renderer.ZoomToCenter(c.renderer.Cam.Scale + 0.1)
			}
		}
	}

	return nil
}

func (c *Cropper) indexOfCornerClicked(cx, cy int) int {
	iw, ih := c.markImg.Size()
	w2 := (c.marks[1].X - c.marks[0].X + float64(iw)) / 2
	h2 := (c.marks[2].Y - c.marks[0].Y + float64(ih)) / 2

	pt := image.Pt(cx, cy)

	// check top left quadrant
	m0 := c.marks[0]
	r0 := image.Rect(int(m0.X), int(m0.Y), int(m0.X + w2), int(m0.Y + h2))

	if pt.In(r0) {
		return 0
	}

	// check top right quadrant
	m1 := c.marks[1]
	r1 := image.Rect(int(m1.X - w2), int(m1.Y), int(m1.X), int(m1.Y + h2))

	if pt.In(r1) {
		return 1
	}

	// check bottom right quadrant
	m2 := c.marks[2]
	r2 := image.Rect(int(m2.X), int(m2.Y - h2), int(m2.X + w2), int(m2.Y))

	if pt.In(r2) {
		return 2
	}

	// check bottom right quadrant
	m3 := c.marks[3]
	r3 := image.Rect(int(m3.X - w2), int(m3.Y - h2), int(m3.X), int(m3.Y))

	if pt.In(r3) {
		return 3
	}

	return -1
}

func (c *Cropper) drawMarks(screen *ebiten.Image) {

	offsets := [4]image.Point{
		image.Pt(0, 0),
		image.Pt(8, 0),
		image.Pt(0, 8),
		image.Pt(8, 8),
	}

	for i, mark := range c.marks {
		opt := &ebiten.DrawImageOptions{}
		if c.clickedIndex == i {
			//opt.ColorM.Scale(2, 2, 2, 2)
			opt.ColorM.Translate(2, 2, 2, 2)
		}
		offset := offsets[i]
		c.renderer.Draw(&pok.RenderTarget{
			Op: opt,
			Src: c.markImg,
			SubImage: nil,
			X: mark.X - float64(offset.X),
			Y: mark.Y - float64(offset.Y),
			Z: 10,
		})
	}

	// top left to top right
	c.renderer.DrawLine(pok.DebugLine{
		X1: c.marks[0].X,
		Y1: c.marks[0].Y,
		X2: c.marks[1].X,
		Y2: c.marks[1].Y,
		Clr: MarkColor,
	})

	// bottom left to bottom right
	c.renderer.DrawLine(pok.DebugLine{
		X1: c.marks[2].X,
		Y1: c.marks[2].Y,
		X2: c.marks[3].X,
		Y2: c.marks[3].Y,
		Clr: MarkColor,
	})

	// top left to bottom left
	c.renderer.DrawLine(pok.DebugLine{
		X1: c.marks[0].X,
		Y1: c.marks[0].Y,
		X2: c.marks[2].X,
		Y2: c.marks[2].Y,
		Clr: MarkColor,
	})

	// top right to bottom right
	c.renderer.DrawLine(pok.DebugLine{
		X1: c.marks[1].X,
		Y1: c.marks[1].Y,
		X2: c.marks[3].X,
		Y2: c.marks[3].Y,
		Clr: MarkColor,
	})
}

func (c *Cropper) Draw(screen *ebiten.Image) {
	c.renderer.Draw(&pok.RenderTarget{
		Op: &ebiten.DrawImageOptions{},
		Src: c.image,
		SubImage: nil,
		X: 0,
		Y: 0,
		Z: 0,
	})

	c.drawMarks(screen)

	c.renderer.Display(screen)
}

func (c *Cropper) Layout(outsideWidth, outsideHeight int) (int, int) {
	return constants.WindowSizeX, constants.WindowSizeY
}

func main() {

	ebiten.SetWindowSize(constants.WindowSizeX, constants.WindowSizeY)
	ebiten.SetWindowTitle("replace this")
	cropper := NewCropper()

	if err := ebiten.RunGame(cropper); err != nil {
		panic(err)
	}
}
