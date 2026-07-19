package gui

import (
	"bytes"
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2/audio"

	iaudio "github.com/danielriddell21/nemesis/internal/audio"
	"github.com/danielriddell21/nemesis/internal/sim"
	"github.com/danielriddell21/nemesis/internal/telemetry"
)

const (
	maxAudible   = 18.0
	sfxLevel     = 0.6
	ambientLevel = 0.3
)

type Audio struct {
	ctx      *audio.Context
	players  map[iaudio.Cue]*audio.Player
	ambient  map[int]*audio.Player
	band     int
	listener sim.Vec2
}

func NewAudio() (*Audio, error) {
	if !audioDeviceLikely() {
		return nil, ErrAudioUnavailable
	}
	a := &Audio{
		ctx:     audio.NewContext(iaudio.SampleRate),
		players: make(map[iaudio.Cue]*audio.Player),
		ambient: make(map[int]*audio.Player),
		band:    -1,
	}
	for cue, pcm := range iaudio.Synth() {
		a.players[cue] = a.ctx.NewPlayerFromBytes(pcm)
	}
	return a, nil
}

func (a *Audio) SetListener(pos sim.Vec2) {
	a.listener = pos
}

func (a *Audio) PlayEvent(e telemetry.Event) {
	kind, ok := e.Kind()
	if !ok {
		return
	}
	cue, ok := iaudio.CueFor(kind)
	if !ok {
		return
	}
	p, ok := a.players[cue]
	if !ok {
		return
	}
	d := math.Hypot(a.listener.X-float64(e.X)-0.5, a.listener.Y-float64(e.Y)-0.5)
	vol := sfxLevel * (1 - d/maxAudible)
	if vol <= 0 {
		return
	}
	p.SetVolume(vol)
	if err := p.Rewind(); err != nil {
		return
	}
	p.Play()
}

func (a *Audio) TickAmbient(menace float64) {
	band := iaudio.MenaceBand(menace)
	if band == a.band {
		return
	}
	if cur, ok := a.ambient[a.band]; ok {
		cur.Pause()
	}
	p, ok := a.ambient[band]
	if !ok {
		pcm := iaudio.Ambient(band)
		loop := audio.NewInfiniteLoop(bytes.NewReader(pcm), int64(len(pcm)))
		var err error
		p, err = a.ctx.NewPlayer(loop)
		if err != nil {
			fmt.Println("audio: ambient band:", err)
			return
		}
		p.SetVolume(ambientLevel)
		a.ambient[band] = p
	}
	p.Play()
	a.band = band
}
