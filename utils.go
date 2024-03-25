package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"strconv"
	"strings"
)

// loadImage loads the specified image from disk. Supported file types are png and jpg
func loadImage(path string) (*image.NRGBA, int, int, error) {
	srcReader, err := os.Open(path)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to read input image: %s, %s", path, err)
	}
	defer srcReader.Close()
	img, _, err := image.Decode(srcReader)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to decode image on load: %s, %s", path, err)
	}
	// Make sure image is NRGBA, with bounds starting at 0,0
	inputImage := ImageToNRGBA(img)
	bounds := inputImage.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y
	return inputImage, width, height, nil
}

// ImgToNRGBA converts any image type to *image.NRGBA with min-point at (0, 0).
// Copied from https://github.com/esimov/gomp/blob/master/image.go  April 2023
// via flyinggoat/utils/imageUtils.go
func ImageToNRGBA(img image.Image) *image.NRGBA {
	srcBounds := img.Bounds()
	if srcBounds.Min.X == 0 && srcBounds.Min.Y == 0 {
		if src0, ok := img.(*image.NRGBA); ok {
			//fmt.Println("ITNRGBA: simple conversion")
			return src0
		}
	}
	srcMinX := srcBounds.Min.X
	srcMinY := srcBounds.Min.Y

	dstBounds := srcBounds.Sub(srcBounds.Min)
	dstW := dstBounds.Dx()
	dstH := dstBounds.Dy()
	dst := image.NewNRGBA(dstBounds)

	switch src := img.(type) {
	case *image.NRGBA:
		//fmt.Println("ITNRGBA: converting from NRGBA")
		rowSize := srcBounds.Dx() * 4
		for dstY := 0; dstY < dstH; dstY++ {
			di := dst.PixOffset(0, dstY)
			si := src.PixOffset(srcMinX, srcMinY+dstY)
			for dstX := 0; dstX < dstW; dstX++ {
				copy(dst.Pix[di:di+rowSize], src.Pix[si:si+rowSize])
			}
		}
	case *image.YCbCr: // e.g. JPG
		//fmt.Println("ITNRGBA: converting from YCbCr")
		for dstY := 0; dstY < dstH; dstY++ {
			di := dst.PixOffset(0, dstY)
			for dstX := 0; dstX < dstW; dstX++ {
				srcX := srcMinX + dstX
				srcY := srcMinY + dstY
				siy := src.YOffset(srcX, srcY)
				sic := src.COffset(srcX, srcY)
				r, g, b := color.YCbCrToRGB(src.Y[siy], src.Cb[sic], src.Cr[sic])
				dst.Pix[di+0] = r
				dst.Pix[di+1] = g
				dst.Pix[di+2] = b
				dst.Pix[di+3] = 0xff
				di += 4
			}
		}
	default: // e.g. RGBA for PNG
		//fmt.Printf("ITNRGBA: converting from something else %T\n", src)
		for dstY := 0; dstY < dstH; dstY++ {
			di := dst.PixOffset(0, dstY)
			for dstX := 0; dstX < dstW; dstX++ {
				c := color.NRGBAModel.Convert(img.At(srcMinX+dstX, srcMinY+dstY)).(color.NRGBA)
				dst.Pix[di+0] = c.R
				dst.Pix[di+1] = c.G
				dst.Pix[di+2] = c.B
				dst.Pix[di+3] = c.A
				di += 4
			}
		}
	}
	return dst
}

func radians(degrees float64) float64 {
	return degrees * math.Pi / 180
}
func degrees(radians float64) float64 {
	return radians * 180 / math.Pi
}

func intsToString(ints []int) string {
	strs := make([]string, len(ints))
	for i, v := range ints {
		strs[i] = strconv.Itoa(v)
	}
	return strings.Join(strs, ",")
}

// From https://stackoverflow.com/questions/39544571/
func round(x, unit float64) float64 {
	return math.Round(x/unit) * unit
}
