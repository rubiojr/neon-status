package main

// Based on the code from Rémy Mathieu:
//
// https://remy.io/blog/bloom-effect-in-go/

import (
	"bufio"
	"flag"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/anthonynsimon/bild/blur"
	"github.com/anthonynsimon/bild/effect"
	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"github.com/goki/freetype/truetype"
)

func main() {
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Println("Please specify source file.\n")
		flag.Usage()
		os.Exit(1)
	}

	input := flag.Args()[0]

	var fb []byte
	var err error

	if font == "" {
		fb, err = ioutil.ReadFile("fonts/Sportrop.ttf")
	} else {
		fb, err = ioutil.ReadFile(font)
	}
	if err != nil {
		fmt.Println("Error reading font file.")
		os.Exit(1)
	}

	font, err := truetype.Parse(fb)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Open(input)
	if err != nil {
		fmt.Println("Error opening input file")
		os.Exit(1)
	}

	face := truetype.NewFace(font, &truetype.Options{Size: 48})
	dc := gg.NewContext(canvasWidth, canvasHeight)
	dc.SetLineWidth(3.0)
	//dc.SetRGB255(147, 112, 219)
	//dc.SetRGB255(178, 0, 255)
	//dc.SetRGB255(255, 0, 0)
	parsedRGB := parseColor(rgb)
	dc.SetRGB255(parsedRGB[0], parsedRGB[1], parsedRGB[2])
	dc.SetFontFace(face)

	scanner := bufio.NewScanner(f)
	count := topMargin + 32
	for scanner.Scan() {
		dc.DrawString(scanner.Text(), float64(leftMargin), float64(topMargin+count))
		count += 32
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	dc.Stroke()

	original := dc.Image()

	bloomed := Bloom(dc.Image())

	dc = gg.NewContext(canvasWidth, canvasHeight)
	dc.SetRGB255(0, 0, 0)
	dc.DrawRectangle(0, 0, float64(canvasWidth), float64(canvasHeight))
	dc.Fill()

	dc.DrawImage(bloomed, 0, 0)

	dc.DrawImage(original, 10, 10)

	dc.SavePNG(output)
	if resizePercent != 1.0 {
		targetWidth := resizePercent * float64(canvasWidth)
		err = resizeOutput(output, int(targetWidth))
		if err != nil {
			fmt.Println("Error resizing output file")
			os.Exit(1)
		}
	}
}

func Bloom(img image.Image) image.Image {
	size := img.Bounds().Size()
	newSize := image.Rect(0, 0, size.X+20, size.Y+20)

	var extended image.Image
	extended = translateImage(img, newSize, 10, 10)

	dilated := effect.Dilate(extended, 0.5)

	bloomed := blur.Gaussian(dilated, 10.0)

	return bloomed
}

func resizeOutput(output string, width int) error {
	src, err := imaging.Open(output)
	if err != nil {
		return err
	}
	rc := imaging.Resize(src, width, 0, imaging.Lanczos)
	return imaging.Save(rc, output)
}

func translateImage(src image.Image, bounds image.Rectangle, xOffset, yOffset int) image.Image {
	rv := image.NewRGBA(bounds)
	size := src.Bounds().Size()
	for x := 0; x < size.X; x++ {
		for y := 0; y < size.Y; y++ {
			rv.Set(xOffset+x, yOffset+y, src.At(x, y))
		}
	}
	return rv
}

func parseColor(s string) [3]int {
	var rgb [3]int
	split := strings.Split(s, ",")
	for i := range split {
		split[i] = strings.TrimSpace(split[i])
	}

	for i := range split {
		val, err := strconv.Atoi(split[i])
		if err != nil {
			log.Fatal(err)
		}
		rgb[i] = val
	}

	return rgb
}

var file string
var rgb string
var output string
var canvasWidth int
var canvasHeight int
var resizePercent float64
var leftMargin int
var topMargin int
var font string

func init() {
	flag.StringVar(&font, "font", "", "Font file to use")
	flag.IntVar(&leftMargin, "margin-left", 10, "Text left margin")
	flag.IntVar(&topMargin, "margin-top", 10, "Text top margin")
	flag.IntVar(&canvasWidth, "width", 1024, "Canvas width")
	flag.IntVar(&canvasHeight, "height", 400, "Canvas height")
	flag.Float64Var(&resizePercent, "resize", 1.0, "Resoize percent")
	flag.StringVar(&output, "output", "output.png", "PNG file to write")
	flag.StringVar(&file, "file", "", "file with the text to render")
	flag.StringVar(&rgb, "rgb", "178,0,255", "the RGB color to use")
}
