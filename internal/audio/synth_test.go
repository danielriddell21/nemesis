package audio

import (
	"bytes"
	"math"
	"testing"

	"github.com/danielriddell21/nemesis/internal/sim"
)

func TestSynthCoversAllCues(t *testing.T) {
	sounds := Synth()
	cues := []Cue{CuePing, CueStep, CueCreak, CueDoor, CueConsole, CueHiss, CueScreech, CueDeath, CueEscape, CueDecoy}
	for _, c := range cues {
		pcm, ok := sounds[c]
		if !ok || len(pcm) == 0 {
			t.Errorf("cue %d has no PCM", c)
		}
		if len(pcm)%bytesPerFrame != 0 {
			t.Errorf("cue %d PCM not frame-aligned", c)
		}
	}
}

func TestSynthDeterministic(t *testing.T) {
	a, b := Synth(), Synth()
	for c := range a {
		if !bytes.Equal(a[c], b[c]) {
			t.Errorf("cue %d differs between builds", c)
		}
	}
}

func TestSynthNotSilent(t *testing.T) {
	for c, pcm := range Synth() {
		peak := 0
		for i := 0; i+1 < len(pcm); i += 2 {
			v := int(int16(uint16(pcm[i]) | uint16(pcm[i+1])<<8))
			if v < 0 {
				v = -v
			}
			if v > peak {
				peak = v
			}
		}
		if peak < 1000 {
			t.Errorf("cue %d peaks at %d, effectively silent", c, peak)
		}
		if peak > 32700 {
			t.Errorf("cue %d peaks at %d, clipping", c, peak)
		}
	}
}

func TestCueFor(t *testing.T) {
	tests := []struct {
		kind sim.ObservationKind
		cue  Cue
		ok   bool
	}{
		{sim.ObsTrackerPing, CuePing, true},
		{sim.ObsStep, CueStep, true},
		{sim.ObsVentCreak, CueCreak, true},
		{sim.ObsAlienCreak, CueCreak, true},
		{sim.ObsDoorOpen, CueDoor, true},
		{sim.ObsConsole, CueConsole, true},
		{sim.ObsAlienHeard, CueHiss, true},
		{sim.ObsAlienSeen, CueScreech, true},
		{sim.ObsDeath, CueDeath, true},
		{sim.ObsEscape, CueEscape, true},
		{sim.ObsAlienState, 0, false},
		{sim.ObsDirectorNudge, 0, false},
	}
	for _, tt := range tests {
		cue, ok := CueFor(tt.kind)
		if ok != tt.ok || (ok && cue != tt.cue) {
			t.Errorf("CueFor(%v) = %v,%v want %v,%v", tt.kind, cue, ok, tt.cue, tt.ok)
		}
	}
}

func TestMenaceBand(t *testing.T) {
	tests := []struct {
		menace float64
		want   int
	}{
		{0, 0},
		{0.19, 0},
		{0.21, 1},
		{0.5, 2},
		{0.99, 4},
		{1, 4},
		{-1, 0},
	}
	for _, tt := range tests {
		if got := MenaceBand(tt.menace); got != tt.want {
			t.Errorf("MenaceBand(%v) = %d, want %d", tt.menace, got, tt.want)
		}
	}
}

func TestAmbientBandsDiffer(t *testing.T) {
	calm := Ambient(0)
	tense := Ambient(maxMenaceBand)
	if len(calm) != len(tense) {
		t.Fatal("ambient loops should share a length")
	}
	if bytes.Equal(calm, tense) {
		t.Error("menace bands should sound different")
	}
	if !bytes.Equal(Ambient(-3), calm) || !bytes.Equal(Ambient(99), tense) {
		t.Error("out-of-range bands should clamp")
	}
}

func TestCueForDecoy(t *testing.T) {
	cue, ok := CueFor(sim.ObsDecoy)
	if !ok || cue != CueDecoy {
		t.Errorf("CueFor(ObsDecoy) = %v,%v want %v,true", cue, ok, CueDecoy)
	}
}

func TestPanCentresAndSwings(t *testing.T) {
	// Dead ahead: balanced. A sound to the right favours the right channel.
	l, r := Pan(0)
	if math.Abs(l-r) > 1e-9 {
		t.Errorf("ahead should be balanced, got L=%v R=%v", l, r)
	}
	l, r = Pan(math.Pi / 2) // hard right
	if r <= l {
		t.Errorf("hard right should favour the right channel, got L=%v R=%v", l, r)
	}
	l, r = Pan(-math.Pi / 2) // hard left
	if l <= r {
		t.Errorf("hard left should favour the left channel, got L=%v R=%v", l, r)
	}
	// Constant power: L^2 + R^2 stays ~1 across the arc.
	for _, b := range []float64{-1.2, -0.4, 0, 0.7, 1.5} {
		l, r = Pan(b)
		if p := l*l + r*r; math.Abs(p-1) > 1e-9 {
			t.Errorf("Pan(%v): power %v, want 1", b, p)
		}
	}
}

func TestPannedScalesChannels(t *testing.T) {
	// One stereo frame at full scale on both channels; mute the left, keep right.
	pcm := []byte{0xff, 0x7f, 0xff, 0x7f} // L=+32767, R=+32767
	out := Panned(pcm, 0, 1)
	l := int16(uint16(out[0]) | uint16(out[1])<<8)
	r := int16(uint16(out[2]) | uint16(out[3])<<8)
	if l != 0 {
		t.Errorf("left channel should be muted, got %d", l)
	}
	if r != 32767 {
		t.Errorf("right channel should be untouched, got %d", r)
	}
}
