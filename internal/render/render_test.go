package render

import (
	"bytes"
	"math"
	"testing"

	"github.com/danielriddell21/crucible/hud"
	"github.com/danielriddell21/crucible/level"

	"github.com/danielriddell21/nemesis/internal/sim"
	"github.com/danielriddell21/nemesis/internal/world"
)

func testGame(t *testing.T) *sim.Game {
	t.Helper()
	l, err := world.Generate(world.Config{Width: 32, Height: 24, Seed: 7})
	if err != nil {
		t.Fatal(err)
	}
	return sim.New(l)
}

func TestFrameShape(t *testing.T) {
	g := testGame(t)
	r := NewRenderer(Config{Width: 160, Height: 100, FOV: 1.152})
	fb := r.Frame(g, 0)
	if len(fb) != 160*100*4 {
		t.Fatalf("framebuffer is %d bytes, want %d", len(fb), 160*100*4)
	}
	for i := 3; i < len(fb); i += 4 {
		if fb[i] != 255 {
			t.Fatalf("pixel %d not opaque", i/4)
		}
	}
}

func TestFrameDeterministic(t *testing.T) {
	g := testGame(t)
	r1 := NewRenderer(Config{Width: 160, Height: 100, FOV: 1.152})
	r2 := NewRenderer(Config{Width: 160, Height: 100, FOV: 1.152})
	a := append([]byte(nil), r1.Frame(g, 1.5)...)
	b := r2.Frame(g, 1.5)
	if !bytes.Equal(a, b) {
		t.Error("same state should render the same frame")
	}
}

func TestFrameShowsWalls(t *testing.T) {
	// A bare lit box, facing the east wall: the centre pixel must be the wall
	// texture's cool grey (blue channel above red) at a sensible brightness.
	g := sim.New(boxLevel())
	g.Player.Pos = sim.Vec2{X: 4.5, Y: 4.5}
	g.Player.Angle = 0
	r := NewRenderer(Config{Width: 160, Height: 100, FOV: 1.152})
	fb := r.Frame(g, 0)
	i := (50*160 + 80) * 4
	red, blue := fb[i], fb[i+2]
	if blue <= red || blue < 30 {
		t.Errorf("centre pixel R=%d B=%d, want a lit wall grey", red, blue)
	}
}

func boxLevel() *world.Level {
	base := level.New(8, 8, 0)
	for y := range 8 {
		for x := range 8 {
			t := world.TileFloor
			if x == 0 || y == 0 || x == 7 || y == 7 {
				t = world.TileWall
			}
			base.Set(x, y, t)
			base.Light[base.Index(x, y)] = 0.8
		}
	}
	base.Spawn = world.Coord{X: 1, Y: 1}
	base.Exit = world.Coord{X: 6, Y: 6}
	return &world.Level{Level: base}
}

func TestTrackerDrawsWhenRaised(t *testing.T) {
	g := testGame(t)
	cfg := Config{Width: 160, Height: 100, FOV: 1.152}
	plain := NewRenderer(cfg).Frame(g, 0)
	plainCopy := append([]byte(nil), plain...)
	g.Tick(sim.Input{Tracker: true}, 1.0/60)
	raised := NewRenderer(cfg).Frame(g, 0)
	if bytes.Equal(plainCopy, raised) {
		t.Error("raising the tracker should change the frame")
	}
}

func TestDeathTintsFrame(t *testing.T) {
	g := testGame(t)
	cfg := Config{Width: 160, Height: 100, FOV: 1.152}
	before := append([]byte(nil), NewRenderer(cfg).Frame(g, 0)...)
	g.Alien.Pos = g.Player.Pos
	g.Tick(sim.Input{}, 1.0/60)
	if !g.Dead() {
		t.Fatal("expected the alien on top of the player to kill")
	}
	after := NewRenderer(cfg).Frame(g, 0)
	redBefore, redAfter := 0, 0
	for i := 0; i < len(after); i += 4 {
		redBefore += int(before[i])
		redAfter += int(after[i])
	}
	if redAfter <= redBefore {
		t.Error("death should tint the frame red")
	}
}

func TestOverlayMessageDrawn(t *testing.T) {
	g := testGame(t)
	cfg := Config{Width: 160, Height: 100, FOV: 1.152}
	o := hud.New()
	bare := append([]byte(nil), NewRenderer(cfg, WithOverlay(o)).Frame(g, 0)...)
	o.Post("VENT CREAK AT 12 4", 60, hud.Notice)
	posted := NewRenderer(cfg, WithOverlay(o)).Frame(g, 0)
	if bytes.Equal(bare, posted) {
		t.Error("posting an overlay message should change the frame")
	}
}

func TestCastRayHitsBorder(t *testing.T) {
	g := sim.New(boxLevel())
	hit := castRay(g, sim.Vec2{X: 4.5, Y: 4.5}, 1, 0)
	if hit.Cell != (world.Coord{X: 7, Y: 4}) {
		t.Errorf("ray hit %v, want the east border wall", hit.Cell)
	}
	if math.Abs(hit.Dist-2.5) > 1e-9 {
		t.Errorf("hit distance %v, want 2.5", hit.Dist)
	}
}

func TestSlidingDoorReveals(t *testing.T) {
	// A door dead ahead: shut it occludes the room; fully open the ray passes
	// through to the wall beyond, so the mid-column changes as it slides.
	l := boxLevel()
	l.Tiles[4*8+2] = world.TileDoor // a door in the west part of the box
	g := sim.New(l)
	g.Player.Pos = sim.Vec2{X: 1.5, Y: 4.5}
	g.Player.Angle = 0 // facing the door to the east
	shut := NewRenderer(Config{Width: 64, Height: 48, FOV: 1.152})
	shutFrame := append([]byte(nil), shut.Frame(g, 0)...)

	g.World.OpenDoor(world.Coord{X: 2, Y: 4})
	for range 40 { // let it slide fully open
		g.Tick(sim.Input{}, 1.0/60)
	}
	openFrame := NewRenderer(Config{Width: 64, Height: 48, FOV: 1.152}).Frame(g, 0)
	if bytes.Equal(shutFrame, openFrame) {
		t.Error("a sliding door should change the view between shut and open")
	}
}

func TestHiddenPlayerDarkensView(t *testing.T) {
	l := boxLevel()
	l.Tiles[4*8+2] = world.TileLocker
	l.Lockers = []world.Coord{{X: 2, Y: 4}}
	g := sim.New(l)
	g.Player.Pos = sim.Vec2{X: 2.5, Y: 4.5}
	cfg := Config{Width: 64, Height: 48, FOV: 1.152}
	before := brightness(NewRenderer(cfg).Frame(g, 0))
	g.Tick(sim.Input{Hide: true}, 1.0/60)
	if !g.Player.Hidden {
		t.Fatal("player should be hidden")
	}
	after := brightness(NewRenderer(cfg).Frame(g, 0))
	if after >= before {
		t.Error("the locker view should darken the frame")
	}
}

func brightness(fb []byte) int {
	total := 0
	for i := 0; i < len(fb); i += 4 {
		total += int(fb[i]) + int(fb[i+1]) + int(fb[i+2])
	}
	return total
}
