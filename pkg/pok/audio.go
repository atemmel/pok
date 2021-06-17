package pok

import (
	"io/ioutil"
	"github.com/atemmel/pok/pkg/constants"
	"github.com/atemmel/pok/pkg/debug"
	"github.com/hajimehoshi/ebiten/audio"
	"github.com/hajimehoshi/ebiten/audio/mp3"
	"github.com/hajimehoshi/ebiten/audio/vorbis"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

const volume = 0.2

type Audio struct {
	audioContext *audio.Context
	audioPlayer *audio.Player
	thudPlayer *audio.Player
	doorPlayer *audio.Player
	playerJumpPlayer *audio.Player
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

func (a *Audio) PlayPlayerJump() {
	a.playerJumpPlayer.Rewind()
	a.playerJumpPlayer.Play()
}

func NewAudio() Audio {
	ctx, err := audio.NewContext(44100)
	debug.Assert(err)
	msrc, err := loadMp3(ctx, constants.AudioDir + "apple.mp3")
	debug.Assert(err)
	loop := audio.NewInfiniteLoop(msrc, msrc.Length() - 2500000)
	player, err := audio.NewPlayer(ctx, loop)
	debug.Assert(err)
	msrc, err = loadMp3(ctx, constants.AudioDir + "thud.mp3")
	debug.Assert(err)
	thud, err := audio.NewPlayer(ctx, msrc)
	debug.Assert(err)
	msrc, err = loadMp3(ctx, constants.AudioDir + "door.mp3")
	debug.Assert(err)
	door, err := audio.NewPlayer(ctx, msrc)
	debug.Assert(err)
	osrc, err := loadOgg(ctx, constants.AudioDir + "player_jump.ogg")
	debug.Assert(err)
	jump, err := audio.NewPlayer(ctx, osrc)

	player.SetVolume(volume)
	thud.SetVolume(volume)
	door.SetVolume(volume)
	jump.SetVolume(volume)

	return Audio{
		ctx,
		player,
		thud,
		door,
		jump,
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

func loadOgg(ctx *audio.Context, str string) (*vorbis.Stream, error) {
	stream, err := ebitenutil.OpenFile(str)
	if err != nil {
		return nil, err
	}
	src, err := vorbis.Decode(ctx, stream)
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
