package pok

import(
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"
)

const Unused = -1

type AutoTileInfo struct {
	UpperLeft int
	Upper int
	UpperRight int
	Left int
	Center int
	Right int
	LowerLeft int
	Lower int
	LowerRight int
	CurveUpperLeft int
	CurveUpperRight int
	CurveLowerLeft int
	CurveLowerRight int
}

func ReadAllAutoTileInfo(directory string) ([]AutoTileInfo, error) {
	dirs, err := ioutil.ReadDir(directory)
	if err != nil {
		log.Println("Could not open dir", directory)
		return make([]AutoTileInfo, 0), nil
	}

	atis := make([]AutoTileInfo, 0)
	for i := range dirs {
		if dirs[i].IsDir() || !strings.HasSuffix(dirs[i].Name(), ".ati") {
			continue
		}

		bytes, err := ioutil.ReadFile(directory + dirs[i].Name())
		if err != nil {
			return nil, err
		}

		ati := AutoTileInfo{}
		err = json.Unmarshal(bytes, &ati)
		if err != nil {
			return nil, err
		}

		atis = append(atis, ati)
	}

	return atis, nil
}

func BuildNeighbors(tileMap *TileMap, tile, depth, texture int) [][]int {
	mat := make([][]int, 0)
	xStart := tile % tileMap.Width
	yStart := tile / tileMap.Width

	for i := -1; i < 2; i++ {
		y := yStart + i
		if y < 0 || y >= tileMap.Height {
			continue
		}
		row := make([]int, 0)
		for j := -1; j < 2; j++ {
			x := xStart + j
			if x < 0 || x >= tileMap.Width {
				continue
			}

			index := y * tileMap.Width + x

			// Given a texture match, provide exact texture index
			if tileMap.TextureIndicies[depth][index] == texture {
				row = append(row, tileMap.Tiles[depth][index])
			} else {	// Otherwise, explicitly state the disinterest in this tile
				row = append(row, Unused)
			}
		}

		if len(row) > 0 {
			mat = append(mat, row)
		}
	}

	return mat
}

func DecideTileIndicies(neighbors [][]int, ati *AutoTileInfo) int {
	//TODO: Actually decide tile indices

	return ati.UpperRight
}
