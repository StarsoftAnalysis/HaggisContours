package main

// This file is part of hcontours -- HarrisContours.
// Copyright (C) 2024 Chris Dennis, chris@starsoftanalysis.co.uk
//
// hcontours is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

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

// Return true if the two angles (in radians) are close enough
func sameAngle(a1, a2 float64) bool {
	// FIXME should do mod(2pi)?
	return math.Abs(a1-a2) < 0.01
}

func mmOrInch(val, limit float64) float64 {
	if val > limit {
		return val
	}
	return val * 25.4
}

// Make sure n is within [low, high]
func limitInt(n, low, high int) int {
	if n < low {
		n = low
	} else if n > high {
		n = high
	}
	return n
}

// Return a slice of integers evenly spaced from 1 to 255.
// Assumes n is within the range 1 to 255
func evenThresholds(n int) []int {
	step := 256.0 / float64(n+1)
	thresholds := make([]int, n)
	for i := range n {
		thresholds[i] = int(math.Round(step * float64((i + 1))))
	}
	return thresholds
}

func almostEqual(a, b float64, epsilon float64) bool {
	return math.Abs(a-b) <= epsilon
}

func equalStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
