package main

import (
	"errors"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/atemmel/pok/pkg/pok"
	"io/ioutil"
	"strconv"
	"strings"
)

var files []string
var doValidate bool

func init() {
	flag.BoolVar(&doValidate, "validate", false, "Validate files to be correct JSON")
	flag.Parse()
	files = flag.Args()
}

func ptr(value int) *int {
	return &value
}

func validateFile(str string) {
	bytes, err := ioutil.ReadFile(str)
	if err != nil {
		panic(err)
	}

	tree := &pok.DialogTree{}
	err = json.Unmarshal(bytes, tree)
	if err != nil {
		panic(err)
	}

	fmt.Println(tree)

	printer := pok.DialogTreePrinter{}
	printer.Print(tree)
}

func transpileToDialogJson(path string) pok.DialogTree {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	str := string(bytes)
	err = validateLines(str)
	if err != nil {
		panic(err)
	}

	tree := make(pok.DialogTree, 0)
	checkpoint := 0
	outerCheckpoint := 0
	nodeNo := 0
	foundNewline := false

	mknode := func(i int) *pok.DialogNode {
		nodeNo++
		return &pok.DialogNode{
			Dialog: str[checkpoint:i],
			Next: ptr(nodeNo),
		}
	}

	for i := range str {
		if str[i] == '\n' {
			if foundNewline {
				foundNewline = false
				tree = append(tree, mknode(i))
				outerCheckpoint = checkpoint
				checkpoint = i + 1
			} else {
				foundNewline = true
			}
		}
	}

	if foundNewline { // uneven line count
		tree = append(tree,
			&pok.DialogNode{
				Dialog: str[checkpoint:],
				Next: nil,
			})
	} else {	// even line count
		tree[len(tree)-1] = &pok.DialogNode{
			Dialog: str[outerCheckpoint:],
			Next: nil,
		}
	}

	return tree
}

func genFilename(original string, extension string) string {
	other := original
	if i := strings.Index(original, "."); i != -1 {
		other = other[:i] + ".dialog"
	} else {
		other += ".dialog"
	}
	return other
}

func validateLines(lines string) error {
	counter := 0
	lineNo := 1
	for i := range lines {
		if counter > pok.MaxLetters {
			return errors.New("Too long line encountered on line " + strconv.Itoa(lineNo))
		}

		counter++
		if lines[i] == '\n' {
			lineNo++
			counter = 0
		}
	}

	return nil
}

func main() {
	if files != nil && len(files) > 0 {
		if doValidate {
			for _, s := range files {
				validateFile(s)
			}
		} else {
			for _, s := range files {
				tree := transpileToDialogJson(s)
				bytes, err := json.Marshal(&tree)
				if err != nil {
					panic(err)
				}
				filename := genFilename(s, ".dialog")
				ioutil.WriteFile(filename, bytes, 0644)
			}
		}
	}
}
