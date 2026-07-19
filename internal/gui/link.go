package gui

import "github.com/danielriddell21/nemesis/internal/telemetry"

type Msg struct {
	Type     string            `json:"t"`
	Seed     int64             `json:"seed,omitempty"`
	Width    int               `json:"w,omitempty"`
	Height   int               `json:"h,omitempty"`
	Consoles int               `json:"consoles,omitempty"`
	State    *StateMsg         `json:"state,omitempty"`
	Events   []telemetry.Event `json:"events,omitempty"`
}

type StateMsg struct {
	Tick       uint64   `json:"tick"`
	PlayerX    float64  `json:"px"`
	PlayerY    float64  `json:"py"`
	PlayerA    float64  `json:"pa"`
	AlienX     float64  `json:"ax"`
	AlienY     float64  `json:"ay"`
	AlienState string   `json:"as"`
	TargetX    int      `json:"tgx"`
	TargetY    int      `json:"tgy"`
	Path       [][2]int `json:"path,omitempty"`
	Menace     float64  `json:"menace"`
	Done       int      `json:"done"`
	Total      int      `json:"total"`
	PingTier   int      `json:"lp,omitempty"`
	VentTier   int      `json:"lv,omitempty"`
	SearchTier int      `json:"ls,omitempty"`
	Hot        [][2]int `json:"hot,omitempty"`
	Unlocked   bool     `json:"unlocked,omitempty"`
	Dead       bool     `json:"dead,omitempty"`
	Escaped    bool     `json:"escaped,omitempty"`
}

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
