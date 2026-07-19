package world

// This file assigns per-tile lighting. The station is kept dim — rooms each get
// their own gloom, corridors are darker still, and the vents are near-black —
// with a scattering of failing fixtures that stutter, because a light that
// cannot be trusted is worth two that can.

const (
	// corridorLight is the brightness of corridor and door cells.
	corridorLight = 0.35
	// ventLight is the brightness inside the ducts.
	ventLight = 0.15
	// flickerChance is the probability a room's fixtures are failing.
	flickerChance = 0.18
)

// assignLight stamps a brightness level onto every walkable cell: a per-room
// roll for rooms, uniform gloom for corridors and vents. Rooms that roll a
// failing fixture flicker at render time.
func assignLight(l *Level, g *rng, rooms []Room) {
	for _, r := range rooms {
		base := g.betweenF(0.45, 0.85)
		flicker := g.chance(flickerChance)
		for y := r.Y; y < r.Y+r.H; y++ {
			for x := r.X; x < r.X+r.W; x++ {
				i := y*l.Width + x
				l.Light[i] = base
				l.Flicker[i] = flicker
			}
		}
	}
	for i, t := range l.Tiles {
		if l.Light[i] > 0 {
			continue
		}
		switch t {
		case TileVent:
			l.Light[i] = ventLight
		case TileFloor, TileDoor, TileSpawn, TileExit:
			l.Light[i] = corridorLight
		}
	}
}
