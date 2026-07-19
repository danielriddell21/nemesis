package render

import (
	"image/color"
	"math"

	"github.com/danielriddell21/nemesis/internal/hud"
	"github.com/danielriddell21/nemesis/internal/sim"
)

type Renderer struct {
	cfg     Config
	fb      []byte
	zbuf    []float64
	tex     *textureSet
	overlay *hud.Overlay
}

type Option func(*Renderer)

func WithOverlay(o *hud.Overlay) Option {
	return func(r *Renderer) { r.overlay = o }
}

func NewRenderer(cfg Config, opts ...Option) *Renderer {
	r := &Renderer{
		cfg:  cfg,
		fb:   make([]byte, cfg.Width*cfg.Height*4),
		zbuf: make([]float64, cfg.Width),
		tex:  buildTextures(),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *Renderer) Config() Config { return r.cfg }

func (r *Renderer) Frame(g *sim.Game, now float64) []byte {
	cam := newCamera(g.Player.Pos, g.Player.Angle, r.cfg.FOV)
	r.drawBackdrop()
	r.drawWalls(g, cam, now)
	r.drawSprites(g, cam)
	r.drawEffects(g, now)
	r.drawHUD(g, now)
	return r.fb
}

type camera struct {
	pos            sim.Vec2
	dirX, dirY     float64
	planeX, planeY float64
}

func newCamera(pos sim.Vec2, angle, fov float64) camera {
	planeLen := math.Tan(fov / 2)
	dirX, dirY := math.Cos(angle), math.Sin(angle)
	return camera{
		pos:    pos,
		dirX:   dirX,
		dirY:   dirY,
		planeX: -dirY * planeLen,
		planeY: dirX * planeLen,
	}
}

func (r *Renderer) drawBackdrop() {
	w, h := r.cfg.Width, r.cfg.Height
	half := h / 2
	for y := range h {
		// Rows darken toward the horizon: distant floor and ceiling recede
		// into the station gloom.
		var c [4]byte
		if y < half {
			depth := float64(half-y) / float64(half)
			c = shadeBytes(palette.ceiling, 0.25+0.75*depth)
		} else {
			depth := float64(y-half+1) / float64(half)
			c = shadeBytes(palette.floor, 0.25+0.75*depth)
		}
		row := r.fb[y*w*4 : (y+1)*w*4]
		for x := 0; x < w*4; x += 4 {
			row[x], row[x+1], row[x+2], row[x+3] = c[0], c[1], c[2], c[3]
		}
	}
}

func (r *Renderer) drawEffects(g *sim.Game, now float64) {
	inVent := g.Player.InVent(g.World)
	if inVent {
		r.letterbox()
	}
	switch {
	case g.Dead():
		r.tint(palette.deathTint, 0.55)
	case g.Escaped():
		r.tint(palette.escapeTint, 0.45)
	case g.Menace() > 0.65:
		// The closer the hunter prowls, the harder the screen pulses.
		alpha := (g.Menace() - 0.65) / 0.35 * (0.08 + 0.05*math.Sin(now*6))
		if alpha > 0 {
			r.tint(palette.deathTint, alpha)
		}
	}
}

func (r *Renderer) letterbox() {
	// Crawling through a duct squeezes the view between black bars.
	w, h := r.cfg.Width, r.cfg.Height
	bar := h / 6
	for y := range h {
		if y >= bar && y < h-bar {
			continue
		}
		row := r.fb[y*w*4 : (y+1)*w*4]
		for x := range row {
			if x%4 != 3 {
				row[x] = 0
			}
		}
	}
}

func (r *Renderer) tint(c color.RGBA, alpha float64) {
	for i := 0; i < len(r.fb); i += 4 {
		r.fb[i] = blendByte(r.fb[i], c.R, alpha)
		r.fb[i+1] = blendByte(r.fb[i+1], c.G, alpha)
		r.fb[i+2] = blendByte(r.fb[i+2], c.B, alpha)
	}
}

func blendByte(dst, src uint8, alpha float64) uint8 {
	return uint8(float64(dst)*(1-alpha) + float64(src)*alpha)
}

func shadeBytes(c color.RGBA, k float64) [4]byte {
	return [4]byte{
		uint8(float64(c.R) * k),
		uint8(float64(c.G) * k),
		uint8(float64(c.B) * k),
		255,
	}
}
