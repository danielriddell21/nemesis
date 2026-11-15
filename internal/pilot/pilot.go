// Package pilot is a scripted bot that plays a nemesis run: it pursues
// objectives, routes around the hunter, flees or hides from it, works the
// torch and tracker and lobs decoys. It reads only the simulation, so it drives both the headless demo
// generator and the in-app attract-mode recording without a display.
package pilot

import (
	"math"
	"slices"

	"github.com/danielriddell21/nemesis/internal/sim"
	"github.com/danielriddell21/nemesis/internal/world"
)

// Pilot holds the bot's small amount of cadence state between ticks.
type Pilot struct {
	trackerClock   float64
	pulse          bool
	throwTimer     float64
	torchPulse     bool
	fleeing        bool
	startedFleeing bool
	goal           world.Coord
	hasGoal        bool

	plan plan
}

type plan struct {
	from   world.Coord
	target world.Coord
	path   []world.Coord
	ok     bool
	age    int
}

const replanTicks = 6

const (
	runSafeDist    = 12.5
	walkSafeDist   = 5.5
	pingSafeDist   = 7.5
	torchSafeDist  = 17.0
	avoidRadius    = 5.0
	fleeDist       = 6.5
	goalDangerDist = 9.0
)

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
	target, next, ok := p.replan(g)
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
	// Entering the access cell is not enough: close in until the cabinet is
	// actually within reach.
	inReach := consoleClose && g.FacingInteractable() == sim.InteractConsole
	if math.Abs(diff) < 0.7 && !inReach {
		in.Forward = 1
	}
	in.Mode = p.gait(g)
	// The sim edge-latches Use, so the pilot pulses the key rather than
	// holding it down.
	p.pulse = !p.pulse
	wantUse := inReach || p.doorAhead(g, next)
	in.Use = wantUse && p.pulse
	in.Tracker = p.trackerRaised(g, dt)
	// Only worth a decoy when the retreat really does turn away from the old
	// heading; thrown straight ahead it would call the hunter after us.
	in.Throw = math.Abs(diff) > 1 && p.wantThrow(g)
	in.Torch = p.wantTorch(g)
	return in
}

func (p *Pilot) replan(g *sim.Game) (target, next world.Coord, ok bool) {
	// Route-finding is the pilot's whole cost, so a plan is kept for a few
	// ticks unless the pilot has stepped into a new cell.
	pl := &p.plan
	p.startedFleeing = false
	here := g.Player.Pos.Cell()
	if pl.age > 0 && pl.from == here {
		pl.age--
		return pl.target, pl.steer(g), pl.ok
	}
	*pl = plan{from: here, age: replanTicks}
	if pl.target, pl.ok = p.destination(g); !pl.ok {
		return pl.target, pl.target, false
	}
	pl.path, pl.ok = p.path(g, pl.target)
	return pl.target, pl.steer(g), pl.ok
}

func (pl *plan) steer(g *sim.Game) world.Coord {
	// Close enough to the next cell's centre? Aim one further ahead so the
	// walk stays smooth.
	if len(pl.path) == 0 {
		return pl.target
	}
	if len(pl.path) > 1 && cellMid(pl.path[0]).Sub(g.Player.Pos).Len() < 0.35 {
		return pl.path[1]
	}
	return pl.path[0]
}

func (p *Pilot) wantTorch(g *sim.Game) bool {
	// Light the way while the hunter is far off, and go dark well before the
	// glow could reach it. Returning true only when intent and reality differ
	// presses the toggle exactly once.
	want := g.Player.Battery > 0 && hunterDist(g) > torchSafeDist
	p.torchPulse = !p.torchPulse
	return want != g.Player.Torch && p.torchPulse
}

func hunterDist(g *sim.Game) float64 {
	return g.Alien.Pos.Sub(g.Player.Pos).Len()
}

func (p *Pilot) wantHide(g *sim.Game) bool {
	// Duck into a locker while the hunter is after us; stay put until it has
	// lost interest and wandered off, then slip out. Returning true only when
	// intent and reality differ presses the toggle exactly once.
	c := g.Player.Pos.Cell()
	onLocker := g.World.Level.At(c.X, c.Y) == world.TileLocker
	want := onLocker && p.threatened(g) && !g.Exposed(c)
	if g.Player.Hidden {
		want = p.threatened(g) || hunterDist(g) < 10
	}
	return want != g.Player.Hidden
}

func (p *Pilot) threatened(g *sim.Game) bool {
	// In danger when the hunter is close, or is coming to look at somewhere
	// close.
	d := hunterDist(g)
	switch g.Alien.State {
	case sim.StateLurk:
		return false
	case sim.StatePatrol:
		return d < fleeDist
	default:
		return d < fleeDist || (d < 16 && cellMid(g.Alien.Target).Sub(g.Player.Pos).Len() < 6)
	}
}

func (p *Pilot) destination(g *sim.Game) (world.Coord, bool) {
	// Retreat when threatened, or when the hunter stands on every way to the
	// objective; otherwise press on with the objective.
	goal, haveGoal := p.objective(g)
	blocked := haveGoal && g.Alien.State != sim.StateHunt && !p.safeRoute(g, goal)
	if p.threatened(g) || blocked {
		if c, ok := p.refuge(g); ok {
			// Flagged for the one tick the pilot turns to run: the moment
			// to drop a decoy on the heading it is abandoning.
			p.startedFleeing = !p.fleeing
			p.fleeing = true
			return c, true
		}
	}
	p.startedFleeing, p.fleeing = false, false
	if blocked {
		// Nowhere better to be: keep still and let the hunter move on.
		return world.Coord{}, false
	}
	return goal, haveGoal
}

func (p *Pilot) safeRoute(g *sim.Game, target world.Coord) bool {
	from, dest := g.Player.Pos.Cell(), walkTarget(g, target)
	return from == dest || len(bfs(g, from, dest, nearHunter(g))) > 0
}

func route(g *sim.Game, from, dest world.Coord) []world.Coord {
	// Best first: clear of the hunter and of the loud, slow ducts; then just
	// clear of the hunter; then anything at all.
	danger := nearHunter(g)
	l := g.World.Level
	quiet := func(c world.Coord) bool {
		return l.At(c.X, c.Y) == world.TileVent || (danger != nil && danger(c))
	}
	if path := bfs(g, from, dest, quiet); len(path) > 0 {
		return path
	}
	if danger != nil {
		if path := bfs(g, from, dest, danger); len(path) > 0 {
			return path
		}
	}
	return bfs(g, from, dest, nil)
}

func (p *Pilot) wantThrow(g *sim.Game) bool {
	// Lob a noisemaker down the old heading at the moment of turning to flee,
	// so the chirp calls the hunter somewhere the pilot is leaving. A cooldown
	// keeps decoys meaningful.
	if !p.startedFleeing {
		return false
	}
	if p.throwTimer > 0 || g.Player.Decoys == 0 || g.DecoyState().Active {
		return false
	}
	// A sleeping hunter needs no distraction, and one in full chase ignores it.
	if g.Alien.State == sim.StateLurk || g.Alien.State == sim.StateHunt {
		return false
	}
	if hunterDist(g) < 3 {
		return false
	}
	p.throwTimer = 9
	return true
}

func (p *Pilot) gait(g *sim.Game) sim.MoveMode {
	// Sprint only when the hunter is out of earshot or already after us; creep
	// whenever footsteps could carry to it. The safe distances are where it
	// cannot hear the gait even at its most alert.
	d := hunterDist(g)
	switch {
	case g.Alien.State == sim.StateHunt && (d < 6 || g.Exposed(g.Player.Pos.Cell())):
		return sim.ModeRun
	case g.Alien.State == sim.StateHunt:
		// Out of its sight with some room: go quiet so it loses the trail.
		return sim.ModeSneak
	case d > runSafeDist:
		return sim.ModeRun
	case d > walkSafeDist:
		return sim.ModeWalk
	default:
		return sim.ModeSneak
	}
}

func (p *Pilot) refuge(g *sim.Game) (world.Coord, bool) {
	// Somewhere out of the hunter's sight, far from it and from wherever it is
	// headed, preferring a locker to hide in. Never a vent: the crawl is slow
	// and loud, and the hunter is quicker in the ducts than out of them.
	l := g.World.Level
	heading := cellMid(g.Alien.Target)
	toHunter := g.Alien.Pos.Sub(g.Player.Pos)
	best, bestScore, found := world.Coord{}, -math.MaxFloat64, false
	consider := func(c world.Coord, bonus float64) {
		fromAlien := cellMid(c).Sub(g.Alien.Pos).Len()
		fromHeading := cellMid(c).Sub(heading).Len()
		if fromAlien < 8 || fromHeading < 6 || g.Exposed(c) {
			return
		}
		toRefuge := cellMid(c).Sub(g.Player.Pos)
		fromMe := toRefuge.Len()
		score := math.Min(fromAlien, fromHeading+2) - 0.6*fromMe + bonus
		if toRefuge.X*toHunter.X+toRefuge.Y*toHunter.Y > 0 {
			// It lies on the hunter's side of us: getting there means
			// closing the gap first.
			score -= 12
		}
		if score > bestScore {
			best, bestScore, found = c, score, true
		}
	}
	for _, c := range l.Lockers {
		consider(c, 4)
	}
	for _, r := range l.Rooms {
		consider(r.Center(), 0)
	}
	return best, found
}

func (p *Pilot) trackerRaised(g *sim.Game, dt float64) bool {
	// Check the scope on a nervous cadence, but never where the hunter could
	// hear the chirp.
	p.trackerClock += dt
	d := hunterDist(g)
	phase := math.Mod(p.trackerClock, 7)
	return d > pingSafeDist && d < 14 && phase < 3
}

func (p *Pilot) objective(g *sim.Game) (world.Coord, bool) {
	l := g.World.Level
	// Stay committed to a console until it is lit: re-picking the nearest
	// every tick dithers forever between two that are equally far.
	if p.hasGoal && !g.ConsoleActivated(p.goal) && !p.goalCompromised(g) {
		return p.goal, true
	}
	p.hasGoal = false
	// Nearest console still dark, then the airlock. One the hunter is sitting
	// on counts as far away, so a safer console is taken first.
	best, bestD, found := world.Coord{}, math.MaxFloat64, false
	for _, c := range l.Consoles {
		if g.ConsoleActivated(c) {
			continue
		}
		d := cellMid(c).Sub(g.Player.Pos).Len()
		if cellMid(c).Sub(g.Alien.Pos).Len() < goalDangerDist {
			d += 1000
		}
		if d < bestD {
			best, bestD, found = c, d, true
		}
	}
	if found {
		p.goal, p.hasGoal = best, true
		return best, true
	}
	if !g.Escaped() {
		return l.Exit, true
	}
	return world.Coord{}, false
}

func (p *Pilot) goalCompromised(g *sim.Game) bool {
	// Give up a console the hunter has moved onto while we are still far from
	// it, as long as there is another to go for instead.
	if cellMid(p.goal).Sub(g.Alien.Pos).Len() >= goalDangerDist {
		return false
	}
	if cellMid(p.goal).Sub(g.Player.Pos).Len() < 4 {
		return false
	}
	for _, c := range g.World.Level.Consoles {
		if c != p.goal && !g.ConsoleActivated(c) && cellMid(c).Sub(g.Alien.Pos).Len() >= goalDangerDist {
			return true
		}
	}
	return false
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

func (p *Pilot) path(g *sim.Game, target world.Coord) ([]world.Coord, bool) {
	from := g.Player.Pos.Cell()
	dest := walkTarget(g, target)
	if from == dest {
		// Already at the access cell: nothing to walk, aim at the target.
		return nil, true
	}
	path := route(g, from, dest)
	return path, len(path) > 0
}

func (p *Pilot) nextCell(g *sim.Game, target world.Coord) (world.Coord, bool) {
	path, ok := p.path(g, target)
	pl := plan{target: target, path: path}
	return pl.steer(g), ok
}

func (p *Pilot) doorAhead(g *sim.Game, next world.Coord) bool {
	l := g.World.Level
	return l.At(next.X, next.Y) == world.TileDoor && !g.World.DoorOpen(next) &&
		cellMid(next).Sub(g.Player.Pos).Len() < 1.0
}

func nearHunter(g *sim.Game) func(world.Coord) bool {
	// While the hunter sleeps or prowls, cells within its reach or in its
	// sight are off the route. Once it is in full chase there is no safe side to pick.
	if g.Alien.State == sim.StateHunt {
		return nil
	}
	return func(c world.Coord) bool {
		return cellMid(c).Sub(g.Alien.Pos).Len() < avoidRadius || g.Exposed(c)
	}
}

func bfs(g *sim.Game, src, dst world.Coord, avoid func(world.Coord) bool) []world.Coord {
	// Doors count as open because the pilot opens them on the way through.
	// Cells avoid reports are skipped, except the one being stood on.
	l := g.World.Level
	if src == dst || !l.InBounds(src.X, src.Y) || !l.InBounds(dst.X, dst.Y) {
		return nil
	}
	// prev holds each reached cell's predecessor index plus one, so zero
	// means unvisited.
	prev := make([]int32, l.W*l.H)
	start, goal := l.Index(src.X, src.Y), l.Index(dst.X, dst.Y)
	prev[start] = int32(start) + 1
	queue := []world.Coord{src}
	for head := 0; head < len(queue) && prev[goal] == 0; head++ {
		c := queue[head]
		ci := int32(l.Index(c.X, c.Y)) + 1
		for _, n := range [4]world.Coord{
			{X: c.X + 1, Y: c.Y},
			{X: c.X - 1, Y: c.Y},
			{X: c.X, Y: c.Y + 1},
			{X: c.X, Y: c.Y - 1},
		} {
			if !l.InBounds(n.X, n.Y) || prev[l.Index(n.X, n.Y)] != 0 {
				continue
			}
			if !l.At(n.X, n.Y).Walkable() || (avoid != nil && avoid(n)) {
				continue
			}
			prev[l.Index(n.X, n.Y)] = ci
			queue = append(queue, n)
		}
	}
	if prev[goal] == 0 {
		return nil
	}
	var rev []world.Coord
	for i := goal; i != start; i = int(prev[i]) - 1 {
		rev = append(rev, world.Coord{X: i % l.W, Y: i / l.W})
	}
	slices.Reverse(rev)
	return rev
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
