package vis

import (
	"testing"

	"github.com/danielriddell21/nemesis/internal/sim"
	"github.com/danielriddell21/nemesis/internal/world"
)

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
	snap := Snapshot(s)
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
	snap := Snapshot(s)
	if snap.PingTier != 2 || snap.VentTier != 2 || snap.SearchTier != 2 {
		t.Errorf("tiers = %d/%d/%d, want 2/2/2", snap.PingTier, snap.VentTier, snap.SearchTier)
	}
}
