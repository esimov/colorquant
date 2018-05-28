package colorquant

import (
	"image"
	"image/color"
	"math"
	"image/draw"
)

// Dither is a two dimensional slice for storing different dithering methods.
type Dither struct {
	Filter [][]float32
}

// NoDither is used to call the default quantize method without applying dithering.
var NoDither Quantizer = Dither{}

// Empty check if dither struct is empty. If empty this means we are not using any dithering method.
func (dither Dither) Empty() bool {
	if len(dither.Filter) > 0 {
		return false
	}
	return true
}

// Quantize takes as parameter the original image and returns the processed image with or without dithering applied.
func (dither Dither) Quantize(src image.Image, dst draw.Image, nq int, useDither bool, useQuantizer bool) image.Image {
	var quant image.Image
	var er, eg, eb, ea int32
	dx, dy := src.Bounds().Dx(), src.Bounds().Dy()

	// Import the quantized image and specify the quantization level
	if useQuantizer {
		quant = Quant{}.Quantize(src, nq)
	}

	// Prepopulate a multidimensional slice. We will use this to store the quantization level.
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

	// If dst is an *image.Paletted, we have a fast path for dst.Set and dst.At.
	palette, pix, stride := [][4]int32(nil), []byte(nil), 0
	if p, ok := dst.(*image.Paletted); ok {
		palette = make([][4]int32, len(p.Palette))
		for i, col := range p.Palette {
			r, g, b, a := col.RGBA()
			palette[i][0] = int32(r)
			palette[i][1] = int32(g)
			palette[i][2] = int32(b)
			palette[i][3] = int32(a)
		}
		pix, stride = p.Pix[p.PixOffset(0, 0):], p.Stride
	}

	out := color.RGBA{A:0xff}

	// Loop trough the image and process each pixel individually.
	for x := 0; x != dx; x++ {
		for y := 0; y != dy; y++ {
			if !useQuantizer {
				r1, g1, b1, a1 := src.At(x, y).RGBA()
				// er, eg and eb are the pixel's R,G,B values
				er, eg, eb, ea = int32(r1), int32(g1), int32(b1), int32(a1)

				if useDither {
					er = clamp(er + int32(rErr[x][y] * 1.12))
					eg = clamp(eg + int32(gErr[x][y] * 1.12))
					eb = clamp(eb + int32(bErr[x][y] * 1.12))
				}

				// Find the closest palette color in Euclidean R,G,B,A space:
				// the one that minimizes sum-squared-difference.
				bestIndex, bestSum := 0, uint32(1<<32-1)
				for index, p := range palette {
					sum := sqDiff(er, p[0]) + sqDiff(eg, p[1]) + sqDiff(eb, p[2]) + sqDiff(ea, p[3])
					if sum < bestSum {
						bestIndex, bestSum = index, sum
						if sum == 0 {
							break
						}
					}
				}
				pix[y*stride+x] = byte(bestIndex)

				if !useDither {
					continue
				}
				er -= palette[bestIndex][0]
				eg -= palette[bestIndex][1]
				eb -= palette[bestIndex][2]
				ea -= palette[bestIndex][3]

				// Diffuse error in two dimension
				ydim := len(dither.Filter) - 1
				xdim := len(dither.Filter[0]) / 2 // split the X dimension in two halves
				for xx := 0; xx < ydim + 1; xx++ {
					for yy := -xdim; yy <= xdim - 1; yy++ {
						if y + yy < 0 || dy <= y + yy || x + xx < 0 || dx <= x + xx {
							continue
						}
						// Propagate the quantization error
						rErr[x+xx][y+yy] += float32(er) * dither.Filter[xx][yy + ydim]
						gErr[x+xx][y+yy] += float32(eg) * dither.Filter[xx][yy + ydim]
						bErr[x+xx][y+yy] += float32(eb) * dither.Filter[xx][yy + ydim]
					}
				}
			} else {
				// Find the closest pixel color between the paletted image and the original image.
				r1, g1, b1, a1 := findClosestColor(quant.(*image.Paletted), src.At(x, y)).RGBA()
				// er, eg and eb are the pixel's R,G,B values
				er, eg, eb, ea := int32(r1), int32(g1), int32(b1), int32(a1)

				if useDither && !dither.Empty() {
					er = clamp(er + int32(rErr[x][y] * 1.12))
					eg = clamp(eg + int32(gErr[x][y] * 1.12))
					eb = clamp(eb + int32(bErr[x][y] * 1.12))
				}
				out.R = uint8(er>>8)
				out.G = uint8(eg>>8)
				out.B = uint8(eb>>8)
				out.A = uint8(ea>>8)

				// Set the resulting pixel colors in the destination image.
				dst.Set(x, y, &out)

				if !useDither || dither.Empty() {
					continue
				}
				sr, sg, sb, sa := dst.At(x, y).RGBA()
				er -= int32(sr)
				eg -= int32(sg)
				eb -= int32(sb)
				ea -= int32(sa)

				// Diffuse error in two dimension
				ydim := len(dither.Filter) - 1
				xdim := len(dither.Filter[0]) / 2 // split the X dimension in two halves
				for xx := 0; xx < ydim + 1; xx++ {
					for yy := -xdim; yy <= xdim - 1; yy++ {
						if y + yy < 0 || dy <= y + yy || x + xx < 0 || dx <= x + xx {
							continue
						}
						// Propagate the quantization error
						rErr[x+xx][y+yy] += float32(er) * dither.Filter[xx][yy + ydim]
						gErr[x+xx][y+yy] += float32(eg) * dither.Filter[xx][yy + ydim]
						bErr[x+xx][y+yy] += float32(eb) * dither.Filter[xx][yy + ydim]
					}
				}
			}
		}
	}
	return dst
}

// Returns the index of the palette color closest to source color in Euclidean R,G,B,A space.
func findClosestColor(palette *image.Paletted, src color.Color) color.Color {
	var pr, pg, pb, pa float64
	if len(palette.Palette) == 0 {
		return nil
	}
	cr, cg, cb, ca := src.RGBA()
	idx, bestSum := 0, uint32(1<<32-1)

	// Rec. 709 (sRGB) luma coef.
	pr = .2126
	pg = .7152
	pb = .0722
	pa = 1.0

	// Get the square root of euclidean distance.
	euclMax := math.Sqrt(pr * 255 * 255 + pg * 255 * 255 + pb * 255 * 255)
	for index, v := range palette.Palette {
		vr, vg, vb, va := v.RGBA()
		// Get the color distance.
		sum := math.Sqrt(
			pr * sqDiffFloat(float64(cr), float64(vr)) +
			pg * sqDiffFloat(float64(cg), float64(vg)) +
			pb * sqDiffFloat(float64(cb), float64(vb)) +
			pa * sqDiffFloat(float64(ca), float64(va))) / euclMax
		// Get the min value.
		if uint32(sum) < bestSum {
			idx, bestSum = index, uint32(sum)
			if sum == 0 {
				break
			}
		}
	}
	// Return the colors most closely identical to the original pixel colors.
	return palette.Palette[idx]
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

// sqDiffFloat returns the squared-difference of X and Y.
func sqDiffFloat(x, y float64) float64 {
	var d float64

	if x > y {
		d = float64(x - y)
	} else {
		d = float64(x - y)
	}
	return float64(d * d)
}

// sqDiff returns the squared-difference of X and Y.
func sqDiff(x, y int32) uint32 {
	var d uint32
	if x > y {
		d = uint32(x - y)
	} else {
		d = uint32(y - x)
	}
	return (d * d) >> 2
}


// min returns the smallest number between two numbers.
func min(x, y uint32) uint32 {
	if x < y {
		return x
	}
	return y
}

// max returns the biggest number between two numbers.
func max(x, y uint32) uint32 {
	if x > y {
		return x
	}
	return y
}