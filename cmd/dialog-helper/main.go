package main

import (
	"errors"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/atemmel/pok/pkg/dialog"
	"io/ioutil"
	"strconv"
	"strings"
)

var files []string
var doValidate *bool

func init() {
	doValidate = flag.Bool("validate", false, "Validates files to be correct JSON")
}

func ptr(value int) *int {
	return &value
}

func validateFile(str string) {
	bytes, err := ioutil.ReadFile(str)
	if err != nil {
		panic(err)
	}

	tree := &dialog.DialogTree{}
	err = json.Unmarshal(bytes, tree)
	if err != nil {
		panic(err)
	}

	fmt.Println(tree)

	printer := dialog.DialogTreePrinter{}
	printer.Print(tree)
}

func transpileToDialogJson(path string) dialog.DialogTree {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	str := string(bytes)
	err = validateLines(str)
	if err != nil {
		panic(err)
	}

	tree := make(dialog.DialogTree, 0)
	checkpoint := 0
	outerCheckpoint := 0
	nodeNo := 0
	foundNewline := false

	mknode := func(i int) *dialog.DialogNode {
		nodeNo++
		return &dialog.DialogNode{
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
			&dialog.DialogNode{
				Dialog: str[checkpoint:],
				Next: nil,
			})
	} else {	// even line count
		tree[len(tree)-1] = &dialog.DialogNode{
			Dialog: str[outerCheckpoint:],
			Next: nil,
		}
	}

	return tree
}

func genFilename(original string, extension string) string {
	other := original
	if i := strings.Index(original, "."); i != -1 {
		other = other[:i] + extension
	} else {
		other += extension
	}
	return other
}

func validateLines(lines string) error {
	counter := 0
	lineNo := 1
	for i := range lines {
		if counter > dialog.MaxLetters {
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
	flag.Parse()
	files = flag.Args()
	if files != nil && len(files) > 0 {
		if *doValidate {
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
