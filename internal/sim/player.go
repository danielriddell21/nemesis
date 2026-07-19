package sim

import (
	"math"

	"github.com/danielriddell21/nemesis/internal/world"
)

const (
	sneakSpeed = 1.3
	walkSpeed  = 2.6
	runSpeed   = 4.2
	crawlSpeed = 1.1
	turnSpeed  = 2.6
	reach      = 0.9

	sneakNoise   = 1.5
	walkNoise    = 4.0
	runNoise     = 9.0
	crawlNoise   = 6.0
	doorNoise    = 6.0
	pingNoise    = 5.0
	consoleNoise = 12.0

	killRange = 0.7
)

type Player struct {
	Pos    Vec2
	Angle  float64
	Moving bool
	Mode   MoveMode

	stepAcc float64
}

func (p Player) Dir() Vec2 {
	return Vec2{X: math.Cos(p.Angle), Y: math.Sin(p.Angle)}
}

func (p Player) InVent(w *World) bool {
	c := p.Pos.Cell()
	return w.Level.At(c.X, c.Y) == world.TileVent
}

func (g *Game) tickPlayer(in Input, dt float64) noiseEvent {
	p := &g.Player
	p.Angle = normalizeAngle(p.Angle + in.Turn*turnSpeed*dt + in.TurnDelta)
	p.Mode = in.Mode

	speed, loud := moveProfile(in.Mode)
	if p.InVent(g.World) {
		speed, loud = crawlSpeed, crawlNoise
	}
	if in.Tracker {
		// The tracker takes both hands: raising it caps you to a creep.
		speed = math.Min(speed, sneakSpeed)
	}

	dir := p.Dir()
	strafeX, strafeY := -dir.Y, dir.X
	dx := (dir.X*in.Forward + strafeX*in.Strafe) * speed * dt
	dy := (dir.Y*in.Forward + strafeY*in.Strafe) * speed * dt
	moved := resolveMove(g.World, p.Pos, dx, dy)
	p.Moving = moved.Sub(p.Pos).Len() > 1e-9
	p.Pos = moved

	if !p.Moving {
		p.stepAcc = 0
		return noiseEvent{}
	}
	kind := ObsStep
	if p.InVent(g.World) {
		kind = ObsVentCreak
	}
	// Footsteps land on a cadence proportional to speed; each one is a noise
	// event the hunter can hear and the audio layer can play.
	p.stepAcc += speed * dt
	if p.stepAcc < stepLength {
		return noiseEvent{at: p.Pos, radius: loud * continuousNoiseScale, kind: kind}
	}
	p.stepAcc -= stepLength
	g.observe(Observation{Kind: kind, At: p.Pos.Cell(), Radius: loud})
	return noiseEvent{at: p.Pos, radius: loud, kind: kind}
}

const (
	stepLength = 0.75
	// Between footsteps the player still leaks a fraction of the step noise, so
	// creeping right next to the hunter is never entirely free.
	continuousNoiseScale = 0.35
)

func moveProfile(m MoveMode) (speed, noise float64) {
	switch m {
	case ModeSneak:
		return sneakSpeed, sneakNoise
	case ModeRun:
		return runSpeed, runNoise
	default:
		return walkSpeed, walkNoise
	}
}

func (g *Game) interact(noise *noiseEvent) {
	p := g.Player
	ahead := Vec2{X: p.Pos.X + p.Dir().X*reach, Y: p.Pos.Y + p.Dir().Y*reach}
	c := ahead.Cell()
	switch g.World.Level.At(c.X, c.Y) {
	case world.TileDoor:
		if !g.World.DoorOpen(c) {
			g.World.OpenDoor(c)
			g.observe(Observation{Kind: ObsDoorOpen, At: c, Radius: doorNoise})
			noise.merge(noiseEvent{at: cellCenter(c), radius: doorNoise, kind: ObsDoorOpen})
		}
	case world.TileConsole:
		if !g.activated[c] {
			g.activated[c] = true
			g.director.escalate(g)
			g.observe(Observation{Kind: ObsConsole, At: c, Radius: consoleNoise})
			noise.merge(noiseEvent{at: cellCenter(c), radius: consoleNoise, kind: ObsConsole})
		}
	}
}
