package gui

import "github.com/danielriddell21/nemesis/internal/vis"

// Msg and StateMsg are the visualiser's message types. They are aliased here
// because the window hub speaks them, while the visualiser that consumes them
// lives in a display-free package.
type (
	Msg      = vis.Msg
	StateMsg = vis.StateMsg
)

type Link struct {
	In  <-chan Msg
	Out chan<- Msg
}

func trySend(ch chan<- Msg, m Msg) {
	select {
	case ch <- m:
	default:
	}
}
