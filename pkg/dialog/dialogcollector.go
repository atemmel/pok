package dialog

type DialogTreeCollector struct {
	tree *DialogTree
	current *int
	nextResult *DialogTreeCollectorResult
}

type DialogTreeCollectorResult struct {
	Dialog string
	NodeId NodeId
	Opt string
}

func MakeDialogTreeCollector(tree *DialogTree) DialogTreeCollector {
	if len(*tree) == 0 {
		return DialogTreeCollector{
			tree,
			nil,
			nil,
		}
	}

	return DialogTreeCollector{
		tree,
		new(int),
		nil,
	}
}

func (coll *DialogTreeCollector) CollectOnce() *DialogTreeCollectorResult {
	if coll.current == nil {
		return nil
	}
	coll.nextResult = &DialogTreeCollectorResult{}
	(*coll.tree)[*coll.current].Visit(coll)
	return coll.nextResult
}

func (coll *DialogTreeCollector) Peek() *DialogTreeCollectorResult {
	old := coll.current
	res := coll.CollectOnce()
	coll.current = old
	return res
}

func (coll *DialogTreeCollector) VisitDialog(d *DialogNode) {
	coll.nextResult.Dialog = d.Dialog
	coll.nextResult.NodeId = d.GetNodeId()
	coll.current = d.Next
}

func (coll *DialogTreeCollector) VisitBinary(b *BinaryDialogNode) {
	coll.nextResult.Dialog = b.Dialog
	coll.nextResult.NodeId = b.GetNodeId()
}

func (coll *DialogTreeCollector) VisitChoice(c *ChoiceDialogNode) {
}

func (coll *DialogTreeCollector) VisitEffect(e *EffectDialogNode) {
	coll.nextResult.NodeId = e.GetNodeId()
	coll.nextResult.Opt = e.Effect
	coll.current = e.Next
}
