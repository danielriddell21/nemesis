package sim

import "testing"

func TestDepthWakesHunterSooner(t *testing.T) {
	shallow := New(flatLevel(24, 16))
	deep := New(flatLevel(24, 16), WithDepth(4))
	if deep.director.wake >= shallow.director.wake {
		t.Errorf("deeper decks should wake the hunter sooner: deep=%v shallow=%v",
			deep.director.wake, shallow.director.wake)
	}
	if deep.director.aggression <= shallow.director.aggression {
		t.Error("deeper decks should start the hunter warier")
	}
	if deep.threat <= 1 {
		t.Errorf("deeper decks should raise the threat multiplier, got %v", deep.threat)
	}
	if deep.Depth() != 4 {
		t.Errorf("Depth() = %d, want 4", deep.Depth())
	}
}

func TestDepthZeroIsBaseline(t *testing.T) {
	g := New(flatLevel(24, 16), WithDepth(0))
	if g.threat != 1 || g.Depth() != 0 {
		t.Errorf("depth 0 should be the baseline: threat=%v depth=%d", g.threat, g.Depth())
	}
}
