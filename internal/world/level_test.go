package world

import (
	"strings"
	"testing"
)

func TestTileWalkable(t *testing.T) {
	tests := []struct {
		tile TileType
		want bool
	}{
		{TileFloor, true},
		{TileDoor, true},
		{TileVent, true},
		{TileSpawn, true},
		{TileExit, true},
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
	l := newLevel(4, 4, 0)
	if l.At(-1, 0) != TileWall || l.At(0, -1) != TileWall || l.At(4, 0) != TileWall || l.At(0, 4) != TileWall {
		t.Error("out-of-bounds reads should return TileWall")
	}
	if l.LightAt(-1, 0) != 0 {
		t.Error("out-of-bounds light should be dark")
	}
	if l.FlickerAt(-1, 0) {
		t.Error("out-of-bounds flicker should be false")
	}
}

func TestLevelSolid(t *testing.T) {
	l := newLevel(4, 4, 0)
	l.set(1, 1, TileFloor)
	l.set(2, 1, TileDoor)
	l.set(1, 2, TileConsole)
	l.set(2, 2, TileVent)
	if l.Solid(1, 1) || l.Solid(2, 2) {
		t.Error("floor and vent should not be solid")
	}
	if !l.Solid(2, 1) || !l.Solid(1, 2) || !l.Solid(0, 0) {
		t.Error("door, console and wall should be solid")
	}
}

func TestRoomAt(t *testing.T) {
	l := newLevel(8, 8, 0)
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
	if len(lines) != l.Height {
		t.Fatalf("String() has %d lines, want %d", len(lines), l.Height)
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

func TestReachableTreatsDoorsSolid(t *testing.T) {
	// A corridor blocked by a door: Reachable (doors closed) says no,
	// StepsBetween (doors open) finds the path.
	l := newLevel(7, 3, 0)
	for x := 1; x <= 5; x++ {
		l.set(x, 1, TileFloor)
	}
	l.set(3, 1, TileDoor)
	src, dst := Coord{X: 1, Y: 1}, Coord{X: 5, Y: 1}
	if Reachable(l, src, dst) {
		t.Error("closed door should block Reachable")
	}
	if got := StepsBetween(l, src, dst); got != 4 {
		t.Errorf("StepsBetween = %d, want 4", got)
	}
}
