package main

import (
	"fmt"
	"image"
	"image/color"

	"github.com/danielriddell21/crucible/demo"
	"github.com/danielriddell21/crucible/record"

	"github.com/danielriddell21/nemesis/internal/world"
)

const (
	montageTile = 6
	montageCols = 3
	montageGap  = 10
)

// montageBG is the dark backdrop behind and between the station cells.
var montageBG = color.RGBA{R: 12, G: 14, B: 18, A: 255}

var montagePalette = map[world.Tile]color.RGBA{
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
	cells := make([]image.Image, 0, len(seeds))
	for _, seed := range seeds {
		l, err := world.Generate(world.Config{Width: demoWidth, Height: demoHeight, Seed: seed})
		if err != nil {
			return fmt.Errorf("montage seed %d: %w", seed, err)
		}
		cells = append(cells, stationCell(l))
	}
	if err := record.SavePNG(path, demo.Montage(cells, montageCols, montageGap, montageBG)); err != nil {
		return fmt.Errorf("write montage: %w", err)
	}
	fmt.Println(path)
	return nil
}

// stationCell paints one generated level as a tile map, for the contact sheet.
func stationCell(l *world.Level) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, l.W*montageTile, l.H*montageTile))
	for i := range img.Pix {
		switch i % 4 {
		case 0:
			img.Pix[i] = montageBG.R
		case 1:
			img.Pix[i] = montageBG.G
		case 2:
			img.Pix[i] = montageBG.B
		default:
			img.Pix[i] = montageBG.A
		}
	}
	drawStation(img, l, 0, 0)
	return img
}

func drawStation(img *image.RGBA, l *world.Level, offX, offY int) {
	for y := range l.H {
		for x := range l.W {
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
