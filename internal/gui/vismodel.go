package gui

import (
	"fmt"

	"github.com/danielriddell21/nemesis/internal/telemetry"
	"github.com/danielriddell21/nemesis/internal/world"
)

const (
	visFeedLines = 16
	rippleLife   = 1.6
	visWindowW   = 736
	visWindowH   = 352
	visPanelW    = 300
)

type ripple struct {
	x, y   float64
	radius float64
	age    float64
}

type visModel struct {
	level     *world.Level
	state     StateMsg
	haveState bool
	activated map[world.Coord]bool
	feed      []string
	ripples   []ripple
}

func newVisModel() *visModel {
	return &visModel{activated: make(map[world.Coord]bool)}
}

func (m *visModel) apply(msg Msg) {
	switch msg.Type {
	case "hello":
		m.applyHello(msg)
	case "state":
		if msg.State != nil {
			m.state = *msg.State
			m.haveState = true
		}
	case "events":
		for _, e := range msg.Events {
			m.applyEvent(e)
		}
	}
}

func (m *visModel) applyHello(msg Msg) {
	// The level is deterministic from its seed, so the game only ships the
	// generation parameters and the visualiser regrows an identical map.
	l, err := world.Generate(world.Config{
		Width:    msg.Width,
		Height:   msg.Height,
		Seed:     msg.Seed,
		Consoles: msg.Consoles,
	})
	if err != nil {
		m.feed = append(m.feed, fmt.Sprintf("LEVEL REGEN FAILED: %v", err))
		return
	}
	m.level = l
	m.state = StateMsg{}
	m.haveState = false
	m.activated = make(map[world.Coord]bool)
	m.ripples = nil
	m.feed = append(m.feed, fmt.Sprintf("RUN START - SEED %d", msg.Seed))
	m.trimFeed()
}

func (m *visModel) applyEvent(e telemetry.Event) {
	if line := e.Line(); line != "" {
		m.feed = append(m.feed, line)
		m.trimFeed()
	}
	if kind, ok := e.Kind(); ok && kind.String() == "console" {
		m.activated[world.Coord{X: e.X, Y: e.Y}] = true
	}
	if e.Radius > 0 {
		m.ripples = append(m.ripples, ripple{
			x:      float64(e.X) + 0.5,
			y:      float64(e.Y) + 0.5,
			radius: e.Radius,
		})
	}
}

func (m *visModel) trimFeed() {
	if len(m.feed) > visFeedLines {
		m.feed = m.feed[len(m.feed)-visFeedLines:]
	}
}

func (m *visModel) tick(dt float64) {
	kept := m.ripples[:0]
	for _, r := range m.ripples {
		r.age += dt
		if r.age < rippleLife {
			kept = append(kept, r)
		}
	}
	m.ripples = kept
}
