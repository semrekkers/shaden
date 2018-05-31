package unit

import (
	"fmt"

	"github.com/brettbuddin/shaden/dsp"
)

func newMux(io *IO, c Config) (*Unit, error) {
	var config struct {
		Size int
	}
	if err := c.Decode(&config); err != nil {
		return nil, err
	}

	if config.Size == 0 {
		config.Size = 2
	}

	inputs := make([]*In, config.Size)
	for i := range inputs {
		inputs[i] = io.NewIn(fmt.Sprintf("%d", i), dsp.Float64(0))
	}

	return NewUnit(io, &mux{
		selection: io.NewIn("select", dsp.Float64(1)),
		out:       io.NewOut("out"),
		inputs:    inputs,
	}), nil
}

type mux struct {
	inputs    []*In
	selection *In
	out       *Out
}

func (m *mux) ProcessSample(i int) {
	max := float64(len(m.inputs) - 1)
	s := int(dsp.Clamp(m.selection.Read(i), 0, max))
	m.out.Write(i, m.inputs[s].Read(i))
}
