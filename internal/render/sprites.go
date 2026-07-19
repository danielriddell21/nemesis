package render

import (
	"image/color"

	"github.com/danielriddell21/nemesis/internal/sim"
)

type billboard struct {
	pos     sim.Vec2
	tex     *texture
	tint    color.RGBA
	useTint bool
	scale   float64
}

func (r *Renderer) drawSprites(g *sim.Game, cam camera) {
	boards := []billboard{
		{pos: g.Alien.Pos, tex: r.tex.alien, scale: 1},
		r.exitBeacon(g),
	}
	for _, b := range boards {
		r.drawBillboard(b, cam)
	}
}

func (r *Renderer) exitBeacon(g *sim.Game) billboard {
	c := palette.exitLocked
	if g.ExitUnlocked() {
		c = palette.exitOpen
	}
	return billboard{
		pos:     sim.Vec2{X: float64(g.World.Level.Exit.X) + 0.5, Y: float64(g.World.Level.Exit.Y) + 0.5},
		tint:    c,
		useTint: true,
		scale:   0.6,
	}
}

func (r *Renderer) drawBillboard(b billboard, cam camera) {
	w, h := r.cfg.Width, r.cfg.Height
	relX := b.pos.X - cam.pos.X
	relY := b.pos.Y - cam.pos.Y

	// Inverse camera transform into screen space.
	invDet := 1 / (cam.planeX*cam.dirY - cam.dirX*cam.planeY)
	transX := invDet * (cam.dirY*relX - cam.dirX*relY)
	transY := invDet * (-cam.planeY*relX + cam.planeX*relY)
	if transY <= 0.1 {
		return
	}
	screenX := int(float64(w) / 2 * (1 + transX/transY))
	size := int(float64(h) / transY * b.scale)
	if size < 2 {
		return
	}
	top := h/2 + int(float64(h)/transY)/2 - size
	drawStart := max(top, 0)
	drawEnd := min(top+size, h)
	left := screenX - size/4
	right := screenX + size/4
	for x := max(left, 0); x < min(right, w); x++ {
		if transY >= r.zbuf[x] {
			continue
		}
		texX := (x - left) * 32 / max(right-left, 1)
		for y := drawStart; y < drawEnd; y++ {
			c, ok := r.spriteTexel(b, texX, (y-top)*64/size)
			if !ok {
				continue
			}
			r.putShaded(x, y, c, 0.9, transY)
		}
	}
}

func (r *Renderer) spriteTexel(b billboard, u, v int) (color.RGBA, bool) {
	if b.useTint {
		// Beacon: a translucent light column rendered as a dithered pillar.
		if (u+v)%2 == 0 {
			return color.RGBA{}, false
		}
		return b.tint, true
	}
	c := b.tex.at(u, v)
	if c.A == 0 {
		return color.RGBA{}, false
	}
	return c, true
}
