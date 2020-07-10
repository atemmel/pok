package pok

import (
	"errors"
	"github.com/hajimehoshi/ebiten"
	"image/color"
)

var selectionX int
var selectionY int
var m2Pressed = false
var m3Pressed = false
var copyBuffer = 0
var selectedTile = 0
var currentLayer = 0

var plusPressed = false
var minusPressed = false
var pPressed = false
var uPressed = false
var iPressed = false
var drawOnlyCurrentLayer = false
var drawUi = false

type Editor struct {
	tileMap TileMap
	rend Renderer
	selection *ebiten.Image
	collisionMarker *ebiten.Image
	exitMarker *ebiten.Image
}

func NewEditor() *Editor {
	var err error
	es := &Editor{}

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

	return es;
}

func (e *Editor) Update(screen *ebiten.Image) error {
	err := e.handleInputs()
	return err
}

func (e *Editor) Draw(screen *ebiten.Image) {
	e.tileMap.Draw(&e.rend)
	if drawUi {
		e.DrawTileMapDetail()
	}
}

func (e *Editor) DrawTileMapDetail() {
	for j := range e.tileMap.Tiles {
		if drawOnlyCurrentLayer && j != currentLayer {
			continue
		}
		for i := range e.tileMap.Tiles[j] {
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
	}

	if drawUi {
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

func (e *Editor) handleInputs() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return errors.New("")	//TODO Gotta be a better way to do this
	}

	_, dy := ebiten.Wheel()
	if dy != 0. && len(e.tileMap.Tiles[currentLayer]) > selectedTile && selectedTile >= 0 {
		if dy < 0 {
			e.tileMap.Tiles[currentLayer][selectedTile]--
		} else {
			e.tileMap.Tiles[currentLayer][selectedTile]++
		}
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButton(0)) {
		cx, cy := ebiten.CursorPosition();
		e.SelectTileFromMouse(cx, cy)
	}

	if !m2Pressed && ebiten.IsMouseButtonPressed(ebiten.MouseButton(1)) {
		m2Pressed = true
		cx, cy := ebiten.CursorPosition();
		e.SelectTileFromMouse(cx, cy)
		if 0 <= selectedTile && selectedTile < len(e.tileMap.Tiles[currentLayer]) {
			e.tileMap.Collision[currentLayer][selectedTile] = !e.tileMap.Collision[currentLayer][selectedTile]
		}
	} else if !ebiten.IsMouseButtonPressed(ebiten.MouseButton(1)) {
		m2Pressed = false
	}

	if !m3Pressed && ebiten.IsMouseButtonPressed(ebiten.MouseButton(2)) {
		m3Pressed = true
		cx, cy := ebiten.CursorPosition();
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
	} else if !ebiten.IsMouseButtonPressed(ebiten.MouseButton(2)) {
		m3Pressed = false
	}

	if ebiten.IsKeyPressed(ebiten.KeyC) {
		if 0 <= selectedTile && selectedTile < len(e.tileMap.Tiles[currentLayer]) {
			copyBuffer = e.tileMap.Tiles[currentLayer][selectedTile]
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyV) {
		if 0 <= selectedTile && selectedTile < len(e.tileMap.Tiles[currentLayer]) {
			e.tileMap.Tiles[currentLayer][selectedTile] = copyBuffer
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyMinus) && !plusPressed {	// Plus
		if currentLayer + 1 < len(e.tileMap.Tiles) {
			currentLayer++
		}
		plusPressed = true
	} else if !ebiten.IsKeyPressed(ebiten.KeyMinus) {
		plusPressed = false
	}

	if ebiten.IsKeyPressed(ebiten.KeySlash) && !minusPressed {	// Minus
		if currentLayer > 0 {
			currentLayer--
		}
		minusPressed = true
	} else if !ebiten.IsKeyPressed(ebiten.KeySlash) {
		minusPressed = false
	}

	if ebiten.IsKeyPressed(ebiten.KeyP) && !pPressed {	// Minus
		e.tileMap.Tiles = append(e.tileMap.Tiles, make([]int, len(e.tileMap.Tiles[0])))
		e.tileMap.Collision = append(e.tileMap.Collision, make([]bool, len(e.tileMap.Collision[0])))
		pPressed = true
	} else if !ebiten.IsKeyPressed(ebiten.KeyP) {
		pPressed = false
	}

	if ebiten.IsKeyPressed(ebiten.KeyU) && !uPressed {
		drawOnlyCurrentLayer = !drawOnlyCurrentLayer
		uPressed = true
	} else if !ebiten.IsKeyPressed(ebiten.KeyU) {
		uPressed = false
	}

	if ebiten.IsKeyPressed(ebiten.KeyI) && !iPressed {
		drawUi = !drawUi
		iPressed = true
	} else if !ebiten.IsKeyPressed(ebiten.KeyI) {
		iPressed = false
	}

	return nil
}

func (e *Editor) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return DisplaySizeX, DisplaySizeY
}
