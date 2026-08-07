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
}

func Available() bool { return true }
