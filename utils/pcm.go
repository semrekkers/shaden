package utils

import (
	"encoding/binary"
	"io"
	"math"
)

const sampleSize = 8

type PcmWriter struct {
	w   io.Writer
	buf []byte
	pos int
}

func NewPcmWriter(w io.Writer, bufSize int) *PcmWriter {
	return &PcmWriter{
		w:   w,
		buf: make([]byte, bufSize*sampleSize),
	}
}

func (pw *PcmWriter) WriteSample(v float64) error {
	binary.LittleEndian.PutUint64(pw.buf[pw.pos:], math.Float64bits(v))
	pw.pos += sampleSize
	if pw.pos == len(pw.buf) {
		return pw.Flush()
	}
	return nil
}

func (pw *PcmWriter) Flush() error {
	if _, err := pw.w.Write(pw.buf[:pw.pos]); err != nil {
		return err
	}
	pw.pos = 0
	return nil
}
