package pok

func CreateRainWeather(r *Renderer) *RainWeather {
	var _ error
	rain := &RainWeather{}
	return rain
}

type RainWeather struct {
	renderer *Renderer
}
