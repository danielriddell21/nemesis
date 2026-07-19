package telemetry

import (
	"fmt"

	"github.com/danielriddell21/nemesis/internal/sim"
)

type Event struct {
	Tick    uint64  `json:"tick"`
	Type    string  `json:"type"`
	X       int     `json:"x"`
	Y       int     `json:"y"`
	Radius  float64 `json:"radius,omitempty"`
	State   string  `json:"state,omitempty"`
	TargetX int     `json:"tx,omitempty"`
	TargetY int     `json:"ty,omitempty"`
}

func FromObservation(o sim.Observation) Event {
	e := Event{
		Tick:   o.Tick,
		Type:   o.Kind.String(),
		X:      o.At.X,
		Y:      o.At.Y,
		Radius: o.Radius,
	}
	switch o.Kind {
	case sim.ObsAlienState:
		e.State = o.State.String()
		e.TargetX, e.TargetY = o.Target.X, o.Target.Y
	case sim.ObsAlienHeard, sim.ObsDirectorNudge:
		e.TargetX, e.TargetY = o.Target.X, o.Target.Y
	}
	return e
}

func (e Event) Line() string {
	switch e.Type {
	case "vent-creak", "alien-creak":
		return fmt.Sprintf("VENT CREAK AT %d %d", e.X, e.Y)
	case "door":
		return fmt.Sprintf("BULKHEAD OPENED AT %d %d", e.X, e.Y)
	case "console":
		return fmt.Sprintf("GENERATOR ONLINE AT %d %d", e.X, e.Y)
	case "ping":
		return fmt.Sprintf("TRACKER PING AT %d %d", e.X, e.Y)
	case "alien-state":
		return fmt.Sprintf("HUNTER %s -> %d %d", e.State, e.TargetX, e.TargetY)
	case "alien-heard":
		return fmt.Sprintf("HUNTER HEARD NOISE AT %d %d", e.TargetX, e.TargetY)
	case "alien-seen":
		return fmt.Sprintf("HUNTER SPOTTED PREY AT %d %d", e.X, e.Y)
	case "director-nudge":
		return fmt.Sprintf("DIRECTOR STEERS HUNT TO %d %d", e.TargetX, e.TargetY)
	case "escape":
		return "PREY ESCAPED THROUGH THE AIRLOCK"
	case "death":
		return fmt.Sprintf("PREY KILLED AT %d %d", e.X, e.Y)
	default:
		return ""
	}
}

type Subscriber interface {
	OnEvent(Event)
}

const feedDepth = 64

type Bus struct {
	subs   []Subscriber
	recent []Event
}

func NewBus(subs ...Subscriber) *Bus {
	kept := make([]Subscriber, 0, len(subs))
	for _, s := range subs {
		if s != nil {
			kept = append(kept, s)
		}
	}
	return &Bus{subs: kept}
}

func (b *Bus) Observe(o sim.Observation) {
	if o.Kind == sim.ObsStep {
		// Footsteps are constant; the feed and subscribers care about the
		// discrete triggers, not the drumbeat.
		return
	}
	e := FromObservation(o)
	b.recent = append(b.recent, e)
	if len(b.recent) > feedDepth {
		b.recent = b.recent[len(b.recent)-feedDepth:]
	}
	for _, s := range b.subs {
		s.OnEvent(e)
	}
}

func (b *Bus) Recent() []Event {
	return b.recent
}

func (e Event) Kind() (sim.ObservationKind, bool) {
	for k := sim.ObsStep; k <= sim.ObsDirectorNudge; k++ {
		if k.String() == e.Type {
			return k, true
		}
	}
	return 0, false
}
