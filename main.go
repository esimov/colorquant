package main

import (
	"image"
	_ "image/png"
	_ "image/jpeg"
	"log"
	"os"
	//"image"
	//"image/color/palette"
	"image/jpeg"
)

func main() {
	var dither map[string]Dither = map[string]Dither{
		"FloydSteinberg" : Dither{
			[][]float32{
				[]float32{ 7.0 / 16.0, 1.0, 0.0 },
				[]float32{ 3.0 / 16.0, -1.0, 1.0 },
				[]float32{ 5.0 / 16.0, 0.0, 1.0 },
				[]float32{ 1.0 / 16.0, 1.0, 1.0 },
				[]float32{ 3.0 / 16.0, 1.0, -1.0 },
			},
		},
		"Burkes" : Dither{
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
	f, err := os.Open("david.jpg")
	if err != nil {
		log.Fatal(err)
	}
	img, _, err := image.Decode(f)
	if ec := f.Close(); err != nil {
		log.Fatal(err)
	} else if ec != nil {
		log.Fatal(ec)
	}
	fq, err := os.Create("frog16.jpg")
	if err != nil {
		log.Fatal(err)
	}
	quant, err := dither["FloydSteinberg"].Process(img)
	if err != nil {
		log.Fatal(err)
	}
	//FloydSteinberg.Draw(p8, p8.Bounds(), img, image.ZP)

	if err = jpeg.Encode(fq, quant, &jpeg.Options{100}); err != nil {
		log.Fatal(err)
	}
}