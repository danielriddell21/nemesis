package gui

import "fmt"

func (g *Game) buildMenus() {
	g.titleMenu = &menu{
		title:    "N E M E S I S",
		subtitle: append([]string{"you are not alone on this station", ""}, g.records.summary()...),
		items: []menuItem{
			{label: func() string { return "DESCEND" }, action: g.newRun},
			{label: func() string { return "SETTINGS" }, action: func() { g.openSettings(stateTitle) }},
			{label: func() string { return "QUIT" }, action: func() { g.quit = true }},
		},
	}
	g.pauseMenu = &menu{
		title: "PAUSED",
		items: []menuItem{
			{label: func() string { return "RESUME" }, action: func() { g.state = statePlaying }},
			{label: func() string { return "SETTINGS" }, action: func() { g.openSettings(statePaused) }},
			{label: func() string { return "ABANDON RUN" }, action: g.toTitle},
		},
	}
	g.settingsMenu = g.buildSettingsMenu()
}

func (g *Game) buildSettingsMenu() *menu {
	onOff := func(b bool) string {
		if b {
			return "ON"
		}
		return "OFF"
	}
	return &menu{
		title:    "SETTINGS",
		subtitle: []string{"left / right adjust     enter or esc: back", ""},
		items: []menuItem{
			{
				label:  func() string { return "SOUND            " + onOff(g.settings.Sound) },
				adjust: func(int) { g.settings.Sound = !g.settings.Sound; g.applySettings() },
			},
			{
				label:  func() string { return "SFX VOLUME       " + bar(g.settings.SFXVolume) },
				adjust: func(d int) { g.settings.SFXVolume = clamp01(g.settings.SFXVolume + float64(d)*0.1); g.applySettings() },
			},
			{
				label: func() string { return "AMBIENT VOLUME   " + bar(g.settings.AmbientVolume) },
				adjust: func(d int) {
					g.settings.AmbientVolume = clamp01(g.settings.AmbientVolume + float64(d)*0.1)
					g.applySettings()
				},
			},
			{
				label:  func() string { return "MOUSE SENS       " + fmt.Sprintf("%.1f", g.settings.Sensitivity) },
				adjust: func(d int) { g.settings.Sensitivity = clampRange(g.settings.Sensitivity+float64(d)*0.1, 0.2, 3.0) },
			},
			{
				label:  func() string { return "FIELD OF VIEW    " + fmt.Sprintf("%d", int(g.settings.FOV*100)) },
				adjust: func(d int) { g.settings.FOV = clampRange(g.settings.FOV+float64(d)*0.05, 0.8, 1.5); g.applySettings() },
			},
			{label: func() string { return "BACK" }, action: g.leaveSettings},
		},
	}
}

func (g *Game) openSettings(from state) {
	g.settingsReturn = from
	g.settingsMenu.sel = 0
	g.state = stateSettings
}

func (g *Game) leaveSettings() {
	g.applySettings()
	_ = g.settings.save()
	g.state = g.settingsReturn
	if g.state == stateTitle {
		g.buildMenus()
	}
}

func (g *Game) applySettings() {
	g.audio.Configure(g.settings)
	g.renderer.SetFOV(g.settings.FOV)
}

// bar renders a 0..1 value as a ten-cell meter for the settings screen.
func bar(v float64) string {
	filled := int(v*10 + 0.5)
	out := make([]byte, 10)
	for i := range out {
		if i < filled {
			out[i] = '#'
		} else {
			out[i] = '-'
		}
	}
	return string(out)
}
