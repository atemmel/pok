package pok

import (
	"errors"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/inpututil"
	"image/color"
)

var selectionX int
var selectionY int
var copyBuffer = 0
var selectedTile = 0
var currentLayer = 0

var drawOnlyCurrentLayer = false
var drawUi = false

type Editor struct {
	tileMap TileMap
	rend Renderer
	dialog DialogBox
	selection *ebiten.Image
	collisionMarker *ebiten.Image
	exitMarker *ebiten.Image
	activeFile string
	tw typewriter
}

type typewriterResult int

const (
	None typewriterResult = 0
	Success typewriterResult = 1
	Abort typewriterResult = 2
)

type typewriter struct {
	Mode bool
	Input string
	Result typewriterResult
}

func (tw *typewriter) Start() {
	tw.Mode = true
	tw.Result = None
	tw.Input = ""
}

func (tw *typewriter) HandleInputs() {
	tw.Input += string(ebiten.InputChars())
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		if len(tw.Input) > 0 {
			tw.Input = tw.Input[:len(tw.Input)-1]
		}
	} else if ebiten.IsKeyPressed(ebiten.KeyEnter) {
		tw.Result = Success
		tw.Mode = false;
	} else if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		tw.Result = Abort
		tw.Mode = false;
	}
}

func NewEditor() *Editor {
	var err error
	es := &Editor{}

	es.dialog = NewDialogBox()

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
	if e.tw.Mode {
		e.tw.HandleInputs();
		e.dialog.SetString("Enter name of file to open:\n" + e.tw.Input);
		if e.tw.Result != None {
			e.dialog.Hidden = true
		}
		return nil
	}
	err := e.handleInputs()
	return err
}

func (e *Editor) Draw(screen *ebiten.Image) {
	e.tileMap.Draw(&e.rend)
	if drawUi {
		e.DrawTileMapDetail()
	}
	e.dialog.Draw(screen)
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
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return errors.New("")	//TODO Gotta be a better way to do this
	}

	if e.activeFile != "" {
		e.handleInputs()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyI) {
		drawUi = !drawUi
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyO) {
		e.dialog.Hidden = false
		e.tw.Start()

	}

	return nil
}

func (e *Editor) handleMapInputs() {
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

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton(1)) {
		cx, cy := ebiten.CursorPosition();
		e.SelectTileFromMouse(cx, cy)
		if 0 <= selectedTile && selectedTile < len(e.tileMap.Tiles[currentLayer]) {
			e.tileMap.Collision[currentLayer][selectedTile] = !e.tileMap.Collision[currentLayer][selectedTile]
		}
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton(2)) {
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
		e.tileMap.Collision = append(e.tileMap.Collision, make([]bool, len(e.tileMap.Collision[0])))
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyU) {
		drawOnlyCurrentLayer = !drawOnlyCurrentLayer
	}
}

func (e *Editor) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return DisplaySizeX, DisplaySizeY
}
