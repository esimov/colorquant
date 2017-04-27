package colorquant

import (
	"image"
	"image/color"
	"fmt"
	"math"
)

// Struct containing a two dimensional slice for storing different dithering methods.
type Dither struct {
	Filter [][]float32
}

// Used to call the default quantize method without applying dithering.
var Default Quantizer = Dither{}

// The Quantize method takes as parameter the original image and returns the processed image with dithering.
func (dither Dither) Quantize(input image.Image, nq int) image.Image {
	res := ditherImage(input, nq, dither)
	return res
}

// Private function to call error quantization method (dithering) over an image.
func ditherImage(input image.Image, nq int, dither Dither) image.Image {
	var quant image.Image
	var r4, g4, b4, a4 float32

	// Create a new empty RGBA image. This will be the destination of the new processed image.
	output := image.NewRGBA(image.Rect(0, 0, input.Bounds().Dx(), input.Bounds().Dy()))
	dx, dy := input.Bounds().Dx(), input.Bounds().Dy()

	quantErrorCurr := make([][4]int32, dx+2)
	quantErrorNext := make([][4]int32, dx+2)

	// Import the quantized image.
	// The first parameter is the destination image. The second parameter is the quantization level (the number of colors).
	quant = Quant{}.Quantize(input, nq)

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

	out := color.RGBA{A:0xff}
	// Loop trough the image and process each pixel individually.
	for x := 0; x != dx; x++ {
		for y := 0; y != dy; y++ {
			r1, g1, b1, a1 := output.At(x, y).RGBA()
			// Find the closest pixel color between the paletted image and the original image.
			r2, g2, b2, a2 := findPaletteColor(quant.(*image.Paletted), input.At(x, y)).RGBA()

			// Set the pixel results in each color channel separately.
			// We need to right shift the resulting colors with 8 bits.
			out.R = uint8(r2>>8)
			out.G = uint8(g2>>8)
			out.B = uint8(b2>>8)
			out.A = uint8(a2>>8)

			// Set the resulting pixel colors in the destination image.
			output.Set(x, y, &out)

			// Take the color difference between the paletted image and the original image.
			er := uint8(r1>>8) - uint8(r2>>8)
			eg := uint8(g1>>8) - uint8(g2>>8)
			eb := uint8(b1>>8) - uint8(b2>>8)
			ea := uint8(a1>>8) - uint8(a2>>8)

			for i := 0; i != len(dither.Filter); i++ {
				y1 := dither.Filter[i][1] // Y value of the dithering method (between -1, 1)
				x1 := dither.Filter[i][2] // X value of the dithering method (between -1, 1)

				// Get the X and Y value from the original image and sum up with the dithering level
				var xt int = int(x1) + x
				var yt int = int(y1) + y
				if xt >= 0 && xt < dx && yt >= 0 && yt < dy {
					d := dither.Filter[i][0]
					r3, g3, b3, a3 := output.At(xt, yt).RGBA()

					// Quantize the resulting image with the error level multiplied with the dithering value.
					r4 = float32(uint8(r3)) + (float32(er) * d)
					g4 = float32(uint8(g3)) + (float32(eg) * d)
					b4 = float32(uint8(b3)) + (float32(eb) * d)
					a4 = float32(uint8(a3)) + (float32(ea) * d)
					if r4 > 255 {
						fmt.Println("R3: ", float32(uint8(r3)))
						fmt.Println("ER: ", float32(er) * d)
						fmt.Println("Final: ", r4)
						fmt.Println("==========================")
					}
					r := max(0, min(255, int(r4)))
					g := max(0, min(255, int(g4)))
					b := max(0, min(255, int(b4)))
					a := max(0, min(255, int(a4)))

					// Set the final colors in the destination image.
					output.Set(xt, yt, color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)})
				}
			}
		}
		quantErrorCurr, quantErrorNext = quantErrorNext, quantErrorCurr
		for i := range quantErrorNext {
			quantErrorNext[i] = [4]int32{}
		}
	}
	return output
}

// Returns the index of the palette color closest to quantizedImg in Euclidean R,G,B,A space.
func findPaletteColor(palette *image.Paletted, quantized color.Color) color.Color {
	var pr, pg, pb, pa float64
	if len(palette.Palette) == 0 {
		return nil
	}
	cr, cg, cb, ca := quantized.RGBA()
	idx, min := 0, uint32(1<<32-1)
	// Some arbitrary values.
	pr = .2126
	pg = .7152
	pb = .0722
	pa = 1.0

	// Get the square root of euclidean distance.
	euclMax := math.Sqrt(pr * 255 + pg * 255 + pb * 255)
	for index, v := range palette.Palette {
		vr, vg, vb, va := v.RGBA()
		// Get the color distance.
		dist := math.Sqrt(
			pr * sqDiff(float64(cr), float64(vr)) +
			pg * sqDiff(float64(cg), float64(vg)) +
			pb * sqDiff(float64(cb), float64(vb)) +
			pa * sqDiff(float64(ca), float64(va))) / euclMax
		// Get the min value.
		if uint32(dist) < min {
			idx, min = index, uint32(dist)
		}
	}
	// Return the colors most closely identical to the original pixel colors.
	return palette.Palette[idx]
}

// Clamp clamps i to the interval [0, 0xffff].
func clamp(i int32) int32 {
	if i < 0 {
		return 0
	}
	if i > 0xffff {
		return 0xffff
	}
	return i
}

// Returns the squared-difference of X and Y.
func sqDiff(x, y float64) float64 {
	var d float64

	if x > y {
		d = float64(x - y)
	} else {
		d = float64(y - x)
	}
	return float64(uint32(d * d) >> 2)
}

// Returns the smallest number between two numbers.
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// Returns the biggest number between two numbers.
func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}