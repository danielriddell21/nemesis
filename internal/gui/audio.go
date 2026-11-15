package gui

import (
	"bytes"
	"fmt"
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2/audio"

	"github.com/danielriddell21/crucible/synth"

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

	heart float64
	growl float64
	groan float64
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
	a.playAt(cue, sim.Vec2{X: float64(e.X) + 0.5, Y: float64(e.Y) + 0.5}, 1)
}

func (a *Audio) playAt(cue iaudio.Cue, at sim.Vec2, level float64) {
	pcm, ok := a.pcm[cue]
	if !ok {
		return
	}
	dx, dy := at.X-a.listener.X, at.Y-a.listener.Y
	d := math.Hypot(dx, dy)
	gain := level * sfxBase * a.sfxGain * (1 - d/maxAudible)
	if gain <= 0 {
		return
	}
	// Pan the sound by its bearing relative to where the player is looking.
	bearing := math.Atan2(dy, dx) - a.facing
	left, right := synth.Pan(bearing)
	p := a.ctx.NewPlayerFromBytes(synth.Panned(pcm, left, right))
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

const (
	groanGap    = 14.0
	groanSpread = 20.0
	groanReach  = 9.0
)

func (a *Audio) TickTension(dt float64, s *sim.Game) {
	if a.muted {
		return
	}
	// The sounds with no single event behind them: the player's own
	// heartbeat, the hunter's breathing, and the hull settling.
	dist := s.Alien.Pos.Sub(s.Player.Pos).Len()
	a.heart -= dt
	fear := iaudio.Fear(dist, s.Alien.State == sim.StateHunt)
	if every, ok := iaudio.HeartbeatInterval(fear); ok && a.heart <= 0 {
		a.heart = every
		a.playAt(iaudio.CueHeartbeat, a.listener, 0.5+0.5*fear)
	}
	a.growl -= dt
	if every, ok := iaudio.GrowlInterval(dist); ok && a.growl <= 0 && s.Alien.State != sim.StateLurk {
		a.growl = every
		a.playAt(iaudio.CueGrowl, s.Alien.Pos, 1)
	}
	a.groan -= dt
	if a.groan <= 0 {
		// Somewhere off in the structure, never the same place twice.
		a.groan = groanGap + rand.Float64()*groanSpread
		bearing := rand.Float64() * 2 * math.Pi
		from := sim.Vec2{X: a.listener.X + math.Cos(bearing)*groanReach, Y: a.listener.Y + math.Sin(bearing)*groanReach}
		a.playAt(iaudio.CueGroan, from, 1)
	}
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
