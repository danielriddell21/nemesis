package eval

import (
	"strings"
	"testing"

	"github.com/danielriddell21/nemesis/internal/world"
)

func TestPlayWithoutTheHunterEscapes(t *testing.T) {
	l, err := world.Generate(world.Config{Width: 48, Height: 32, Seed: 1, Consoles: 3})
	if err != nil {
		t.Fatal(err)
	}
	if got := Play(l, 600, false); got != Escaped {
		t.Errorf("outcome with the hunter parked = %v, want Escaped", got)
	}
	if got := Play(l, 0.5, false); got != TimedOut {
		t.Errorf("outcome with half a second on the clock = %v, want TimedOut", got)
	}
}

func TestRunCountsEverySeedAndIsDeterministic(t *testing.T) {
	cfg := Config{Seeds: 6, Width: 48, Height: 32, Consoles: 3, MaxSeconds: 600}
	a, b := Run(cfg), Run(cfg)
	if a != b {
		t.Errorf("two runs differ: %+v vs %+v", a, b)
	}
	if a.Escaped+a.Died+a.TimedOut+a.GenFailures != a.Seeds {
		t.Errorf("outcomes do not add up to the seeds: %+v", a)
	}
	if a.Solvable != a.Seeds {
		t.Errorf("solvable %d of %d: every deck should be clearable with the hunter parked", a.Solvable, a.Seeds)
	}
}

func TestFailuresApplyThePassMarks(t *testing.T) {
	cases := []struct {
		name string
		r    Report
		want string
	}{
		{"clean pass", Report{Seeds: 100, Solvable: 100, Escaped: 20, Died: 80}, ""},
		{"nothing run", Report{}, "no seeds"},
		{"generation failure", Report{Seeds: 100, GenFailures: 1, Solvable: 99, Escaped: 20}, "failed to generate"},
		{"unsolvable decks", Report{Seeds: 100, Solvable: 94, Escaped: 20}, "solvable rate"},
		{"unfair hunter", Report{Seeds: 100, Solvable: 100, Escaped: 7}, "unfair"},
		{"toothless hunter", Report{Seeds: 100, Solvable: 100, Escaped: 61}, "toothless"},
		{"on the floor", Report{Seeds: 100, Solvable: 95, Escaped: 8}, ""},
		{"on the ceiling", Report{Seeds: 100, Solvable: 100, Escaped: 60}, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := strings.Join(c.r.Failures(), "; ")
			if c.want == "" && got != "" {
				t.Errorf("unexpected failures: %s", got)
			}
			if c.want != "" && !strings.Contains(got, c.want) {
				t.Errorf("failures %q do not mention %q", got, c.want)
			}
		})
	}
}

func TestReportStringCarriesTheMarks(t *testing.T) {
	s := Report{Seeds: 100, Solvable: 100, Escaped: 20, Died: 80}.String()
	for _, want := range []string{"seeds 100", "escaped 20 (20%", "pass >= 95%", "pass 8-60%"} {
		if !strings.Contains(s, want) {
			t.Errorf("report %q is missing %q", s, want)
		}
	}
}
