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

func (ati *AutoTileInfo) HasIndex(index int) bool {
	return ati.UpperLeft == index || ati.Upper == index || ati.UpperRight == index || ati.Left == index || ati.Center == index || ati.Right == index || ati.LowerLeft == index || ati.Lower == index || ati.LowerRight == index || ati.CurveUpperLeft == index || ati.CurveUpperRight == index || ati.CurveLowerLeft == index || ati.CurveLowerRight == index
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

func BuildNeighbors(tileMap *TileMap, tile, depth, texture int, ati *AutoTileInfo) [][]int {
	mat := make([][]int, 0)
	xStart := tile % tileMap.Width
	yStart := tile / tileMap.Width

	for i := -1; i < 2; i++ {
		row := make([]int, 0)
		y := yStart + i
		if y < 0 || y >= tileMap.Height {
			row := append(row, Unused, Unused, Unused)
			mat = append(mat, row)
			continue
		}
		for j := -1; j < 2; j++ {
			x := xStart + j
			if x < 0 || x >= tileMap.Width {
				row = append(row, Unused)
				continue
			}

			index := y * tileMap.Width + x

			// Given a texture match, provide exact texture index
			if tileMap.TextureIndicies[depth][index] == texture && ati.HasIndex(tileMap.Tiles[depth][index]) {
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

	// If something directly above and below
	if neighbors[0][1] != Unused && neighbors[2][1] != Unused {
		if neighbors[1][0] == Unused && neighbors[1][2] != Unused {
			return ati.Left
		}

		if neighbors[1][2] == Unused && neighbors[1][0] != Unused {
			return ati.Right
		}

		// If something directly left and right
		if neighbors[1][2] != Unused && neighbors[1][0] != Unused {
			// If something in upper corners
			if neighbors[0][0] != Unused && neighbors[0][2] != Unused {
				if neighbors[2][0] == Unused {
					return ati.CurveLowerLeft
				} else if neighbors[2][2] == Unused {
					return ati.CurveLowerRight
				}
			} else if neighbors[2][0] != Unused || neighbors[2][2] != Unused {
				if neighbors[0][0] == Unused {
					return ati.CurveUpperLeft
				} else if neighbors[0][2] == Unused {
					return ati.CurveUpperRight
				}
			}
		}
	}

	// If nothing above but something below
	if neighbors[0][1] == Unused && neighbors[2][1] != Unused {
		// If something to right but nothing to the left
		if neighbors[1][2] != Unused && neighbors[1][0] == Unused {
			return ati.UpperLeft
		} else if neighbors[1][2] == Unused && neighbors[1][0] != Unused {
			return ati.UpperRight
		}
		return ati.Upper
	}

	if neighbors[2][1] == Unused {
		if neighbors[1][2] != Unused && neighbors[1][0] == Unused {
			return ati.LowerLeft
		} else if neighbors[1][2] == Unused && neighbors[1][0] != Unused {
			return ati.LowerRight
		}
		return ati.Lower
	}

	return ati.Center
}
