package gui

import (
	"errors"
	"fmt"
	"testing"
)

func TestHandleRunError(t *testing.T) {
	if err := handleRunError(nil); err != nil {
		t.Errorf("nil should pass through, got %v", err)
	}
	audioErr := fmt.Errorf("run: %w", ErrAudioUnavailable)
	if err := handleRunError(audioErr); err != nil {
		t.Errorf("audio unavailability should be swallowed, got %v", err)
	}
	otoErr := errors.New(`audio error: oto: ALSA error at snd_pcm_open: "default": No such file or directory`)
	if err := handleRunError(otoErr); err != nil {
		t.Errorf("async oto failure should be swallowed, got %v", err)
	}
	real := errors.New("the GPU fell out")
	if err := handleRunError(real); !errors.Is(err, real) {
		t.Errorf("real errors must propagate, got %v", err)
	}
}
