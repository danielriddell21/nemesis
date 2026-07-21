package world

import (
	"strings"
	"testing"

	"github.com/danielriddell21/crucible/level"
)

func newTestLevel(w, h int) *Level {
	return &Level{Level: level.New(w, h, 0)}
}

func TestTileWalkable(t *testing.T) {
	tests := []struct {
		tile Tile
		want bool
	}{
		{TileFloor, true},
		{TileDoor, true},
		{TileVent, true},
		{TileSpawn, true},
		{TileExit, true},
		{TileLocker, true},
		{TileWall, false},
		{TileConsole, false},
	}
	for _, tt := range tests {
		if got := tt.tile.Walkable(); got != tt.want {
			t.Errorf("%c.Walkable() = %v, want %v", tt.tile.Rune(), got, tt.want)
		}
	}
}

func TestLevelAtOutOfBounds(t *testing.T) {
	l := newTestLevel(4, 4)
	if l.At(-1, 0) != TileWall || l.At(0, -1) != TileWall || l.At(4, 0) != TileWall || l.At(0, 4) != TileWall {
		t.Error("out-of-bounds reads should return TileWall")
	}
	if l.FlickerAt(-1, 0) {
		t.Error("out-of-bounds flicker should be false")
	}
}

func TestLevelSolid(t *testing.T) {
	// Solidity is the engine's !walkable rule: walls and consoles block,
	// doors do not — the closed-door block is sim.World's runtime concern.
	l := newTestLevel(4, 4)
	l.Set(1, 1, TileFloor)
	l.Set(2, 1, TileDoor)
	l.Set(1, 2, TileConsole)
	l.Set(2, 2, TileVent)
	if l.Solid(1, 1) || l.Solid(2, 2) || l.Solid(2, 1) {
		t.Error("floor, vent and door should not be solid at the level layer")
	}
	if !l.Solid(1, 2) || !l.Solid(0, 0) {
		t.Error("console and wall should be solid")
	}
}

func TestRoomAt(t *testing.T) {
	l := newTestLevel(8, 8)
	l.Rooms = []Room{{X: 1, Y: 1, W: 3, H: 3}, {X: 5, Y: 5, W: 2, H: 2}}
	if got := l.RoomAt(Coord{X: 2, Y: 2}); got != 0 {
		t.Errorf("RoomAt(2,2) = %d, want 0", got)
	}
	if got := l.RoomAt(Coord{X: 5, Y: 6}); got != 1 {
		t.Errorf("RoomAt(5,6) = %d, want 1", got)
	}
	if got := l.RoomAt(Coord{X: 4, Y: 4}); got != -1 {
		t.Errorf("RoomAt(4,4) = %d, want -1", got)
	}
}

func TestLevelString(t *testing.T) {
	l, err := Generate(Config{Width: 24, Height: 20, Seed: 4})
	if err != nil {
		t.Fatal(err)
	}
	s := l.String()
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) != l.H {
		t.Fatalf("String() has %d lines, want %d", len(lines), l.H)
	}
	if !strings.ContainsRune(s, 'S') || !strings.ContainsRune(s, 'E') {
		t.Error("String() should mark spawn and exit")
	}
	if !strings.ContainsRune(s, '~') {
		t.Error("String() should contain vent cells")
	}
	if !strings.ContainsRune(s, '!') {
		t.Error("String() should contain consoles")
	}
}

func TestStepsBetweenWalksDoors(t *testing.T) {
	// A straight corridor with a door in the middle: the step distance counts
	// through it, since walkability treats doors as passable.
	l := newTestLevel(7, 3)
	for x := 1; x <= 5; x++ {
		l.Set(x, 1, TileFloor)
	}
	l.Set(3, 1, TileDoor)
	src, dst := Coord{X: 1, Y: 1}, Coord{X: 5, Y: 1}
	if got := StepsBetween(l, src, dst); got != 4 {
		t.Errorf("StepsBetween = %d, want 4", got)
	}
}
