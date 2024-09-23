package adsr

// ADSR ... A data structure to store Attack, Decay, Sustain, and Release settings for a sound. expects normalized float values
type ADSR struct {
	Attack  float64
	Decay   float64
	Sustain float64
	Release float64
}