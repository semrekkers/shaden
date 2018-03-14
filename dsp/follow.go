package dsp

import "math"

// Follow is an envelope follower
type Follow struct {
	env float64
}

// Tick advances the operation
func (f *Follow) Tick(in, rise, fall float64) float64 {
	in = math.Abs(in)
	if in == f.env {
		return f.env
	}
	if in > f.env {
		fall = rise
	}
	f.env = math.Pow(0.01, 1.0/fall)*(f.env-in) + in
	return f.env
}
