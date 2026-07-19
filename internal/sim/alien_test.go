package sim

import (
	"math"
	"testing"

	"github.com/danielriddell21/nemesis/internal/world"
)

func TestAlienWakesAfterLurk(t *testing.T) {
	g := New(flatLevel(24, 16))
	g.Alien.Pos = Vec2{X: 20.5, Y: 12.5}
	stepN(g, Input{}, int(lurkTime*60)+120)
	if g.Alien.State == StateLurk {
		t.Error("alien should wake and patrol after the lurk period")
	}
}

func TestAlienOpensDoorsOnPath(t *testing.T) {
	l := flatLevel(24, 5)
	door := world.Coord{X: 12, Y: 2}
	for y := 1; y < 4; y++ {
		l.Tiles[y*24+12] = world.TileWall
	}
	l.Tiles[door.Y*24+door.X] = world.TileDoor
	rec := &recorder{}
	g := New(l, WithObserver(rec))
	g.Player.Pos = Vec2{X: 22.5, Y: 2.5} // out of the way beyond the door
	g.Alien.Pos = Vec2{X: 2.5, Y: 2.5}
	g.Alien.State = StatePatrol
	g.setAlienTarget(world.Coord{X: 20, Y: 2})
	stepN(g, Input{}, 20*60)
	if !g.World.DoorOpen(door) {
		t.Error("alien should shoulder doors open on its path")
	}
}

func TestAlienPrefersVents(t *testing.T) {
	// Two parallel routes to the target: a long floor detour and a straight
	// vent duct. The A* path must take the duct.
	l := flatLevel(20, 7)
	for x := 2; x < 18; x++ {
		l.Tiles[3*20+x] = world.TileVent
	}
	w := NewWorld(l)
	path := findPath(w, world.Coord{X: 1, Y: 3}, world.Coord{X: 18, Y: 3})
	if len(path) == 0 {
		t.Fatal("no path found")
	}
	vents := 0
	for _, c := range path {
		if l.At(c.X, c.Y) == world.TileVent {
			vents++
		}
	}
	if vents < 10 {
		t.Errorf("path used %d vent cells, want the duct route", vents)
	}
}

func TestFindPathUnreachable(t *testing.T) {
	l := flatLevel(16, 12)
	for y := 1; y < 11; y++ {
		l.Tiles[y*16+8] = world.TileWall
	}
	w := NewWorld(l)
	if p := findPath(w, world.Coord{X: 1, Y: 1}, world.Coord{X: 14, Y: 1}); p != nil {
		t.Errorf("walled-off target should have no path, got %v", p)
	}
	if p := findPath(w, world.Coord{X: 1, Y: 1}, world.Coord{X: 1, Y: 1}); p != nil {
		t.Errorf("path to self should be nil, got %v", p)
	}
}

func TestHuntFollowsThenSearches(t *testing.T) {
	g := New(flatLevel(24, 12))
	g.Alien.Pos = Vec2{X: 8.5, Y: 1.5}
	g.Alien.Facing = math.Pi
	stepN(g, Input{}, 5)
	if g.Alien.State != StateHunt {
		t.Fatal("alien should be hunting")
	}
	// Teleport the player far away and out of sight behind a wall (but not
	// onto the exit tile); the alien must fall back to the last known position
	// and then search.
	for y := 1; y < 11; y++ {
		g.World.Level.Tiles[y*24+16] = world.TileWall
	}
	g.Player.Pos = Vec2{X: 22.5, Y: 2.5}
	stepN(g, Input{}, 30*60)
	if g.Alien.State == StateHunt {
		t.Error("alien should give up a cold trail")
	}
}

func TestDirectorLeashesCamping(t *testing.T) {
	rec := &recorder{}
	g := New(flatLevel(48, 16), WithObserver(rec))
	// Park the alien next to the player in a non-hunt state and keep it there.
	g.Alien.Pos = Vec2{X: 4.5, Y: 1.5}
	g.Alien.Facing = 0 // facing away, never sees the player
	g.Alien.State = StateSearch
	g.Alien.searchesLeft = math.MaxInt32
	for range int((campLimit + 3) * 60) {
		g.Tick(Input{}, 1.0/60)
		// Pin the search around the player so the camp timer accumulates.
		if g.Alien.Pos.Sub(g.Player.Pos).Len() > campRange-1 {
			g.Alien.Pos = Vec2{X: 4.5, Y: 1.5}
		}
	}
	if rec.count(ObsDirectorNudge) == 0 {
		t.Error("director should leash a camping alien away")
	}
}

func TestDirectorNudgesTowardPlayer(t *testing.T) {
	rec := &recorder{}
	g := New(flatLevel(48, 16), WithObserver(rec))
	g.Alien.Pos = Vec2{X: 45.5, Y: 13.5}
	g.Alien.State = StatePatrol
	g.director.wake = 0
	stepN(g, Input{}, int(nudgeInterval*60)+240)
	if rec.count(ObsDirectorNudge) == 0 {
		t.Error("director should periodically nudge the alien toward the player")
	}
}

func TestConsoleEscalatesDirector(t *testing.T) {
	l := flatLevel(16, 12)
	console := world.Coord{X: 3, Y: 0}
	l.Tiles[console.Y*16+console.X] = world.TileConsole
	l.Consoles = []world.Coord{console}
	g := New(l)
	g.Alien.Pos = Vec2{X: 14.5, Y: 10.5}
	before := g.director.aggression
	g.Player.Pos = Vec2{X: 3.5, Y: 1.5}
	g.Player.Angle = -math.Pi / 2
	g.Tick(Input{Use: true}, 1.0/60)
	if g.director.aggression <= before {
		t.Error("activating a console should raise the director's aggression")
	}
	if g.Alien.State != StateInvestigate {
		t.Errorf("console noise should send the alien investigating, got %v", g.Alien.State)
	}
}

func TestAlienStateStrings(t *testing.T) {
	states := map[AlienState]string{
		StateLurk:        "lurk",
		StatePatrol:      "patrol",
		StateInvestigate: "investigate",
		StateHunt:        "hunt",
		StateSearch:      "search",
		AlienState(99):   "unknown",
	}
	for s, want := range states {
		if got := s.String(); got != want {
			t.Errorf("state %d = %q, want %q", s, got, want)
		}
	}
}
