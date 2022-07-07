package textures

import(
	"encoding/json"
	"errors"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/atemmel/pok/pkg/constants"
	"github.com/atemmel/pok/pkg/debug"
	"github.com/atemmel/pok/pkg/jobs"

	"io/ioutil"
	_ "image/png"
)

type textureMeta struct {
	texture *ebiten.Image
	animated bool
}

var(
	aliases map[string]int
	textures []textureMeta

	baseTextureIndex = InvalidIndex
	waterTextureIndex = InvalidIndex
	stairTextureIndex = InvalidIndex
)

type textureAnimationMeta struct {
	Texture string
	Frames int
	FramesPerStep uint
	TilesToSkip int

	step int
	textureId int
}

type textureAnimations []textureAnimationMeta

const(
	InvalidIndex = -1

	preAlloc = 8
	baseTextureStr = constants.TileMapImagesDir + "base.png"
	waterTextureStr = constants.TileMapImagesDir + "water.png"
	stairTextureStr = constants.TileMapImagesDir + "stairs.png"

	animationManifestStr = constants.TileMapImagesDir + "animation_manifest.json"

	rockTextureStr = constants.PropsImagesDir + "object_rock.png"
	cutTextureStr = constants.PropsImagesDir + "object_cut.png"
	boulderTextureStr = constants.PropsImagesDir + "object_boulder.png"
)

var(
	rockImg *ebiten.Image = nil
	cutImg *ebiten.Image = nil
	boulderImg *ebiten.Image = nil
	animations textureAnimations = nil
)

func Init() {
	aliases = make(map[string]int, preAlloc)
	textures = make([]textureMeta, 0, preAlloc)
	animations = make(textureAnimations, 0)
	var err error

	rockImg, _, err = ebitenutil.NewImageFromFile(rockTextureStr)
	debug.Assert(err)
	cutImg, _, err = ebitenutil.NewImageFromFile(cutTextureStr)
	debug.Assert(err)
	boulderImg, _, err = ebitenutil.NewImageFromFile(boulderTextureStr)
	debug.Assert(err)

	bytes, err := ioutil.ReadFile(animationManifestStr)
	debug.Assert(err)

	err = json.Unmarshal(bytes, &animations)
	debug.Assert(err)
	setupAnimation()
}

/*
func Animate() {
	if animations == nil {
		return
	}

	for i := range animations {
		anim := &animations[i]
		anim.step++
		if anim.step >= anim.Frames {
			anim.step = 0
		}
	}
}
*/

func setupAnimation() {
	if animations == nil {
		return
	}

	animate := func(anim *textureAnimationMeta) {
		anim.step++
		if anim.step >= anim.Frames {
			anim.step = 0
		}
	}

	for i := range animations {
		anim := &animations[i]
		jobs.Add(jobs.Job{
			Do: func() {
				animate(anim)
			},
			When: anim.FramesPerStep,
		})
	}
}

func Load(path string) (*ebiten.Image, int) {
	index, ok := aliases[path]
	if !ok {
		return insertNewTexture(path);
	}
	return Access(index), index
}

func GetRockImage() *ebiten.Image {
	return rockImg
}

func GetCutImage() *ebiten.Image {
	return cutImg
}

func GetBoulderImage() *ebiten.Image {
	return boulderImg
}

func LoadWithError(path string) (*ebiten.Image, error) {
	img, _, err := ebitenutil.NewImageFromFile(path)
	return img, err
}

func Access(index int) *ebiten.Image {
	return textures[index].texture;
}

func IsWater(index int) bool {
	return index == waterTextureIndex
}

func IsBase(index int) bool {
	return index == baseTextureIndex
}

func IsStair(index int) bool {
	return index == stairTextureIndex
}

func IsAnimated(index int) bool {
	return textures[index].animated
}

func GetStepAndSkip(index int) (int, int) {
	for i := range animations {
		if animations[i].textureId == index {
			return animations[i].step, animations[i].TilesToSkip
		}
	}
	debug.Assert(errors.New("Invalid animation index"))
	return 0, 0
}

func GetStepScale(index int) float64 {
	for i := range animations {
		anim := &animations[i]
		if anim.textureId == index {
			s := float64(anim.step) / float64(anim.Frames)
			return s
		}
	}
	debug.Assert(errors.New("Invalid animation index"))
	return 0
}

func GetWaterStepScale() float64 {
	return GetStepScale(waterTextureIndex)
}

func insertNewTexture(path string) (*ebiten.Image, int) {
	img, _, err := ebitenutil.NewImageFromFile(path)
	debug.Assert(err)

	for i, t := range textures {
		if t.texture == nil {
			aliases[path] = i
			textures[i].texture = img
			checkForSpecialTextures(path, i)
			return img, i
		}
	}

	i := len(textures)
	aliases[path] = i
	textures = append(textures, textureMeta{
		texture: img,
		animated: false,
	})
	checkForSpecialTextures(path, i)
	return img, i
}

func checkForSpecialTextures(path string, index int) {
	if baseTextureIndex == InvalidIndex && baseTextureStr == path {
		baseTextureIndex = index
	} else if waterTextureIndex == InvalidIndex && waterTextureStr == path {
		waterTextureIndex = index
	} else if stairTextureIndex == InvalidIndex && stairTextureStr == path {
		stairTextureIndex = index
	}

	for i := range animations {
		anim := &animations[i]
		if constants.TileMapImagesDir + anim.Texture == path {
			anim.textureId = index
			textures[index].animated = true
			break
		}
	}
}
