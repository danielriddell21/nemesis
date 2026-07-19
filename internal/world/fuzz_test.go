package world

import "testing"

// FuzzGenerate throws arbitrary configs at the generator: whatever the inputs,
// it must either fail cleanly or return a level that honours the guarantee.
func FuzzGenerate(f *testing.F) {
	f.Add(48, 32, int64(1), 3)
	f.Add(16, 16, int64(0), 1)
	f.Add(200, 5, int64(-9), 99)
	f.Fuzz(func(t *testing.T, w, h int, seed int64, consoles int) {
		if w > 128 || h > 128 || w < -128 || h < -128 {
			t.Skip("grid too large for fuzz budget")
		}
		l, err := Generate(Config{Width: w, Height: h, Seed: seed, Consoles: consoles})
		if err != nil {
			return
		}
		if !reachable(l, l.Spawn, l.Exit, blocksWalls(l)) {
			t.Fatalf("exit unreachable: w=%d h=%d seed=%d\n%s", w, h, seed, l)
		}
		if len(l.Consoles) == 0 {
			t.Fatalf("no consoles placed: w=%d h=%d seed=%d", w, h, seed)
		}
		if len(l.VentMouths) < 2 {
			t.Fatalf("no vent system: w=%d h=%d seed=%d", w, h, seed)
		}
	})
}
