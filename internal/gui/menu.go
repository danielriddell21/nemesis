package gui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type menuItem struct {
	label  func() string
	action func()
	adjust func(dir int) // nil for plain items; set for adjustable settings rows
}

type menu struct {
	title    string
	subtitle []string
	items    []menuItem
	sel      int
}

type menuSound int

const (
	soundNone menuSound = iota
	soundMove
	soundSelect
)

func (m *menu) moveSel(dir int) {
	m.sel = (m.sel + dir + len(m.items)) % len(m.items)
}

func (m *menu) update() menuSound {
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		m.moveSel(1)
		return soundMove
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		m.moveSel(-1)
		return soundMove
	}
	cur := m.items[m.sel]
	if cur.adjust != nil {
		if inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
			cur.adjust(-1)
			return soundMove
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
			cur.adjust(1)
			return soundMove
		}
	}
	if (inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace)) && cur.action != nil {
		cur.action()
		return soundSelect
	}
	return soundNone
}

var (
	menuBG       = color.RGBA{R: 10, G: 12, B: 16, A: 255}
	menuTitle    = color.RGBA{R: 210, G: 70, B: 60, A: 255}
	menuText     = color.RGBA{R: 200, G: 214, B: 208, A: 255}
	menuDim      = color.RGBA{R: 120, G: 134, B: 128, A: 255}
	menuSelected = color.RGBA{R: 90, G: 220, B: 160, A: 255}
)

func (m *menu) draw(c *canvas) {
	c.textCentered(70, m.title, menuTitle)
	y := 108
	for _, line := range m.subtitle {
		c.textCentered(y, line, menuDim)
		y += 15
	}
	y = max(y+16, 176)
	for i, it := range m.items {
		col := menuText
		label := it.label()
		if i == m.sel {
			col = menuSelected
			label = "> " + label + " <"
			c.rect(0, y-13, c.w, 18, color.RGBA{R: 24, G: 40, B: 34, A: 255})
		}
		c.textCentered(y, label, col)
		y += 22
	}
}
