package world

import (
	"reflect"
	"testing"

	"github.com/danielriddell21/crucible/worldgen"
)

// exitReachable reports whether the exit can be reached from the spawn with
// doors treated as passable — the deck's basic connectivity guarantee.
func exitReachable(l *Level) bool {
	return worldgen.Reachable(l.W, l.H, l.Spawn, l.Exit, blocksWalls(l.Level))
}

func TestGenerateConnectivity(t *testing.T) {
	sizes := []struct {
		w, h int
	}{
		{16, 16},
		{32, 24},
		{48, 32},
		{64, 48},
	}
	// For each size, many seeds must all yield an exit and consoles reachable
	// from spawn.
	for _, s := range sizes {
		for seed := int64(0); seed < 150; seed++ {
			l, err := Generate(Config{Width: s.w, Height: s.h, Seed: seed})
			if err != nil {
				t.Fatalf("%dx%d seed=%d: %v", s.w, s.h, seed, err)
			}
			if l.At(l.Spawn.X, l.Spawn.Y) != TileSpawn {
				t.Errorf("%dx%d seed=%d: spawn tile not marked", s.w, s.h, seed)
			}
			if l.At(l.Exit.X, l.Exit.Y) != TileExit {
				t.Errorf("%dx%d seed=%d: exit tile not marked", s.w, s.h, seed)
			}
			if !exitReachable(l) {
				t.Errorf("%dx%d seed=%d: exit unreachable from spawn\n%s", s.w, s.h, seed, l)
			}
			dist := distanceField(l.Level, l.Spawn)
			for _, c := range l.Consoles {
				if l.At(c.X, c.Y) != TileConsole {
					t.Errorf("%dx%d seed=%d: console cell %v not marked", s.w, s.h, seed, c)
				}
				if !faceReachable(dist, c) {
					t.Errorf("%dx%d seed=%d: console %v has no reachable face\n%s", s.w, s.h, seed, c, l)
				}
			}
		}
	}
}

func TestGenerateDeterminism(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
	}{
		{"small", Config{Width: 24, Height: 24, Seed: 1}},
		{"wide", Config{Width: 64, Height: 32, Seed: 7}},
		{"large", Config{Width: 80, Height: 60, Seed: 123456}},
		{"zero seed", Config{Width: 40, Height: 30, Seed: 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := Generate(tt.cfg)
			if err != nil {
				t.Fatal(err)
			}
			b, err := Generate(tt.cfg)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(a, b) {
				t.Error("same config generated different levels")
			}
		})
	}
}

func TestGenerateSeedsDiffer(t *testing.T) {
	a, err := Generate(Config{Width: 48, Height: 32, Seed: 1})
	if err != nil {
		t.Fatal(err)
	}
	b, err := Generate(Config{Width: 48, Height: 32, Seed: 2})
	if err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(a.Tiles, b.Tiles) {
		t.Error("different seeds generated identical tile grids")
	}
}

func TestGenerateConsoleCount(t *testing.T) {
	for _, want := range []int{1, 2, 3, 4, 5} {
		l, err := Generate(Config{Width: 48, Height: 32, Seed: 11, Consoles: want})
		if err != nil {
			t.Fatalf("consoles=%d: %v", want, err)
		}
		if len(l.Consoles) != want {
			t.Errorf("consoles=%d: got %d placed", want, len(l.Consoles))
		}
	}
}

func TestGenerateNormalizesConfig(t *testing.T) {
	// Degenerate requests are clamped up to the minimum viable grid rather than
	// erroring, and silly console counts are clamped into range.
	l, err := Generate(Config{Width: 1, Height: 1, Seed: 3, Consoles: 99})
	if err != nil {
		t.Fatal(err)
	}
	if l.W != minDimension || l.H != minDimension {
		t.Errorf("got %dx%d, want clamped to %dx%d", l.W, l.H, minDimension, minDimension)
	}
	if len(l.Consoles) != 5 {
		t.Errorf("got %d consoles, want clamped to 5", len(l.Consoles))
	}
}

func TestGenerateSpawnLit(t *testing.T) {
	l, err := Generate(Config{Width: 48, Height: 32, Seed: 5})
	if err != nil {
		t.Fatal(err)
	}
	if l.LightAt(l.Spawn.X, l.Spawn.Y) <= 0 {
		t.Error("spawn tile has no light")
	}
	for y := range l.H {
		for x := range l.W {
			if l.At(x, y).Walkable() && l.LightAt(x, y) <= 0 {
				t.Errorf("walkable cell (%d,%d) unlit", x, y)
			}
		}
	}
}
