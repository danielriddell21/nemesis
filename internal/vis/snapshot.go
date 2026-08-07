package vis

import "github.com/danielriddell21/nemesis/internal/sim"

// Snapshot reads the game's current state into the message the visualiser
// draws from. It is the one place that knows how a running game maps onto the
// wire format, so the live window and a recording see exactly the same thing.
func Snapshot(s *sim.Game) StateMsg {
	path := s.Alien.Path()
	cells := make([][2]int, 0, len(path))
	for _, c := range path {
		cells = append(cells, [2]int{c.X, c.Y})
	}
	ping, vent, search := s.LearnTiers()
	decoy, locker := s.LearnExtras()
	var hot [][2]int
	for i, h := range s.RoomHeat() {
		if h > 0.5 {
			hot = append(hot, [2]int{i, int(h * 100)})
		}
	}
	d := s.DecoyState()
	return StateMsg{
		Tick:        s.TickCount(),
		PlayerX:     s.Player.Pos.X,
		PlayerY:     s.Player.Pos.Y,
		PlayerA:     s.Player.Angle,
		Hidden:      s.Player.Hidden,
		AlienX:      s.Alien.Pos.X,
		AlienY:      s.Alien.Pos.Y,
		AlienA:      s.Alien.Facing,
		AlienState:  s.Alien.State.String(),
		Vision:      s.VisionRange(),
		VisionFOV:   s.VisionFOV(),
		TargetX:     s.Alien.Target.X,
		TargetY:     s.Alien.Target.Y,
		Path:        cells,
		Menace:      s.Menace(),
		Done:        s.ObjectivesDone(),
		Total:       s.ObjectivesTotal(),
		Deck:        s.Depth() + 1,
		Unlocked:    s.ExitUnlocked(),
		PingTier:    ping,
		VentTier:    vent,
		SearchTier:  search,
		DecoyTier:   decoy,
		LockerTier:  locker,
		Hot:         hot,
		DecoyActive: d.Active,
		DecoyX:      d.Pos.X,
		DecoyY:      d.Pos.Y,
		Dead:        s.Dead(),
		Escaped:     s.Escaped(),
	}
}
