package gui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/danielriddell21/nemesis/internal/sim"
)

const mouseSensitivity = 0.0035

func (g *Game) readInput() sim.Input {
	var in sim.Input

	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		in.Forward++
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		in.Forward--
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		in.Strafe++
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		in.Strafe--
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		in.Turn++
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		in.Turn--
	}
	in.TurnDelta = g.mouseTurn()
	in.Mode = readMode()

	if inpututil.IsKeyJustPressed(ebiten.KeyE) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		in.Use = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyT) || ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		in.Tracker = true
	}
	return in
}

func readMode() sim.MoveMode {
	switch {
	case ebiten.IsKeyPressed(ebiten.KeyShiftLeft) || ebiten.IsKeyPressed(ebiten.KeyShiftRight):
		return sim.ModeRun
	case ebiten.IsKeyPressed(ebiten.KeyControlLeft) || ebiten.IsKeyPressed(ebiten.KeyC):
		return sim.ModeSneak
	default:
		return sim.ModeWalk
	}
}

func (g *Game) mouseTurn() float64 {
	mx, _ := ebiten.CursorPosition()
	if !g.haveMouse {
		g.haveMouse = true
		g.lastMouseX = mx
		return 0
	}
	delta := mx - g.lastMouseX
	g.lastMouseX = mx
	return float64(delta) * mouseSensitivity
}
