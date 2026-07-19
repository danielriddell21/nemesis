package sim

import "github.com/danielriddell21/nemesis/internal/world"

type ObservationKind uint8

const (
	ObsStep ObservationKind = iota
	ObsVentCreak
	ObsDoorOpen
	ObsConsole
	ObsTrackerPing
	ObsEscape
	ObsDeath
	ObsAlienState
	ObsAlienHeard
	ObsAlienSeen
	ObsAlienCreak
	ObsDirectorNudge
	ObsAlienLearn
)

func (k ObservationKind) String() string {
	switch k {
	case ObsStep:
		return "step"
	case ObsVentCreak:
		return "vent-creak"
	case ObsDoorOpen:
		return "door"
	case ObsConsole:
		return "console"
	case ObsTrackerPing:
		return "ping"
	case ObsEscape:
		return "escape"
	case ObsDeath:
		return "death"
	case ObsAlienState:
		return "alien-state"
	case ObsAlienHeard:
		return "alien-heard"
	case ObsAlienSeen:
		return "alien-seen"
	case ObsAlienCreak:
		return "alien-creak"
	case ObsDirectorNudge:
		return "director-nudge"
	case ObsAlienLearn:
		return "alien-learn"
	default:
		return "unknown"
	}
}

type LearnKind uint8

const (
	LearnPing LearnKind = iota
	LearnVent
	LearnSearch
)

func (k LearnKind) String() string {
	switch k {
	case LearnPing:
		return "tracker"
	case LearnVent:
		return "vents"
	case LearnSearch:
		return "search"
	default:
		return "unknown"
	}
}

type Observation struct {
	Tick   uint64
	Kind   ObservationKind
	At     world.Coord
	Radius float64
	State  AlienState
	Target world.Coord
	Learn  LearnKind
	Tier   int
}

type Observer interface {
	Observe(Observation)
}

type nopObserver struct{}

func (nopObserver) Observe(Observation) {}

func Fanout(obs ...Observer) Observer {
	kept := make([]Observer, 0, len(obs))
	for _, o := range obs {
		if o != nil {
			kept = append(kept, o)
		}
	}
	return fanout(kept)
}

type fanout []Observer

func (f fanout) Observe(o Observation) {
	for _, s := range f {
		s.Observe(o)
	}
}
