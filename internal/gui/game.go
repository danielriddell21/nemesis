package gui

import (
	"fmt"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/danielriddell21/nemesis/internal/hud"
	"github.com/danielriddell21/nemesis/internal/render"
	"github.com/danielriddell21/nemesis/internal/sim"
	"github.com/danielriddell21/nemesis/internal/telemetry"
	"github.com/danielriddell21/nemesis/internal/world"
)

const (
	tickDT        = 1.0 / 60
	stateEveryNth = 4
)

type Game struct {
	cfg      Config
	sim      *sim.Game
	renderer *render.Renderer
	overlay  *hud.Overlay
	audio    *Audio
	records  Records
	link     *Link

	baseSeed int64
	runIndex int
	carried  sim.Learned
	events   []telemetry.Event
	now      float64
	recorded bool
	quit     bool

	haveMouse  bool
	lastMouseX int
}

var _ ebiten.Game = (*Game)(nil)

func NewGame(cfg Config, seed int64, overlay *hud.Overlay, audio *Audio, records Records) (*Game, error) {
	g := &Game{
		cfg:      cfg,
		renderer: render.NewRenderer(render.DefaultConfig(), render.WithOverlay(overlay)),
		overlay:  overlay,
		audio:    audio,
		records:  records,
		link:     cfg.Link,
		baseSeed: seed,
	}
	if err := g.startRun(); err != nil {
		return nil, err
	}
	return g, nil
}

func (g *Game) OnEvent(e telemetry.Event) {
	g.events = append(g.events, e)
}

func (g *Game) runSeed() int64 {
	// Every restart gets a fresh, reproducible seed derived from the base.
	return g.baseSeed + int64(g.runIndex)*0x9e3779b9
}

func (g *Game) startRun() error {
	level, err := world.Generate(world.Config{
		Width:    g.cfg.Width,
		Height:   g.cfg.Height,
		Seed:     g.runSeed(),
		Consoles: g.cfg.Consoles,
	})
	if err != nil {
		return fmt.Errorf("generate level: %w", err)
	}
	fmt.Printf("nemesis — run %d, seed %d\n", g.runIndex+1, g.runSeed())
	g.sim = sim.New(level, sim.WithObserver(telemetry.NewBus(g)), sim.WithLearned(g.carried))
	g.recorded = false
	g.events = g.events[:0]
	g.overlay.Post("BRING THE GENERATORS ONLINE - THEN THE AIRLOCK", 240, hud.Notice)
	g.sendHello()
	return nil
}

func (g *Game) Update() error {
	if g.quit || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
	g.drainLink()
	g.now += tickDT

	if g.sim.Dead() || g.sim.Escaped() {
		g.finishRun()
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyR) {
			// The hunter remembers: what it learned about this prey survives
			// into the next station.
			g.carried = g.sim.Learned()
			g.runIndex++
			if err := g.startRun(); err != nil {
				return err
			}
		}
		g.overlay.Tick()
		return nil
	}

	g.sim.Tick(g.readInput(), tickDT)
	g.postNotices()
	g.playCues()
	g.publish()
	g.events = g.events[:0]
	g.overlay.Tick()
	return nil
}

func (g *Game) finishRun() {
	if g.recorded {
		return
	}
	g.recorded = true
	g.records.record(g.sim.Escaped(), g.sim.Elapsed())
	if err := g.records.save(); err != nil {
		fmt.Println("records:", err)
	}
	if g.sim.Escaped() {
		g.overlay.Post("PRESS ENTER FOR THE NEXT STATION", 100000, hud.Notice)
	} else {
		g.overlay.Post("PRESS ENTER TO TRY AGAIN", 100000, hud.Notice)
	}
	g.publish()
}

func (g *Game) postNotices() {
	for _, e := range g.events {
		text, frames, ok := notice(e, g.sim)
		if ok {
			g.overlay.Post(text, frames, hud.Notice)
		}
	}
}

func (g *Game) playCues() {
	if g.audio == nil {
		return
	}
	g.audio.SetListener(g.sim.Player.Pos)
	for _, e := range g.events {
		g.audio.PlayEvent(e)
	}
	g.audio.TickAmbient(g.sim.Menace())
}

func (g *Game) publish() {
	if g.link == nil {
		return
	}
	if len(g.events) > 0 {
		trySend(g.link.Out, Msg{Type: "events", Events: append([]telemetry.Event(nil), g.events...)})
	}
	if g.sim.TickCount()%stateEveryNth == 0 || g.sim.Dead() || g.sim.Escaped() {
		s := snapshot(g.sim)
		trySend(g.link.Out, Msg{Type: "state", State: &s})
	}
}

func (g *Game) sendHello() {
	if g.link == nil {
		return
	}
	trySend(g.link.Out, Msg{
		Type:     "hello",
		Seed:     g.runSeed(),
		Width:    g.cfg.Width,
		Height:   g.cfg.Height,
		Consoles: g.cfg.Consoles,
	})
}

func (g *Game) drainLink() {
	if g.link == nil {
		return
	}
	for {
		select {
		case _, ok := <-g.link.In:
			if !ok {
				// The coordinator is gone; keep playing without a link.
				g.link = nil
				return
			}
		default:
			return
		}
	}
}

func snapshot(s *sim.Game) StateMsg {
	path := s.Alien.Path()
	cells := make([][2]int, 0, len(path))
	for _, c := range path {
		cells = append(cells, [2]int{c.X, c.Y})
	}
	ping, vent, search := s.LearnTiers()
	var hot [][2]int
	for i, h := range s.RoomHeat() {
		if h > 0.5 {
			hot = append(hot, [2]int{i, int(h * 100)})
		}
	}
	return StateMsg{
		Tick:       s.TickCount(),
		PlayerX:    s.Player.Pos.X,
		PlayerY:    s.Player.Pos.Y,
		PlayerA:    s.Player.Angle,
		AlienX:     s.Alien.Pos.X,
		AlienY:     s.Alien.Pos.Y,
		AlienState: s.Alien.State.String(),
		TargetX:    s.Alien.Target.X,
		TargetY:    s.Alien.Target.Y,
		Path:       cells,
		Menace:     s.Menace(),
		Done:       s.ObjectivesDone(),
		Total:      s.ObjectivesTotal(),
		Unlocked:   s.ExitUnlocked(),
		PingTier:   ping,
		VentTier:   vent,
		SearchTier: search,
		Hot:        hot,
		Dead:       s.Dead(),
		Escaped:    s.Escaped(),
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.WritePixels(g.renderer.Frame(g.sim, g.now))
}

func (g *Game) Layout(_, _ int) (int, int) {
	cfg := g.renderer.Config()
	return cfg.Width, cfg.Height
}

func randomSeed() int64 {
	return int64(rand.Uint64() >> 1)
}
