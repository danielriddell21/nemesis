package world

import "testing"

func TestLockersPlaced(t *testing.T) {
	// Most levels should place some lockers, each a walkable floor recess
	// backing onto a wall and clear of the spawn.
	withLockers := 0
	for seed := int64(0); seed < 80; seed++ {
		l, err := Generate(Config{Width: 48, Height: 32, Seed: seed})
		if err != nil {
			t.Fatalf("seed=%d: %v", seed, err)
		}
		if len(l.Lockers) > 0 {
			withLockers++
		}
		for _, c := range l.Lockers {
			if l.At(c.X, c.Y) != TileLocker {
				t.Errorf("seed=%d: locker cell %v not marked", seed, c)
			}
			if !l.At(c.X, c.Y).Walkable() {
				t.Errorf("seed=%d: locker %v not walkable", seed, c)
			}
			if !backsOntoWall(l.Level, c) {
				t.Errorf("seed=%d: locker %v does not back onto a wall", seed, c)
			}
			if c == l.Spawn {
				t.Errorf("seed=%d: locker placed on the spawn", seed)
			}
		}
	}
	if withLockers < 60 {
		t.Errorf("only %d/80 levels grew lockers, want most", withLockers)
	}
}

func TestLockerCountClamped(t *testing.T) {
	l, err := Generate(Config{Width: 64, Height: 40, Seed: 3, Lockers: 3})
	if err != nil {
		t.Fatal(err)
	}
	if len(l.Lockers) > 3 {
		t.Errorf("got %d lockers, want at most 3", len(l.Lockers))
	}
	none, err := Generate(Config{Width: 48, Height: 32, Seed: 3, Lockers: -1})
	if err != nil {
		t.Fatal(err)
	}
	if len(none.Lockers) != 0 {
		t.Errorf("negative locker request should place none, got %d", len(none.Lockers))
	}
}

func TestLockersDeterministic(t *testing.T) {
	cfg := Config{Width: 48, Height: 32, Seed: 17}
	a, _ := Generate(cfg)
	b, _ := Generate(cfg)
	if len(a.Lockers) != len(b.Lockers) {
		t.Fatalf("locker counts differ: %d vs %d", len(a.Lockers), len(b.Lockers))
	}
	for i := range a.Lockers {
		if a.Lockers[i] != b.Lockers[i] {
			t.Errorf("locker %d differs: %v vs %v", i, a.Lockers[i], b.Lockers[i])
		}
	}
}
