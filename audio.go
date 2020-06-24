package main

import (
	"github.com/hajimehoshi/ebiten/audio"
	"github.com/hajimehoshi/ebiten/audio/mp3"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

type Audio struct {
	audioContext *audio.Context
	audioPlayer *audio.Player
}

func NewAudio() Audio {
	ctx, err := audio.NewContext(44100)
	if err != nil {
		panic(err)
	}
	stream, err := ebitenutil.OpenFile("resources/audio/apple.mp3")
	if err != nil {
		panic(err)
	}
	mp3, err := mp3.Decode(ctx, stream)
	if err != nil {
		panic(err)
	}
	player, err := audio.NewPlayer(ctx, mp3)
	return Audio{
		ctx,
		player,
	}
}
