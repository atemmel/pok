package pok

import (
	"errors"
	"fmt"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/inpututil"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"image"
	"image/color"
	"io/ioutil"
	"strconv"
	"strings"
	"math"
)

var selectionX int
var selectionY int
var copyBuffer = 0
var selectedTile = 0
var currentLayer = 0
var baseTextureIndex = 0
var activeObjsIndex = -1

var drawOnlyCurrentLayer = false
var drawUi = false
var activeTool = Pencil
var placedObjects []PlacedEditorObject = make([]PlacedEditorObject, 0)

const(
	IconOffsetX = 2
	IconOffsetY = 70
	IconPadding = 2

	Pencil = 0
	Eraser = 1
	Object = 2
	Bucket = 3
	Link = 4
	NIcons = 5

)

var ToolNames = [NIcons]string{
	"Pencil",
	"Eraser",
	"Object",
	"Bucket",
	"Link",
}

type Vec2 struct {
	X, Y float64
}

type Editor struct {
	tileMaps []TileMap
	tileMapOffsets  []Vec2
	activeTileMapIndex int

	activeTileMap *TileMap
	lastSavedTileMaps []TileMap
	rend Renderer
	dialog DialogBox
	grid Grid
	objectGrid ObjectGrid
	selection *ebiten.Image
	collisionMarker *ebiten.Image
	deleteableMarker *ebiten.Image
	exitMarker *ebiten.Image
	icons *ebiten.Image
	activeFile string
	nextFile string
	tw Typewriter
	clickStartX float64
	clickStartY float64
	resize Resize
	dieOnNextTick bool
}

func NewEditor() *Editor {
	var err error
	es := &Editor{}

	es.dialog = NewDialogBox()
	es.dieOnNextTick = false

	es.selection, _ = ebiten.NewImage(TileSize, TileSize, ebiten.FilterDefault)
	es.collisionMarker, _ = ebiten.NewImage(TileSize, TileSize, ebiten.FilterDefault)
	es.exitMarker, _ = ebiten.NewImage(TileSize, TileSize, ebiten.FilterDefault)
	es.deleteableMarker, _ = ebiten.NewImage(TileSize, TileSize, ebiten.FilterDefault)

	selectionClr := color.RGBA{255, 0, 0, 255}
	collisionClr := color.RGBA{255, 0, 255, 255}
	exitClr := color.RGBA{0, 0, 255, 255}
	deleteableClr := color.RGBA{150, 0, 0, 255}

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

	for p := 0; p < 4; p++ {
		for q := 0; q < 4; q++ {
			es.exitMarker.Set(p + 14, q, exitClr)
		}
	}

	for p := 0; p < 4; p++ {
		for q := 0; q < 4; q++ {
			es.deleteableMarker.Set(p + 6, q + 6, deleteableClr)
		}
	}

	es.rend = NewRenderer(DisplaySizeX, DisplaySizeY, 1)

	es.clickStartX = -1
	es.clickStartY = -1

	es.icons, _, err = ebitenutil.NewImageFromFile("./editorresources/images/editoricons.png", ebiten.FilterDefault)
	if err != nil {
		panic(err)
	}

	es.tileMaps = make([]TileMap, 0)
	es.tileMapOffsets = make([]Vec2, 0)

	//TODO: Make constant
	//basedir := "editorresources/overworldobjects/"

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
	for i := range e.tileMaps {
		offset := e.tileMapOffsets[i]
		e.tileMaps[i].DrawWithOffset(&e.rend, offset.X, offset.Y)
	}
	if drawUi && e.activeFile != "" {
		e.DrawTileMapDetail()
		e.resize.Draw(&e.rend)
	}
	e.rend.Display(screen)
	e.dialog.Draw(screen)

	if drawUi && e.activeFile != "" {
		if e.gridIsVisible() {
			e.grid.Draw(screen)
		} else if e.objectGridIsVisible() {
			e.objectGrid.Draw(screen)
		}
		e.drawIcons(screen)
	}

	debugStr := ""
	if e.activeFile == "" {
		debugStr += "(No file)"
	} else {
		debugStr += e.activeFile
	}
	debugStr += fmt.Sprintf(`
x: %f, y: %f
zoom: %d%%
%s`, e.rend.Cam.X, e.rend.Cam.Y, int(e.rend.Cam.Scale * 100), ToolNames[activeTool])
	ebitenutil.DebugPrint(screen, debugStr)
}

func (e *Editor) DrawTileMapDetail() {
	for j := range e.activeTileMap.Collision {
		if drawOnlyCurrentLayer && j != currentLayer {
			continue
		}
		for i := range e.activeTileMap.Collision[j] {
			x := float64(i % e.activeTileMap.Width) * TileSize
			y := float64(i / e.activeTileMap.Width) * TileSize

			if currentLayer == j && e.activeTileMap.Collision[j][i] {
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
		for i := range e.activeTileMap.Exits {
			e.rend.Draw(&RenderTarget{
				&ebiten.DrawImageOptions{},
				e.exitMarker,
				nil,
				float64(e.activeTileMap.Exits[i].X * TileSize),
				float64(e.activeTileMap.Exits[i].Y * TileSize),
				100,
			})
		}

		if activeTool == Eraser {
			for i := range placedObjects {
				e.rend.Draw(&RenderTarget{
					&ebiten.DrawImageOptions{},
					e.deleteableMarker,
					nil,
					float64(placedObjects[i].X * TileSize),
					float64(placedObjects[i].Y * TileSize),
					100,
				})
			}
		}

		offset := e.tileMapOffsets[e.activeTileMapIndex]
		e.rend.Draw(&RenderTarget{
			&ebiten.DrawImageOptions{},
			e.selection,
			nil,
			float64(selectionX * TileSize) + offset.X,
			float64(selectionY * TileSize) + offset.Y,
			100,
		})
	}
}

func (e *Editor) SelectTileFromMouse(cx, cy int) {
	offset := e.tileMapOffsets[e.activeTileMapIndex]
	cx, cy = e.TransformPointToCam(cx, cy)
	cx += int(e.rend.Cam.X)
	cy += int(e.rend.Cam.Y)
	cx -= int(offset.X)
	cy -= int(offset.Y)

	cx -= cx % TileSize
	cy -= cy % TileSize
	selectionX = cx / TileSize
	selectionY = cy / TileSize
	selectedTile =  selectionX + selectionY * e.activeTileMap.Width
}

func (e *Editor) TransformPointToCam(cx, cy int) (int, int) {
	ds := 1 / (e.rend.Cam.Scale)
	cx = int(float64(cx) * ds)
	cy = int(float64(cy) * ds)
	return cx, cy
}

func (e *Editor) loadFile() {
	e.dialog.Hidden = false
	e.tw.Start("Enter name of file to open:\n", func(str string) {
		e.dialog.Hidden = true
		if str == "" {
			return
		}

		e.nextFile = str
		tm := &TileMap{}
		err := tm.OpenFile(str);
		if err != nil {
			e.dialog.Hidden = false
			e.tw.Start("Could not open file " + e.tw.Input + ". Create new file? (y/n):", func(str string) {
				e.dialog.Hidden = true
				if str == "" || str == "y" || str == "Y" {
					// create new file
					tm = CreateTileMap(2, 2, listPngs("resources/images/overworld/"))
					e.updateEditorWithNewTileMap(tm)
					return
				}

			})
		} else {
			e.updateEditorWithNewTileMap(tm)
		}
	})
}

func (e *Editor) updateEditorWithNewTileMap(tileMap *TileMap) {
	e.appendTileMap(tileMap)
	e.activeFile = e.nextFile
	drawUi = true
	e.grid = NewGrid(tileMap.images[0])
	e.fillObjectGrid("editorresources/overworldobjects/")
	e.resize = NewResize(e.activeTileMap, &e.tileMapOffsets[len(e.tileMapOffsets) - 1])
}

func (e *Editor) appendTileMap(tileMap *TileMap) {
	e.lastSavedTileMaps = append(e.lastSavedTileMaps, TileMap{})
	e.tileMaps = append(e.tileMaps, *tileMap)
	e.tileMapOffsets = append(e.tileMapOffsets, Vec2{0, 0})
	e.activeTileMap = &e.tileMaps[len(e.tileMaps)-1]
}

func (e *Editor) saveFile() {
	err := e.activeTileMap.SaveToFile(e.activeFile)
	if err != nil {
		e.dialog.Hidden = false
		e.tw.Start("Could not save file " + err.Error(), func(str string) {
			e.dialog.Hidden = true
		})
	}

	copy(e.lastSavedTileMaps, e.tileMaps)
}

func (e *Editor) hasSaved() bool {
	for i := range e.tileMaps {
		opt := cmpopts.IgnoreFields(e.tileMaps[i], "images", "nTilesX")
		if !cmp.Equal(e.lastSavedTileMaps[i], e.tileMaps[i], opt) {
			return false
		}
	}
	return true
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
		index := e.getTileMapIndexAtCoord(cx, cy)
		if index != -1 {
			e.setActiveTileMap(index)
		}
		if e.gridIsVisible() && e.grid.Contains(image.Point{cx, cy}) {
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
		} else if e.objectGridIsVisible() && e.objectGrid.Contains(image.Point{cx, cy}) {
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton(0)) {
				obj := e.objectGrid.Select(cx, cy)
				if obj != -1 {
					activeObjsIndex = obj
				}
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
			copyBuffer = e.activeTileMap.Tiles[currentLayer][selectedTile]
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyV) {
		if e.selectedTileIsValid() {
			e.activeTileMap.Tiles[currentLayer][selectedTile] = copyBuffer
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) {	// Plus
		if currentLayer + 1 < len(e.activeTileMap.Tiles) {
			currentLayer++
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySlash) {	// Minus
		if currentLayer > 0 {
			currentLayer--
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		e.activeTileMap.AppendLayer()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyL) {
		if(len(e.activeTileMap.Tiles) > 1) {
			e.activeTileMap.Tiles = e.activeTileMap.Tiles[:len(e.activeTileMap.Tiles)-1]
			e.activeTileMap.Collision = e.activeTileMap.Collision[:len(e.activeTileMap.Collision)-1]
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyU) {
		drawOnlyCurrentLayer = !drawOnlyCurrentLayer
	}
}

func (e *Editor) handleMapMouseInputs() {
	_, dy := ebiten.Wheel()
	if dy != 0. {
		if dy < 0 {
			e.rend.ZoomToCenter(e.rend.Cam.Scale - 0.1)
		} else {
			e.rend.ZoomToCenter(e.rend.Cam.Scale + 0.1)
		}
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton(0)) {
		cx, cy := ebiten.CursorPosition();
		cx, cy = e.TransformPointToCam(cx, cy)
		e.resize.tryClick(cx, cy, &e.rend.Cam)
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButton(0)) && !ebiten.IsKeyPressed(ebiten.KeyControl) && !ebiten.IsKeyPressed(ebiten.KeyShift) {
		cx, cy := ebiten.CursorPosition();
		//if !e.isAlreadyClicking() && (e.resize.IsHolding() || e.resize.tryClick(cx, cy, &e.rend.Cam)) {
		if !e.isAlreadyClicking() && e.resize.IsHolding() {
			e.resize.Hold()
		} else {
			e.SelectTileFromMouse(cx, cy)
			if e.selectedTileIsValid() {
				if activeTool == Pencil {
					i := e.grid.GetIndex()
					e.activeTileMap.Tiles[currentLayer][selectedTile] = i
					e.activeTileMap.TextureIndicies[currentLayer][selectedTile] = baseTextureIndex
				} else if activeTool == Eraser {
					if ebiten.IsKeyPressed(ebiten.KeyShift) {
						col := selectedTile % e.activeTileMap.Width
						row := selectedTile / e.activeTileMap.Width
						i := HasPlacedObjectAt(placedObjects, col, row)
						if i != -1 {
							e.activeTileMap.EraseObject(placedObjects[i], &e.objectGrid.objs[placedObjects[i].Index])
							placedObjects[i] = placedObjects[len(placedObjects) - 1]
							placedObjects = placedObjects[:len(placedObjects) - 1]
						}
					} else {
						e.activeTileMap.Tiles[currentLayer][selectedTile] = -1
						e.activeTileMap.TextureIndicies[currentLayer][selectedTile] = baseTextureIndex
					}
				}
			}
		}
	} else {
		x, y, origin := e.resize.Release()
		if origin != -1 {
			e.activeTileMap.Resize(x, y, origin)
		}
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton(0)) && !ebiten.IsKeyPressed(ebiten.KeyControl) {
		cx, cy := ebiten.CursorPosition();
		e.SelectTileFromMouse(cx, cy)
		switch activeTool {
			case Object:
				if e.selectedTileIsValid() {
					obj := &e.objectGrid.objs[activeObjsIndex]
					e.activeTileMap.InsertObject(obj, activeObjsIndex, selectedTile, currentLayer, &placedObjects)
					fmt.Println(placedObjects)
				}
			case Link:
				if e.selectedTileIsValid() {
					// Open dialog
					e.dialog.Hidden = false
					e.tw.Start("Which file does this tile link to?", func(str string) {
						exit := Exit{}
						if str == "" {
							e.dialog.Hidden = true
							return
						}

						exit.Target = str

						e.tw.Start("Which id does this link have?", func(str string) {
							e.dialog.Hidden = true
							if str == "" {
								return
							}

							id, err := strconv.Atoi(str)
							if err != nil {
								return
							}

							exit.Id = id
							exit.X = selectedTile % e.activeTileMap.Width
							exit.Y = selectedTile / e.activeTileMap.Height
							e.activeTileMap.PlaceExit(exit)
						})

					})
				}
		}
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton(1)) {
		cx, cy := ebiten.CursorPosition();
		e.SelectTileFromMouse(cx, cy)
		if e.selectedTileIsValid() {
			e.activeTileMap.Collision[currentLayer][selectedTile] = !e.activeTileMap.Collision[currentLayer][selectedTile]
		}
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButton(2)) || (ebiten.IsMouseButtonPressed(ebiten.MouseButton(0)) && ebiten.IsKeyPressed(ebiten.KeyControl)) {
		cx, cy := ebiten.CursorPosition();
		if !e.isAlreadyClicking() {
			e.clickStartY = float64(cy)
			e.clickStartX = float64(cx)
		} else {
			e.rend.Cam.X -= (float64(cx) - e.clickStartX) / e.rend.Cam.Scale
			e.rend.Cam.Y -= (float64(cy) - e.clickStartY) / e.rend.Cam.Scale
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
	} else if ebiten.IsMouseButtonPressed(ebiten.MouseButton(0)) && ebiten.IsKeyPressed(ebiten.KeyShift) {
		cx, cy := ebiten.CursorPosition();
		if !e.isAlreadyClicking() {
			e.clickStartY = float64(cy)
			e.clickStartX = float64(cx)
		} else {
			index := len(e.tileMaps) - 1
			offset := &e.tileMapOffsets[index]
			offset.X += (float64(cx) - e.clickStartX) / e.rend.Cam.Scale
			offset.Y += (float64(cy) - e.clickStartY) / e.rend.Cam.Scale
			e.clickStartX = float64(cx)
			e.clickStartY = float64(cy)
		}
	} else {
		e.clickStartX = -1
		e.clickStartY = -1
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButton(0)) {
		offset := &e.tileMapOffsets[e.activeTileMapIndex]
		offset.X = math.Round(offset.X / TileSize) * TileSize
		offset.Y = math.Round(offset.Y / TileSize) * TileSize
	}
}

func (e *Editor) isAlreadyClicking() bool {
	return e.clickStartX != -1 && e.clickStartY != -1
}

func (e *Editor) selectedTileIsValid() bool {
	return 0 <= selectedTile && selectedTile < len(e.activeTileMap.Tiles[currentLayer])
	//cx, cy := ebiten.CursorPosition()
	//return e.getTileMapIndexAtCoord(cx, cy) != -1
}

func (e *Editor) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return DisplaySizeX, DisplaySizeY
}

func (e *Editor) gridIsVisible() bool {
	return activeTool == Pencil || activeTool == Bucket
}

func (e *Editor) objectGridIsVisible() bool {
	return activeTool == Object
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

func (e *Editor) fillObjectGrid(dir string) {
	objs, err := ReadAllObjects(dir)
	if err != nil {
		panic(err)
	}

	for i := range objs {
		for j := range e.activeTileMap.Textures {
			if e.activeTileMap.Textures[j] == objs[i].Texture {
				objs[i].textureIndex = j
			}
		}
	}

	for i := range e.activeTileMap.Textures {
		if e.activeTileMap.Textures[i] == "base.png" {
			baseTextureIndex = i
		}
	}

	//fmt.Println(objs)
	e.objectGrid = NewObjectGrid(e.activeTileMap, objs)
}

func (e *Editor) setActiveTileMap(index int) {
	e.activeTileMap = &e.tileMaps[index]
	e.activeTileMapIndex = index
}

func (e *Editor) getTileMapIndexAtCoord(cx, cy int) int {
	p := image.Point{cx, cy}
	for i := range e.tileMaps {
		w := int(float64(e.tileMaps[i].Width * TileSize) * e.rend.Cam.Scale)
		h := int(float64(e.tileMaps[i].Height * TileSize) * e.rend.Cam.Scale)
		x := int(e.tileMapOffsets[i].X - e.rend.Cam.X)
		y := int(e.tileMapOffsets[i].Y - e.rend.Cam.Y)

		r := image.Rect(x, y, x + w, y + h)
		if p.In(r) {
			return i
		}
	}
	return -1
}

func listPngs(dir string) []string {
	dirs, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	valid := make([]string, 0)
	for i := range dirs {
		if dirs[i].IsDir() || !strings.HasSuffix(dirs[i].Name(), ".png") {
			continue
		}
		valid = append(valid, dirs[i].Name())
	}
	return valid
}
