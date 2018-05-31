package engine

import (
	"fmt"
	"testing"
	"time"

	"github.com/brettbuddin/shaden/unit"
	"github.com/stretchr/testify/require"
)

const (
	sampleRate = 44100
	frameSize  = 256
)

func TestEngine_Stop(t *testing.T) {
	be := backend{
		start:     func(func([]float32, [][]float32)) error { return nil },
		stop:      func() error { return nil },
		frameSize: frameSize * 2,
	}
	e, err := New(be, frameSize)
	require.NoError(t, err)
	go e.Run()
	go func() {
		for range e.Errors() {
		}
	}()
	require.NoError(t, e.Stop())
}

func TestEngine_StopProxyBackendError(t *testing.T) {
	be := backend{
		start:     func(func([]float32, [][]float32)) error { return nil },
		stop:      func() error { return fmt.Errorf("exploded") },
		frameSize: frameSize * 2,
	}
	e, err := New(be, frameSize)
	require.NoError(t, err)
	go e.Run()
	go func() {
		for range e.Errors() {
		}
	}()
	require.Error(t, e.Stop())
}

func TestEngine_StartError(t *testing.T) {
	be := backend{
		start:     func(func([]float32, [][]float32)) error { return fmt.Errorf("exploded") },
		stop:      func() error { return nil },
		frameSize: frameSize * 2,
	}
	e, err := New(be, frameSize)
	require.NoError(t, err)
	go e.Run()

	select {
	case err = <-e.Errors():
	case <-time.After(5 * time.Second):
		t.Error("timeout waiting for start error")
	}

	require.Error(t, err)
	require.NoError(t, e.Stop())
}

func TestEngine_MountAndUnmount(t *testing.T) {
	size := frameSize * 2

	be := backend{
		start: func(cb func([]float32, [][]float32)) error {
			out := make([][]float32, 2)
			for i := 0; i < size; i++ {
				out[0] = make([]float32, size)
				out[1] = make([]float32, size)
			}
			cb(make([]float32, size), out) // receive mount message
			cb(make([]float32, size), out) // receive unmount message
			return nil
		},
		stop:      func() error { return nil },
		frameSize: size,
	}
	e, err := New(be, frameSize)
	require.NoError(t, err)
	require.Equal(t, 3, e.graph.Size())

	go e.Run()
	go func() {
		for range e.Errors() {
		}
	}()

	// Unit with no inputs and outputs
	io := unit.NewIO("example", frameSize)
	u := unit.NewUnit(io, nil)

	// Send a MountUnit message to the engine
	msg := NewMessage(MountUnit(u))
	err = e.SendMessage(msg)
	require.NoError(t, err)

	var reply *Reply
	select {
	case reply = <-msg.Reply:
	case <-time.After(5 * time.Second):
		t.Error("timeout waiting for receive reply")
	}
	require.Nil(t, reply.Data)
	require.NoError(t, reply.Error)

	require.Equal(t, 4, e.graph.Size())

	// Send UnmountUnit message to the engine
	msg = NewMessage(UnmountUnit(u))

	err = e.SendMessage(msg)
	require.NoError(t, err)

	select {
	case reply = <-msg.Reply:
	case <-time.After(5 * time.Second):
		t.Error("timeout waiting for receive reply")
	}
	require.Nil(t, reply.Data)
	require.NoError(t, reply.Error)

	require.Equal(t, 3, e.graph.Size())
	require.NoError(t, e.Stop())
}

func TestEngine_MountAndReset(t *testing.T) {
	size := frameSize * 2

	be := backend{
		start: func(cb func([]float32, [][]float32)) error {
			out := make([][]float32, 2)
			for i := 0; i < size; i++ {
				out[0] = make([]float32, size)
				out[1] = make([]float32, size)
			}
			cb(make([]float32, size), out)
			cb(make([]float32, size), out)
			return nil
		},
		stop:      func() error { return nil },
		frameSize: size,
	}
	e, err := New(be, frameSize)
	require.NoError(t, err)
	require.Equal(t, 3, e.graph.Size())

	go e.Run()
	go func() {
		for range e.Errors() {
		}
	}()

	io := unit.NewIO("example", frameSize)
	u := unit.NewUnit(io, nil)

	msg := NewMessage(MountUnit(u))
	err = e.SendMessage(msg)
	require.NoError(t, err)

	var reply *Reply
	select {
	case reply = <-msg.Reply:
	case <-time.After(5 * time.Second):
		t.Error("timeout waiting for receive reply")
	}
	require.Nil(t, reply.Data)
	require.NoError(t, reply.Error)

	require.Equal(t, 4, e.graph.Size())

	msg = NewMessage(Clear)
	err = e.SendMessage(msg)
	require.NoError(t, err)

	select {
	case reply = <-msg.Reply:
	case <-time.After(5 * time.Second):
		t.Error("timeout waiting for receive reply")
	}
	require.NoError(t, reply.Error)

	require.Equal(t, 3, e.graph.Size())
	require.NoError(t, e.Stop())
}

type backend struct {
	start                 func(func([]float32, [][]float32)) error
	stop                  func() error
	sampleRate, frameSize int
}

func (b backend) Start(cb func([]float32, [][]float32)) error { return b.start(cb) }
func (b backend) Stop() error                                 { return b.stop() }
func (b backend) FrameSize() int                              { return b.frameSize }
func (b backend) SampleRate() int                             { return b.sampleRate }
