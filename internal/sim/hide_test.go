package sim

import (
	"math"
	"testing"

	"github.com/danielriddell21/nemesis/internal/world"
)

func lockerLevel() *world.Level {
	l := flatLevel(24, 12)
	l.Tiles[6*24+6] = world.TileLocker
	l.Lockers = []world.Coord{{X: 6, Y: 6}}
	l.Rooms = []world.Room{{X: 1, Y: 1, W: 22, H: 10}}
	return l
}

func TestHideOnlyOnLocker(t *testing.T) {
	g := New(lockerLevel())
	g.Player.Pos = Vec2{X: 2.5, Y: 2.5} // plain floor
	g.Tick(Input{Hide: true}, 1.0/60)
	if g.Player.Hidden {
		t.Error("cannot hide on open floor")
	}
	g.Tick(Input{Hide: false}, 1.0/60)  // release the button
	g.Player.Pos = Vec2{X: 6.5, Y: 6.5} // on the locker
	g.Tick(Input{Hide: true}, 1.0/60)
	if !g.Player.Hidden {
		t.Error("should hide when standing on a locker")
	}
	// A second press (after releasing) unhides.
	g.Tick(Input{Hide: false}, 1.0/60)
	g.Tick(Input{Hide: true}, 1.0/60)
	if g.Player.Hidden {
		t.Error("pressing hide again should leave the locker")
	}
}

func TestHiddenPlayerUnseenButMoveless(t *testing.T) {
	g := New(lockerLevel())
	g.Player.Pos = Vec2{X: 6.5, Y: 6.5}
	g.Tick(Input{Hide: true}, 1.0/60)
	// Alien staring straight at the locker from close range: hidden means unseen.
	g.Alien.Pos = Vec2{X: 10.5, Y: 6.5}
	g.Alien.Facing = math.Pi
	stepN(g, Input{}, 10)
	if g.Alien.State == StateHunt {
		t.Error("a hidden player should not be seen at range")
	}
	// Movement input does nothing while hidden.
	before := g.Player.Pos
	stepN(g, Input{Forward: 1}, 30)
	if g.Player.Pos != before {
		t.Error("a hidden player should not move")
	}
}

func TestHunterBreachesLocker(t *testing.T) {
	g := New(lockerLevel())
	g.Player.Pos = Vec2{X: 6.5, Y: 6.5}
	g.Tick(Input{Hide: true}, 1.0/60)
	// Alien right on top of the locker: it wrenches the door open.
	g.Alien.Pos = Vec2{X: 6.5, Y: 6.5}
	g.Tick(Input{}, 1.0/60)
	if !g.Dead() {
		t.Error("a hunter reaching the locker should breach it")
	}
}

func TestHunterLearnsLockersOnHide(t *testing.T) {
	rec := &recorder{}
	g := New(lockerLevel(), WithObserver(rec))
	// Hunter searching nearby when the prey ducks in.
	g.Alien.Pos = Vec2{X: 9.5, Y: 6.5}
	g.Alien.State = StateSearch
	g.Player.Pos = Vec2{X: 6.5, Y: 6.5}
	g.Tick(Input{Hide: true}, 1.0/60)
	if g.Learned().Lockers == 0 {
		t.Error("ducking into a locker under the hunter's nose should teach it")
	}
}

func TestLockerAwareHunterChecksLockers(t *testing.T) {
	// A hunter that has learned lockers detours to one while searching.
	g := New(lockerLevel(), WithLearned(Learned{Lockers: lockerTier1}))
	g.Alien.Pos = Vec2{X: 14.5, Y: 6.5}
	g.Alien.lastKnown = Vec2{X: 14.5, Y: 6.5}
	g.Alien.State = StateSearch
	g.Alien.searchesLeft = 3
	g.Alien.ventChecked = true // skip vent step
	target := g.nextSearchTarget()
	if target != (world.Coord{X: 6, Y: 6}) {
		t.Errorf("locker-aware search target = %v, want the locker at 6,6", target)
	}
}

func TestNaiveHunterSkipsLockers(t *testing.T) {
	g := New(lockerLevel())
	g.Alien.ventChecked = true
	if _, ok := g.lockerCheckPoint(); ok {
		t.Error("a naive hunter should not check lockers")
	}
}
