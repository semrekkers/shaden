package midi

import (
	"fmt"
	"time"

	"github.com/rakyll/portmidi"

	"buddin.us/musictheory"
	"buddin.us/shaden/dsp"
	"buddin.us/shaden/unit"
)

func newInput(creator streamCreator, receiver eventReceiver) func(*unit.IO, unit.Config) (*unit.Unit, error) {
	return func(io *unit.IO, c unit.Config) (*unit.Unit, error) {
		var (
			config struct {
				Rate     int
				Device   int
				Channels []int
				Feedback int
			}
			pitches = map[int]float64{}
			p       = musictheory.NewPitch(musictheory.C, musictheory.Natural, 0)
		)

		for i := 12; i < 127; i++ {
			pitches[i] = dsp.Frequency(p.Freq(), c.SampleRate).Float64()
			p = p.Transpose(musictheory.Minor(2))
		}

		config.Feedback = -1
		if err := c.Decode(&config); err != nil {
			return nil, err
		}

		stream, err := creator.NewStream(portmidi.DeviceID(config.Device), int64(c.FrameSize))
		if err != nil {
			return nil, err
		}

		var feedbackStream *portmidi.Stream
		if config.Feedback != -1 {
			feedbackStream, err = portmidi.NewOutputStream(portmidi.DeviceID(config.Feedback), int64(dsp.FrameSize), 0)
			if err != nil {
				return nil, err
			}
		}

		if len(config.Channels) == 0 {
			config.Channels = []int{1}
		}

		if config.Rate == 0 {
			config.Rate = 10
		}

		ctrl := &input{
			stream:    stream,
			eventChan: stream.Channel(time.Duration(config.Rate) * time.Millisecond),
			receiver:  receiver,
			events:    make([]portmidi.Event, c.FrameSize),
			pitches:   pitches,
			frameSize: c.FrameSize,
		}

		for _, ch := range config.Channels {
			io.ExposeOutputProcessor(ctrl.newPitch(ch))
			io.ExposeOutputProcessor(ctrl.newPitchRaw(ch))
			io.ExposeOutputProcessor(ctrl.newGate(ch))
			io.ExposeOutputProcessor(ctrl.newBend(ch))
			for i := 1; i < 128; i++ {
				io.ExposeOutputProcessor(ctrl.newCC(ch, i))
			}
			if feedbackStream != nil {
				for i := 0; i < 64; i++ {
					// Reset button
					feedbackStream.WriteShort(statusNoteOn+int64(ch)-1, int64(i), buttonColorOff)

					io.ExposeOutputProcessor(ctrl.newToggle(ch, i, feedbackStream))
					io.ExposeOutputProcessor(ctrl.newButton(ch, i, feedbackStream))
				}
			}
		}

		return unit.NewUnit(io, ctrl), nil
	}
}

type input struct {
	stream    eventStream
	eventChan <-chan portmidi.Event
	receiver  eventReceiver
	events    []portmidi.Event
	pitches   map[int]float64
	frameSize int
}

func (in *input) newPitch(ch int) *pitch {
	return &pitch{
		input: in,
		ch:    ch,
		out:   unit.NewOut(fmt.Sprintf("%d/pitch", ch), make([]float64, in.frameSize)),
	}
}

func (in *input) newPitchRaw(ch int) *pitchRaw {
	return &pitchRaw{
		input: in,
		ch:    ch,
		out:   unit.NewOut(fmt.Sprintf("%d/pitchraw", ch), make([]float64, in.frameSize)),
	}
}

func (in *input) newGate(ch int) *gate {
	return &gate{
		input:   in,
		stateFn: gateUp,
		state:   &gateState{which: -1, chOffset: int64(ch) - 1},
		out:     unit.NewOut(fmt.Sprintf("%d/gate", ch), make([]float64, in.frameSize)),
	}
}

func (in *input) newCC(ch, num int) *cc {
	return &cc{
		input: in,
		ch:    int64(ch),
		num:   int64(num),
		out:   unit.NewOut(fmt.Sprintf("%d/cc/%d", ch, num), make([]float64, in.frameSize)),
	}
}

func (in *input) newBend(ch int) *bend {
	return &bend{
		input: in,
		ch:    int64(ch),
		out:   unit.NewOut(fmt.Sprintf("%d/bend", ch), make([]float64, in.frameSize)),
	}
}

func (in *input) newToggle(ch, num int, feedback *portmidi.Stream) *toggle {
	return &toggle{
		input:    in,
		ch:       int64(ch),
		num:      int64(num),
		feedback: feedback,
		out:      unit.NewOut(fmt.Sprintf("%d/toggle/%d", ch, num), make([]float64, dsp.FrameSize)),
	}
}

func (in *input) newButton(ch, num int, feedback *portmidi.Stream) *button {
	return &button{
		input:    in,
		ch:       int64(ch),
		num:      int64(num),
		state:    buttonStateReleased,
		value:    -1,
		feedback: feedback,
		out:      unit.NewOut(fmt.Sprintf("%d/button/%d", ch, num), make([]float64, dsp.FrameSize)),
	}
}

func (in *input) IsProcessable() bool {
	return true
}

func (in *input) ProcessSample(i int) {
	if in.stream == nil {
		return
	}
	in.events[i] = in.receiver(in.eventChan)
}

func (in *input) Close() error {
	err := in.stream.Close()
	in.stream = nil
	return err
}

const (
	statusNoteOn     = 144
	statusNoteOff    = 128
	statusCC         = 176
	statusPitchWheel = 224
)

type pitch struct {
	input *input
	ch    int
	freq  float64
	out   *unit.Out
}

func (o *pitch) IsProcessable() bool { return o.out.ExternalNeighborCount() > 0 }
func (o *pitch) Out() *unit.Out      { return o.out }

func (o *pitch) ProcessFrame(n int) {
	for i := 0; i < n; i++ {
		o.ProcessSample(i)
	}
}

func (o *pitch) ProcessSample(i int) {
	if e := o.input.events[i]; e.Status == int64(statusNoteOn+o.ch-1) {
		if v, ok := o.input.pitches[int(e.Data1)]; ok && e.Data2 > 0 {
			o.freq = v
		}
	}
	o.out.Write(i, o.freq)
}

type pitchRaw struct {
	input *input
	ch    int
	note  float64
	out   *unit.Out
}

func (o *pitchRaw) IsProcessable() bool { return o.out.ExternalNeighborCount() > 0 }
func (o *pitchRaw) Out() *unit.Out      { return o.out }

func (o *pitchRaw) ProcessFrame(n int) {
	for i := 0; i < n; i++ {
		o.ProcessSample(i)
	}
}

func (o *pitchRaw) ProcessSample(i int) {
	if e := o.input.events[i]; e.Status == int64(statusNoteOn+o.ch-1) {
		o.note = float64(e.Data1)
	}
	o.out.Write(i, o.note)
}

type gate struct {
	input   *input
	state   *gateState
	stateFn gateStateFunc
	out     *unit.Out
}

type gateState struct {
	event           portmidi.Event
	which, chOffset int64
	value           float64
}

type gateStateFunc func(*gateState) gateStateFunc

func gateRolling(s *gateState) gateStateFunc {
	s.value = -1
	return gateDown
}

func gateDown(s *gateState) gateStateFunc {
	s.value = 1
	if s.event.Status == 0 && s.event.Timestamp == 0 {
		return gateDown
	}

	which := s.event.Data1

	switch s.event.Status {
	case statusNoteOn + s.chOffset:
		if s.event.Data2 > 0 {
			if which != s.which {
				s.which = which
				return gateRolling
			}
			s.which = -1
			return gateUp
		}
		if which == s.which {
			s.which = -1
			return gateUp
		}
	case statusNoteOff + s.chOffset:
		if which == s.which {
			s.which = -1
			return gateUp
		}
	}
	return gateDown
}

func gateUp(s *gateState) gateStateFunc {
	s.value = -1
	if s.event.Status == 0 && s.event.Timestamp == 0 {
		return gateUp
	}
	if s.event.Status == statusNoteOn+s.chOffset && s.event.Data2 > 0 {
		s.which = s.event.Data1
		return gateDown
	}
	return gateUp
}

func (o *gate) IsProcessable() bool { return o.out.ExternalNeighborCount() > 0 }
func (o *gate) Out() *unit.Out      { return o.out }

func (o *gate) ProcessFrame(n int) {
	for i := 0; i < n; i++ {
		o.ProcessSample(i)
	}
}

func (o *gate) ProcessSample(i int) {
	o.state.event = o.input.events[i]
	o.stateFn = o.stateFn(o.state)
	o.out.Write(i, o.state.value)
}

type cc struct {
	input   *input
	ch, num int64
	value   float64
	out     *unit.Out
}

func (o *cc) IsProcessable() bool { return o.out.ExternalNeighborCount() > 0 }
func (o *cc) Out() *unit.Out      { return o.out }

func (o *cc) ProcessFrame(n int) {
	for i := 0; i < n; i++ {
		o.ProcessSample(i)
	}
}

func (o *cc) ProcessSample(i int) {
	if e := o.input.events[i]; e.Status == statusCC+o.ch-1 && e.Data1 == o.num {
		o.value = float64(e.Data2) / 127
	}
	o.out.Write(i, o.value)
}

type bend struct {
	input *input
	ch    int64
	value float64
	out   *unit.Out
}

func (o *bend) IsProcessable() bool { return o.out.ExternalNeighborCount() > 0 }
func (o *bend) Out() *unit.Out      { return o.out }

func (o *bend) ProcessFrame(n int) {
	for i := 0; i < n; i++ {
		o.ProcessSample(i)
	}
}

func (o *bend) ProcessSample(i int) {
	if e := o.input.events[i]; e.Status == statusPitchWheel+o.ch-1 && e.Data1 == 0 {
		o.value = float64(e.Data2) / 127
	}
	o.out.Write(i, o.value)
}

const (
	buttonColorOff    = 0
	buttonColorYellow = 5
	buttonColorGreen  = 1
)

type toggle struct {
	input       *input
	ch, num     int64
	value, used bool
	feedback    *portmidi.Stream
	out         *unit.Out
}

func (o *toggle) IsProcessable() bool { return o.out.ExternalNeighborCount() > 0 }
func (o *toggle) Out() *unit.Out      { return o.out }

func (o *toggle) ProcessFrame(n int) {
	if !o.used {
		o.feedback.WriteShort(statusNoteOn+o.ch-1, o.num, buttonColorYellow)
		o.used = true
	}
	for i := 0; i < n; i++ {
		o.ProcessSample(i)
	}
}

func (o *toggle) ProcessSample(i int) {
	if e := o.input.events[i]; e.Status == statusNoteOn+o.ch-1 && e.Data1 == o.num {
		buttonColor := int64(buttonColorYellow)
		if !o.value {
			buttonColor = buttonColorGreen
		}
		o.feedback.WriteShort(statusNoteOn+o.ch-1, o.num, buttonColor)
		o.value = !o.value
	}

	out := -1.0
	if o.value {
		out = 1.0
	}
	o.out.Write(i, out)
}

type button struct {
	input    *input
	ch, num  int64
	used     bool
	state    buttonStateFunc
	value    float64
	feedback *portmidi.Stream
	out      *unit.Out
}

type buttonStateFunc func(*button, *portmidi.Event)

func (b *button) IsProcessable() bool { return b.out.ExternalNeighborCount() > 0 }
func (b *button) Out() *unit.Out      { return b.out }

func (b *button) ProcessFrame(n int) {
	if !b.used {
		b.feedback.WriteShort(statusNoteOn+b.ch-1, b.num, buttonColorYellow)
		b.used = true
	}
	for i := 0; i < n; i++ {
		b.ProcessSample(i)
	}
}

func (b *button) ProcessSample(i int) {
	b.state(b, &b.input.events[i])
	b.out.Write(i, b.value)
}

func buttonStateReleased(b *button, e *portmidi.Event) {
	if e.Status == statusNoteOn+b.ch-1 && e.Data1 == b.num {
		b.state = buttonStatePressed
		b.value = 1
		b.feedback.WriteShort(statusNoteOn+b.ch-1, b.num, buttonColorGreen)
	}
}

func buttonStatePressed(b *button, e *portmidi.Event) {
	if e.Status == statusNoteOff+b.ch-1 && e.Data1 == b.num {
		b.state = buttonStateReleased
		b.value = -1
		b.feedback.WriteShort(statusNoteOn+b.ch-1, b.num, buttonColorYellow)
	}
}
