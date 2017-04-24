package main

import (
	"image"
	_ "image/png"
	"image/jpeg"
	"log"
	"os"
	"github.com/esimov/colorquant"
)

func main() {
	var dither map[string]colorquant.Dither = map[string]colorquant.Dither{
		"FloydSteinberg" : colorquant.Dither{
			[][]float32{
				[]float32{ 7.0 / 32.0, 1.0, 0.0 },
				[]float32{ 3.0 / 32.0, -1.0, 1.0 },
				[]float32{ 5.0 / 32.0, 0.0, 1.0 },
				[]float32{ 1.0 / 32.0, 1.0, 1.0 },
				[]float32{ 3.0 / 32.0, 1.0, -1.0 },
			},
		},
		"Burkes" : colorquant.Dither{
			[][]float32{
				[]float32{ 8.0 / 32.0, 1.0, 0.0 },
				[]float32{ 4.0 / 32.0, 2.0, 0.0 },
				[]float32{ 2.0 / 32.0, -2.0, 1.0 },
				[]float32{ 4.0 / 32.0, -1.0, 1.0 },
				[]float32{ 8.0 / 32.0, 0.0, 1.0 },
				[]float32{ 4.0 / 32.0, 1.0, 1.0 },
				[]float32{ 2.0 / 32.0, 2.0, 1.0 },
				[]float32{ 4.0 / 32.0, 1.0, -1.0 },
			},
		},
	}
	f, err := os.Open("../input/treefrog.jpg")
	if err != nil {
		log.Fatal(err)
	}
	img, _, err := image.Decode(f)
	if ec := f.Close(); err != nil {
		log.Fatal(err)
	} else if ec != nil {
		log.Fatal(ec)
	}
	fq, err := os.Create("output.jpg")
	if err != nil {
		log.Fatal(err)
	}
	floydSteinberg := dither["FloydSteinberg"]
	quant := floydSteinberg.Quantize(img, 256)
	if err != nil {
		log.Fatal(err)
	}
	//FloydSteinberg.Draw(p8, p8.Bounds(), img, image.ZP)

	if err = jpeg.Encode(fq, quant, &jpeg.Options{100}); err != nil {
		log.Fatal(err)
	}
}