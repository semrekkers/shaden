package unit

import (
	"os"

	"buddin.us/shaden/dsp"
	"buddin.us/shaden/utils"
)

func newRecord(name string, c Config) (*Unit, error) {
	var config struct {
		FileName string
	}
	if err := c.Decode(&config); err != nil {
		return nil, err
	}
	if config.FileName == "" {
		config.FileName = "out.pcm"
	}

	// TODO: Close the file descriptor when the unit unmounts.
	file, err := os.Create(config.FileName)
	if err != nil {
		return nil, err
	}

	io := NewIO()
	return NewUnit(io, name, &record{
		in:       io.NewIn("in", dsp.Float64(0)),
		record:   io.NewIn("record", dsp.Float64(0)),
		reset:    io.NewIn("reset", dsp.Float64(0)),
		position: io.NewOut("position"),
		out:      io.NewOut("out"),

		file:   file,
		writer: utils.NewPcmWriter(file, dsp.FrameSize),
	}), nil
}

type record struct {
	in, record, reset *In
	file              *os.File
	writer            *utils.PcmWriter
	length            float64
	out, position     *Out
}

func (r *record) ProcessSample(i int) {
	if r.reset.Read(i) > 0 {
		r.file.Truncate(0)
		r.length = 0
	}
	if r.record.Read(i) <= 0 {
		return
	}
	sample := r.in.Read(i)
	r.writer.WriteSample(sample)
	r.out.Write(i, sample)
	r.position.Write(i, r.length)
	r.length += 1000 / dsp.SampleRate
}
