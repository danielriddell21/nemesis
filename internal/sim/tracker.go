package sim

import "math"

const (
	trackerRange = 14.0
	pingInterval = 1.1
)

type Tracker struct {
	Raised  bool
	Contact Contact

	pingTimer float64
}

type Contact struct {
	Valid   bool
	Bearing float64
	Dist    float64
	Age     float64
}

func newTracker() Tracker {
	return Tracker{}
}

func (g *Game) tickTracker(in Input, dt float64, noise *noiseEvent) {
	t := &g.Tracker
	t.Contact.Age += dt
	if !in.Tracker {
		t.Raised = false
		t.pingTimer = 0
		return
	}
	if !t.Raised {
		t.Raised = true
		t.pingTimer = 0
	}
	t.pingTimer -= dt
	if t.pingTimer > 0 {
		return
	}
	t.pingTimer = pingInterval

	// The ping is a real sound in the world: the hunter can hear it too.
	g.observe(Observation{Kind: ObsTrackerPing, At: g.Player.Pos.Cell(), Radius: pingNoise})
	noise.merge(noiseEvent{at: g.Player.Pos, radius: pingNoise})

	// The tracker senses motion, not bodies: a still hunter paints no blip.
	to := g.Alien.Pos.Sub(g.Player.Pos)
	dist := to.Len()
	if !g.Alien.Moving || dist > trackerRange {
		t.Contact = Contact{}
		return
	}
	t.Contact = Contact{
		Valid:   true,
		Bearing: normalizeAngle(math.Atan2(to.Y, to.X) - g.Player.Angle),
		Dist:    dist,
	}
}
