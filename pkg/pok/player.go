package pok

import (
	"github.com/atemmel/pok/pkg/constants"
	"github.com/hajimehoshi/ebiten/v2"
)

type Player struct {
	Id int
	Char Character
	Connected bool
	Location string
}

const hmAnimFramesPerStep = 8
const totalHMAnimSteps = 4

func (player *Player) Update(g *Game) {
	stepDone := player.Char.Update(g)

	if player.Char.isBiking {
		activePlayerImg = playerBikingImg
	} else if player.Char.isSurfing {
		activePlayerImg = playerSurfingImg
	} else if player.Char.isWalking && player.Char.velocity > WalkVelocity {
		activePlayerImg = playerRunningImg
	} else {
		activePlayerImg = playerImg
	}

	if stepDone {
		if selectedHm == Surf {
			selectedHm = None
			player.Char.isSurfing = true
		}

		player.Char.isWalking = false
		if i := g.Ows.tileMap.HasExitAt(player.Char.X, player.Char.Y, player.Char.Z); i > -1 {
			if g.Ows.tileMap.Exits[i].Target != "" {
				img := ebiten.NewImage(constants.DisplaySizeX, constants.DisplaySizeY)
				g.As.Draw(g, img)
				g.As = NewTransitionState(img, constants.TileMapDir + g.Ows.tileMap.Exits[i].Target, g.Ows.tileMap.Exits[i].Id)
				g.Audio.PlayDoor()
			}
		}
	}
}
