package gui

import (
	"fmt"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/danielriddell21/crucible/hud"

	iaudio "github.com/danielriddell21/nemesis/internal/audio"
	"github.com/danielriddell21/nemesis/internal/render"
	"github.com/danielriddell21/nemesis/internal/sim"
	"github.com/danielriddell21/nemesis/internal/telemetry"
	"github.com/danielriddell21/nemesis/internal/world"
)

const (
	tickDT        = 1.0 / 60
	stateEveryNth = 4
)

type state int

const (
	stateTitle state = iota
	statePlaying
	statePaused
	stateSettings
)

type Game struct {
	cfg      Config
	sim      *sim.Game
	renderer *render.Renderer
	overlay  *hud.Overlay
	audio    *Audio
	records  Records
	settings Settings
	link     *Link

	state          state
	settingsReturn state
	baseSeed       int64
	deck           int
	carried        sim.Learned
	events         []telemetry.Event
	now            float64
	recorded       bool
	quit           bool

	canvas       *canvas
	lastFrame    []byte
	titleMenu    *menu
	pauseMenu    *menu
	settingsMenu *menu

	haveMouse  bool
	lastMouseX int
}

var _ ebiten.Game = (*Game)(nil)

func NewGame(cfg Config, seed int64, overlay *hud.Overlay, audio *Audio, records Records, settings Settings) (*Game, error) {
	rc := render.DefaultConfig()
	rc.FOV = settings.FOV
	g := &Game{
		cfg:      cfg,
		renderer: render.NewRenderer(rc, render.WithOverlay(overlay)),
		overlay:  overlay,
		audio:    audio,
		records:  records,
		settings: settings,
		link:     cfg.Link,
		baseSeed: seed,
		state:    stateTitle,
		canvas:   newCanvas(rc.Width, rc.Height),
	}
	g.audio.Configure(settings)
	g.buildMenus()
	return g, nil
}

func (g *Game) OnEvent(e telemetry.Event) {
	g.events = append(g.events, e)
}

func (g *Game) Update() error {
	if g.quit {
		return ebiten.Termination
	}
	g.setCursor()
	switch g.state {
	case stateTitle:
		g.playMenuSound(g.titleMenu.update())
	case statePlaying:
		return g.updatePlaying()
	case statePaused:
		g.updatePaused()
	case stateSettings:
		g.updateSettings()
	}
	return nil
}

func (g *Game) setCursor() {
	if g.state == statePlaying {
		ebiten.SetCursorMode(ebiten.CursorModeCaptured)
		return
	}
	ebiten.SetCursorMode(ebiten.CursorModeVisible)
}

func (g *Game) updatePlaying() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
	g.drainLink()
	g.now += tickDT

	if g.sim.Dead() || g.sim.Escaped() {
		g.finishDeck()
		g.advanceOnEnter()
		g.overlay.Tick()
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.state = statePaused
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

func (g *Game) advanceOnEnter() {
	if !inpututil.IsKeyJustPressed(ebiten.KeyEnter) && !inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		return
	}
	if g.sim.Escaped() {
		g.carried = g.sim.Learned()
		g.deck++
		g.startDeck()
		return
	}
	g.toTitle()
}

func (g *Game) updatePaused() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.playMenuSound(soundSelect)
		g.state = statePlaying
		return
	}
	g.playMenuSound(g.pauseMenu.update())
}

func (g *Game) updateSettings() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.playMenuSound(soundSelect)
		g.leaveSettings()
		return
	}
	g.playMenuSound(g.settingsMenu.update())
}

func (g *Game) playMenuSound(s menuSound) {
	switch s {
	case soundMove:
		g.audio.PlayUI(iaudio.CueMenuMove)
	case soundSelect:
		g.audio.PlayUI(iaudio.CueMenuSelect)
	}
}

func (g *Game) newRun() {
	g.deck = 0
	g.carried = sim.Learned{}
	g.records.startedRun()
	_ = g.records.save()
	g.startDeck()
}

func (g *Game) toTitle() {
	g.state = stateTitle
	g.buildMenus()
}

func (g *Game) startDeck() {
	consoles := 3 + g.deck
	width := min(g.cfg.Width+g.deck*4, 80)
	height := min(g.cfg.Height+g.deck*2, 56)
	seed := g.baseSeed + int64(g.deck)*0x9e3779b9
	level, err := world.Generate(world.Config{Width: width, Height: height, Seed: seed, Consoles: consoles})
	if err != nil {
		fmt.Println("generate deck:", err)
		g.toTitle()
		return
	}
	fmt.Printf("nemesis — deck %d, seed %d\n", g.deck+1, seed)
	g.sim = sim.New(level,
		sim.WithObserver(telemetry.NewBus(g)),
		sim.WithLearned(g.carried),
		sim.WithDepth(g.deck),
	)
	g.recorded = false
	g.events = g.events[:0]
	g.overlay.Post("BRING THE GENERATORS ONLINE - THEN THE AIRLOCK", 240, hud.Notice)
	g.state = statePlaying
	g.sendHello()
}

func (g *Game) finishDeck() {
	if g.recorded {
		return
	}
	g.recorded = true
	if g.sim.Escaped() {
		g.records.clearedDeck(g.deck, g.sim.Elapsed())
		g.overlay.Post(fmt.Sprintf("DECK %d CLEARED - [ENTER] DESCEND DEEPER", g.deck+1), 100000, hud.Notice)
	} else {
		g.records.diedOn(g.deck)
		g.overlay.Post("IT FOUND YOU - [ENTER] RETURN TO TITLE", 100000, hud.Notice)
	}
	_ = g.records.save()
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
	g.audio.SetListener(g.sim.Player.Pos, g.sim.Player.Angle)
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
		s := Snapshot(g.sim)
		trySend(g.link.Out, Msg{Type: "state", State: &s})
	}
}

func (g *Game) sendHello() {
	if g.link == nil {
		return
	}
	trySend(g.link.Out, Msg{
		Type:     "hello",
		Seed:     g.baseSeed + int64(g.deck)*0x9e3779b9,
		Width:    min(g.cfg.Width+g.deck*4, 80),
		Height:   min(g.cfg.Height+g.deck*2, 56),
		Consoles: 3 + g.deck,
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
				g.link = nil
				return
			}
		default:
			return
		}
	}
}

func Snapshot(s *sim.Game) StateMsg {
	path := s.Alien.Path()
	cells := make([][2]int, 0, len(path))
	for _, c := range path {
		cells = append(cells, [2]int{c.X, c.Y})
	}
	ping, vent, search := s.LearnTiers()
	decoy, locker := s.LearnExtras()
	var hot [][2]int
	for i, h := range s.RoomHeat() {
		if h > 0.5 {
			hot = append(hot, [2]int{i, int(h * 100)})
		}
	}
	d := s.DecoyState()
	return StateMsg{
		Tick:        s.TickCount(),
		PlayerX:     s.Player.Pos.X,
		PlayerY:     s.Player.Pos.Y,
		PlayerA:     s.Player.Angle,
		Hidden:      s.Player.Hidden,
		AlienX:      s.Alien.Pos.X,
		AlienY:      s.Alien.Pos.Y,
		AlienA:      s.Alien.Facing,
		AlienState:  s.Alien.State.String(),
		Vision:      s.VisionRange(),
		VisionFOV:   s.VisionFOV(),
		TargetX:     s.Alien.Target.X,
		TargetY:     s.Alien.Target.Y,
		Path:        cells,
		Menace:      s.Menace(),
		Done:        s.ObjectivesDone(),
		Total:       s.ObjectivesTotal(),
		Deck:        s.Depth() + 1,
		Unlocked:    s.ExitUnlocked(),
		PingTier:    ping,
		VentTier:    vent,
		SearchTier:  search,
		DecoyTier:   decoy,
		LockerTier:  locker,
		Hot:         hot,
		DecoyActive: d.Active,
		DecoyX:      d.Pos.X,
		DecoyY:      d.Pos.Y,
		Dead:        s.Dead(),
		Escaped:     s.Escaped(),
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.state {
	case statePlaying:
		frame := g.renderer.Frame(g.sim, g.now)
		g.lastFrame = append(g.lastFrame[:0], frame...)
		screen.WritePixels(frame)
	case stateTitle:
		g.canvas.fill(menuBG)
		g.titleMenu.draw(g.canvas)
		screen.WritePixels(g.canvas.pixels())
	case statePaused:
		g.drawMenuOverlay(screen, g.pauseMenu)
	case stateSettings:
		g.drawMenuOverlay(screen, g.settingsMenu)
	}
}

func (g *Game) drawMenuOverlay(screen *ebiten.Image, m *menu) {
	g.composeMenu(m)
	screen.WritePixels(g.canvas.pixels())
}

func (g *Game) composeMenu(m *menu) {
	// Settings can be opened straight from the title, before any game frame
	// exists to dim behind the menu; fall back to a solid backdrop so the menu
	// does not draw over stale pixels.
	if len(g.lastFrame) == 0 {
		g.canvas.fill(menuBG)
	} else {
		g.canvas.dimFrom(g.lastFrame, 0.3)
	}
	m.draw(g.canvas)
}

func (g *Game) Layout(_, _ int) (int, int) {
	cfg := g.renderer.Config()
	return cfg.Width, cfg.Height
}

func randomSeed() int64 {
	return int64(rand.Uint64() >> 1)
}
