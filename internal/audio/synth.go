package audio

import (
	"math"

	"github.com/danielriddell21/crucible/synth"

	"github.com/danielriddell21/nemesis/internal/sim"
)

// SampleRate is the playback rate the oto context is configured with.
const SampleRate = synth.SampleRate

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
	return synth.Render(ambientLoop, func(t float64) float64 {
		pulse := 0.6 + 0.4*math.Sin(2*math.Pi*(pulses/ambientLoop)*t-math.Pi/2)
		v := 0.16*math.Sin(2*math.Pi*(48-detune)*t) +
			0.10*math.Sin(2*math.Pi*(72-detune)*t) +
			0.06*math.Sin(2*math.Pi*(96-detune*2)*t)
		edge := edgeAmp * math.Sin(2*math.Pi*233*t) * math.Sin(2*math.Pi*0.9*t)
		return v*pulse + edge
	})
}

func synthPing() []byte {
	return synth.Render(0.5, func(t float64) float64 {
		blip := 0.5 * synth.Env(t, 18) * math.Sin(2*math.Pi*1150*t)
		echo := 0.18 * synth.Env(t-0.12, 22) * math.Sin(2*math.Pi*1725*t)
		if t < 0.12 {
			echo = 0
		}
		return blip + echo
	})
}

func synthStep() []byte {
	noise := synth.Noise(0x51ee7, 0xbeef)
	return synth.Render(0.09, func(t float64) float64 {
		return synth.Env(t, 55) * (0.22*math.Sin(2*math.Pi*70*t) + 0.08*noise())
	})
}

func synthCreak() []byte {
	return synth.Render(0.45, func(t float64) float64 {
		f := 320 - 170*t/0.45
		wobble := 1 + 0.04*math.Sin(2*math.Pi*13*t)
		v := math.Sin(2 * math.Pi * f * wobble * t)
		// Squeeze the sine toward a squeaky edge.
		v = math.Copysign(math.Pow(math.Abs(v), 0.6), v)
		return 0.28 * synth.Env(t, 5) * v
	})
}

func synthDoor() []byte {
	noise := synth.Noise(0x51ee7, 0xbeef)
	return synth.Render(0.6, func(t float64) float64 {
		hiss := 0.16 * synth.Env(t, 7) * noise()
		clunk := 0.0
		if t > 0.35 {
			clunk = 0.4 * synth.Env(t-0.35, 30) * math.Sin(2*math.Pi*55*(t-0.35))
		}
		return hiss + clunk
	})
}

func synthConsole() []byte {
	return synth.Render(1.1, func(t float64) float64 {
		rise := 50 + 70*math.Min(t/0.8, 1)
		hum := 0.3 * math.Sin(2*math.Pi*rise*t) * math.Min(t*4, 1) * synth.Env(t, 1.2)
		click := 0.2 * synth.Env(t, 60) * math.Sin(2*math.Pi*900*t)
		return hum + click
	})
}

func synthHiss() []byte {
	noise := synth.Noise(0x51ee7, 0xbeef)
	return synth.Render(0.6, func(t float64) float64 {
		shape := math.Sin(math.Pi * math.Min(t/0.6, 1))
		return 0.2 * shape * noise() * (0.6 + 0.4*math.Sin(2*math.Pi*300*t))
	})
}

func synthScreech() []byte {
	return synth.Render(0.9, func(t float64) float64 {
		f := 1400 + 500*math.Sin(2*math.Pi*7*t)
		v := math.Sin(2*math.Pi*f*t) * math.Sin(2*math.Pi*11*t)
		return 0.4 * synth.Env(t, 3) * math.Copysign(math.Pow(math.Abs(v), 0.5), v)
	})
}

func synthDeath() []byte {
	noise := synth.Noise(0x51ee7, 0xbeef)
	return synth.Render(1.4, func(t float64) float64 {
		f := 110 * math.Exp(-t*1.6)
		return synth.Env(t, 2.2) * (0.5*math.Sin(2*math.Pi*f*t) + 0.15*noise())
	})
}

func synthDecoy() []byte {
	noise := synth.Noise(0x51ee7, 0xbeef)
	return synth.Render(0.35, func(t float64) float64 {
		// A tinny two-tone chirp over a little clatter: the noisemaker calling
		// for attention.
		tone := 880.0
		if math.Mod(t, 0.16) > 0.08 {
			tone = 1180
		}
		beep := 0.3 * synth.Env(t, 6) * square(2*math.Pi*tone*t)
		clatter := 0.08 * synth.Env(t, 40) * noise()
		return beep + clatter
	})
}

func synthMenuMove() []byte {
	return synth.Render(0.06, func(t float64) float64 {
		return 0.2 * synth.Env(t, 42) * math.Sin(2*math.Pi*660*t)
	})
}

func synthMenuSelect() []byte {
	return synth.Render(0.16, func(t float64) float64 {
		// Two quick rising blips: a soft confirm.
		f := 620.0
		if t > 0.06 {
			f = 930
		}
		return 0.22 * synth.Env(t, 14) * square(2*math.Pi*f*t)
	})
}

func square(phase float64) float64 {
	if math.Sin(phase) >= 0 {
		return 1
	}
	return -1
}

func synthEscape() []byte {
	noise := synth.Noise(0x51ee7, 0xbeef)
	return synth.Render(1.6, func(t float64) float64 {
		whoosh := 0.2 * synth.Env(t, 4) * noise()
		chord := 0.0
		if t > 0.5 {
			u := t - 0.5
			chord = 0.12 * synth.Env(u, 1.5) * (math.Sin(2*math.Pi*220*u) +
				math.Sin(2*math.Pi*277*u) + math.Sin(2*math.Pi*330*u))
		}
		return whoosh + chord
	})
}
