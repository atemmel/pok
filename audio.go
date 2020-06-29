package main

import (
	"io/ioutil"
	"github.com/hajimehoshi/ebiten/audio"
	"github.com/hajimehoshi/ebiten/audio/mp3"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

type Audio struct {
	audioContext *audio.Context
	audioPlayer *audio.Player
	thudPlayer *audio.Player
	doorPlayer *audio.Player
}

func (a *Audio) PlayThud() {
	if a.thudPlayer.IsPlaying() {
		return
	}
	a.thudPlayer.Rewind()
	a.thudPlayer.Play()
}

func (a *Audio) PlayDoor() {
	//if a.doorPlayer.IsPlaying() {
	//}
	a.doorPlayer.Rewind()
	a.doorPlayer.Play()
}

func NewAudio() Audio {
	ctx, err := audio.NewContext(44100)
	if err != nil {
		panic(err)
	}
	src, err := loadMp3(ctx, "resources/audio/apple.mp3")
	if err != nil {
		panic(err)
	}
	loop := audio.NewInfiniteLoop(src, src.Length() - 2500000)
	player, err := audio.NewPlayer(ctx, loop)
	if err != nil {
		panic(err)
	}
	src, err = loadMp3(ctx, "resources/audio/thud.mp3")
	if err != nil {
		panic(err)
	}
	thud, err := audio.NewPlayer(ctx, src)
	if err != nil {
		panic(err)
	}
	src, err = loadMp3(ctx, "resources/audio/door.mp3")
	if err != nil {
		panic(err)
	}
	door, err := audio.NewPlayer(ctx, src)
	if err != nil {
		panic(err)
	}
	return Audio{
		ctx,
		player,
		thud,
		door,
	}
}

func loadMp3(ctx *audio.Context, str string) (*mp3.Stream, error) {
	stream, err := ebitenutil.OpenFile(str)
	if err != nil {
		return nil, err
	}
	src, err := mp3.Decode(ctx, stream)
	if err != nil {
		return nil, err
	}
	return src, nil
}

func loadMp3AsBytes(ctx *audio.Context, str string) ([]byte, error) {
	stream, err := ebitenutil.OpenFile(str)
	if err != nil {
		return nil, err
	}
	src, err := mp3.Decode(ctx, stream)
	if err != nil {
		return nil, err
	}
	bytes, err := ioutil.ReadAll(src)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
