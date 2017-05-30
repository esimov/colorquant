package colorquant

import (
	"testing"
	"image"
	"image/color/palette"
	"image/color"
	"math/rand"
)

func TestQuant_Paletted(t *testing.T) {
	img := image.NewPaletted(image.Rect(0, 0, 10, 10), palette.WebSafe)

	res := Quant{}.Quantize(img, 10)
	if p, ok := res.(*image.Paletted); ok {
		palette := make([][4]int32, len(p.Palette))
		for i, col := range p.Palette {
			r, g, b, a := col.RGBA()
			palette[i][0] = int32(r)
			palette[i][1] = int32(g)
			palette[i][2] = int32(b)
			palette[i][3] = int32(a)
		}

		if palette == nil {
			t.Errorf("The expected image should be a paletted image!")
		}
	}
}

func TestQuant_Median(t *testing.T) {
	img := image.NewPaletted(image.Rect(0, 0, 10, 10), palette.WebSafe)
	for i := 0; i < img.Bounds().Dx(); i++ {
		for j := 0; j < img.Bounds().Dy(); j ++ {
			img.Set(i, j, color.RGBA{0xff, 0, 0, 0})
		}
	}
	qz := newQuantizer(img, 2)
	qz.cluster()

	cls := &cluster{
		[]point{
			{0, 0},
			{1, 0},
			{0, 1},
			{1, 1},
		}, 1, 1,
	}
	res := qz.Median(cls)
	if res != 0 {
		t.Errorf("The expected result should be 0, got %d", res)
	}
}

func TestQuant_Level (t *testing.T) {
	quantLevel := 10
	reds := []uint32{}
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	for i := 0; i < img.Bounds().Dx(); i++ {
		for j := 0; j < img.Bounds().Dy(); j ++ {
			col := rand.Intn(255)
			img.Set(i, j, color.RGBA{uint8(col), 0, 0, 0})
		}
	}

	res := Quant{}.Quantize(img, quantLevel)

	if p, ok := res.(*image.Paletted); ok {
		for _, col := range p.Palette {
			r, _, _, _ := col.RGBA()
			reds = append(reds, r)
		}

		if len(reds) != quantLevel {
			t.Errorf("The quantization level should be %d, got %d", quantLevel, len(reds))
		}
	}
}