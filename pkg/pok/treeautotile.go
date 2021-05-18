package pok

import(
	"image"
)

const(
	SingleTreeWidth = 4
	SingleTreeHeight = 4
	CrowdTreeWidth = 8
	CrowdTreeHeight = 6
	CrowdTreeSpaceX = 2
	CrowdTreeSpaceY = 2

	TreeDepthOffset = 3
)

type TreeAutoTileInfo struct {
	Single image.Point
	Crowd image.Point
	single int
	crowd int
	textureWidth int
}

func (self *TreeAutoTileInfo) FillArea(tm *TileMap, x, y, nx, ny, depth int) {
	outerRightBorder := tati.GetCrowd(CrowdTreeWidth - 2, 2)
	innerRightBorder := tati.GetCrowd(CrowdTreeWidth - 1, 2)

	for j := 0; j < ny; j++ {
		ypos := y + CrowdTreeSpaceY * j

		for i := 0; i < nx; i++ {
			xpos := x + CrowdTreeSpaceX * i

			self.PlaceSingularTree(tm, xpos, ypos, depth)

			if i > 0 {
				if j == ny - 1 {
					self.DoTreeDownBorder(tm, xpos, ypos, depth)
					self.JoinTreesLeft(tm, xpos, ypos, depth)
				} else {
					self.JoinTreesLeftDown(tm, xpos, ypos, depth)
				}
			}

		}

		if j > 0 {
			self.DoTreeLeftBorder(tm, x, ypos - CrowdTreeSpaceY, depth)

			for i := 0; i < nx; i++ {
				xpos := x + CrowdTreeSpaceX * i
				self.JoinTreesUp(tm, self, xpos, ypos, depth)
			}

			ex1 := x + (nx - 1) * CrowdTreeSpaceX + 2
			ex2 := x + (nx - 1) * CrowdTreeSpaceX + 3
			ey := ypos

			if tm.Contains(ex1, ey) {
				index := ey * tm.Width + ex1
				tm.Tiles[depth][index] = outerRightBorder
			}
			if tm.Contains(ex2, ey) {
				index := ey * tm.Width + ex2
				tm.Tiles[depth][index] = innerRightBorder
			}
		}
	}

	ypos := y + CrowdTreeSpaceY * ny + CrowdTreeSpaceY - 1
	for i := 0; i < nx; i++ {
		xpos := x + CrowdTreeSpaceX * i + 1
		if tm.Contains(xpos, ypos) {
			index := ypos * tm.Width + xpos
			tile := tm.Tiles[depth][index]
			tm.Tiles[depth][index] = -1
			tm.Tiles[depth - 2][index] = tile
		}

		xpos++
		if tm.Contains(xpos, ypos) {
			index := ypos * tm.Width + xpos
			tile := tm.Tiles[depth][index]
			tm.Tiles[depth][index] = -1
			tm.Tiles[depth - 2][index] = tile
		}
	}

	w := nx * CrowdTreeSpaceX
	h := ny * CrowdTreeSpaceY

	tm.FillCollision(x + 1, y + 2, w, h, depth - TreeDepthOffset)
}

func (self *TreeAutoTileInfo) GetSingle(x, y int) int {
	offset := self.single
	return y * self.textureWidth + x + offset
}

func (self *TreeAutoTileInfo) GetCrowd(x, y int) int {
	offset := self.crowd
	return y * self.textureWidth + x + offset
}

func SelectJoinPatternFromX(x int) (int, int) {
	// every other front and back
	/*
	if x % 4 > 1 {
		return 2, 3
	}
	return 4, 5
	*/

	// constantly overlapping
	return 2, 3
}

func (self *TreeAutoTileInfo) JoinTreesUp(tileMap *TileMap, tati *TreeAutoTileInfo, x, y, depth int) {
	for tx := 1; tx < SingleTreeWidth - 1; tx++ {
		tile := self.GetCrowd(tx, 2)
		ex, ey := tx + x, 0 + y
		if tileMap.Contains(ex, ey) {
			index := tileMap.Index(ex, ey)
			tileMap.Tiles[depth][index] = tile
		}
	}

	for tx := 1; tx < SingleTreeWidth - 1; tx++ {
		tile := self.GetCrowd(tx, 3)
		ex, ey := tx + x, 1 + y
		if tileMap.Contains(ex, ey) {
			index := tileMap.Index(ex, ey)
			tileMap.Tiles[depth][index] = tile
		}
	}
}

func (self *TreeAutoTileInfo) JoinTreesLeftDown(tileMap *TileMap, x, y, depth int) {
	ltile, rtile := SelectJoinPatternFromX(x)

	for ty := 0; ty < SingleTreeHeight - 1; ty++ {
		tile := self.GetCrowd(ltile, ty)
		ex, ey := 0 + x, ty + y
		if tileMap.Contains(ex, ey) {
			index := tileMap.Index(ex, ey)
			tileMap.Tiles[depth][index] = tile
		}
	}

	for ty := 0; ty < SingleTreeHeight - 1; ty++ {
		tile := tati.GetCrowd(rtile, ty)
		ex, ey := 1 + x, ty + y
		if tileMap.Contains(ex, ey) {
			index := tileMap.Index(ex, ey)
			tileMap.Tiles[depth][index] = tile
		}
	}
}

func (self *TreeAutoTileInfo) JoinTreesLeft(tileMap *TileMap, x, y, depth int) {
	tile1, tile2 := SelectJoinPatternFromX(x)

	for ty := 0; ty < SingleTreeHeight - 2; ty++ {
		ltile := self.GetCrowd(tile1, ty)
		rtile := self.GetCrowd(tile2, ty)
		ex, ey := 0 + x, ty + y
		if tileMap.Contains(ex, ey) {
			index := tileMap.Index(ex, ey)
			tileMap.Tiles[depth][index] = ltile
		}
		ex += 1
		if tileMap.Contains(ex, ey) {
			index := tileMap.Index(ex, ey)
			tileMap.Tiles[depth][index] = rtile
		}
	}

	ty := SingleTreeHeight - 2
	ltile := tati.GetCrowd(tile1, ty + 2)
	rtile := tati.GetCrowd(tile2, ty + 2)
	ex, ey := 0 + x, ty + y
	if tileMap.Contains(ex, ey) {
		index := tileMap.Index(ex, ey)
		tileMap.Tiles[depth][index] = ltile
	}
	ex += 1
	if tileMap.Contains(ex, ey) {
		index := tileMap.Index(ex, ey)
		tileMap.Tiles[depth][index] = rtile
	}
}

func (self *TreeAutoTileInfo) DoTreeDownBorder(tileMap *TileMap, x, y, depth int) {
	for tx := 1; tx < SingleTreeWidth; tx++ {
		tile := self.GetCrowd(tx, 5)
		ex, ey := tx + x - 2, y + 3
		if tileMap.Contains(ex, ey) {
			index := tileMap.Index(ex, ey)
			tileMap.Tiles[depth][index] = tile
		}
	}
}

func (self *TreeAutoTileInfo) DoTreeLeftBorder(tileMap *TileMap, x, y, depth int) {
	tile := self.GetCrowd(0, 2)
	ex, ey := x, y + 2
	if tileMap.Contains(ex, ey) {
		index := tileMap.Index(ex, ey)
		tileMap.Tiles[depth][index] = tile
	}
}

func (self *TreeAutoTileInfo) PlaceSingularTree(tileMap *TileMap, x, y, depth int) {
	for tx := 0; tx < SingleTreeWidth; tx++ {
		for ty := 0; ty < SingleTreeHeight; ty++ {
			tile := self.GetSingle(tx, ty)
			ex, ey := tx + x, ty + y
			if tileMap.Contains(ex, ey) {
				index := tileMap.Index(ex, ey)
				tileMap.Tiles[depth][index] = tile
			}
		}
	}
}

func (self *TreeAutoTileInfo) prepare() {
	self.Single = image.Point{0, 38*2}
	self.Crowd = image.Point{0, 16*2}
	self.textureWidth = 128 / TileSize

	// do single tree
	x := self.Single.X
	y := self.Single.Y

	base := y * SingleTreeWidth + x

	self.single = base

	// do crowd tree
	x = self.Crowd.X
	y = self.Crowd.Y

	base = y * CrowdTreeWidth + x

	self.crowd = base
}
