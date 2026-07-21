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
		for x := range l.W {
			if l.At(x, 0) != TileWall || l.At(x, l.H-1) != TileWall {
				t.Fatalf("seed=%d: border breached at x=%d", seed, x)
			}
		}
		for y := range l.H {
			if l.At(0, y) != TileWall || l.At(l.W-1, y) != TileWall {
				t.Fatalf("seed=%d: border breached at y=%d", seed, y)
			}
		}
	}
}
