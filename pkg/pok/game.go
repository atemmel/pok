package pok

import (
	"github.com/atemmel/pok/pkg/constants"
	"github.com/atemmel/pok/pkg/debug"
	"github.com/atemmel/pok/pkg/textures"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"math"
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
	playerImg, err = textures.LoadWithError(constants.CharacterImagesDir + "trchar000.png")
	debug.Assert(err)
	playerRunningImg, err = textures.LoadWithError(constants.CharacterImagesDir + "boy_run.png")
	debug.Assert(err)
	beachSplashImg, err = textures.LoadWithError(constants.ImagesDir + "water_effect.png")
	debug.Assert(err)
	playerSurfingImg, err = textures.LoadWithError(constants.CharacterImagesDir + "boy_surf.png")
	debug.Assert(err)
	sharpedoImg, err = textures.LoadWithError(constants.ImagesDir + "surf_sharpedo.png")
	debug.Assert(err)
	sharpedoImg, err = textures.LoadWithError(constants.ImagesDir + "surf_sharpedo.png")
	debug.Assert(err)
	playerUsingHMImg, err = textures.LoadWithError(constants.ImagesDir + "hm_anim.png")
	debug.Assert(err)

	activePlayerImg = playerImg
	g.Dialog = NewDialogBox()
	drawUi = false

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

func (g *Game) Update() error {
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
	debug.Assert(err)
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
	g.Player.Char.Gx = float64(g.Player.Char.X * constants.TileSize)
	g.Player.Char.Gy = float64(g.Player.Char.Y * constants.TileSize)
	g.Rend = NewRenderer(
		constants.DisplaySizeX,
		constants.DisplaySizeY,
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
		player.Char.Tx + (constants.TileSize * 2),
		player.Char.Ty + (constants.TileSize * 2),
	)


	waterBobOffsetY := 0.0
	if player.Char.isSurfing {
		scale := float64(step) / float64(nWaterFrames)
		waterBobOffsetY = math.Sin(scale * math.Pi) * 4
	}

	g.Rend.Draw(&RenderTarget{
		playerOpt,
		activePlayerImg,
		&playerRect,
		x,
		y + waterBobOffsetY,
		2,
	})

	nx, ny := player.Char.X, player.Char.Y
	n := ny * g.Ows.tileMap.Width + nx
	index := g.Ows.tileMap.textureMapping[g.Ows.tileMap.TextureIndicies[player.Char.Z][n]]

	// splash effect
	if textures.IsWater(index) && !player.Char.isSurfing && !player.Char.isJumping {
		splashOpt := &ebiten.DrawImageOptions{}
		w, h := beachSplashImg.Size()
		sx := w / nWaterSplashFrames

		splashRect := image.Rect(
			sx * (step % nWaterSplashFrames),
			0,
			sx * (step % nWaterSplashFrames) + sx,
			h,
		)

		g.Rend.Draw(&RenderTarget{
			splashOpt,
			beachSplashImg,
			&splashRect,
			x + waterSplashOffsetX,
			y + waterSplashOffsetY,
			2 + 1,
		})
	}

	// surfing mount
	if player.Char.isSurfing {
		w, h := sharpedoImg.Size()

		animWidth := w / 2
		animHeight := h / 4

		stepW := player.Char.Tx / (constants.TileSize * 4)
		stepH := player.Char.Ty / (constants.TileSize * 2)

		sharpedoRect := image.Rect(
			animWidth * stepW,
			stepH * animHeight,
			animWidth * stepW + animWidth,
			stepH * animHeight + animHeight,
		)

		sharpedoOpt := &ebiten.DrawImageOptions{}

		g.Rend.Draw(&RenderTarget{
			sharpedoOpt,
			sharpedoImg,
			&sharpedoRect,
			x,
			y + waterBobOffsetY,
			1,
		})
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return constants.DisplaySizeX, constants.DisplaySizeY
}

func (g *Game) PlayAudio() {
	g.Audio.audioPlayer.Play()
}
