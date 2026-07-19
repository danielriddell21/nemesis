package sim

type noiseEvent struct {
	at     Vec2
	radius float64
	kind   ObservationKind
}

func (n *noiseEvent) merge(o noiseEvent) {
	// The hunter reacts to the loudest thing it can hear, so overlapping noise
	// keeps only the widest event.
	if o.radius > n.radius {
		*n = o
	}
}
