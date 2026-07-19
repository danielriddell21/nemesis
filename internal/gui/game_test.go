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

func TestSnapshotMirrorsSim(t *testing.T) {
	s := testSim(t)
	for range 30 {
		s.Tick(sim.Input{Forward: 1}, 1.0/60)
	}
	snap := snapshot(s)
	if snap.PlayerX != s.Player.Pos.X || snap.PlayerY != s.Player.Pos.Y {
		t.Error("snapshot player position out of sync")
	}
	if snap.AlienState != s.Alien.State.String() {
		t.Error("snapshot alien state out of sync")
	}
	if snap.Total != s.ObjectivesTotal() || snap.Done != s.ObjectivesDone() {
		t.Error("snapshot objectives out of sync")
	}
}

func TestSnapshotCarriesLearning(t *testing.T) {
	l, err := world.Generate(world.Config{Width: 32, Height: 24, Seed: 9})
	if err != nil {
		t.Fatal(err)
	}
	s := sim.New(l, sim.WithLearned(sim.Learned{Pings: 9, Creaks: 11, Cold: 2}))
	snap := snapshot(s)
	if snap.PingTier != 2 || snap.VentTier != 2 || snap.SearchTier != 2 {
		t.Errorf("tiers = %d/%d/%d, want 2/2/2", snap.PingTier, snap.VentTier, snap.SearchTier)
	}
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
