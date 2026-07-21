package render

import (
	"image/color"
	"math"

	"github.com/danielriddell21/crucible/geom"
	"github.com/danielriddell21/crucible/raycast"

	"github.com/danielriddell21/nemesis/internal/sim"
	"github.com/danielriddell21/nemesis/internal/world"
)

func (r *Renderer) drawWalls(g *sim.Game, cam raycast.Camera, now float64) {
	w, h := r.cfg.Width, r.cfg.Height
	for x := range w {
		rayX, rayY := cam.RayDir(x, w)
		hit := castRay(g, cam.Pos, rayX, rayY)
		r.zbuf[x] = hit.Dist

		lineHeight := int(float64(h) / hit.Dist)
		top := h/2 - lineHeight/2
		bottom := h/2 + lineHeight/2

		tex := r.wallTexture(g, hit.Cell)
		texX := int(hit.WallX * texSize)
		light := r.faceLight(g, hit, now)
		if hit.Side == 1 {
			light *= 0.8
		}
		step := float64(tex.h) / float64(lineHeight)
		texPos := 0.0
		if top < 0 {
			texPos = float64(-top) * step
			top = 0
		}
		if bottom > h {
			bottom = h
		}
		for y := top; y < bottom; y++ {
			c := tex.at(texX, int(texPos))
			texPos += step
			r.putShaded(x, y, c, light, hit.Dist)
		}
	}
}

func castRay(g *sim.Game, origin geom.Vec2, rayX, rayY float64) raycast.Hit {
	mapX, mapY := int(math.Floor(origin.X)), int(math.Floor(origin.Y))
	deltaX, deltaY := math.Abs(1/rayX), math.Abs(1/rayY)
	var stepX, stepY int
	var sideX, sideY float64
	if rayX < 0 {
		stepX, sideX = -1, (origin.X-float64(mapX))*deltaX
	} else {
		stepX, sideX = 1, (float64(mapX)+1-origin.X)*deltaX
	}
	if rayY < 0 {
		stepY, sideY = -1, (origin.Y-float64(mapY))*deltaY
	} else {
		stepY, sideY = 1, (float64(mapY)+1-origin.Y)*deltaY
	}
	side := 0
	prevX, prevY := mapX, mapY
	for range 256 {
		prevX, prevY = mapX, mapY
		if sideX < sideY {
			sideX += deltaX
			mapX += stepX
			side = 0
		} else {
			sideY += deltaY
			mapY += stepY
			side = 1
		}
		cell := world.Coord{X: mapX, Y: mapY}
		if g.World.Level.At(mapX, mapY) == world.TileDoor && g.World.DoorOpen(cell) {
			if hit, blocked := doorColumn(g, origin, rayX, rayY, sideX, sideY, deltaX, deltaY, side, cell, prevX, prevY); blocked {
				return hit
			}
			continue
		}
		if g.World.Solid(mapX, mapY) {
			break
		}
	}
	dist := max(raycast.BoundaryDist(sideX, sideY, deltaX, deltaY, side), 1e-4)
	return raycast.Hit{
		Cell:  world.Coord{X: mapX, Y: mapY},
		Prev:  world.Coord{X: prevX, Y: prevY},
		Dist:  dist,
		Side:  side,
		WallX: raycast.BoundaryWallX(origin, rayX, rayY, dist, side),
	}
}

// doorColumn decides whether a ray crossing an opening door is stopped by the
// sliding panel (returning the hit) or slips through its retracted part.
func doorColumn(g *sim.Game, origin geom.Vec2, rayX, rayY, sideX, sideY, deltaX, deltaY float64, side int, cell world.Coord, prevX, prevY int) (raycast.Hit, bool) {
	slide := g.World.DoorSlide(cell)
	if slide >= 1 {
		return raycast.Hit{}, false // fully retracted: the doorway is open
	}
	d := raycast.BoundaryDist(sideX, sideY, deltaX, deltaY, side)
	wx := raycast.BoundaryWallX(origin, rayX, rayY, d, side)
	if wx < slide {
		return raycast.Hit{}, false // this column has retracted; the ray passes
	}
	return raycast.Hit{Cell: cell, Prev: world.Coord{X: prevX, Y: prevY}, Dist: max(d, 1e-4), Side: side, WallX: wx}, true
}

func (r *Renderer) wallTexture(g *sim.Game, cell world.Coord) *texture {
	switch g.World.Level.At(cell.X, cell.Y) {
	case world.TileDoor:
		return r.tex.door
	case world.TileConsole:
		if g.ConsoleActivated(cell) {
			return r.tex.consoleActive
		}
		return r.tex.console
	default:
		return r.tex.wall
	}
}

func (r *Renderer) faceLight(g *sim.Game, hit raycast.Hit, now float64) float64 {
	// A wall face borrows the light of the open cell the ray crossed last, so
	// faces bounding a dark corridor stay dark.
	l := g.World.Level
	light := l.LightAt(hit.Prev.X, hit.Prev.Y)
	if light <= 0 {
		light = 0.3
	}
	if l.At(hit.Prev.X, hit.Prev.Y) == world.TileVent {
		light = math.Min(light, 0.2)
	}
	if l.FlickerAt(hit.Prev.X, hit.Prev.Y) {
		light *= flicker(now)
	}
	return light
}

func flicker(now float64) float64 {
	// A failing fixture: mostly on, with irregular dips driven by beating
	// sine waves so the stutter never quite repeats.
	v := math.Sin(now*13) + math.Sin(now*7.3+1.7) + math.Sin(now*29*math.Pi/10)
	if v < -1.2 {
		return 0.25
	}
	return 1
}

func (r *Renderer) putShaded(x, y int, c color.RGBA, light, dist float64) {
	k := light * 1.25 / (1 + dist*dist*0.035)
	if k > 1 {
		k = 1
	}
	i := (y*r.cfg.Width + x) * 4
	r.fb[i] = uint8(float64(c.R) * k)
	r.fb[i+1] = uint8(float64(c.G) * k)
	r.fb[i+2] = uint8(float64(c.B) * k)
	r.fb[i+3] = 255
}
