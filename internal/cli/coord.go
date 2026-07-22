package cli

import (
	"fmt"
	"os"

	"github.com/danielriddell21/crucible/hub"

	"github.com/danielriddell21/nemesis/internal/gui"
)

func hubConfig(self string) hub.Config[gui.Msg] {
	return hub.Config[gui.Msg]{
		Self:      self,
		ChildArgs: func(idx int) []string { return []string{fmt.Sprintf("--child=%d", idx)} },
		Route:     route,
	}
}

// route sorts the game's messages into hub routes: the opening hello is the
// shared state a late-joining visualiser needs, so it is cached and replayed;
// per-tick state and event batches fan out to the other windows.
func route(m gui.Msg) hub.Route {
	switch m.Type {
	case "hello":
		return hub.RouteState
	case "state", "events":
		return hub.RouteBroadcast
	default:
		return hub.RouteNone
	}
}

func runRecord(cfg gui.Config) error {
	if err := gui.RunRecord(cfg); err != nil {
		return fmt.Errorf("record visualiser: %w", err)
	}
	return nil
}

func runLeader(cfg gui.Config) error {
	if !cfg.Visualiser {
		cfg.Role = gui.RoleGame
		if err := gui.Run(cfg); err != nil {
			return fmt.Errorf("run gui: %w", err)
		}
		return nil
	}
	return lead(cfg, os.Args[0], gui.Run)
}

// lead runs the game window as the hub leader and spawns the visualiser as a
// child window at startup, relaying messages between the two.
func lead(cfg gui.Config, self string, runWindow func(gui.Config) error) error {
	h := hub.New(hubConfig(self))
	leaderIn := make(chan gui.Msg, 64)
	leaderOut := make(chan gui.Msg, 64)
	h.AddParticipant(leaderIn, nil) // the game window is id 0
	go h.Run()
	go func() {
		for m := range leaderOut {
			h.Inject(0, m)
		}
	}()
	h.SpawnChild() // the visualiser

	cfg.Role = gui.RoleGame
	cfg.Link = &gui.Link{In: leaderIn, Out: leaderOut}
	err := runWindow(cfg)
	close(leaderOut)
	h.Shutdown()
	if err != nil {
		return fmt.Errorf("run game: %w", err)
	}
	return nil
}

func runChild(cfg gui.Config) error {
	err := hub.RunChild(func(l hub.Link[gui.Msg]) error {
		cfg.Role = gui.RoleVisualiser
		cfg.Link = &gui.Link{In: l.In, Out: l.Out}
		if err := gui.Run(cfg); err != nil {
			return fmt.Errorf("run visualiser: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("run visualiser window: %w", err)
	}
	return nil
}
