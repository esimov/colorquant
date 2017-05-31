# colorquant

<strong>Colorquant</strong> is an image / color quantization library written in Go. It can be considered as a replacement for the core image <a href="https://golang.org/pkg/image/draw/">draw</a> method, for various reasons (see below).

### The purpose

The purpose of color quantization is to reduce the color palette of an image to a fraction of it's initial colors (usually 256), but to preserve it's representative colors and to elliminate visual artifacts at the same time. Even with the best set of 256 colors, there are many images that look bad. They have visible contouring in regions where the color changes slowly. 

To create a smoother transition between colors and to wash out the edges various dithering methods can be plugged in.

### Implementation

The implementation is mainly based on the article from <a href="http://www.leptonica.com/color-quantization.html">Leptonica</a>.

The reason why I opted for a custom quantization and dithering algorithm are twofold:
* First, even the core draw method does provide an error quantization algorithm, it does not implement the support for quantization level (to how many colors we wish to reduce the original image).
* Second, the dithering method is based exclusively on Floyd-Steinberg dithering method, but there are other dithering algorithm, which can be used (ex. Burkes, Stucki, Atkinson, Sierra etc.).

### Installation

`go get -u github.com/esimov/colorquant`

### Running

The library provides a CLI method to generate the quantified images. Type `go run cli.go --help` to get the supported commands.

```
Usage of commands:
  -compression int
    	JPEG compression. (default 100)
  -ditherer string
    	Dithering method. (default "FloydSteinberg")
  -no-dither
    	Use image quantizer without dithering.
  -output string
    	Output directory. (default "output")
  -palette int
    	The number of palette colors. (default 256)
  -type string
    	Image type. Possible options .jpg, .png (default "jpg")

```
The generated images will be exported into the `output` folder. By default the <i><strong>Floyd-Steinberg</strong></i> dithering method is applied, but if you whish to <strong>not</strong> use any dithering algorithm use the `--no-dither` flag.

### Usage

##### ➤ Without dither
This is main method to generate a non-dithered quantified image:

```go
colorquant.NoDither.Quantize(src, dst, numColors, false, true)
```
where the last paremeter means either to use the library quantization algorithm (if the parameter is <i>true</i>), otherwise use the quantization level provided by the paletted image (if the paramater is <i>false</i>).

##### ➤ With dither
The same, but this time using a dithering method:

```go
ditherer.Quantize(src, dst, numColors, true, true)
```
where ditherer is struct with the form of:

```go
"FloydSteinberg" : colorquant.Dither{
	[][]float32{
		[]float32{ 0.0, 0.0, 0.0, 7.0 / 48.0, 5.0 / 48.0 },
		[]float32{ 3.0 / 48.0, 5.0 / 48.0, 7.0 / 48.0, 5.0 / 48.0, 3.0 / 48.0 },
		[]float32{ 1.0 / 48.0, 3.0 / 48.0, 5.0 / 48.0, 3.0 / 48.0, 1.0 / 48.0 },
	},
},
```
### Examples

All the examples below are generated using *Floyd-Steinberg* dithering method with the following command line as an example:

`go run cli.go ../input/treefrog.jpg -compression 100 -ditherer FloydSteinberg -palette 128`

| Number of colors | Without dither | With Dither
|:--:|:--:|:--:|
| *128* | <img src="https://cloud.githubusercontent.com/assets/883386/26618632/b0e865b2-45e3-11e7-9312-c66f5d690312.jpg"> | <img src="https://cloud.githubusercontent.com/assets/883386/26618639/b623c77e-45e3-11e7-8900-2850bb8a0a9d.jpg"> |
| *256* | <img src="https://cloud.githubusercontent.com/assets/883386/26618480/2f9b1158-45e3-11e7-9851-742a21e1f8af.jpg"> | <img src="https://cloud.githubusercontent.com/assets/883386/26618461/229eb626-45e3-11e7-8fa4-9eaeeeb55712.jpg"> | 
| *512* | <img src="https://cloud.githubusercontent.com/assets/883386/26617946/a860de4a-45e0-11e7-8c19-1e6fa3fd6e19.jpg" > | <img src="https://cloud.githubusercontent.com/assets/883386/26618118/9cce2046-45e1-11e7-8945-4bea76459741.jpg"> | 
| *1024* | <img src="https://cloud.githubusercontent.com/assets/883386/26619097/a27027de-45e5-11e7-83b3-cb5b9e7d7079.jpg" > | <img src="https://cloud.githubusercontent.com/assets/883386/26619106/a8ec32b0-45e5-11e7-9642-c0f74a384544.jpg"> | 

## License

This software is distributed under the MIT license found in the LICENSE file.
