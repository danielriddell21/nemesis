package gui

import (
	"fmt"
	"strings"

	"github.com/danielriddell21/crucible/keymap"
	"github.com/danielriddell21/crucible/menu"
)

// controlHints is nemesis's control scheme, shown on the pause menu and shared
// through crucible/keymap so its "key: action" format matches the family.
var controlHints = []keymap.Binding{
	{Key: "WASD", Action: "MOVE"},
	{Key: "MOUSE", Action: "LOOK"},
	{Key: "SHIFT", Action: "RUN"},
	{Key: "CTRL", Action: "SNEAK"},
	{Key: "E", Action: "USE"},
	{Key: "F", Action: "HIDE"},
	{Key: "T", Action: "TRACKER"},
	{Key: "Q", Action: "DECOY"},
}

// controlsSubtitle formats the control scheme as two centred rows for a menu
// subtitle.
func controlsSubtitle() []string {
	row := func(bs []keymap.Binding) string {
		parts := make([]string, len(bs))
		for i, b := range bs {
			parts[i] = b.Label()
		}
		return strings.Join(parts, "    ")
	}
	return []string{row(controlHints[:4]), row(controlHints[4:]), ""}
}

func (g *Game) buildMenus() {
	g.titleMenu = &menu.Menu{
		Title:    "N E M E S I S",
		Subtitle: append([]string{"you are not alone on this station", ""}, g.records.summary()...),
		Items: []menu.Item{
			{Label: func() string { return "DESCEND" }, Action: g.newRun},
			{Label: func() string { return "SETTINGS" }, Action: func() { g.openSettings(stateTitle) }},
			{Label: func() string { return "QUIT" }, Action: func() { g.quit = true }},
		},
	}
	g.pauseMenu = &menu.Menu{
		Title:    "PAUSED",
		Subtitle: controlsSubtitle(),
		Items: []menu.Item{
			{Label: func() string { return "RESUME" }, Action: func() { g.state = statePlaying }},
			{Label: func() string { return "SETTINGS" }, Action: func() { g.openSettings(statePaused) }},
			{Label: func() string { return "ABANDON RUN" }, Action: g.toTitle},
		},
	}
	g.settingsMenu = g.buildSettingsMenu()
}

func (g *Game) buildSettingsMenu() *menu.Menu {
	return &menu.Menu{
		Title:    "SETTINGS",
		Subtitle: []string{"left / right adjust     enter or esc: back", ""},
		Items: []menu.Item{
			{
				Label:  func() string { return "SOUND            " + menu.OnOff(g.settings.Sound) },
				Adjust: func(int) { g.settings.Sound = !g.settings.Sound; g.applySettings() },
			},
			{
				Label:  func() string { return "SFX VOLUME       " + menu.Bar(g.settings.SFXVolume) },
				Adjust: func(d int) { g.settings.SFXVolume = clamp01(g.settings.SFXVolume + float64(d)*0.1); g.applySettings() },
			},
			{
				Label: func() string { return "AMBIENT VOLUME   " + menu.Bar(g.settings.AmbientVolume) },
				Adjust: func(d int) {
					g.settings.AmbientVolume = clamp01(g.settings.AmbientVolume + float64(d)*0.1)
					g.applySettings()
				},
			},
			{
				Label:  func() string { return "MOUSE SENS       " + fmt.Sprintf("%.1f", g.settings.Sensitivity) },
				Adjust: func(d int) { g.settings.Sensitivity = clampRange(g.settings.Sensitivity+float64(d)*0.1, 0.2, 3.0) },
			},
			{
				Label:  func() string { return "FIELD OF VIEW    " + fmt.Sprintf("%d", int(g.settings.FOV*100)) },
				Adjust: func(d int) { g.settings.FOV = clampRange(g.settings.FOV+float64(d)*0.05, 0.8, 1.5); g.applySettings() },
			},
			{Label: func() string { return "BACK" }, Action: g.leaveSettings},
		},
	}
}

func (g *Game) openSettings(from state) {
	g.settingsReturn = from
	g.settingsMenu.Sel = 0
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
