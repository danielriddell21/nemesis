package sim

const lockerNoticeRange = 9.0

func (g *Game) tickHide(in Input) {
	if !in.Hide || g.hideLatch {
		g.hideLatch = in.Hide
		return
	}
	g.hideLatch = true
	p := &g.Player
	if p.Hidden {
		p.Hidden = false
		return
	}
	if !p.OnLocker(g.World) {
		return
	}
	p.Hidden = true
	g.observe(Observation{Kind: ObsHide, At: p.Pos.Cell()})
	// If the hunter is actively looking and close by when the prey ducks in, it
	// learns that lockers are worth checking.
	if g.hunterAware() && g.Alien.Pos.Sub(p.Pos).Len() < lockerNoticeRange {
		g.learnLocker()
	}
}

func (g *Game) hunterAware() bool {
	switch g.Alien.State {
	case StateHunt, StateInvestigate, StateSearch:
		return true
	default:
		return false
	}
}
