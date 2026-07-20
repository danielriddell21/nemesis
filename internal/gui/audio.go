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

const maxAudible = 18.0

type Audio struct {
	ctx      *audio.Context
	pcm      map[iaudio.Cue][]byte
	voices   []*audio.Player
	ambient  map[int]*audio.Player
	band     int
	listener sim.Vec2
	facing   float64

	sfxGain     float64
	ambientGain float64
	muted       bool
}

func NewAudio() (*Audio, error) {
	if !audioDeviceLikely() {
		return nil, ErrAudioUnavailable
	}
	a := &Audio{
		ctx:         audio.NewContext(iaudio.SampleRate),
		pcm:         iaudio.Synth(),
		ambient:     make(map[int]*audio.Player),
		band:        -1,
		sfxGain:     0.8,
		ambientGain: 0.6,
	}
	return a, nil
}

func (a *Audio) Configure(s Settings) {
	if a == nil {
		return
	}
	a.sfxGain = s.SFXVolume
	a.ambientGain = s.AmbientVolume
	a.muted = !s.Sound
	if cur, ok := a.ambient[a.band]; ok {
		if a.muted {
			cur.Pause()
		} else {
			cur.SetVolume(ambientBase * a.ambientGain)
			cur.Play()
		}
	}
}

func (a *Audio) SetListener(pos sim.Vec2, facing float64) {
	a.listener = pos
	a.facing = facing
}

const (
	sfxBase     = 0.6
	ambientBase = 0.5
)

func (a *Audio) PlayEvent(e telemetry.Event) {
	if a.muted {
		return
	}
	kind, ok := e.Kind()
	if !ok {
		return
	}
	cue, ok := iaudio.CueFor(kind)
	if !ok {
		return
	}
	pcm, ok := a.pcm[cue]
	if !ok {
		return
	}
	dx, dy := float64(e.X)+0.5-a.listener.X, float64(e.Y)+0.5-a.listener.Y
	d := math.Hypot(dx, dy)
	gain := sfxBase * a.sfxGain * (1 - d/maxAudible)
	if gain <= 0 {
		return
	}
	// Pan the sound by its bearing relative to where the player is looking.
	bearing := math.Atan2(dy, dx) - a.facing
	left, right := iaudio.Pan(bearing)
	p := a.ctx.NewPlayerFromBytes(iaudio.Panned(pcm, left, right))
	p.SetVolume(gain)
	p.Play()
	a.reap(p)
}

const uiBase = 0.5

// PlayUI plays a menu blip: no panning or distance falloff, since the interface
// has no place in the world.
func (a *Audio) PlayUI(cue iaudio.Cue) {
	if a == nil || a.muted {
		return
	}
	pcm, ok := a.pcm[cue]
	if !ok {
		return
	}
	p := a.ctx.NewPlayerFromBytes(pcm)
	p.SetVolume(uiBase * a.sfxGain)
	p.Play()
	a.reap(p)
}

// reap keeps a short list of the most recent one-shot voices alive until they
// finish, then lets them be collected — ebiten players stop when unreferenced.
func (a *Audio) reap(p *audio.Player) {
	live := a.voices[:0]
	for _, v := range a.voices {
		if v.IsPlaying() {
			live = append(live, v)
		}
	}
	live = append(live, p)
	a.voices = live
}

func (a *Audio) TickAmbient(menace float64) {
	if a.muted {
		return
	}
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
		a.ambient[band] = p
	}
	p.SetVolume(ambientBase * a.ambientGain)
	p.Play()
	a.band = band
}
