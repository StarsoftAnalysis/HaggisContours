// hcontours.go

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

package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
)

const hcVersion = "0.1.1"

// Get the pixel value (0..255) at the given coordinates in the image
// Grey: Y = 0.299 R + 0.587 G + 0.114 B
// FIXME use Pix or At ?
func getPix(imageData *image.NRGBA, width, height int, p PointT) int {
	if p.x < 0 || p.y < 0 || p.x >= width || p.y >= height {
		//fmt.Printf("gP: p=%v  off edge, returning %v\n", p, white)
		return white
	}
	pixIndex := p.y*imageData.Stride + p.x*4
	//at := imageData.At(p.x, p.y)
	pixVal := int(math.Round(0.299*float64(imageData.Pix[pixIndex]) + 0.587*float64(imageData.Pix[pixIndex+1]) + 0.114*float64(imageData.Pix[pixIndex+2])))
	//fmt.Printf("gP: p=%v  len(Pix)=%v stride=%v  pixIndex=%v  pixVal=%v  at=%v\n", p, len(imageData.Pix), imageData.Stride, pixIndex, pixVal, at)
	return pixVal
}

// Calculate the weighted average between points 'out' and 'in',
// based on where the threshold lies between the two pixel values.
// We expect the out pixel to have a higher value (i.e. be lighter)
// than the in pixel, and the threshold to be in the range [inPix, outPix].
// The answer is shifted by 0.5 in each direction to account for
// the fence-post error: we're moving from the centres of pixels to the edges.
func pointWeightedAvg(out, in PointT, outPix, inPix, threshold int, width, height int) Point64T {
	if outPix == inPix || outPix < threshold || threshold < inPix {
		panic(fmt.Sprintf("pointWeightedAvg: invalid values for outPix (%v), threshold (%v), and inPix (%v)\n", outPix, threshold, inPix))
	}
	proportion := float64(outPix-threshold) / float64(outPix-inPix)
	var pwa Point64T
	// Have to deal with edges separately: make the average slightly off-image
	const slightly = 0.001
	if out.x < 0 {
		pwa.x = -slightly
	} else if out.x >= width {
		pwa.x = float64(width) + slightly
	} else {
		pwa.x = float64(out.x) + float64(in.x-out.x)*proportion + 0.5
	}
	if out.y < 0 {
		pwa.y = -slightly
	} else if out.y >= height {
		pwa.y = float64(height) + slightly
	} else {
		pwa.y = float64(out.y) + float64(in.y-out.y)*proportion + 0.5
	}
	//fmt.Printf("pWA: out=%v in=%v  outPix=%v threshold=%v inPix=%v  wd/ht=%v/%v  prop=%v  returning %v\n", out, in, outPix, threshold, inPix, width, height, proportion, pwa)
	return pwa
}

// Contour-finding strategy:
// * scan across width to first pixel >= threshold
// * turn left -- now have in-pixel on left, out-pixel on right
// * look at pixels ahead:
// - if left one is in, turn left
// - else if right one is in, straight on
// - else turn right
// (i.e. just a line-following thing)
// * accumulate weighted mid-points of each in/out pair
func traceContour(imageData *image.NRGBA, width, height int, threshold int, start PointT, svgF *SVGfile) (ContourT, []PointT, float64) {
	contour := make(ContourT, 0, 10)
	contourLen := 0.0
	seen := make([]PointT, 1, 10) // Annoyingly, we need to also return a list of in-shape pixels
	seen[0] = start
	direction := DirectionT(approachDir) // we bumped into start pixel moving in +v x direction
	in := start                          // pixel in the shape
	out := in.Backstep(direction)        // one step back -- gives pixel outside the shape
	inPix := getPix(imageData, width, height, in)
	outPix := getPix(imageData, width, height, out)
	prevPoint := pointWeightedAvg(out, in, outPix, inPix, threshold, width, height)
	contour = append(contour, prevPoint)
	direction.TurnLeft()
	startDir := direction // The direction we'll be facing when the contour is complete
	//fmt.Printf("\ntE: start=%v startDir=%v   in=%v inPix=%v  out=%v outPix=%v  contour=%v\n", start, startDir, in, inPix, out, outPix, contour)
	for {
		// Look ahead:
		// +------+------+
		// | Next | Next |
		// | Out  | In   |  ^
		// +------+------|  | Direction
		// | Out  |  In  |
		// +------+------+
		nextOut := out.Step(direction)
		nextIn := in.Step(direction)

		nextOutPix := getPix(imageData, width, height, nextOut)
		nextInPix := getPix(imageData, width, height, nextIn)

		if nextOutPix < threshold { // If next cell on the left is in the shape, turn left
			in = nextOut
			seen = append(seen, in)
			inPix = nextOutPix
			direction.TurnLeft()
			//fmt.Printf("tE: turn left: now in=%v inPix=%v  out=%v outPix=%v  dir=%v\n", in, inPix, out, outPix, direction)
		} else if nextInPix >= threshold { // If next cell on the right is not in the shape, turn right
			out = nextIn
			outPix = nextInPix
			direction.TurnRight()
			//fmt.Printf("tE: turn right: now in=%v inPix=%v  out=%v outPix=%v  dir=%v\n", in, inPix, out, outPix, direction)
		} else { // Otherwise, go straight on
			out = nextOut
			in = nextIn
			seen = append(seen, in)
			inPix = nextOutPix
			outPix = nextOutPix
			inPix = nextInPix
			//fmt.Printf("tE: straight on: now in=%v inPix=%v  out=%v outPix=%v  dir=%v\n", in, inPix, out, outPix, direction)
		}

		// Add point to the contour (including the repeated point that closes the loop)
		nextPoint := pointWeightedAvg(out, in, outPix, inPix, threshold, width, height)
		contour = append(contour, nextPoint)
		contourLen += prevPoint.Distance(nextPoint)
		prevPoint = nextPoint
		// Break if back at beginning
		//fmt.Printf("tE: break?  in=%v start=%v   dir=%v startDir=%v\n", in, start, direction, startDir)
		if in.Equal(start) && direction == startDir {
			break
		}
	}

	return contour, seen, contourLen
}

func b2c(b bool) string {
	if b {
		return "t"
	}
	return "f"
}

func contourFinder(imageData *image.NRGBA, width, height int, threshold int, clip bool, svgF *SVGfile) (ContourS, float64) {
	seen := make([]bool, width*height)
	skipping := false
	contourCount := 0
	contours := make(ContourS, 0, 3)
	totalLen := 0.0
	if clip {
		// start the path -- single path for all contours.   TODO maybe like this for non-clipped paths
		svgF.closedPathStart("")
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			p := PointT{x, y}
			if getPix(imageData, width, height, p) < threshold {
				if !seen[x+y*width] && !skipping {
					contour, moreSeen, contourLen := traceContour(imageData, width, height, threshold, p, svgF)
					contourCount += 1
					contours = append(contours, contour)
					totalLen += contourLen
					// this could be a _lot_ more efficient
					for _, p := range moreSeen {
						seen[p.x+p.y*width] = true
					}
					if svgF != nil {
						if clip {
							svgF.plotContourClip(contour, width, height)
						} else {
							svgF.plotContour(contour, width, height)
						}
					}
				}
				skipping = true
			} else {
				skipping = false
			}
		}
	}
	if clip {
		// finish the (closed) path
		svgF.closedPathStop()
	}
	// contours are really only returned for test cases
	// FIXME some contours may not have generated an SVG e.g. heightmap1
	//  could return a value from plotContour
	return contours, totalLen
}

func parsePaperSize(opts *OptsT) bool {
	valid := true
	ps := strings.ToUpper((*opts).paper)
	dims := strings.Split(ps, "X")
	//fmt.Printf("pPS: ps=%v dims=%v\n", ps, dims)
	if len(dims) == 1 {
		// no 'X' -- should be a standard size
		size, ok := paperSizes[ps]
		if !ok {
			valid = false
		} else {
			opts.paperSize = size
		}
	} else if len(dims) == 2 {
		// something like 123x45
		paperWidth, err := strconv.ParseFloat(dims[0], 64)
		paperHeight, err := strconv.ParseFloat(dims[1], 64)
		fmt.Printf("pps: pW=%v pH=%v err=%v\n", paperWidth, paperHeight, err)
		if err != nil {
			valid = false
		} else {
			paperWidth = mmOrInch(paperWidth, 30)
			paperHeight = mmOrInch(paperHeight, 30)
			(*opts).paperSize = RectangleT{width: paperWidth, height: paperHeight}
		}
	} else {
		// too many X's
		valid = false
	}
	if !valid {
		fmt.Printf("Can't make head nor tail of paper size '%s'\n", opts.paper)
	}
	return valid
}

func parseArgs(args []string) (OptsT, bool) {
	var opts OptsT
	pf := pflag.NewFlagSet("contours", pflag.ExitOnError)
	pf.IntSliceVarP(&opts.thresholds, "threshold", "t", []int{128}, "Threshold levels, each 0..255, separated by commas.")
	pf.IntVarP(&opts.tcount, "tcount", "T", 1, "Number of evenly-spaced threshold levels (unless overridden by --threshold).")
	pf.Float64VarP(&opts.margin, "margin", "m", 15, "Minimum margin (in mm).")
	pf.StringVarP(&opts.paper, "paper", "p", "A4L", "Paper size and orientation.  A4L | A4P | A3L | A3P.")
	pf.Float64VarP(&opts.linewidth, "linewidth", "l", 0.5, "Width of contour lines, in mm.")
	pf.Float64VarP(&opts.framewidth, "framewidth", "f", 0.0, "Width of frame lines, if any, in mm.")
	pf.BoolVarP(&opts.image, "image", "i", false, "Use the original image as a background in the SVG image.")
	pf.BoolVarP(&opts.clip, "clip", "c", false, "Clip borders of image, rather than breaking contours.")
	pf.BoolVarP(&opts.debug, "debug", "d", false, "Add extra bits to the SVG -- intended for developer use only.")
	pf.SortFlags = false
	if args == nil {
		pf.Parse(os.Args[1:]) // don't pass program name
	} else {
		pf.Parse(args) // args passed as a string (for testing)
	}
	ok := true
	if pf.NArg() < 1 {
		fmt.Println("No input file name given")
		ok = false
	}
	if pf.Changed("threshold") {
		// User has set thresholds -- don't use tcount
		opts.tcount = -1
	} else {
		opts.tcount = limitInt(opts.tcount, 1, 255)
		opts.thresholds = evenThresholds(opts.tcount)
	}
	opts.infile = pf.Arg(0)
	ok = ok && parsePaperSize(&opts)
	if ok {
		opts.margin = mmOrInch(opts.margin, 2)
		if opts.paperSize.width < opts.margin*3 || opts.paperSize.height < opts.margin*3 {
			fmt.Printf("Margin %g mm is too big for paper size %g x %g mm\n", opts.margin, opts.paperSize.width, opts.paperSize.height)
			ok = false
		}
	}
	return opts, ok
}

func buildSVGfilename(opts OptsT) string {
	frameString := ""
	if opts.framewidth > 0.0 {
		frameString = fmt.Sprintf("F%g", opts.framewidth)
	}
	imageString := ""
	if opts.image {
		imageString = "I"
	}
	clipString := ""
	if opts.clip {
		clipString = "C"
	}
	tString := ""
	if opts.tcount == -1 {
		tString = "t" + intsToString(opts.thresholds)
	} else {
		tString = fmt.Sprintf("T%d", opts.tcount)
	}
	optString := fmt.Sprintf("-hc-%sm%gp%s%s%s%s", tString, opts.margin, opts.paper, frameString, imageString, clipString)
	ext := filepath.Ext(opts.infile)
	filename := strings.TrimSuffix(opts.infile, ext) + optString + ".svg"
	return filename
}

func createSVG(opts OptsT) string {
	var svgF *SVGfile = new(SVGfile)
	img, width, height, err := loadImage(opts.infile)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	opts.width = width
	opts.height = height
	svgFilename := buildSVGfilename(opts)
	scale := svgF.openStart(svgFilename, opts)
	contourText := make([]string, len(opts.thresholds))
	totalLen := 0.0
	for i, threshold := range opts.thresholds {
		svgF.layer(threshold, "contour")
		contours, thresholdLen := contourFinder(img, opts.width, opts.height, threshold, opts.clip, svgF)
		contourText[i] = fmt.Sprintf("%d contours found at threshold %d, with length %.2fm", len(contours), threshold, thresholdLen*scale/1000)
		totalLen += thresholdLen
	}
	svgF.endLayer()
	for _, text := range contourText {
		fmt.Println(text)
		svgF.writeComment(text)
	}
	text := fmt.Sprintf("Total contour length: %.2fm", totalLen*scale/1000)
	fmt.Println(text)
	svgF.writeComment(text)
	svgF.stopSave()
	return svgFilename
}

func main() {
	opts, ok := parseArgs(nil)
	if !ok {
		os.Exit(1)
	}
	fmt.Printf("hcontours: processing '%s'\n", opts.infile)
	//fmt.Printf("\t%+v\n", opts)
	//fmt.Printf("options: %#v\n", opts)
	_ = createSVG(opts)
}
