package gui

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

	// RecordPath, when set, runs a bot-driven visualiser that captures frames
	// to a GIF at this path and then exits — the AI-visualiser demo.
	RecordPath   string
	RecordFrames int
}

func Available() bool { return true }
