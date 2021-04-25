package pok

type Delta interface {
	Undo(ed *Editor)
	Redo(ed *Editor)
}

const preAllocDelta = 16
var UndoStack = make([]Delta, 0, preAllocDelta)
var RedoStack = make([]Delta, 0, preAllocDelta)

var CurrentPencilDelta *PencilDelta = &PencilDelta{}
var CurrentEraserDelta *EraserDelta = &EraserDelta{}
var CurrentObjectDelta *ObjectDelta = &ObjectDelta{}
var CurrentLinkDelta *LinkDelta = &LinkDelta{}

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


func (dp *PencilDelta) Undo(ed *Editor) {
	tm := ed.tileMaps[dp.tileMapIndex]
	for i, j := range dp.indicies {
		tm.Tiles[dp.z][j] = dp.oldTiles[i]
		tm.TextureIndicies[dp.z][j] = dp.oldTextureIndicies[i]
	}
}

func (dp *PencilDelta) Redo(ed *Editor) {
	tm := ed.tileMaps[dp.tileMapIndex]
	for _, j := range dp.indicies {
		tm.Tiles[dp.z][j] = dp.newTile
		tm.TextureIndicies[dp.z][j] = dp.newTextureIndex
	}
}

type EraserDelta struct {
	indicies []int
	oldTiles []int
	oldTextureIndicies []int
	z int
	tileMapIndex int
	newTextureIndex int
}

func (de *EraserDelta) Undo(ed *Editor) {
	tm := ed.tileMaps[de.tileMapIndex]
	for i, j := range de.indicies {
		tm.Tiles[de.z][j] = de.oldTiles[i]
		tm.TextureIndicies[de.z][j] = de.oldTextureIndicies[i]
	}
}

func (de *EraserDelta) Redo(ed *Editor) {
	tm := ed.tileMaps[de.tileMapIndex]
	for _, j := range de.indicies {
		tm.Tiles[de.z][j] = -1
		tm.TextureIndicies[de.z][j] = 0
	}
}

type ObjectDelta struct {
	placedObjectIndex int
	objectIndex int
	tileMapIndex int
	origin int
	z int
}

func (do *ObjectDelta) Undo(ed *Editor) {
	tm := ed.tileMaps[do.tileMapIndex]
	pobj := placedObjects[do.tileMapIndex][do.placedObjectIndex]
	obj := &ed.objectGrid.objs[do.objectIndex]
	tm.EraseObject(pobj, obj)
	copy(placedObjects[do.tileMapIndex][do.placedObjectIndex:], placedObjects[do.tileMapIndex][do.placedObjectIndex+1:])
	placedObjects[do.tileMapIndex] = placedObjects[do.tileMapIndex][:len(placedObjects[do.tileMapIndex])-1]
}

func (do *ObjectDelta) Redo(ed *Editor) {
	tm := ed.tileMaps[do.tileMapIndex]
	obj := &ed.objectGrid.objs[do.objectIndex]
	tm.InsertObject(obj, do.placedObjectIndex, do.origin, do.z, &placedObjects[do.tileMapIndex])
}

type LinkDelta struct {
	linkBegin *LinkData
	linkEnd *LinkData
	linkIdA int
	linkIdB int
}

func (dl *LinkDelta) Undo(ed *Editor) {
	tmA := ed.tileMaps[dl.linkBegin.TileMapIndex]
	tmB := ed.tileMaps[dl.linkEnd.TileMapIndex]

	for i := range tmA.Exits {
		if tmA.Exits[i].Id == dl.linkIdA {
			ed.removeLink(dl.linkBegin.TileMapIndex, i)
			break
		}
	}

	for i := range tmB.Exits {
		if tmB.Exits[i].Id == dl.linkIdB {
			ed.removeLink(dl.linkEnd.TileMapIndex, i)
			break
		}
	}

}

func (dl *LinkDelta) Redo(ed *Editor) {
	ed.tryConnectTileMaps(dl.linkBegin, dl.linkEnd)
}
