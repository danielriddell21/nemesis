package sim

import (
	"testing"
)

func TestTrackerPingsOnCadence(t *testing.T) {
	rec := &recorder{}
	g := New(flatLevel(32, 12), WithObserver(rec))
	stepN(g, Input{Tracker: true}, int(3.5*pingInterval*60))
	if got := rec.count(ObsTrackerPing); got < 3 || got > 4 {
		t.Errorf("got %d pings over 3.5 intervals, want 3 or 4", got)
	}
	if !g.Tracker.Raised {
		t.Error("tracker should be raised")
	}
}

func TestTrackerBlipsOnlyOnMovement(t *testing.T) {
	g := New(flatLevel(32, 12))
	g.Alien.Pos = Vec2{X: 8.5, Y: 1.5}
	g.Alien.Moving = false
	g.Alien.State = StateLurk // stays put
	stepN(g, Input{Tracker: true}, int(2*pingInterval*60))
	if g.Tracker.Contact.Valid {
		t.Error("a motionless hunter should paint no blip")
	}
	// March the alien: the next ping must find it.
	g.Alien.State = StatePatrol
	g.setAlienTarget(g.World.Level.Exit)
	stepN(g, Input{Tracker: true}, int(2*pingInterval*60))
	if !g.Tracker.Contact.Valid {
		t.Error("a moving hunter in range should paint a blip")
	}
	if g.Tracker.Contact.Dist <= 0 {
		t.Error("blip should carry a distance")
	}
}

func TestTrackerLoweredStopsPings(t *testing.T) {
	rec := &recorder{}
	g := New(flatLevel(32, 12), WithObserver(rec))
	stepN(g, Input{}, int(3*pingInterval*60))
	if rec.count(ObsTrackerPing) != 0 {
		t.Error("lowered tracker should not ping")
	}
}

func TestTrackerPingIsAudible(t *testing.T) {
	rec := &recorder{}
	g := New(flatLevel(24, 12), WithObserver(rec))
	// Alien close enough to hear the ping, facing away so it cannot see.
	g.Alien.Pos = Vec2{X: 4.5, Y: 1.5}
	g.Alien.Facing = 0
	g.Alien.State = StatePatrol
	stepN(g, Input{Tracker: true}, int(1.5*pingInterval*60))
	if rec.count(ObsAlienHeard) == 0 {
		t.Error("a nearby hunter should hear the tracker ping")
	}
}
