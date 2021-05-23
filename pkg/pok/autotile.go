package pok

import(
	"encoding/json"
	"io/ioutil"
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
	files := listWithExtension(directory, ".ati")
	atis := make([]AutoTileInfo, 0)
	for i := range files {
		bytes, err := ioutil.ReadFile(directory + files[i])
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

func DecideTileIndicies(tileMap *TileMap, tile, depth, texture int, ati *AutoTileInfo) *AutotileDelta {
	neighbors := BuildNeighbors(tileMap, tile, depth, texture, ati)
	xStart := tile % tileMap.Width - 1
	yStart := tile / tileMap.Width - 1

	oldData := make(map[int]ModifiedTile)
	newData := make(map[int]ModifiedTile)

	oldTile := tileMap.Tiles[depth][tile]
	oldTextureIndex := tileMap.TextureIndicies[depth][tile]

	oldData[tile] = ModifiedTile{
		oldTile,
		oldTextureIndex,
	}

	ripple := func(x, y int) {
		newTile := y * tileMap.Width + x

		newNeighbors := BuildNeighbors(tileMap, newTile, depth, texture, ati)
		newIndex := DecideTileIndex(newNeighbors, ati)

		oldTile = tileMap.Tiles[depth][newTile]
		oldTextureIndex = tileMap.TextureIndicies[depth][newTile]

		if _, ok := oldData[newTile]; !ok {
			oldData[newTile] = ModifiedTile{
				oldTile,
				tileMap.TextureIndicies[depth][newTile],
			}
		}

		newData[newTile] = ModifiedTile{
			newIndex,
			texture,
		}

		tileMap.Tiles[depth][newTile] = newIndex
		tileMap.TextureIndicies[depth][newTile] = texture
	}

	tileMap.Tiles[depth][tile] = ati.Center
	tileMap.TextureIndicies[depth][tile] = texture

	newData[tile] = ModifiedTile{
		ati.Center,
		texture,
	}

	for i := range neighbors {
		for j := range neighbors[i] {
			if neighbors[i][j] != Unused {
				x := xStart + j
				y := yStart + i
				ripple(x, y)
			}
		}
	}

	neighbors = BuildNeighbors(tileMap, tile, depth, texture, ati)
	tileMap.Tiles[depth][tile] = DecideTileIndex(neighbors, ati)

	return &AutotileDelta{
		oldData,
		newData,
		0,
		0,
	}
}

func DecideTileIndex(neighbors [][]int, ati *AutoTileInfo) int {

	// If something directly above and below
	if neighbors[0][1] != Unused && neighbors[2][1] != Unused {

		// If nothing to the left and something to the right
		if neighbors[1][0] == Unused && neighbors[1][2] != Unused {
			return ati.Left
		}

		// If nothing to the right and something to the left
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
