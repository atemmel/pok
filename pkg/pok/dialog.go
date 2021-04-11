package pok

import(
	"encoding/json"
	"errors"
	"io/ioutil"
)

type DialogTreeVisitor interface {
	VisitDialog(*DialogNode)
	VisitBinary(*BinaryDialogNode)
	VisitChoice(*ChoiceDialogNode)
}

type DialogNodeInterface interface {
	Visit(DialogTreeVisitor)
}

type BinaryDialogNode struct {
	Dialog string
	True, False *int
}

type ChoiceDialogNode struct {
	Dialog string
	Choices []string
	Results []*int
}

type DialogBranchNode struct {
	Value1, Value2 string
	Operation string
	True, False *int
}

type DialogEffectNode struct {
	Set, To string
	Next *int
}

type DialogNode struct {
	Dialog string
	Next *int
}

func (d *DialogNode) Visit(visitor DialogTreeVisitor) {
	visitor.VisitDialog(d)
}

func (b *BinaryDialogNode) Visit(visitor DialogTreeVisitor) {
	visitor.VisitBinary(b)
}

func (c *ChoiceDialogNode) Visit(visitor DialogTreeVisitor) {
	visitor.VisitChoice(c)
}

type DialogTree []DialogNodeInterface

type DialogTreeNodeData struct {
	Type string
	Data json.RawMessage
}

type DialogNodeData DialogNode

type BinaryDialogNodeData BinaryDialogNode

type ChoiceDialogNodeData ChoiceDialogNode

type DialogTreeData []DialogTreeNodeData

func (d *DialogNode) MarshalJSON() ([]byte, error) {
	diag := DialogNodeData{
		d.Dialog,
		d.Next,
	}

	bytes, err := json.Marshal(diag)
	if err != nil {
		return nil, err
	}

	data := DialogTreeNodeData{
		"Dialog",
		bytes,
	}

	return json.Marshal(data)
}

func (b *BinaryDialogNode) MarshalJSON() ([]byte, error) {
	diag := BinaryDialogNodeData{
		b.Dialog,
		b.True,
		b.False,
	}

	bytes, err := json.Marshal(diag)
	if err != nil {
		return nil, err
	}

	data := DialogTreeNodeData{
		"Binary",
		bytes,
	}

	return json.Marshal(data)
}

func (m *ChoiceDialogNode) MarshalJSON() ([]byte, error) {
	diag := ChoiceDialogNodeData{
		m.Dialog,
		m.Choices,
		m.Results,
	}

	bytes, err := json.Marshal(diag)
	if err != nil {
		return nil, err
	}

	data := DialogTreeNodeData{
		"Choice",
		bytes,
	}

	return json.Marshal(data)
}

func (dt *DialogTree) UnmarshalJSON(bytes []byte) error {
	intermediate := make(DialogTreeData, 0)
	err := json.Unmarshal(bytes, &intermediate)
	if err != nil {
		return err
	}

	*dt = make(DialogTree, len(intermediate))

	for i, s := range intermediate {
		var node DialogNodeInterface
		switch s.Type {
		case "Dialog":
			dialog := &DialogNode{}
			err = json.Unmarshal(s.Data, dialog)
			node = dialog
		case "Binary":
			binary := &BinaryDialogNode{}
			err = json.Unmarshal(s.Data, binary)
			node = binary
		case "Choice":
			choice := &ChoiceDialogNode{}
			err = json.Unmarshal(s.Data, choice)
			node = choice
		default:
			return errors.New("Unrecognized node type, " + s.Type)
		}
		if err != nil {
			return err
		}
		(*dt)[i] = node
	}

	return nil
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
