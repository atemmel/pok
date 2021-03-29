package pok

import(
	"github.com/hajimehoshi/ebiten"
	"image"
)

type NpcGrid struct {
	grid Grid
}

func NewNpcGrid(npcImages []*ebiten.Image) NpcGrid {
	w := 32
	h := 32

	x := 0
	y := 0

	tx := 0
	ty := 2

	rect := image.Rect(tx, ty, tx + w, ty + h)
	n := maxGridWidth / w

	img, _ := ebiten.NewImage(maxGridWidth, len(npcImages) / n * h, ebiten.FilterDefault)

	for _, i := range npcImages {

		opt := &ebiten.DrawImageOptions{}

		opt.GeoM.Translate(float64(x * w), float64(y * h))
		img.DrawImage(i.SubImage(rect).(*ebiten.Image), opt)

		x++
		if x >= n {
			x = 0
			y++
		}
	}

	grid := NewGrid(img, w)
	return NpcGrid{
		grid,
	}
}

func (npcg *NpcGrid) Draw(target *ebiten.Image) {
	npcg.grid.Draw(target)
}

func (npcg *NpcGrid) Scroll(dir ScrollDirection) {
	npcg.grid.Scroll(dir)
}

func (npcg *NpcGrid) Select(cx, cy int) {
	npcg.grid.Select(cx, cy)
}

func (npcg *NpcGrid) Contains(p image.Point) bool {
	return npcg.grid.Contains(p)
}

func (npcg *NpcGrid) GetIndex() int {
	return npcg.grid.GetIndex()
}
