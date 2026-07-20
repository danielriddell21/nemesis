package sim

import (
	"math"

	"github.com/danielriddell21/nemesis/internal/world"
)

type AlienState uint8

const (
	StateLurk AlienState = iota
	StatePatrol
	StateInvestigate
	StateHunt
	StateSearch
)

func (s AlienState) String() string {
	switch s {
	case StateLurk:
		return "lurk"
	case StatePatrol:
		return "patrol"
	case StateInvestigate:
		return "investigate"
	case StateHunt:
		return "hunt"
	case StateSearch:
		return "search"
	default:
		return "unknown"
	}
}

const (
	patrolSpeed      = 1.7
	investigateSpeed = 2.3
	searchSpeed      = 2.0
	huntSpeed        = 4.4
	ventSpeedBoost   = 1.5

	huntRepath   = 0.4
	doorShoulder = 0.8
	arriveDwell  = 2.5
	searchRounds = 3
)

type Alien struct {
	Pos    Vec2
	Facing float64
	State  AlienState
	Target world.Coord
	Moving bool

	path          []world.Coord
	pathIdx       int
	repath        float64
	dwell         float64
	searchesLeft  int
	lastKnown     Vec2
	tracked       bool
	wasVent       bool
	ventChecked   bool
	lockerChecked bool
}

func newAlien(pos Vec2) Alien {
	return Alien{Pos: pos, State: StateLurk, Target: pos.Cell()}
}

func (a *Alien) InVent(w *World) bool {
	c := a.Pos.Cell()
	return w.Level.At(c.X, c.Y) == world.TileVent
}

func (a *Alien) Path() []world.Coord {
	if a.pathIdx >= len(a.path) {
		return nil
	}
	return a.path[a.pathIdx:]
}

func (g *Game) setAlienState(s AlienState) {
	if g.Alien.State == s {
		return
	}
	g.Alien.State = s
	g.observe(Observation{Kind: ObsAlienState, At: g.Alien.Pos.Cell(), State: s, Target: g.Alien.Target})
}

func (g *Game) setAlienTarget(c world.Coord) {
	g.Alien.Target = nearestWalkable(g.World.Level, c)
	g.Alien.path = nil
	g.Alien.pathIdx = 0
	g.Alien.repath = 0
}

func (g *Game) tickAlien(dt float64, noise noiseEvent) {
	a := &g.Alien
	g.senseTick(noise)

	if a.dwell > 0 {
		a.dwell -= dt
		a.Moving = false
		g.sweepFacing(dt)
		return
	}

	switch a.State {
	case StateLurk:
		a.Moving = false
		return
	case StateHunt:
		g.tickHunt(dt)
	case StateInvestigate, StateSearch, StatePatrol:
		g.tickTravel()
	}
	g.moveAlien(dt)
}

func (g *Game) senseTick(noise noiseEvent) {
	a := &g.Alien
	if g.alienSees() {
		if a.State != StateHunt {
			g.observe(Observation{Kind: ObsAlienSeen, At: g.Player.Pos.Cell()})
			g.depositHeat(g.Player.Pos.Cell())
			g.setAlienState(StateHunt)
		}
		a.lastKnown = g.Player.Pos
		a.tracked = true
		return
	}
	if a.State == StateHunt {
		return
	}
	if g.alienHears(noise) {
		c := noise.at.Cell()
		g.observe(Observation{Kind: ObsAlienHeard, At: a.Pos.Cell(), Target: c, Radius: noise.radius})
		g.hearNoise(noise.kind, c)
		if a.State == StateLurk {
			g.director.wake = 0
		}
		if g.pingRush(noise) {
			return
		}
		g.setAlienTarget(c)
		g.setAlienState(StateInvestigate)
	}
}

func (g *Game) pingRush(noise noiseEvent) bool {
	// A hunter that has heard enough pings knows exactly what that chirp is:
	// close ones are answered with a dead sprint, not curiosity.
	if noise.kind != ObsTrackerPing || g.learning.pingTier() < 2 {
		return false
	}
	if noise.at.Sub(g.Alien.Pos).Len() > pingRushRange {
		return false
	}
	g.Alien.lastKnown = noise.at
	g.Alien.tracked = false
	g.setAlienState(StateHunt)
	return true
}

func (g *Game) tickHunt(dt float64) {
	a := &g.Alien
	a.repath -= dt
	if a.tracked {
		// While the trail is warm the hunter re-paths straight at the player.
		if a.repath <= 0 {
			a.repath = huntRepath
			a.path = findPath(g.World, a.Pos.Cell(), g.Player.Pos.Cell())
			a.pathIdx = 0
		}
		if !g.alienSees() && g.Player.Pos.Sub(a.lastKnown).Len() > 1.5 {
			a.tracked = false
		}
		return
	}
	// Trail cold: head to the last known position, then start searching.
	if a.pathIdx >= len(a.path) {
		last := a.lastKnown.Cell()
		if a.Pos.Cell() == last {
			a.searchesLeft = g.searchRoundsLearned()
			a.dwell = g.searchDwell()
			a.ventChecked = false
			a.lockerChecked = false
			g.coldTrail()
			g.setAlienState(StateSearch)
			return
		}
		g.setAlienTarget(last)
		a.path = findPath(g.World, a.Pos.Cell(), last)
		a.pathIdx = 0
	}
}

func (g *Game) tickTravel() {
	a := &g.Alien
	if a.pathIdx < len(a.path) {
		return
	}
	if a.Pos.Cell() != a.Target {
		a.path = findPath(g.World, a.Pos.Cell(), a.Target)
		a.pathIdx = 0
		if len(a.path) > 0 {
			return
		}
	}
	g.arrive()
}

func (g *Game) arrive() {
	a := &g.Alien
	switch a.State {
	case StateInvestigate:
		a.searchesLeft = g.searchRoundsLearned()
		a.dwell = g.searchDwell()
		a.ventChecked = false
		a.lockerChecked = false
		g.setAlienState(StateSearch)
	case StateSearch:
		if a.searchesLeft > 0 {
			a.searchesLeft--
			g.setAlienTarget(g.nextSearchTarget())
			a.dwell = g.searchDwell() / 2
			return
		}
		g.setAlienState(StatePatrol)
		g.setAlienTarget(g.patrolPoint())
	case StatePatrol:
		a.dwell = g.searchDwell() / 2
		g.setAlienTarget(g.patrolPoint())
	}
}

func (g *Game) nextSearchTarget() world.Coord {
	// The first sweeps of a search go to the places the hunter has learned prey
	// disappears — the vents, then the lockers — before it fans out at random.
	a := &g.Alien
	if !a.ventChecked {
		a.ventChecked = true
		if m, ok := g.ventCheckPoint(); ok {
			return m
		}
	}
	if !a.lockerChecked {
		a.lockerChecked = true
		if m, ok := g.lockerCheckPoint(); ok {
			return m
		}
	}
	return g.searchPoint(a.Pos.Cell())
}

func (g *Game) searchPoint(near world.Coord) world.Coord {
	l := g.World.Level
	for range 16 {
		c := world.Coord{
			X: near.X + g.rng.IntN(9) - 4,
			Y: near.Y + g.rng.IntN(9) - 4,
		}
		if l.At(c.X, c.Y).Walkable() {
			return c
		}
	}
	return near
}

func (g *Game) patrolPoint() world.Coord {
	l := g.World.Level
	if len(l.Rooms) == 0 {
		return g.Alien.Pos.Cell()
	}
	// Prefer a room the director has hinted at, then the habits the hunter has
	// learned; otherwise roam anywhere but the room it is already in.
	if g.director.hintValid {
		g.director.hintValid = false
		return g.director.hint
	}
	if c, ok := g.hotRoom(); ok && g.rng.Float64() < 0.5 {
		return c
	}
	cur := l.RoomAt(g.Alien.Pos.Cell())
	for range 8 {
		i := g.rng.IntN(len(l.Rooms))
		if i != cur {
			return l.Rooms[i].Center()
		}
	}
	return l.Rooms[g.rng.IntN(len(l.Rooms))].Center()
}

func (g *Game) moveAlien(dt float64) {
	a := &g.Alien
	if a.pathIdx >= len(a.path) {
		a.Moving = false
		return
	}
	next := a.path[a.pathIdx]
	if g.World.Level.At(next.X, next.Y) == world.TileDoor && !g.World.DoorOpen(next) {
		// Shoulder the bulkhead open: a pause, then a bang the player can hear.
		a.Moving = false
		a.dwell = doorShoulder
		g.World.OpenDoor(next)
		g.observe(Observation{Kind: ObsDoorOpen, At: next, Radius: doorNoise})
		return
	}
	dest := cellCenter(next)
	to := dest.Sub(a.Pos)
	dist := to.Len()
	speed := g.alienSpeed()
	step := speed * dt
	a.Moving = true
	a.Facing = math.Atan2(to.Y, to.X)
	if dist <= step {
		a.Pos = dest
		a.pathIdx++
		g.creakOnVentCrossing(next)
		return
	}
	a.Pos = Vec2{X: a.Pos.X + to.X/dist*step, Y: a.Pos.Y + to.Y/dist*step}
}

func (g *Game) creakOnVentCrossing(c world.Coord) {
	a := &g.Alien
	inVent := g.World.Level.At(c.X, c.Y) == world.TileVent
	if inVent != a.wasVent {
		// The duct announces the hunter at the grate: entering or leaving a
		// vent creaks, and the visualiser logs where.
		g.observe(Observation{Kind: ObsAlienCreak, At: c, Radius: crawlNoise})
	}
	a.wasVent = inVent
}

func (g *Game) alienSpeed() float64 {
	var s float64
	switch g.Alien.State {
	case StateHunt:
		s = huntSpeed
	case StateInvestigate:
		s = investigateSpeed
	case StateSearch:
		s = searchSpeed
	default:
		s = patrolSpeed
	}
	if g.Alien.InVent(g.World) {
		s *= ventSpeedBoost
	}
	return s * g.threat
}

func (g *Game) sweepFacing(dt float64) {
	// While dwelling the hunter scans its surroundings.
	g.Alien.Facing = normalizeAngle(g.Alien.Facing + 1.8*dt)
}
