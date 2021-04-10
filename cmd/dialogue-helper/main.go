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

func outputDummy(str string) {
	postBranchB := &pok.DialogData{
		"Donezo!",
		nil,
	}

	postBranchA := &pok.MultipleChoiceDialogData{
		"This is a multiple choice dialog",
		[]string{ "Red", "Blue", "Green" },
		[]pok.DialogNode{
			&pok.DialogData{
				"You chose Red!",
				postBranchB,
			},
			&pok.DialogData{ "You chose Blue!", postBranchB },
			&pok.DialogData{ "You chose Green!", postBranchB },
		},
	}

	root := &pok.DialogData{
		"This is a basic dialog.",
		&pok.BinaryDialogData{
			"This is a yes/no dialog",
			&pok.DialogData{
				"You chose yes!",
				postBranchA,
			},
			&pok.DialogData{
				"You chose no!",
				postBranchB,
			},
		},
	}

	bytes, err := json.MarshalIndent(root, "", "\t")
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

	var node pok.DialogNode = &pok.DialogData{}
	fmt.Println(node)
	err = json.Unmarshal(bytes, node)
	if err != nil {
		panic(err)
	}

	fmt.Println(node)
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
