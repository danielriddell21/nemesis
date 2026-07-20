package render

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"github.com/danielriddell21/nemesis/internal/hud"
	"github.com/danielriddell21/nemesis/internal/sim"
)

const (
	hudMarginX  = 4
	hudBaseline = 13
	glyphWidth  = 7
)

func (r *Renderer) drawHUD(g *sim.Game, now float64) {
	r.drawStatusLine(g)
	r.drawOverlayMessage()
	r.drawTracker(g, now)
	switch {
	case g.Dead():
		r.drawTextCentered(r.cfg.Height/2-8, "IT FOUND YOU", palette.exitLocked)
	case g.Escaped():
		r.drawTextCentered(r.cfg.Height/2-8, "AIRLOCK CYCLED - YOU ESCAPED", palette.exitOpen)
	}
}

func (r *Renderer) drawStatusLine(g *sim.Game) {
	status := fmt.Sprintf("%s   SYSTEMS %d/%d   DECOYS %d", gaitLabel(g), g.ObjectivesDone(), g.ObjectivesTotal(), g.Player.Decoys)
	if g.ExitUnlocked() {
		status += "   AIRLOCK OPEN"
	}
	if g.Depth() > 0 {
		status = fmt.Sprintf("DECK %d   %s", g.Depth()+1, status)
	}
	y := r.cfg.Height - 8
	r.drawText(hudMarginX+1, y+1, status, palette.hudDrop)
	r.drawText(hudMarginX, y, status, palette.hudDim)

	if prompt := actionPrompt(g); prompt != "" {
		x := (r.cfg.Width - len(prompt)*glyphWidth) / 2
		r.drawText(x+1, r.cfg.Height-24+1, prompt, palette.hudDrop)
		r.drawText(x, r.cfg.Height-24, prompt, palette.tracker)
	}
}

func actionPrompt(g *sim.Game) string {
	if g.Player.Hidden {
		return "[F] LEAVE LOCKER"
	}
	if g.Player.OnLocker(g.World) {
		return "[F] HIDE"
	}
	return ""
}

func gaitLabel(g *sim.Game) string {
	if g.Player.Hidden {
		return "HIDDEN"
	}
	if g.Player.InVent(g.World) {
		return "CRAWL"
	}
	switch g.Player.Mode {
	case sim.ModeSneak:
		return "SNEAK"
	case sim.ModeRun:
		return "RUN"
	default:
		return "WALK"
	}
}

func (r *Renderer) drawOverlayMessage() {
	if r.overlay == nil {
		return
	}
	msg, ch, ok := r.overlay.Active()
	if !ok {
		return
	}
	fg := palette.hudText
	if ch == hud.Diagnostic {
		fg = palette.hudDim
	}
	r.drawText(hudMarginX+1, hudBaseline+1, msg, palette.hudDrop)
	r.drawText(hudMarginX, hudBaseline, msg, fg)
}

func (r *Renderer) drawTracker(g *sim.Game, now float64) {
	if !g.Tracker.Raised {
		return
	}
	cx := r.cfg.Width / 2
	cy := r.cfg.Height - 30
	radius := r.cfg.Height / 5

	r.drawArc(cx, cy, radius, palette.tracker)
	r.drawArc(cx, cy, radius/2, dimmed(palette.tracker, 0.5))
	// Forward tick so the scope reads as facing up-screen.
	for d := 0; d <= radius; d++ {
		r.putPixel(cx, cy-d, dimmed(palette.tracker, 0.35))
	}

	c := g.Tracker.Contact
	if c.Valid && c.Age < sim.PingInterval() {
		// Bearing 0 is dead ahead (up-screen); fade the blip as it ages.
		k := 1 - c.Age/sim.PingInterval()
		px := cx + int(math.Sin(c.Bearing)*c.Dist/sim.TrackerRange()*float64(radius))
		py := cy - int(math.Cos(c.Bearing)*c.Dist/sim.TrackerRange()*float64(radius))
		r.drawBlob(px, py, 3, dimmed(palette.blip, 0.3+0.7*k))
		label := fmt.Sprintf("%2.0fM", c.Dist)
		r.drawTextCentered(cy+12, label, palette.tracker)
	} else if math.Mod(now, sim.PingInterval()) < 0.1 {
		// The sweep flash on an empty return.
		r.drawBlob(cx, cy, 2, dimmed(palette.tracker, 0.8))
	}
}

func (r *Renderer) drawArc(cx, cy, radius int, c color.RGBA) {
	for i := 0; i <= 180; i++ {
		a := float64(i) / 180 * math.Pi
		x := cx + int(math.Cos(a)*float64(radius))
		y := cy - int(math.Sin(a)*float64(radius))
		r.putPixel(x, y, c)
	}
}

func (r *Renderer) drawBlob(cx, cy, radius int, c color.RGBA) {
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy <= radius*radius {
				r.putPixel(cx+dx, cy+dy, c)
			}
		}
	}
}

func (r *Renderer) putPixel(x, y int, c color.RGBA) {
	if x < 0 || y < 0 || x >= r.cfg.Width || y >= r.cfg.Height {
		return
	}
	i := (y*r.cfg.Width + x) * 4
	r.fb[i], r.fb[i+1], r.fb[i+2], r.fb[i+3] = c.R, c.G, c.B, 255
}

func (r *Renderer) drawText(x, y int, s string, c color.RGBA) {
	d := &font.Drawer{
		Dst:  r.framebufferImage(),
		Src:  image.NewUniform(c),
		Face: basicfont.Face7x13,
		Dot:  fixed.P(x, y),
	}
	d.DrawString(s)
}

func (r *Renderer) drawTextCentered(y int, s string, c color.RGBA) {
	x := (r.cfg.Width - len(s)*glyphWidth) / 2
	r.drawText(x+1, y+1, s, palette.hudDrop)
	r.drawText(x, y, s, c)
}

func (r *Renderer) framebufferImage() *image.RGBA {
	return &image.RGBA{
		Pix:    r.fb,
		Stride: r.cfg.Width * 4,
		Rect:   image.Rect(0, 0, r.cfg.Width, r.cfg.Height),
	}
}

func dimmed(c color.RGBA, k float64) color.RGBA {
	return color.RGBA{R: uint8(float64(c.R) * k), G: uint8(float64(c.G) * k), B: uint8(float64(c.B) * k), A: 255}
}
