package unit

import "github.com/mitchellh/mapstructure"

var (
	builders = map[string]IOBuilder{
		"abs":      buildUnary(unaryAbs),
		"bipolar":  buildUnary(unaryBipolar),
		"ceil":     buildUnary(unaryCeil),
		"floor":    buildUnary(unaryFloor),
		"invert":   buildUnary(unaryInv),
		"not":      buildUnary(unaryNOT),
		"noop":     buildUnary(unaryNoop),
		"unipolar": buildUnary(unaryUnipolar),

		"and":   buildBinary(binaryAND),
		"diff":  buildBinary(binaryDiff),
		"div":   buildBinary(binaryDiv),
		"gt":    buildBinary(binaryGT),
		"imply": buildBinary(binaryIMPLY),
		"lt":    buildBinary(binaryLT),
		"max":   buildBinary(binaryMax),
		"min":   buildBinary(binaryMin),
		"mod":   buildBinary(binaryMod),
		"mult":  buildBinary(binaryMult),
		"nand":  buildBinary(binaryNAND),
		"nor":   buildBinary(binaryNOR),
		"or":    buildBinary(binaryOR),
		"sum":   buildBinary(binarySum),
		"xnor":  buildBinary(binaryXNOR),
		"xor":   buildBinary(binaryXOR),

		"adjust":             newAdjust,
		"adsr":               newAdsr,
		"center":             newCenter,
		"chance":             newChance,
		"chebyshev":          newChebyshev,
		"clip":               newClip,
		"clock":              newClock,
		"clock-div":          newClockDiv,
		"clock-mult":         newClockMult,
		"cluster":            newCluster,
		"cond":               newCond,
		"count":              newCount,
		"debug":              newDebug,
		"decimate":           newDecimate,
		"delay":              newDelay,
		"demux":              newDemux,
		"dynamics":           newDynamics,
		"euclid":             newEuclid,
		"filter":             newFilter,
		"fold":               newFold,
		"gate":               newGate,
		"gate-mix":           newGateMix,
		"gate-series":        newGateSeries,
		"gen":                newGen,
		"lag":                newLag,
		"latch":              newLatch,
		"lerp":               newInterpolate,
		"logic":              newLogic,
		"low-gen":            newLowGen,
		"midi-hz":            newMIDIToHz,
		"mix":                newMix,
		"morph":              newMorph,
		"mux":                newMux,
		"overload":           newOverload,
		"pan":                newPan,
		"panmix":             newPanMix,
		"pitch":              newPitch,
		"quantize":           newQuantize,
		"random-series":      newRandomSeries,
		"rcd":                newRCD,
		"reverb":             newReverb,
		"sample":             newWAVSample,
		"step":               newStep,
		"shift":              newShift,
		"slope":              newSlope,
		"smooth":             newSmooth,
		"stages":             newStages,
		"switch":             newSwitch,
		"toggle":             newToggle,
		"transpose":          newTranspose,
		"transpose-interval": newTransposeInterval,
		"val-gate":           newValToGate,
		"xfade":              newCrossfade,
		"xfeed":              newCrossfeed,
	}
)

// IOBuilder provides an IO, containing identifying information, for a Unit to be constructed around.
type IOBuilder func(*IO, Config) (*Unit, error)

// Builder constructs a Unit of some type.
type Builder func(Config) (*Unit, error)

// Config is a map that's used to provide configuration options to Builders.
type Config struct {
	Values                map[string]interface{}
	SampleRate, FrameSize int
}

// Decode loads a struct with the contents of the raw Config object.
func (c Config) Decode(v interface{}) error {
	return mapstructure.Decode(c.Values, v)
}

// Builders returns all Builders for all Units provided by this package.
func Builders() map[string]Builder {
	return PrepareBuilders(builders)
}

// PrepareBuilders converts a set of IOBuilders to a set of Builders.
func PrepareBuilders(builders map[string]IOBuilder) map[string]Builder {
	m := map[string]Builder{}
	for k, v := range builders {
		m[k] = func(typ string, f IOBuilder) Builder {
			return func(c Config) (*Unit, error) {
				return f(NewIO(typ, c.FrameSize), c)
			}
		}(k, v)
	}
	return m
}

func buildUnary(op unaryOp) IOBuilder {
	return func(io *IO, c Config) (*Unit, error) {
		return newUnary(io, op)
	}
}

func buildBinary(op binaryOp) IOBuilder {
	return func(io *IO, c Config) (*Unit, error) {
		return newBinary(io, op)
	}
}
