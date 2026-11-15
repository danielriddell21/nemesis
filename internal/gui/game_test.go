package gui

import (
	"os"
	"testing"

	"github.com/danielriddell21/nemesis/internal/sim"
	"github.com/danielriddell21/nemesis/internal/telemetry"
	"github.com/danielriddell21/nemesis/internal/world"
)

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}

func testSim(t *testing.T) *sim.Game {
	t.Helper()
	l, err := world.Generate(world.Config{Width: 32, Height: 24, Seed: 9})
	if err != nil {
		t.Fatal(err)
	}
	return sim.New(l)
}

func TestNoticeFiltersByEarshot(t *testing.T) {
	s := testSim(t)
	near := telemetry.Event{Type: "alien-creak", X: s.Player.Pos.Cell().X + 2, Y: s.Player.Pos.Cell().Y}
	if _, _, ok := notice(near, s); !ok {
		t.Error("a creak two tiles away should produce a notice")
	}
	far := telemetry.Event{Type: "alien-creak", X: s.Player.Pos.Cell().X + 25, Y: s.Player.Pos.Cell().Y + 20}
	if _, _, ok := notice(far, s); ok {
		t.Error("a creak across the map should stay silent")
	}
	if _, _, ok := notice(telemetry.Event{Type: "director-nudge"}, s); ok {
		t.Error("director internals must never reach the player's HUD")
	}
	if text, _, ok := notice(telemetry.Event{Type: "alien-seen", X: 1, Y: 1}, s); !ok || text == "" {
		t.Error("being spotted should warn the player")
	}
}

func TestTrySendNeverBlocks(t *testing.T) {
	ch := make(chan Msg, 1)
	trySend(ch, Msg{Type: "state"})
	trySend(ch, Msg{Type: "state"}) // full: must drop, not block
	if len(ch) != 1 {
		t.Fatalf("channel holds %d messages, want 1", len(ch))
	}
}

func TestNoticeCoversTheTorch(t *testing.T) {
	s := testSim(t)
	here := s.Player.Pos.Cell()
	if text, _, ok := notice(telemetry.Event{Type: "battery", X: here.X, Y: here.Y}, s); !ok || text == "" {
		t.Error("picking up a cell should be announced")
	}
	if text, _, ok := notice(telemetry.Event{Type: "alien-glow", X: 1, Y: 1}, s); !ok || text == "" {
		t.Error("the hunter catching the torchlight should warn the player")
	}
	if _, _, ok := notice(telemetry.Event{Type: "torch", X: here.X, Y: here.Y}, s); ok {
		t.Error("an ordinary flick of the switch needs no caption")
	}
	s.Player.Battery = 0
	if text, _, ok := notice(telemetry.Event{Type: "torch", X: here.X, Y: here.Y}, s); !ok || text == "" {
		t.Error("a cell giving out should be announced")
	}
}

func TestTorchFailingWarnsOncePerCell(t *testing.T) {
	s := testSim(t)
	s.Player.Torch = true
	s.Player.Battery = 0.5
	warned, post := torchFailing(s, false)
	if warned || post {
		t.Error("a healthy cell should not warn")
	}
	s.Player.Battery = 0.1
	warned, post = torchFailing(s, warned)
	if !warned || !post {
		t.Error("a cell in its last fifth should warn")
	}
	if _, post = torchFailing(s, warned); post {
		t.Error("the warning should post once, not every tick")
	}
	s.Player.Battery = 0.6
	warned, _ = torchFailing(s, warned)
	s.Player.Battery = 0.1
	if _, post = torchFailing(s, warned); !post {
		t.Error("a fresh cell should re-arm the warning")
	}
	s.Player.Torch = false
	if _, post = torchFailing(s, false); post {
		t.Error("a torch that is off should not warn")
	}
}

func TestStepTapFeedsFootstepsToTheAudioQueue(t *testing.T) {
	g := &Game{}
	tap := stepTap{g}
	tap.Observe(sim.Observation{Kind: sim.ObsStep, At: world.Coord{X: 3, Y: 4}})
	tap.Observe(sim.Observation{Kind: sim.ObsDoorOpen})
	if len(g.steps) != 1 || g.steps[0].X != 3 || g.steps[0].Y != 4 {
		t.Errorf("steps = %+v, want the one footstep at 3,4", g.steps)
	}
	if kind, ok := g.steps[0].Kind(); !ok || kind != sim.ObsStep {
		t.Errorf("queued event kind = %v,%v, want a step", kind, ok)
	}
}
