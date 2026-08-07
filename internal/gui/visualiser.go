package gui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/danielriddell21/nemesis/internal/vis"
)

// Visualiser is the hunter-AI window: it takes messages from the game window
// over the link and hands them to the display-free renderer, then blits the
// frame it draws.
type Visualiser struct {
	link *Link
	r    *vis.Renderer
	gone bool
}

var _ ebiten.Game = (*Visualiser)(nil)

func newVisualiser(cfg Config) *Visualiser {
	return &Visualiser{link: cfg.Link, r: vis.New()}
}

func (v *Visualiser) Update() error {
	if v.gone || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	if v.link != nil {
		for {
			select {
			case msg, ok := <-v.link.In:
				if !ok {
					// The game window is gone: close this one too.
					v.gone = true
					return nil
				}
				v.r.Apply(msg)
			default:
				v.r.Tick(tickDT)
				return nil
			}
		}
	}
	v.r.Tick(tickDT)
	return nil
}

func (v *Visualiser) Draw(screen *ebiten.Image) {
	screen.WritePixels(v.r.Frame())
}

func (v *Visualiser) Layout(_, _ int) (int, int) {
	w, h := vis.Size()
	return w, h
}
