package colorquant

import (
	"testing"
	"image"
	"image/color"
	"image/color/palette"
)

func Test_IsDitherUsed(t *testing.T) {
	ditherer := Dither{
		[][]float32{
			[]float32{0.0, 0.0, 0.0, 7.0 / 48.0, 5.0 / 48.0 },
			[]float32{3.0 / 48.0, 5.0 / 48.0, 7.0 / 48.0, 5.0 / 48.0, 3.0 / 48.0 },
			[]float32{1.0 / 48.0, 3.0 / 48.0, 5.0 / 48.0, 3.0 / 48.0, 1.0 / 48.0 },
		},
	}

	if ditherer.Empty() {
		t.Error("Dither method is not used!")
	}
}

func Test_DitherMethod(t *testing.T) {
	validDitherMethods := map[string]int{
		"FloydSteinberg" : 1,
		"Burkes" : 2,
		"Stucki" : 3,
		"Atkinson" : 4,
		"Sierra-3" : 5,
		"Sierra-2" : 6,
		"Sierra-Lite" : 7,
	}
	testMethod := "FloydSteinberg"
	if _, ok := validDitherMethods[testMethod]; !ok {
		t.Error("Invalid dithering method!")
	}
}

func Test_QuantizationLevel(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	quantLevel := 10
	Quant{}.Quantize(img, quantLevel)

	if quantLevel < 2 {
		t.Error("Image quantization level should be > 1!")
	}
}

func Test_OutputColor(t *testing.T) {
	img := image.NewRGBA64(image.Rect(0, 0, 1, 2))
	img.SetRGBA64(0, 0, color.RGBA64{0xff, 0x00, 0x00, 0x00})

	r1, g1, b1, a1 := img.At(0, 0).RGBA()
	er, eg, eb, ea := int32(r1), int32(g1), int32(b1), int32(a1)

	if er > 0xff || eg > 0xff || eb > 0xff || ea > 0xff {
		t.Errorf("The expected color should be less than 0xff, got R:%d G:%d B:%d !", er, eg, eb)
	}
}

func Test_PalettedImage(t *testing.T) {
	ditherer := Dither{
		[][]float32{
			[]float32{0.0, 0.0, 0.0, 7.0 / 48.0, 5.0 / 48.0 },
			[]float32{3.0 / 48.0, 5.0 / 48.0, 7.0 / 48.0, 5.0 / 48.0, 3.0 / 48.0 },
			[]float32{1.0 / 48.0, 3.0 / 48.0, 5.0 / 48.0, 3.0 / 48.0, 1.0 / 48.0 },
		},
	}
	src := image.NewRGBA(image.Rect(0, 0, 10, 10))
	dst := image.NewPaletted(image.Rect(0, 0, 10, 10), palette.WebSafe)
	ditherer.Quantize(src, dst, 16, true, true)

	palette := make([][4]int32, len(dst.Palette))
	for i, col := range dst.Palette {
		r, g, b, a := col.RGBA()
		palette[i][0] = int32(r)
		palette[i][1] = int32(g)
		palette[i][2] = int32(b)
		palette[i][3] = int32(a)
	}

	if palette == nil {
		t.Error("Destination image should be of paletted type!")
	}
}