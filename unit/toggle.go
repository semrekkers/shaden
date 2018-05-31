package unit

import (
	"github.com/brettbuddin/shaden/dsp"
)

func newToggle(io *IO, _ Config) (*Unit, error) {
	return NewUnit(io, &toggle{
		trigger: io.NewIn("trigger", dsp.Float64(-1)),
		out:     io.NewOut("out"),
	}), nil
}

type toggle struct {
	trigger         *In
	out             *Out
	value, lastTrig float64
}

func (t *toggle) ProcessSample(i int) {
	trig := t.trigger.Read(i)
	if isTrig(t.lastTrig, trig) {
		if t.value > 0 {
			t.value = -1
		} else {
			t.value = 1
		}
	}
	t.lastTrig = trig
	t.out.Write(i, t.value)
}
