package world

import "math/rand/v2"

// rng is a small deterministic random source used throughout generation. Seeding
// it with the same value always yields the same sequence, which is what makes a
// level reproducible from its seed.
type rng struct {
	r *rand.Rand
}

// newRNG builds a deterministic source from a seed. The two PCG stream words are
// derived from the seed so a single int64 fully determines the sequence.
func newRNG(seed int64) *rng {
	u := uint64(seed)
	return &rng{r: rand.New(rand.NewPCG(u, u^0x9e3779b97f4a7c15))}
}

// intn returns a pseudo-random int in [0, n). It panics if n <= 0, matching the
// standard library's contract.
func (g *rng) intn(n int) int {
	return g.r.IntN(n)
}

// between returns a pseudo-random int in [lo, hi]. If hi <= lo it returns lo.
func (g *rng) between(lo, hi int) int {
	if hi <= lo {
		return lo
	}
	return lo + g.r.IntN(hi-lo+1)
}

// betweenF returns a pseudo-random float64 in [lo, hi). If hi <= lo it returns lo.
func (g *rng) betweenF(lo, hi float64) float64 {
	if hi <= lo {
		return lo
	}
	return lo + g.r.Float64()*(hi-lo)
}

// chance reports true with probability p (clamped to [0, 1]).
func (g *rng) chance(p float64) bool {
	switch {
	case p <= 0:
		return false
	case p >= 1:
		return true
	default:
		return g.r.Float64() < p
	}
}
