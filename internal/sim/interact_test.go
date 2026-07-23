package sim

import (
	"math"
	"testing"

	"github.com/danielriddell21/nemesis/internal/world"
)

func TestFacingInteractable(t *testing.T) {
	l := flatLevel(16, 12)
	console := world.Coord{X: 4, Y: 3}
	l.Tiles[console.Y*16+console.X] = world.TileConsole
	g := New(l)

	// Stand just west of the console, facing east toward it.
	g.Player.Pos = Vec2{X: float64(console.X) - 0.4, Y: float64(console.Y) + 0.5}
	g.Player.Angle = 0 // east
	if got := g.FacingInteractable(); got != InteractConsole {
		t.Fatalf("facing the console = %v, want InteractConsole", got)
	}

	// Facing away, nothing is in reach.
	g.Player.Angle = math.Pi
	if got := g.FacingInteractable(); got != InteractNone {
		t.Errorf("facing away = %v, want InteractNone", got)
	}

	// Once activated, the console no longer prompts.
	g.Player.Angle = 0
	g.activated[console] = true
	if got := g.FacingInteractable(); got != InteractNone {
		t.Errorf("activated console = %v, want InteractNone", got)
	}
}
