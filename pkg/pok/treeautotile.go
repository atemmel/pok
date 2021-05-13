package pok

import(
	"image"
)

const(
	SingleTreeWidth = 4
	SingleTreeHeight = 4
	CrowdTreeWidth = 8
	CrowdTreeHeight = 8
)

type TreeAutoTileInfo struct {
	Single image.Point
	Crowd image.Point
	singleArr [SingleTreeWidth*SingleTreeHeight]int
	crowdArr [CrowdTreeWidth*CrowdTreeHeight]int
	textureWidth int
}

func (tati *TreeAutoTileInfo) HasIndex(index int) bool {
	// Prefer checking the crowdArr, as it has a higher probability of occuring
	for _, i := range tati.crowdArr {
		if i == index {
			return true
		}
	}

	for _, i := range tati.singleArr {
		if i == index {
			return true
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
	// Place singular tree
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
	//self.Single = image.Rect(8, 608, 56, 672)
	self.Single = image.Point{0, 38*2}
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
}
