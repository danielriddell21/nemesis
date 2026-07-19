package render

import (
	"testing"

	"github.com/danielriddell21/nemesis/internal/sim"
	"github.com/danielriddell21/nemesis/internal/world"
)

func BenchmarkFrame(b *testing.B) {
	l, err := world.Generate(world.Config{Width: 48, Height: 32, Seed: 3})
	if err != nil {
		b.Fatal(err)
	}
	g := sim.New(l)
	r := NewRenderer(DefaultConfig())
	b.ResetTimer()
	for b.Loop() {
		r.Frame(g, 0)
	}
}
