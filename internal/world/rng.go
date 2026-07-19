package world

import "math/rand/v2"

type rng struct {
	r *rand.Rand
}

func newRNG(seed int64) *rng {
	u := uint64(seed)
	return &rng{r: rand.New(rand.NewPCG(u, u^0x9e3779b97f4a7c15))}
}

func (g *rng) intn(n int) int {
	return g.r.IntN(n)
}

func (g *rng) between(lo, hi int) int {
	if hi <= lo {
		return lo
	}
	return lo + g.r.IntN(hi-lo+1)
}

func (g *rng) betweenF(lo, hi float64) float64 {
	if hi <= lo {
		return lo
	}
	return lo + g.r.Float64()*(hi-lo)
}

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
