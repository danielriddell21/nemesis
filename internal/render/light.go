package render

import (
	"image/color"
	"math"

	"github.com/danielriddell21/crucible/raycast"

	"github.com/danielriddell21/nemesis/internal/sim"
	"github.com/danielriddell21/nemesis/internal/world"
)

const (
	ambientGain    = 1.25
	ambientFalloff = 0.06
	beamGain       = 1.7
	beamFalloff    = 0.02
	beamSpill      = 0.05
	fogReach       = 9.0
	fogWash        = 0.6
	unlitCell      = 0.1
	ventCeiling    = 0.1
	grainDepth     = 5
	grainRate      = 24
)

var (
	ambientTint = [3]float64{0.92, 1, 0.96}
	beamTint    = [3]float64{1, 0.96, 0.82}
	murk        = [3]float64{10, 16, 14}
)

func (r *Renderer) buildMasks() {
	// Both masks depend only on the frame size, so they are laid out once: the
	// torch's cone on the screen, and the darkening toward the corners.
	w, h := r.cfg.Width, r.cfg.Height
	r.cone = make([]float32, w*h)
	r.vignette = make([]uint16, w*h)
	r.wallTop = make([]int, w)
	r.wallEnd = make([]int, w)
	cx, cy := float64(w)/2, float64(h)*0.52
	for y := range h {
		for x := range w {
			dx, dy := float64(x)-cx, float64(y)-cy
			beam := math.Hypot(dx, dy) / (float64(h) * 0.5)
			pool := 1 - smoothstep(0.45, 1.05, beam)
			hot := 0.25 * (1 - smoothstep(0, 0.3, beam))
			r.cone[y*w+x] = float32(beamSpill + (1-beamSpill)*pool + hot)

			edge := math.Hypot(dx/float64(w), (float64(y)-float64(h)/2)/float64(h)) * 2
			r.vignette[y*w+x] = uint16(256 * (1 - 0.5*smoothstep(0.85, 1.45, edge)))
		}
	}
}

func smoothstep(lo, hi, v float64) float64 {
	t := (v - lo) / (hi - lo)
	if t <= 0 {
		return 0
	}
	if t >= 1 {
		return 1
	}
	return t * t * (3 - 2*t)
}

func (r *Renderer) lightScene(g *sim.Game, now float64) {
	// One pass over the deck settles this frame's light per cell — the base
	// glow, with failing fixtures stuttered — so every pixel is a lookup.
	l := g.World.Level
	r.levelW, r.levelH = l.W, l.H
	if len(r.cellLight) != l.W*l.H {
		r.cellLight = make([]float64, l.W*l.H)
	}
	dip := flicker(now)
	for y := range l.H {
		for x := range l.W {
			v := l.LightAt(x, y)
			if l.At(x, y) == world.TileVent {
				v = math.Min(v, ventCeiling)
			}
			if l.FlickerAt(x, y) {
				v *= dip
			}
			r.cellLight[y*l.W+x] = v
		}
	}
	r.torch = g.TorchStrength()
}

func (r *Renderer) lightAt(x, y int) float64 {
	if x < 0 || y < 0 || x >= r.levelW || y >= r.levelH {
		return 0
	}
	return r.cellLight[y*r.levelW+x]
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
	// Two lights reach a surface: the station's failing ambient, cool and
	// short-ranged, and the torch's warm cone, which carries further. Distance
	// then washes the result toward the murk of stale air.
	d2 := dist * dist
	amb := light * ambientGain / (1 + d2*ambientFalloff)
	beam := r.torch * beamGain * float64(r.cone[y*r.cfg.Width+x]) / (1 + d2*beamFalloff)
	fog := fogWash * dist / (dist + fogReach)
	keep := 1 - fog
	lit := amb + beam
	i := (y*r.cfg.Width + x) * 4
	r.fb[i] = toByte(float64(c.R)*keep*(amb*ambientTint[0]+beam*beamTint[0]) + murk[0]*fog*lit)
	r.fb[i+1] = toByte(float64(c.G)*keep*(amb*ambientTint[1]+beam*beamTint[1]) + murk[1]*fog*lit)
	r.fb[i+2] = toByte(float64(c.B)*keep*(amb*ambientTint[2]+beam*beamTint[2]) + murk[2]*fog*lit)
	r.fb[i+3] = 255
}

func (r *Renderer) putGlow(x, y int, c color.RGBA, dist float64) {
	// Self-lit texels — an indicator lamp, an eye catching what light there
	// is — ignore the room and only fade with distance.
	k := 1 / (1 + dist*dist*0.012)
	i := (y*r.cfg.Width + x) * 4
	r.fb[i] = toByte(float64(c.R) * k)
	r.fb[i+1] = toByte(float64(c.G) * k)
	r.fb[i+2] = toByte(float64(c.B) * k)
	r.fb[i+3] = 255
}

func toByte(v float64) uint8 {
	if v >= 255 {
		return 255
	}
	if v <= 0 {
		return 0
	}
	return uint8(v)
}

func (r *Renderer) drawFloorAndCeiling(g *sim.Game, cam raycast.Camera) {
	// Floor casting: each screen row below the horizon is a line across the
	// deck at one distance, so the world position steps evenly along it. The
	// walls are already down, so only the pixels they left are filled.
	w, h := r.cfg.Width, r.cfg.Height
	leftX, leftY := cam.RayDir(0, w)
	rightX, rightY := cam.RayDir(w, w)
	l := g.World.Level
	for y := range h {
		p := float64(y) - float64(h)/2 + 0.5
		tex := r.tex.floor
		if p < 0 {
			p, tex = -p, r.tex.ceiling
		}
		dist := 0.5 * float64(h) / p
		stepX := dist * (rightX - leftX) / float64(w)
		stepY := dist * (rightY - leftY) / float64(w)
		fx, fy := cam.Pos.X+dist*leftX, cam.Pos.Y+dist*leftY
		for x := range w {
			if y >= r.wallTop[x] && y < r.wallEnd[x] {
				fx += stepX
				fy += stepY
				continue
			}
			cx, cy := int(math.Floor(fx)), int(math.Floor(fy))
			t := tex
			if l.At(cx, cy) == world.TileVent {
				t = r.tex.vent
				if tex == r.tex.ceiling {
					t = r.tex.ventCeiling
				}
			}
			c := t.tile(int((fx-float64(cx))*texSize), int((fy-float64(cy))*texSize))
			r.putShaded(x, y, c, r.lightAt(cx, cy), dist)
			fx += stepX
			fy += stepY
		}
	}
}

func (r *Renderer) grade(now float64) {
	// The final pass: corners fall away into shadow and, when grain is on, a
	// little sensor noise crawls over everything so even a still frame never
	// sits quite still.
	seed := uint32(int(now*grainRate)) * 0x9e3779b1
	for p, v := range r.vignette {
		// The vignette is 8.8 fixed point, so the whole pass stays in integers.
		k := int(v)
		n := 0
		if r.cfg.Grain {
			n = int(hash(uint32(p), seed)%(2*grainDepth+1)) - grainDepth
		}
		px := r.fb[p*4 : p*4+3 : p*4+3]
		px[0] = clampByte(int(px[0])*k>>8 + n)
		px[1] = clampByte(int(px[1])*k>>8 + n)
		px[2] = clampByte(int(px[2])*k>>8 + n)
	}
}

func hash(a, b uint32) uint32 {
	h := a*0x85ebca6b ^ b
	h ^= h >> 15
	h *= 0xc2b2ae35
	h ^= h >> 13
	return h
}
