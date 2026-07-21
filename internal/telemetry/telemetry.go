package telemetry

import (
	"fmt"
	"strings"

	ctel "github.com/danielriddell21/crucible/telemetry"

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
	Learn   string  `json:"learn,omitempty"`
	Tier    int     `json:"tier,omitempty"`
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
	case sim.ObsAlienLearn:
		e.Learn = o.Learn.String()
		e.Tier = o.Tier
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
	case "decoy":
		return fmt.Sprintf("DECOY CHIRPS AT %d %d", e.X, e.Y)
	case "hide":
		return fmt.Sprintf("PREY DUCKS INTO LOCKER AT %d %d", e.X, e.Y)
	case "alien-state":
		return fmt.Sprintf("HUNTER %s -> %d %d", e.State, e.TargetX, e.TargetY)
	case "alien-heard":
		return fmt.Sprintf("HUNTER HEARD NOISE AT %d %d", e.TargetX, e.TargetY)
	case "alien-seen":
		return fmt.Sprintf("HUNTER SPOTTED PREY AT %d %d", e.X, e.Y)
	case "director-nudge":
		return fmt.Sprintf("DIRECTOR STEERS HUNT TO %d %d", e.TargetX, e.TargetY)
	case "alien-learn":
		return fmt.Sprintf("HUNTER LEARNED %s %s", strings.ToUpper(e.Learn), roman(e.Tier))
	case "escape":
		return "PREY ESCAPED THROUGH THE AIRLOCK"
	case "death":
		return fmt.Sprintf("PREY KILLED AT %d %d", e.X, e.Y)
	default:
		return ""
	}
}

// Subscriber receives each published event. It aliases the engine's
// generic subscriber specialised to Event.
type Subscriber = ctel.Subscriber[Event]

const feedDepth = 64

// Bus adapts the game's sim.Observer stream onto the engine's generic
// event bus: it converts observations to events, mutes the constant
// footstep drumbeat, and lets the engine handle fan-out and the recent
// feed.
type Bus struct {
	inner *ctel.Bus[Event]
}

func NewBus(subs ...Subscriber) *Bus {
	return &Bus{inner: ctel.NewBus(subs...).Configure(ctel.WithFeedDepth[Event](feedDepth))}
}

func (b *Bus) Observe(o sim.Observation) {
	if o.Kind == sim.ObsStep {
		// Footsteps are constant; the feed and subscribers care about the
		// discrete triggers, not the drumbeat.
		return
	}
	b.inner.Publish(FromObservation(o))
}

func (b *Bus) Recent() []Event {
	return b.inner.Recent()
}

func (e Event) Kind() (sim.ObservationKind, bool) {
	for k := sim.ObsStep; k <= sim.ObsAlienLearn; k++ {
		if k.String() == e.Type {
			return k, true
		}
	}
	return 0, false
}

func roman(n int) string {
	numerals := []string{"0", "I", "II", "III", "IV", "V"}
	if n < 0 || n >= len(numerals) {
		return "V+"
	}
	return numerals[n]
}
