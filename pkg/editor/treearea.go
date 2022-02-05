package editor

import (
	"github.com/atemmel/pok/pkg/pok"
	"github.com/atemmel/pok/pkg/constants"
	"image/color"
)

type TreeAreaSelection struct {
	BeginX, BeginY *int
	EndX, EndY int
	TreeInfo *TreeAutoTileInfo
}

func (t *TreeAreaSelection) Hold(x, y int) {
	if t.BeginX == nil && t.BeginY == nil {
		t.BeginX = &x
		t.BeginY = &y
	} else {
		t.EndX = x
		t.EndY = y
	}
}

func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

func (t *TreeAreaSelection) Draw(rend *pok.Renderer, offset Vec2) {
	if !t.IsHolding() {
		return
	}

	const cornerOffset = 4

	lineX, lineY := t.CountBoundingTiles()
	px, py := t.Polarity()

	x0 := float64(*t.BeginX)
	y0 := float64(*t.BeginY)

	x0 = (x0 * constants.TileSize + offset.X)
	y0 = (y0 * constants.TileSize + offset.Y)

	x1 := x0 + float64(lineX * px * constants.TileSize)
	y1 := y0 + float64(lineY * py * constants.TileSize)

	if x1 < x0 {
		x1, x0 = x0, x1
	}

	if y1 < y0 {
		y1, y0 = y0, y1
	}

	clr := color.RGBA{255, 0, 0, 255}

	line := pok.DebugLine{}
	line.Clr = clr
	line.X1 = x0 + cornerOffset
	line.Y1 = y0
	line.X2 = x1
	line.Y2 = y0

	rend.DrawLine(line)

	line.X1 = x0
	line.Y1 = y1
	line.X2 = x1 - cornerOffset
	line.Y2 = y1

	rend.DrawLine(line)

	line.X1 = x0
	line.Y1 = y0 + cornerOffset
	line.X2 = x0
	line.Y2 = y1

	rend.DrawLine(line)

	line.X1 = x1
	line.Y1 = y0
	line.X2 = x1
	line.Y2 = y1 - cornerOffset

	rend.DrawLine(line)

	t.drawCorner(x0, y0, cornerOffset, clr, rend)
	if x1 == x0 && y1 == y0 {
		x1 += constants.TileSize
		y1 += constants.TileSize
	}
	t.drawCorner(x1, y1, cornerOffset, clr, rend)
}

func (t *TreeAreaSelection) drawCorner(x0, y0, cornerOffset float64, clr color.RGBA, rend *pok.Renderer) {

	line := pok.DebugLine{}
	line.Clr = clr

	line.X1 = x0
	line.Y1 = y0 - cornerOffset
	line.X2 = x0 + cornerOffset
	line.Y2 = y0

	rend.DrawLine(line)

	line.X1 = x0 + cornerOffset
	line.Y1 = y0
	line.X2 = x0
	line.Y2 = y0 + cornerOffset

	rend.DrawLine(line)

	line.X1 = x0
	line.Y1 = y0 + cornerOffset
	line.X2 = x0 - cornerOffset
	line.Y2 = y0

	rend.DrawLine(line)

	line.X1 = x0 - cornerOffset
	line.Y1 = y0
	line.X2 = x0
	line.Y2 = y0 - cornerOffset

	rend.DrawLine(line)
}

func (t *TreeAreaSelection) CountBoundingTiles() (int, int) {
	lineX := abs(*t.BeginX - t.EndX)
	lineY := abs(*t.BeginY - t.EndY)

	if lineX >= SingleTreeWidth {
		lineX = SingleTreeWidth + ((lineX - SingleTreeWidth) / CrowdTreeSpaceX) * CrowdTreeSpaceX
	} else {
		lineX = 0
	}

	if lineY >= SingleTreeHeight {
		lineY = SingleTreeHeight + ((lineY - SingleTreeHeight) / CrowdTreeSpaceY) * CrowdTreeSpaceY
	} else {
		lineY = 0
	}

	return lineX, lineY
}

func (t *TreeAreaSelection) CountBoundingTrees() (int, int) {
	lineX := abs(*t.BeginX - t.EndX)
	lineY := abs(*t.BeginY - t.EndY)

	if lineX >= SingleTreeWidth {
		lineX = 1 + ((lineX - SingleTreeWidth) / CrowdTreeSpaceX)
	} else {
		lineX = 0
	}

	if lineY >= SingleTreeHeight {
		lineY = 1 + ((lineY - SingleTreeHeight) / CrowdTreeSpaceY)
	} else {
		lineY = 0
	}

	return lineX, lineY
}

func (t *TreeAreaSelection) Polarity() (int, int) {
	xp, yp := 1, 1
	if *t.BeginX > t.EndX {
		xp = -1
	}

	if *t.BeginY > t.EndY {
		yp = -1
	}

	return xp, yp
}

func (t *TreeAreaSelection) IsHolding() bool {
	return t.BeginX != nil && t.BeginY != nil
}

func (t *TreeAreaSelection) Release(tm *pok.TileMap, depth int) {
	if t.BeginX == nil || t.BeginY == nil {
		return
	}

	t.PopulateWithTrees(tm, depth)
	t.BeginX = nil
	t.BeginY = nil
}

func (t *TreeAreaSelection) PopulateWithTrees(tm *pok.TileMap, depth int) {
	tx, ty := t.CountBoundingTiles()

	if tx <= 0 || ty <= 0 {
		return
	}

	nx, ny := t.CountBoundingTrees()

	px, py := t.Polarity()
	rx, ry := *t.BeginX + tx * px, *t.BeginY + ty * py

	x0, y0 := *t.BeginX, *t.BeginY

	if px == -1 {
		x0 = rx
	}

	if py == -1 {
		y0 = ry
	}

	for len(tm.Tiles) <= depth + TreeDepthOffset {
		tm.AppendLayer()
	}

	t.TreeInfo.FillArea(tm, x0, y0, nx, ny, depth + TreeDepthOffset)
}
