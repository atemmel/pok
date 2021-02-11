package pok

/*
import(
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"
)
*/

import(
	"image"
)

const(
	SingleTreeWidth = 4
	SingleTreeHeight = 3
	CrowdTreeWidth = 8
	CrowdTreeHeight = 6
)

type TreeAutoTileInfo struct {
	Single image.Rectangle
	Crowd image.Rectangle
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
	// do single tree
	x := self.Single.Min.X
	y := self.Single.Min.Y
	w := self.Single.Dx()

	base := y * self.textureWidth + x

	ty := 0
	tx := 0
	for i := range self.singleArr {
		if tx > w {
			ty++
			tx = 0
		}
		self.singleArr[i] = ty * self.textureWidth + base + tx
		tx++
	}

	// do crowd tree
	x = self.Crowd.Min.X
	y = self.Crowd.Min.Y
	w = self.Single.Dx()

	base = y * self.textureWidth + x

	ty = 0
	tx = 0
	for i := range self.crowdArr {
		if tx > w {
			ty++
			tx = 0
		}
		self.crowdArr[i] = ty * self.textureWidth + base + tx
		tx++
	}
}
