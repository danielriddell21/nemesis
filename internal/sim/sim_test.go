package sim

import (
	"math"
	"testing"

	"github.com/danielriddell21/crucible/level"

	"github.com/danielriddell21/nemesis/internal/world"
)

// flatLevel builds a hand-made open arena with a wall border, spawn in the
// north-west and exit in the south-east.
func flatLevel(w, h int) *world.Level {
	base := level.New(w, h, 0)
	for y := range h {
		for x := range w {
			t := world.TileFloor
			if x == 0 || y == 0 || x == w-1 || y == h-1 {
				t = world.TileWall
			}
			base.Set(x, y, t)
			base.Light[base.Index(x, y)] = 0.6
		}
	}
	base.Spawn = world.Coord{X: 1, Y: 1}
	base.Exit = world.Coord{X: w - 2, Y: h - 2}
	base.Set(base.Spawn.X, base.Spawn.Y, world.TileSpawn)
	base.Set(base.Exit.X, base.Exit.Y, world.TileExit)
	return &world.Level{Level: base}
}

type recorder struct {
	events []Observation
}

func (r *recorder) Observe(o Observation) { r.events = append(r.events, o) }

func (r *recorder) count(k ObservationKind) int {
	n := 0
	for _, e := range r.events {
		if e.Kind == k {
			n++
		}
	}
	return n
}

func stepN(g *Game, in Input, n int) {
	for range n {
		g.Tick(in, 1.0/60)
	}
}

func TestNewPlacesActors(t *testing.T) {
	l := flatLevel(16, 12)
	g := New(l)
	if got := g.Player.Pos.Cell(); got != l.Spawn {
		t.Errorf("player spawned at %v, want %v", got, l.Spawn)
	}
	if g.Alien.State != StateLurk {
		t.Errorf("alien should start lurking, got %v", g.Alien.State)
	}
	if g.Dead() || g.Escaped() {
		t.Error("fresh game should be neither dead nor escaped")
	}
}

func TestPlayerMovesAndCollides(t *testing.T) {
	g := New(flatLevel(16, 12))
	g.Player.Angle = 0 // facing +X
	start := g.Player.Pos
	stepN(g, Input{Forward: 1}, 30)
	if g.Player.Pos.X <= start.X {
		t.Error("player did not move forward")
	}
	// Drive into the east wall: movement must stop at the border.
	stepN(g, Input{Forward: 1, Mode: ModeRun}, 600)
	if g.Player.Pos.X > 15 {
		t.Errorf("player clipped through the wall to x=%v", g.Player.Pos.X)
	}
}

func TestMoveModesChangeSpeed(t *testing.T) {
	dist := func(mode MoveMode) float64 {
		g := New(flatLevel(32, 12))
		g.Player.Angle = 0
		start := g.Player.Pos
		stepN(g, Input{Forward: 1, Mode: mode}, 60)
		return g.Player.Pos.Sub(start).Len()
	}
	sneak, walk, run := dist(ModeSneak), dist(ModeWalk), dist(ModeRun)
	if !(sneak < walk && walk < run) {
		t.Errorf("speed order wrong: sneak=%v walk=%v run=%v", sneak, walk, run)
	}
}

func TestTrackerCapsSpeed(t *testing.T) {
	g := New(flatLevel(32, 12))
	g.Player.Angle = 0
	start := g.Player.Pos
	stepN(g, Input{Forward: 1, Mode: ModeRun, Tracker: true}, 60)
	got := g.Player.Pos.Sub(start).Len()
	if got > sneakSpeed*1.05 {
		t.Errorf("running with tracker up moved %v tiles/s, want <= %v", got, sneakSpeed)
	}
}

func TestFootstepsObserved(t *testing.T) {
	rec := &recorder{}
	g := New(flatLevel(32, 12), WithObserver(rec))
	g.Player.Angle = 0
	stepN(g, Input{Forward: 1, Mode: ModeRun}, 120)
	if rec.count(ObsStep) == 0 {
		t.Error("running should emit footstep observations")
	}
}

func TestAlienHearsRunning(t *testing.T) {
	rec := &recorder{}
	g := New(flatLevel(24, 12), WithObserver(rec))
	g.Alien.Pos = Vec2{X: 8.5, Y: 1.5}
	g.Alien.Facing = 0 // facing away from the player
	g.Player.Angle = 0
	stepN(g, Input{Forward: 1, Mode: ModeRun}, 30)
	if rec.count(ObsAlienHeard) == 0 {
		t.Error("alien should hear running six tiles away")
	}
	if g.Alien.State != StateInvestigate && g.Alien.State != StateHunt {
		t.Errorf("heard alien should investigate, got %v", g.Alien.State)
	}
}

func TestAlienIgnoresDistantSneaking(t *testing.T) {
	rec := &recorder{}
	g := New(flatLevel(32, 12), WithObserver(rec))
	g.Alien.Pos = Vec2{X: 25.5, Y: 1.5}
	g.Alien.Facing = 0 // facing away: no vision contact either
	g.Player.Angle = 0
	stepN(g, Input{Forward: 1, Mode: ModeSneak}, 30)
	if rec.count(ObsAlienHeard) != 0 {
		t.Error("alien should not hear sneaking from across the map")
	}
}

func TestAlienSeesPlayerAndHunts(t *testing.T) {
	rec := &recorder{}
	g := New(flatLevel(24, 12), WithObserver(rec))
	g.Alien.Pos = Vec2{X: 6.5, Y: 1.5}
	g.Alien.Facing = math.Pi // looking straight at the player at (1.5, 1.5)
	stepN(g, Input{}, 5)
	if g.Alien.State != StateHunt {
		t.Errorf("alien with line of sight should hunt, got %v", g.Alien.State)
	}
	if rec.count(ObsAlienSeen) == 0 {
		t.Error("spotting the player should be observed")
	}
}

func TestWallBlocksVision(t *testing.T) {
	l := flatLevel(24, 12)
	// Wall column between player and alien.
	for y := 1; y < 11; y++ {
		l.Tiles[y*24+4] = world.TileWall
	}
	g := New(l)
	g.Alien.Pos = Vec2{X: 6.5, Y: 1.5}
	g.Alien.Facing = math.Pi
	stepN(g, Input{}, 5)
	if g.Alien.State == StateHunt {
		t.Error("alien should not see through walls")
	}
}

func TestAlienKillsOnContact(t *testing.T) {
	rec := &recorder{}
	g := New(flatLevel(16, 12), WithObserver(rec))
	g.Alien.Pos = Vec2{X: g.Player.Pos.X + 0.5, Y: g.Player.Pos.Y}
	g.Tick(Input{}, 1.0/60)
	if !g.Dead() {
		t.Fatal("adjacent alien with line of sight should kill")
	}
	if rec.count(ObsDeath) != 1 {
		t.Error("death should be observed once")
	}
	// Further ticks are inert after death.
	before := g.Player.Pos
	g.Tick(Input{Forward: 1}, 1.0/60)
	if g.Player.Pos != before {
		t.Error("dead player should not move")
	}
}

func TestDoorInteraction(t *testing.T) {
	l := flatLevel(16, 12)
	door := world.Coord{X: 3, Y: 1}
	l.Tiles[door.Y*16+door.X] = world.TileDoor
	rec := &recorder{}
	g := New(l, WithObserver(rec))
	g.Player.Angle = 0
	// Walk into the closed door: blocked.
	stepN(g, Input{Forward: 1}, 60)
	if g.Player.Pos.X >= 3 {
		t.Fatal("closed door should block movement")
	}
	g.Tick(Input{Use: true}, 1.0/60)
	if !g.World.DoorOpen(door) {
		t.Fatal("use in front of a door should open it")
	}
	if rec.count(ObsDoorOpen) != 1 {
		t.Error("door opening should be observed")
	}
	stepN(g, Input{Forward: 1}, 120)
	if g.Player.Pos.X <= 3 {
		t.Error("open door should be passable")
	}
}

func TestConsoleActivationUnlocksExit(t *testing.T) {
	l := flatLevel(16, 12)
	console := world.Coord{X: 3, Y: 0}
	l.Tiles[console.Y*16+console.X] = world.TileConsole
	l.Consoles = []world.Coord{console}
	rec := &recorder{}
	g := New(l, WithObserver(rec))
	if g.ExitUnlocked() {
		t.Fatal("exit should start locked")
	}
	// Stand under the console facing up (-Y) and use it.
	g.Player.Pos = Vec2{X: 3.5, Y: 1.5}
	g.Player.Angle = -math.Pi / 2
	g.Tick(Input{Use: true}, 1.0/60)
	if !g.ConsoleActivated(console) || g.ObjectivesDone() != 1 {
		t.Fatal("console should activate on use")
	}
	if !g.ExitUnlocked() {
		t.Error("activating every console should unlock the exit")
	}
	if rec.count(ObsConsole) != 1 {
		t.Error("console activation should be observed")
	}
	// A second use is a no-op.
	g.Tick(Input{Use: false}, 1.0/60)
	g.Tick(Input{Use: true}, 1.0/60)
	if rec.count(ObsConsole) != 1 {
		t.Error("console should activate only once")
	}
}

func TestEscapeRequiresUnlockedExit(t *testing.T) {
	l := flatLevel(16, 12)
	console := world.Coord{X: 3, Y: 0}
	l.Tiles[console.Y*16+console.X] = world.TileConsole
	l.Consoles = []world.Coord{console}
	g := New(l)
	g.Alien.Pos = Vec2{X: 8.5, Y: 5.5} // clear of the exit tile it wakes on
	g.Player.Pos = cellCenter(l.Exit)
	g.Tick(Input{}, 1.0/60)
	if g.Escaped() {
		t.Fatal("locked exit should not complete the level")
	}
	g.Player.Pos = Vec2{X: 3.5, Y: 1.5}
	g.Player.Angle = -math.Pi / 2
	g.Tick(Input{Use: true}, 1.0/60)
	g.Player.Pos = cellCenter(l.Exit)
	g.Tick(Input{}, 1.0/60)
	if !g.Escaped() {
		t.Error("unlocked exit should complete the level")
	}
}

func TestDeterminism(t *testing.T) {
	l, err := world.Generate(world.Config{Width: 32, Height: 24, Seed: 42})
	if err != nil {
		t.Fatal(err)
	}
	run := func() (Vec2, Vec2) {
		g := New(l)
		for i := range 600 {
			in := Input{Forward: 1, Mode: ModeRun}
			if i%120 < 30 {
				in.Turn = 1
			}
			g.Tick(in, 1.0/60)
		}
		return g.Player.Pos, g.Alien.Pos
	}
	p1, a1 := run()
	p2, a2 := run()
	if p1 != p2 || a1 != a2 {
		t.Errorf("same inputs diverged: player %v vs %v, alien %v vs %v", p1, p2, a1, a2)
	}
}
