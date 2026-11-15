package telemetry

import (
	"strings"
	"testing"

	"github.com/danielriddell21/nemesis/internal/sim"
	"github.com/danielriddell21/nemesis/internal/world"
)

type capture struct {
	events []Event
}

func (c *capture) OnEvent(e Event) { c.events = append(c.events, e) }

func TestBusForwardsAndFilters(t *testing.T) {
	c := &capture{}
	b := NewBus(c, nil)
	b.Observe(sim.Observation{Kind: sim.ObsStep, At: world.Coord{X: 1, Y: 1}})
	b.Observe(sim.Observation{Kind: sim.ObsVentCreak, At: world.Coord{X: 12, Y: 4}, Radius: 6})
	if len(c.events) != 1 {
		t.Fatalf("got %d events, want footsteps filtered and the creak kept", len(c.events))
	}
	if c.events[0].Type != "vent-creak" || c.events[0].X != 12 || c.events[0].Y != 4 {
		t.Errorf("unexpected event %+v", c.events[0])
	}
	if got := len(b.Recent()); got != 1 {
		t.Errorf("recent feed has %d entries, want 1", got)
	}
}

func TestBusBoundsFeed(t *testing.T) {
	b := NewBus()
	for range feedDepth * 2 {
		b.Observe(sim.Observation{Kind: sim.ObsDoorOpen})
	}
	if got := len(b.Recent()); got != feedDepth {
		t.Errorf("feed grew to %d, want capped at %d", got, feedDepth)
	}
}

func TestEventLines(t *testing.T) {
	tests := []struct {
		obs  sim.Observation
		want string
	}{
		{sim.Observation{Kind: sim.ObsVentCreak, At: world.Coord{X: 12, Y: 4}}, "VENT CREAK AT 12 4"},
		{sim.Observation{Kind: sim.ObsAlienCreak, At: world.Coord{X: 3, Y: 9}}, "VENT CREAK AT 3 9"},
		{sim.Observation{Kind: sim.ObsDoorOpen, At: world.Coord{X: 5, Y: 6}}, "BULKHEAD OPENED AT 5 6"},
		{sim.Observation{Kind: sim.ObsConsole, At: world.Coord{X: 7, Y: 2}}, "GENERATOR ONLINE AT 7 2"},
		{sim.Observation{Kind: sim.ObsAlienState, State: sim.StateHunt, Target: world.Coord{X: 8, Y: 8}}, "HUNTER hunt -> 8 8"},
		{sim.Observation{Kind: sim.ObsAlienHeard, Target: world.Coord{X: 2, Y: 3}}, "HUNTER HEARD NOISE AT 2 3"},
		{sim.Observation{Kind: sim.ObsDirectorNudge, Target: world.Coord{X: 4, Y: 4}}, "DIRECTOR STEERS HUNT TO 4 4"},
		{sim.Observation{Kind: sim.ObsEscape}, "PREY ESCAPED THROUGH THE AIRLOCK"},
		{sim.Observation{Kind: sim.ObsAlienLearn, Learn: sim.LearnPing, Tier: 2}, "HUNTER LEARNED TRACKER II"},
		{sim.Observation{Kind: sim.ObsAlienLearn, Learn: sim.LearnVent, Tier: 1}, "HUNTER LEARNED VENTS I"},
		{sim.Observation{Kind: sim.ObsAlienLearn, Learn: sim.LearnSearch, Tier: 4}, "HUNTER LEARNED SEARCH IV"},
	}
	for _, tt := range tests {
		if got := FromObservation(tt.obs).Line(); got != tt.want {
			t.Errorf("Line() = %q, want %q", got, tt.want)
		}
	}
	if line := FromObservation(sim.Observation{Kind: sim.ObsStep}).Line(); line != "" {
		t.Errorf("steps should have no feed line, got %q", line)
	}
}

func TestLiveGameFeedsBus(t *testing.T) {
	l, err := world.Generate(world.Config{Width: 32, Height: 24, Seed: 5})
	if err != nil {
		t.Fatal(err)
	}
	c := &capture{}
	g := sim.New(l, sim.WithObserver(NewBus(c)))
	for range 60 {
		g.Tick(sim.Input{Tracker: true}, 1.0/60)
	}
	found := false
	for _, e := range c.events {
		if e.Type == "ping" {
			found = true
			if !strings.HasPrefix(e.Line(), "TRACKER PING AT ") {
				t.Errorf("ping line %q malformed", e.Line())
			}
		}
	}
	if !found {
		t.Error("a raised tracker should feed ping events through the bus")
	}
}

func TestTorchEventsRoundTripThroughTheBus(t *testing.T) {
	cases := []struct {
		obs  sim.Observation
		want string
	}{
		// The switch is flicked constantly and triggers nothing by itself, so
		// it stays out of the hunter's feed.
		{sim.Observation{Kind: sim.ObsTorch, At: world.Coord{X: 3, Y: 4}}, ""},
		{sim.Observation{Kind: sim.ObsBattery, At: world.Coord{X: 5, Y: 6}}, "TORCH CELL TAKEN AT 5 6"},
		{sim.Observation{Kind: sim.ObsAlienGlow, At: world.Coord{X: 1, Y: 1}, Target: world.Coord{X: 7, Y: 8}}, "HUNTER NOTICED TORCHLIGHT AT 7 8"},
	}
	for _, c := range cases {
		e := FromObservation(c.obs)
		if got := e.Line(); got != c.want {
			t.Errorf("Line() = %q, want %q", got, c.want)
		}
		if kind, ok := e.Kind(); !ok || kind != c.obs.Kind {
			t.Errorf("Kind() of %q = %v,%v, want %v", e.Type, kind, ok, c.obs.Kind)
		}
	}
}

func TestEveryObservationKindHasANameAndRoundTrips(t *testing.T) {
	for k := sim.ObsStep; k < sim.ObsKindCount; k++ {
		if k.String() == "unknown" {
			t.Errorf("kind %d has no name", k)
		}
		if got, ok := FromObservation(sim.Observation{Kind: k}).Kind(); !ok || got != k {
			t.Errorf("kind %v does not survive the event round trip: %v,%v", k, got, ok)
		}
	}
}
