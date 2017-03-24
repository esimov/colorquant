package main

import (
	"image/png"
	"log"
	"os"
	//"image"
	//"image/color/palette"
)

func main() {
	var dither Dither = Dither{
		"FloydSteinberg",
		settings{
			[][]float32{
				[]float32{ 0.0, 0.0, 0.0, 7.0 / 48.0, 5.0 / 48.0 },
				[]float32{ 3.0 / 48.0, 5.0 / 48.0, 7.0 / 48.0, 5.0 / 48.0, 3.0 / 48.0 },
				[]float32{ 1.0 / 48.0, 3.0 / 48.0, 5.0 / 48.0, 3.0 / 48.0, 1.0 / 48.0 },
			},
		},
	}
	f, err := os.Open("Quantum_frog.png")
	if err != nil {
		log.Fatal(err)
	}
	img, err := png.Decode(f)
	if ec := f.Close(); err != nil {
		log.Fatal(err)
	} else if ec != nil {
		log.Fatal(ec)
	}
	fq, err := os.Create("frog16.png")
	if err != nil {
		log.Fatal(err)
	}
	quant := dither.Process(img, 1.18)
	//p8 := image.NewPaletted(image.Rect(0, 0, quant.Bounds().Dx(), quant.Bounds().Dy()), palette.WebSafe)
	//FloydSteinberg.Draw(p8, p8.Bounds(), img, image.ZP)
	if err = png.Encode(fq, quant); err != nil {
		log.Fatal(err)
	}
}