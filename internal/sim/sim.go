package sim

import (
	"math"
	"math/rand/v2"

	"github.com/danielriddell21/nemesis/internal/world"
)

type Vec2 struct {
	X, Y float64
}

func (v Vec2) Sub(o Vec2) Vec2 { return Vec2{X: v.X - o.X, Y: v.Y - o.Y} }

func (v Vec2) Len() float64 { return math.Hypot(v.X, v.Y) }

func (v Vec2) Cell() world.Coord {
	return world.Coord{X: int(math.Floor(v.X)), Y: int(math.Floor(v.Y))}
}

func cellCenter(c world.Coord) Vec2 {
	return Vec2{X: float64(c.X) + 0.5, Y: float64(c.Y) + 0.5}
}

type MoveMode uint8

const (
	ModeWalk MoveMode = iota
	ModeSneak
	ModeRun
)

type Input struct {
	Forward, Strafe float64
	Turn, TurnDelta float64
	Mode            MoveMode
	Tracker         bool
	Use             bool
	Hide            bool
	Throw           bool
}

type Game struct {
	World   *World
	Player  Player
	Alien   Alien
	Tracker Tracker

	tick     uint64
	elapsed  float64
	depth    int
	threat   float64
	rng      *rand.Rand
	observer Observer
	director director

	activated map[world.Coord]bool
	learning  learning
	decoy     Decoy
	escaped   bool
	dead      bool

	useLatch   bool
	hideLatch  bool
	throwLatch bool
}

type Option func(*Game)

func WithObserver(o Observer) Option {
	return func(g *Game) {
		if o != nil {
			g.observer = o
		}
	}
}

func WithLearned(l Learned) Option {
	return func(g *Game) {
		g.learning.Learned = l
	}
}

func WithDepth(d int) Option {
	return func(g *Game) {
		if d > 0 {
			g.depth = d
		}
	}
}

func New(l *world.Level, opts ...Option) *Game {
	u := uint64(l.Seed)
	g := &Game{
		World: NewWorld(l),
		Player: Player{
			Pos:    cellCenter(l.Spawn),
			Angle:  spawnFacing(l),
			Decoys: startingDecoys,
		},
		Tracker:   newTracker(),
		threat:    1,
		rng:       rand.New(rand.NewPCG(u^0xa5a5a5a5, u+0x9e3779b97f4a7c15)),
		observer:  nopObserver{},
		activated: make(map[world.Coord]bool),
		learning:  newLearning(Learned{}, len(l.Rooms)),
		director:  newDirector(),
	}
	for _, opt := range opts {
		opt(g)
	}
	g.applyDepth()
	g.Alien = newAlien(alienStart(l))
	return g
}

func (g *Game) applyDepth() {
	if g.depth <= 0 {
		return
	}
	// Deeper decks wake the hunter sooner, start it warier, and quicken it a
	// touch — the station gets less forgiving the further you descend.
	g.threat = 1 + 0.04*float64(g.depth)
	g.director.wake = math.Max(6, lurkTime-float64(g.depth)*4)
	g.director.aggression = math.Min(0.5, 0.1*float64(g.depth))
}

func spawnFacing(l *world.Level) float64 {
	d := cellCenter(l.Exit).Sub(cellCenter(l.Spawn))
	return math.Atan2(d.Y, d.X)
}

func alienStart(l *world.Level) Vec2 {
	// The hunter wakes in the vent mouth farthest from the spawn, so the
	// opening moments are quiet.
	best, bestDist := l.Exit, -1
	for _, m := range l.VentMouths {
		if d := world.StepsBetween(l, l.Spawn, m); d > bestDist {
			best, bestDist = m, d
		}
	}
	return cellCenter(best)
}

func (g *Game) Tick(in Input, dt float64) {
	if g.escaped || g.dead {
		return
	}
	g.tick++
	g.elapsed += dt

	g.learning.decay(dt)
	noise := g.tickPlayer(in, dt)
	g.tickTracker(in, dt, &noise)
	g.tickHide(in)
	if in.Use && !g.useLatch && !g.Player.Hidden {
		g.interact(&noise)
	}
	g.useLatch = in.Use
	g.tickThrow(in)
	noise.merge(g.tickDecoy(dt))

	g.tickAlien(dt, noise)
	g.checkDecoyReached()
	g.director.tick(g, dt)
	g.checkKill()
	g.checkEscape()
}

func (g *Game) Escaped() bool { return g.escaped }

func (g *Game) Dead() bool { return g.dead }

func (g *Game) Elapsed() float64 { return g.elapsed }

func (g *Game) TickCount() uint64 { return g.tick }

func (g *Game) Menace() float64 { return g.director.menace }

func (g *Game) ObjectivesDone() int { return len(g.activated) }

func (g *Game) ObjectivesTotal() int { return len(g.World.Level.Consoles) }

func (g *Game) ExitUnlocked() bool { return g.ObjectivesDone() >= g.ObjectivesTotal() }

func (g *Game) ConsoleActivated(c world.Coord) bool { return g.activated[c] }

func (g *Game) observe(o Observation) {
	o.Tick = g.tick
	g.observer.Observe(o)
}

func (g *Game) checkKill() {
	if g.dead {
		return
	}
	d := g.Alien.Pos.Sub(g.Player.Pos).Len()
	if g.Player.Hidden {
		// A locker only hides you until the hunter reaches it and wrenches the
		// door open.
		if d < lockerBreach {
			g.die()
		}
		return
	}
	if d < killRange && g.lineOfSight(g.Alien.Pos, g.Player.Pos) {
		g.die()
	}
}

func (g *Game) die() {
	g.dead = true
	g.observe(Observation{Kind: ObsDeath, At: g.Player.Pos.Cell()})
}

func (g *Game) Depth() int { return g.depth }

func (g *Game) VisionRange() float64 { return g.effectiveVisionRange() }

func (g *Game) VisionFOV() float64 { return visionFOV }

func (g *Game) checkEscape() {
	if g.escaped || g.dead {
		return
	}
	c := g.Player.Pos.Cell()
	if g.World.Level.At(c.X, c.Y) == world.TileExit && g.ExitUnlocked() {
		g.escaped = true
		g.observe(Observation{Kind: ObsEscape, At: c})
	}
}

func normalizeAngle(a float64) float64 {
	for a < -math.Pi {
		a += 2 * math.Pi
	}
	for a > math.Pi {
		a -= 2 * math.Pi
	}
	return a
}
