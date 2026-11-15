package audio

import (
	"bytes"
	"math"
	"testing"

	"github.com/danielriddell21/crucible/synth"

	"github.com/danielriddell21/nemesis/internal/sim"
)

func TestSynthCoversAllCues(t *testing.T) {
	sounds := Synth()
	cues := []Cue{
		CuePing, CueStep, CueCreak, CueDoor, CueConsole, CueHiss, CueScreech, CueDeath, CueEscape, CueDecoy,
		CueMenuMove, CueMenuSelect, CueTorch, CueBattery, CueHeartbeat, CueGrowl, CueGroan,
	}
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
		{sim.ObsAlienGlow, CueHiss, true},
		{sim.ObsTorch, CueTorch, true},
		{sim.ObsBattery, CueBattery, true},
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

func sample(pcm []byte, frame int) float64 {
	i := frame * synth.BytesPerFrame
	return float64(int16(uint16(pcm[i])|uint16(pcm[i+1])<<8)) / 32768
}

func TestAmbientLoopSeamIsSilent(t *testing.T) {
	for band := 0; band <= maxMenaceBand; band++ {
		pcm := Ambient(band)
		frames := len(pcm) / synth.BytesPerFrame
		// The jump across the loop point must be no bigger than the largest
		// step between neighbouring samples inside the loop.
		worst := 0.0
		for f := 1; f < frames; f++ {
			worst = math.Max(worst, math.Abs(sample(pcm, f)-sample(pcm, f-1)))
		}
		seam := math.Abs(sample(pcm, 0) - sample(pcm, frames-1))
		if seam > worst*1.5 {
			t.Errorf("band %d: loop seam jumps %.4f, the largest step inside the loop is %.4f", band, seam, worst)
		}
	}
}

func TestFear(t *testing.T) {
	if got := Fear(30, false); got != 0 {
		t.Errorf("fear with the hunter far away = %v, want 0", got)
	}
	if near, mid := Fear(2, false), Fear(8, false); near <= mid || mid <= 0 {
		t.Errorf("fear should rise as the hunter closes: near=%v mid=%v", near, mid)
	}
	if got := Fear(30, true); got < 0.75 {
		t.Errorf("fear during a chase = %v, want at least 0.75", got)
	}
	if got := Fear(-5, true); got != 1 {
		t.Errorf("fear clamps at 1, got %v", got)
	}
}

func TestHeartbeatQuickensWithFear(t *testing.T) {
	if _, ok := HeartbeatInterval(0.1); ok {
		t.Error("a calm heart should stay silent")
	}
	calm, ok1 := HeartbeatInterval(0.2)
	panic, ok2 := HeartbeatInterval(1)
	if !ok1 || !ok2 || panic >= calm {
		t.Errorf("beat should quicken with fear: calm=%v panic=%v", calm, panic)
	}
	if math.Abs(panic-fastestBeat) > 1e-9 {
		t.Errorf("full fear beats every %v, want %v", panic, fastestBeat)
	}
}

func TestGrowlInterval(t *testing.T) {
	if _, ok := GrowlInterval(growlRange + 1); ok {
		t.Error("a distant hunter should not be heard breathing")
	}
	near, _ := GrowlInterval(1)
	far, ok := GrowlInterval(growlRange)
	if !ok || near >= far {
		t.Errorf("breaths should come quicker up close: near=%v far=%v", near, far)
	}
}
