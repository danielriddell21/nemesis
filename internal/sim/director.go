package sim

import "github.com/danielriddell21/nemesis/internal/world"

const (
	lurkTime      = 25.0
	nudgeInterval = 30.0
	campLimit     = 18.0
	campRange     = 7.0
	menaceRange   = 9.0
	leashMinDist  = 14
)

type director struct {
	menace     float64
	aggression float64
	wake       float64
	nudge      float64
	camp       float64
	hint       world.Coord
	hintValid  bool
}

func newDirector() director {
	return director{wake: lurkTime, nudge: nudgeInterval}
}

func (d *director) tick(g *Game, dt float64) {
	d.tickWake(g, dt)
	d.tickMenace(g, dt)
	d.tickCamp(g, dt)
	d.tickNudge(g, dt)
}

func (d *director) tickWake(g *Game, dt float64) {
	if g.Alien.State != StateLurk {
		return
	}
	d.wake -= dt
	if d.wake <= 0 {
		g.setAlienState(StatePatrol)
		g.setAlienTarget(g.patrolPoint())
	}
}

func (d *director) tickMenace(g *Game, dt float64) {
	// Menace is the tension dial: it climbs while the hunter is close and
	// bleeds away when it prowls elsewhere. The audio bed follows it.
	dist := g.Alien.Pos.Sub(g.Player.Pos).Len()
	if g.Alien.State == StateHunt {
		d.menace = clamp01(d.menace + dt*0.5)
		return
	}
	if dist < menaceRange && g.Alien.State != StateLurk {
		d.menace = clamp01(d.menace + dt*(menaceRange-dist)/menaceRange*0.25)
	} else {
		d.menace = clamp01(d.menace - dt*0.06)
	}
}

func (d *director) tickCamp(g *Game, dt float64) {
	// Fairness leash: a hunter that loiters near an unseen player too long is
	// sent to a far room so runs never stall in a corner.
	dist := g.Alien.Pos.Sub(g.Player.Pos).Len()
	if g.Alien.State == StateHunt || g.Alien.State == StateLurk || dist > campRange {
		d.camp = 0
		return
	}
	d.camp += dt
	if d.camp < campLimit {
		return
	}
	d.camp = 0
	far := d.farRoom(g)
	g.setAlienTarget(far)
	g.setAlienState(StatePatrol)
	g.observe(Observation{Kind: ObsDirectorNudge, At: g.Alien.Pos.Cell(), Target: far})
}

func (d *director) tickNudge(g *Game, dt float64) {
	if g.Alien.State == StateLurk {
		return
	}
	d.nudge -= dt
	if d.nudge > 0 {
		return
	}
	d.nudge = nudgeInterval * (1.2 - d.aggression*0.6)
	if g.Alien.State == StateHunt || g.Alien.State == StateInvestigate {
		return
	}
	// The director always knows where the player is, but it only gestures: the
	// hint is the player's neighbourhood, never the exact cell.
	hint := d.nearPlayerRoom(g)
	d.hint = hint
	d.hintValid = true
	g.setAlienTarget(hint)
	g.setAlienState(StateInvestigate)
	g.observe(Observation{Kind: ObsDirectorNudge, At: g.Alien.Pos.Cell(), Target: hint})
}

func (d *director) escalate(g *Game) {
	total := len(g.World.Level.Consoles)
	if total > 0 {
		d.aggression = clamp01(d.aggression + 1/float64(total))
	}
	d.menace = clamp01(d.menace + 0.3)
	// A generator roaring back to life always draws the hunter's attention.
	if g.Alien.State != StateHunt {
		g.setAlienTarget(g.Player.Pos.Cell())
		g.setAlienState(StateInvestigate)
	}
	if g.Alien.State == StateLurk {
		d.wake = 0
	}
}

func (d *director) nearPlayerRoom(g *Game) world.Coord {
	l := g.World.Level
	if ri := l.RoomAt(g.Player.Pos.Cell()); ri >= 0 {
		return l.Rooms[ri].Center()
	}
	return g.Player.Pos.Cell()
}

func (d *director) farRoom(g *Game) world.Coord {
	l := g.World.Level
	p := g.Player.Pos.Cell()
	for i := range l.Rooms {
		c := l.Rooms[i].Center()
		if absInt(c.X-p.X)+absInt(c.Y-p.Y) >= leashMinDist {
			return c
		}
	}
	if len(l.Rooms) > 0 {
		return l.Rooms[0].Center()
	}
	return p
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
