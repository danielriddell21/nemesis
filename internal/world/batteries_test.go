package world

import (
	"slices"
	"testing"
)

func TestBatteriesPlacedOnReachableFloor(t *testing.T) {
	for seed := int64(1); seed <= 40; seed++ {
		l, err := Generate(Config{Width: 48, Height: 32, Seed: seed})
		if err != nil {
			t.Fatal(err)
		}
		if len(l.Batteries) == 0 {
			t.Errorf("seed %d: no batteries placed", seed)
		}
		for _, c := range l.Batteries {
			if l.At(c.X, c.Y) != TileFloor {
				t.Errorf("seed %d: battery at %v sits on %v, want floor", seed, c, l.At(c.X, c.Y))
			}
			if StepsBetween(l, l.Spawn, c) < 0 {
				t.Errorf("seed %d: battery at %v is unreachable", seed, c)
			}
			if !l.BatteryAt(c) {
				t.Errorf("seed %d: BatteryAt(%v) = false", seed, c)
			}
		}
	}
}

func TestBatteryCountClampedAndDeterministic(t *testing.T) {
	none, err := Generate(Config{Width: 48, Height: 32, Seed: 3, Batteries: -1})
	if err != nil {
		t.Fatal(err)
	}
	if len(none.Batteries) != 0 {
		t.Errorf("negative count placed %d batteries", len(none.Batteries))
	}
	a, _ := Generate(Config{Width: 48, Height: 32, Seed: 3, Batteries: 2})
	b, _ := Generate(Config{Width: 48, Height: 32, Seed: 3, Batteries: 2})
	if len(a.Batteries) > 2 || !slices.Equal(a.Batteries, b.Batteries) {
		t.Errorf("batteries %v vs %v: want the same, at most 2", a.Batteries, b.Batteries)
	}
	if a.BatteryAt(a.Spawn) {
		t.Error("battery placed on the spawn")
	}
}

func TestStationIsDim(t *testing.T) {
	l, err := Generate(Config{Width: 48, Height: 32, Seed: 9})
	if err != nil {
		t.Fatal(err)
	}
	for y := range l.H {
		for x := range l.W {
			if v := l.LightAt(x, y); l.At(x, y).Walkable() && (v <= 0 || v > 0.6) {
				t.Fatalf("light at %d,%d = %.2f, want dim but never black", x, y, v)
			}
		}
	}
}
