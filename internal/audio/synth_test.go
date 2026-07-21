package audio

import (
	"bytes"
	"testing"

	"github.com/danielriddell21/crucible/synth"

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
		if len(pcm)%synth.BytesPerFrame != 0 {
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

// Pan and Panned now live in crucible/synth and are covered by its tests.
