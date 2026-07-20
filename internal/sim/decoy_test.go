package sim

import (
	"testing"

	"github.com/danielriddell21/nemesis/internal/world"
)

func TestThrowDecoyConsumesInventory(t *testing.T) {
	g := New(flatLevel(24, 12))
	if g.Player.Decoys != startingDecoys {
		t.Fatalf("start decoys = %d, want %d", g.Player.Decoys, startingDecoys)
	}
	g.Player.Angle = 0
	g.Tick(Input{Throw: true}, 1.0/60)
	if g.Player.Decoys != startingDecoys-1 {
		t.Errorf("throwing should consume a decoy, got %d", g.Player.Decoys)
	}
	if !g.DecoyState().Active {
		t.Error("a decoy should be active after throwing")
	}
	// Holding throw does not spam more decoys, and a second active one is barred.
	g.Tick(Input{Throw: true}, 1.0/60)
	if g.Player.Decoys != startingDecoys-1 {
		t.Error("a held throw or an already-active decoy should not consume more")
	}
}

func TestDecoyLuresThenIsFound(t *testing.T) {
	rec := &recorder{}
	g := New(flatLevel(40, 12), WithObserver(rec))
	g.Alien.Pos = Vec2{X: 20.5, Y: 1.5}
	g.Alien.Facing = 0
	g.Alien.State = StatePatrol
	g.Alien.dwell = 2 // scanning in place, within earshot of where the decoy lands
	g.Player.Pos = Vec2{X: 14.5, Y: 1.5}
	g.Player.Angle = 0 // throws east, toward the alien
	g.Tick(Input{Throw: true}, 1.0/60)
	// Slip away to the far end, out of the hunter's sight, so the decoy is the
	// only thing drawing it.
	g.Player.Pos = Vec2{X: 2.5, Y: 1.5}
	stepN(g, Input{}, 60) // let it land and chirp
	if rec.count(ObsDecoy) == 0 {
		t.Fatal("a landed decoy should chirp")
	}
	stepN(g, Input{}, 20*60)
	if g.Learned().Decoys == 0 {
		t.Error("the hunter should reach the decoy and learn from it")
	}
}

func TestDecoyIgnoredWhenLearned(t *testing.T) {
	rec := &recorder{}
	g := New(flatLevel(40, 12), WithObserver(rec), WithLearned(Learned{Decoys: decoyTier2}))
	g.Alien.Pos = Vec2{X: 20.5, Y: 1.5}
	g.Alien.Facing = 0
	g.Alien.State = StatePatrol
	g.Alien.dwell = 5 // parked within earshot: only its deafness to decoys matters
	g.Player.Pos = Vec2{X: 14.5, Y: 1.5}
	g.Player.Angle = 0
	g.Tick(Input{Throw: true}, 1.0/60)
	g.Player.Pos = Vec2{X: 2.5, Y: 1.5}
	stepN(g, Input{}, 3*60)
	if g.Alien.State == StateInvestigate {
		t.Error("a decoy-wise hunter should not be lured by the chirp")
	}
}

func TestDecoyStopsAtWall(t *testing.T) {
	g := New(flatLevel(16, 12))
	g.Player.Pos = Vec2{X: 13.5, Y: 1.5}
	g.Player.Angle = 0 // toward the east wall
	g.Tick(Input{Throw: true}, 1.0/60)
	stepN(g, Input{}, 30)
	d := g.DecoyState()
	if !d.Landed {
		t.Fatal("decoy should land")
	}
	if d.Pos.X > 15 {
		t.Errorf("decoy passed through the wall to x=%v", d.Pos.X)
	}
}

func TestHiddenPlayerCannotThrow(t *testing.T) {
	l := flatLevel(16, 12)
	l.Tiles[1*16+3] = world.TileLocker
	l.Lockers = []world.Coord{{X: 3, Y: 1}}
	g := New(l)
	g.Player.Pos = Vec2{X: 3.5, Y: 1.5}
	g.Tick(Input{Hide: true}, 1.0/60)
	if !g.Player.Hidden {
		t.Fatal("player should be hidden")
	}
	before := g.Player.Decoys
	g.Tick(Input{Throw: true}, 1.0/60)
	if g.Player.Decoys != before {
		t.Error("a hidden player should not be able to throw")
	}
}
