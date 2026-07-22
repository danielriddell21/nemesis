package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"

	"github.com/danielriddell21/crucible/hud"

	"github.com/danielriddell21/nemesis/internal/render"
	"github.com/danielriddell21/nemesis/internal/sim"
	"github.com/danielriddell21/nemesis/internal/world"
)

func stills() error {
	l, err := world.Generate(world.Config{Width: demoWidth, Height: demoHeight, Seed: demoSeed})
	if err != nil {
		return fmt.Errorf("generate: %w", err)
	}
	g := sim.New(l)
	g.Player.Angle = longestSightline(g, l)
	d := g.Player.Dir()
	g.Alien.Pos = sim.Vec2{X: g.Player.Pos.X + d.X*7, Y: g.Player.Pos.Y + d.Y*7}
	for range 40 {
		g.Tick(sim.Input{Tracker: true}, tickDT)
	}
	r := render.NewRenderer(render.DefaultConfig(), render.WithOverlay(hud.New()))
	fb := r.Frame(g, 3.0)
	cfg := r.Config()
	return writePNG("docs/demos/game.png", fb, cfg.Width, cfg.Height)
}

func longestSightline(g *sim.Game, l *world.Level) float64 {
	best, bestD := 0.0, -1.0
	for i := range 64 {
		a := float64(i) / 64 * 2 * math.Pi
		dx, dy := math.Cos(a), math.Sin(a)
		d := 0.0
		for d < 15 {
			d += 0.25
			c := world.Coord{X: int(g.Player.Pos.X + dx*d), Y: int(g.Player.Pos.Y + dy*d)}
			if l.Solid(c.X, c.Y) {
				break
			}
		}
		if d > bestD {
			bestD, best = d, a
		}
	}
	return best
}

func writePNG(path string, fb []byte, w, h int) error {
	img := &image.RGBA{Pix: fb, Stride: w * 4, Rect: image.Rect(0, 0, w, h)}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return fmt.Errorf("encode %s: %w", path, err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	fmt.Println(path)
	return nil
}
