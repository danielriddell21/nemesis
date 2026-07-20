package render

import (
	"image/color"
	"math"

	"github.com/danielriddell21/nemesis/internal/sim"
	"github.com/danielriddell21/nemesis/internal/world"
)

func (r *Renderer) drawWalls(g *sim.Game, cam camera, now float64) {
	w, h := r.cfg.Width, r.cfg.Height
	for x := range w {
		cameraX := 2*float64(x)/float64(w) - 1
		rayX := cam.dirX + cam.planeX*cameraX
		rayY := cam.dirY + cam.planeY*cameraX
		hit := castRay(g, cam.pos, rayX, rayY)
		r.zbuf[x] = hit.dist

		lineHeight := int(float64(h) / hit.dist)
		top := h/2 - lineHeight/2
		bottom := h/2 + lineHeight/2

		tex := r.wallTexture(g, hit.cell)
		texX := int(hit.wallX * texSize)
		light := r.faceLight(g, hit, now)
		if hit.side == 1 {
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
			r.putShaded(x, y, c, light, hit.dist)
		}
	}
}

type rayHit struct {
	cell  world.Coord
	prev  world.Coord
	dist  float64
	side  int
	wallX float64
}

func castRay(g *sim.Game, pos sim.Vec2, rayX, rayY float64) rayHit {
	mapX, mapY := int(math.Floor(pos.X)), int(math.Floor(pos.Y))
	deltaX, deltaY := math.Abs(1/rayX), math.Abs(1/rayY)
	var stepX, stepY int
	var sideX, sideY float64
	if rayX < 0 {
		stepX, sideX = -1, (pos.X-float64(mapX))*deltaX
	} else {
		stepX, sideX = 1, (float64(mapX)+1-pos.X)*deltaX
	}
	if rayY < 0 {
		stepY, sideY = -1, (pos.Y-float64(mapY))*deltaY
	} else {
		stepY, sideY = 1, (float64(mapY)+1-pos.Y)*deltaY
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
			if hit, blocked := doorColumn(g, pos, rayX, rayY, sideX, sideY, deltaX, deltaY, side, cell, prevX, prevY); blocked {
				return hit
			}
			continue
		}
		if g.World.Solid(mapX, mapY) {
			break
		}
	}
	dist := max(boundaryDist(sideX, sideY, deltaX, deltaY, side), 1e-4)
	return rayHit{
		cell:  world.Coord{X: mapX, Y: mapY},
		prev:  world.Coord{X: prevX, Y: prevY},
		dist:  dist,
		side:  side,
		wallX: boundaryWallX(pos, rayX, rayY, dist, side),
	}
}

// doorColumn decides whether a ray crossing an opening door is stopped by the
// sliding panel (returning the hit) or slips through its retracted part.
func doorColumn(g *sim.Game, pos sim.Vec2, rayX, rayY, sideX, sideY, deltaX, deltaY float64, side int, cell world.Coord, prevX, prevY int) (rayHit, bool) {
	slide := g.World.DoorSlide(cell)
	if slide >= 1 {
		return rayHit{}, false // fully retracted: the doorway is open
	}
	d := boundaryDist(sideX, sideY, deltaX, deltaY, side)
	wx := boundaryWallX(pos, rayX, rayY, d, side)
	if wx < slide {
		return rayHit{}, false // this column has retracted; the ray passes
	}
	return rayHit{cell: cell, prev: world.Coord{X: prevX, Y: prevY}, dist: max(d, 1e-4), side: side, wallX: wx}, true
}

func boundaryDist(sideX, sideY, deltaX, deltaY float64, side int) float64 {
	if side == 0 {
		return sideX - deltaX
	}
	return sideY - deltaY
}

func boundaryWallX(pos sim.Vec2, rayX, rayY, dist float64, side int) float64 {
	var wx float64
	if side == 0 {
		wx = pos.Y + dist*rayY
	} else {
		wx = pos.X + dist*rayX
	}
	return wx - math.Floor(wx)
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

func (r *Renderer) faceLight(g *sim.Game, hit rayHit, now float64) float64 {
	// A wall face borrows the light of the open cell the ray crossed last, so
	// faces bounding a dark corridor stay dark.
	l := g.World.Level
	light := l.LightAt(hit.prev.X, hit.prev.Y)
	if light <= 0 {
		light = 0.3
	}
	if l.At(hit.prev.X, hit.prev.Y) == world.TileVent {
		light = math.Min(light, 0.2)
	}
	if l.FlickerAt(hit.prev.X, hit.prev.Y) {
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
