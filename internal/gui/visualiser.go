package gui

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"github.com/danielriddell21/nemesis/internal/world"
)

var visPalette = struct {
	background color.RGBA
	wall       color.RGBA
	floor      color.RGBA
	vent       color.RGBA
	door       color.RGBA
	console    color.RGBA
	consoleOn  color.RGBA
	exitSealed color.RGBA
	exitOpen   color.RGBA
	player     color.RGBA
	path       color.RGBA
	target     color.RGBA
	text       color.RGBA
	dim        color.RGBA
	ripple     color.RGBA
	states     map[string]color.RGBA
}{
	background: color.RGBA{R: 12, G: 14, B: 18, A: 255},
	wall:       color.RGBA{R: 38, G: 43, B: 51, A: 255},
	floor:      color.RGBA{R: 64, G: 70, B: 80, A: 255},
	vent:       color.RGBA{R: 96, G: 82, B: 60, A: 255},
	door:       color.RGBA{R: 168, G: 132, B: 72, A: 255},
	console:    color.RGBA{R: 90, G: 120, B: 128, A: 255},
	consoleOn:  color.RGBA{R: 90, G: 210, B: 130, A: 255},
	exitSealed: color.RGBA{R: 190, G: 70, B: 60, A: 255},
	exitOpen:   color.RGBA{R: 90, G: 230, B: 130, A: 255},
	player:     color.RGBA{R: 120, G: 230, B: 160, A: 255},
	path:       color.RGBA{R: 130, G: 50, B: 50, A: 255},
	target:     color.RGBA{R: 240, G: 120, B: 90, A: 255},
	text:       color.RGBA{R: 205, G: 220, B: 212, A: 255},
	dim:        color.RGBA{R: 110, G: 125, B: 118, A: 255},
	ripple:     color.RGBA{R: 200, G: 200, B: 120, A: 255},
	states: map[string]color.RGBA{
		"lurk":        {R: 110, G: 110, B: 120, A: 255},
		"patrol":      {R: 90, G: 140, B: 220, A: 255},
		"investigate": {R: 230, G: 180, B: 70, A: 255},
		"hunt":        {R: 235, G: 70, B: 55, A: 255},
		"search":      {R: 180, G: 110, B: 220, A: 255},
	},
}

type Visualiser struct {
	link *Link
	m    *visModel
	fb   []byte
	gone bool
}

var _ ebiten.Game = (*Visualiser)(nil)

func newVisualiser(cfg Config) *Visualiser {
	return &Visualiser{
		link: cfg.Link,
		m:    newVisModel(),
		fb:   make([]byte, visWindowW*visWindowH*4),
	}
}

func (v *Visualiser) Update() error {
	if v.gone || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	if v.link != nil {
		for {
			select {
			case msg, ok := <-v.link.In:
				if !ok {
					// The game window is gone: close this one too.
					v.gone = true
					return nil
				}
				v.m.apply(msg)
			default:
				v.m.tick(tickDT)
				return nil
			}
		}
	}
	v.m.tick(tickDT)
	return nil
}

func (v *Visualiser) Draw(screen *ebiten.Image) {
	v.render()
	screen.WritePixels(v.fb)
}

func (v *Visualiser) Layout(_, _ int) (int, int) {
	return visWindowW, visWindowH
}

func (v *Visualiser) render() {
	v.fill(visPalette.background)
	if v.m.level == nil {
		v.text(20, 24, "WAITING FOR THE GAME WINDOW...", visPalette.dim)
		return
	}
	ts, offX, offY := v.mapTransform()
	v.renderTiles(ts, offX, offY)
	v.renderHeat(ts, offX, offY)
	v.renderPath(ts, offX, offY)
	v.renderRipples(ts, offX, offY)
	v.renderActors(ts, offX, offY)
	v.renderPanel()
}

func (v *Visualiser) mapTransform() (ts, offX, offY int) {
	l := v.m.level
	availW := visWindowW - visPanelW
	ts = min(availW/l.W, visWindowH/l.H)
	if ts < 1 {
		ts = 1
	}
	offX = (availW - l.W*ts) / 2
	offY = (visWindowH - l.H*ts) / 2
	return ts, offX, offY
}

func (v *Visualiser) renderTiles(ts, offX, offY int) {
	l := v.m.level
	for y := range l.H {
		for x := range l.W {
			v.rect(offX+x*ts, offY+y*ts, ts, ts, v.tileColor(l, x, y))
		}
	}
}

func (v *Visualiser) tileColor(l *world.Level, x, y int) color.RGBA {
	switch l.At(x, y) {
	case world.TileWall:
		return visPalette.wall
	case world.TileVent:
		return visPalette.vent
	case world.TileDoor:
		return visPalette.door
	case world.TileConsole:
		if v.m.activated[world.Coord{X: x, Y: y}] {
			return visPalette.consoleOn
		}
		return visPalette.console
	case world.TileExit:
		if v.m.state.Unlocked {
			return visPalette.exitOpen
		}
		return visPalette.exitSealed
	default:
		c := visPalette.floor
		k := 0.5 + 0.5*l.LightAt(x, y)
		return color.RGBA{
			R: uint8(float64(c.R) * k),
			G: uint8(float64(c.G) * k),
			B: uint8(float64(c.B) * k),
			A: 255,
		}
	}
}

func (v *Visualiser) renderHeat(ts, offX, offY int) {
	// Hot rooms are where the hunter keeps detecting prey: its learned habits.
	l := v.m.level
	for _, h := range v.m.state.Hot {
		if h[0] < 0 || h[0] >= len(l.Rooms) {
			continue
		}
		r := l.Rooms[h[0]]
		alpha := min(float64(h[1])/100*0.12, 0.45)
		for y := r.Y; y < r.Y+r.H; y++ {
			for x := r.X; x < r.X+r.W; x++ {
				v.blendRect(offX+x*ts, offY+y*ts, ts, ts, visPalette.exitSealed, alpha)
			}
		}
	}
}

func (v *Visualiser) blendRect(x, y, w, h int, c color.RGBA, alpha float64) {
	for dy := range h {
		for dx := range w {
			px, py := x+dx, y+dy
			if px < 0 || py < 0 || px >= visWindowW || py >= visWindowH {
				continue
			}
			i := (py*visWindowW + px) * 4
			v.fb[i] = uint8(float64(v.fb[i])*(1-alpha) + float64(c.R)*alpha)
			v.fb[i+1] = uint8(float64(v.fb[i+1])*(1-alpha) + float64(c.G)*alpha)
			v.fb[i+2] = uint8(float64(v.fb[i+2])*(1-alpha) + float64(c.B)*alpha)
		}
	}
}

func (v *Visualiser) renderPath(ts, offX, offY int) {
	for _, c := range v.m.state.Path {
		v.rect(offX+c[0]*ts+ts/4, offY+c[1]*ts+ts/4, ts/2, ts/2, visPalette.path)
	}
	if v.m.haveState {
		tx, ty := v.m.state.TargetX, v.m.state.TargetY
		v.cross(offX+tx*ts+ts/2, offY+ty*ts+ts/2, ts/2+1, visPalette.target)
	}
}

func (v *Visualiser) renderRipples(ts, offX, offY int) {
	for _, r := range v.m.ripples {
		progress := r.age / rippleLife
		radius := r.radius * progress * float64(ts)
		fade := 1 - progress
		c := visPalette.ripple
		c.R = uint8(float64(c.R) * fade)
		c.G = uint8(float64(c.G) * fade)
		c.B = uint8(float64(c.B) * fade)
		v.circle(offX+int(r.x*float64(ts)), offY+int(r.y*float64(ts)), int(radius), c)
	}
}

func (v *Visualiser) renderActors(ts, offX, offY int) {
	if !v.m.haveState {
		return
	}
	s := v.m.state
	ax, ay := offX+int(s.AlienX*float64(ts)), offY+int(s.AlienY*float64(ts))
	v.renderVisionCone(s, ax, ay, ts)
	if s.DecoyActive {
		v.blob(offX+int(s.DecoyX*float64(ts)), offY+int(s.DecoyY*float64(ts)), ts/3, visPalette.ripple)
	}
	// The hunter, tinted by its state.
	sc, ok := visPalette.states[s.AlienState]
	if !ok {
		sc = visPalette.dim
	}
	v.blob(ax, ay, ts/2+2, sc)
	// The player, with a facing tick — dimmed while hidden in a locker.
	pc := visPalette.player
	if s.Hidden {
		pc = visPalette.dim
	}
	px, py := offX+int(s.PlayerX*float64(ts)), offY+int(s.PlayerY*float64(ts))
	v.blob(px, py, ts/3+1, pc)
	fx := px + int(math.Cos(s.PlayerA)*float64(ts))
	fy := py + int(math.Sin(s.PlayerA)*float64(ts))
	v.line(px, py, fx, fy, pc)
}

func (v *Visualiser) renderVisionCone(s StateMsg, ax, ay, ts int) {
	if s.Vision <= 0 {
		return
	}
	sc, ok := visPalette.states[s.AlienState]
	if !ok {
		sc = visPalette.dim
	}
	edge := color.RGBA{R: sc.R / 2, G: sc.G / 2, B: sc.B / 2, A: 255}
	for i := 0; i <= 12; i++ {
		a := s.AlienA - s.VisionFOV/2 + s.VisionFOV*float64(i)/12
		ex := ax + int(math.Cos(a)*s.Vision*float64(ts))
		ey := ay + int(math.Sin(a)*s.Vision*float64(ts))
		v.line(ax, ay, ex, ey, edge)
	}
}

func (v *Visualiser) renderPanel() {
	x := visWindowW - visPanelW + 14
	y := 24
	s := v.m.state
	title := "HUNTER AI"
	if s.Deck > 0 {
		title = fmt.Sprintf("HUNTER AI   DECK %d", s.Deck)
	}
	v.text(x, y, title, visPalette.text)
	y += 20
	sc, ok := visPalette.states[s.AlienState]
	if !ok {
		sc = visPalette.dim
	}
	state := s.AlienState
	if state == "" {
		state = "offline"
	}
	if s.Hidden {
		state += "  (prey hidden)"
	}
	v.text(x, y, "STATE  "+state, sc)
	y += 16
	v.text(x, y, fmt.Sprintf("TARGET %d %d", s.TargetX, s.TargetY), visPalette.dim)
	y += 16
	v.text(x, y, fmt.Sprintf("MENACE %3.0f%%", s.Menace*100), visPalette.dim)
	v.rect(x+90, y-8, int(s.Menace*140), 8, visPalette.exitSealed)
	y += 16
	airlock := "AIRLOCK SEALED"
	if s.Unlocked {
		airlock = "AIRLOCK OPEN"
	}
	v.text(x, y, fmt.Sprintf("SYSTEMS %d/%d  %s", s.Done, s.Total, airlock), visPalette.dim)
	y += 16
	v.text(x, y, "LEARNED", visPalette.text)
	y += 14
	v.text(x, y, fmt.Sprintf(" TRACKER %d  VENTS %d  SEARCH %d", s.PingTier, s.VentTier, s.SearchTier), visPalette.dim)
	y += 14
	v.text(x, y, fmt.Sprintf(" DECOYS %d  LOCKERS %d", s.DecoyTier, s.LockerTier), visPalette.dim)
	y += 22
	v.text(x, y, "TRIGGERS", visPalette.text)
	y += 16
	for _, line := range v.m.feed {
		v.text(x, y, line, visPalette.dim)
		y += 13
	}
}

func (v *Visualiser) fill(c color.RGBA) {
	for i := 0; i < len(v.fb); i += 4 {
		v.fb[i], v.fb[i+1], v.fb[i+2], v.fb[i+3] = c.R, c.G, c.B, 255
	}
}

func (v *Visualiser) rect(x, y, w, h int, c color.RGBA) {
	for dy := range h {
		for dx := range w {
			v.pixel(x+dx, y+dy, c)
		}
	}
}

func (v *Visualiser) blob(cx, cy, radius int, c color.RGBA) {
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy <= radius*radius {
				v.pixel(cx+dx, cy+dy, c)
			}
		}
	}
}

func (v *Visualiser) circle(cx, cy, radius int, c color.RGBA) {
	if radius < 1 {
		return
	}
	steps := 8 * radius
	for i := range steps {
		a := float64(i) / float64(steps) * 2 * math.Pi
		v.pixel(cx+int(math.Cos(a)*float64(radius)), cy+int(math.Sin(a)*float64(radius)), c)
	}
}

func (v *Visualiser) cross(cx, cy, arm int, c color.RGBA) {
	for d := -arm; d <= arm; d++ {
		v.pixel(cx+d, cy+d, c)
		v.pixel(cx+d, cy-d, c)
	}
}

func (v *Visualiser) line(x0, y0, x1, y1 int, c color.RGBA) {
	steps := max(absDelta(x1, x0), absDelta(y1, y0))
	if steps == 0 {
		v.pixel(x0, y0, c)
		return
	}
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		v.pixel(x0+int(t*float64(x1-x0)), y0+int(t*float64(y1-y0)), c)
	}
}

func absDelta(a, b int) int {
	if a > b {
		return a - b
	}
	return b - a
}

func (v *Visualiser) pixel(x, y int, c color.RGBA) {
	if x < 0 || y < 0 || x >= visWindowW || y >= visWindowH {
		return
	}
	i := (y*visWindowW + x) * 4
	v.fb[i], v.fb[i+1], v.fb[i+2], v.fb[i+3] = c.R, c.G, c.B, 255
}

func (v *Visualiser) text(x, y int, s string, c color.RGBA) {
	d := &font.Drawer{
		Dst: &image.RGBA{
			Pix:    v.fb,
			Stride: visWindowW * 4,
			Rect:   image.Rect(0, 0, visWindowW, visWindowH),
		},
		Src:  image.NewUniform(c),
		Face: basicfont.Face7x13,
		Dot:  fixed.P(x, y),
	}
	d.DrawString(s)
}
