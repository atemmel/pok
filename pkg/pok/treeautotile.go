package pok

import(
	"image"
)

const(
	SingleTreeWidth = 4
	SingleTreeHeight = 4
	CrowdTreeWidth = 8
	CrowdTreeHeight = 6
)

type TreeAutoTileInfo struct {
	Single image.Point
	Crowd image.Point
	singleArr [SingleTreeWidth*SingleTreeHeight]int
	crowdArr [CrowdTreeWidth*CrowdTreeHeight]int
	textureWidth int
}

// Unnecesary extensibility, Lmao q8^)

func (self *TreeAutoTileInfo) IsLeftJoinable(index int, isCrowd bool) bool {
	return self.IsJoinableX(3, 1, index, isCrowd) || self.IsJoinableX(5, 1, index, isCrowd)
}

func (self *TreeAutoTileInfo) IsRightJoinable(index int, isCrowd bool) bool {
	return self.IsJoinableX(3, 1, index, isCrowd) || self.IsJoinableX(5, 1, index, isCrowd)
}

func (self *TreeAutoTileInfo) IsJoinableX(crowdX, singleX, index int, isCrowd bool) bool {
	if isCrowd {
		for i := 0; i < CrowdTreeHeight; i++ {
			j := self.GetCrowd(crowdX, i)
			if j == self.crowdArr[index] {
				return true
			}
		}
	} else {
		for i := 0; i < SingleTreeHeight; i++ {
			j:= self.GetSingle(singleX, i)
			if j == self.singleArr[index] {
				return true
			}
		}
	}

	return false
}

func (self *TreeAutoTileInfo) GetSingle(x, y int) int {
	offset := self.singleArr[0]
	return y * self.textureWidth + x + offset
}

func (self *TreeAutoTileInfo) GetCrowd(x, y int) int {
	offset := self.crowdArr[0]
	return y * self.textureWidth + x + offset
}

func PlaceTree(tileMap *TileMap, tati *TreeAutoTileInfo, x, y, depth int) {
	depth += 1
	for depth >= len(tileMap.Tiles) {
		tileMap.AppendLayer()
	}

	// Find nearby trees
	i, tx, _, crowdFound := FindNearbyTrees(tileMap, tati, x, y, depth)
	if i != -1 {
		if tx < 0 {
			if tati.IsLeftJoinable(i, crowdFound) {
				PlaceBaselessSingularTree(tileMap, tati, x, y, depth)
				JoinTreesLeft(tileMap, tati, x, y, depth)
			}
		} else if tx > 0 {
			if tati.IsRightJoinable(i, crowdFound) {
				PlaceBaselessSingularTree(tileMap, tati, x, y, depth)
				JoinTreesRight(tileMap, tati, x, y, depth)
			}
		}

	} else {
		// Place singular tree
		PlaceSingularTree(tileMap, tati, x, y, depth)
	}
}

func SelectJoinPatternfromX(x int) (int, int) {
	if x % 4 > 2 {
		return 2, 3
	}
	return 4, 5
}

func JoinTreesLeft(tileMap *TileMap, tati *TreeAutoTileInfo, x, y, depth int) {
	ltile, rtile := SelectJoinPatternfromX(x)

	for ty := 0; ty < SingleTreeHeight - 1; ty++ {
		tile := tati.GetCrowd(ltile, ty)
		ex, ey := 0 + x, ty + y
		if tileMap.Within(ex, ey) {
			index := ey * tileMap.Width + ex
			tileMap.Tiles[depth][index] = tile
		}
	}

	for ty := 0; ty < SingleTreeHeight - 1; ty++ {
		tile := tati.GetCrowd(rtile, ty)
		ex, ey := 1 + x, ty + y
		if tileMap.Within(ex, ey) {
			index := ey * tileMap.Width + ex
			tileMap.Tiles[depth][index] = tile
		}
	}
}

func JoinTreesRight(tileMap *TileMap, tati *TreeAutoTileInfo, x, y, depth int) {
	ltile, rtile := SelectJoinPatternfromX(x)

	for ty := 0; ty < SingleTreeHeight - 1; ty++ {
		tile := tati.GetCrowd(ltile, ty)
		ex, ey := 2 + x, ty + y
		if tileMap.Within(ex, ey) {
			index := ey * tileMap.Width + ex
			tileMap.Tiles[depth][index] = tile
		}
	}

	for ty := 0; ty < SingleTreeHeight - 1; ty++ {
		tile := tati.GetCrowd(rtile, ty)
		ex, ey := 3 + x, ty + y
		if tileMap.Within(ex, ey) {
			index := ey * tileMap.Width + ex
			tileMap.Tiles[depth][index] = tile
		}
	}
}

func FindNearbyTrees(tileMap *TileMap, tati *TreeAutoTileInfo, x, y, depth int) (int, int, int, bool) {
	offsetX := []int{-1, 0, 3}
	offsetY := []int{-1, 0, 3}

	for _, ox := range offsetX {
		for _, oy := range offsetY {
			if ox == 0 && oy == 0 {
				continue
			}

			tx, ty := x + ox, y + oy
			if tileMap.Within(tx, ty) {
				// more likely to be within crowd than single
				ti := ty * tileMap.Width + tx
				for t, tv := range tati.crowdArr {
					if tv == tileMap.Tiles[depth][ti] {
						return t, ox, oy, true
					}
				}

				// also check single
				for t, tv := range tati.singleArr {
					if tv == tileMap.Tiles[depth][ti] {
						return t, ox, oy, false
					}
				}
			}
		}
	}
	return -1, 0, 0, false
}

func PlaceBaselessSingularTree(tileMap *TileMap, tati *TreeAutoTileInfo, x, y, depth int) {
	for tx := 1; tx < SingleTreeWidth; tx++ {
		for ty := 0; ty < SingleTreeHeight - 1; ty++ {
			tile := tati.GetSingle(tx, ty)
			ex, ey := tx + x, ty + y
			if tileMap.Within(ex, ey) {
				index := ey * tileMap.Width + ex
				tileMap.Tiles[depth][index] = tile
			}
		}
	}
}

func PlaceSingularTree(tileMap *TileMap, tati *TreeAutoTileInfo, x, y, depth int) {
	for tx := 0; tx < SingleTreeWidth; tx++ {
		for ty := 0; ty < SingleTreeHeight; ty++ {
			tile := tati.GetSingle(tx, ty)
			ex, ey := tx + x, ty + y
			if tileMap.Within(ex, ey) {
				index := ey * tileMap.Width + ex
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

	ty := 0
	tx := 0
	for i := range self.singleArr {
		if tx > self.textureWidth {
			ty++
			tx = 0
		}
		self.singleArr[i] = ty * self.textureWidth + base + tx
		tx++
	}

	// do crowd tree
	x = self.Crowd.X
	y = self.Crowd.Y

	base = y * CrowdTreeWidth + x

	ty = 0
	tx = 0
	for i := range self.crowdArr {
		if tx > self.textureWidth {
			ty++
			tx = 0
		}
		self.crowdArr[i] = ty * self.textureWidth + base + tx
		tx++
	}

}
