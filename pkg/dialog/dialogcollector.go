package dialog

type DialogTreeCollector struct {
	tree *DialogTree
	current *int
	nextResult *DialogTreeCollectorResult
}

type DialogTreeCollectorResult struct {
	Dialog string
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

func (coll *DialogTreeCollector) VisitDialog(d *DialogNode) {
	coll.nextResult.Dialog = d.Dialog
	coll.current = d.Next
}

func (coll *DialogTreeCollector) VisitBinary(b *BinaryDialogNode) {
}

func (coll *DialogTreeCollector) VisitChoice(c *ChoiceDialogNode) {
}
