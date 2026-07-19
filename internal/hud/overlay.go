package hud

import "sync"

type Channel uint8

const (
	Diagnostic Channel = iota
	Notice
)

type Overlay struct {
	mu        sync.Mutex
	text      string
	channel   Channel
	remaining int
}

func New() *Overlay {
	return &Overlay{}
}

func (o *Overlay) Post(text string, frames int, ch Channel) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if text == "" || frames <= 0 {
		o.text, o.channel, o.remaining = "", Diagnostic, 0
		return
	}
	o.text, o.channel, o.remaining = text, ch, frames
}

func (o *Overlay) Tick() {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.remaining > 0 {
		o.remaining--
		if o.remaining == 0 {
			o.text = ""
		}
	}
}

func (o *Overlay) Active() (string, Channel, bool) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.text, o.channel, o.remaining > 0
}
