// Package pilot is a scripted bot that plays a nemesis run: it pursues
// objectives, flees or hides from the hunter, works the tracker and lobs
// decoys. It reads only the simulation, so it drives both the headless demo
// generator and the in-app attract-mode recording without a display.
package pilot

import (
	"math"

	"github.com/danielriddell21/nemesis/internal/sim"
	"github.com/danielriddell21/nemesis/internal/world"
)

// Pilot holds the bot's small amount of cadence state between ticks.
type Pilot struct {
	trackerClock float64
	pulse        bool
	throwTimer   float64
}

// New returns a fresh pilot.
func New() *Pilot { return &Pilot{} }

// Input decides the run's input for one tick of length dt.
func (p *Pilot) Input(g *sim.Game, dt float64) sim.Input {
	var in sim.Input
	p.throwTimer -= dt
	in.Hide = p.wantHide(g)
	if g.Player.Hidden {
		// Tucked away: keep still and wait the hunter out.
		return in
	}
	target, fleeing := p.fleeTarget(g)
	if !fleeing {
		var ok bool
		target, ok = p.objective(g)
		if !ok {
			return in
		}
	}
	next, ok := p.nextCell(g, target)
	if !ok {
		return in
	}

	// Steer toward the centre of the next path cell, walking once the nose is
	// roughly on the line. Beside a console, face the cabinet instead.
	aim := cellMid(next)
	consoleClose := g.World.Level.At(target.X, target.Y) == world.TileConsole &&
		manhattan(g.Player.Pos.Cell(), target) <= 1
	if consoleClose {
		aim = cellMid(target)
	}
	to := aim.Sub(g.Player.Pos)
	want := math.Atan2(to.Y, to.X)
	diff := angleDiff(want, g.Player.Angle)
	in.TurnDelta = clamp(diff, -0.12, 0.12)
	if math.Abs(diff) < 0.7 && !consoleClose {
		in.Forward = 1
	}
	in.Mode = p.gait(g)
	// The sim edge-latches Use, so the pilot pulses the key rather than
	// holding it down.
	p.pulse = !p.pulse
	wantUse := (consoleClose && math.Abs(diff) < 0.3) || p.doorAhead(g, next)
	in.Use = wantUse && p.pulse
	in.Tracker = p.trackerRaised(g, dt)
	in.Throw = p.wantThrow(g)
	return in
}

func (p *Pilot) wantHide(g *sim.Game) bool {
	// Duck into a locker when the hunter is on the prowl and close; stay put
	// until it has wandered well away, then slip out. Returning true only when
	// intent and reality differ presses the toggle exactly once.
	c := g.Player.Pos.Cell()
	onLocker := g.World.Level.At(c.X, c.Y) == world.TileLocker
	d := g.Alien.Pos.Sub(g.Player.Pos).Len()
	hunted := g.Alien.State == sim.StateHunt || g.Alien.State == sim.StateSearch || g.Alien.State == sim.StateInvestigate
	want := onLocker && hunted && d < 7
	if g.Player.Hidden {
		want = d < 11 // stay hidden while it lingers
	}
	return want != g.Player.Hidden
}

func (p *Pilot) wantThrow(g *sim.Game) bool {
	// Lob a noisemaker when the hunter is within earshot but not yet on top of
	// us, on a cooldown so decoys stay meaningful.
	if p.throwTimer > 0 || g.Player.Decoys == 0 || g.DecoyState().Active {
		return false
	}
	if d := g.Alien.Pos.Sub(g.Player.Pos).Len(); d < 6 || d > 13 {
		return false
	}
	p.throwTimer = 9
	return true
}

func (p *Pilot) gait(g *sim.Game) sim.MoveMode {
	// Creep when the hunter is close, sprint when it is far or already after
	// us: exactly the habits the hunter's learning is built to punish.
	d := g.Alien.Pos.Sub(g.Player.Pos).Len()
	switch {
	case g.Alien.State == sim.StateHunt:
		return sim.ModeRun
	case d < 7:
		return sim.ModeSneak
	case d > 15:
		return sim.ModeRun
	default:
		return sim.ModeWalk
	}
}

func (p *Pilot) fleeTarget(g *sim.Game) (world.Coord, bool) {
	// Break away when the hunter presses in, preferring a vent dive: the
	// crawl is slower but the creaks it teaches the hunter are the point of
	// the demo.
	d := g.Alien.Pos.Sub(g.Player.Pos).Len()
	if d > 6.5 || g.Alien.State == sim.StateLurk {
		return world.Coord{}, false
	}
	l := g.World.Level
	best, bestScore, found := world.Coord{}, -math.MaxFloat64, false
	consider := func(c world.Coord, bonus float64) {
		fromAlien := cellMid(c).Sub(g.Alien.Pos).Len()
		fromMe := cellMid(c).Sub(g.Player.Pos).Len()
		if fromAlien < 8 || fromMe < 2 {
			return
		}
		score := fromAlien - 0.6*fromMe + bonus
		if score > bestScore {
			best, bestScore, found = c, score, true
		}
	}
	for _, c := range l.Lockers {
		consider(c, 9) // a locker to hide in is the best escape
	}
	for _, m := range l.VentMouths {
		consider(m, 6)
	}
	for _, r := range l.Rooms {
		consider(r.Center(), 0)
	}
	return best, found
}

func (p *Pilot) trackerRaised(g *sim.Game, dt float64) bool {
	// Check the scope on a nervous cadence, and cling to it when the hunter
	// is close — which is precisely when it can hear the chirp and learn it.
	p.trackerClock += dt
	d := g.Alien.Pos.Sub(g.Player.Pos).Len()
	if d < 5.5 {
		return true
	}
	phase := math.Mod(p.trackerClock, 7)
	return d < 13 && phase < 3
}

func (p *Pilot) objective(g *sim.Game) (world.Coord, bool) {
	l := g.World.Level
	// Nearest console still dark, then the airlock.
	best, bestD, found := world.Coord{}, math.MaxFloat64, false
	for _, c := range l.Consoles {
		if g.ConsoleActivated(c) {
			continue
		}
		if d := cellMid(c).Sub(g.Player.Pos).Len(); d < bestD {
			best, bestD, found = c, d, true
		}
	}
	if found {
		return best, true
	}
	if !g.Escaped() {
		return l.Exit, true
	}
	return world.Coord{}, false
}

func walkTarget(g *sim.Game, target world.Coord) world.Coord {
	// A console is solid: walk to the floor cell in front of it instead.
	l := g.World.Level
	if l.At(target.X, target.Y) != world.TileConsole {
		return target
	}
	for _, n := range [4]world.Coord{
		{X: target.X + 1, Y: target.Y},
		{X: target.X - 1, Y: target.Y},
		{X: target.X, Y: target.Y + 1},
		{X: target.X, Y: target.Y - 1},
	} {
		if l.At(n.X, n.Y).Walkable() {
			return n
		}
	}
	return target
}

func (p *Pilot) nextCell(g *sim.Game, target world.Coord) (world.Coord, bool) {
	from := g.Player.Pos.Cell()
	path := bfs(g, from, walkTarget(g, target))
	if len(path) == 0 {
		return world.Coord{}, false
	}
	next := path[0]
	// Close enough to the current cell centre? Aim one further ahead so the
	// walk stays smooth.
	if len(path) > 1 && cellMid(next).Sub(g.Player.Pos).Len() < 0.35 {
		next = path[1]
	}
	return next, true
}

func (p *Pilot) doorAhead(g *sim.Game, next world.Coord) bool {
	l := g.World.Level
	return l.At(next.X, next.Y) == world.TileDoor && !g.World.DoorOpen(next) &&
		cellMid(next).Sub(g.Player.Pos).Len() < 1.0
}

func bfs(g *sim.Game, src, dst world.Coord) []world.Coord {
	// Doors count as open because the pilot opens them on the way through.
	l := g.World.Level
	if src == dst {
		return nil
	}
	prev := map[world.Coord]world.Coord{src: src}
	queue := []world.Coord{src}
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		if c == dst {
			break
		}
		for _, n := range [4]world.Coord{
			{X: c.X + 1, Y: c.Y},
			{X: c.X - 1, Y: c.Y},
			{X: c.X, Y: c.Y + 1},
			{X: c.X, Y: c.Y - 1},
		} {
			if _, seen := prev[n]; seen {
				continue
			}
			if !l.At(n.X, n.Y).Walkable() {
				continue
			}
			prev[n] = c
			queue = append(queue, n)
		}
	}
	if _, ok := prev[dst]; !ok {
		return nil
	}
	var rev []world.Coord
	for c := dst; c != src; c = prev[c] {
		rev = append(rev, c)
	}
	path := make([]world.Coord, 0, len(rev))
	for i := len(rev) - 1; i >= 0; i-- {
		path = append(path, rev[i])
	}
	return path
}

func cellMid(c world.Coord) sim.Vec2 {
	return sim.Vec2{X: float64(c.X) + 0.5, Y: float64(c.Y) + 0.5}
}

func manhattan(a, b world.Coord) int {
	return absInt(a.X-b.X) + absInt(a.Y-b.Y)
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func angleDiff(a, b float64) float64 {
	d := a - b
	for d > math.Pi {
		d -= 2 * math.Pi
	}
	for d < -math.Pi {
		d += 2 * math.Pi
	}
	return d
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
