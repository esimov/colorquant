package main

import (
	"image"
	_ "image/png"
	"image/jpeg"
	"log"
	"os"
	"github.com/esimov/colorquant"
	"flag"
	"fmt"
	"time"
	"errors"
	"path/filepath"
	"image/png"
	"image/color/palette"
)

type file struct {
	name string
}

// Command line flags
var (
	input		string
	output		string
	ditherMethod	string
	imageType	string
	noDither	bool
	compression 	int
	numColors	int
	commands 	flag.FlagSet
)

const helper = `
Usage of commands:
  -compression int
    	JPEG compression. (default 100)
  -dither string
    	Dithering method. (default "FloydSteinberg")
  -no-dither
    	Use image quantizer without dithering.
  -output string
    	Output directory. (default "output")
  -palette int
    	The number of palette colors. (default 256)
  -type string
    	Image type. Possible options .jpg, .png (default "jpg")
`

var dither map[string]colorquant.Dither = map[string]colorquant.Dither{
	"FloydSteinberg" : colorquant.Dither{
		[][]float32{
			[]float32{ 0.0, 0.0, 0.0, 7.0 / 48.0, 5.0 / 48.0 },
			[]float32{ 3.0 / 48.0, 5.0 / 48.0, 7.0 / 48.0, 5.0 / 48.0, 3.0 / 48.0 },
			[]float32{ 1.0 / 48.0, 3.0 / 48.0, 5.0 / 48.0, 3.0 / 48.0, 1.0 / 48.0 },
			//[]float32{ 5.0 / 48.0, 2.0 / 48.0, 1.0 / 48.0, 0.0, 0.0 },
		},
	},
	"Burkes" : colorquant.Dither{
		[][]float32{
			[]float32{ 0.0, 0.0, 0.0, 8.0 / 32.0, 4.0 / 32.0 },
			[]float32{ 2.0 / 32.0, 4.0 / 32.0, 8.0 / 32.0, 4.0 / 32.0, 2.0 / 32.0 },
			[]float32{ 0.0, 0.0, 0.0, 0.0, 0.0 },
			[]float32{ 4.0 / 32.0, 8.0 / 32.0, 0.0, 0.0, 0.0 },
		},
	},
	"Stucki" : colorquant.Dither{
		[][]float32{
			[]float32{ 0.0, 0.0, 0.0, 8.0 / 42.0, 4.0 / 42.0 },
			[]float32{ 2.0 / 42.0, 4.0 / 42.0, 8.0 / 42.0, 4.0 / 42.0, 2.0 / 42.0 },
			[]float32{ 1.0 / 42.0, 2.0 / 42.0, 4.0 / 42.0, 2.0 / 42.0, 1.0 / 42.0 },
		},
	},
	"Sierra-3" : colorquant.Dither{
		[][]float32{
			[]float32{ 0.0, 0.0, 0.0, 5.0 / 32.0, 3.0 / 32.0 },
			[]float32{ 2.0 / 32.0, 4.0 / 32.0, 5.0 / 32.0, 4.0 / 32.0, 2.0 / 32.0 },
			[]float32{ 0.0, 2.0 / 32.0, 3.0 / 32.0, 2.0 / 32.0, 0.0 },
		},
	},
}

// Open image
func (file *file) Open() (image.Image, error) {
	f, err := os.Open(file.name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	return img, err
}

// Save the generated image
func (file *file) Quantify(src image.Image, output string) (image.Image, error) {
	var err error
	var quant image.Image

	dst := image.NewPaletted(image.Rect(0, 0, src.Bounds().Dx(), src.Bounds().Dy()), palette.WebSafe)
	if noDither {
		quant = colorquant.NoDither.Quantize(src, dst, numColors, false)
	} else {
		if _, ok := dither[ditherMethod]; !ok {
			log.Fatal("\nInvalid dithering method!")
			return nil, err
		}

		ditherer := dither[ditherMethod]
		quant = ditherer.Quantize(src, dst, numColors, true)
	}

	fq, err := os.Create(output)
	if err != nil {
		return nil, err
	}
	defer fq.Close()

	switch imageType {
	case "jpg" :
		if err = jpeg.Encode(fq, quant, &jpeg.Options{compression}); err != nil {
			log.Fatal(err)
			return nil, err
		}
	case "png" :
		if err = png.Encode(fq, quant); err != nil {
			log.Fatal(err)
			return nil, err
		}
	}
	return quant, nil
}

func main() {
	commands = *flag.NewFlagSet("commands", flag.ExitOnError)
	commands.StringVar(&output, "output", "output", "Output directory.")
	commands.StringVar(&ditherMethod, "dither", "FloydSteinberg", "Dithering method.")
	commands.StringVar(&imageType, "type", "jpg", "Image type. Possible options .jpg, .png")
	commands.BoolVar(&noDither, "no-dither", false, "Use image quantizer without dithering.")
	commands.IntVar(&compression, "compression", 100, "JPEG compression.")
	commands.IntVar(&numColors, "palette", 256, "The number of palette colors.")

	if len(os.Args) <= 1 || (os.Args[1] == "--help" || os.Args[1] == "-h") {
		fmt.Println(errors.New(helper))
		os.Exit(1)
	}

	// Parse flags before to use them
	commands.Parse(os.Args[2:])

	// Channel used to signal the completion event
	done := make(chan struct{})
	input := &file{name: string(os.Args[1])}
	img, _ := input.Open()

	if commands.Parsed() {
		cwd, err := filepath.Abs(filepath.Dir(input.name))
		if err != nil {
			log.Fatal(err)
		}
		newDir := filepath.Dir(cwd) + "/" + output

		os.Mkdir(newDir, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		os.Chdir(newDir)

		fmt.Print("Rendering image...")
		now := time.Now()
		progress(done)

		// Process the image
		func(input *file, done chan struct{}) {
			switch imageType {
			case "jpg" :
				if noDither {
					input.Quantify(img, "output.jpg")
				} else {
					input.Quantify(img, ditherMethod + ".jpg")
				}
			case "png" :
				if noDither {
					input.Quantify(img, "output.png")
				} else {
					input.Quantify(img, ditherMethod + ".png")
				}
			}
			done <- struct{}{}
		}(input, done)

		since := time.Since(now)
		fmt.Println("\nDone✓")
		fmt.Printf("Rendered in: %.2fs\n", since.Seconds())
	}
}

// Function to visualize the rendering progress
func progress(done chan struct{}) {
	ticker := time.NewTicker(time.Millisecond * 200)

	go func() {
		for {
			select {
			case <-ticker.C:
				fmt.Print(".")
			case <-done:
				ticker.Stop()
			}
		}
	}()
}