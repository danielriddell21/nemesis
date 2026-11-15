package render

import (
	"bytes"
	"math"
	"testing"

	"github.com/danielriddell21/crucible/geom"
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
	hit := castRay(g, geom.Vec2{X: 4.5, Y: 4.5}, 1, 0)
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

func darkBox() *world.Level {
	l := boxLevel()
	for i := range l.Light {
		l.Light[i] = 0.05
	}
	return l
}

func pixel(fb []byte, w, x, y int) (r, g, b int) {
	i := (y*w + x) * 4
	return int(fb[i]), int(fb[i+1]), int(fb[i+2])
}

func TestTorchLightsADarkRoom(t *testing.T) {
	g := sim.New(darkBox())
	g.Player.Pos = sim.Vec2{X: 2.5, Y: 4.5}
	g.Player.Angle = 0
	cfg := Config{Width: 160, Height: 100, FOV: 1.152}
	dark := append([]byte(nil), NewRenderer(cfg).Frame(g, 0)...)
	g.Player.Torch = true
	lit := NewRenderer(cfg).Frame(g, 0)
	if brightness(lit) < brightness(dark)*3 {
		t.Errorf("torch frame brightness %d vs dark %d: the beam should transform a dark room", brightness(lit), brightness(dark))
	}
	// The beam is a cone: the centre of the far wall outshines its edge.
	cr, cg, cb := pixel(lit, 160, 80, 50)
	er, eg, eb := pixel(lit, 160, 4, 50)
	if cr+cg+cb <= er+eg+eb {
		t.Errorf("beam centre %d is no brighter than the screen edge %d", cr+cg+cb, er+eg+eb)
	}
	// And it is warm where the station's own light is cool: the same wall
	// under ambient light alone leans further to blue.
	amb := sim.New(boxLevel())
	amb.Player.Pos = sim.Vec2{X: 2.5, Y: 4.5}
	ar, _, ab := pixel(NewRenderer(cfg).Frame(amb, 0), 160, 80, 50)
	if cr*ab <= ar*cb {
		t.Errorf("torchlit wall R/B = %d/%d, ambient wall R/B = %d/%d: the beam should be warmer", cr, cb, ar, ab)
	}
}

func TestDeadCellLeavesTheRoomDark(t *testing.T) {
	g := sim.New(darkBox())
	g.Player.Pos = sim.Vec2{X: 2.5, Y: 4.5}
	cfg := Config{Width: 160, Height: 100, FOV: 1.152}
	dark := brightness(NewRenderer(cfg).Frame(g, 0))
	g.Player.Torch = true
	g.Player.Battery = 0
	if got := brightness(NewRenderer(cfg).Frame(g, 0)); got > dark*11/10 {
		t.Errorf("a flat torch lit the room: %d vs %d", got, dark)
	}
}

func TestHunterEyesShowInTheDark(t *testing.T) {
	l := darkBox()
	// Park the airlock beacon out of view so the only glow is the hunter's.
	l.Exit = world.Coord{X: 1, Y: 1}
	g := sim.New(l)
	g.Player.Pos = sim.Vec2{X: 1.5, Y: 4.5}
	g.Player.Angle = 0
	cfg := Config{Width: 320, Height: 200, FOV: 1.152}
	g.Alien.Pos = sim.Vec2{X: -50, Y: -50}
	empty := peak(NewRenderer(cfg).Frame(g, 0), 320, 40, 140)
	g.Alien.Pos = sim.Vec2{X: 4.5, Y: 4.5}
	withHunter := peak(NewRenderer(cfg).Frame(g, 0), 320, 40, 140)
	if withHunter < empty+80 {
		t.Errorf("brightest pixel with the hunter ahead = %d, empty room = %d: its eyes should glow", withHunter, empty)
	}
}

func peak(fb []byte, w, y0, y1 int) int {
	best := 0
	for y := y0; y < y1; y++ {
		for x := range w {
			r, g, b := pixel(fb, w, x, y)
			best = max(best, r, g, b)
		}
	}
	return best
}

func TestFloorIsTexturedAndFadesWithDistance(t *testing.T) {
	g := sim.New(boxLevel())
	g.Player.Pos = sim.Vec2{X: 1.5, Y: 4.5}
	g.Player.Angle = 0
	fb := NewRenderer(Config{Width: 160, Height: 100, FOV: 1.152}).Frame(g, 0)
	// Far rows near the horizon are dimmer than the deck at the player's feet.
	mean := func(y0, y1 int) int {
		total, n := 0, 0
		for y := y0; y < y1; y++ {
			for x := 60; x < 100; x++ {
				r, gg, b := pixel(fb, 160, x, y)
				total += r + gg + b
				n++
			}
		}
		return total / n
	}
	if near, far := mean(84, 96), mean(60, 63); near <= far {
		t.Errorf("near floor %d should be brighter than far floor %d", near, far)
	}
	// Plating seams and stains: a floor row is not one flat colour.
	seen := map[int]bool{}
	for x := 40; x < 120; x++ {
		r, _, _ := pixel(fb, 160, x, 92)
		seen[r/8] = true
	}
	if len(seen) < 2 {
		t.Error("the floor row is a single flat tone; want plating texture")
	}
}

func TestBatteryGauge(t *testing.T) {
	cases := []struct {
		charge float64
		want   int
	}{{1, 10}, {0.95, 10}, {0.5, 5}, {0.41, 5}, {0.05, 1}, {0.001, 1}, {0, 0}}
	for _, c := range cases {
		if got := batteryLit(c.charge); got != c.want {
			t.Errorf("batteryLit(%v) = %d, want %d", c.charge, got, c.want)
		}
	}
	g := testGame(t)
	cfg := Config{Width: 160, Height: 100, FOV: 1.152}
	full := append([]byte(nil), NewRenderer(cfg).Frame(g, 0)...)
	g.Player.Battery = 0.3
	if bytes.Equal(full, NewRenderer(cfg).Frame(g, 0)) {
		t.Error("draining the battery should change the gauge")
	}
}

func TestBatteryPickupDrawnUntilTaken(t *testing.T) {
	l := boxLevel()
	cell := world.Coord{X: 4, Y: 4}
	l.Batteries = []world.Coord{cell}
	g := sim.New(l)
	g.Alien.Pos = sim.Vec2{X: -50, Y: -50}
	g.Player.Pos = sim.Vec2{X: 1.5, Y: 4.5}
	g.Player.Angle = 0
	cfg := Config{Width: 160, Height: 100, FOV: 1.152}
	before := append([]byte(nil), NewRenderer(cfg).Frame(g, 0)...)
	g.Player.Battery = 0.2
	g.Player.Pos = sim.Vec2{X: 4.5, Y: 4.5}
	g.Tick(sim.Input{}, 1.0/60)
	if !g.BatteryTaken(cell) {
		t.Fatal("expected the cell to be picked up")
	}
	g.Player.Pos = sim.Vec2{X: 1.5, Y: 4.5}
	g.Player.Battery = 1
	if bytes.Equal(before, NewRenderer(cfg).Frame(g, 0)) {
		t.Error("a taken cell should vanish from the deck")
	}
}

func TestMasksAreBoundedAndCentred(t *testing.T) {
	r := NewRenderer(Config{Width: 160, Height: 100, FOV: 1.152})
	centre, corner := r.cone[52*160+80], r.cone[0]
	if centre <= corner || corner < beamSpill-1e-6 {
		t.Errorf("cone centre %.2f corner %.2f: want a bright core and a faint spill", centre, corner)
	}
	if v := r.vignette[50*160+80]; v != 256 {
		t.Errorf("vignette at the centre = %d, want full strength 256", v)
	}
	if v := r.vignette[0]; v >= 256 || v < 100 {
		t.Errorf("vignette at the corner = %d, want darker but not black", v)
	}
	if smoothstep(0, 1, -1) != 0 || smoothstep(0, 1, 2) != 1 || smoothstep(0, 1, 0.5) != 0.5 {
		t.Error("smoothstep should clamp and pass through the midpoint")
	}
}

func TestGrainCrawlsButStaysDeterministic(t *testing.T) {
	g := sim.New(boxLevel())
	g.Player.Pos = sim.Vec2{X: 2.5, Y: 4.5}
	cfg := Config{Width: 160, Height: 100, FOV: 1.152, Grain: true}
	a := append([]byte(nil), NewRenderer(cfg).Frame(g, 1)...)
	b := append([]byte(nil), NewRenderer(cfg).Frame(g, 1)...)
	if !bytes.Equal(a, b) {
		t.Error("grain at the same instant should be identical")
	}
	if bytes.Equal(a, NewRenderer(cfg).Frame(g, 2)) {
		t.Error("grain should crawl from one moment to the next")
	}
	clean := Config{Width: 160, Height: 100, FOV: 1.152}
	c := append([]byte(nil), NewRenderer(clean).Frame(g, 1)...)
	if !bytes.Equal(c, NewRenderer(clean).Frame(g, 2)) {
		t.Error("without grain a still scene should not change over time")
	}
	if !DefaultConfig().Grain {
		t.Error("the game's default config should have grain on")
	}
}
