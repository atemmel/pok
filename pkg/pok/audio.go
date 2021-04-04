package pok

import (
	"io/ioutil"
	"github.com/hajimehoshi/ebiten/audio"
	"github.com/hajimehoshi/ebiten/audio/mp3"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

const volume = 0.2

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
	a.doorPlayer.Rewind()
	a.doorPlayer.Play()
}

func NewAudio() Audio {
	ctx, err := audio.NewContext(44100)
	Assert(err)
	src, err := loadMp3(ctx, AudioDir + "apple.mp3")
	Assert(err)
	loop := audio.NewInfiniteLoop(src, src.Length() - 2500000)
	player, err := audio.NewPlayer(ctx, loop)
	Assert(err)
	src, err = loadMp3(ctx, AudioDir + "thud.mp3")
	Assert(err)
	thud, err := audio.NewPlayer(ctx, src)
	Assert(err)
	src, err = loadMp3(ctx, AudioDir + "door.mp3")
	Assert(err)
	door, err := audio.NewPlayer(ctx, src)
	Assert(err)

	player.SetVolume(volume)
	thud.SetVolume(volume)
	door.SetVolume(volume)

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
