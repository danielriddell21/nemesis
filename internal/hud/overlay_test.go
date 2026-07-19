package hud

import "testing"

func TestOverlayLifecycle(t *testing.T) {
	o := New()
	if _, _, ok := o.Active(); ok {
		t.Fatal("fresh overlay should be inactive")
	}
	o.Post("VENT CREAK", 2, Notice)
	text, ch, ok := o.Active()
	if !ok || text != "VENT CREAK" || ch != Notice {
		t.Fatalf("got %q/%v/%v", text, ch, ok)
	}
	o.Tick()
	if _, _, ok := o.Active(); !ok {
		t.Fatal("message should survive its first frame")
	}
	o.Tick()
	if _, _, ok := o.Active(); ok {
		t.Fatal("message should expire after its frames run out")
	}
}

func TestOverlayClearOnEmptyPost(t *testing.T) {
	o := New()
	o.Post("HELLO", 100, Diagnostic)
	o.Post("", 100, Diagnostic)
	if _, _, ok := o.Active(); ok {
		t.Fatal("posting empty text should clear the overlay")
	}
}
