package pok

import(
	"fmt"
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

func (tati *TreeAutoTileInfo) IsLeftJoinable(index int, isCrowd bool) bool {
	if isCrowd {
		for i := 0; i < CrowdTreeHeight; i++ {
			j := tati.GetCrowd(CrowdTreeWidth - 3, i)
			if j == tati.crowdArr[index] {
				return true
			}
		}
	} else {
		for i := 0; i < SingleTreeHeight; i++ {
			j:= tati.GetSingle(SingleTreeWidth - 3, i)
			if j == tati.singleArr[index] {
				return true
			}
			fmt.Println(j, index)
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
				fmt.Println("YES")
				PlaceBaselessSingularTree(tileMap, tati, x, y, depth)
				JoinTreesLeft(tileMap, tati, x, y, depth)
			}
		}

	} else {
		// Place singular tree
		PlaceSingularTree(tileMap, tati, x, y, depth)
	}
}

func JoinTreesLeft(tileMap *TileMap, tati *TreeAutoTileInfo, x, y, depth int) {
	for ty := 0; ty < SingleTreeHeight - 1; ty++ {
		tile := tati.GetCrowd(4, ty)
		ex, ey := 0 + x, ty + y
		if tileMap.Within(ex, ey) {
			index := ey * tileMap.Width + ex
			tileMap.Tiles[depth][index] = tile
		}
	}

	for ty := 0; ty < SingleTreeHeight - 1; ty++ {
		tile := tati.GetCrowd(5, ty)
		ex, ey := 1 + x, ty + y
		if tileMap.Within(ex, ey) {
			index := ey * tileMap.Width + ex
			tileMap.Tiles[depth][index] = tile
		}
	}
}

func FindNearbyTrees(tileMap *TileMap, tati *TreeAutoTileInfo, x, y, depth int) (int, int, int, bool) {
	for tx := x -1; tx < x+1; tx++ {
		for ty := y-1; ty < y+1; ty++ {
			if tileMap.Within(tx, ty) {
				// more likely to be within crowd than single
				ti := ty * tileMap.Width + tx
				for t, tv := range tati.crowdArr {
					if tv == tileMap.Tiles[depth][ti] {
						fmt.Println("MHM")
						return t, tx - x, ty - y, true
					}
				}

				// also check single
				for t, tv := range tati.singleArr {
					if tv == tileMap.Tiles[depth][ti] {
						return t, tx - x, ty - y, false
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

	fmt.Println(self.crowdArr)
	fmt.Println(self.singleArr)
}
