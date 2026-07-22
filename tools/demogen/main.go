// Command demogen renders nemesis's documentation media headlessly: a
// first-person gameplay clip, a station montage, and feature stills. Like
// pandemonium's generator it drives the simulation and the software renderer
// directly and never imports the GUI, so it needs no display. The AI
// visualiser demo is recorded separately from the real window (see the
// justfile's `demos` recipe).
package main

import (
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"os"

	"github.com/danielriddell21/nemesis/internal/pilot"
	"github.com/danielriddell21/nemesis/internal/sim"
	"github.com/danielriddell21/nemesis/internal/world"
)

const (
	demoSeed     = 42
	demoWidth    = 48
	demoHeight   = 32
	demoConsoles = 3

	tickDT     = 1.0 / 60
	frameDelay = 12 // hundredths of a second, unless a clip overrides it
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "demogen:", err)
		os.Exit(1)
	}
}

func run() error {
	if err := os.MkdirAll("docs/demos", 0o750); err != nil {
		return fmt.Errorf("demos dir: %w", err)
	}
	if err := recordCorridors("docs/demos/corridors.gif"); err != nil {
		return err
	}
	if err := stationsMontage("docs/demos/stations.png"); err != nil {
		return err
	}
	return stills()
}

// session drives one continuing run: when the pilot dies or escapes, the next
// station carries the hunter's learning forward, exactly like the real game.
type session struct {
	game    *sim.Game
	pilot   *pilot.Pilot
	carried sim.Learned
	run     int
}

func (s *session) start() error {
	l, err := world.Generate(world.Config{
		Width: demoWidth, Height: demoHeight,
		Seed: demoSeed + int64(s.run)*0x9e3779b9, Consoles: demoConsoles,
	})
	if err != nil {
		return fmt.Errorf("generate: %w", err)
	}
	s.game = sim.New(l, sim.WithLearned(s.carried), sim.WithDepth(s.run))
	return nil
}

func (s *session) tick() error {
	if s.game.Dead() || s.game.Escaped() {
		s.carried = s.game.Learned()
		s.run++
		return s.start()
	}
	s.game.Tick(s.pilot.Input(s.game, tickDT), tickDT)
	return nil
}

func appendFrame(anim *gif.GIF, prev *image.Paletted, fb []byte, w, h int) *image.Paletted {
	src := &image.RGBA{Pix: fb, Stride: w * 4, Rect: image.Rect(0, 0, w, h)}
	full := image.NewPaletted(src.Rect, palette.Plan9)
	draw.Draw(full, src.Rect, src, image.Point{}, draw.Src)

	frame := full
	if prev != nil {
		// Encode only the rectangle that changed since the previous frame.
		box, changed := diffBox(prev, full)
		if !changed {
			box = image.Rect(0, 0, 1, 1)
		}
		sub := image.NewPaletted(box, palette.Plan9)
		draw.Draw(sub, box, full, box.Min, draw.Src)
		frame = sub
	}
	anim.Image = append(anim.Image, frame)
	anim.Delay = append(anim.Delay, frameDelay)
	anim.Disposal = append(anim.Disposal, gif.DisposalNone)
	return full
}

func diffBox(a, b *image.Paletted) (image.Rectangle, bool) {
	minX, minY := b.Rect.Max.X, b.Rect.Max.Y
	maxX, maxY := -1, -1
	for y := range b.Rect.Max.Y {
		rowA := a.Pix[y*a.Stride : y*a.Stride+b.Rect.Max.X]
		rowB := b.Pix[y*b.Stride : y*b.Stride+b.Rect.Max.X]
		for x := range b.Rect.Max.X {
			if rowA[x] == rowB[x] {
				continue
			}
			if x < minX {
				minX = x
			}
			if x > maxX {
				maxX = x
			}
			if y < minY {
				minY = y
			}
			maxY = y
		}
	}
	if maxX < 0 {
		return image.Rectangle{}, false
	}
	return image.Rect(minX, minY, maxX+1, maxY+1), true
}
