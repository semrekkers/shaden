package utils

import (
	"bytes"
	"testing"
)

func TestPcmWriter_WriteSample(t *testing.T) {
	tests := []struct {
		input  []float64
		output []byte
	}{
		{[]float64{0, 1}, []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf0, 0x3f}},
		{[]float64{0.67436244, -0.86745342}, []byte{0xa5, 0x6c, 0x2e, 0x8a, 0x60, 0x94, 0xe5, 0x3f, 0xdd, 0x81, 0xb6, 0xac, 0x2d, 0xc2, 0xeb, 0xbf}},
		{[]float64{0, 0, 0}, []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
	}

	var buffer bytes.Buffer
	for _, tt := range tests {
		pw := NewPcmWriter(&buffer, 2)
		for _, sample := range tt.input {
			if err := pw.WriteSample(sample); err != nil {
				t.Errorf("PcmWriter.WriteSample() error = %v", err)
			}
		}
		result := buffer.Bytes()
		if len(result) != len(tt.output) {
			t.Errorf("result is not equal to output")
		}
		for i := 0; i < len(tt.output); i++ {
			if result[i] != tt.output[i] {
				t.Errorf("result[%d] = %#x, want %#x", i, result[i], tt.output[i])
			}
		}

		buffer.Reset()
	}
}
