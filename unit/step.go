package unit

import (
	"github.com/brettbuddin/shaden/dsp"
	"github.com/brettbuddin/shaden/errors"
)

func newStep(io *IO, c Config) (*Unit, error) {
	return NewUnit(io, &step{
		in:      io.NewIn("in", dsp.Float64(1)),
		advance: io.NewIn("advance", dsp.Float64(-1)),
		reset:   io.NewIn("reset", dsp.Float64(-1)),
		serie:   io.NewProp("serie", "", stringPropSetter),
		out:     io.NewOut("out"),
	}), nil
}

type step struct {
	in, advance, reset     *In
	serie                  *Prop
	target                 int
	out                    *Out
	lastAdvance, lastReset float64
}

const stepOn = '1'

func (s *step) ProcessSample(i int) {
	var (
		serie   = s.serie.Value().(string)
		advance = s.advance.Read(i)
		reset   = s.reset.Read(i)
	)

	if isTrig(s.lastAdvance, advance) {
		s.target++
	}
	s.lastAdvance = advance
	if isTrig(s.lastReset, reset) {
		s.target = 0
	}
	s.lastReset = reset

	out := -1.0
	if len(serie) > 0 {
		// When patched, len(serie) can be changed so always check bounds.
		s.target %= len(serie)
		if serie[s.target] == ' ' {
			// Ignore a single space.
			s.target = (s.target + 1) % len(serie)
		} else if serie[s.target] == stepOn {
			out = s.in.Read(i)
		}
	}
	s.out.Write(i, out)
}

func stringPropSetter(p *Prop, v interface{}) error {
	if _, ok := v.(string); !ok {
		return errors.Errorf("property %q is not a string", p.name)
	}
	p.value = v
	return nil
}
