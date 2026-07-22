package gui

import "github.com/danielriddell21/crucible/record"

type Role uint8

const (
	RoleGame Role = iota
	RoleVisualiser
)

type Config struct {
	Seed     int64
	Width    int
	Height   int
	Consoles int

	Visualiser bool
	Role       Role
	Link       *Link

	// Rec holds the shared --record flags; when its path is set a bot-driven
	// visualiser captures frames to a GIF and exits — the AI-visualiser demo.
	// The demo is frame-delay paced (like gambit), so only --record and
	// --record-frames are exposed; playback rate and scale are fixed.
	Rec record.Options
}

func Available() bool { return true }
