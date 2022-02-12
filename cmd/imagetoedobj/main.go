package main

import(
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"github.com/atemmel/pok/pkg/constants"
	"github.com/atemmel/pok/pkg/editor"
	"image"
	"os"
)

var outputDir string
var filesToProcess []string

func processFile(path string) error {
	img := openImage(path)

	pixelWidth := img.Bounds().Max.X
	pixelHeight := img.Bounds().Max.Y

	if pixelWidth % constants.TileSize != 0 {
		return errors.New(fmt.Sprint(
			"Image", 
			path, 
			"is weirdly dimensioned, as its width of", 
			pixelWidth, 
			"is not a factor of", 
			constants.TileSize,
		))
	}

	if pixelHeight % constants.TileSize != 0 {
		return errors.New(fmt.Sprint(
			"Image", 
			path, 
			"is weirdly dimensioned, as its height of", 
			pixelHeight, 
			"is not a factor of", 
			constants.TileSize,
		))
	}

	width := pixelWidth / constants.TileSize
	height := pixelHeight / constants.TileSize

	const baseDepth = 1

	depths := make([]int, width * height)

	for i := range depths {
		depths[i] = baseDepth
	}

	for i := 0; i < width; i++ {
		depths[i] = baseDepth + 1
	}

	fileName := getFilename(path)
	extension := getExtension(path)

	texture := fileName + "." + extension

	edobj := editor.EditorObject{
		Texture: texture,
		X: 0,
		Y: 0,
		W: width,
		H: height,
		Z: depths,
	}

	bytes, err := json.MarshalIndent(&edobj, "", "\t")
	if err != nil {
		return err
	}

	outputFile := outputDir + "/" + fileName + ".edobj"
	return ioutil.WriteFile(outputFile, bytes, 0644)
}

func getFilename(path string) string {
	lastFileSep := 0
	lastDot := len(path) - 1

	// find file separator
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			lastFileSep = i + 1
			break
		}
	}

	// find last dot
	for ; lastDot > lastFileSep; lastDot-- {
		if path[lastDot] == '.' {
			break
		}
	}

	// no dot found
	if lastDot == lastFileSep {
		// leave as it is
		lastDot = len(path)
	}

	return path[lastFileSep:lastDot]
}

func getExtension(path string) string {
	lastDot := len(path) - 1
	for ; lastDot >= 0; lastDot-- {
		if path[lastDot] == '.' {
			lastDot += 1
			break
		}
	}

	return path[lastDot:]
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

func init() {
	flag.StringVar(&outputDir, "output-dir", ".", "Set output directory")
	flag.Parse()

	filesToProcess = flag.Args()
}

func main() {
	if len(filesToProcess) == 0 {
		fmt.Println("Error: No files to process")
		return
	}

	for _, file := range filesToProcess {
		err := processFile(file)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}
