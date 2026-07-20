package main

import (
	"fmt"
	"image"
	"image/color"

	"github.com/danielriddell21/nemesis/internal/world"
)

const (
	montageTile = 6
	montageCols = 3
	montageRows = 2
	montageGap  = 10
)

var montagePalette = map[world.TileType]color.RGBA{
	world.TileWall:    {R: 38, G: 43, B: 51, A: 255},
	world.TileFloor:   {R: 74, G: 80, B: 90, A: 255},
	world.TileVent:    {R: 110, G: 92, B: 64, A: 255},
	world.TileDoor:    {R: 168, G: 132, B: 72, A: 255},
	world.TileSpawn:   {R: 120, G: 230, B: 160, A: 255},
	world.TileExit:    {R: 190, G: 70, B: 60, A: 255},
	world.TileConsole: {R: 200, G: 190, B: 90, A: 255},
	world.TileLocker:  {R: 90, G: 180, B: 200, A: 255},
}

func stationsMontage(path string) error {
	seeds := []int64{7, 21, 42, 99, 1234, 31337}
	cellW := demoWidth*montageTile + montageGap
	cellH := demoHeight*montageTile + montageGap
	img := image.NewRGBA(image.Rect(0, 0, montageCols*cellW+montageGap, montageRows*cellH+montageGap))
	bg := [4]uint8{12, 14, 18, 255}
	for i := range img.Pix {
		img.Pix[i] = bg[i%4]
	}
	for i, seed := range seeds {
		l, err := world.Generate(world.Config{Width: demoWidth, Height: demoHeight, Seed: seed})
		if err != nil {
			return fmt.Errorf("montage seed %d: %w", seed, err)
		}
		offX := montageGap + (i%montageCols)*cellW
		offY := montageGap + (i/montageCols)*cellH
		drawStation(img, l, offX, offY)
	}
	return writePNG(path, img.Pix, img.Rect.Dx(), img.Rect.Dy())
}

func drawStation(img *image.RGBA, l *world.Level, offX, offY int) {
	for y := range l.Height {
		for x := range l.Width {
			c, ok := montagePalette[l.At(x, y)]
			if !ok {
				continue
			}
			// Floors carry their room's gloom so the lighting variety shows.
			if l.At(x, y) == world.TileFloor {
				k := 0.5 + 0.5*l.LightAt(x, y)
				c = color.RGBA{
					R: uint8(float64(c.R) * k),
					G: uint8(float64(c.G) * k),
					B: uint8(float64(c.B) * k),
					A: 255,
				}
			}
			for dy := range montageTile {
				for dx := range montageTile {
					img.SetRGBA(offX+x*montageTile+dx, offY+y*montageTile+dy, c)
				}
			}
		}
	}
}
