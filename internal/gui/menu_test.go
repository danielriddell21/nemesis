package gui

import (
	"image/color"
	"testing"

	"github.com/danielriddell21/crucible/hud"
)

func testGameForMenus(t *testing.T) *Game {
	t.Helper()
	g, err := NewGame(Config{Width: 32, Height: 24}, 1, hud.New(), nil, Records{}, defaultSettings())
	if err != nil {
		t.Fatal(err)
	}
	return g
}

// Menu selection wraparound is covered by crucible/menu's own tests.

func TestMenuBackdropClearsWithoutGameFrame(t *testing.T) {
	g := testGameForMenus(t)
	// Stand in for a previously drawn screen: paint the whole canvas red.
	g.canvas.Fill(color.RGBA{R: 255, A: 255})
	g.lastFrame = nil // settings opened from the title, no game rendered yet
	g.composeMenu(g.settingsMenu)
	// A top corner the menu text never reaches must be repainted to the menu
	// backdrop, not left showing the stale red frame.
	px := g.canvas.Pixels()[:4]
	if px[0] != menuBG.R || px[1] != menuBG.G || px[2] != menuBG.B {
		t.Errorf("backdrop not cleared: corner = %v, want menuBG %v", px[:3], menuBG)
	}
}

func TestTitleMenuStartsRun(t *testing.T) {
	g := testGameForMenus(t)
	if g.state != stateTitle {
		t.Fatal("game should open on the title screen")
	}
	g.titleMenu.Items[0].Action() // DESCEND
	if g.state != statePlaying {
		t.Errorf("descend should start play, state = %v", g.state)
	}
	if g.deck != 0 || g.records.Runs != 1 {
		t.Errorf("a new run should be deck 0 and count a run, got deck %d runs %d", g.deck, g.records.Runs)
	}
}

func TestSettingsAdjustAndReturn(t *testing.T) {
	g := testGameForMenus(t)
	g.openSettings(stateTitle)
	if g.state != stateSettings {
		t.Fatal("openSettings should switch to the settings screen")
	}
	// SOUND toggle is the first row.
	before := g.settings.Sound
	g.settingsMenu.Items[0].Adjust(0)
	if g.settings.Sound == before {
		t.Error("adjusting sound should toggle it")
	}
	// SFX volume down clamps at zero.
	g.settings.SFXVolume = 0.05
	g.settingsMenu.Items[1].Adjust(-1)
	if g.settings.SFXVolume != 0 {
		t.Errorf("sfx volume should clamp to 0, got %v", g.settings.SFXVolume)
	}
	g.leaveSettings()
	if g.state != stateTitle {
		t.Errorf("leaving settings should return to the opener, got %v", g.state)
	}
}

func TestDescendCarriesLearningAcrossDecks(t *testing.T) {
	g := testGameForMenus(t)
	g.newRun()
	// Reach into the sim, mark some learning, then simulate an escape descent.
	g.deck = 0
	g.startDeck()
	if g.deck != 0 {
		t.Fatal("startDeck should not change the deck number")
	}
	g.carried.Pings = 9
	g.deck++
	g.startDeck()
	if g.deck != 1 {
		t.Errorf("descend should increment the deck, got %d", g.deck)
	}
	if g.sim.Learned().Pings != 9 {
		t.Error("learning should carry into the next deck")
	}
	if g.sim.Depth() != 1 {
		t.Errorf("deck 1 should run at depth 1, got %d", g.sim.Depth())
	}
}
