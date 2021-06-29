package pok

import(
	"time"
	"math"
)

type TimeOfDay int

const(
	Morning TimeOfDay = iota
	Day
	Night
)

type TimeRenderEffect struct {
	R, G, B float64
	At int // hour of the day
}

var availableEffects = []TimeRenderEffect{
	{	// Middle of night
		R: 0.5,
		G: 0.5,
		B: 0.8,
		At: 0,
	},
	{	// Before morning
		R: 0.8,
		G: 0.8,
		B: 0.8,
		At: 4,
	},
	{	// Morning
		R: 1.051,
		G: 1.051,
		B: 1.051,
		At: 6,
	},
	{	// Day
		R: 1,
		G: 1,
		B: 1,
		At: 12,
	},
	{	// Early evening
		R: 0.95,
		G: 0.95,
		B: 0.9,
		At: 17,
	},
	{	// Mid evening
		R: 0.9,
		G: 0.8,
		B: 0.8,
		At: 18,
	},
	{	// Late evening
		R:0.8,
		G: 0.8,
		B: 0.8,
		At: 20,
	},
	{	// Night (again)
		R: 0.5,
		G: 0.5,
		B: 0.8,
		At: 23,
	},
}

func GetTimeOfDay() TimeOfDay {
	now := time.Now()
	hour := now.Hour()

	if 4 <= hour && hour < 10 {
		return Morning
	} else if 10 <= hour && hour < 20 {
		return Day
	}

	return Night
}

func GetActiveEffect() (float64, float64, float64) {
	now := time.Now()
	hour := now.Hour()
	minute := now.Minute()
	second := now.Second()

	ongoingIndex := 0
	incomingIndex := 0

	for i := range availableEffects {
		if hour < availableEffects[i].At {
			incomingIndex = i
			ongoingIndex = i - 1
			break
		}
	}

	if incomingIndex == 0 {
		ongoingIndex = len(availableEffects)-1
	}

	progress := float64(hour) + (float64(minute) / 60) + (float64(second) / 60 / 60)
	target := float64(availableEffects[incomingIndex].At)


	frac := progress / target
	if math.IsInf(frac, 0) {
		frac = 1
	}

	from := availableEffects[ongoingIndex]
	to := availableEffects[incomingIndex]

	r := lerp(from.R, to.R, frac)
	g := lerp(from.G, to.G, frac)
	b := lerp(from.B, to.B, frac)

	return r, g, b
}

func lerp(a, b float64, frac float64) float64 {
	return a + frac * (b - a)
}
