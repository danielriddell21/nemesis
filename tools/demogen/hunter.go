package main

import (
	"fmt"
	"image"

	"github.com/danielriddell21/crucible/demo"
	"github.com/danielriddell21/crucible/record"

	"github.com/danielriddell21/nemesis/internal/pilot"
	"github.com/danielriddell21/nemesis/internal/telemetry"
	"github.com/danielriddell21/nemesis/internal/vis"
)

const (
	hunterEvery  = 20  // sim ticks per captured frame
	hunterDelay  = 12  // hundredths of a second per frame: ~2.8x speed playback
	hunterFrames = 140 // long enough to show a hunt turn into a search
)

// hunterClip watches a bot-driven run through the AI visualiser. It collects
// each tick's events so they can be forwarded to the renderer, the same way the
// live visualiser window receives them over the link.
type hunterClip struct {
	r      *vis.Renderer
	events []telemetry.Event
}

// OnEvent buffers one event for the current tick.
func (h *hunterClip) OnEvent(e telemetry.Event) { h.events = append(h.events, e) }

// recordHunter records the visualiser watching a bot play. It drives the same
// renderer the visualiser window draws with, so the clip shows what a player
// running `nemesis --visualiser` would see.
func recordHunter(path string) error {
	h := &hunterClip{r: vis.New()}
	s := &session{
		pilot:   pilot.New(),
		observe: h,
		onStart: func(seed int64) {
			h.r.Apply(vis.Msg{
				Type: "hello", Seed: seed,
				Width: demoWidth, Height: demoHeight, Consoles: demoConsoles,
			})
		},
	}
	if err := s.start(); err != nil {
		return err
	}

	w, hgt := vis.Size()
	// Frames are already the visualiser's own size, so the recorder keeps
	// scale 1; the map is mostly still between frames, which delta encoding
	// exploits.
	rec := record.NewRecorder(0, 1, hunterFrames, record.WithFrameDelay(hunterDelay), record.WithFrameDiff())

	clip := demo.Clip{
		Frames: hunterFrames,
		Every:  hunterEvery,
		Step: func(int) error {
			h.events = h.events[:0]
			if err := s.tick(); err != nil {
				return err
			}
			state := vis.Snapshot(s.game)
			h.r.Apply(vis.Msg{Type: "state", State: &state})
			if len(h.events) > 0 {
				h.r.Apply(vis.Msg{Type: "events", Events: h.events})
			}
			h.r.Tick(tickDT)
			return nil
		},
		Frame: func(int) image.Image { return record.FromRGBA(h.r.Frame(), w, hgt) },
	}
	if _, err := clip.Record(rec); err != nil {
		return fmt.Errorf("capture clip: %w", err)
	}
	if err := rec.Save(path); err != nil {
		return fmt.Errorf("save gif: %w", err)
	}
	fmt.Printf("%s: %d frames\n", path, rec.Len())
	return nil
}
