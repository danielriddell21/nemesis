package sim

import (
	"math"

	"github.com/danielriddell21/nemesis/internal/world"
)

const (
	visionRange = 13.0
	visionFOV   = 1.9

	hearingFloor = 0.85
)

func (g *Game) lineOfSight(a, b Vec2) bool {
	// Amanatides & Woo grid march from a to b; sight is clear when no cell on
	// the segment blocks it.
	d := b.Sub(a)
	dist := d.Len()
	if dist < 1e-9 {
		return true
	}
	x, y := int(math.Floor(a.X)), int(math.Floor(a.Y))
	endX, endY := int(math.Floor(b.X)), int(math.Floor(b.Y))
	stepX, tMaxX, tDeltaX := gridStep(a.X, d.X/dist)
	stepY, tMaxY, tDeltaY := gridStep(a.Y, d.Y/dist)
	for {
		if x == endX && y == endY {
			return true
		}
		if tMaxX < tMaxY {
			x += stepX
			tMaxX += tDeltaX
		} else {
			y += stepY
			tMaxY += tDeltaY
		}
		if math.Min(tMaxX, tMaxY) > dist+1 {
			return true
		}
		if (x != endX || y != endY) && g.World.BlocksSight(x, y) {
			return false
		}
	}
}

func gridStep(origin, dir float64) (step int, tMax, tDelta float64) {
	cell := math.Floor(origin)
	switch {
	case dir > 0:
		return 1, (cell + 1 - origin) / dir, 1 / dir
	case dir < 0:
		return -1, (origin - cell) / -dir, 1 / -dir
	default:
		return 0, math.Inf(1), math.Inf(1)
	}
}

func (g *Game) alienSees() bool {
	if g.Player.Hidden {
		// Inside a locker there is nothing to see; only a breach finds you.
		return false
	}
	return g.inSight(g.Player.Pos, g.effectiveVisionRange())
}

func (g *Game) inSight(pos Vec2, reach float64) bool {
	to := pos.Sub(g.Alien.Pos)
	dist := to.Len()
	if dist > reach {
		return false
	}
	if dist > 1e-9 {
		bearing := normalizeAngle(math.Atan2(to.Y, to.X) - g.Alien.Facing)
		if math.Abs(bearing) > visionFOV/2 {
			return false
		}
	}
	return g.lineOfSight(g.Alien.Pos, pos)
}

func (g *Game) Exposed(c world.Coord) bool {
	return g.inSight(cellCenter(c), g.visionRangeAt(c, false))
}

func (g *Game) effectiveVisionRange() float64 {
	still := !g.Player.Moving && g.Player.Mode == ModeSneak
	return g.visionRangeAt(g.Player.Pos.Cell(), still)
}

func (g *Game) visionRangeAt(c world.Coord, still bool) float64 {
	light := g.World.Level.LightAt(c.X, c.Y)
	// Darkness shortens the hunter's sight; holding still while sneaking
	// shortens it further. It always spots you at arm's length.
	r := visionRange * (0.4 + light)
	if still {
		r *= 0.55
	}
	if g.TorchLit() {
		// A lit torch is a beacon: no shadow or stillness hides the hand
		// holding it.
		r = math.Max(r, torchSightRange)
	}
	return math.Max(r, 1.6)
}

func (g *Game) alienHears(n noiseEvent) bool {
	if n.radius <= 0 {
		return false
	}
	if n.kind == ObsDecoy && g.decoyIgnored() {
		// A hunter that has seen through the trick no longer chases the chirp.
		return false
	}
	radius := n.radius * (hearingFloor + g.director.aggression*0.5)
	return g.Alien.Pos.Sub(n.at).Len() <= radius
}

func nearestWalkable(l *world.Level, c world.Coord) world.Coord {
	if l.At(c.X, c.Y).Walkable() {
		return c
	}
	best, bestD := c, math.MaxFloat64
	for y := range l.H {
		for x := range l.W {
			if !l.At(x, y).Walkable() {
				continue
			}
			d := math.Hypot(float64(x-c.X), float64(y-c.Y))
			if d < bestD {
				best, bestD = world.Coord{X: x, Y: y}, d
			}
		}
	}
	return best
}
