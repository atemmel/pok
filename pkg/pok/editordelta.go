package pok

type Delta interface {
	Undo(ed *Editor)
	Redo(ed *Editor)
}

const preAllocDelta = 16
var UndoStack = make([]Delta, 0, preAllocDelta)
var RedoStack = make([]Delta, 0, preAllocDelta)

func PerformUndo(ed *Editor) {
	if len(UndoStack) > 0 {
		top := UndoStack[len(UndoStack) - 1]
		top.Undo(ed)
		RedoStack = append(RedoStack, top)
		UndoStack = UndoStack[:len(UndoStack) - 1]
	}
}

func PerformRedo(ed *Editor) {
	if len(RedoStack) > 0 {
		top := RedoStack[len(RedoStack)-1]
		top.Redo(ed)
		UndoStack = append(UndoStack, top)
		RedoStack = RedoStack[:len(RedoStack) - 1]
	}
}

type PencilDelta struct {
	indicies []int
	oldTiles []int
	oldTextureIndicies []int
	z int
	tileMapIndex int
	newTile int
	newTextureIndex int
}

var CurrentPencilDelta *PencilDelta = &PencilDelta{}

func (pd *PencilDelta) Undo(ed *Editor) {
	tm := ed.tileMaps[pd.tileMapIndex]
	for i, j := range pd.indicies {
		tm.Tiles[pd.z][j] = pd.oldTiles[i]
		tm.TextureIndicies[pd.z][j] = pd.oldTextureIndicies[i]
	}
}

func (pd *PencilDelta) Redo(ed *Editor) {
	tm := ed.tileMaps[pd.tileMapIndex]
	for _, j := range pd.indicies {
		tm.Tiles[pd.z][j] = pd.newTile
		tm.TextureIndicies[pd.z][j] = pd.newTextureIndex
	}
}

type EraserDelta struct {
	tiles []Vec2
	oldTiles []int
	oldTextureIndicies []int
	tileMapIndex int
}
