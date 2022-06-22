package main

import(
	"image"
	"github.com/atemmel/pok/pkg/editor"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/sqweek/dialog"
)

const (
	ListItemHeight = 18
)

type List struct {
	items []listItem
	first, last int
	cachedImg *ebiten.Image
	x, y float64
	h int
	selectedItemIndex int
	OnRemove func(id int)
}

type ListItem struct {
	Id int
	Name string
}

type listItem struct {
	item ListItem
	rect image.Rectangle
}

func NewList(x, y, n int) List {
	return List{
		items: make([]listItem, 0),
		first: 0,
		last: 0,
		cachedImg: nil,
		x: float64(x),
		y: float64(y),
		h: n,
		selectedItemIndex: 0,
		OnRemove: nil,
	}
}

func (l *List) Append(item ListItem) {
	i := listItem{
		item: item,
		rect: image.Rectangle{},
	}

	l.items = append(l.items, i)
	if len(l.items) >= l.last {
		l.last = len(l.items)
		if l.last - l.first >= l.h {
			l.first = len(l.items) - l.h
		}
	}
	l.selectedItemIndex = len(l.items) - 1
	l.update()
}

func (l *List) MaybeDelete(id int) {
	index := -1
	for i := range l.items {
		if l.items[i].item.Id == id {
			index = i
			break
		}
	}

	if index == -1 {
		return
	}

	if l.OnRemove != nil {
		l.OnRemove(id)
	}

	l.items = append(l.items[:index], l.items[index+1:]...)
	if l.selectedItemIndex >= len(l.items) {
		l.selectedItemIndex--
	}
	l.last--
	if l.first > 0 {
		l.first--
	}
	l.update()
}

func (l *List) HasSelectedId() bool {
	return len(l.items) > 0
}

func (l *List) GetSelectedId() int {
	return l.items[l.selectedItemIndex].item.Id
}

func (l *List) Clear() {
	l.cachedImg = nil
	l.items = l.items[:0]
}

func (l *List) PollInputs() bool {
	cx, cy := ebiten.CursorPosition()
	c := image.Pt(cx, cy)
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		for i := l.first; i < l.last; i++ {
			delta := image.Pt(int(l.x), int(l.y))
			r := l.items[i].rect.Add(delta)
			if c.In(r) {
				l.selectedItemIndex = i
				return true
			}
		}
	} else if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		for i := l.first; i < l.last; i++ {
			delta := image.Pt(int(l.x), int(l.y))
			r := l.items[i].rect.Add(delta)
			if c.In(r) {
				name := l.items[i].item.Name
				id := l.items[i].item.Id
				doRemove := dialog.Message("Remove %v?", name).YesNo()
				if doRemove {
					l.MaybeDelete(id)
				}
				return true;
			}
		}
	}

	if l.cachedImg == nil {
		return false
	}

	_, dy := ebiten.Wheel()
	r := l.cachedImg.Bounds().Add(image.Pt(
		int(l.x), int(l.y),
	))
	if c.In(r) {
		if dy > 0 {
			shown := l.last - l.first - 1
			if l.first > 0 && shown < l.h {
				l.first--
				l.last--
				l.update()
			}
			return true
		} else if dy < 0 {
			shown := l.last - l.first - 1
			if l.last < len(l.items) && shown < l.h {
				l.first++
				l.last++
				l.update()
			}
			return true
		}
	}

	return false
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
		length := len(l.items[i].item.Name)
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

	wBounds := text.BoundString(buttonFont, l.items[maxLenIndex].item.Name)
	w := wBounds.Dx() + (paddingX * 2)
	h := (l.last - l.first) * ListItemHeight

	// build image in regular ram first
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	eimg := ebiten.NewImageFromImage(img)

	for i := l.first; i < l.last; i++ {
		yB := float64(i - l.first) * ListItemHeight
		base := editor.CreateNeatImageWithBorder(w, ListItemHeight)
		src := ebiten.NewImageFromImage(base)

		name := l.items[i].item.Name

		const extraOffset = 9
		text.Draw(src, name, buttonFont, paddingX + 1, paddingY + extraOffset + 1, editor.FgShadow)
		text.Draw(src, name, buttonFont, paddingX, paddingY + extraOffset, editor.Fg)

		opt := &ebiten.DrawImageOptions{}
		opt.GeoM.Translate(0, yB)
		eimg.DrawImage(src, opt)

		iw, ih := src.Size()
		l.items[i].rect = image.Rect(
			0, int(yB), iw, int(yB) + ih,
		)
	}

	l.cachedImg = eimg
}
