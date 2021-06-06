package pok

import (
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"image"
)

type Game struct {
	Ows OverworldState
	As GameState
	Player Player
	Client Client
	Rend Renderer
	Audio Audio
	Dialog DialogBox
}

func CreateGame() *Game {
	g := &Game{}
	g.As = &g.Ows
	var err error
	playerImg, _, err = ebitenutil.NewImageFromFile(CharacterImagesDir + "trchar000.png", ebiten.FilterDefault)
	Assert(err)
	playerRunningImg, _, err = ebitenutil.NewImageFromFile(CharacterImagesDir + "boy_run.png", ebiten.FilterDefault)
	Assert(err)
	activePlayerImg = playerImg
	g.Dialog = NewDialogBox()
	drawUi = true

	return g
}

func (g *Game) TileIsOccupied(x int, y int, z int) bool {
	if x < 0 || x >= g.Ows.tileMap.Width || y < 0 ||  y >= g.Ows.tileMap.Height {
		return true
	}

	index := y * g.Ows.tileMap.Width + x

	// Out of bounds check
	if z < 0 || z >= len(g.Ows.tileMap.Tiles) {
		return true
	}

	if index >= len(g.Ows.tileMap.Tiles[z]) || index < 0 {
		return true
	}

	if g.Ows.tileMap.Collision[z][index] {
		return true
	}

	for _, p := range g.Client.playerMap.players {
		if p.Char.X == x && p.Char.Y == y {
			return true
		}
	}

	for i := range g.Ows.tileMap.npcs {
		c := &g.Ows.tileMap.npcs[i].Char
		if c.X == x && c.Y == y && c.Z == z {
			return true
		}
	}

	if g.Player.Char.X == x && g.Player.Char.Y == y && g.Player.Char.Z == z {
		return true
	}

	return false
}

func (g *Game) Update(screen *ebiten.Image) error {
	err := g.As.GetInputs(g)
	if err != nil {
		return err
	}
	err = g.As.Update(g)
	if err != nil {
		return err
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.As.Draw(g, screen)
}

func (g *Game) Load(str string, entrypoint int) {
	err := g.Ows.tileMap.OpenFile(str)
	Assert(err)
	currentLayer = 0
	selectedTile = 0
	g.Player.Location = str
	index := g.Ows.tileMap.GetEntryWithId(entrypoint)
	if index >= 0 {
		g.Player.Char.X = g.Ows.tileMap.Entries[index].X
		g.Player.Char.Y = g.Ows.tileMap.Entries[index].Y
	} else {
		g.Player.Char.X = 0
		g.Player.Char.Y = 0
	}
	g.Player.Char.Gx = float64(g.Player.Char.X * TileSize)
	g.Player.Char.Gy = float64(g.Player.Char.Y * TileSize)
	g.Rend = NewRenderer(
		DisplaySizeX,
		DisplaySizeY,
		2,
	)
}

func (g *Game) Save() {
	/*
	bytes, err := json.Marshal(g.Ows.tileMap)
	if err != nil {
		fmt.Println(err)
	}
	ioutil.WriteFile(g.Player.Location, bytes, 0644)
	*/
}

//TODO: Maybe throw away?
func (g *Game) DrawPlayer(player *Player) {
	playerOpt := &ebiten.DrawImageOptions{}

	x := player.Char.Gx + NpcOffsetX
	y := player.Char.Gy + NpcOffsetY + player.Char.OffsetY

	playerRect := image.Rect(
		player.Char.Tx,
		player.Char.Ty,
		player.Char.Tx + (TileSize * 2),
		player.Char.Ty + (TileSize * 2),
	)

	g.Rend.Draw(&RenderTarget{
		playerOpt,
		activePlayerImg,
		&playerRect,
		x,
		y,
		2,
	})
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return DisplaySizeX, DisplaySizeY
}

func (g *Game) PlayAudio() {
	g.Audio.audioPlayer.Play()
}
