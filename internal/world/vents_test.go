package world

import "testing"

func TestVentsCarved(t *testing.T) {
	// Every generated level must carry a vent system: at least two mouths, and
	// every vent cell part of a tunnel that reaches a mouth.
	for seed := int64(0); seed < 100; seed++ {
		l, err := Generate(Config{Width: 48, Height: 32, Seed: seed})
		if err != nil {
			t.Fatalf("seed=%d: %v", seed, err)
		}
		if len(l.VentMouths) < 2 {
			t.Fatalf("seed=%d: %d vent mouths, want >= 2", seed, len(l.VentMouths))
		}
		for _, m := range l.VentMouths {
			if l.At(m.X, m.Y) != TileVent {
				t.Errorf("seed=%d: mouth %v is not a vent cell", seed, m)
			}
		}
	}
}

func TestVentMouthsOpenOntoFloor(t *testing.T) {
	l, err := Generate(Config{Width: 48, Height: 32, Seed: 9})
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range l.VentMouths {
		open := false
		for _, n := range neighbors4(m) {
			if t := l.At(n.X, n.Y); t.Walkable() && t != TileVent {
				open = true
			}
		}
		if !open {
			t.Errorf("mouth %v does not open onto a non-vent walkable cell", m)
		}
	}
}

func TestVentsNeverBreachBorder(t *testing.T) {
	for seed := int64(0); seed < 50; seed++ {
		l, err := Generate(Config{Width: 32, Height: 24, Seed: seed})
		if err != nil {
			t.Fatalf("seed=%d: %v", seed, err)
		}
		for x := range l.Width {
			if l.At(x, 0) != TileWall || l.At(x, l.Height-1) != TileWall {
				t.Fatalf("seed=%d: border breached at x=%d", seed, x)
			}
		}
		for y := range l.Height {
			if l.At(0, y) != TileWall || l.At(l.Width-1, y) != TileWall {
				t.Fatalf("seed=%d: border breached at y=%d", seed, y)
			}
		}
	}
}

func TestBFSPathEndpoints(t *testing.T) {
	l := newLevel(8, 8, 0)
	// Straight strip of wall from (1,1) to (5,1) inside the border.
	ok := func(c Coord) bool { return c.Y == 1 && c.X >= 1 && c.X <= 5 }
	path := bfsPath(l, Coord{X: 1, Y: 1}, Coord{X: 5, Y: 1}, ok)
	if len(path) != 5 {
		t.Fatalf("path length %d, want 5", len(path))
	}
	if path[0] != (Coord{X: 5, Y: 1}) || path[len(path)-1] != (Coord{X: 1, Y: 1}) {
		t.Errorf("path endpoints wrong: %v", path)
	}
	if bfsPath(l, Coord{X: 1, Y: 1}, Coord{X: 6, Y: 6}, ok) != nil {
		t.Error("unreachable destination should yield nil path")
	}
}
