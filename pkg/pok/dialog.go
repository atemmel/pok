package pok

import(
	"encoding/json"
	"io/ioutil"
)

type DialogNode interface {
	Child() DialogNode
}

type BinaryDialogData struct {
	Dialog string
	//True, False int
	True, False DialogNode
}

type MultipleChoiceDialogData struct {
	Dialog string
	Choices []string
	Results []DialogNode
}

type DialogBranchData struct {
	Value1, Value2 string
	Operation string
	//True, False int
	True, False DialogNode
}

type DialogEffectData struct {
	Set, To string
	Next DialogNode
}

//TODO: Improve this, dialog tree maybe?
// https://i.stack.imgur.com/fkNWM.png
type DialogData struct {
	Dialog string
	Next DialogNode
}

func (d *DialogData) Child() DialogNode {
	return nil
}
func (d *BinaryDialogData) Child() DialogNode {
	return nil
}
func (d *MultipleChoiceDialogData) Child() DialogNode {
	return nil
}
func (d *DialogBranchData) Child() DialogNode {
	return nil
}
func (d *DialogEffectData) Child() DialogNode {
	return nil
}
type DialogTree struct {
	Root DialogNode
}

type Dialog struct {
	Dialog string
}

func ReadDialogTreeFromFile(path string) DialogTree {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	var tree DialogTree
	err = json.Unmarshal(data, &tree)
	if err != nil {
		panic(err)
	}

	return tree
}

func ReadDialogFromFile(path string) Dialog {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return Dialog{
			"Could not load: '" + path + "'",
		}
	}

	var dialog Dialog
	err = json.Unmarshal(data, &dialog)
	if err != nil {
		return Dialog{
			"Could not parse: '" + path + "'",
		}
	}

	return dialog
}
