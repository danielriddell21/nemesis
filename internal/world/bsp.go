package world

const (
	minLeaf  = 7
	maxLeaf  = 16
	minRoom  = 3
	roomPad  = 1
	maxDepth = 7
)

type bspNode struct {
	bounds      Room
	left, right *bspNode
	room        *Room
}

func (n *bspNode) leaf() bool {
	return n.left == nil && n.right == nil
}

func (g *rng) split(n *bspNode, depth int) {
	if depth >= maxDepth {
		return
	}
	w, h := n.bounds.W, n.bounds.H
	canV := w >= 2*minLeaf // vertical cut -> left/right
	canH := h >= 2*minLeaf // horizontal cut -> top/bottom
	if !canV && !canH {
		return
	}
	// Small enough regions stop splitting most of the time, for size variety.
	// The root always splits so even small maps contain more than one room.
	if depth > 0 && w <= maxLeaf && h <= maxLeaf && g.chance(0.3) {
		return
	}

	vertical := chooseSplitAxis(g, w, h, canV, canH)
	if vertical {
		at := g.between(minLeaf, w-minLeaf)
		n.left = &bspNode{bounds: Room{X: n.bounds.X, Y: n.bounds.Y, W: at, H: h}}
		n.right = &bspNode{bounds: Room{X: n.bounds.X + at, Y: n.bounds.Y, W: w - at, H: h}}
	} else {
		at := g.between(minLeaf, h-minLeaf)
		n.left = &bspNode{bounds: Room{X: n.bounds.X, Y: n.bounds.Y, W: w, H: at}}
		n.right = &bspNode{bounds: Room{X: n.bounds.X, Y: n.bounds.Y + at, W: w, H: h - at}}
	}
	g.split(n.left, depth+1)
	g.split(n.right, depth+1)
}

func chooseSplitAxis(g *rng, w, h int, canV, canH bool) bool {
	switch {
	case canV && !canH:
		return true
	case canH && !canV:
		return false
	case w > h:
		return true
	case h > w:
		return false
	default:
		return g.chance(0.5)
	}
}

func (g *rng) carveRooms(n *bspNode, l *Level) {
	if !n.leaf() {
		if n.left != nil {
			g.carveRooms(n.left, l)
		}
		if n.right != nil {
			g.carveRooms(n.right, l)
		}
		return
	}
	b := n.bounds
	maxW, maxH := b.W-2*roomPad, b.H-2*roomPad
	if maxW < minRoom || maxH < minRoom {
		return
	}
	rw := g.between(minRoom, maxW)
	rh := g.between(minRoom, maxH)
	rx := b.X + roomPad + g.intn(maxW-rw+1)
	ry := b.Y + roomPad + g.intn(maxH-rh+1)
	room := Room{X: rx, Y: ry, W: rw, H: rh}
	n.room = &room
	for y := ry; y < ry+rh; y++ {
		for x := rx; x < rx+rw; x++ {
			l.set(x, y, TileFloor)
		}
	}
}

func (g *rng) connect(n *bspNode, l *Level) *Room {
	if n.room != nil {
		return n.room
	}
	var lr, rr *Room
	if n.left != nil {
		lr = g.connect(n.left, l)
	}
	if n.right != nil {
		rr = g.connect(n.right, l)
	}
	if lr != nil && rr != nil {
		g.carveCorridor(lr.Center(), rr.Center(), l)
	}
	if lr != nil {
		return lr
	}
	return rr
}

func (g *rng) carveCorridor(a, b Coord, l *Level) {
	if g.chance(0.5) {
		carveH(l, a.X, b.X, a.Y)
		carveV(l, a.Y, b.Y, b.X)
	} else {
		carveV(l, a.Y, b.Y, a.X)
		carveH(l, a.X, b.X, b.Y)
	}
}

func carveH(l *Level, x1, x2, y int) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	for x := x1; x <= x2; x++ {
		if l.At(x, y) == TileWall {
			l.set(x, y, TileFloor)
		}
	}
}

func carveV(l *Level, y1, y2, x int) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	for y := y1; y <= y2; y++ {
		if l.At(x, y) == TileWall {
			l.set(x, y, TileFloor)
		}
	}
}

func (g *rng) carveStubs(l *Level) {
	const (
		attempts = 24
		maxStubs = 4
	)
	dirs := [4]Coord{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}}
	carved := 0
	for range attempts {
		if carved >= maxStubs {
			return
		}
		x := g.between(1, l.Width-2)
		y := g.between(1, l.Height-2)
		if l.At(x, y) != TileFloor {
			continue
		}
		d := dirs[g.intn(len(dirs))]
		length := g.between(1, 3)
		if digStub(l, Coord{X: x, Y: y}, d, length) {
			carved++
		}
	}
}

func digStub(l *Level, f, d Coord, length int) bool {
	dug := 0
	c := f
	for range length {
		c = Coord{X: c.X + d.X, Y: c.Y + d.Y}
		if c.X < 1 || c.Y < 1 || c.X >= l.Width-1 || c.Y >= l.Height-1 {
			break
		}
		if l.At(c.X, c.Y) != TileWall {
			break
		}
		open := 0
		for _, n := range neighbors4(c) {
			if l.At(n.X, n.Y).Walkable() {
				open++
			}
		}
		if open != 1 {
			break
		}
		l.set(c.X, c.Y, TileFloor)
		dug++
	}
	return dug > 0
}

func collectRooms(n *bspNode) []Room {
	if n == nil {
		return nil
	}
	if n.room != nil {
		return []Room{*n.room}
	}
	return append(collectRooms(n.left), collectRooms(n.right)...)
}
