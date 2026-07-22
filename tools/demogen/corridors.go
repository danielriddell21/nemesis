package main

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"math"
	"os"

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
	anim := &gif.GIF{}
	var prev *image.Paletted
	now, frames, recording := 0.0, 0, false
	for i := 0; frames < povFrames && now < povDeadline; i++ {
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
		prev = appendFrame(anim, prev, small, w, h)
		anim.Delay[len(anim.Delay)-1] = povDelay
		frames++
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
