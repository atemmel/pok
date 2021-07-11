package main

import(
	"encoding/json"
	"flag"
	"fmt"
	"github.com/atemmel/pok/pkg/constants"
	"github.com/atemmel/pok/pkg/pok"
	"io/ioutil"
	"image"
	"image/color"
	"image/png"
	"os"
	"strconv"
	"strings"
)

var shouldOutputImage bool = false

func init() {
	flag.BoolVar(
		&shouldOutputImage,
		"image",
		false,
		"Output an example image describing which objects were found",
	)
	flag.Parse()
}

func openImage(path string) image.Image {
	handle, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	img, _, err := image.Decode(handle)
	if err != nil {
		panic(err)
	}
	return img
}

func saveImage(img image.Image, path string) {
	handle, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	err = png.Encode(handle, img)
	if err != nil {
		panic(err)
	}
}

func inAnyBox(boxes []image.Rectangle, x, y int) bool {
	pt := image.Pt(x, y)
	for i := range boxes {
		if pt.In(boxes[i]) {
			return true
		}
	}
	return false
}

func expandBox(rect image.Rectangle, pt image.Point) image.Rectangle {
	if pt.X < rect.Min.X {
		rect.Min.X = pt.X
	}
	if pt.X > rect.Max.X {
		rect.Max.X = pt.X
	}
	if pt.Y < rect.Min.Y {
		rect.Min.Y = pt.Y
	}
	if pt.Y > rect.Max.Y {
		rect.Max.Y = pt.Y
	}

	return rect
}

func neighbour(pt image.Point, img image.Image) bool {
	x, y := pt.X, pt.Y
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	return (x >= 0 && x < w) && (y >= 0 && y < h) && img.At(x, y) != color.NRGBA{0, 0, 0,0}
}

func getNeighbours(pt image.Point, img image.Image) []image.Point {
	neighbours := make([]image.Point, 0, 4)

	pts := []image.Point{
		{pt.X + 1, pt.Y},
		{pt.X - 1, pt.Y},
		{pt.X, pt.Y + 1},
		{pt.X, pt.Y - 1},
	}

	for _, p := range pts {
		if neighbour(p, img) {
			neighbours = append(neighbours, p)
		}
	}

	return neighbours
}

func exploreBoundingBox(x, y int, img image.Image, visited []bool) image.Rectangle {
	rect := image.Rect(x, y, x, y)
	queue := make([]image.Point, 0, 32)
	w := img.Bounds().Dx()

	visitedIndex := func(x, y int) int {
		return y * w + x
	}

	visited[visitedIndex(x, y)] = true
	queue = append(queue, image.Pt(x, y))

	for len(queue) > 0 {
		pt := queue[0]
		queue = queue[1:]
		rect = expandBox(rect, pt)

		neighbours := getNeighbours(pt, img)
		for _, n := range neighbours {
			if !visited[visitedIndex(n.X, n.Y)] {
				visited[visitedIndex(n.X, n.Y)] = true
				queue = append(queue, n)
			}
		}
	}

	return rect
}

func process(img image.Image) []image.Rectangle {
	boxes := make([]image.Rectangle, 0)
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	visited := make([]bool, w*h)

	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			clr := img.At(x, y)
			transparent := color.NRGBA{0,0,0,0}
			if clr != transparent && !inAnyBox(boxes, x, y) {
				box := exploreBoundingBox(
					x,
					y,
					img,
					visited,
				)

				if box.Dx() == 0 || box.Dy() == 0 {
					continue
				}

				boxes = append(boxes, box)
			}
		}
	}

	return boxes
}

func outputImage(boxes []image.Rectangle, edobjs []pok.EditorObject, img image.Image, str string) {
	result := image.NewNRGBA(img.Bounds())
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			result.Set(x, y, img.At(x, y))
		}
	}

	gridColor := color.NRGBA{0, 255, 0, 255}
	for x := constants.TileSize; x < w + constants.TileSize; x += constants.TileSize {
		for y := 0; y < h; y++ {
			result.Set(x, y, gridColor)
		}
	}

	for y := constants.TileSize; y < h + constants.TileSize; y += constants.TileSize {
		for x := 0; x < w; x++ {
			result.Set(x, y, gridColor)
		}
	}

	markerColor := color.NRGBA{255, 0, 0, 255}
	for _, box := range boxes {
		x0 := box.Min.X
		x1 := box.Max.X
		y0 := box.Min.Y
		y1 := box.Max.Y

		for x := x0; x < x1; x++ {
			result.Set(x, y0, markerColor)
			result.Set(x, y1, markerColor)
		}
		for y := y0; y < y1; y++ {
			result.Set(x0, y, markerColor)
			result.Set(x1, y, markerColor)
		}
	}

	otherColor := color.NRGBA{0, 0, 255, 255}
	for i := range edobjs {
		x0 := edobjs[i].X * constants.TileSize
		y0 := edobjs[i].Y * constants.TileSize
		x1 := (edobjs[i].X + edobjs[i].W) * constants.TileSize
		y1 := (edobjs[i].Y + edobjs[i].H) * constants.TileSize

		for x := x0; x < x1; x++ {
			result.Set(x, y0, otherColor)
			result.Set(x, y1, otherColor)
		}
		for y := y0; y < y1; y++ {
			result.Set(x0, y, otherColor)
			result.Set(x1, y, otherColor)
		}
	}

	index := strings.LastIndex(str, ".")
	outpath := ""
	if index == -1 {
		outpath = str + "_marked.png"
	} else {
		outpath = str[:index] + "_marked.png"
	}

	saveImage(result, outpath)
}

func writeEditorObjectToFile(edobj *pok.EditorObject, path string) {
	bytes, err := json.Marshal(edobj)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(path, bytes, 0666)
	if err != nil {
		panic(err)
	}
}

func outputEditorObjects(boxes []image.Rectangle, s string) []pok.EditorObject {
	edobj := pok.EditorObject{
		Texture: "",
		X: 0,
		Y: 0,
		W: 0,
		H: 0,
		Z: nil,
	}

	index := strings.LastIndex(s, "/")
	if index != -1 {
		s = s[index + 1:]
	}

	edobj.Texture = s

	edobjs := make([]pok.EditorObject, 0, len(boxes))

	for _, box := range boxes {
		edobj.X = box.Min.X / constants.TileSize
		edobj.Y = box.Min.Y / constants.TileSize

		dx, dy := box.Dx(), box.Dy()
		edobj.W = dx / constants.TileSize
		edobj.H = dy / constants.TileSize
		if dx % constants.TileSize != 0 {
			edobj.W++
		}
		if dy % constants.TileSize != 0 {
			edobj.H++
		}

		edobj.Z = make([]int, edobj.W*edobj.H)
		for j := range edobj.Z {
			if j < edobj.W {
				edobj.Z[j] = 2
			} else {
				edobj.Z[j] = 1
			}
		}

		edobjs = append(edobjs, edobj)
	}
	return edobjs
}

func main() {
	args := flag.Args()

	for _, s := range args {
		img := openImage(s)
		boxes := process(img)
		edobjs := outputEditorObjects(boxes, s)
		if shouldOutputImage {
			outputImage(boxes, edobjs, img, s)
			fmt.Println(boxes, edobjs)
		}

		preStr := ""
		index := strings.Index(s, ".")
		if index == -1 {
			preStr = s
		} else {
			preStr = s[:index]
		}

		for i := range edobjs {
			edobj := &edobjs[i]
			edobjPath := preStr + strconv.Itoa(i + 1) + ".edobj"
			writeEditorObjectToFile(edobj, edobjPath)
		}
	}
}
