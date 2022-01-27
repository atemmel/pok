package pok

const downPourZ = 9001;

type WeatherKind int

const (
	Regular WeatherKind = iota
	Hail
	Rain
)

type Weather interface {
	Update()
	Draw(rend *Renderer)
}
