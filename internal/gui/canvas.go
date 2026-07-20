package gui

import (
	"image"
	"image/color"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const glyphWidth = 7

// canvas is a small software drawing surface for the menus and overlays, so the
// front-end can compose text screens the same way the renderer composes frames
// — as raw pixels handed to ebiten via WritePixels.
type canvas struct {
	img  *image.RGBA
	w, h int
}

func newCanvas(w, h int) *canvas {
	return &canvas{img: image.NewRGBA(image.Rect(0, 0, w, h)), w: w, h: h}
}

func (c *canvas) pixels() []byte { return c.img.Pix }

func (c *canvas) fill(col color.RGBA) {
	for i := 0; i < len(c.img.Pix); i += 4 {
		c.img.Pix[i], c.img.Pix[i+1], c.img.Pix[i+2], c.img.Pix[i+3] = col.R, col.G, col.B, 255
	}
}

// dimFrom copies another framebuffer scaled toward black, so a paused game
// shows faintly behind its menu.
func (c *canvas) dimFrom(fb []byte, k float64) {
	n := min(len(fb), len(c.img.Pix))
	for i := 0; i < n; i += 4 {
		c.img.Pix[i] = uint8(float64(fb[i]) * k)
		c.img.Pix[i+1] = uint8(float64(fb[i+1]) * k)
		c.img.Pix[i+2] = uint8(float64(fb[i+2]) * k)
		c.img.Pix[i+3] = 255
	}
}

func (c *canvas) rect(x, y, w, h int, col color.RGBA) {
	for dy := range h {
		py := y + dy
		if py < 0 || py >= c.h {
			continue
		}
		for dx := range w {
			px := x + dx
			if px < 0 || px >= c.w {
				continue
			}
			i := (py*c.w + px) * 4
			c.img.Pix[i], c.img.Pix[i+1], c.img.Pix[i+2], c.img.Pix[i+3] = col.R, col.G, col.B, 255
		}
	}
}

func (c *canvas) text(x, y int, s string, col color.RGBA) {
	d := &font.Drawer{
		Dst:  c.img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  fixed.P(x, y),
	}
	d.DrawString(s)
}

func (c *canvas) textCentered(y int, s string, col color.RGBA) {
	x := (c.w - len(s)*glyphWidth) / 2
	c.text(x+1, y+1, s, color.RGBA{A: 255}) // drop shadow
	c.text(x, y, s, col)
}
