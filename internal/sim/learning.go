package sim

import "github.com/danielriddell21/nemesis/internal/world"

const (
	pingTier1     = 3
	pingTier2     = 8
	creakTier1    = 4
	creakTier2    = 10
	maxSearchTier = 4

	pingRushRange  = 9.0
	ventCheckRange = 10.0

	heatDeposit  = 1.0
	heatHalfLife = 45.0
)

type Learned struct {
	Pings  int `json:"pings"`
	Creaks int `json:"creaks"`
	Cold   int `json:"cold"`
}

type learning struct {
	Learned
	heat []float64
}

func newLearning(carried Learned, rooms int) learning {
	return learning{Learned: carried, heat: make([]float64, rooms)}
}

func (l *learning) pingTier() int {
	switch {
	case l.Pings >= pingTier2:
		return 2
	case l.Pings >= pingTier1:
		return 1
	default:
		return 0
	}
}

func (l *learning) ventTier() int {
	switch {
	case l.Creaks >= creakTier2:
		return 2
	case l.Creaks >= creakTier1:
		return 1
	default:
		return 0
	}
}

func (l *learning) searchTier() int {
	if l.Cold > maxSearchTier {
		return maxSearchTier
	}
	return l.Cold
}

func (l *learning) decay(dt float64) {
	k := 1 - dt/heatHalfLife*0.693
	if k < 0 {
		k = 0
	}
	for i := range l.heat {
		l.heat[i] *= k
	}
}

func (g *Game) depositHeat(c world.Coord) {
	if ri := g.World.Level.RoomAt(c); ri >= 0 && ri < len(g.learning.heat) {
		g.learning.heat[ri] += heatDeposit
	}
}

func (g *Game) hearNoise(kind ObservationKind, at world.Coord) {
	l := &g.learning
	switch kind {
	case ObsTrackerPing:
		l.Pings++
		g.noteLearn(LearnPing, l.pingTier(), l.Pings == pingTier1 || l.Pings == pingTier2)
	case ObsVentCreak:
		l.Creaks++
		g.noteLearn(LearnVent, l.ventTier(), l.Creaks == creakTier1 || l.Creaks == creakTier2)
	}
	g.depositHeat(at)
}

func (g *Game) coldTrail() {
	g.learning.Cold++
	g.noteLearn(LearnSearch, g.learning.searchTier(), g.learning.Cold <= maxSearchTier)
}

func (g *Game) noteLearn(what LearnKind, tier int, unlocked bool) {
	if unlocked {
		g.observe(Observation{Kind: ObsAlienLearn, At: g.Alien.Pos.Cell(), Learn: what, Tier: tier})
	}
}

func (g *Game) Learned() Learned {
	return g.learning.Learned
}

func (g *Game) LearnTiers() (ping, vent, search int) {
	return g.learning.pingTier(), g.learning.ventTier(), g.learning.searchTier()
}

func (g *Game) RoomHeat() []float64 {
	return g.learning.heat
}

func (g *Game) searchRoundsLearned() int {
	return searchRounds + g.learning.searchTier()
}

func (g *Game) searchDwell() float64 {
	// A hunter that has lost prey before sweeps faster and lingers less
	// between rounds — practice makes the search sharper.
	d := arriveDwell * (1 - 0.15*float64(g.learning.searchTier()))
	if d < 1 {
		return 1
	}
	return d
}

func (g *Game) ventCheckPoint() (world.Coord, bool) {
	// A duct-literate hunter checks the nearest grate while searching.
	if g.learning.ventTier() == 0 {
		return world.Coord{}, false
	}
	best, bestD, found := world.Coord{}, ventCheckRange, false
	for _, m := range g.World.Level.VentMouths {
		if d := cellCenter(m).Sub(g.Alien.Pos).Len(); d < bestD {
			best, bestD, found = m, d, true
		}
	}
	return best, found
}

func (g *Game) hotRoom() (world.Coord, bool) {
	// The heat map is where prey keeps being detected; patrols drift there.
	l := g.World.Level
	bestI, bestH := -1, 0.5
	for i, h := range g.learning.heat {
		if h > bestH {
			bestI, bestH = i, h
		}
	}
	if bestI < 0 || bestI >= len(l.Rooms) {
		return world.Coord{}, false
	}
	return l.Rooms[bestI].Center(), true
}
