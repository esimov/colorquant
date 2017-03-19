package main

import (
	"image"
	"image/draw"
	"image/color"
)

type Settings struct {
	Filter [][]float32
}

type Dither struct {
	Type string
	Settings
}

func (dither Dither) Process(input image.Image, mul float32) image.Image {
	bounds := input.Bounds()
	img := image.NewRGBA(bounds)
	for x := bounds.Min.X; x < bounds.Dx(); x++ {
		for y := bounds.Min.Y; y < bounds.Dy(); y++ {
			pixel := input.At(x, y)
			img.Set(x, y, pixel)
		}
	}
	dx, dy := input.Bounds().Dx(), input.Bounds().Dy()

	// Prepopulate multidimensional slices
	rErr := make([][]float32, dx)
	gErr := make([][]float32, dx)
	bErr := make([][]float32, dx)
	for x := 0; x < dx; x++ {
		rErr[x]	= make([]float32, dy)
		gErr[x]	= make([]float32, dy)
		bErr[x]	= make([]float32, dy)
		for y := 0; y < dy; y++ {
			rErr[x][y] = 0
			gErr[x][y] = 0
			bErr[x][y] = 0
		}
	}

	var qrr, qrg, qrb float32
	for x := 0; x < dx; x++ {
		for y := 0; y < dy; y++ {
			r32, g32, b32, a := img.At(x, y).RGBA()
			r, g, b := float32(uint8(r32)), float32(uint8(g32)), float32(uint8(b32))
			r -= rErr[x][y] * mul
			g -= gErr[x][y] * mul
			b -= bErr[x][y] * mul

			// Diffuse the error of each calculation to the neighboring pixels
			if r < 128 {
				qrr = -r
				r = 0
			} else {
				qrr = 255 - r
				r = 255
			}
			if g < 128 {
				qrg = -g
				g = 0
			} else {
				qrg = 255 - g
				g = 255
			}
			if b < 128 {
				qrb = -b
				b = 0
			} else {
				qrb = 255 - b
				b = 255
			}
			img.Set(x, y, color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)})

			// Diffuse error in two dimension
			ydim := len(dither.Filter) - 1
			xdim := len(dither.Filter[0]) / 2
			for xx := 0; xx < ydim + 1; xx++ {
				for yy := -xdim; yy <= xdim - 1; yy++ {
					if y + yy < 0 || dy <= y + yy || x + xx < 0 || dx <= x + xx {
						continue
					}
					// Adds the error of the previous pixel to the current pixel
					rErr[x+xx][y+yy] += qrr * dither.Filter[xx][yy + ydim]
					gErr[x+xx][y+yy] += qrg * dither.Filter[xx][yy + ydim]
					bErr[x+xx][y+yy] += qrb * dither.Filter[xx][yy + ydim]
				}
			}
		}
	}

	newimg := image.NewRGBA(bounds)
	palette := Quant(input, 16)
	draw.FloydSteinberg.Draw(newimg, input.Bounds(), palette, image.ZP)
	for x := 0; x < dx; x++ {
		for y := 0; y < dy; y++ {
			r1, g1, b1, _ := img.At(x, y).RGBA()
			//println(r1,g1,b1)
			r2, g2, b2, _ := palette.At(x, y).RGBA()
			er := r1 - r2
			eg := g1 - g2
			eb := b1 - b2
			//println(r2)

			for i := 0; i < len(dither.Filter); i++ {
				x1 := dither.Filter[i][1]
				y1 := dither.Filter[i][2]

				d := dither.Filter[i][0]
				//println(d)
				r1, g1, b1, _ = img.At(x + int(x1), int(y1)).RGBA()
				r1 = r1 + (er * uint32(d))
				g1 = g1 + (eg * uint32(d))
				b1 = b1 + (eb * uint32(d))
				//println(r1)
			}
			newimg.Set(x, y, color.RGBA{uint8(r1 >> 24), uint8(g1 >> 16), uint8(b2 >> 8), uint8(0xff)})
		}
	}
	return newimg
}
