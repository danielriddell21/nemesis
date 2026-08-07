package vis

import (
	"testing"

	"github.com/danielriddell21/nemesis/internal/telemetry"
	"github.com/danielriddell21/nemesis/internal/world"
)

func TestVisModelHelloRegrowsLevel(t *testing.T) {
	m := newVisModel()
	m.apply(Msg{Type: "hello", Seed: 11, Width: 32, Height: 24, Consoles: 2})
	if m.level == nil {
		t.Fatal("hello should regenerate the level")
	}
	want, err := world.Generate(world.Config{Width: 32, Height: 24, Seed: 11, Consoles: 2})
	if err != nil {
		t.Fatal(err)
	}
	if m.level.String() != want.String() {
		t.Error("visualiser level differs from the game's for the same seed")
	}
	if len(m.feed) == 0 {
		t.Error("hello should log a run-start line")
	}
}

func TestVisModelStateAndEvents(t *testing.T) {
	m := newVisModel()
	m.apply(Msg{Type: "hello", Seed: 1, Width: 24, Height: 20})
	m.apply(Msg{Type: "state", State: &StateMsg{Tick: 5, AlienState: "hunt", Menace: 0.9}})
	if !m.haveState || m.state.AlienState != "hunt" {
		t.Fatalf("state not applied: %+v", m.state)
	}
	m.apply(Msg{Type: "events", Events: []telemetry.Event{
		{Type: "vent-creak", X: 12, Y: 4, Radius: 6},
		{Type: "console", X: 3, Y: 7, Radius: 12},
		{Type: "step", X: 1, Y: 1},
	}})
	if len(m.ripples) != 2 {
		t.Errorf("noise events should ripple, got %d", len(m.ripples))
	}
	if !m.activated[world.Coord{X: 3, Y: 7}] {
		t.Error("console event should mark the console active")
	}
	found := false
	for _, line := range m.feed {
		if line == "VENT CREAK AT 12 4" {
			found = true
		}
	}
	if !found {
		t.Errorf("feed missing creak line: %v", m.feed)
	}
}

func TestVisModelRipplesAge(t *testing.T) {
	m := newVisModel()
	m.applyEvent(telemetry.Event{Type: "ping", X: 2, Y: 2, Radius: 5})
	for range int(rippleLife/0.016) + 2 {
		m.tick(0.016)
	}
	if len(m.ripples) != 0 {
		t.Errorf("ripples should expire, %d left", len(m.ripples))
	}
}

func TestVisModelFeedBounded(t *testing.T) {
	m := newVisModel()
	for range visFeedLines * 3 {
		m.applyEvent(telemetry.Event{Type: "door", X: 1, Y: 1})
	}
	if len(m.feed) != visFeedLines {
		t.Errorf("feed length %d, want capped at %d", len(m.feed), visFeedLines)
	}
}
