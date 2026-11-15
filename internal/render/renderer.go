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

	cone     []float32
	vignette []uint16
	wallTop  []int
	wallEnd  []int

	cellLight      []float64
	levelW, levelH int
	torch          float64
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
	r.buildMasks()
	return r
}

func (r *Renderer) Config() Config { return r.cfg }

func (r *Renderer) SetFOV(fov float64) { r.cfg.FOV = fov }

func (r *Renderer) Frame(g *sim.Game, now float64) []byte {
	cam := raycast.NewCamera(geom.Vec2{X: g.Player.Pos.X, Y: g.Player.Pos.Y}, g.Player.Angle, r.cfg.FOV)
	r.lightScene(g, now)
	r.drawWalls(g, cam)
	r.drawFloorAndCeiling(g, cam)
	r.drawSprites(g, cam, now)
	r.grade(now)
	r.drawEffects(g, now)
	r.drawHUD(g, now)
	return r.fb
}

func (r *Renderer) drawEffects(g *sim.Game, now float64) {
	if g.Player.Hidden {
		r.lockerView()
	} else if g.Player.InVent(g.World) {
		r.letterbox()
		r.ventPress()
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

func (r *Renderer) ventPress() {
	// The walls close in: the edges of the view fall away to black, as if the
	// duct were pressing on both shoulders.
	w, h := r.cfg.Width, r.cfg.Height
	for y := range h {
		for x := range w {
			edge := math.Abs(float64(x)/float64(w)*2 - 1)
			k := 1 - 0.65*smoothstep(0.45, 1, edge)
			i := (y*w + x) * 4
			r.fb[i] = uint8(float64(r.fb[i]) * k)
			r.fb[i+1] = uint8(float64(r.fb[i+1]) * k)
			r.fb[i+2] = uint8(float64(r.fb[i+2]) * k)
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
