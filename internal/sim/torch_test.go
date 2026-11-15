package sim

import (
	"math"
	"testing"

	"github.com/danielriddell21/nemesis/internal/world"
)

func tap(g *Game, in Input) {
	g.Tick(in, 1.0/60)
	g.Tick(Input{}, 1.0/60)
}

func TestDiagonalIsNoFasterThanStraight(t *testing.T) {
	dist := func(in Input) float64 {
		g := New(flatLevel(32, 32))
		g.Player.Pos = Vec2{X: 16.5, Y: 16.5}
		g.Alien.Pos = Vec2{X: 30.5, Y: 30.5}
		start := g.Player.Pos
		stepN(g, in, 60)
		return g.Player.Pos.Sub(start).Len()
	}
	straight := dist(Input{Forward: 1})
	diagonal := dist(Input{Forward: 1, Strafe: 1})
	if math.Abs(diagonal-straight) > 1e-6 {
		t.Errorf("diagonal moved %.4f, straight %.4f: want equal", diagonal, straight)
	}
	if half := dist(Input{Forward: 0.5}); half >= straight {
		t.Errorf("a half push moved %.4f, want less than %.4f", half, straight)
	}
}

func TestTorchTogglesOnEachPress(t *testing.T) {
	rec := &recorder{}
	g := New(flatLevel(16, 12), WithObserver(rec))
	if g.Player.Torch || g.Player.Battery != 1 {
		t.Fatalf("fresh player torch=%v battery=%v, want off and full", g.Player.Torch, g.Player.Battery)
	}
	stepN(g, Input{Torch: true}, 10)
	if !g.Player.Torch || !g.TorchLit() {
		t.Error("holding the key should switch the torch on once")
	}
	g.Tick(Input{}, 1.0/60)
	tap(g, Input{Torch: true})
	if g.Player.Torch || g.TorchLit() {
		t.Error("a second press should switch the torch off")
	}
	if got := rec.count(ObsTorch); got != 2 {
		t.Errorf("observed %d torch switches, want 2", got)
	}
}

func TestTorchDrainsOnlyWhileLitAndDies(t *testing.T) {
	g := New(flatLevel(16, 12))
	stepN(g, Input{}, 120)
	if g.Player.Battery != 1 {
		t.Errorf("battery drained to %v with the torch off", g.Player.Battery)
	}
	tap(g, Input{Torch: true})
	stepN(g, Input{}, 60)
	if want := 1 - 1/torchLife; math.Abs(g.Player.Battery-want) > 0.01 {
		t.Errorf("battery after a second = %v, want about %v", g.Player.Battery, want)
	}
	g.Player.Battery = 0.001
	stepN(g, Input{}, 30)
	if g.Player.Torch || g.Player.Battery != 0 {
		t.Errorf("a flat cell should cut the torch: torch=%v battery=%v", g.Player.Torch, g.Player.Battery)
	}
	tap(g, Input{Torch: true})
	if g.Player.Torch {
		t.Error("a flat torch should not switch back on")
	}
}

func TestLowBatteryStutters(t *testing.T) {
	g := New(flatLevel(16, 12))
	g.Player.Torch = true
	lit := func(battery float64) (on, off int) {
		for i := range 600 {
			g.Player.Battery = battery
			g.elapsed = float64(i) / 60
			if g.TorchLit() {
				on++
			} else {
				off++
			}
		}
		return on, off
	}
	if _, off := lit(0.5); off != 0 {
		t.Errorf("a healthy cell cut out %d times", off)
	}
	on, off := lit(0.02)
	if on == 0 || off == 0 {
		t.Errorf("a dying cell should stutter: on=%d off=%d", on, off)
	}
	g.Player.Battery = 0.02
	if s := g.TorchStrength(); s > 0 && s >= 1 {
		t.Errorf("a dying beam should be dimmer than full, got %v", s)
	}
}

func TestHidingKillsTheTorch(t *testing.T) {
	l := flatLevel(16, 12)
	l.Set(3, 3, world.TileLocker)
	g := New(l)
	g.Player.Pos = Vec2{X: 3.5, Y: 3.5}
	g.Alien.Pos = Vec2{X: 13.5, Y: 9.5}
	tap(g, Input{Torch: true})
	tap(g, Input{Hide: true})
	if !g.Player.Hidden || g.Player.Torch {
		t.Errorf("hidden=%v torch=%v, want hidden with the torch off", g.Player.Hidden, g.Player.Torch)
	}
	tap(g, Input{Torch: true})
	if g.Player.Torch {
		t.Error("the torch should stay off inside a locker")
	}
}

func TestBatteryPickup(t *testing.T) {
	rec := &recorder{}
	l := flatLevel(16, 12)
	cell := world.Coord{X: 3, Y: 1}
	l.Batteries = []world.Coord{cell}
	g := New(l, WithObserver(rec))
	g.Alien.Pos = Vec2{X: 13.5, Y: 9.5}
	g.Player.Angle = 0

	// A full torch leaves the spare where it lies.
	g.Player.Pos = cellCenter(cell)
	g.Tick(Input{}, 1.0/60)
	if g.BatteryTaken(cell) {
		t.Fatal("picked up a cell with a full battery")
	}
	g.Player.Battery = 0.3
	g.Tick(Input{}, 1.0/60)
	if !g.BatteryTaken(cell) || math.Abs(g.Player.Battery-0.8) > 1e-9 {
		t.Errorf("taken=%v battery=%v, want taken and 0.8", g.BatteryTaken(cell), g.Player.Battery)
	}
	g.Player.Battery = 0.1
	g.Tick(Input{}, 1.0/60)
	if g.Player.Battery != 0.1 || rec.count(ObsBattery) != 1 {
		t.Errorf("a spent cell recharged again: battery=%v pickups=%d", g.Player.Battery, rec.count(ObsBattery))
	}
}

func TestTorchBetraysADarkStillPlayer(t *testing.T) {
	l := flatLevel(32, 12)
	for i := range l.Light {
		l.Light[i] = 0.1
	}
	setup := func(torch bool) *Game {
		g := New(l)
		g.Player.Pos = Vec2{X: 2.5, Y: 5.5}
		g.Alien.Pos = Vec2{X: 14.5, Y: 5.5}
		g.Alien.Facing = math.Pi
		g.Player.Torch = torch
		stepN(g, Input{Mode: ModeSneak}, 5)
		return g
	}
	if g := setup(false); g.Alien.State == StateHunt {
		t.Error("a still player in the dark twelve tiles off should go unseen")
	}
	if g := setup(true); g.Alien.State != StateHunt {
		t.Errorf("a lit torch in plain view should be spotted, got %v", g.Alien.State)
	}
}

func TestGlowDrawsAHunterLookingAway(t *testing.T) {
	rec := &recorder{}
	g := New(flatLevel(32, 12), WithObserver(rec))
	g.Player.Pos = Vec2{X: 2.5, Y: 5.5}
	g.Alien.Pos = Vec2{X: 8.5, Y: 5.5}
	g.Alien.Facing = 0 // back turned
	stepN(g, Input{}, 30)
	if rec.count(ObsAlienGlow) != 0 || g.Alien.State != StateLurk {
		t.Fatalf("no torch, no glow: state=%v", g.Alien.State)
	}
	g.Player.Torch = true
	stepN(g, Input{}, 2)
	if rec.count(ObsAlienGlow) != 1 {
		t.Errorf("glow observed %d times, want 1", rec.count(ObsAlienGlow))
	}
	if g.Alien.State != StateInvestigate && g.Alien.State != StateHunt {
		t.Errorf("a hunter that caught the glow should come looking, got %v", g.Alien.State)
	}
}

func TestGlowNeedsRangeAndLineOfSight(t *testing.T) {
	far := New(flatLevel(40, 12))
	far.Player.Pos = Vec2{X: 2.5, Y: 5.5}
	far.Alien.Pos = Vec2{X: 2.5 + glowRange + 4, Y: 5.5}
	far.Player.Torch = true
	stepN(far, Input{}, 5)
	if far.Alien.State != StateLurk {
		t.Errorf("glow carried past its range: %v", far.Alien.State)
	}

	l := flatLevel(32, 12)
	for y := 1; y < 11; y++ {
		l.Set(5, y, world.TileWall)
	}
	walled := New(l)
	walled.Player.Pos = Vec2{X: 2.5, Y: 5.5}
	walled.Alien.Pos = Vec2{X: 8.5, Y: 5.5}
	walled.Player.Torch = true
	stepN(walled, Input{}, 5)
	if walled.Alien.State != StateLurk {
		t.Errorf("glow passed through a wall: %v", walled.Alien.State)
	}
}

func TestColdHuntFollowsWhatItHears(t *testing.T) {
	l := flatLevel(32, 12)
	for y := 1; y < 9; y++ {
		l.Set(10, y, world.TileWall)
	}
	g := New(l)
	g.Player.Pos = Vec2{X: 7.5, Y: 3.5}
	g.Player.Angle = math.Pi
	g.Alien.Pos = Vec2{X: 11.5, Y: 3.5}
	g.Alien.State = StateHunt
	g.Alien.tracked = false
	g.Alien.lastKnown = Vec2{X: 20.5, Y: 9.5}

	stepN(g, Input{Mode: ModeSneak}, 3)
	if g.Alien.lastKnown.Cell() != (world.Coord{X: 20, Y: 9}) {
		t.Fatal("a silent player should not redirect the chase")
	}
	// A footfall lands within the first stride.
	stepN(g, Input{Forward: 1, Mode: ModeRun}, 15)
	if got := g.Alien.lastKnown.Cell(); manhattanCells(got, g.Player.Pos.Cell()) > 1 {
		t.Errorf("after hearing a sprint the chase heads to %v, want the player near %v", got, g.Player.Pos.Cell())
	}
}

func manhattanCells(a, b world.Coord) int {
	return absInt(a.X-b.X) + absInt(a.Y-b.Y)
}

func TestExposedMatchesTheHuntersSight(t *testing.T) {
	l := flatLevel(32, 12)
	for y := 1; y < 9; y++ {
		l.Set(10, y, world.TileWall)
	}
	g := New(l)
	g.Alien.Pos = Vec2{X: 6.5, Y: 3.5}
	g.Alien.Facing = math.Pi
	cases := []struct {
		name string
		cell world.Coord
		want bool
	}{
		{"in front and close", world.Coord{X: 3, Y: 3}, true},
		{"behind its back", world.Coord{X: 9, Y: 3}, false},
		{"beyond the wall", world.Coord{X: 13, Y: 3}, false},
	}
	for _, c := range cases {
		if got := g.Exposed(c.cell); got != c.want {
			t.Errorf("%s: Exposed(%v) = %v, want %v", c.name, c.cell, got, c.want)
		}
	}
	g.Alien.Facing = 0
	if !g.Exposed(world.Coord{X: 9, Y: 3}) {
		t.Error("turning the hunter should expose the cell now in front of it")
	}
}
