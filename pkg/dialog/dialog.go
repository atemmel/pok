package dialog

import(
	"encoding/json"
	"errors"
	"io/ioutil"
)

const(
	MaxLetters = 44
)

type NodeId int

const(
	DialogNodeId NodeId = iota
	BinaryDialogNodeId
	ChoiceDialogNodeId
	EffectDialogNodeId
)

type DialogTreeVisitor interface {
	VisitDialog(*DialogNode)
	VisitBinary(*BinaryDialogNode)
	VisitChoice(*ChoiceDialogNode)
	VisitEffect(*EffectDialogNode)
}

type DialogNodeInterface interface {
	Visit(DialogTreeVisitor)
	GetNodeId() NodeId
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

type EffectDialogNode struct {
	Effect string
	Next *int
}

type DialogAssignNode struct {
	Set, To string
	Next *int
}

type DialogNode struct {
	Dialog string
	Next *int
}

func Link(index int) *int {
	return &index
}

func (d *DialogNode) Visit(visitor DialogTreeVisitor) {
	visitor.VisitDialog(d)
}

func (d *DialogNode) GetNodeId() NodeId {
	return DialogNodeId
}

func (b *BinaryDialogNode) Visit(visitor DialogTreeVisitor) {
	visitor.VisitBinary(b)
}

func (b *BinaryDialogNode) GetNodeId() NodeId {
	return BinaryDialogNodeId
}

func (c *ChoiceDialogNode) Visit(visitor DialogTreeVisitor) {
	visitor.VisitChoice(c)
}

func (c *ChoiceDialogNode) GetNodeId() NodeId {
	return ChoiceDialogNodeId
}

func (e *EffectDialogNode) Visit(visitor DialogTreeVisitor) {
	visitor.VisitEffect(e)
}

func (e *EffectDialogNode) GetNodeId() NodeId {
	return EffectDialogNodeId
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

func ReadDialogTreeFromFile(path string) (*DialogTree, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	tree := &DialogTree{}
	err = json.Unmarshal(data, tree)
	if err != nil {
		return nil, err
	}

	return tree, nil
}
