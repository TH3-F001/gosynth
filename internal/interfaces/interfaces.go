package interfaces

type AudioSource interface {
	Sample(deltaT float64) float64
}
