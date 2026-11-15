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

func (t *texture) tile(u, v int) color.RGBA {
	// Deck textures are square and a power of two wide, so wrapping is a mask.
	return t.pix[(v&(texSize-1))*texSize+u&(texSize-1)]
}

type textureSet struct {
	wall          *texture
	floor         *texture
	ceiling       *texture
	door          *texture
	vent          *texture
	duct          *texture
	ventCeiling   *texture
	console       *texture
	consoleActive *texture
	alien         [2]*texture
	locker        *texture
	decoy         *texture
	battery       [2]*texture
}

func buildTextures() *textureSet {
	return &textureSet{
		wall:          wallTexture(),
		floor:         floorTexture(),
		ceiling:       ceilingTexture(),
		door:          doorTexture(),
		vent:          ventTexture(),
		duct:          ductTexture(),
		ventCeiling:   ventCeilingTexture(),
		console:       consoleTexture(palette.console, false),
		consoleActive: consoleTexture(palette.consoleActive, true),
		alien:         [2]*texture{alienTexture(-1), alienTexture(1)},
		locker:        lockerTexture(),
		decoy:         decoyTexture(),
		battery:       [2]*texture{batteryTexture(false), batteryTexture(true)},
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
			t.set(x, y, grime(c, x, y))
		}
	}
	return t
}

func grime(c color.RGBA, x, y int) color.RGBA {
	// Years of neglect: damp creeping up from the deck, rust weeping down
	// from a few seams, and a fine mottling of dirt over everything.
	c = adjust(c, -int(hash(uint32(x), uint32(y))%14))
	if y > 44 {
		c = adjust(c, -(y-44)*2)
	}
	if streak := hash(uint32(x/2), 77) % 13; streak == 0 && y > 16 && y < 16+int(hash(uint32(x/2), 5)%40) {
		c = color.RGBA{R: clampByte(int(c.R) - 6), G: clampByte(int(c.G) - 26), B: clampByte(int(c.B) - 38), A: 255}
	}
	return c
}

func floorTexture() *texture {
	t := newTexture(texSize)
	for y := range texSize {
		for x := range texSize {
			c := palette.floor
			// Deck plating: a seam round each tile, a tread pattern across
			// it, and stains pooled at random.
			if x < 2 || y < 2 {
				c = adjust(c, -22)
			} else if (x+y)%8 == 0 {
				c = adjust(c, 10)
			}
			c = adjust(c, -int(hash(uint32(x/4), uint32(y/4))%16))
			if hash(uint32(x/8), uint32(y/8)+31)%7 == 0 {
				c = adjust(c, -18)
			}
			t.set(x, y, c)
		}
	}
	return t
}

func ceilingTexture() *texture {
	t := newTexture(texSize)
	for y := range texSize {
		for x := range texSize {
			c := palette.ceiling
			// Ceiling panels in half-tile squares with a conduit run down one
			// side.
			if x%32 < 2 || y%32 < 2 {
				c = adjust(c, -18)
			}
			if x > 44 && x < 52 {
				c = adjust(c, 14)
			}
			c = adjust(c, -int(hash(uint32(x/4)+9, uint32(y/4))%10))
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
			// A floor grating: bright bars over a black void, so the duct
			// reads as hollow under your feet.
			c := adjust(palette.vent, 14)
			switch {
			case x%16 < 3 || y%32 < 3:
				c = adjust(palette.vent, -6)
			case x%8 >= 3 && y%32 >= 8 && y%32 < 28:
				c = color.RGBA{R: 4, G: 5, B: 5, A: 255}
			}
			t.set(x, y, c)
		}
	}
	return t
}

func ductTexture() *texture {
	t := newTexture(texSize)
	for y := range texSize {
		for x := range texSize {
			// Corrugated sheet metal: horizontal ribs with a lit crest and a
			// dark trough, riveted seams every 32 texels, and dust.
			c := palette.vent
			switch y % 8 {
			case 0, 1:
				c = adjust(c, 22)
			case 4, 5:
				c = adjust(c, -22)
			}
			if x%32 < 2 {
				c = adjust(c, -28)
			}
			if x%32 == 4 && (y%16 == 4 || y%16 == 12) {
				c = adjust(c, 38)
			}
			t.set(x, y, grime(c, x, y))
		}
	}
	return t
}

func ventCeilingTexture() *texture {
	t := newTexture(texSize)
	for y := range texSize {
		for x := range texSize {
			// The duct's roof: the same sheet, seamed lengthways and darker.
			c := adjust(palette.vent, -18)
			if x%16 < 2 {
				c = adjust(c, -16)
			}
			t.set(x, y, adjust(c, -int(hash(uint32(x), uint32(y)+3)%8)))
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

func alienTexture(sway int) *texture {
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
	eye := color.RGBA{R: 196, G: 226, B: 178, A: glowAlpha}
	// A tall silhouette: elongated head, hunched torso, whip tail. The legs and
	// tail lean by `sway` so two frames give a slinking walk cycle.
	for y := range h {
		half := silhouetteHalfWidth(y)
		lean := 0
		if y >= 40 {
			lean = sway * (y - 40) / 8
		}
		for dx := -half; dx <= half; dx++ {
			x := w/2 + dx + lean
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
	// Two pale points in the skull: in the dark they are all that shows.
	for _, dx := range [2]int{-2, 2} {
		t.set(w/2+dx, 6, eye)
		t.set(w/2+dx, 7, eye)
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

func lockerTexture() *texture {
	t := newTexture(texSize)
	for y := range texSize {
		for x := range texSize {
			t.set(x, y, lockerTexel(x, y))
		}
	}
	return t
}

func lockerTexel(x, y int) color.RGBA {
	if x == 8 || x == 56 || y == 6 || y == 60 {
		return adjust(palette.wall, 16) // the frame edge
	}
	if x <= 8 || x >= 56 || y <= 6 || y >= 60 {
		return adjust(palette.wall, -28) // recessed surround
	}
	door := adjust(palette.wall, -6)
	if x > 46 && x < 50 && y > 26 && y < 40 {
		return adjust(door, 30) // handle
	}
	if y%6 < 2 {
		return adjust(door, -18) // louvred slats
	}
	return door
}

func decoyTexture() *texture {
	const w = 32
	t := newTexture(w)
	transparent := color.RGBA{}
	glow := color.RGBA{R: 90, G: 220, B: 160, A: 255}
	shell := color.RGBA{R: 66, G: 72, B: 82, A: 255}
	// A small canister resting on the floor: the device sits in the lower band
	// of the sprite so it reads as a thrown object, not a standing figure.
	for y := range texSize {
		for x := range w {
			c := transparent
			dx, dy := x-w/2, y-52
			if dy >= -10 && dy <= 8 && dx*dx*2+dy*dy <= 80 {
				c = shell
				if dx*dx+dy*dy <= 6 {
					c = glow // blinking core
				}
			}
			t.set(x, y, c)
		}
	}
	return t
}

func batteryTexture(blink bool) *texture {
	const w = 32
	t := newTexture(w)
	shell := color.RGBA{R: 74, G: 70, B: 48, A: 255}
	band := color.RGBA{R: 150, G: 132, B: 44, A: 255}
	lamp := color.RGBA{R: 120, G: 240, B: 130, A: glowAlpha}
	if !blink {
		lamp = color.RGBA{R: 40, G: 96, B: 52, A: glowAlpha}
	}
	// A squat power cell lying on the deck, with a charge lamp that blinks so
	// it can be found in the dark.
	for y := 46; y < 62; y++ {
		for x := 9; x < 23; x++ {
			c := shell
			switch {
			case y > 50 && y < 54:
				c = band
			case y < 49 && x > 13 && x < 18:
				c = lamp
			}
			t.set(x, y, c)
		}
	}
	return t
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
