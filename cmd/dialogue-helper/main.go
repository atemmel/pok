package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/atemmel/pok/pkg/pok"
	"io/ioutil"
)

var files []string

func init() {
	flag.Parse()
	files = flag.Args()
}

func ptr(value int) *int {
	return &value
}

func outputDummy(str string) {

	tree := pok.DialogTree{
		&pok.BinaryDialogNode{
			Dialog: "This is a yes/no dialog",
			True: ptr(1),
			False: ptr(2),
		},
		&pok.DialogNode{
			Dialog:"You chose yes!",
			Next: ptr(4),
		},
		&pok.DialogNode{
			Dialog: "You chose no!",
			Next: ptr(3),
		},
		&pok.DialogNode{
			Dialog: "Donezo!",
			Next: nil,
		},
		&pok.ChoiceDialogNode{
			Dialog: "This is a multiple choice dialog",
			Choices: []string{ "Red", "Blue", "Green" },
			Results: []*int{ ptr(5), ptr(6), ptr(7) },
		},
		&pok.DialogNode{
			Dialog: "You chose Red!",
			Next: ptr(3),
		},
		&pok.DialogNode{
			Dialog: "You chose Blue!",
			Next: ptr(3),
		},
		&pok.DialogNode{
			Dialog: "You chose Green!",
			Next: ptr(3),
		},
	}

	bytes, err := json.MarshalIndent(tree, "", "\t")
	if err != nil {
		panic(err)
	}

	ioutil.WriteFile(str, bytes, 0644)
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

func main() {
	if files != nil && len(files) > 0 {
		for _, s := range files {
			validateFile(s)
		}
	} else {
		outputDummy("test.txt")
	}
}
