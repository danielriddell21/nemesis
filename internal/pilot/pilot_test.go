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
	path := bfs(g, src, l.Exit, nil)
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
	if path := bfs(g, src, src, nil); path != nil {
		t.Errorf("bfs to the same cell = %v, want nil", path)
	}
	if path := bfs(g, src, world.Coord{X: -5, Y: -5}, nil); path != nil {
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
		{"hunted but out of sight goes quiet", 40, sim.StateHunt, sim.ModeSneak},
		{"close sneaks", 5, sim.StateLurk, sim.ModeSneak},
		{"within earshot of a walk sneaks", 5, sim.StatePatrol, sim.ModeSneak},
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
	turning := func() *Pilot { return &Pilot{startedFleeing: true} }

	placeAlien(g, 5, sim.StateSearch)
	if New().wantThrow(g) {
		t.Error("threw without turning to flee")
	}
	placeAlien(g, 2, sim.StateSearch)
	if turning().wantThrow(g) {
		t.Error("threw with the hunter on top of us")
	}
	placeAlien(g, 5, sim.StateLurk)
	if turning().wantThrow(g) {
		t.Error("threw at a sleeping hunter")
	}
	placeAlien(g, 5, sim.StateHunt)
	if turning().wantThrow(g) {
		t.Error("threw at a hunter in full chase")
	}

	placeAlien(g, 5, sim.StateSearch)
	p := turning()
	if !p.wantThrow(g) {
		t.Fatal("did not drop a decoy on turning to flee")
	}
	if p.wantThrow(g) {
		t.Error("threw again during the cooldown")
	}

	g.Player.Decoys = 0
	if turning().wantThrow(g) {
		t.Error("threw with no decoys left")
	}
}

func TestDestinationFlagsTheTurnToFlee(t *testing.T) {
	g := newGame(t, 42)
	p := New()
	placeAlien(g, 4, sim.StateSearch)
	if _, ok := p.refuge(g); !ok {
		t.Skip("nowhere to flee on this layout")
	}
	p.destination(g)
	if !p.fleeing || !p.startedFleeing {
		t.Error("first flee tick should flag the turn")
	}
	p.destination(g)
	if p.startedFleeing {
		t.Error("the turn should be flagged for one tick only")
	}
	placeAlien(g, 30, sim.StatePatrol)
	p.destination(g)
	if p.fleeing {
		t.Error("still fleeing with the hunter far away")
	}
	placeAlien(g, 4, sim.StateSearch)
	p.destination(g)
	if !p.startedFleeing {
		t.Error("a fresh flee after breaking away should flag the turn again")
	}
}

func TestThreatened(t *testing.T) {
	g := newGame(t, 42)
	p := New()
	placeAlien(g, 3, sim.StateLurk)
	if p.threatened(g) {
		t.Error("a sleeping hunter is no threat")
	}
	placeAlien(g, 3, sim.StatePatrol)
	if !p.threatened(g) {
		t.Error("a prowling hunter three tiles off is a threat")
	}
	placeAlien(g, 12, sim.StatePatrol)
	if p.threatened(g) {
		t.Error("a patrol twelve tiles off is not a threat")
	}
	placeAlien(g, 12, sim.StateInvestigate)
	g.Alien.Target = g.Player.Pos.Cell()
	if !p.threatened(g) {
		t.Error("a hunter coming to look at our cell is a threat")
	}
	g.Alien.Target = world.Coord{X: -40, Y: -40}
	if p.threatened(g) {
		t.Error("a hunter investigating elsewhere is not a threat")
	}
}

func TestRefugeIsFarHiddenAndNeverAVent(t *testing.T) {
	g := newGame(t, 42)
	placeAlien(g, 4, sim.StateSearch)
	g.Alien.Target = g.Player.Pos.Cell()
	c, ok := New().refuge(g)
	if !ok {
		t.Skip("nowhere to flee on this layout")
	}
	if d := cellMid(c).Sub(g.Alien.Pos).Len(); d < 8 {
		t.Errorf("refuge %v is only %.1f from the hunter", c, d)
	}
	if g.Exposed(c) {
		t.Errorf("refuge %v is in the hunter's sight", c)
	}
	if g.World.Level.At(c.X, c.Y) == world.TileVent {
		t.Errorf("refuge %v is a vent", c)
	}
}

func TestTrackerRaised(t *testing.T) {
	g := newGame(t, 42)

	placeAlien(g, 3, sim.StateLurk)
	if New().trackerRaised(g, 0.1) {
		t.Error("tracker raised where the hunter can hear the ping")
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

func TestObjectiveIsTheNearestDarkConsole(t *testing.T) {
	g := newGame(t, 42)
	g.Alien.Pos = sim.Vec2{X: -50, Y: -50}
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

func TestNextCellAtTheConsoleAccessCellAimsAtTheConsole(t *testing.T) {
	g := newGame(t, 42)
	console := g.World.Level.Consoles[0]
	access := walkTarget(g, console)
	g.Player.Pos = cellMid(access)
	got, ok := New().nextCell(g, console)
	if !ok || got != console {
		t.Errorf("nextCell at the access cell = %v, %v; want %v, true", got, ok, console)
	}
}

func TestPilotClosesInUntilTheConsoleIsInReach(t *testing.T) {
	g := newGame(t, 42)
	console := g.World.Level.Consoles[0]
	access := walkTarget(g, console)
	// The far edge of the access cell: beside the console, but out of reach.
	away := cellMid(access).Sub(cellMid(console))
	g.Player.Pos = sim.Vec2{X: cellMid(access).X + away.X*0.45, Y: cellMid(access).Y + away.Y*0.45}
	g.Player.Angle = math.Atan2(-away.Y, -away.X)
	g.Alien.Pos = sim.Vec2{X: -50, Y: -50}
	p := New()
	const dt = 1.0 / 60
	for i := 0; i < 5*60 && !g.ConsoleActivated(console); i++ {
		g.Tick(p.Input(g, dt), dt)
		g.Alien.Pos = sim.Vec2{X: -50, Y: -50}
		g.Alien.State = sim.StateLurk
	}
	if !g.ConsoleActivated(console) {
		t.Errorf("pilot at %v never lit the console at %v", g.Player.Pos, console)
	}
}

func TestObjectiveStaysCommittedUntilTheConsoleIsLit(t *testing.T) {
	g := newGame(t, 42)
	p := New()
	first, ok := p.objective(g)
	if !ok {
		t.Fatal("no objective")
	}
	// Stand beside a different console: the commitment must not flip.
	for _, c := range g.World.Level.Consoles {
		if c != first {
			g.Player.Pos = cellMid(walkTarget(g, c))
		}
	}
	if got, _ := p.objective(g); got != first {
		t.Errorf("objective flipped from %v to %v before the console was lit", first, got)
	}
}

func TestWantTorchGoesDarkNearTheHunter(t *testing.T) {
	g := newGame(t, 42)
	// The switch is pulsed, so a wanted press lands within two ticks.
	press := func(p *Pilot) bool {
		first := p.wantTorch(g)
		second := p.wantTorch(g)
		return first || second
	}

	placeAlien(g, 25, sim.StatePatrol)
	if !press(New()) {
		t.Error("did not switch the torch on with the hunter far away")
	}
	g.Player.Torch = true
	if press(New()) {
		t.Error("pressed the switch with the torch already on")
	}
	placeAlien(g, 6, sim.StatePatrol)
	if !press(New()) {
		t.Error("did not kill the torch with the hunter close")
	}
	g.Player.Torch = false
	g.Player.Battery = 0
	placeAlien(g, 25, sim.StatePatrol)
	if press(New()) {
		t.Error("pressed the switch on a flat battery")
	}
}

func TestBFSRoutesAroundAvoidedCells(t *testing.T) {
	g := newGame(t, 42)
	l := g.World.Level
	src := g.Player.Pos.Cell()
	direct := bfs(g, src, l.Exit, nil)
	if len(direct) < 3 {
		t.Fatalf("direct path too short to block: %v", direct)
	}
	blocked := direct[len(direct)/2]
	path := bfs(g, src, l.Exit, func(c world.Coord) bool { return c == blocked })
	for _, c := range path {
		if c == blocked {
			t.Fatalf("path crosses the avoided cell %v", blocked)
		}
	}
	if all := bfs(g, src, l.Exit, func(world.Coord) bool { return true }); all != nil {
		t.Errorf("found a path %v with every cell avoided", all)
	}
}

func TestDestinationBacksOffWhenTheHunterBlocksTheOnlyWay(t *testing.T) {
	g := newGame(t, 42)
	p := New()
	goal, _ := p.objective(g)
	// The hunter sits on the access cell: every route ends inside its reach.
	g.Alien.Pos = cellMid(walkTarget(g, goal))
	g.Alien.State = sim.StatePatrol
	p.goal, p.hasGoal = goal, true
	for _, c := range g.World.Level.Consoles {
		if c != goal {
			p.goal = goal
		}
	}
	if got, ok := p.destination(g); ok && walkTarget(g, got) == walkTarget(g, goal) {
		t.Error("walked toward a hunter sitting on the objective")
	}
	if p.safeRoute(g, goal) {
		t.Error("a route ending under the hunter was called safe")
	}
}

func TestObjectivePrefersAConsoleTheHunterIsNotSittingOn(t *testing.T) {
	g := newGame(t, 42)
	p := New()
	nearest, _ := p.objective(g)
	g.Alien.Pos = cellMid(nearest)
	got, ok := New().objective(g)
	if !ok || got == nearest {
		t.Errorf("objective = %v, want a console away from the hunter at %v", got, nearest)
	}
	// Already committed and far away: the commitment is dropped too.
	if again, _ := p.objective(g); again == nearest {
		t.Errorf("stayed committed to %v with the hunter on it", nearest)
	}
}

func TestRouteKeepsOutOfTheDuctsWhenItCan(t *testing.T) {
	for seed := int64(1); seed <= 25; seed++ {
		g := newGame(t, seed)
		g.Alien.Pos = sim.Vec2{X: -50, Y: -50}
		l := g.World.Level
		path := route(g, g.Player.Pos.Cell(), l.Exit)
		if len(path) == 0 {
			t.Fatalf("seed %d: no route to the exit", seed)
		}
		noVents := func(c world.Coord) bool { return l.At(c.X, c.Y) == world.TileVent }
		if len(bfs(g, g.Player.Pos.Cell(), l.Exit, noVents)) == 0 {
			continue
		}
		for _, c := range path {
			if l.At(c.X, c.Y) == world.TileVent {
				t.Errorf("seed %d: route crawls through the vent at %v though a floor route exists", seed, c)
				break
			}
		}
	}
}

func TestReplanKeepsAPlanUntilTheCellChanges(t *testing.T) {
	g := newGame(t, 42)
	g.Alien.Pos = sim.Vec2{X: -50, Y: -50}
	p := New()
	target, _, ok := p.replan(g)
	if !ok {
		t.Fatal("no plan from the spawn")
	}
	if p.plan.age != replanTicks {
		t.Fatalf("fresh plan age = %d, want %d", p.plan.age, replanTicks)
	}
	// A change of heart inside the same cell waits for the plan to age out.
	p.hasGoal = false
	g.Alien.Pos = cellMid(target)
	if again, _, _ := p.replan(g); again != target || p.plan.age != replanTicks-1 {
		t.Errorf("plan rethought early: target %v -> %v, age %d", target, again, p.plan.age)
	}
	// Stepping into a new cell rethinks at once.
	g.Player.Pos = cellMid(p.plan.path[0])
	if p.replan(g); p.plan.age != replanTicks || p.plan.from != g.Player.Pos.Cell() {
		t.Errorf("no fresh plan after stepping: age %d from %v", p.plan.age, p.plan.from)
	}
	for range replanTicks + 1 {
		p.replan(g)
	}
	if p.plan.age != replanTicks {
		t.Errorf("plan did not age out and renew: age %d", p.plan.age)
	}
}
