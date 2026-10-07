package pilot

import (
	"math"
	"testing"

	"github.com/danielriddell21/nemesis/internal/sim"
	"github.com/danielriddell21/nemesis/internal/world"
)

func newGame(t *testing.T, seed int64) *sim.Game {
	t.Helper()
	l, err := world.Generate(world.Config{Width: 32, Height: 24, Seed: seed, Consoles: 2})
	if err != nil {
		t.Fatal(err)
	}
	return sim.New(l)
}

// placeAlien puts the hunter at distance d east of the player, in state s.
func placeAlien(g *sim.Game, d float64, s sim.AlienState) {
	g.Alien.Pos = sim.Vec2{X: g.Player.Pos.X + d, Y: g.Player.Pos.Y}
	g.Alien.State = s
}

func TestAngleDiffWrapsToTheShortWay(t *testing.T) {
	cases := []struct{ a, b, want float64 }{
		{0, 0, 0},
		{1, 0.5, 0.5},
		{0.5, 1, -0.5},
		{math.Pi - 0.1, -math.Pi + 0.1, -0.2},
		{-math.Pi + 0.1, math.Pi - 0.1, 0.2},
		{5 * math.Pi, 0, math.Pi},
	}
	for _, c := range cases {
		if got := angleDiff(c.a, c.b); math.Abs(got-c.want) > 1e-9 {
			t.Errorf("angleDiff(%v, %v) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestClamp(t *testing.T) {
	for _, c := range []struct{ v, want float64 }{{-2, -1}, {-1, -1}, {0.5, 0.5}, {1, 1}, {3, 1}} {
		if got := clamp(c.v, -1, 1); got != c.want {
			t.Errorf("clamp(%v) = %v, want %v", c.v, got, c.want)
		}
	}
}

func TestManhattanAndCellMid(t *testing.T) {
	if got := manhattan(world.Coord{X: 1, Y: 5}, world.Coord{X: 4, Y: 1}); got != 7 {
		t.Errorf("manhattan = %d, want 7", got)
	}
	if got := manhattan(world.Coord{X: 4, Y: 1}, world.Coord{X: 1, Y: 5}); got != 7 {
		t.Errorf("manhattan reversed = %d, want 7", got)
	}
	if got := cellMid(world.Coord{X: 2, Y: 3}); got != (sim.Vec2{X: 2.5, Y: 3.5}) {
		t.Errorf("cellMid = %v, want {2.5 3.5}", got)
	}
}

func TestBFSFindsAWalkablePathToTheExit(t *testing.T) {
	g := newGame(t, 42)
	l := g.World.Level
	src := g.Player.Pos.Cell()
	path := bfs(g, src, l.Exit)
	if len(path) == 0 {
		t.Fatal("no path from spawn to the exit")
	}
	if path[len(path)-1] != l.Exit {
		t.Errorf("path ends at %v, want the exit %v", path[len(path)-1], l.Exit)
	}
	prev := src
	for _, c := range path {
		if manhattan(prev, c) != 1 {
			t.Fatalf("path jumps from %v to %v", prev, c)
		}
		if !l.At(c.X, c.Y).Walkable() {
			t.Fatalf("path crosses %v, which is not walkable", c)
		}
		prev = c
	}
}

func TestBFSWithNothingToDo(t *testing.T) {
	g := newGame(t, 42)
	src := g.Player.Pos.Cell()
	if path := bfs(g, src, src); path != nil {
		t.Errorf("bfs to the same cell = %v, want nil", path)
	}
	if path := bfs(g, src, world.Coord{X: -5, Y: -5}); path != nil {
		t.Errorf("bfs off the map = %v, want nil", path)
	}
}

func TestWalkTargetStepsInFrontOfAConsole(t *testing.T) {
	g := newGame(t, 42)
	l := g.World.Level
	if len(l.Consoles) == 0 {
		t.Fatal("generated deck has no consoles")
	}
	c := l.Consoles[0]
	w := walkTarget(g, c)
	if manhattan(w, c) != 1 || !l.At(w.X, w.Y).Walkable() {
		t.Errorf("walkTarget(console %v) = %v, want a walkable neighbour", c, w)
	}
	if got := walkTarget(g, l.Exit); got != l.Exit {
		t.Errorf("walkTarget(exit) = %v, want the exit itself", got)
	}
}

func TestGait(t *testing.T) {
	cases := []struct {
		name  string
		d     float64
		state sim.AlienState
		want  sim.MoveMode
	}{
		{"hunted runs even when close", 2, sim.StateHunt, sim.ModeRun},
		{"close sneaks", 5, sim.StateLurk, sim.ModeSneak},
		{"far runs", 20, sim.StateLurk, sim.ModeRun},
		{"between walks", 10, sim.StateLurk, sim.ModeWalk},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			g := newGame(t, 42)
			placeAlien(g, c.d, c.state)
			if got := New().gait(g); got != c.want {
				t.Errorf("gait = %v, want %v", got, c.want)
			}
		})
	}
}

func TestWantThrow(t *testing.T) {
	g := newGame(t, 42)
	g.Player.Decoys = 2

	placeAlien(g, 3, sim.StateHunt)
	p := New()
	if p.wantThrow(g) {
		t.Error("threw with the hunter on top of us")
	}
	placeAlien(g, 20, sim.StateHunt)
	if p.wantThrow(g) {
		t.Error("threw with the hunter out of earshot")
	}

	placeAlien(g, 9, sim.StateHunt)
	if !p.wantThrow(g) {
		t.Fatal("did not throw with the hunter in earshot")
	}
	if p.wantThrow(g) {
		t.Error("threw again during the cooldown")
	}

	g.Player.Decoys = 0
	if New().wantThrow(g) {
		t.Error("threw with no decoys left")
	}
}

func TestTrackerRaised(t *testing.T) {
	g := newGame(t, 42)

	placeAlien(g, 3, sim.StateLurk)
	if !New().trackerRaised(g, 0.1) {
		t.Error("tracker lowered with the hunter close")
	}

	placeAlien(g, 20, sim.StateLurk)
	if New().trackerRaised(g, 0.1) {
		t.Error("tracker raised with the hunter far away")
	}

	placeAlien(g, 9, sim.StateLurk)
	p := New()
	if !p.trackerRaised(g, 1) {
		t.Error("tracker lowered early in the cadence")
	}
	if p.trackerRaised(g, 3) {
		t.Error("tracker raised late in the cadence")
	}
}

func TestWantHide(t *testing.T) {
	g := newGame(t, 42)
	l := g.World.Level
	if len(l.Lockers) == 0 {
		t.Skip("generated deck has no lockers")
	}
	g.Player.Pos = cellMid(l.Lockers[0])

	placeAlien(g, 4, sim.StateHunt)
	if !New().wantHide(g) {
		t.Error("did not hide on a locker with the hunter close")
	}
	placeAlien(g, 4, sim.StateLurk)
	if New().wantHide(g) {
		t.Error("hid while the hunter was not hunting")
	}

	g.Player.Hidden = true
	placeAlien(g, 9, sim.StateHunt)
	if New().wantHide(g) {
		t.Error("left the locker while the hunter lingered")
	}
	placeAlien(g, 15, sim.StateHunt)
	if !New().wantHide(g) {
		t.Error("stayed hidden after the hunter wandered off")
	}
}

func TestFleeTargetOnlyWhenPressed(t *testing.T) {
	g := newGame(t, 42)
	placeAlien(g, 12, sim.StateHunt)
	if _, ok := New().fleeTarget(g); ok {
		t.Error("fled from a distant hunter")
	}
	placeAlien(g, 3, sim.StateLurk)
	if _, ok := New().fleeTarget(g); ok {
		t.Error("fled from a lurking hunter")
	}
	placeAlien(g, 3, sim.StateHunt)
	c, ok := New().fleeTarget(g)
	if !ok {
		t.Fatal("did not flee a close hunter")
	}
	if d := cellMid(c).Sub(g.Alien.Pos).Len(); d < 8 {
		t.Errorf("fled to %v, %.1f from the hunter, want at least 8", c, d)
	}
}

func TestObjectiveIsTheNearestDarkConsole(t *testing.T) {
	g := newGame(t, 42)
	l := g.World.Level
	got, ok := New().objective(g)
	if !ok {
		t.Fatal("no objective at the start of a run")
	}
	best := math.MaxFloat64
	for _, c := range l.Consoles {
		best = math.Min(best, cellMid(c).Sub(g.Player.Pos).Len())
	}
	if d := cellMid(got).Sub(g.Player.Pos).Len(); d != best {
		t.Errorf("objective %v is %.2f away, the nearest console is %.2f", got, d, best)
	}
}

func TestPilotMakesProgress(t *testing.T) {
	g := newGame(t, 42)
	p := New()
	start := g.Player.Pos
	const dt = 1.0 / 30
	for i := 0; i < 30*60 && !g.Dead() && !g.Escaped(); i++ {
		g.Tick(p.Input(g, dt), dt)
	}
	activated := 0
	for _, c := range g.World.Level.Consoles {
		if g.ConsoleActivated(c) {
			activated++
		}
	}
	if activated == 0 && !g.Escaped() && g.Player.Pos.Sub(start).Len() < 3 {
		t.Errorf("a minute in, the pilot has lit no console and moved %.1f", g.Player.Pos.Sub(start).Len())
	}
}
