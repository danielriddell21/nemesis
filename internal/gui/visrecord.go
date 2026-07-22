package gui

import (
	"fmt"
	"image"
	"os"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/danielriddell21/crucible/record"

	"github.com/danielriddell21/nemesis/internal/pilot"
	"github.com/danielriddell21/nemesis/internal/sim"
	"github.com/danielriddell21/nemesis/internal/telemetry"
	"github.com/danielriddell21/nemesis/internal/world"
)

const (
	recCaptureEvery  = 20 // sim ticks per captured frame
	recFrameDelay    = 12 // hundredths of a second: ~2.8x speed playback
	recDefaultFrames = 150
)

// recVisualiser drives a bot through fresh stations and records the AI
// visualiser to a GIF, then exits — the "run the real window and capture it"
// approach the family's other GUIs use for their demos.
type recVisualiser struct {
	cfg     Config
	v       *Visualiser
	game    *sim.Game
	pilot   *pilot.Pilot
	carried sim.Learned
	run     int
	events  []telemetry.Event
	rec     *record.Recorder
	done    bool
}

// RunRecord opens a hidden window that plays itself and records the visualiser
// to cfg.Rec.Path, capturing cfg.Rec.Frames frames before exiting.
func RunRecord(cfg Config) error {
	frames := cfg.Rec.Frames
	if frames <= 0 {
		frames = recDefaultFrames
	}
	rv := &recVisualiser{
		cfg:   cfg,
		v:     newVisualiser(cfg),
		pilot: pilot.New(),
		rec:   record.NewRecorder(0, 1, frames, record.WithFrameDelay(recFrameDelay), record.WithFrameDiff()),
	}
	rv.startRun()
	ebiten.SetWindowTitle("nemesis — recording")
	ebiten.SetWindowSize(visWindowW, visWindowH)
	if err := ebiten.RunGame(rv); err != nil {
		return fmt.Errorf("record visualiser: %w", err)
	}
	return nil
}

// OnEvent collects the tick's events so they can be forwarded to the model.
func (r *recVisualiser) OnEvent(e telemetry.Event) { r.events = append(r.events, e) }

func (r *recVisualiser) startRun() {
	seed := r.cfg.Seed + int64(r.run)*0x9e3779b9
	l, err := world.Generate(world.Config{
		Width: r.cfg.Width, Height: r.cfg.Height, Seed: seed, Consoles: r.cfg.Consoles,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "record: generate:", err)
		r.done = true
		return
	}
	r.game = sim.New(l,
		sim.WithObserver(telemetry.NewBus(r)),
		sim.WithLearned(r.carried),
		sim.WithDepth(r.run),
	)
	r.v.m.apply(Msg{Type: "hello", Seed: seed, Width: r.cfg.Width, Height: r.cfg.Height, Consoles: r.cfg.Consoles})
}

func (r *recVisualiser) Update() error {
	if r.done {
		return ebiten.Termination
	}
	for range recCaptureEvery {
		if r.game.Dead() || r.game.Escaped() {
			// Carry the hunter's learning into the next station, like the game.
			r.carried = r.game.Learned()
			r.run++
			r.startRun()
			continue
		}
		r.events = r.events[:0]
		r.game.Tick(r.pilot.Input(r.game, tickDT), tickDT)
		state := Snapshot(r.game)
		r.v.m.apply(Msg{Type: "state", State: &state})
		if len(r.events) > 0 {
			r.v.m.apply(Msg{Type: "events", Events: r.events})
		}
		r.v.m.tick(tickDT)
	}
	return nil
}

func (r *recVisualiser) Draw(screen *ebiten.Image) {
	r.v.render()
	screen.WritePixels(r.v.fb)
	if !r.rec.Done() {
		r.rec.Add(&image.RGBA{Pix: r.v.fb, Stride: visWindowW * 4, Rect: image.Rect(0, 0, visWindowW, visWindowH)})
		return
	}
	if !r.done {
		r.done = true
		if err := r.rec.Save(r.cfg.Rec.Path); err != nil {
			fmt.Fprintln(os.Stderr, "record: save:", err)
			return
		}
		fmt.Printf("%s: %d frames\n", r.cfg.Rec.Path, r.rec.Len())
	}
}

func (r *recVisualiser) Layout(_, _ int) (int, int) { return visWindowW, visWindowH }
