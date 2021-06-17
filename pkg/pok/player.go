package pok

import (
	"github.com/atemmel/pok/pkg/constants"
	"github.com/hajimehoshi/ebiten"
)

type Player struct {
	Id int
	Char Character
	Connected bool
	Location string
}

func (player *Player) Update(g *Game) {
	stepDone := player.Char.Update(g)

	if player.Char.isWalking && player.Char.velocity > WalkVelocity {
		activePlayerImg = playerRunningImg
	} else {
		activePlayerImg = playerImg
	}

	if stepDone {
		player.Char.isWalking = false
		if i := g.Ows.tileMap.HasExitAt(player.Char.X, player.Char.Y, player.Char.Z); i > -1 {
			if g.Ows.tileMap.Exits[i].Target != "" {
				img, _ := ebiten.NewImage(constants.DisplaySizeX, constants.DisplaySizeY, ebiten.FilterDefault);
				g.As.Draw(g, img)
				g.As = NewTransitionState(img, constants.TileMapDir + g.Ows.tileMap.Exits[i].Target, g.Ows.tileMap.Exits[i].Id)
				g.Audio.PlayDoor()
			}
		}
	}
}
