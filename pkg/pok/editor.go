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
	//"strconv"
	"log"
	"strings"
	"math"
)

var selectionX int
var selectionY int
var copyBuffer = 0
var selectedTile = 0
var currentLayer = 0
var baseTextureIndex = 0
var activeObjsIndex = 0
var activeAtiIndex = -1

var drawOnlyCurrentLayer = false
var drawUi = false
var activeTool = Pencil
var placedObjects [][]PlacedEditorObject = make([][]PlacedEditorObject, 0)
var linkBegin *LinkData

type LinkData struct {
	X, Y int
	TileMapIndex int
}

const(
	IconOffsetX = 2
	IconOffsetY = 70
	IconPadding = 2
)

// Tools
const(
	Pencil = iota
	Eraser
	Object
	Bucket
	Link
	AutoTile
	Tree
	PlaceNpc
	NIcons
)

var ToolNames = [NIcons]string{
	"Pencil",
	"Eraser",
	"Object",
	"Bucket",
	"Link",
	"Autotile",
	"Tree",
	"Npc",
}

type Vec2 struct {
	X, Y float64
}

type Editor struct {
	tileMaps []*TileMap
	tileMapOffsets  []*Vec2
	activeTileMapIndex int

	activeTileMap *TileMap
	lastSavedTileMaps []*TileMap
	rend Renderer
	dialog DialogBox
	grid Grid
	objectGrid ObjectGrid
	selection *ebiten.Image
	collisionMarker *ebiten.Image
	deleteableMarker *ebiten.Image
	exitMarker *ebiten.Image
	icons *ebiten.Image
	activeFiles []string
	nextFile string
	tw Typewriter
	clickStartX float64
	clickStartY float64
	resizers []Resize
	autoTileInfo []AutoTileInfo
	autoTileGrid AutoTileGrid
	npcImages []*ebiten.Image
	npcImagesStrings []string
	npcGrid NpcGrid
	dieOnNextTick bool
}

func NewEditor() *Editor {
	var err error
	es := &Editor{}

	es.dialog = NewDialogBox()
	es.dialog.speed = TextInstant
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

	es.icons, _, err = ebitenutil.NewImageFromFile(EditorImagesDir + "editoricons.png", ebiten.FilterDefault)
	if err != nil {
		panic(err)
	}

	es.tileMaps = make([]*TileMap, 0)
	es.tileMapOffsets = make([]*Vec2, 0)

	es.npcImagesStrings = listPngs(CharacterImagesDir)
	for i := range es.npcImagesStrings {
		es.npcImagesStrings[i] = CharacterImagesDir + es.npcImagesStrings[i]
	}
	es.npcImages = loadImages(es.npcImagesStrings)
	es.npcGrid = NewNpcGrid(es.npcImages)

	return es;
}

func (e *Editor) Update(screen *ebiten.Image) error {
	e.dialog.Update()
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
	if drawUi && len(e.activeFiles) != 0 {
		e.drawLinksFromActiveTileMap()
		e.DrawTileMapDetail()
		//e.resizers[e.activeTileMapIndex].Draw(&e.rend)
		for i := range e.resizers {
			e.resizers[i].Draw(&e.rend)
		}
	}
	e.rend.Display(screen)
	e.dialog.Draw(screen)

	if drawUi && len(e.activeFiles) != 0 {
		if e.gridIsVisible() {
			e.grid.Draw(screen)
		} else if e.objectGridIsVisible() {
			e.objectGrid.Draw(screen)
		} else if e.autoTileGridIsVisible() {
			e.autoTileGrid.Draw(screen)
		} else if e.npcGridIsVisible() {
			e.npcGrid.Draw(screen)
		}
		e.drawIcons(screen)
	}

	debugStr := ""
	if len(e.activeFiles) == 0 {
		debugStr += "(No files)"
	} else {
		debugStr += e.activeFiles[e.activeTileMapIndex]
	}
	debugStr += fmt.Sprintf(`
x: %f, y: %f
zoom: %d%%
%s`, e.rend.Cam.X, e.rend.Cam.Y, int(e.rend.Cam.Scale * 100), ToolNames[activeTool])
	ebitenutil.DebugPrint(screen, debugStr)
}

func (e *Editor) DrawTileMapDetail() {
	offset := e.tileMapOffsets[e.activeTileMapIndex]
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
					x + offset.X,
					y + offset.Y,
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
				float64(e.activeTileMap.Exits[i].X * TileSize) + offset.X,
				float64(e.activeTileMap.Exits[i].Y * TileSize) + offset.Y,
				100,
			})
		}

		if activeTool == Eraser {
			for i := range placedObjects[e.activeTileMapIndex] {
				e.rend.Draw(&RenderTarget{
					&ebiten.DrawImageOptions{},
					e.deleteableMarker,
					nil,
					float64(placedObjects[e.activeTileMapIndex][i].X * TileSize) + offset.X,
					float64(placedObjects[e.activeTileMapIndex][i].Y * TileSize) + offset.Y,
					100,
				})
			}
		}

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
					tm = CreateTileMap(2, 2, listPngs(TileMapImagesDir))
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
	e.activeFiles = append(e.activeFiles, e.nextFile)
	drawUi = true
	e.grid = NewGrid(tileMap.images[0], TileSize)
	e.fillObjectGrid(OverworldObjectsDir)
	var err error
	e.autoTileInfo, err = ReadAllAutoTileInfo(AutotileInfoDir)
	e.autoTileGrid = NewAutoTileGrid(tileMap.images[0], tileMap.nTilesX[0], e.autoTileInfo)
	if err != nil {
		panic(err)
	}
}

func (e *Editor) appendTileMap(tileMap *TileMap) {
	placedObjects = append(placedObjects, make([]PlacedEditorObject, 0))
	e.lastSavedTileMaps = append(e.lastSavedTileMaps, &TileMap{})
	e.tileMaps = append(e.tileMaps, tileMap)
	e.tileMapOffsets = append(e.tileMapOffsets, &Vec2{0, 0})
	e.activeTileMap = e.tileMaps[len(e.tileMaps)-1]
	e.resizers = append(e.resizers, NewResize(e.tileMaps[len(e.tileMaps)-1], e.tileMapOffsets[len(e.tileMapOffsets) - 1]))
}

func (e *Editor) saveFile() {
	err := e.activeTileMap.SaveToFile(e.activeFiles[e.activeTileMapIndex])
	if err != nil {
		e.dialog.Hidden = false
		e.tw.Start("Could not save file " + err.Error(), func(str string) {
			e.dialog.Hidden = true
		})
	}

	//TODO: Rethink this, save each file individually or all at once?
	for i := range e.tileMaps {
		tm := *e.tileMaps[i]
		e.lastSavedTileMaps[i] = &tm
	}
}

func (e *Editor) hasSaved() bool {
	for i := range e.tileMaps {
		opt := cmpopts.IgnoreFields(*e.tileMaps[i], "images", "nTilesX", "npcImages", "npcImagesStrings", "npcs")
		if !cmp.Equal(*e.lastSavedTileMaps[i], *e.tileMaps[i], opt) {
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

	if len(e.activeFiles) != 0 {
		cx, cy := ebiten.CursorPosition()
		index := e.getTileMapIndexAtCoord(cx, cy)
		if index != -1 && !e.isAlreadyClicking() {
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
				//cx, cy := ebiten.CursorPosition()
				e.grid.Select(cx, cy)
			}
		} else if e.autoTileGridIsVisible() && e.autoTileGrid.Contains(image.Point{cx, cy}) {
			_, sy := ebiten.Wheel()
			if sy < 0 {
				e.autoTileGrid.Scroll(ScrollDown)
			} else if sy > 0 {
				e.autoTileGrid.Scroll(ScrollUp)
			}
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton(0)) {
				e.autoTileGrid.Select(cx, cy)
			}
		} else if e.npcGridIsVisible() && e.npcGrid.Contains(image.Point{cx, cy}) {
			_, sy := ebiten.Wheel()
			if sy < 0 {
				e.npcGrid.Scroll(ScrollDown)
			} else if sy > 0 {
				e.npcGrid.Scroll(ScrollUp)
			}
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton(0)) {
				e.npcGrid.Select(cx, cy)
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
				e.switchActiveTool(i)
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
		e.resizers[e.activeTileMapIndex].tryClick(cx, cy, &e.rend.Cam)
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButton(0)) && !ebiten.IsKeyPressed(ebiten.KeyControl) {
		cx, cy := ebiten.CursorPosition();
		if !e.isAlreadyClicking() && e.resizers[e.activeTileMapIndex].IsHolding() {
			e.resizers[e.activeTileMapIndex].Hold()
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
						i := HasPlacedObjectAt(placedObjects[e.activeTileMapIndex], col, row)
						if i != -1 {
							e.activeTileMap.EraseObject(placedObjects[e.activeTileMapIndex][i], &e.objectGrid.objs[placedObjects[e.activeTileMapIndex][i].Index])
							placedObjects[e.activeTileMapIndex][i] = placedObjects[e.activeTileMapIndex][len(placedObjects[e.activeTileMapIndex]) - 1]
							placedObjects[e.activeTileMapIndex] = placedObjects[e.activeTileMapIndex][:len(placedObjects[e.activeTileMapIndex]) - 1]
						}
					} else {
						e.activeTileMap.Tiles[currentLayer][selectedTile] = -1
						e.activeTileMap.TextureIndicies[currentLayer][selectedTile] = baseTextureIndex
					}
				} else if activeTool == AutoTile {
					ati := &e.autoTileInfo[e.autoTileGrid.GetIndex()]
					DecideTileIndicies(e.activeTileMap, selectedTile, currentLayer, baseTextureIndex, ati)
				} else if activeTool == Tree {
					//TODO: perform tree logic
				} else if activeTool == PlaceNpc {
					e.tryPlaceNpc()
				}
			}
		}
	} else {
		x, y, origin := e.resizers[e.activeTileMapIndex].Release()
		if origin != -1 {
			e.activeTileMap.Resize(x, y, origin)
			e.removeInvalidLinks()

			if origin == TopLeftCorner || origin == TopRightCorner {
				e.tileMapOffsets[e.activeTileMapIndex].Y -= float64(y * TileSize)
			}

			if origin == TopLeftCorner || origin == BotLeftCorner {
				e.tileMapOffsets[e.activeTileMapIndex].X -= float64(x * TileSize)
			}
		}
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton(0)) && !ebiten.IsKeyPressed(ebiten.KeyControl) {
		cx, cy := ebiten.CursorPosition();
		e.SelectTileFromMouse(cx, cy)
		switch activeTool {
			case Object:
				if e.selectedTileIsValid() {
					obj := &e.objectGrid.objs[activeObjsIndex]
					e.activeTileMap.InsertObject(obj, activeObjsIndex, selectedTile, currentLayer, &placedObjects[e.activeTileMapIndex])
				}
			case Link:
				if e.selectedTileIsValid() {
					if linkBegin == nil {
						linkBegin = &LinkData{
							selectionX,
							selectionY,
							e.activeTileMapIndex,
						}
					} else {
						linkEnd := &LinkData{
							selectionX,
							selectionY,
							e.activeTileMapIndex,
						}
						e.tryConnectTileMaps(linkBegin, linkEnd)
						linkBegin = nil
					}
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
	} else if ebiten.IsMouseButtonPressed(ebiten.MouseButton(0)) && ebiten.IsKeyPressed(ebiten.KeyShift) {
		cx, cy := ebiten.CursorPosition();
		if !e.isAlreadyClicking() {
			e.clickStartY = float64(cy)
			e.clickStartX = float64(cx)
		} else {
			offset := e.tileMapOffsets[e.activeTileMapIndex]
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
		offset := e.tileMapOffsets[e.activeTileMapIndex]
		offset.X = math.Round(offset.X / TileSize) * TileSize
		offset.Y = math.Round(offset.Y / TileSize) * TileSize
	}
}

func (e *Editor) isAlreadyClicking() bool {
	return e.clickStartX != -1 && e.clickStartY != -1
}

func (e *Editor) selectedTileIsValid() bool {
	//cx, cy := ebiten.CursorPosition()
	return 0 <= selectedTile && selectedTile < len(e.activeTileMap.Tiles[currentLayer]) //&& e.getTileMapIndexAtCoord(cx, cy) != -1
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

func (e *Editor) autoTileGridIsVisible() bool {
	return activeTool == AutoTile
}

func (e *Editor) npcGridIsVisible() bool {
	return activeTool == PlaceNpc
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
		x1 := IconOffsetX
		y1 := IconOffsetY + i * (h + IconPadding)

		x2 := x1 + w
		y2 := y1 + h

		r := image.Rect(x1, y1, x2, y2)
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

	e.objectGrid = NewObjectGrid(e.activeTileMap, objs)
}

func (e *Editor) setActiveTileMap(index int) {
	e.activeTileMap = e.tileMaps[index]
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

func (e *Editor) switchActiveTool(newTool int) {
	activeTool = newTool
	linkBegin = nil
}

func (e *Editor) npcAtPosition(x, y int) bool {
	if e.activeTileMap == nil {
		return true
	}

	for i := range e.activeTileMap.NpcInfo {
		otherX := e.activeTileMap.NpcInfo[i].X
		otherY := e.activeTileMap.NpcInfo[i].Y

		if x == otherX && y == otherY {
			return true
		}
	}

	return false
}

func (e *Editor) tryConnectTileMaps(start, end *LinkData) {
	if !e.validateLink(start) || !e.validateLink(end) {
		// abort
		return
	}

	if *start == *end {
		// also abort
		return
	}

	startEntryIndex := 0
	endEntryIndex := 0

	startEntries := e.tileMaps[start.TileMapIndex].Entries
	endEntries := e.tileMaps[end.TileMapIndex].Entries

	for i := range startEntries {
		if startEntryIndex != startEntries[i].Id {
			break
		}
		startEntryIndex++
	}

	for i := range endEntries {
		if endEntryIndex != endEntries[i].Id {
			break
		}
		endEntryIndex++
	}

	entryA := Entry{
		startEntryIndex,
		start.X,
		start.Y,
		currentLayer,
	}

	exitA := Exit{
		e.activeFiles[end.TileMapIndex],
		endEntryIndex,
		start.X,
		start.Y,
		currentLayer,
	}

	entryB := Entry{
		endEntryIndex,
		end.X,
		end.Y,
		currentLayer,
	}

	exitB := Exit{
		e.activeFiles[start.TileMapIndex],
		startEntryIndex,
		end.X,
		end.Y,
		currentLayer,
	}

	e.tileMaps[start.TileMapIndex].PlaceEntry(entryA)
	e.tileMaps[start.TileMapIndex].PlaceExit(exitA)
	e.tileMaps[end.TileMapIndex].PlaceEntry(entryB)
	e.tileMaps[end.TileMapIndex].PlaceExit(exitB)

}

func (e *Editor) validateLink(link *LinkData) bool {
	w := e.tileMaps[link.TileMapIndex].Width
	index := link.Y * w + link.X

	// cannot put link on tile with collision
	if e.tileMaps[link.TileMapIndex].Collision[currentLayer][index] {
		return false
	}

	exits := e.tileMaps[link.TileMapIndex].Exits
	for i := range exits {
		if exits[i].X == link.X && exits[i].Y == link.Y {
			return false
		}
	}
	entries := e.tileMaps[link.TileMapIndex].Entries
	for i := range entries {
		if entries[i].X == link.X && entries[i].Y == link.Y {
			return false
		}
	}
	return true
}

func (e *Editor) drawLinksFromActiveTileMap() {
	clr := color.RGBA{
		245,
		173,
		66,
		255,
	}

	for _, ex := range e.activeTileMap.Exits {
		for i := range e.activeFiles {
			if e.activeFiles[i] == ex.Target {
				line := DebugLine{}
				line.Clr = clr
				line.X1 = float64(ex.X) * TileSize + e.tileMapOffsets[e.activeTileMapIndex].X + TileSize / 2
				line.Y1 = float64(ex.Y) * TileSize + e.tileMapOffsets[e.activeTileMapIndex].Y + TileSize / 2

				for _, en := range e.tileMaps[i].Entries {
					if en.Id == ex.Id {
						line.X2 = float64(en.X) * TileSize + e.tileMapOffsets[i].X + TileSize / 2
						line.Y2 = float64(en.Y) * TileSize + e.tileMapOffsets[i].Y + TileSize / 2
						break
					}
				}

				e.rend.DrawLine(line)
			}
		}
	}
}

func (e *Editor) removeInvalidLinks() {
	ex := e.activeTileMap.Exits[:]

	for i := range ex {
		if ex[i].X <= e.activeTileMap.Width || ex[i].Y <= e.activeTileMap.Height {
			e.removeLink(e.activeTileMapIndex, i)
		}
	}

	en := e.activeTileMap.Entries[:]
	for i := range en {
		if en[i].X <= e.activeTileMap.Width || en[i].Y <= e.activeTileMap.Height {
			e.removeLinkFromEntry(e.activeTileMapIndex, i)
		}
	}
}

func (e *Editor) removeLink(tileMapIndex, exitIndex int) {
	exs := e.tileMaps[tileMapIndex].Exits[:]
	ex := exs[exitIndex]
	e.tileMaps[tileMapIndex].Exits = append(exs[:exitIndex], exs[exitIndex+1:]...)

	var otherTileMap *TileMap
	for i := range e.activeFiles {
		if e.activeFiles[i] == ex.Target {
			otherTileMap = e.tileMaps[i]
			break
		}
	}

	if otherTileMap == nil {
		return
	}

	for i := range otherTileMap.Entries {
		if otherTileMap.Entries[i].Id == ex.Id {
			otherTileMap.Entries = append(otherTileMap.Entries[:i], otherTileMap.Entries[i+1:]...)
			break
		}
	}
}

func (e *Editor) removeLinkFromEntry(tileMapIndex, entryIndex int) {
	tm := e.tileMaps[tileMapIndex]
	en := tm.Entries[entryIndex]
	for _, tm = range e.tileMaps {
		for i := range tm.Exits {
			target := e.activeFiles[tileMapIndex]
			if tm.Exits[i].Target == target && tm.Exits[i].Id == en.Id {
				e.removeLink(tileMapIndex, i)
				break
			}
		}
	}
}

func (e *Editor) tryPlaceNpc() {
	x := selectedTile % e.activeTileMap.Width
	y := selectedTile / e.activeTileMap.Width
	if !e.npcAtPosition(x, y) {
		i := e.npcGrid.GetIndex()
		e.dialog.Hidden = false
		e.tw.Start("Enter name of dialogue file:", func (str string) {
			e.dialog.Hidden = true
			if str == "" {
				return
			}

			ni := &NpcInfo{
				e.npcImagesStrings[i],
				str,
				x,
				y,
				currentLayer,
			}

			e.activeTileMap.PlaceNpc(ni)
		})
	}
}

func listPngs(dir string) []string {
	dirs, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Println("Could not open dir", dir)
		return make([]string, 0)
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

func loadImages(images []string) []*ebiten.Image {
	imgs := make([]*ebiten.Image, 0, len(images))

	for _, s := range images {
		img, _, err := ebitenutil.NewImageFromFile(s, ebiten.FilterDefault)
		if err != nil {
			log.Println("Could not load image", s)
		}
		imgs = append(imgs, img)
	}

	return imgs
}
