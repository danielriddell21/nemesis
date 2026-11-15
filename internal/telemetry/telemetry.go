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
	case sim.ObsAlienHeard, sim.ObsAlienGlow, sim.ObsDirectorNudge:
		e.TargetX, e.TargetY = o.Target.X, o.Target.Y
	case sim.ObsAlienLearn:
		e.Learn = o.Learn.String()
		e.Tier = o.Tier
	}
	return e
}

var lines = map[string]func(Event) string{
	"vent-creak":     func(e Event) string { return fmt.Sprintf("VENT CREAK AT %d %d", e.X, e.Y) },
	"alien-creak":    func(e Event) string { return fmt.Sprintf("VENT CREAK AT %d %d", e.X, e.Y) },
	"door":           func(e Event) string { return fmt.Sprintf("BULKHEAD OPENED AT %d %d", e.X, e.Y) },
	"console":        func(e Event) string { return fmt.Sprintf("GENERATOR ONLINE AT %d %d", e.X, e.Y) },
	"ping":           func(e Event) string { return fmt.Sprintf("TRACKER PING AT %d %d", e.X, e.Y) },
	"decoy":          func(e Event) string { return fmt.Sprintf("DECOY CHIRPS AT %d %d", e.X, e.Y) },
	"hide":           func(e Event) string { return fmt.Sprintf("PREY DUCKS INTO LOCKER AT %d %d", e.X, e.Y) },
	"battery":        func(e Event) string { return fmt.Sprintf("TORCH CELL TAKEN AT %d %d", e.X, e.Y) },
	"alien-state":    func(e Event) string { return fmt.Sprintf("HUNTER %s -> %d %d", e.State, e.TargetX, e.TargetY) },
	"alien-heard":    func(e Event) string { return fmt.Sprintf("HUNTER HEARD NOISE AT %d %d", e.TargetX, e.TargetY) },
	"alien-glow":     func(e Event) string { return fmt.Sprintf("HUNTER NOTICED TORCHLIGHT AT %d %d", e.TargetX, e.TargetY) },
	"alien-seen":     func(e Event) string { return fmt.Sprintf("HUNTER SPOTTED PREY AT %d %d", e.X, e.Y) },
	"director-nudge": func(e Event) string { return fmt.Sprintf("DIRECTOR STEERS HUNT TO %d %d", e.TargetX, e.TargetY) },
	"alien-learn": func(e Event) string {
		return fmt.Sprintf("HUNTER LEARNED %s %s", strings.ToUpper(e.Learn), roman(e.Tier))
	},
	"escape": func(Event) string { return "PREY ESCAPED THROUGH THE AIRLOCK" },
	"death":  func(e Event) string { return fmt.Sprintf("PREY KILLED AT %d %d", e.X, e.Y) },
}

func (e Event) Line() string {
	format, ok := lines[e.Type]
	if !ok {
		return ""
	}
	return format(e)
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
	for k := sim.ObsStep; k < sim.ObsKindCount; k++ {
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
