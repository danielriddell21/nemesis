package sim

import (
	"math"

	"github.com/danielriddell21/nemesis/internal/world"
)

const (
	torchLife       = 80.0
	TorchLowMark    = 0.2
	batteryCharge   = 0.5
	torchSightRange = 16.0
	glowRange       = 8.0
	glowCooldown    = 1.5
)

func (g *Game) tickTorch(in Input, dt float64) {
	p := &g.Player
	pressed := in.Torch && !g.torchLatch
	g.torchLatch = in.Torch
	if pressed && !p.Hidden {
		g.setTorch(!p.Torch && p.Battery > 0)
	}
	if !p.Torch {
		return
	}
	p.Battery = math.Max(0, p.Battery-dt/torchLife)
	if p.Battery == 0 {
		g.setTorch(false)
	}
}

func (g *Game) setTorch(on bool) {
	if g.Player.Torch == on {
		return
	}
	g.Player.Torch = on
	g.observe(Observation{Kind: ObsTorch, At: g.Player.Pos.Cell()})
}

func (g *Game) TorchLit() bool {
	return g.TorchStrength() > 0
}

func (g *Game) TorchStrength() float64 {
	p := g.Player
	if !p.Torch || p.Battery <= 0 {
		return 0
	}
	if p.Battery >= TorchLowMark {
		return 1
	}
	// A dying cell: the beam dims and cuts out in irregular stutters that
	// grow longer as the charge runs down.
	drained := 1 - p.Battery/TorchLowMark
	t := g.elapsed
	beat := math.Sin(t*13) + math.Sin(t*7.3+1.7) + math.Sin(t*29*math.Pi/10)
	if beat < -1.2+1.4*drained {
		return 0
	}
	return 1 - 0.4*drained
}

func (g *Game) BatteryTaken(c world.Coord) bool { return g.taken[c] }

func (g *Game) collectBattery() {
	p := &g.Player
	c := p.Pos.Cell()
	if p.Hidden || p.Battery >= 1 || g.taken[c] || !g.World.Level.BatteryAt(c) {
		return
	}
	g.taken[c] = true
	p.Battery = math.Min(1, p.Battery+batteryCharge)
	g.observe(Observation{Kind: ObsBattery, At: c})
}

func (g *Game) alienNoticesGlow(dt float64) bool {
	// A lit torch throws light the hunter can catch from any angle: it does
	// not know what the glow is, only that it is worth a look.
	g.glowTimer = math.Max(0, g.glowTimer-dt)
	if g.glowTimer > 0 || !g.TorchLit() {
		return false
	}
	if g.Alien.Pos.Sub(g.Player.Pos).Len() > glowRange || !g.lineOfSight(g.Alien.Pos, g.Player.Pos) {
		return false
	}
	g.glowTimer = glowCooldown
	return true
}
