package gui

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
)

var ErrAudioUnavailable = errors.New("audio device unavailable")

func audioDeviceLikely() bool {
	// Ebiten's audio context is a process-wide singleton: once it fails on a
	// machine with no sound device the whole game loop errors out. On Linux a
	// missing /dev/snd is a cheap, reliable tell to stay silent instead.
	if runtime.GOOS != "linux" {
		return true
	}
	_, err := os.Stat("/dev/snd")
	return err == nil
}

func handleRunError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrAudioUnavailable) || strings.Contains(err.Error(), "oto:") {
		fmt.Fprintln(os.Stderr, "audio device unavailable — sound disabled; relaunch to keep playing")
		return nil
	}
	return err
}
