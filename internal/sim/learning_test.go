package sim

import (
	"math"
	"testing"

	"github.com/danielriddell21/nemesis/internal/world"
)

func TestLearnedCarriesIntoNewGame(t *testing.T) {
	carried := Learned{Pings: 9, Creaks: 11, Cold: 6}
	g := New(flatLevel(16, 12), WithLearned(carried))
	if g.Learned() != carried {
		t.Fatalf("got %+v, want carried %+v", g.Learned(), carried)
	}
	ping, vent, search := g.LearnTiers()
	if ping != 2 || vent != 2 || search != maxSearchTier {
		t.Errorf("tiers = %d/%d/%d, want 2/2/%d", ping, vent, search, maxSearchTier)
	}
}

func TestHunterLearnsPings(t *testing.T) {
	rec := &recorder{}
	g := New(flatLevel(24, 12), WithObserver(rec))
	// Hold the tracker with the hunter in earshot; pin it in place so it
	// cannot close in or spot the player while it racks up pings.
	pin := Vec2{X: 4.5, Y: 1.5}
	for range int(pingTier1*pingInterval*60) + 120 {
		g.Alien.Pos = pin
		g.Alien.Facing = 0
		g.Alien.State = StatePatrol
		g.Tick(Input{Tracker: true}, 1.0/60)
	}
	if g.Learned().Pings < pingTier1 {
		t.Fatalf("hunter heard %d pings, want >= %d", g.Learned().Pings, pingTier1)
	}
	if rec.count(ObsAlienLearn) == 0 {
		t.Error("crossing a learning tier should be observed")
	}
}

func TestPingRushAtTierTwo(t *testing.T) {
	g := New(flatLevel(24, 12), WithLearned(Learned{Pings: pingTier2}))
	g.Alien.Pos = Vec2{X: 4.5, Y: 1.5}
	g.Alien.Facing = 0 // cannot see the player
	g.Alien.State = StatePatrol
	for range int(pingInterval*60) + 10 {
		g.Tick(Input{Tracker: true}, 1.0/60)
		if g.Dead() {
			return // rushed in and made contact: learning worked
		}
	}
	if g.Alien.State != StateHunt {
		t.Errorf("a ping-literate hunter should rush the chirp, got %v", g.Alien.State)
	}
}

func TestNaiveHunterInvestigatesPings(t *testing.T) {
	g := New(flatLevel(24, 12))
	g.Alien.Pos = Vec2{X: 4.5, Y: 1.5}
	g.Alien.Facing = 0
	g.Alien.State = StatePatrol
	g.Tick(Input{Tracker: true}, 1.0/60)
	if g.Alien.State == StateHunt {
		t.Error("a naive hunter should investigate a ping, not hunt it")
	}
}

func TestColdTrailsSharpenSearches(t *testing.T) {
	g := New(flatLevel(16, 12))
	if g.searchRoundsLearned() != searchRounds {
		t.Fatalf("naive rounds = %d, want %d", g.searchRoundsLearned(), searchRounds)
	}
	base := g.searchDwell()
	for range 3 {
		g.coldTrail()
	}
	if g.searchRoundsLearned() != searchRounds+3 {
		t.Errorf("rounds after 3 cold trails = %d, want %d", g.searchRoundsLearned(), searchRounds+3)
	}
	if g.searchDwell() >= base {
		t.Error("cold trails should shorten the search dwell")
	}
	for range 20 {
		g.coldTrail()
	}
	if g.searchRoundsLearned() != searchRounds+maxSearchTier {
		t.Errorf("rounds should cap at %d, got %d", searchRounds+maxSearchTier, g.searchRoundsLearned())
	}
	if g.searchDwell() < 1 {
		t.Error("search dwell should never collapse below a second")
	}
}

func TestVentCheckPoint(t *testing.T) {
	l := flatLevel(24, 12)
	l.Tiles[5*24+10] = world.TileVent
	l.VentMouths = []world.Coord{{X: 10, Y: 5}}
	g := New(l)
	g.Alien.Pos = Vec2{X: 8.5, Y: 5.5}
	if _, ok := g.ventCheckPoint(); ok {
		t.Fatal("a naive hunter should not check vents")
	}
	g.learning.Creaks = creakTier1
	m, ok := g.ventCheckPoint()
	if !ok || m != (world.Coord{X: 10, Y: 5}) {
		t.Errorf("duct-literate hunter should check the nearby grate, got %v %v", m, ok)
	}
	g.Alien.Pos = Vec2{X: 22.5, Y: 10.5} // 13 tiles out, beyond ventCheckRange
	if _, ok := g.ventCheckPoint(); ok {
		t.Error("grates beyond range should not be checked")
	}
}

func TestHeatMapDrawsPatrols(t *testing.T) {
	l, err := world.Generate(world.Config{Width: 32, Height: 24, Seed: 21})
	if err != nil {
		t.Fatal(err)
	}
	g := New(l)
	if _, ok := g.hotRoom(); ok {
		t.Fatal("fresh hunter should have no hot rooms")
	}
	room := l.Rooms[0]
	for range 3 {
		g.depositHeat(room.Center())
	}
	c, ok := g.hotRoom()
	if !ok || c != room.Center() {
		t.Fatalf("hot room = %v %v, want %v", c, ok, room.Center())
	}
	// Heat decays back to nothing over time.
	for range 60 * 240 {
		g.learning.decay(1.0 / 60)
	}
	if _, ok := g.hotRoom(); ok {
		t.Error("heat should decay away")
	}
}

func TestHeatFromSighting(t *testing.T) {
	l, err := world.Generate(world.Config{Width: 32, Height: 24, Seed: 3})
	if err != nil {
		t.Fatal(err)
	}
	g := New(l)
	// Stand the hunter in the spawn room facing the player point-blank.
	g.Alien.Pos = Vec2{X: g.Player.Pos.X + 1.5, Y: g.Player.Pos.Y}
	g.Alien.Facing = math.Pi
	g.Tick(Input{}, 1.0/60)
	total := 0.0
	for _, h := range g.RoomHeat() {
		total += h
	}
	if total <= 0 {
		t.Error("a sighting should deposit heat")
	}
}
