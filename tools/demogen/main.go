package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"os"

	"github.com/danielriddell21/nemesis/internal/gui"
	"github.com/danielriddell21/nemesis/internal/sim"
	"github.com/danielriddell21/nemesis/internal/telemetry"
	"github.com/danielriddell21/nemesis/internal/world"
)

const (
	demoSeed     = 42
	demoWidth    = 48
	demoHeight   = 32
	demoConsoles = 3

	simSeconds   = 150.0
	captureEvery = 20 // sim ticks per GIF frame
	frameDelay   = 12 // hundredths of a second: ~2.8x speed playback
	tickDT       = 1.0 / 60
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
	if err := record("docs/demos/hunter.gif"); err != nil {
		return err
	}
	return stills()
}

type session struct {
	game    *sim.Game
	vis     *gui.Visualiser
	pilot   pilot
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
	s.game = sim.New(l, sim.WithObserver(telemetry.NewBus(s)), sim.WithLearned(s.carried))
	s.vis.Apply(gui.Msg{
		Type: "hello",
		Seed: l.Seed, Width: demoWidth, Height: demoHeight, Consoles: demoConsoles,
	})
	return nil
}

func (s *session) OnEvent(e telemetry.Event) {
	s.vis.Apply(gui.Msg{Type: "events", Events: []telemetry.Event{e}})
}

func (s *session) tick() error {
	if s.game.Dead() || s.game.Escaped() {
		// Same rhythm as the real game: the next station gets the same,
		// smarter hunter.
		s.carried = s.game.Learned()
		s.run++
		return s.start()
	}
	s.game.Tick(s.pilot.input(s.game, tickDT), tickDT)
	s.vis.TickModel(tickDT)
	return nil
}

func record(path string) error {
	s := &session{vis: gui.NewOffscreenVisualiser()}
	if err := s.start(); err != nil {
		return err
	}
	anim := &gif.GIF{}
	var prev *image.Paletted
	for i := range int(simSeconds / tickDT) {
		if err := s.tick(); err != nil {
			return err
		}
		if i%captureEvery != 0 {
			continue
		}
		state := gui.Snapshot(s.game)
		s.vis.Apply(gui.Msg{Type: "state", State: &state})
		fb, w, h := s.vis.RenderFrame()
		prev = appendFrame(anim, prev, fb, w, h)
	}
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, anim); err != nil {
		return fmt.Errorf("encode gif: %w", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		return fmt.Errorf("write gif: %w", err)
	}
	fmt.Printf("%s: %d frames\n", path, len(anim.Image))
	return nil
}

func appendFrame(anim *gif.GIF, prev *image.Paletted, fb []byte, w, h int) *image.Paletted {
	src := &image.RGBA{Pix: fb, Stride: w * 4, Rect: image.Rect(0, 0, w, h)}
	full := image.NewPaletted(src.Rect, palette.Plan9)
	draw.Draw(full, src.Rect, src, image.Point{}, draw.Src)

	frame := full
	if prev != nil {
		// Encode only the rectangle that changed since the previous frame:
		// the map area moves, the panel text mostly stands still.
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
