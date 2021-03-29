package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"strings"
)

var output string
var scale float64
var verbose bool
var filesToProcess []string

func init() {
	flag.StringVar(&output, "output", "", "Specify output file/directory of processed files")
	flag.Float64Var(&scale, "scale", 1.0, "Specify scale to use (0.5 halves the size, 2.0 doubles it, etc)")
	flag.BoolVar(&verbose, "verbose", false, "Print additional messages")
	flag.Parse()
	filesToProcess = flag.Args()
}

func validatePaths(paths []string) {
	for _, s := range paths {
		if _, err := os.Stat(s); os.IsNotExist(err) {
			fmt.Println("Could not find", s)
			panic(err)
		}
	}
}

func processImages(paths []string) {
	for _, p := range paths {
		if verbose {
			fmt.Println("Currently processing:", p)
		}
		processImage(p)
	}
}

func processImage(path string) {
	img, err := openImage(path)
	if err != nil {
		panic(err)
	}

	w1 := img.Bounds().Max.X
	h1 := img.Bounds().Max.Y

	w2 := int(float64(w1) * scale)
	h2 := int(float64(h1) * scale)

	var processedImage image.Image
	if scale < 1 {
		processedImage = downscaleImage(img, w2, h2)
	} else if scale > 1 {
		processedImage = upscaleImage(img, w2, h2)
	} else {
		processedImage = img
	}

	index := strings.LastIndex(path, "/")
	outPath := output + "/"
	if index == -1 {
		outPath += path
	} else {
		outPath += path[index + 1:]
	}

	err = saveImage(processedImage, outPath)
	if err != nil {
		panic(err)
	}
}

func upscaleImage(img image.Image, w, h int) image.Image {
	result := image.NewRGBA(image.Rect(0, 0, w, h))

	w1 := img.Bounds().Max.X
	h1 := img.Bounds().Max.Y

	dx := int(float64(w) / float64(w1))
	dy := int(float64(h) / float64(h1))

	for y := 0; y < h1; y++ {
		uy := y * dy
		for x := 0; x < w1; x++ {
			ux := x * dx
			p := img.At(x, y)
			for iy := 0; iy < dy; iy++ {
				for ix := 0; ix < dx; ix++ {
					result.Set(ux + ix, uy + iy, p)
				}
			}
		}
	}

	return result
}

func downscaleImage(img image.Image, w, h int) image.Image {
	result := image.NewRGBA(image.Rect(0, 0, w, h))

	w1 := img.Bounds().Max.X
	h1 := img.Bounds().Max.Y

	dx := float64(w) / float64(w1)
	dy := float64(h) / float64(h1)

	carryOverX := 0.0
	carryOverY := 0.0
	for y := 0; y < h1; {
		sy := float64(y) / float64(h1)
		oldY := y
		for x := 0; x < w1; {
			sx := float64(x) / float64(w1)
			p := img.At(x, y)

			xDest := int(sx * float64(w))
			yDest := int(sy * float64(h))
			result.Set(xDest, yDest, p)

			oldX := x
			x += int(dx)
			if x == oldX {
				carryOverX += dx
				if carryOverX >= 1 {
					carryOverX--
					x++
				}
			}
		}

		y += int(dy)
		if y == oldY {
			carryOverY += dy
			if carryOverY >= 1 {
				carryOverY--
				y++
			}
		}

	}

	return result
}

func openImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	image, _, err := image.Decode(f)
	return image, err
}

func saveImage(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	err = png.Encode(f, img)
	return err
}

func main() {
	if scale < 0 {
		fmt.Println("Error: scale must not be negative")
		return
	}

	validatePaths(filesToProcess)
	processImages(filesToProcess)
}
