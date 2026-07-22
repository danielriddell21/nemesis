package main

import (
	"fmt"
	"image"
	"math"

	"github.com/danielriddell21/crucible/record"

	"github.com/danielriddell21/nemesis/internal/pilot"
	"github.com/danielriddell21/nemesis/internal/render"
)

const (
	povScale    = 2
	povEvery    = 7   // sim ticks per frame: ~8.6 fps at real speed
	povFrames   = 160 // ~18 seconds of play
	povDelay    = 12  // hundredths of a second per frame
	povLeadIn   = 13.0
	povForceAt  = 60.0
	povDeadline = 150.0
)

func recordCorridors(path string) error {
	s := &session{pilot: pilot.New()}
	if err := s.start(); err != nil {
		return err
	}
	r := render.NewRenderer(render.DefaultConfig())
	cfg := r.Config()
	// Frames are pre-downscaled with a gamma lift below, so the recorder keeps
	// scale 1 and only handles the delta-frame GIF encoding.
	rec := record.NewRecorder(0, 1, povFrames, record.WithFrameDelay(povDelay), record.WithFrameDiff())
	now, recording := 0.0, false
	for i := 0; !rec.Done() && now < povDeadline; i++ {
		if err := s.tick(); err != nil {
			return err
		}
		now += tickDT
		// Start rolling once the hunter is close enough to matter, so the clip
		// opens on the interesting part of the run.
		if !recording {
			d := s.game.Alien.Pos.Sub(s.game.Player.Pos).Len()
			recording = d < povLeadIn || now > povForceAt
		}
		if !recording || i%povEvery != 0 {
			continue
		}
		fb := r.Frame(s.game, now)
		small, w, h := downscale(fb, cfg.Width, cfg.Height, povScale)
		rec.Add(&image.RGBA{Pix: small, Stride: w * 4, Rect: image.Rect(0, 0, w, h)})
	}
	if err := rec.Save(path); err != nil {
		return fmt.Errorf("save gif: %w", err)
	}
	fmt.Printf("%s: %d frames\n", path, rec.Len())
	return nil
}

func downscale(fb []byte, w, h, k int) ([]byte, int, int) {
	// Box-average k x k blocks, then lift the gamma: the GIF palette spaces
	// its dark shades coarsely, so the station's gloom needs a boost to
	// survive quantisation.
	ow, oh := w/k, h/k
	out := make([]byte, ow*oh*4)
	lift := gammaLUT(0.62)
	for y := range oh {
		for x := range ow {
			var r, g, b int
			for dy := range k {
				for dx := range k {
					i := ((y*k+dy)*w + x*k + dx) * 4
					r += int(fb[i])
					g += int(fb[i+1])
					b += int(fb[i+2])
				}
			}
			n := k * k
			o := (y*ow + x) * 4
			out[o] = lift[r/n]
			out[o+1] = lift[g/n]
			out[o+2] = lift[b/n]
			out[o+3] = 255
		}
	}
	return out, ow, oh
}

func gammaLUT(gamma float64) [256]uint8 {
	var lut [256]uint8
	for i := range lut {
		lut[i] = uint8(255*math.Pow(float64(i)/255, gamma) + 0.5)
	}
	return lut
}
