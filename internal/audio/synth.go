package audio

import (
	"math"
	"math/rand/v2"

	"github.com/danielriddell21/nemesis/internal/sim"
)

const (
	SampleRate      = 44100
	ChannelCount    = 2
	BitDepthInBytes = 2
	bytesPerFrame   = ChannelCount * BitDepthInBytes
)

type Cue int

const (
	CuePing Cue = iota
	CueStep
	CueCreak
	CueDoor
	CueConsole
	CueHiss
	CueScreech
	CueDeath
	CueEscape
	CueDecoy
	CueMenuMove
	CueMenuSelect
)

func CueFor(k sim.ObservationKind) (Cue, bool) {
	switch k {
	case sim.ObsTrackerPing:
		return CuePing, true
	case sim.ObsStep:
		return CueStep, true
	case sim.ObsVentCreak, sim.ObsAlienCreak:
		return CueCreak, true
	case sim.ObsDoorOpen:
		return CueDoor, true
	case sim.ObsConsole:
		return CueConsole, true
	case sim.ObsDecoy:
		return CueDecoy, true
	case sim.ObsAlienHeard:
		return CueHiss, true
	case sim.ObsAlienSeen:
		return CueScreech, true
	case sim.ObsDeath:
		return CueDeath, true
	case sim.ObsEscape:
		return CueEscape, true
	default:
		return 0, false
	}
}

func Synth() map[Cue][]byte {
	return map[Cue][]byte{
		CuePing:       synthPing(),
		CueStep:       synthStep(),
		CueCreak:      synthCreak(),
		CueDoor:       synthDoor(),
		CueConsole:    synthConsole(),
		CueHiss:       synthHiss(),
		CueScreech:    synthScreech(),
		CueDeath:      synthDeath(),
		CueEscape:     synthEscape(),
		CueDecoy:      synthDecoy(),
		CueMenuMove:   synthMenuMove(),
		CueMenuSelect: synthMenuSelect(),
	}
}

// Pan converts a bearing relative to the listener's facing (0 = dead ahead,
// positive to the right) into constant-power left/right channel gains in
// [0, 1]. Sounds ahead or behind sit centred; sounds to a side swing toward
// that ear.
func Pan(bearing float64) (left, right float64) {
	// Fold front/back onto the same left-right axis: a sound directly behind
	// pans the same as one directly ahead (centred).
	x := math.Sin(bearing) // -1 hard left, +1 hard right
	angle := (x + 1) * (math.Pi / 4)
	return math.Cos(angle), math.Sin(angle)
}

// Panned returns a copy of interleaved 16-bit stereo PCM with the left and
// right channels scaled by the given gains.
func Panned(pcm []byte, left, right float64) []byte {
	out := make([]byte, len(pcm))
	for i := 0; i+bytesPerFrame <= len(pcm); i += bytesPerFrame {
		l := int16(uint16(pcm[i]) | uint16(pcm[i+1])<<8)
		r := int16(uint16(pcm[i+2]) | uint16(pcm[i+3])<<8)
		l = int16(float64(l) * left)
		r = int16(float64(r) * right)
		out[i], out[i+1] = byte(l), byte(uint16(l)>>8)
		out[i+2], out[i+3] = byte(r), byte(uint16(r)>>8)
	}
	return out
}

const maxMenaceBand = 4

func MenaceBand(menace float64) int {
	b := int(menace * (maxMenaceBand + 1))
	if b > maxMenaceBand {
		b = maxMenaceBand
	}
	if b < 0 {
		b = 0
	}
	return b
}

const ambientLoop = 8.0

func Ambient(band int) []byte {
	if band < 0 {
		band = 0
	} else if band > maxMenaceBand {
		band = maxMenaceBand
	}
	// The station's air handlers: a low drone that flattens, quickens and
	// gains a metallic edge as the hunter closes in.
	detune := float64(band) * 0.8
	pulses := float64(2 + band)
	edgeAmp := 0.02 * float64(band)
	return renderPCM(ambientLoop, func(t float64) float64 {
		pulse := 0.6 + 0.4*math.Sin(2*math.Pi*(pulses/ambientLoop)*t-math.Pi/2)
		v := 0.16*math.Sin(2*math.Pi*(48-detune)*t) +
			0.10*math.Sin(2*math.Pi*(72-detune)*t) +
			0.06*math.Sin(2*math.Pi*(96-detune*2)*t)
		edge := edgeAmp * math.Sin(2*math.Pi*233*t) * math.Sin(2*math.Pi*0.9*t)
		return v*pulse + edge
	})
}

func renderPCM(dur float64, gen func(t float64) float64) []byte {
	n := int(dur * SampleRate)
	buf := make([]byte, n*bytesPerFrame)
	for i := range n {
		v := gen(float64(i) / SampleRate)
		if v > 1 {
			v = 1
		} else if v < -1 {
			v = -1
		}
		s := int16(v * 32767)
		lo, hi := byte(s), byte(s>>8)
		off := i * bytesPerFrame
		buf[off], buf[off+1] = lo, hi   // left
		buf[off+2], buf[off+3] = lo, hi // right
	}
	return buf
}

func env(t, decay float64) float64 { return math.Exp(-t * decay) }

func noiseSource() func() float64 {
	// Deterministic noise so every build ships identical sounds.
	r := rand.New(rand.NewPCG(0x51ee7, 0xbeef))
	return func() float64 { return r.Float64()*2 - 1 }
}

func synthPing() []byte {
	return renderPCM(0.5, func(t float64) float64 {
		blip := 0.5 * env(t, 18) * math.Sin(2*math.Pi*1150*t)
		echo := 0.18 * env(t-0.12, 22) * math.Sin(2*math.Pi*1725*t)
		if t < 0.12 {
			echo = 0
		}
		return blip + echo
	})
}

func synthStep() []byte {
	noise := noiseSource()
	return renderPCM(0.09, func(t float64) float64 {
		return env(t, 55) * (0.22*math.Sin(2*math.Pi*70*t) + 0.08*noise())
	})
}

func synthCreak() []byte {
	return renderPCM(0.45, func(t float64) float64 {
		f := 320 - 170*t/0.45
		wobble := 1 + 0.04*math.Sin(2*math.Pi*13*t)
		v := math.Sin(2 * math.Pi * f * wobble * t)
		// Squeeze the sine toward a squeaky edge.
		v = math.Copysign(math.Pow(math.Abs(v), 0.6), v)
		return 0.28 * env(t, 5) * v
	})
}

func synthDoor() []byte {
	noise := noiseSource()
	return renderPCM(0.6, func(t float64) float64 {
		hiss := 0.16 * env(t, 7) * noise()
		clunk := 0.0
		if t > 0.35 {
			clunk = 0.4 * env(t-0.35, 30) * math.Sin(2*math.Pi*55*(t-0.35))
		}
		return hiss + clunk
	})
}

func synthConsole() []byte {
	return renderPCM(1.1, func(t float64) float64 {
		rise := 50 + 70*math.Min(t/0.8, 1)
		hum := 0.3 * math.Sin(2*math.Pi*rise*t) * math.Min(t*4, 1) * env(t, 1.2)
		click := 0.2 * env(t, 60) * math.Sin(2*math.Pi*900*t)
		return hum + click
	})
}

func synthHiss() []byte {
	noise := noiseSource()
	return renderPCM(0.6, func(t float64) float64 {
		shape := math.Sin(math.Pi * math.Min(t/0.6, 1))
		return 0.2 * shape * noise() * (0.6 + 0.4*math.Sin(2*math.Pi*300*t))
	})
}

func synthScreech() []byte {
	return renderPCM(0.9, func(t float64) float64 {
		f := 1400 + 500*math.Sin(2*math.Pi*7*t)
		v := math.Sin(2*math.Pi*f*t) * math.Sin(2*math.Pi*11*t)
		return 0.4 * env(t, 3) * math.Copysign(math.Pow(math.Abs(v), 0.5), v)
	})
}

func synthDeath() []byte {
	noise := noiseSource()
	return renderPCM(1.4, func(t float64) float64 {
		f := 110 * math.Exp(-t*1.6)
		return env(t, 2.2) * (0.5*math.Sin(2*math.Pi*f*t) + 0.15*noise())
	})
}

func synthDecoy() []byte {
	noise := noiseSource()
	return renderPCM(0.35, func(t float64) float64 {
		// A tinny two-tone chirp over a little clatter: the noisemaker calling
		// for attention.
		tone := 880.0
		if math.Mod(t, 0.16) > 0.08 {
			tone = 1180
		}
		beep := 0.3 * env(t, 6) * square(2*math.Pi*tone*t)
		clatter := 0.08 * env(t, 40) * noise()
		return beep + clatter
	})
}

func synthMenuMove() []byte {
	return renderPCM(0.06, func(t float64) float64 {
		return 0.2 * env(t, 42) * math.Sin(2*math.Pi*660*t)
	})
}

func synthMenuSelect() []byte {
	return renderPCM(0.16, func(t float64) float64 {
		// Two quick rising blips: a soft confirm.
		f := 620.0
		if t > 0.06 {
			f = 930
		}
		return 0.22 * env(t, 14) * square(2*math.Pi*f*t)
	})
}

func square(phase float64) float64 {
	if math.Sin(phase) >= 0 {
		return 1
	}
	return -1
}

func synthEscape() []byte {
	noise := noiseSource()
	return renderPCM(1.6, func(t float64) float64 {
		whoosh := 0.2 * env(t, 4) * noise()
		chord := 0.0
		if t > 0.5 {
			u := t - 0.5
			chord = 0.12 * env(u, 1.5) * (math.Sin(2*math.Pi*220*u) +
				math.Sin(2*math.Pi*277*u) + math.Sin(2*math.Pi*330*u))
		}
		return whoosh + chord
	})
}
