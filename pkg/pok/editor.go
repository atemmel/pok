package pok

import (
	"errors"
	"fmt"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/inpututil"
	"github.com/google/go-cmp/cmp"
	"image"
	"image/color"
)

var selectionX int
var selectionY int
var copyBuffer = 0
var selectedTile = 0
var currentLayer = 0

var drawOnlyCurrentLayer = false
var drawUi = false
var activeTool = Pencil

const(
	IconOffsetX = 2
	IconOffsetY = 70
	IconPadding = 2

	Pencil = 0
	Eraser = 1
	Object = 2
	Bucket = 3
	NIcons = 4

)

var ToolNames = [NIcons]string{
	"Pencil",
	"Eraser",
	"Object",
	"Bucket",
}

type Editor struct {
	tileMap TileMap
	lastSavedTileMap TileMap
	rend Renderer
	dialog DialogBox
	grid Grid
	selection *ebiten.Image
	collisionMarker *ebiten.Image
	exitMarker *ebiten.Image
	icons *ebiten.Image
	activeFile string
	nextFile string
	tw Typewriter
	clickStartX float64
	clickStartY float64
	resize Resize
	dieOnNextTick bool

	// TODO: Refactor
	tileset *ebiten.Image
}

func NewEditor() *Editor {
	var err error
	es := &Editor{}

	es.dialog = NewDialogBox()
	es.dieOnNextTick = false

	es.selection, err = ebiten.NewImage(TileSize, TileSize, ebiten.FilterDefault)
	if err != nil {
		panic(err)
	}

	es.collisionMarker, err = ebiten.NewImage(TileSize, TileSize, ebiten.FilterDefault)
	if err != nil {
		panic(err)
	}

	es.exitMarker, err = ebiten.NewImage(TileSize, TileSize, ebiten.FilterDefault)
	if err != nil {
		panic(err)
	}

	selectionClr := color.RGBA{255, 0, 0, 255}
	collisionClr := color.RGBA{255, 0, 255, 255}
	exitClr := color.RGBA{0, 0, 255, 255}

	for p := 0; p < es.selection.Bounds().Max.X; p++ {
		es.selection.Set(p, 0, selectionClr)
		es.selection.Set(p, es.selection.Bounds().Max.Y - 1, selectionClr)
	}

	for p := 1; p < es.selection.Bounds().Max.Y - 1; p++ {
		es.selection.Set(0, p, selectionClr)
		es.selection.Set(es.selection.Bounds().Max.Y - 1, p, selectionClr)
	}

	for p := 0; p < 4; p++ {
		for q := 0; q < 4; q++ {
			es.collisionMarker.Set(p, q, collisionClr)
		}
	}

	for p:= 0; p < 4; p++ {
		for q := 0; q < 4; q++ {
			es.exitMarker.Set(p + 14, q, exitClr)
		}
	}

	//es.tileset, _, err = ebitenutil.NewImageFromFile("./resources/images/tileset1.png", ebiten.FilterDefault)
	es.tileset, _, err = ebitenutil.NewImageFromFile("./resources/images/base.png", ebiten.FilterDefault)

	if err != nil {
		panic(err)
	}

	es.grid = NewGrid(es.tileset)

	es.rend = NewRenderer(DisplaySizeX, DisplaySizeY, DisplaySizeX, DisplaySizeY)

	es.clickStartX = -1
	es.clickStartY = -1
	es.resize = NewResize(&es.tileMap)

	es.icons, _, err = ebitenutil.NewImageFromFile("./resources/images/editoricons.png", ebiten.FilterDefault)
	if err != nil {
		panic(err)
	}

	return es;
}

func (e *Editor) Update(screen *ebiten.Image) error {
	if e.tw.Active {
		e.tw.HandleInputs();
		e.dialog.SetString(e.tw.GetDisplayString());
		return nil
	}
	err := e.handleInputs()
	return err
}

func (e *Editor) Draw(screen *ebiten.Image) {
	e.tileMap.Draw(&e.rend, e.tileset)
	if drawUi {
		e.DrawTileMapDetail()
		e.resize.Draw(&e.rend)
	}
	e.rend.Display(screen)
	e.dialog.Draw(screen)

	if drawUi {
		if e.gridIsVisible() {
			e.grid.Draw(screen)
		}
		e.drawIcons(screen)
	}

	debugStr := ""
	if e.activeFile == "" {
		debugStr += "(No file)"
	} else {
		debugStr += e.activeFile
	}
	debugStr += fmt.Sprintf("\nx: %f, y: %f\n%s", e.rend.Cam.X, e.rend.Cam.Y, ToolNames[activeTool])
	ebitenutil.DebugPrint(screen, debugStr)
}

func (e *Editor) DrawTileMapDetail() {
	for j := range e.tileMap.Collision {
		if drawOnlyCurrentLayer && j != currentLayer {
			continue
		}
		for i := range e.tileMap.Collision[j] {
			x := float64(i % e.tileMap.Width) * TileSize
			y := float64(i / e.tileMap.Width) * TileSize

			if currentLayer == j && e.tileMap.Collision[j][i] {
				e.rend.Draw(&RenderTarget{
					&ebiten.DrawImageOptions{},
					e.collisionMarker,
					nil,
					x,
					y,
					100,
				})
			}
		}
	}

	if drawUi {
		for i := range e.tileMap.Exits {
			e.rend.Draw(&RenderTarget{
				&ebiten.DrawImageOptions{},
				e.exitMarker,
				nil,
				float64(e.tileMap.Exits[i].X * TileSize),
				float64(e.tileMap.Exits[i].Y * TileSize),
				100,
			})
		}

		e.rend.Draw(&RenderTarget{
			&ebiten.DrawImageOptions{},
			e.selection,
			nil,
			float64(selectionX * TileSize),
			float64(selectionY * TileSize),
			100,
		})
	}
}

func (e *Editor) SelectTileFromMouse(cx, cy int) {
	cx += int(e.rend.Cam.X)
	cy += int(e.rend.Cam.Y)
	cx -= cx % TileSize
	cy -= cy % TileSize
	selectionX = cx / TileSize
	selectionY = cy / TileSize
	selectedTile =  selectionX + selectionY * e.tileMap.Width
}

func (e *Editor) loadFile() {
	e.dialog.Hidden = false
	e.tw.Start("Enter name of file to open:\n", func(str string) {
		e.dialog.Hidden = true
		if str == "" {
			return
		}

		e.nextFile = str
		err := e.tileMap.OpenFile(str);
		if err != nil {
			e.dialog.Hidden = false
			e.tw.Start("Could not open file " + e.tw.Input + ". Create new file? (y/n):", func(str string) {
				e.dialog.Hidden = true
				if str == "" || str == "y" || str == "Y" {
					// create new file
					e.activeFile = e.nextFile
					drawUi = true
					e.tileMap = CreateTileMap(2, 2)
					return
				}

			})
		} else {
			e.activeFile = e.nextFile
			drawUi = true
		}
	})
}

func (e *Editor) saveFile() {
	err := e.tileMap.SaveToFile(e.activeFile)
	if err != nil {
		e.dialog.Hidden = false
		e.tw.Start("Could not save file " + err.Error(), func(str string) {
			e.dialog.Hidden = true
		})
	}
	e.lastSavedTileMap = e.tileMap
}

func (e *Editor) hasSaved() bool {
	return cmp.Equal(e.lastSavedTileMap, e.tileMap)
}

func (e *Editor) unsavedWorkDialog() {
	e.dialog.Hidden = false
	e.tw.Start("You have unsaved work. Are you sure you want to exit?:", func(str string) {
		e.dialog.Hidden = true
		if str == "y" || str == "Y" {
			e.dieOnNextTick = true
		}
	})
}

func (e *Editor) handleInputs() error {
	if e.dieOnNextTick {
		return errors.New("")
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if !e.hasSaved() {
			e.unsavedWorkDialog()
		} else {
			e.dieOnNextTick = true
		}
	}

	if e.activeFile != "" {
		cx, cy := ebiten.CursorPosition()
		if e.grid.Contains(image.Point{cx, cy}) {
			_, sy := ebiten.Wheel()
			if sy < 0 {
				e.grid.Scroll(ScrollDown)
			} else if sy > 0 {
				e.grid.Scroll(ScrollUp)
			}
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton(0)) {
				cx, cy := ebiten.CursorPosition()
				e.grid.Select(cx, cy)
			}
		} else if i := e.containsIcon(cx, cy); i != NIcons {
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton(0)) {
				activeTool = i
			}
		} else {
			e.handleMapMouseInputs()
		}
		e.handleMapInputs()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyI) {
		drawUi = !drawUi
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyO) {
		e.loadFile()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		e.saveFile()
	}

	return nil
}

func (e *Editor) handleMapInputs() {
	if ebiten.IsKeyPressed(ebiten.KeyC) {
		if e.selectedTileIsValid() {
			copyBuffer = e.tileMap.Tiles[currentLayer][selectedTile]
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyV) {
		if e.selectedTileIsValid() {
			e.tileMap.Tiles[currentLayer][selectedTile] = copyBuffer
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) {	// Plus
		if currentLayer + 1 < len(e.tileMap.Tiles) {
			currentLayer++
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySlash) {	// Minus
		if currentLayer > 0 {
			currentLayer--
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		e.tileMap.Tiles = append(e.tileMap.Tiles, make([]int, len(e.tileMap.Tiles[0])))
		for i := range e.tileMap.Tiles[len(e.tileMap.Tiles) - 1] {
			e.tileMap.Tiles[len(e.tileMap.Tiles)-1][i] = -1
		}
		e.tileMap.Collision = append(e.tileMap.Collision, make([]bool, len(e.tileMap.Collision[0])))
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyL) {
		if(len(e.tileMap.Tiles) > 1) {
			e.tileMap.Tiles = e.tileMap.Tiles[:len(e.tileMap.Tiles)-1]
			e.tileMap.Collision = e.tileMap.Collision[:len(e.tileMap.Collision)-1]
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyU) {
		drawOnlyCurrentLayer = !drawOnlyCurrentLayer
	}
}

func (e *Editor) handleMapMouseInputs() {
	_, dy := ebiten.Wheel()
	if dy != 0. && e.selectedTileIsValid() {
		if dy < 0 {
			e.tileMap.Tiles[currentLayer][selectedTile]--
		} else {
			e.tileMap.Tiles[currentLayer][selectedTile]++
		}
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButton(0)) && !ebiten.IsKeyPressed(ebiten.KeyControl) {
		cx, cy := ebiten.CursorPosition();
		if e.resize.IsHolding() || e.resize.tryClick(cx, cy, &e.rend.Cam) {
			e.resize.Hold()
		} else {
			e.SelectTileFromMouse(cx, cy)
			if e.selectedTileIsValid() {
				i := e.grid.GetIndex()
				e.tileMap.Tiles[currentLayer][selectedTile] = i
			}
		}
	} else {
		x, y, origin := e.resize.Release()
		e.tileMap.Resize(x, y, origin)
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton(1)) {
		cx, cy := ebiten.CursorPosition();
		e.SelectTileFromMouse(cx, cy)
		if e.selectedTileIsValid() {
			e.tileMap.Collision[currentLayer][selectedTile] = !e.tileMap.Collision[currentLayer][selectedTile]
		}
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButton(2)) || (ebiten.IsMouseButtonPressed(ebiten.MouseButton(0)) && ebiten.IsKeyPressed(ebiten.KeyControl)) {
		cx, cy := ebiten.CursorPosition();
		if(e.clickStartX == -1 && e.clickStartY == -1) {
			e.clickStartY = float64(cy)
			e.clickStartX = float64(cx)
		} else {
			e.rend.Cam.X -= (float64(cx) - e.clickStartX)
			e.rend.Cam.Y -= (float64(cy) - e.clickStartY)
			e.clickStartX = float64(cx)
			e.clickStartY = float64(cy)
		}
		/*
		e.SelectTileFromMouse(cx, cy)
		if 0 <= selectedTile && selectedTile < len(e.tileMap.Tiles[currentLayer]) {
			if i := e.tileMap.HasExitAt(selectionX, selectionY, currentLayer); i != -1 {
				e.tileMap.Exits[i] = e.tileMap.Exits[len(e.tileMap.Exits) - 1]
				e.tileMap.Exits = e.tileMap.Exits[:len(e.tileMap.Exits) - 1]
			} else {
				e.tileMap.Exits = append(e.tileMap.Exits, Exit{
					"",
					0,
					selectionX,
					selectionY,
					currentLayer,
				})
			}
		}
		*/
	} else {
		e.clickStartX = -1
		e.clickStartY = -1
	}
}

func (e *Editor) selectedTileIsValid() bool {
	return 0 <= selectedTile && selectedTile < len(e.tileMap.Tiles[currentLayer])
}

func (e *Editor) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return DisplaySizeX, DisplaySizeY
}

func (e *Editor) gridIsVisible() bool {
	return activeTool == Pencil || activeTool == Bucket
}

func (e *Editor) drawIcons(screen *ebiten.Image) {
	w, h := e.icons.Size()
	h /= NIcons
	for i := 0; i < NIcons; i++ {
		opt := &ebiten.DrawImageOptions{}
		opt.GeoM.Translate(IconOffsetX, IconOffsetY + float64(i * (h + IconPadding)))
		r := image.Rect(0, i * h, w, i * h + h)
		screen.DrawImage(e.icons.SubImage(r).(*ebiten.Image), opt)
	}
}

func (e *Editor) containsIcon(x, y int) int {
	w, h := e.icons.Size()
	h /= NIcons
	p := image.Point{x, y}

	for i := 0; i < NIcons; i++ {
		r := image.Rect(IconOffsetX, IconOffsetY + i * h, w, IconOffsetY + i * h + h)
		if p.In(r) {
			return i
		}
	}

	return NIcons
}
