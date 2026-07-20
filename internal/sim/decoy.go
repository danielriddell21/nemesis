package sim

const (
	startingDecoys  = 2
	decoyThrowSpeed = 9.0
	decoyFlight     = 0.45
	decoyLifetime   = 5.0
	decoyPulse      = 0.6
	decoyNoise      = 8.0
	decoyReach      = 1.4
)

type Decoy struct {
	Pos    Vec2
	Active bool
	Landed bool

	vel    Vec2
	flight float64
	ttl    float64
	pulse  float64
}

func (g *Game) tickThrow(in Input) {
	p := &g.Player
	if !in.Throw || g.throwLatch {
		g.throwLatch = in.Throw
		return
	}
	g.throwLatch = true
	if p.Hidden || p.Decoys <= 0 || g.decoy.Active {
		return
	}
	p.Decoys--
	dir := p.Dir()
	g.decoy = Decoy{
		Pos:    p.Pos,
		Active: true,
		vel:    Vec2{X: dir.X * decoyThrowSpeed, Y: dir.Y * decoyThrowSpeed},
		flight: decoyFlight,
		ttl:    decoyLifetime,
		pulse:  0,
	}
}

// tickDecoy advances the thrown noisemaker and returns the noise it makes this
// step (empty when it is silent). The gadget flies forward until it hits a wall
// or its flight time runs out, then chirps on a cadence until it dies.
func (g *Game) tickDecoy(dt float64) noiseEvent {
	d := &g.decoy
	if !d.Active {
		return noiseEvent{}
	}
	if !d.Landed {
		d.flight -= dt
		next := resolveMove(g.World, d.Pos, d.vel.X*dt, d.vel.Y*dt)
		if next == d.Pos || d.flight <= 0 {
			d.Landed = true
		}
		d.Pos = next
		return noiseEvent{}
	}

	d.ttl -= dt
	if d.ttl <= 0 {
		d.Active = false
		return noiseEvent{}
	}
	d.pulse -= dt
	if d.pulse > 0 {
		return noiseEvent{}
	}
	d.pulse = decoyPulse
	g.observe(Observation{Kind: ObsDecoy, At: d.Pos.Cell(), Radius: decoyNoise})
	return noiseEvent{at: d.Pos, radius: decoyNoise, kind: ObsDecoy}
}

// checkDecoyReached lets a lured hunter discover the gadget: reaching it while
// investigating teaches it the sound is a trick and silences the decoy.
func (g *Game) checkDecoyReached() {
	d := &g.decoy
	if !d.Active || !d.Landed {
		return
	}
	if g.Alien.State != StateInvestigate && g.Alien.State != StateSearch {
		return
	}
	if g.Alien.Pos.Sub(d.Pos).Len() > decoyReach {
		return
	}
	d.Active = false
	g.reachedDecoy()
}

func (g *Game) DecoyState() Decoy { return g.decoy }
