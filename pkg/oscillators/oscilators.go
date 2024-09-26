package oscillators

import (
	"math"
)

//#region SineWave

// Amplitude expects a decimal value between 0 and 1
type SineWave struct {
	Frequency float64
	Amplitude float64
}

func NewSineWave(frequency, amplitude float64) *SineWave {
	return &SineWave{
		Frequency: frequency,
		Amplitude: amplitude,
	}
}

func (sine *SineWave) Sample(deltaT float64) float64 {
	return sine.Amplitude * math.Sin(sine.Frequency*2*math.Pi*deltaT)
}

//#endregion

// #region SquareWave
type SquareWave struct {
	Frequency float64
	Amplitude float64
}

func NewSquareWave(frequency, amplitude float64) *SquareWave {
	return &SquareWave{
		Frequency: frequency,
		Amplitude: amplitude,
	}
}

func (sq *SquareWave) Sample(deltaT float64) float64 {
	sin := sq.Amplitude * math.Sin(sq.Frequency*2*math.Pi*deltaT)
	if sin > 0.0 {
		return sq.Amplitude
	}
	return -sq.Amplitude
}

//#endregion
