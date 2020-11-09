package pok

import(
	"encoding/json"
	"io/ioutil"
	"strings"
)

type EditorObject struct {
	Texture string
	X, Y int
	W, H int
	Z []int

	textureIndex int
}

type PlacedEditorObject struct {
	X, Y int
	Index int
}

func ReadAllObjects(directory string) ([]EditorObject, error) {
	dirs, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	objs := make([]EditorObject, 0)
	for i := range dirs {
		if dirs[i].IsDir() || !strings.HasSuffix(dirs[i].Name(), ".edobj") {
			continue
		}

		bytes, err := ioutil.ReadFile(directory + dirs[i].Name())
		if err != nil {
			return nil, err
		}

		obj := EditorObject{}
		err = json.Unmarshal(bytes, &obj)
		if err != nil {
			return nil, err
		}

		objs = append(objs, obj)
	}

	return objs, nil
}
