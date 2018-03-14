package unit

import "buddin.us/shaden/dsp"

func newFollow(io *IO, _ Config) (*Unit, error) {
	return NewUnit(io, &follow{
		in:   io.NewIn("in", dsp.Float64(0)),
		rise: io.NewIn("rise", dsp.Float64(10)),
		fall: io.NewIn("fall", dsp.Float64(200)),
		out:  io.NewOut("out"),
	}), nil
}

type follow struct {
	in, rise, fall *In
	follow         dsp.Follow
	out            *Out
}

func (f *follow) ProcessSample(i int) {
	var (
		in   = f.in.Read(i)
		rise = f.rise.Read(i)
		fall = f.fall.Read(i)
	)
	f.out.Write(i, f.follow.Tick(in, rise, fall))
}
