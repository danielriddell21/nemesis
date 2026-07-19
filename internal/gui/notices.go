package gui

import (
	"fmt"
	"math"

	"github.com/danielriddell21/nemesis/internal/sim"
	"github.com/danielriddell21/nemesis/internal/telemetry"
)

const (
	noticeFrames = 150
	earshotCreak = 12.0
	earshotDoor  = 10.0
)

func notice(e telemetry.Event, s *sim.Game) (string, int, bool) {
	kind, ok := e.Kind()
	if !ok {
		return "", 0, false
	}
	d := distanceToPlayer(e, s)
	switch kind {
	case sim.ObsConsole:
		return fmt.Sprintf("GENERATOR ONLINE %d/%d", s.ObjectivesDone(), s.ObjectivesTotal()), noticeFrames, true
	case sim.ObsAlienCreak:
		// Only what the player could plausibly hear becomes a message.
		if d <= earshotCreak {
			return "A VENT CREAKS NEARBY", noticeFrames, true
		}
	case sim.ObsDoorOpen:
		if d <= earshotDoor && !nearPlayer(e, s, 1.5) {
			return "A BULKHEAD OPENS SOMEWHERE", noticeFrames, true
		}
	case sim.ObsAlienSeen:
		return "IT SEES YOU - RUN", noticeFrames, true
	}
	return "", 0, false
}

func distanceToPlayer(e telemetry.Event, s *sim.Game) float64 {
	return math.Hypot(s.Player.Pos.X-float64(e.X)-0.5, s.Player.Pos.Y-float64(e.Y)-0.5)
}

func nearPlayer(e telemetry.Event, s *sim.Game, r float64) bool {
	return distanceToPlayer(e, s) <= r
}
