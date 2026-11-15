package eval

import (
	"fmt"

	"github.com/danielriddell21/nemesis/internal/pilot"
	"github.com/danielriddell21/nemesis/internal/sim"
	"github.com/danielriddell21/nemesis/internal/world"
)

const (
	tickDT = 1.0 / 60

	MinSolvable = 0.95
	MinEscape   = 0.08
	MaxEscape   = 0.60
)

type Config struct {
	Seeds      int
	Width      int
	Height     int
	Consoles   int
	MaxSeconds float64
}

func DefaultConfig() Config {
	return Config{Seeds: 100, Width: 48, Height: 32, Consoles: 3, MaxSeconds: 600}
}

type Outcome uint8

const (
	Escaped Outcome = iota
	Died
	TimedOut
)

type Report struct {
	Seeds       int
	GenFailures int
	Solvable    int
	Escaped     int
	Died        int
	TimedOut    int
}

func Run(cfg Config) Report {
	r := Report{Seeds: cfg.Seeds}
	for seed := int64(1); seed <= int64(cfg.Seeds); seed++ {
		l, err := world.Generate(world.Config{Width: cfg.Width, Height: cfg.Height, Seed: seed, Consoles: cfg.Consoles})
		if err != nil {
			r.GenFailures++
			continue
		}
		if Play(l, cfg.MaxSeconds, false) == Escaped {
			r.Solvable++
		}
		switch Play(l, cfg.MaxSeconds, true) {
		case Escaped:
			r.Escaped++
		case Died:
			r.Died++
		case TimedOut:
			r.TimedOut++
		}
	}
	return r
}

func Play(l *world.Level, maxSeconds float64, hunter bool) Outcome {
	g := sim.New(l)
	p := pilot.New()
	for g.Elapsed() < maxSeconds && !g.Escaped() && !g.Dead() {
		if !hunter {
			// Park the hunter off the map, asleep, so only the deck and the
			// pilot are under test.
			g.Alien.Pos = sim.Vec2{X: -1000, Y: -1000}
			g.Alien.State = sim.StateLurk
		}
		g.Tick(p.Input(g, tickDT), tickDT)
	}
	switch {
	case g.Escaped():
		return Escaped
	case g.Dead():
		return Died
	default:
		return TimedOut
	}
}

func (r Report) SolvableRate() float64 { return rate(r.Solvable, r.Seeds) }

func (r Report) EscapeRate() float64 { return rate(r.Escaped, r.Seeds) }

func rate(n, of int) float64 {
	if of == 0 {
		return 0
	}
	return float64(n) / float64(of)
}

func (r Report) Failures() []string {
	var out []string
	if r.Seeds == 0 {
		out = append(out, "no seeds were run")
	}
	if r.GenFailures > 0 {
		out = append(out, fmt.Sprintf("%d of %d seeds failed to generate a deck", r.GenFailures, r.Seeds))
	}
	if got := r.SolvableRate(); got < MinSolvable {
		out = append(out, fmt.Sprintf("solvable rate %.0f%% is under the %.0f%% pass mark", got*100, MinSolvable*100))
	}
	if got := r.EscapeRate(); got < MinEscape {
		out = append(out, fmt.Sprintf("escape rate %.0f%% is under the %.0f%% floor: the hunter is unfair", got*100, MinEscape*100))
	}
	if got := r.EscapeRate(); got > MaxEscape {
		out = append(out, fmt.Sprintf("escape rate %.0f%% is over the %.0f%% ceiling: the hunter is toothless", got*100, MaxEscape*100))
	}
	return out
}

func (r Report) String() string {
	return fmt.Sprintf(
		"seeds %d | generation failures %d | solvable (hunter off) %d (%.0f%%, pass >= %.0f%%) | "+
			"vs hunter: escaped %d (%.0f%%, pass %.0f-%.0f%%), died %d, timed out %d",
		r.Seeds, r.GenFailures, r.Solvable, r.SolvableRate()*100, MinSolvable*100,
		r.Escaped, r.EscapeRate()*100, MinEscape*100, MaxEscape*100, r.Died, r.TimedOut)
}
