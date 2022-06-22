package main

import(
	"encoding/json"
	"errors"
	"io/ioutil"
	"image"
	"image/color"
	"image/png"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/atemmel/pok/pkg/constants"
	"github.com/atemmel/pok/pkg/debug"
	"github.com/atemmel/pok/pkg/fonts"
	"github.com/atemmel/pok/pkg/pok"
	"github.com/sqweek/dialog"
	"math"
	"os"
	"strconv"
	"strings"
)

var(
	MarkColor = color.RGBA{255, 255, 255, 255}
)

type Cropper struct {
	// renderer
	renderer pok.Renderer

	// image to crop
	image *ebiten.Image

	// image used to display mark corner
	markImg *ebiten.Image

	// corners of mark
	marks []Mark

	// is panning camera
	isMovingCamera bool

	// is holding a mark
	isHolding bool

	// is moving marks
	isMovingMarks bool

	// index of corner clicked
	clickedIndex int

	// previous mouse click location
	cx, cy int

	// previous mouse click location (scaled)
	sx, sy int

	// saved subimages
	subImages map [int]SubImage

	// gui list of saved subimages
	guiList List

	// active mark id
	activeMarkId int

	// active file name
	file string
}

type Mark struct {
	X, Y float64
}

type SubImage struct {
	TopLeft Mark
	TopRight Mark
	BottomLeft Mark
	BottomRight Mark
}

func NewCropper() *Cropper {
	list := NewList(6, 32, 16)

	const markDim = 8
	mark := ebiten.NewImage(markDim, markDim)
	mark.Fill(MarkColor)

	marks := make([]Mark, 10)
	resetMarks(marks)

	return &Cropper{
		renderer: pok.NewRenderer(
			constants.WindowSizeX,
			constants.WindowSizeY,
			1.0),
		image: nil,
		markImg: mark,
		marks: marks,
		isMovingCamera: false,
		cx: 0,
		cy: 0,
		sx: 0,
		sy: 0,
		clickedIndex: -1,
		subImages: make(map[int]SubImage, 0),
		guiList: list,
		activeMarkId: 0,
		file: "",
	}
}

func (c *Cropper) LoadFile(str string) {
	c.SaveFile()
	image, _, err := ebitenutil.NewImageFromFile(str)
	debug.Assert(err)
	c.image = image
	c.file = str

	stateFile, _, _ := strings.Cut(c.file, ".")
	stateFile += "_state.json"

	bytes, err := ioutil.ReadFile(stateFile)

	if err != nil {
		return
	}

	c.guiList.Clear()
	c.subImages = make(map[int]SubImage)
	err = json.Unmarshal(bytes, &c.subImages)
	debug.Assert(err)

	for id := range c.subImages {
		c.guiList.Append(ListItem{
			Id: id,
			Name: "Mark no: " + strconv.Itoa(id),
		})
		c.activeMarkId = id
	}

	id := c.activeMarkId
	c.marks = subImageToMarks(c.subImages[id])
}

func (c *Cropper) SaveFile() {
	if !c.HasImage() {
		return
	}

	bytes, err := json.Marshal(c.subImages)
	debug.Assert(err)

	title, _, _ := strings.Cut(c.file, ".")
	title += "_state.json"
	debug.Assert(ioutil.WriteFile(title, bytes, 0o755))
}

func (c *Cropper) ExportFile() {
	if !c.HasImage() {
		return
	}

	base, _, _ := strings.Cut(c.file, ".")

	exportedString := ""

	for id, marks := range c.subImages {
		exported := base + strconv.Itoa(id) + ".png"
		r := image.Rect(
			int(marks.TopLeft.X),
			int(marks.TopLeft.Y),
			int(marks.BottomRight.X),
			int(marks.BottomRight.Y),
		)

		out, err := os.Create(exported)
		debug.Assert(err)
		subImage := c.image.SubImage(r)

		err = png.Encode(out, subImage)
		debug.Assert(err)

		exportedString += exported + " "
	}

	dialog.Message("Exported to files: %v", exportedString).Title("Successful export").Info()
}

func resetMarks(marks []Mark) {
	marks[0].X = 0
	marks[0].Y = 0

	marks[1].X = 64
	marks[1].Y = 0

	marks[2].X = 0
	marks[2].Y = 64

	marks[3].X = 64
	marks[3].Y = 64
}

func (c *Cropper) MarkActive() bool {
	return len(c.guiList.items) > 0
}

func (c *Cropper) HasImage() bool {
	return c.image != nil
}

func (c *Cropper) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return errors.New("")
	}

	if c.guiList.PollInputs() {
		c.activeMarkId = c.guiList.GetSelectedId()
		c.marks = subImageToMarks(c.subImages[c.activeMarkId])
		return nil
	}

	cx, cy := ebiten.CursorPosition()

	if PollButtons(cx, cy) {
		return nil
	}

	if !c.HasImage() {
		return nil
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

		ox := 0.0
		oy := 0.0

		if dx >= constants.TileSize {
			ox = -constants.TileSize
		} else if dx <= -constants.TileSize {
			ox = constants.TileSize
		}
		if dy >= constants.TileSize {
			oy = -constants.TileSize
		} else if dy <= -constants.TileSize {
			oy = constants.TileSize
		}

		for i := range c.marks {
			c.marks[i].X -= float64(ox)
			c.marks[i].Y -= float64(oy)
		}

		if ox != 0 {
			c.sx -= int(math.Round(ox))
		}
		if oy != 0 {
			c.sy -= int(math.Round(oy))
		}
		c.UpdateActiveMark()
		return nil
	}

	if c.MarkActive() {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			c.clickedIndex = c.indexOfCornerClicked(tx, ty)
			if c.clickedIndex != -1 {
				c.isHolding = true
			}
		} else if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			c.isHolding = false
			c.clickedIndex = -1
		}
	}

	if c.isHolding {

		dx := c.marks[c.clickedIndex].X - float64(tx)
		dy := c.marks[c.clickedIndex].Y - float64(ty)

		ox := 0.0
		oy := 0.0

		if dx >= constants.TileSize {
			ox = -constants.TileSize
		} else if dx <= -constants.TileSize{
			ox = constants.TileSize
		}
		if dy >= constants.TileSize {
			oy = -constants.TileSize
		} else if dy <= -constants.TileSize{
			oy = constants.TileSize
		}

		nx := c.marks[c.clickedIndex].X + ox
		ny := c.marks[c.clickedIndex].Y + oy

		switch c.clickedIndex {
		case 0:
			if nx > c.marks[1].X - constants.TileSize || ny > c.marks[2].Y - constants.TileSize {
				goto NO_MARK_MOVE
			}
		case 1:
			if nx < c.marks[0].X + constants.TileSize || ny > c.marks[3].Y - constants.TileSize {
				goto NO_MARK_MOVE
			}
		case 2:
			if nx > c.marks[3].X - constants.TileSize || ny < c.marks[0].Y + constants.TileSize {
				goto NO_MARK_MOVE
			}
		case 3:
			if nx < c.marks[2].X + constants.TileSize || ny < c.marks[1].Y + constants.TileSize{
				goto NO_MARK_MOVE
			}
		}

		c.marks[c.clickedIndex].X = nx
		c.marks[c.clickedIndex].Y = ny

		switch c.clickedIndex {
		case 0:	// top left
			c.marks[2].X += ox
			c.marks[1].Y += oy
		case 1:	// top right
			c.marks[3].X += ox
			c.marks[0].Y += oy
		case 2: // bottom left
			c.marks[0].X += ox
			c.marks[3].Y += oy
		case 3: // bottom right
			c.marks[1].X += ox
			c.marks[2].Y += oy
		}

		c.UpdateActiveMark()
	}

NO_MARK_MOVE:

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

func (c *Cropper) drawSubImage(subImage SubImage, id int, screen *ebiten.Image) {

	marks := subImageToMarks(subImage)
	magenta := color.RGBA{255, 0, 255, 255}
	gray := color.RGBA{128, 0, 128, 255}

	active := id == c.activeMarkId

	offsets := [4]image.Point{
		image.Pt(0, 0),
		image.Pt(8, 0),
		image.Pt(0, 8),
		image.Pt(8, 8),
	}

	for i, mark := range marks {
		opt := &ebiten.DrawImageOptions{}

		if active && c.clickedIndex == i { // keep the mark white
			opt.ColorM.Translate(1, 1, 1, 1)
		} else if !active { // make it gray
			opt.ColorM.Translate(-0.75, -1, -0.75, 1)
		} else { // make it magenta
			opt.ColorM.Translate(1, -1, 1, 1)
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
	l0 := MarkColor
	if active && (c.clickedIndex == 0 || c.clickedIndex == 1) {
		//l0 = white
	} else if !active {
		l0 = gray
	} else {
		l0 = magenta
	}
	c.renderer.DrawLine(pok.DebugLine{
		X1: marks[0].X,
		Y1: marks[0].Y,
		X2: marks[1].X,
		Y2: marks[1].Y,
		Clr: l0,
	})

	// bottom left to bottom right
	l1 := MarkColor
	if active && (c.clickedIndex == 3 || c.clickedIndex == 2) {
		//l1 = white
	} else if !active {
		l1 = gray
	} else {
		l1 = magenta
	}
	c.renderer.DrawLine(pok.DebugLine{
		X1: marks[2].X,
		Y1: marks[2].Y,
		X2: marks[3].X,
		Y2: marks[3].Y,
		Clr: l1,
	})

	// top left to bottom left
	l2 := MarkColor
	if active && (c.clickedIndex == 2 || c.clickedIndex == 0) {
		//l2 = white
	} else if !active {
		l2 = gray
	} else {
		l2 = magenta
	}
	c.renderer.DrawLine(pok.DebugLine{
		X1: marks[0].X,
		Y1: marks[0].Y,
		X2: marks[2].X,
		Y2: marks[2].Y,
		Clr: l2,
	})

	// top right to bottom right
	l3 := MarkColor
	if active && (c.clickedIndex == 3 || c.clickedIndex == 1) {
		//l3 = white
	} else if !active {
		l3 = gray
	} else {
		l3 = magenta
	}
	c.renderer.DrawLine(pok.DebugLine{
		X1: marks[1].X,
		Y1: marks[1].Y,
		X2: marks[3].X,
		Y2: marks[3].Y,
		Clr: l3,
	})
}

func (c *Cropper) drawSubImages(screen *ebiten.Image) {
	for id, subImage := range c.subImages {
		c.drawSubImage(subImage, id, screen)
	}
}

func (c *Cropper) Draw(screen *ebiten.Image) {
	if c.HasImage() {
		c.renderer.Draw(&pok.RenderTarget{
			Op: &ebiten.DrawImageOptions{},
			Src: c.image,
			SubImage: nil,
			X: 0,
			Y: 0,
			Z: 0,
		})
	}

	if c.MarkActive() {
		c.drawSubImages(screen)
	}

	c.renderer.Display(screen)
	DrawButtons(screen)
	c.guiList.Draw(screen)
}

func (c *Cropper) Layout(outsideWidth, outsideHeight int) (int, int) {
	return constants.DisplaySizeX, constants.DisplaySizeY
}

func (c *Cropper) UpdateActiveMark() {
	id := c.activeMarkId
	c.subImages[id] = marksToSubImage(c.marks)
	c.SaveFile()
}

func marksToSubImage(marks []Mark) SubImage {
	return SubImage{
		TopLeft: Mark{
			marks[0].X,
			marks[0].Y,
		},
		TopRight: Mark{
			marks[1].X,
			marks[1].Y,
		},
		BottomLeft: Mark{
			marks[2].X,
			marks[2].Y,

		},
		BottomRight: Mark{
			marks[3].X,
			marks[3].Y,
		},
	}
}

func subImageToMarks(subImage SubImage) []Mark{
	return []Mark{
		subImage.TopLeft,
		subImage.TopRight,
		subImage.BottomLeft,
		subImage.BottomRight,
	}
}

func (c *Cropper) NewMark() {
	if !c.HasImage() {
		return
	}

	i := len(c.guiList.items) + 1
	name := "Mark no: " + strconv.Itoa(i)

	c.activeMarkId = i

	c.guiList.Append(ListItem{
		Id: i,
		Name: name,
	})

	resetMarks(c.marks)
	c.UpdateActiveMark()
}

func setFont(path string) error {
	font, err := fonts.LoadFont(path, 16)
	InitButtons(font)
	return err
}

func setButton(c *Cropper) {
	AddButton(&Button{
		X: 6, Y: 10,
		OnClick: func() {
			c.NewMark()
		},
		Title: "Add Mark",
	})
	AddButton(&Button{
		X: 58, Y: 10,
		OnClick: func() {
			file, err := dialog.File().Title("Select image to open").Filter(".png").Load()
			if err != nil && file == "" {
				return
			} else if err != nil {
				dialog.Message("Could not open file: %s", file).Title("Error").Error()
				return
			}

			c.LoadFile(file)
		},
		Title: "Load Tileset",
	})
	AddButton(&Button{
		X: 128, Y: 10,
		OnClick: func() {
			c.ExportFile()
		},
		Title: "Export",
	})
}

func main() {
	logPath := "imagecrop.log"
	debug.InitAssert(&logPath, true)
	err := setFont(constants.FontsDir + "pokemon_pixel_font.ttf")
	debug.Assert(err)

	ebiten.SetWindowSize(constants.WindowSizeX, constants.WindowSizeY)
	ebiten.SetWindowTitle("imagecrop")
	cropper := NewCropper()

	setButton(cropper)

	defer func() {
		if err := recover(); err != nil {
			s, ok := err.(string)
			if ok {
				debug.Assert(errors.New(s))
			}
			e, ok := err.(error)
			if ok {
				debug.Assert(e)
			}
		}
	}()

	if err := ebiten.RunGame(cropper); err != nil {
		if err.Error() != "" {
			panic(err)
		}
	}
}
