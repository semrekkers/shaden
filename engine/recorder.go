package engine

import (
	"encoding/binary"
	"io"
	"math"
)

type RecorderBackend struct {
	inner   Backend
	innerCb func([]float32, [][]float32)
	w       io.Writer
}

func NewRecorderBackend(w io.Writer, inner Backend) *RecorderBackend {
	return &RecorderBackend{
		w:     w,
		inner: inner,
	}
}

func (b *RecorderBackend) Start(cb func([]float32, [][]float32)) error {
	b.innerCb = cb
	return b.inner.Start(b.callback)
}

func (b *RecorderBackend) Stop() error {
	return b.inner.Stop()
}

func (b *RecorderBackend) SampleRate() int {
	return b.inner.SampleRate()
}

func (b *RecorderBackend) FrameSize() int {
	return b.inner.FrameSize()
}

func (b *RecorderBackend) callback(in []float32, out [][]float32) {
	const float32size = 4

	b.innerCb(in, out)
	if len(out) == 0 {
		return
	}

	// Only stereo for now.
	var (
		frameSize = len(out[0])
		buf       = make([]byte, frameSize*2*float32size)
		cur       = buf
	)
	for i := 0; i < frameSize; i++ {
		binary.LittleEndian.PutUint32(cur, math.Float32bits(out[0][i]))
		cur = cur[float32size:]
		binary.LittleEndian.PutUint32(cur, math.Float32bits(out[1][i]))
		cur = cur[float32size:]
	}

	b.w.Write(buf)
}
