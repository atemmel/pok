package main

import(
	"image"
	"github.com/atemmel/pok/pkg/editor"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
)

const (
	ListItemHeight = 18
)

type List struct {
	items []ListItem
	first, last int
	cachedImg *ebiten.Image
	x, y float64
	h int
}

type ListItem struct {
	Id int
	Name string
}

func NewList(x, y, n int) List {
	return List{
		items: make([]ListItem, 0),
		first: 0,
		last: 0,
		cachedImg: nil,
		x: float64(x),
		y: float64(y),
		h: n,
	}
}

func (l *List) Append(item ListItem) {
	l.items = append(l.items, item)
	if len(l.items) >= l.last {
		l.last = len(l.items)
		if l.last - l.first >= l.h {
			l.first = len(l.items) - l.h
		}
	}
	l.update()
}

func (l *List) MaybeDelete(id int) {
	index := -1
	for i := range l.items {
		if l.items[i].Id == id {
			index = i
			break
		}
	}

	if index == -1 {
		return
	}

	l.items = append(l.items[:index], l.items[index+1:]...)
}

func (l *List) Draw(target *ebiten.Image) {
	if l.cachedImg != nil {
		opt := &ebiten.DrawImageOptions{}
		opt.GeoM.Translate(l.x, l.y)
		target.DrawImage(l.cachedImg, opt);
	}
}

func (l *List) update() {
	maxLenIndex := -1
	maxLen := 0
	for i := range l.items {
		length := len(l.items[i].Name)
		if length > maxLen {
			maxLenIndex = i
			maxLen = length
		}
	}

	if maxLenIndex == -1 {
		return
	}

	const paddingX = 4
	const paddingY = 4

	wBounds := text.BoundString(buttonFont, l.items[maxLenIndex].Name)
	w := wBounds.Dx() + (paddingX * 2)
	h := (l.last - l.first) * ListItemHeight

	// build image in regular ram first
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	eimg := ebiten.NewImageFromImage(img)

	for i := l.first; i < l.last; i++ {
		yB := float64(i - l.first) * ListItemHeight
		base := editor.CreateNeatImageWithBorder(w, ListItemHeight)
		src := ebiten.NewImageFromImage(base)

		name := l.items[i].Name

		const extraOffset = 9
		text.Draw(src, name, buttonFont, paddingX + 1, paddingY + extraOffset + 1, editor.FgShadow)
		text.Draw(src, name, buttonFont, paddingX, paddingY + extraOffset, editor.Fg)

		opt := &ebiten.DrawImageOptions{}
		opt.GeoM.Translate(0, yB)
		eimg.DrawImage(src, opt)
	}

	l.cachedImg = eimg
}
