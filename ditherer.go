package main

import (
	"image"
	"image/color"
	"image/color/palette"
)

type settings struct {
	filter [][]float32
}

type Dither struct {
	method string
	settings
}

func (dither Dither) Process(input image.Image, mul float32) image.Image {
	dst := image.NewPaletted(image.Rect(0, 0, input.Bounds().Dx(), input.Bounds().Dy()), palette.WebSafe)
	dx, dy := input.Bounds().Dx(), input.Bounds().Dy()

	quantErrorCurr := make([][4]int32, dx+2)
	quantErrorNext := make([][4]int32, dx+2)
	quantizedImg := Quant(input, 256)

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
	out := color.RGBA64{A: 0xffff}
	for x := 0; x != dx; x++ {
		for y := 0; y != dy; y++ {
			sr, sg, sb, sa := quantizedImg.At(x, y).RGBA()
			er, eg, eb, ea := int32(sr), int32(sg), int32(sb), int32(sa)
			er = clamp(er + int32(rErr[x][y] * mul))
			eg = clamp(eg + int32(gErr[x][y] * mul))
			eb = clamp(eb + int32(bErr[x][y] * mul))

			out.R = uint16(er)
			out.G = uint16(eg)
			out.B = uint16(eb)
			out.A = uint16(ea)

			// The third argument is &out instead of out (and out is
			// declared outside of the inner loop) to avoid the implicit
			// conversion to color.Color here allocating memory in the
			// inner loop if sizeof(color.RGBA64) > sizeof(uintptr).
			dst.Set(x, y, &out)

			sr, sg, sb, sa = dst.At(x, y).RGBA()
			//r1, g1, b1, a1 := findPaletteColor(dst, quantizedImg.At(x, y)).RGBA()
			er -= int32(sr)
			eg -= int32(sg)
			eb -= int32(sb)
			ea -= int32(sa)

			// Diffuse error in two dimension
			ydim := len(dither.filter) - 1
			xdim := len(dither.filter[0]) / 2
			for xx := 0; xx < ydim + 1; xx++ {
				for yy := -xdim; yy <= xdim - 1; yy++ {
					if y + yy < 0 || dy <= y + yy || x + xx < 0 || dx <= x + xx {
						continue
					}
					// Propagate the quantization error
					rErr[x+xx][y+yy] += float32(er) * dither.filter[xx][yy + ydim]
					gErr[x+xx][y+yy] += float32(eg) * dither.filter[xx][yy + ydim]
					bErr[x+xx][y+yy] += float32(eb) * dither.filter[xx][yy + ydim]
				}
			}
		}
		quantErrorCurr, quantErrorNext = quantErrorNext, quantErrorCurr
		for i := range quantErrorNext {
			quantErrorNext[i] = [4]int32{}
		}
	}
	return dst
}

// Returns the index of the palette color closest to quantizedImg in Euclidean R,G,B,A space.
func findPaletteColor(palette *image.Paletted, quantizedImg color.Color) color.Color {
	if len(palette.Palette) == 0 {
		return nil
	}
	cr, cg, cb, ca := quantizedImg.RGBA()
	ret, bestSum := 0, uint32(1<<32-1)
	for index, v := range palette.Palette {
		vr, vg, vb, va := v.RGBA()
		sum := sqDiff(cr, vr) + sqDiff(cg, vg) + sqDiff(cb, vb) + sqDiff(ca, va)
		if sum < bestSum {
			ret, bestSum = index, sum
		}
	}
	return palette.Palette[ret]
}

// clamp clamps i to the interval [0, 0xffff].
func clamp(i int32) int32 {
	if i < 0 {
		return 0
	}
	if i > 0xffff {
		return 0xffff
	}
	return i
}

// Return the squared-difference of x and y, shifted by 2 so that adding four of those won't overflow a uint32.
func sqDiff(x, y uint32) uint32 {
	var d uint32
	if x > y {
		d = uint32(x - y)
	} else {
		d = uint32(y - x)
	}
	return (d * d) >> 2
}