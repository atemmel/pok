package editor

import (
	"errors"
	"fmt"
	"github.com/atemmel/pok/pkg/pok"
	"github.com/atemmel/pok/pkg/constants"
	"github.com/atemmel/pok/pkg/debug"
	"github.com/atemmel/pok/pkg/fonts"
	"github.com/atemmel/pok/pkg/textures"
	"github.com/atemmel/pok/pkg/jobs"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/sqweek/dialog"
	"io/ioutil"
	"image"
	"image/color"
	"log"
	"math"
	"os"
    "path/filepath"
	"strings"
)

var WorkingDir, _ = os.Getwd()

var selectionX int
var selectionY int
var copyBuffer = 0
var selectedTile = 0
var currentLayer = 0
var baseTextureIndex = 0
var activeObjsIndex = 0
var activeAtiIndex = -1

var drawOnlyCurrentLayer = false
var DrawDebugInfo = false
var activeTool = Pencil
var placedObjects [][]PlacedEditorObject = make([][]PlacedEditorObject, 0)
var linkBegin *LinkData
var lastSavedUndoStackLength = 0

var treeArea = &TreeAreaSelection{}
var activePalette = 0
var availablePalettes []*ebiten.Image
var availablePaletteIndicies []int
var availablePalettesStrings []string

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
	tileMaps []*pok.TileMap
	tileMapOffsets  []*Vec2
	activeTileMapIndex int

	activeTileMap *pok.TileMap
	rend pok.Renderer
	grid Grid
	objectGrid ObjectGrid
	selection *ebiten.Image
	backgroundGrid *ebiten.Image
	collisionMarker *ebiten.Image
	deleteableMarker *ebiten.Image
	exitMarker *ebiten.Image
	icons *ebiten.Image
	activeFiles []string
	activeFullFiles []string
	nextFile string
	clickStartX float64
	clickStartY float64
	resizers []Resize
	autoTileInfo []AutoTileInfo
	autoTileGrid AutoTileGrid
	treeAutoTileInfo []TreeAutoTileInfo
	treeAutoTileGrid TreeAutoTileGrid
	npcImages []*ebiten.Image
	npcImagesStrings []string
	npcGrid NpcGrid
	dieOnNextTick bool
}

func NewEditor(paths []string) *Editor {
	var err error
	es := &Editor{}

	es.dieOnNextTick = false

	es.selection = ebiten.NewImage(constants.TileSize, constants.TileSize)
	es.collisionMarker = ebiten.NewImage(constants.TileSize, constants.TileSize)
	es.exitMarker = ebiten.NewImage(constants.TileSize, constants.TileSize)
	es.deleteableMarker = ebiten.NewImage(constants.TileSize, constants.TileSize)
	es.backgroundGrid = ebiten.NewImage(constants.TileSize, constants.TileSize)

	selectionClr := color.RGBA{255, 0, 0, 255}
	collisionClr := color.RGBA{255, 0, 255, 255}
	exitClr := color.RGBA{0, 0, 255, 255}
	deleteableClr := color.RGBA{150, 0, 0, 255}
	backgroundGridClr := color.RGBA{255, 255, 255, 25}

	for p := 0; p < es.selection.Bounds().Max.X; p++ {
		es.selection.Set(p, 0, selectionClr)
		es.selection.Set(p, es.selection.Bounds().Max.Y - 1, selectionClr)
	}

	for p := 1; p < es.selection.Bounds().Max.Y - 1; p++ {
		es.selection.Set(0, p, selectionClr)
		es.selection.Set(es.selection.Bounds().Max.Y - 1, p, selectionClr)
	}

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

	for p := 0; p < es.backgroundGrid.Bounds().Max.X; p++ {
		es.backgroundGrid.Set(p, 0, backgroundGridClr)
		es.backgroundGrid.Set(p, es.backgroundGrid.Bounds().Max.Y - 1, backgroundGridClr)
	}

	for p := 1; p < es.backgroundGrid.Bounds().Max.Y - 1; p++ {
		es.backgroundGrid.Set(0, p, backgroundGridClr)
		es.backgroundGrid.Set(es.backgroundGrid.Bounds().Max.Y - 1, p, backgroundGridClr)
	}

	es.rend = pok.NewRenderer(constants.DisplaySizeX, constants.DisplaySizeY, 1)

	es.clickStartX = -1
	es.clickStartY = -1

	es.icons, err = textures.LoadWithError(constants.EditorImagesDir + "editoricons.png")
	debug.Assert(err)

	es.tileMaps = make([]*pok.TileMap, 0)
	es.tileMapOffsets = make([]*Vec2, 0)

	es.npcImagesStrings = listPngs(constants.CharacterImagesDir)
	es.npcImages = loadImages(es.npcImagesStrings, constants.CharacterImagesDir)
	es.npcGrid = NewNpcGrid(es.npcImages)

	es.treeAutoTileInfo, err = ReadAllTreeAutoTileInfo(constants.TreeAutotileInfoDir)
	debug.Assert(err)
	if len(es.treeAutoTileInfo) > 0 {
		treeArea.TreeInfo = &es.treeAutoTileInfo[0]
	}

	availablePalettesStrings = listPngs(constants.TileMapImagesDir)
	availablePaletteIndicies = make([]int, len(availablePalettesStrings))
	availablePalettes = make([]*ebiten.Image, len(availablePalettesStrings))

	for i, s := range availablePalettesStrings {
		img, index := textures.Load(constants.TileMapImagesDir + s)
		availablePalettes[i] = img
		availablePaletteIndicies[i] = index
	}

	for _, s := range paths {
		tm, err := es.loadFile(s)
		if err == nil {
			es.updateEditorWithNewTileMap(tm)
		}
	}

	font, err := fonts.LoadFont(constants.FontsDir + "pokemon_pixel_font.ttf", 16)
	debug.Assert(err)

	initButtons(font)

	vis := func() bool {
		return (activeTool == Pencil || activeTool == Bucket) && es.activeTileMap != nil
	}

	AddButton(&ButtonInfo{
		Content: "PREV",
		OnClick: func() {
			activePalette--
			if activePalette < 0 {
				activePalette = len(availablePalettes) - 1
			}

			es.grid = NewGrid(availablePalettes[activePalette], constants.TileSize)
		},
		VisibilityCondition: vis,
		X: xGridPos, Y: yGridPos - 18,
	})

	AddButton(&ButtonInfo{
		Content: "NEXT",
		OnClick: func() {
			activePalette++
			if activePalette >= len(availablePalettes) {
				activePalette = 0
			}

			es.grid = NewGrid(availablePalettes[activePalette], constants.TileSize)
		},
		VisibilityCondition: vis,
		X: xGridPos + 98, Y: yGridPos - 18,
	})

	AddButton(&ButtonInfo{
		Content: "+",
		OnClick: func() {
			es.doIncrementLayer()
		},
		VisibilityCondition: nil,
		X: IconOffsetX, Y: constants.DisplaySizeY - 20 - 20,
	})

	AddButton(&ButtonInfo{
		Content: "-",
		OnClick: func() {
			es.doDecrementLayer()
		},
		VisibilityCondition: nil,
		X: IconOffsetX, Y: constants.DisplaySizeY - 20,
	})

	AddButton(&ButtonInfo{
		Content: "DELETE LAYER",
		OnClick: func() {
			es.doRemoveLayer()
		},
		VisibilityCondition: nil,
		X: IconOffsetX + 16, Y: constants.DisplaySizeY - 20,
	})

	AddButton(&ButtonInfo{
		Content: "ADD LAYER",
		OnClick: func() {
			es.doAppendLayer()
		},
		VisibilityCondition: nil,
		X: IconOffsetX + 16, Y: constants.DisplaySizeY - 20 - 20,
	})

	AddButton(&ButtonInfo{
		Content: "NEW MAP",
		OnClick: func() {
			es.newFile()
		},
		VisibilityCondition: nil,
		X: IconOffsetX + 16 * 13, Y: constants.DisplaySizeY - 20,
	})

	AddButton(&ButtonInfo{
		Content: "OPEN MAP",
		OnClick: func() {
			es.loadFileDialog()
		},
		VisibilityCondition: nil,
		X: IconOffsetX + 16 * 16, Y: constants.DisplaySizeY - 20,
	})

	jobs.Add(jobs.Job{
		Do: pok.WaterAnim,
		When: 11,
	})

	jobs.Add(jobs.Job{
		Do: pok.WaterSplashAnim,
		When: 11,
	})

	return es;
}

func (e *Editor) Update() error {
	err := e.handleInputs()
	jobs.TickAllOneFrame()
	return err
}

func (e *Editor) Draw(screen *ebiten.Image) {
	e.DrawBackgroundGrid()
	if e.activeTileMap != nil {
		vec := e.tileMapOffsets[e.activeTileMapIndex]
		treeArea.Draw(&e.rend, *vec)
	}

	for i := range e.tileMaps {
		offset := e.tileMapOffsets[i]
		e.tileMaps[i].DrawWithOffset(&e.rend, offset.X, offset.Y, drawOnlyCurrentLayer, currentLayer)
	}
	if DrawDebugInfo && len(e.activeFiles) != 0 {
		e.drawLinksFromActiveTileMap()
		e.DrawTileMapDetail()
		e.resizers[e.activeTileMapIndex].Draw(&e.rend)
	}
	e.rend.Display(screen)

	if DrawDebugInfo && len(e.activeFiles) != 0 {
		if e.gridIsVisible() {
			e.grid.Draw(screen)
		} else if e.objectGridIsVisible() {
			e.objectGrid.Draw(screen)
		} else if e.autoTileGridIsVisible() {
			e.autoTileGrid.Draw(screen)
		} else if e.npcGridIsVisible() {
			e.npcGrid.Draw(screen)
		} else if e.treeAutoTileGridIsVisible() {
			e.treeAutoTileGrid.grid.Draw(screen)
		}
		e.drawIcons(screen)
	}

	drawButtons(screen)

	debugStr := ""
	if len(e.activeFiles) == 0 {
		debugStr += "(No files)"
	} else {
		if !e.hasSaved() {
			debugStr += "*"
		}
		debugStr += e.activeFiles[e.activeTileMapIndex]
	}
	debugStr += fmt.Sprintf(`
x: %f, y: %f, z: %d
zoom: %d%%
%s`, e.rend.Cam.X, e.rend.Cam.Y, currentLayer, int(e.rend.Cam.Scale * 100), ToolNames[activeTool])
	ebitenutil.DebugPrint(screen, debugStr)
}

func (e *Editor) DrawBackgroundGrid() {
	xMax := e.rend.Cam.W / e.rend.Cam.Scale
	yMax := e.rend.Cam.H / e.rend.Cam.Scale

	x := e.rend.Cam.X - float64(int(e.rend.Cam.X) % constants.TileSize) - constants.TileSize
	xLeft := x + constants.TileSize

	for x < xLeft + xMax {
		y := e.rend.Cam.Y - float64(int(e.rend.Cam.Y) % constants.TileSize) - constants.TileSize
		yLeft := y + constants.TileSize
		for y < yLeft + yMax {
			e.rend.Draw(&pok.RenderTarget{
				Op: &ebiten.DrawImageOptions{},
				Src: e.backgroundGrid,
				SubImage: nil,
				X: x,
				Y: y,
				Z: -1337,
			})
			y += float64(e.backgroundGrid.Bounds().Max.Y)
		}
		x += float64(e.backgroundGrid.Bounds().Max.X)
	}
}

func (e *Editor) DrawTileMapDetail() {
	offset := e.tileMapOffsets[e.activeTileMapIndex]
	for j := range e.activeTileMap.Collision {
		if drawOnlyCurrentLayer && j != currentLayer {
			continue
		}
		for i := range e.activeTileMap.Collision[j] {
			x := float64(i % e.activeTileMap.Width) * constants.TileSize
			y := float64(i / e.activeTileMap.Width) * constants.TileSize

			if currentLayer == j && e.activeTileMap.Collision[j][i] {
				e.rend.Draw(&pok.RenderTarget{
					Op: &ebiten.DrawImageOptions{},
					Src: e.collisionMarker,
					SubImage: nil,
					X: x + offset.X,
					Y: y + offset.Y,
					Z: 100,
				})
			}
		}
	}

	if DrawDebugInfo {
		for i := range e.activeTileMap.Exits {
			e.rend.Draw(&pok.RenderTarget{
				Op: &ebiten.DrawImageOptions{},
				Src: e.exitMarker,
				SubImage: nil,
				X: float64(e.activeTileMap.Exits[i].X * constants.TileSize) + offset.X,
				Y: float64(e.activeTileMap.Exits[i].Y * constants.TileSize) + offset.Y,
				Z: 100,
			})
		}

		if activeTool == Eraser {
			for i := range placedObjects[e.activeTileMapIndex] {
				e.rend.Draw(&pok.RenderTarget{
					Op: &ebiten.DrawImageOptions{},
					Src: e.deleteableMarker,
					SubImage: nil,
					X: float64(placedObjects[e.activeTileMapIndex][i].X * constants.TileSize) + offset.X,
					Y: float64(placedObjects[e.activeTileMapIndex][i].Y * constants.TileSize) + offset.Y,
					Z: 100,
				})
			}
		}

		e.rend.Draw(&pok.RenderTarget{
			Op: &ebiten.DrawImageOptions{},
			Src: e.selection,
			SubImage: nil,
			X: float64(selectionX * constants.TileSize) + offset.X,
			Y: float64(selectionY * constants.TileSize) + offset.Y,
			Z: 100,
		})
	}
}

func (e *Editor) SelectTileFromMouse(cx, cy int) {
	offset := e.tileMapOffsets[e.activeTileMapIndex]
	cx = int(float64(cx) / e.rend.Cam.Scale)
	cy = int(float64(cy) / e.rend.Cam.Scale)

	cx += int(math.Round(e.rend.Cam.X - offset.X))
	cy += int(math.Round(e.rend.Cam.Y - offset.Y))

	cx -= cx % constants.TileSize
	cy -= cy % constants.TileSize
	selectionX = cx / constants.TileSize
	selectionY = cy / constants.TileSize
	selectedTile =  selectionX + selectionY * e.activeTileMap.Width
}

func (e *Editor) loadFileDialog() {

	file, err := dialog.File().Title("Open map").Filter("All Files", "*").Load()
	os.Chdir(WorkingDir)
	if err != nil && file == ""{
		return
	} else if err != nil {
		dialog.Message("Could not open file: %s", file).Title("Error").Error()
		return
	}

	tm, err := e.loadFile(file)
	if err != nil {
		doNewFile := dialog.Message("Could not open file %s. Create new file?", file).Title("Create new file?").YesNo()
		if doNewFile {
			tm = pok.CreateTileMap(2, 2, listPngs(constants.TileMapImagesDir))
			e.updateEditorWithNewTileMap(tm)
		}
	} else {
		e.updateEditorWithNewTileMap(tm)
	}
}

func (e *Editor) loadFile(file string) (*pok.TileMap, error) {
	e.nextFile = file
	tm := &pok.TileMap{}
	err := tm.OpenFile(file)
	return tm, err
}

func (e *Editor) newFile() {
	file, err := dialog.File().Title("Name new map").Filter("All Files", "*").Save()
	if err != nil && file == ""{
		return
	} else if err != nil {
		dialog.Message("Could not open file: %s", file).Title("Error").Error()
		return
	}

	tm := pok.CreateTileMap(2, 2, []string{"base.png"})
	e.nextFile = file
	e.updateEditorWithNewTileMap(tm)
}

func (e *Editor) updateEditorWithNewTileMap(tileMap *pok.TileMap) {
	e.appendTileMap(tileMap)
	e.activeFullFiles = append(e.activeFiles, e.nextFile)
	e.activeFiles = append(e.activeFiles, filepath.Base(e.nextFile))
	DrawDebugInfo = true
	const baseIndex = 0
	e.grid = NewGrid(textures.Access(tileMap.TextureMapping[activePalette]), constants.TileSize)
	e.fillObjectGrid(constants.OverworldObjectsDir)
	var err error
	e.autoTileInfo, err = ReadAllAutoTileInfo(constants.AutotileInfoDir)
	debug.Assert(err)

	e.autoTileGrid = NewAutoTileGrid(e.autoTileInfo)

	for i := range e.treeAutoTileInfo {
		err := e.treeAutoTileInfo[i].FitToTileMap(tileMap)
		debug.Assert(err)
	}

	e.treeAutoTileGrid = NewTreeAutoTileGrid(textures.Access(tileMap.TextureMapping[baseIndex]), e.treeAutoTileInfo)
}

func (e *Editor) appendTileMap(tileMap *pok.TileMap) {
	placedObjects = append(placedObjects, make([]PlacedEditorObject, 0))
	e.tileMaps = append(e.tileMaps, tileMap)
	e.tileMapOffsets = append(e.tileMapOffsets, &Vec2{0, 0})
	e.activeTileMap = e.tileMaps[len(e.tileMaps)-1]
	e.resizers = append(e.resizers, NewResize(e.tileMaps[len(e.tileMaps)-1], e.tileMapOffsets[len(e.tileMapOffsets) - 1]))
}

func (e *Editor) saveFile() {
	err := e.activeTileMap.SaveToFile(e.activeFullFiles[e.activeTileMapIndex])
	if err != nil {
		dialog.Message("Could not save file %s, %s", e.activeFullFiles[e.activeTileMapIndex], err.Error()).Title("Error").Error()
		return
	}

	lastSavedUndoStackLength = len(UndoStack)
}

func (e *Editor) hasSaved() bool {
	return len(UndoStack) == lastSavedUndoStackLength
}
func (e *Editor) unsavedWorkDialog() {
	shouldDie := dialog.Message("You have unsaved work. Are you sure you want to exit?").Title("Unsaved work").YesNo()
	if shouldDie {
		os.Exit(0)
	}
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

	if ebiten.IsKeyPressed(ebiten.KeyControl) {
		if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
			PerformUndo(e)
		} else if inpututil.IsKeyJustPressed(ebiten.KeyY) {
			PerformRedo(e)
		}
	}

	cx, cy := ebiten.CursorPosition()
	if pollButtons(cx, cy) {
		return nil
	}

	if len(e.activeFiles) != 0 {
		index := e.getTileMapIndexAtCoord(cx, cy)
		if index != -1 && !e.isAlreadyClicking() {
			err := e.setActiveTileMap(index)
			debug.Assert(err)
		}
		if e.gridIsVisible() && e.grid.Contains(image.Point{cx, cy}) {
			_, sy := ebiten.Wheel()
			if sy < 0 {
				e.grid.Scroll(ScrollDown)
			} else if sy > 0 {
				e.grid.Scroll(ScrollUp)
			}
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton(0)) {
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
		} else if e.treeAutoTileGridIsVisible() && e.treeAutoTileGrid.Contains(image.Point{cx, cy}) {
			_, sy := ebiten.Wheel()
			if sy < 0 {
				e.treeAutoTileGrid.Scroll(ScrollDown)
			} else if sy > 0 {
				e.treeAutoTileGrid.Scroll(ScrollUp)
			}
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton(0)) {
				e.treeAutoTileGrid.Select(cx, cy)
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
		DrawDebugInfo = !DrawDebugInfo
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyO) {
		e.loadFileDialog()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyN) {
		e.newFile()
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
			if e.rend.Cam.Scale > 0.50000001 {
				e.rend.ZoomToCenter(e.rend.Cam.Scale - 0.1)
			}
		} else {
			if e.rend.Cam.Scale < 2.0 {
				e.rend.ZoomToCenter(e.rend.Cam.Scale + 0.1)
			}
		}
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton(0)) {
		cx, cy := ebiten.CursorPosition();
		e.resizers[e.activeTileMapIndex].tryClick(cx, cy, &e.rend.Cam)
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButton(0)) && !ebiten.IsKeyPressed(ebiten.KeyControl) && !ebiten.IsKeyPressed(ebiten.KeyShift) {
		cx, cy := ebiten.CursorPosition();
		if !e.isAlreadyClicking() && e.resizers[e.activeTileMapIndex].IsHolding() {
			e.doResize()
		} else {
			e.SelectTileFromMouse(cx, cy)
			if e.selectedTileIsValid() {
				switch activeTool {
					case Pencil:
						e.doPencil()
					case Eraser:
						e.doEraser()
					case AutoTile:
						e.doAutotile()
					case Tree:
						//TODO: perform tree logic
						treeArea.TreeInfo = &e.treeAutoTileInfo[e.treeAutoTileGrid.GetIndex()]
						treeArea.Hold(selectionX, selectionY)
				}
			}
		}
	} else {
		x, y, origin := e.resizers[e.activeTileMapIndex].Release()
		if origin != -1 {
			e.postDoResize(x, y, origin)
		}
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton(0)) && !ebiten.IsKeyPressed(ebiten.KeyControl) && !ebiten.IsKeyPressed(ebiten.KeyShift) {
		cx, cy := ebiten.CursorPosition();
		e.SelectTileFromMouse(cx, cy)
		if e.selectedTileIsValid() {
			switch activeTool {
				case Bucket:
					e.doBucket()
				case Object:
					e.doObject()
				case Link:
					e.doLink()
				case PlaceNpc:
					e.doPlaceNpc()
			}
		}
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton(1)) {
		cx, cy := ebiten.CursorPosition();
		e.SelectTileFromMouse(cx, cy)
		if e.selectedTileIsValid() {
			switch activeTool {
				case Object:
					e.doRemoveObject()
				case Link:
					e.doRemoveLink()
				case PlaceNpc:
					e.doRemoveNpc()
			}
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
			offset.X += float64(cx) - e.clickStartX
			offset.Y += float64(cy) - e.clickStartY
			e.clickStartX = float64(cx)
			e.clickStartY = float64(cy)
		}
	} else {
		e.clickStartX = -1
		e.clickStartY = -1
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButton(0)) && !ebiten.IsKeyPressed(ebiten.KeyControl) && !ebiten.IsKeyPressed(ebiten.KeyShift) {
		switch activeTool {
			case Pencil:
				e.postDoPencil()
			case Eraser:
				e.postDoEraser()
			case Bucket:
				e.postDoBucket()
			case Object:
				e.postDoObject()
			case AutoTile:
				e.postDoAutotile()
			case PlaceNpc:
				e.postDoNpc()
			case Tree:
				treeArea.Release(e.activeTileMap, currentLayer)
		}

		RedoStack = RedoStack[:0]

		offset := e.tileMapOffsets[e.activeTileMapIndex]
		offset.X = math.Round(offset.X / constants.TileSize) * constants.TileSize
		offset.Y = math.Round(offset.Y / constants.TileSize) * constants.TileSize
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButton(1)) {
		switch activeTool {
			case Object:
				e.postDoRemoveObject()
			case Link:
				e.postDoRemoveLink()
			case PlaceNpc:
				e.postDoRemoveNpc()
		}
	}
}

func (e *Editor) isAlreadyClicking() bool {
	return e.clickStartX != -1 && e.clickStartY != -1
}

func (e *Editor) selectedTileIsValid() bool {
	return 0 <= selectedTile && selectedTile < len(e.activeTileMap.Tiles[currentLayer])
}

func (e *Editor) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return constants.DisplaySizeX, constants.DisplaySizeY
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

func (e *Editor) treeAutoTileGridIsVisible() bool {
	return activeTool == Tree
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
	debug.Assert(err)

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

	e.objectGrid = NewObjectGrid(objs, availablePalettesStrings)
}

func (e *Editor) setActiveTileMap(index int) error {
	e.activeTileMap = e.tileMaps[index]
	e.activeTileMapIndex = index
	for i := range e.treeAutoTileInfo {
		err := e.treeAutoTileInfo[i].FitToTileMap(e.activeTileMap)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Editor) getTileMapIndexAtCoord(cx, cy int) int {
	p := image.Point{cx, cy}
	for i := range e.tileMaps {
		w := int(float64(e.tileMaps[i].Width * constants.TileSize) / e.rend.Cam.Scale)
		h := int(float64(e.tileMaps[i].Height * constants.TileSize) / e.rend.Cam.Scale)
		x := int(math.Round((e.tileMapOffsets[i].X - e.rend.Cam.X) / e.rend.Cam.Scale))
		y := int(math.Round((e.tileMapOffsets[i].Y - e.rend.Cam.Y) / e.rend.Cam.Scale))

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

	startEntries := e.tileMaps[start.TileMapIndex].Entries[:]
	endEntries := e.tileMaps[end.TileMapIndex].Entries[:]

	for ; startEntryIndex < len(startEntries); startEntryIndex++ {
		valid := true
		for i := range startEntries {
			if startEntryIndex == startEntries[i].Id {
				valid = false
				break
			}
		}

		if valid {
			break
		}
	}

	for ; endEntryIndex < len(endEntries); endEntryIndex++ {
		valid := true
		for i := range endEntries {
			if endEntryIndex == endEntries[i].Id {
				valid = false
				break
			}
		}

		if valid {
			break
		}
	}

	entryA := pok.Entry{
		Id: startEntryIndex,
		X: start.X,
		Y: start.Y,
		Z: currentLayer,
	}

	exitA := pok.Exit{
		Target: e.activeFiles[end.TileMapIndex],
		Id: endEntryIndex,
		X: start.X,
		Y: start.Y,
		Z: currentLayer,
	}

	entryB := pok.Entry{
		Id: endEntryIndex,
		X: end.X,
		Y: end.Y,
		Z: currentLayer,
	}

	exitB := pok.Exit{
		Target: e.activeFiles[start.TileMapIndex],
		Id: startEntryIndex,
		X: end.X,
		Y: end.Y,
		Z: currentLayer,
	}

	e.tileMaps[start.TileMapIndex].PlaceEntry(entryA)
	e.tileMaps[start.TileMapIndex].PlaceExit(exitA)
	e.tileMaps[end.TileMapIndex].PlaceEntry(entryB)
	e.tileMaps[end.TileMapIndex].PlaceExit(exitB)

	CurrentLinkDelta.linkIdA = startEntryIndex
	CurrentLinkDelta.linkIdB = endEntryIndex
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
				line := pok.DebugLine{}
				line.Clr = clr
				line.X1 = float64(ex.X) * constants.TileSize + e.tileMapOffsets[e.activeTileMapIndex].X + constants.TileSize / 2
				line.Y1 = float64(ex.Y) * constants.TileSize + e.tileMapOffsets[e.activeTileMapIndex].Y + constants.TileSize / 2

				for _, en := range e.tileMaps[i].Entries {
					if en.Id == ex.Id {
						line.X2 = float64(en.X) * constants.TileSize + e.tileMapOffsets[i].X + constants.TileSize / 2
						line.Y2 = float64(en.Y) * constants.TileSize + e.tileMapOffsets[i].Y + constants.TileSize / 2
						break
					}
				}

				e.rend.DrawLine(line)
			}
		}
	}
}

func (e *Editor) removeInvalidLinks(tileMapIndex int) (map[int]pok.Exit, map[int]pok.Entry, []int, []int) {
	tm := e.tileMaps[tileMapIndex]
	exs := tm.Exits[:]

	oldExits := make(map[int]pok.Exit)
	oldEntries := make(map[int]pok.Entry)

	exitsToRemove := make([]int, 0)
	entriesToRemove := make([]int, 0)

	exitIndicies := make([]int, 0)
	entryIndicies := make([]int, 0)

	for i := 0; i < len(exs); i++ {
		if exs[i].X >= tm.Width || exs[i].Y >= tm.Height {
			exitsToRemove = append(exitsToRemove, i)
			exitIndicies = append(exitIndicies, i)
			oldExits[i] = exs[i]
		}
	}

	ens := tm.Entries[:]
	for i := 0; i < len(ens); i++ {
		if ens[i].X >= tm.Width || ens[i].Y >= tm.Height {
			entriesToRemove = append(entriesToRemove, i)
			entryIndicies = append(entryIndicies, i)
			oldEntries[i] = ens[i]
		}
	}

	for i := len(exitsToRemove) - 1; i >= 0; i-- {
		_, ent := e.removeLink(tileMapIndex, exitsToRemove[i] )
		if ent != nil {
			for j := range entriesToRemove {
				if entriesToRemove[j] == *ent {
					entriesToRemove[j] = -1
				}
				if entriesToRemove[j] > *ent {
					entriesToRemove[j]--
				}
			}
		}
	}

	for i := len(entriesToRemove) - 1; i >= 0; i-- {
		if entriesToRemove[i] == -1 {
			continue
		}

		_, ent := e.removeLinkFromEntry(tileMapIndex, entriesToRemove[i])
		if ent != nil {
			for j := range entriesToRemove {
				if entriesToRemove[j] > *ent {
					entriesToRemove[j]--
				}
			}
		}
	}

	return oldExits, oldEntries, exitIndicies, entryIndicies
}

func (e *Editor) removeLink(tileMapIndex, exitIndex int) (*int, *int){
	exs := e.tileMaps[tileMapIndex].Exits[:]
	ex := exs[exitIndex]
	e.tileMaps[tileMapIndex].Exits = append(exs[:exitIndex], exs[exitIndex+1:]...)

	var otherTileMap *pok.TileMap
	for i := range e.activeFiles {
		if e.activeFiles[i] == ex.Target {
			otherTileMap = e.tileMaps[i]
			break
		}
	}

	if otherTileMap == nil {
		return &exitIndex, nil
	}

	var entry *int

	for i := range otherTileMap.Entries {
		if otherTileMap.Entries[i].Id == ex.Id {
			cpy := i
			entry = &cpy
			otherTileMap.Entries = append(otherTileMap.Entries[:i], otherTileMap.Entries[i+1:]...)
			break
		}
	}

	return &exitIndex, entry
}

func (e *Editor) removeLinkFromEntry(tileMapIndex, entryIndex int) (*int, *int) {
	en := e.tileMaps[tileMapIndex].Entries[entryIndex]
	for otherTileMapIndex, tm := range e.tileMaps {
		for i := range tm.Exits {
			target := e.activeFiles[tileMapIndex]
			if tm.Exits[i].Target == target && tm.Exits[i].Id == en.Id {
				return e.removeLink(otherTileMapIndex, i)
			}
		}
	}

	return nil, nil
}

func (e *Editor) doPencil() {
	oldTile := e.activeTileMap.Tiles[currentLayer][selectedTile]
	oldTextureIndex := e.activeTileMap.TextureIndicies[currentLayer][selectedTile]
	i := e.grid.GetIndex()

	// no-op
	if oldTile == i && oldTextureIndex == activePalette {
		return
	}

	CurrentPencilDelta.indicies = append(CurrentPencilDelta.indicies, selectedTile)
	CurrentPencilDelta.oldTiles = append(CurrentPencilDelta.oldTiles, oldTile)
	CurrentPencilDelta.oldTextureIndicies = append(CurrentPencilDelta.oldTextureIndicies, oldTextureIndex)

	if !e.activeTileMap.HasTexture(availablePaletteIndicies[activePalette]) {
		e.activeTileMap.AppendTexture(availablePaletteIndicies[activePalette], availablePalettesStrings[activePalette])
	}

	e.activeTileMap.Tiles[currentLayer][selectedTile] = i
	e.activeTileMap.TextureIndicies[currentLayer][selectedTile] = e.activeTileMap.MapReverse(availablePaletteIndicies[activePalette])
}

func (e *Editor) doEraser() {
	oldTile := e.activeTileMap.Tiles[currentLayer][selectedTile]
	oldTextureIndex := e.activeTileMap.TextureIndicies[currentLayer][selectedTile]

	// no-op
	if oldTile < 0 && oldTextureIndex == baseTextureIndex {
		return
	}

	CurrentEraserDelta.indicies = append(CurrentEraserDelta.indicies, selectedTile)
	CurrentEraserDelta.oldTiles = append(CurrentEraserDelta.oldTiles, oldTile)
	CurrentEraserDelta.oldTextureIndicies = append(CurrentEraserDelta.oldTextureIndicies, oldTextureIndex)

	e.activeTileMap.Tiles[currentLayer][selectedTile] = -1
	e.activeTileMap.TextureIndicies[currentLayer][selectedTile] = baseTextureIndex
}

func (e *Editor) doBucket() {
	oldTile := e.activeTileMap.Tiles[currentLayer][selectedTile]
	oldTextureIndex := e.activeTileMap.TextureIndicies[currentLayer][selectedTile]

	i := e.grid.GetIndex()
	j := e.activeTileMap.MapReverse(availablePaletteIndicies[activePalette])

	// no-op
	if oldTile == i && oldTextureIndex == j {
		return
	}

	x, y := e.activeTileMap.Coords(selectedTile)

	const prefill = 16

	CurrentBucketDelta.indicies = make([]int, 0, prefill)
	CurrentBucketDelta.oldTiles = make([]int, 0, prefill)
	CurrentBucketDelta.oldTextureIndicies = make([]int, 0, prefill)

	stack := make([]image.Point, 0, prefill)
	stack = append(stack, image.Pt(x, y))

	fill := func(index int) {
		e.activeTileMap.Tiles[currentLayer][index] = i
		e.activeTileMap.TextureIndicies[currentLayer][index] = j
	}

	for len(stack) > 0 {
		pt := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if !e.activeTileMap.Contains(pt.X, pt.Y) {
			continue
		}

		x, y = pt.X, pt.Y
		index := e.activeTileMap.Index(x, y)
		theTile := e.activeTileMap.Tiles[currentLayer][index]
		theTextureIndex := e.activeTileMap.TextureIndicies[currentLayer][index]

		if theTile == oldTile && theTextureIndex == oldTextureIndex {
			CurrentBucketDelta.indicies = append(CurrentBucketDelta.indicies, index)
			CurrentBucketDelta.oldTiles = append(CurrentBucketDelta.oldTiles, theTile)
			CurrentBucketDelta.oldTextureIndicies = append(CurrentBucketDelta.oldTextureIndicies, theTextureIndex)

			fill(index)
			stack = append(
				stack,
				image.Point{x, y + 1},
				image.Point{x, y - 1},
				image.Point{x + 1, y},
				image.Point{x - 1, y},
			)
		}
	}

	CurrentBucketDelta.newTile = i
	CurrentBucketDelta.newTextureIndex = j
	CurrentBucketDelta.z = currentLayer
}

func (e *Editor) doObject() {
	obj := &e.objectGrid.objs[activeObjsIndex]
	//e.activeTileMap.InsertObject(obj, activeObjsIndex, selectedTile, currentLayer, &placedObjects[e.activeTileMapIndex])
	obj.InsertObject(e.activeTileMap, activeObjsIndex, selectedTile, currentLayer, &placedObjects[e.activeTileMapIndex])

	CurrentObjectDelta.placedObjectIndex = len(placedObjects[e.activeTileMapIndex]) - 1
	CurrentObjectDelta.objectIndex = activeObjsIndex
	CurrentObjectDelta.tileMapIndex = e.activeTileMapIndex
	CurrentObjectDelta.origin = selectedTile
	CurrentObjectDelta.z = currentLayer
}

func (e *Editor) doRemoveObject() {
	col := selectedTile % e.activeTileMap.Width
	row := selectedTile / e.activeTileMap.Width
	i := HasPlacedObjectAt(placedObjects[e.activeTileMapIndex], col, row)
	if i == -1 {
		return
	}

	od := &ObjectDelta{
		i,
		placedObjects[e.activeTileMapIndex][i].Index,
		e.activeTileMapIndex,
		selectedTile,
		currentLayer,
	}

	//e.activeTileMap.EraseObject(placedObjects[e.activeTileMapIndex][i], &e.objectGrid.objs[placedObjects[e.activeTileMapIndex][i].Index])
	e.objectGrid.objs[placedObjects[e.activeTileMapIndex][i].Index].EraseObject(e.activeTileMap, placedObjects[e.activeTileMapIndex][i])
	placedObjects[e.activeTileMapIndex][i] = placedObjects[e.activeTileMapIndex][len(placedObjects[e.activeTileMapIndex]) - 1]
	placedObjects[e.activeTileMapIndex] = placedObjects[e.activeTileMapIndex][:len(placedObjects[e.activeTileMapIndex]) - 1]

	CurrentRemoveObjectDelta.objectDelta = od
}

func (e *Editor) doLink() {
	if selectionX < 0 || selectionY < 0 {
		return
	}

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

		CurrentLinkDelta.linkBegin = linkBegin
		CurrentLinkDelta.linkEnd = linkEnd
		linkBegin = nil
		linkEnd = nil

		e.postDoLink()
	}
}

func (e *Editor) doRemoveLink() {
	if selectionX < 0 || selectionY < 0 {
		return
	}

	exitIndex := -1
	entryIndex := -1

	for i := range e.activeTileMap.Exits {
		if e.activeTileMap.Exits[i].X == selectionX && e.activeTileMap.Exits[i].Y == selectionY {
			exitIndex = i
			break
		}
	}

	for i := range e.activeTileMap.Entries {
		if e.activeTileMap.Exits[i].X == selectionX && e.activeTileMap.Exits[i].Y == selectionY {
			entryIndex = i
			break
		}
	}

	if exitIndex == -1 && entryIndex == -1 {
		return
	}

	if exitIndex != -1 {
		exit := e.activeTileMap.Exits[exitIndex]
		CurrentRemoveLinkDelta.exit = &exit
		e.activeTileMap.Exits[exitIndex] = e.activeTileMap.Exits[len(e.activeTileMap.Exits)-1]
		e.activeTileMap.Exits = e.activeTileMap.Exits[:len(e.activeTileMap.Exits)-1]
	}

	if entryIndex != -1 {
		entry := e.activeTileMap.Entries[entryIndex]
		CurrentRemoveLinkDelta.entry = &entry
		e.activeTileMap.Entries[entryIndex] = e.activeTileMap.Entries[len(e.activeTileMap.Entries)-1]
		e.activeTileMap.Entries = e.activeTileMap.Entries[:len(e.activeTileMap.Entries)-1]
	}

	CurrentRemoveLinkDelta.tileMapIndex = e.activeTileMapIndex
}

func (e *Editor) doAutotile() {
	ati := &e.autoTileInfo[e.autoTileGrid.GetIndex()]
	atd := DecideTileIndicies(e.activeTileMap, selectedTile, currentLayer, ati)
	CurrentAutotileDelta.Join(atd)
	CurrentAutotileDelta.tileMapIndex = e.activeTileMapIndex
	CurrentAutotileDelta.z = currentLayer
}

func (e *Editor) doResize() {
	e.resizers[e.activeTileMapIndex].Hold()
	CurrentResizeDelta.tileMapIndex = e.activeTileMapIndex
}

func (e *Editor) doIncrementLayer() {
	currentLayer++
	if e.activeTileMap != nil && currentLayer == len(e.activeTileMap.Tiles) {
		currentLayer--
	}
}

func (e *Editor) doAppendLayer() {
	if e.activeTileMap != nil {
		e.activeTileMap.AppendLayer()
	}
}

func (e *Editor) doDecrementLayer() {
	currentLayer--
	if currentLayer < 0 {
		currentLayer = 0
	}
}

func (e *Editor) doRemoveLayer() {
	if e.activeTileMap != nil {
		e.activeTileMap.RemoveLayer(currentLayer)
		if currentLayer > 0 {
			currentLayer--
		}
	}
}

func (e *Editor) postDoPencil() {
	CurrentPencilDelta.z = currentLayer
	CurrentPencilDelta.tileMapIndex = e.activeTileMapIndex
	CurrentPencilDelta.newTile = e.grid.GetIndex()
	CurrentPencilDelta.newTextureIndex = baseTextureIndex
	UndoStack = append(UndoStack, CurrentPencilDelta)
	CurrentPencilDelta = &PencilDelta{}
}

func (e *Editor) postDoEraser() {
	CurrentEraserDelta.z = currentLayer
	CurrentEraserDelta.tileMapIndex = e.activeTileMapIndex
	CurrentEraserDelta.newTextureIndex = baseTextureIndex
	UndoStack = append(UndoStack, CurrentEraserDelta)
	CurrentEraserDelta = &EraserDelta{}
}

func (e *Editor) postDoBucket() {
	UndoStack = append(UndoStack, CurrentBucketDelta)
	CurrentBucketDelta = &BucketDelta{}
}

func (e *Editor) postDoObject() {
	UndoStack = append(UndoStack, CurrentObjectDelta)
	CurrentObjectDelta = &ObjectDelta{}
}

func (e *Editor) postDoRemoveObject() {
	if CurrentRemoveObjectDelta.objectDelta == nil {
		return
	}

	UndoStack = append(UndoStack, CurrentRemoveObjectDelta)
	CurrentRemoveObjectDelta = &RemoveObjectDelta{}
}

func (e *Editor) postDoLink() {
	UndoStack = append(UndoStack, CurrentLinkDelta)
	CurrentLinkDelta = &LinkDelta{}
}

func (e *Editor) postDoRemoveLink() {
	if CurrentRemoveLinkDelta.entry == nil && CurrentRemoveLinkDelta.exit == nil {
		return
	}

	UndoStack = append(UndoStack, CurrentRemoveLinkDelta)
	CurrentRemoveLinkDelta = &RemoveLinkDelta{}
}

func (e *Editor) postDoAutotile() {
	UndoStack = append(UndoStack, CurrentAutotileDelta)
	CurrentAutotileDelta = &AutotileDelta{}
}

func (e *Editor) postDoResize(x, y, origin int) {
	e.activeTileMap.Resize(x, y, origin)
	exits, entries, exitIndicies, entryIndicies := e.removeInvalidLinks(CurrentResizeDelta.tileMapIndex)

	offsetX := 0.0
	offsetY := 0.0

	if origin == constants.TopLeftCorner || origin == constants.TopRightCorner {
		offsetY = -float64(y * constants.TileSize)
		e.tileMapOffsets[e.activeTileMapIndex].Y += offsetY
	}

	if origin == constants.TopLeftCorner || origin == constants.BotLeftCorner {
		offsetX = -float64(x * constants.TileSize)
		e.tileMapOffsets[e.activeTileMapIndex].X += offsetX
	}
	linkBegin = nil

	CurrentResizeDelta.dx = x
	CurrentResizeDelta.dy = y
	CurrentResizeDelta.origin = origin
	CurrentResizeDelta.offsetDeltaX = offsetX
	CurrentResizeDelta.offsetDeltaY = offsetY
	CurrentResizeDelta.oldExits = exits
	CurrentResizeDelta.oldEntries = entries
	CurrentResizeDelta.exitIndicies = exitIndicies
	CurrentResizeDelta.entryIndicies = entryIndicies

	UndoStack = append(UndoStack, CurrentResizeDelta)
	CurrentResizeDelta = &ResizeDelta{}
}

func (e *Editor) postDoNpc() {
	if CurrentNpcDelta.npcInfo == nil {
		return
	}

	UndoStack = append(UndoStack, CurrentNpcDelta)
	CurrentNpcDelta = &NpcDelta{}
}

func (e *Editor) postDoRemoveNpc() {
	if CurrentRemoveNpcDelta.npcDelta == nil {
		return
	}

	UndoStack = append(UndoStack, CurrentRemoveNpcDelta)
	CurrentRemoveNpcDelta = &RemoveNpcDelta{}
}

func (e *Editor) doPlaceNpc() {
	x := selectedTile % e.activeTileMap.Width
	y := selectedTile / e.activeTileMap.Width
	if !e.npcAtPosition(x, y) {
		i := e.npcGrid.GetIndex()
		file, err := dialog.File().Title("Select NPC dialog file").Filter("All Files", "*").Load()
		os.Chdir(WorkingDir)
		if err != nil && file == ""{
			return
		} else if err != nil {
			dialog.Message("Could not open file: %s", file).Title("Error").Error()
			return
		}

		file = filepath.Base(file)
		//TODO: Implement NpcMovementInfo properly
		ni := &pok.NpcInfo{
			Texture: e.npcImagesStrings[i],
			DialogPath: file,
			X: x,
			Y: y,
			Z: currentLayer,
			MovementInfo: pok.NpcMovementInfo{},
		}

		e.activeTileMap.PlaceNpc(ni)

		CurrentNpcDelta.npcInfo = ni
		CurrentNpcDelta.npcIndex = len(e.activeTileMap.Npcs) -1
		CurrentNpcDelta.tileMapIndex = e.activeTileMapIndex
	}
}

func (e *Editor) doRemoveNpc() {
	x := selectedTile % e.activeTileMap.Width
	y := selectedTile / e.activeTileMap.Width
	index := -1
	for i := range e.activeTileMap.NpcInfo {
		if e.activeTileMap.NpcInfo[i].X == x && e.activeTileMap.NpcInfo[i].Y == y {
			index = i
			break
		}
	}

	if index == -1 {
		return
	}

	ni := e.activeTileMap.NpcInfo[index]

	nd := &NpcDelta{
		&ni,
		len(e.activeTileMap.Npcs) - 1,
		e.activeTileMapIndex,
	}

	e.activeTileMap.RemoveNpc(nd.npcIndex)

	CurrentRemoveNpcDelta.npcDelta = nd
}

func listPngs(dir string) []string {
	return listWithExtension(dir, ".png")
}

func listWithExtension(dir string, ext string) []string {
	dirs, err := ioutil.ReadDir(dir)
	if err != nil {
		return make([]string, 0)
	}

	valid := make([]string, 0)
	for i := range dirs {
		if dirs[i].IsDir() || !strings.HasSuffix(dirs[i].Name(), ext) {
			continue
		}
		valid = append(valid, dirs[i].Name())
	}
	return valid
}

func loadImages(images []string, base string) []*ebiten.Image {
	imgs := make([]*ebiten.Image, 0, len(images))

	for _, s := range images {
		img, err := textures.LoadWithError(base + s)
		if err != nil {
			log.Println("Could not load image", s)
		}
		imgs = append(imgs, img)
	}

	return imgs
}
