package render

import "image/color"

type Config struct {
	Width  int
	Height int
	FOV    float64
}

func DefaultConfig() Config {
	return Config{Width: 640, Height: 400, FOV: 1.152}
}

var palette = struct {
	ceiling       color.RGBA
	floor         color.RGBA
	wall          color.RGBA
	door          color.RGBA
	vent          color.RGBA
	console       color.RGBA
	consoleActive color.RGBA
	alien         color.RGBA
	exitLocked    color.RGBA
	exitOpen      color.RGBA
	hudText       color.RGBA
	hudDim        color.RGBA
	hudDrop       color.RGBA
	tracker       color.RGBA
	blip          color.RGBA
	deathTint     color.RGBA
	escapeTint    color.RGBA
}{
	ceiling:       color.RGBA{R: 16, G: 18, B: 22, A: 255},
	floor:         color.RGBA{R: 30, G: 32, B: 36, A: 255},
	wall:          color.RGBA{R: 92, G: 100, B: 110, A: 255},
	door:          color.RGBA{R: 110, G: 92, B: 60, A: 255},
	vent:          color.RGBA{R: 60, G: 54, B: 46, A: 255},
	console:       color.RGBA{R: 70, G: 90, B: 96, A: 255},
	consoleActive: color.RGBA{R: 70, G: 160, B: 96, A: 255},
	alien:         color.RGBA{R: 26, G: 20, B: 30, A: 255},
	exitLocked:    color.RGBA{R: 180, G: 60, B: 50, A: 255},
	exitOpen:      color.RGBA{R: 80, G: 220, B: 120, A: 255},
	hudText:       color.RGBA{R: 200, G: 220, B: 210, A: 255},
	hudDim:        color.RGBA{R: 110, G: 130, B: 120, A: 255},
	hudDrop:       color.RGBA{R: 0, G: 0, B: 0, A: 255},
	tracker:       color.RGBA{R: 90, G: 200, B: 160, A: 255},
	blip:          color.RGBA{R: 220, G: 250, B: 230, A: 255},
	deathTint:     color.RGBA{R: 120, G: 10, B: 10, A: 255},
	escapeTint:    color.RGBA{R: 20, G: 80, B: 40, A: 255},
}
