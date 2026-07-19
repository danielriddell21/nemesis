package world

const (
	corridorLight = 0.4
	ventLight     = 0.15
	flickerChance = 0.18
)

func assignLight(l *Level, g *rng, rooms []Room) {
	for _, r := range rooms {
		base := g.betweenF(0.5, 0.9)
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
