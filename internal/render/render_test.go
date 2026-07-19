package render

import (
	"bytes"
	"math"
	"testing"

	"github.com/danielriddell21/nemesis/internal/hud"
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
	l := &world.Level{
		Width:  8,
		Height: 8,
		Tiles:  make([]world.TileType, 64),
		Light:  make([]float64, 64),
		Spawn:  world.Coord{X: 1, Y: 1},
		Exit:   world.Coord{X: 6, Y: 6},
	}
	for y := range 8 {
		for x := range 8 {
			t := world.TileFloor
			if x == 0 || y == 0 || x == 7 || y == 7 {
				t = world.TileWall
			}
			l.Tiles[y*8+x] = t
			l.Light[y*8+x] = 0.8
		}
	}
	return l
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
	if hit.cell != (world.Coord{X: 7, Y: 4}) {
		t.Errorf("ray hit %v, want the east border wall", hit.cell)
	}
	if math.Abs(hit.dist-2.5) > 1e-9 {
		t.Errorf("hit distance %v, want 2.5", hit.dist)
	}
}
