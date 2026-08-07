package vis

import "github.com/danielriddell21/nemesis/internal/telemetry"

// Msg is one update from the game to the visualiser. It is JSON-encoded when
// the two run as separate coordinated windows.
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
	Tick        uint64   `json:"tick"`
	PlayerX     float64  `json:"px"`
	PlayerY     float64  `json:"py"`
	PlayerA     float64  `json:"pa"`
	Hidden      bool     `json:"hidden,omitempty"`
	AlienX      float64  `json:"ax"`
	AlienY      float64  `json:"ay"`
	AlienA      float64  `json:"aa"`
	AlienState  string   `json:"as"`
	Vision      float64  `json:"vis"`
	VisionFOV   float64  `json:"visfov"`
	TargetX     int      `json:"tgx"`
	TargetY     int      `json:"tgy"`
	Path        [][2]int `json:"path,omitempty"`
	Menace      float64  `json:"menace"`
	Done        int      `json:"done"`
	Total       int      `json:"total"`
	Deck        int      `json:"deck,omitempty"`
	PingTier    int      `json:"lp,omitempty"`
	VentTier    int      `json:"lv,omitempty"`
	SearchTier  int      `json:"ls,omitempty"`
	DecoyTier   int      `json:"ld,omitempty"`
	LockerTier  int      `json:"lk,omitempty"`
	Hot         [][2]int `json:"hot,omitempty"`
	DecoyActive bool     `json:"decoy,omitempty"`
	DecoyX      float64  `json:"dx,omitempty"`
	DecoyY      float64  `json:"dy,omitempty"`
	Unlocked    bool     `json:"unlocked,omitempty"`
	Dead        bool     `json:"dead,omitempty"`
	Escaped     bool     `json:"escaped,omitempty"`
}
