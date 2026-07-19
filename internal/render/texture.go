package render

import "image/color"

const texSize = 64

type texture struct {
	w, h int
	pix  []color.RGBA
}

func newTexture(w int) *texture {
	return &texture{w: w, h: texSize, pix: make([]color.RGBA, w*texSize)}
}

func (t *texture) set(x, y int, c color.RGBA) { t.pix[y*t.w+x] = c }

func (t *texture) at(u, v int) color.RGBA {
	if u < 0 {
		u = 0
	}
	if v < 0 {
		v = 0
	}
	return t.pix[(v%t.h)*t.w+u%t.w]
}

type textureSet struct {
	wall          *texture
	door          *texture
	vent          *texture
	console       *texture
	consoleActive *texture
	alien         *texture
}

func buildTextures() *textureSet {
	return &textureSet{
		wall:          wallTexture(),
		door:          doorTexture(),
		vent:          ventTexture(),
		console:       consoleTexture(palette.console, false),
		consoleActive: consoleTexture(palette.consoleActive, true),
		alien:         alienTexture(),
	}
}

func wallTexture() *texture {
	t := newTexture(texSize)
	for y := range texSize {
		for x := range texSize {
			c := palette.wall
			// Station plating: panel seams every 16 texels with darker grout
			// and a subtle per-panel tone shift.
			panel := (x/16 + y/16) % 2
			if panel == 1 {
				c = adjust(c, -8)
			}
			if x%16 == 0 || y%16 == 0 {
				c = adjust(c, -30)
			}
			if (x*7+y*13)%31 == 0 {
				c = adjust(c, -12)
			}
			t.set(x, y, c)
		}
	}
	return t
}

func doorTexture() *texture {
	t := newTexture(texSize)
	for y := range texSize {
		for x := range texSize {
			c := palette.door
			// Horizontal slats with a warning chevron band across the middle.
			if y%8 < 2 {
				c = adjust(c, -25)
			}
			if y > 28 && y < 36 && (x+y)%12 < 6 {
				c = adjust(c, 40)
			}
			t.set(x, y, c)
		}
	}
	return t
}

func ventTexture() *texture {
	t := newTexture(texSize)
	for y := range texSize {
		for x := range texSize {
			c := palette.vent
			// Ribbed ducting with grate slits.
			if y%6 < 2 {
				c = adjust(c, -20)
			}
			if x%10 < 2 {
				c = adjust(c, -14)
			}
			t.set(x, y, c)
		}
	}
	return t
}

func consoleTexture(base color.RGBA, active bool) *texture {
	t := newTexture(texSize)
	for y := range texSize {
		for x := range texSize {
			t.set(x, y, consoleTexel(base, active, x, y))
		}
	}
	return t
}

func consoleTexel(base color.RGBA, active bool, x, y int) color.RGBA {
	// A dark cabinet with a screen that glows once activated, over a strip of
	// switchgear.
	if x > 10 && x < 54 && y > 12 && y < 44 {
		if active && y%6 < 3 {
			return adjust(base, 35)
		}
		if !active && (x+y*2)%17 == 0 {
			return adjust(base, 25)
		}
		return base
	}
	if x > 14 && x < 50 && y > 48 && y < 56 && x%8 < 5 {
		return adjust(palette.wall, 12)
	}
	return adjust(palette.wall, -20)
}

func alienTexture() *texture {
	const w, h = 32, texSize
	t := newTexture(w)
	transparent := color.RGBA{}
	for y := range h {
		for x := range w {
			t.set(x, y, transparent)
		}
	}
	body := palette.alien
	sheen := adjust(body, 24)
	// A tall silhouette: elongated head, hunched torso, whip tail. Cheap
	// mirrored half-profile so the shape stays symmetric.
	for y := range h {
		half := silhouetteHalfWidth(y)
		for dx := -half; dx <= half; dx++ {
			x := w/2 + dx
			if x < 0 || x >= w {
				continue
			}
			c := body
			if dx == half || dx == -half {
				c = sheen
			}
			t.set(x, y, c)
		}
	}
	return t
}

func silhouetteHalfWidth(y int) int {
	switch {
	case y < 10: // crested head
		return 3 + y/3
	case y < 18: // neck
		return 3
	case y < 40: // torso and arms
		return 8 - (y-18)/8
	case y < 56: // legs
		return 4
	default: // tail sweep
		return 6
	}
}

func adjust(c color.RGBA, d int) color.RGBA {
	return color.RGBA{R: clampByte(int(c.R) + d), G: clampByte(int(c.G) + d), B: clampByte(int(c.B) + d), A: 255}
}

func clampByte(v int) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}
