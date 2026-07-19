package gui

func NewOffscreenVisualiser() *Visualiser {
	return &Visualiser{
		m:  newVisModel(),
		fb: make([]byte, visWindowW*visWindowH*4),
	}
}

func (v *Visualiser) Apply(m Msg) {
	v.m.apply(m)
}

func (v *Visualiser) TickModel(dt float64) {
	v.m.tick(dt)
}

func (v *Visualiser) RenderFrame() (fb []byte, w, h int) {
	v.render()
	return v.fb, visWindowW, visWindowH
}
