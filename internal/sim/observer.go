package sim

import "github.com/danielriddell21/nemesis/internal/world"

type ObservationKind uint8

const (
	ObsStep ObservationKind = iota
	ObsVentCreak
	ObsDoorOpen
	ObsConsole
	ObsTrackerPing
	ObsDecoy
	ObsHide
	ObsEscape
	ObsDeath
	ObsAlienState
	ObsAlienHeard
	ObsAlienSeen
	ObsAlienCreak
	ObsDirectorNudge
	ObsAlienLearn
)

var obsNames = [...]string{
	ObsStep:          "step",
	ObsVentCreak:     "vent-creak",
	ObsDoorOpen:      "door",
	ObsConsole:       "console",
	ObsTrackerPing:   "ping",
	ObsDecoy:         "decoy",
	ObsHide:          "hide",
	ObsEscape:        "escape",
	ObsDeath:         "death",
	ObsAlienState:    "alien-state",
	ObsAlienHeard:    "alien-heard",
	ObsAlienSeen:     "alien-seen",
	ObsAlienCreak:    "alien-creak",
	ObsDirectorNudge: "director-nudge",
	ObsAlienLearn:    "alien-learn",
}

func (k ObservationKind) String() string {
	if int(k) < len(obsNames) && obsNames[k] != "" {
		return obsNames[k]
	}
	return "unknown"
}

type LearnKind uint8

const (
	LearnPing LearnKind = iota
	LearnVent
	LearnSearch
	LearnDecoy
	LearnLocker
)

func (k LearnKind) String() string {
	switch k {
	case LearnPing:
		return "tracker"
	case LearnVent:
		return "vents"
	case LearnSearch:
		return "search"
	case LearnDecoy:
		return "decoys"
	case LearnLocker:
		return "lockers"
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
