package render

import (
	"image/color"
	"math"

	"github.com/danielriddell21/crucible/geom"
	"github.com/danielriddell21/crucible/hud"
	"github.com/danielriddell21/crucible/paint"
	"github.com/danielriddell21/crucible/raycast"

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

func (r *Renderer) SetFOV(fov float64) { r.cfg.FOV = fov }

func (r *Renderer) Frame(g *sim.Game, now float64) []byte {
	cam := raycast.NewCamera(geom.Vec2{X: g.Player.Pos.X, Y: g.Player.Pos.Y}, g.Player.Angle, r.cfg.FOV)
	r.drawBackdrop()
	r.drawWalls(g, cam, now)
	r.drawSprites(g, cam, now)
	r.drawEffects(g, now)
	r.drawHUD(g, now)
	return r.fb
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
			c = paint.ScaleBytes(palette.ceiling, 0.25+0.75*depth)
		} else {
			depth := float64(y-half+1) / float64(half)
			c = paint.ScaleBytes(palette.floor, 0.25+0.75*depth)
		}
		row := r.fb[y*w*4 : (y+1)*w*4]
		for x := 0; x < w*4; x += 4 {
			row[x], row[x+1], row[x+2], row[x+3] = c[0], c[1], c[2], c[3]
		}
	}
}

func (r *Renderer) drawEffects(g *sim.Game, now float64) {
	if g.Player.Hidden {
		r.lockerView()
	} else if g.Player.InVent(g.World) {
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

func (r *Renderer) lockerView() {
	// Peering out through a louvred locker door: the world is darkened and seen
	// between horizontal slats, with a clear viewing gap at eye level.
	w, h := r.cfg.Width, r.cfg.Height
	slitTop, slitBottom := h*2/5, h*3/5
	for y := range h {
		darken := 0.82
		if y >= slitTop && y < slitBottom {
			darken = 0.25 // the eye-level gap you look through
		} else if (y/6)%2 == 0 {
			darken = 0.95 // the solid slats
		}
		row := r.fb[y*w*4 : (y+1)*w*4]
		for x := 0; x < w*4; x += 4 {
			row[x] = uint8(float64(row[x]) * (1 - darken))
			row[x+1] = uint8(float64(row[x+1]) * (1 - darken))
			row[x+2] = uint8(float64(row[x+2]) * (1 - darken))
		}
	}
}

func (r *Renderer) tint(c color.RGBA, alpha float64) {
	paint.BlendOver(r.fb, c, alpha)
}
