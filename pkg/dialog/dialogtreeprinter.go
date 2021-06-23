package dialog

import(
	"fmt"
)

type DialogTreePrinter struct {
	depth uint
	tree *DialogTree
}

func (p *DialogTreePrinter) Print(tree *DialogTree) {
	p.tree = tree
	p.depth = 0

	if len(*p.tree) < 1 {
		return
	}

	p.visit(0)
}

func (p *DialogTreePrinter) pad() {
	for i := uint(0); i < p.depth; i++ {
		fmt.Print("  ")
	}
}

func (p *DialogTreePrinter) visit(index int) {
	(*p.tree)[index].Visit(p)
}

func (p *DialogTreePrinter) end() {
	p.pad()
	fmt.Println(nil)
}

func (p *DialogTreePrinter) VisitDialog(d *DialogNode) {
	p.pad()
	fmt.Println(d.Dialog)
	if d.Next != nil {
		p.depth++
		p.visit(*d.Next)
		p.depth--
	}
}

func (p *DialogTreePrinter) VisitBinary(b *BinaryDialogNode) {
	p.pad()
	fmt.Println(b.Dialog)
	p.depth++
	if b.True != nil {
		p.visit(*b.True)
	} else {
		p.end()
	}
	if b.False != nil {
		p.visit(*b.False)
	} else {
		p.end()
	}
	p.depth--
}

func (p *DialogTreePrinter) VisitChoice(c *ChoiceDialogNode) {
	p.pad()
	fmt.Println(c.Dialog, ":", c.Choices)
	p.depth++
	for _, ptr := range c.Results {
		if ptr != nil {
			p.visit(*ptr)
		} else {
			p.end()
		}
	}
	p.depth--
}


func (p *DialogTreePrinter) VisitEffect(e *EffectDialogNode) {
	p.pad()
	fmt.Println("Effect:", e.Effect)
	if e.Next != nil {
		p.depth++
		p.visit(*e.Next)
		p.depth--
	}
}
