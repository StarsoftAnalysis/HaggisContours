// hcontours.go

package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
)

const version = "0.1.0"

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
	// Fiddle with sign so that proportion always goes in direction from out to in
	//pwa.x = float64(out.x) + math.Abs(float64(in.x-out.x))*float64(sign(in.x, out.x))*proportion // + 0.5
	/*
			xproportion := proportion
			if out.x > in.x {
				xproportion = 1 - xproportion
			}
			yproportion := proportion
			if out.y > in.y {
				yproportion = 1 - yproportion
			}
		//pwa.x = float64(out.x) + math.Abs(float64(in.x-out.x))*xproportion + 0.5
		//pwa.y = float64(out.y) + math.Abs(float64(in.y-out.y))*yproportion + 0.5
	*/
	// Have to deal with edges separately:
	if out.x < 0 {
		pwa.x = -0.5
	} else if out.x >= width {
		pwa.x = float64(width) + 0.5
	} else {
		pwa.x = float64(out.x) + float64(in.x-out.x)*proportion + 0.5
	}
	if out.y < 0 {
		pwa.y = -0.5
	} else if out.y >= height {
		pwa.y = float64(height) + 0.5
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
func traceContour(imageData *image.NRGBA, width, height int, threshold int, start PointT, svgF *SVGfile) (ContourT, []PointT) {
	contour := make(ContourT, 0, 10)
	seen := make([]PointT, 1, 10) // Annoyingly, we need to also return a list of in-shape pixels
	seen[0] = start
	direction := DirectionT(approachDir) // we bumped into start pixel moving in +v x direction
	in := start                          // pixel in the shape
	out := in.Backstep(direction)        // one step back -- gives pixel outside the shape
	inPix := getPix(imageData, width, height, in)
	outPix := getPix(imageData, width, height, out)
	contour = append(contour, pointWeightedAvg(out, in, outPix, inPix, threshold, width, height))
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
		contour = append(contour, pointWeightedAvg(out, in, outPix, inPix, threshold, width, height))
		// Break if back at beginning
		//fmt.Printf("tE: break?  in=%v start=%v   dir=%v startDir=%v\n", in, start, direction, startDir)
		if in.Equal(start) && direction == startDir {
			break
		}
	}

	return contour, seen
}

func b2c(b bool) string {
	if b {
		return "t"
	}
	return "f"
}

func contourFinder(imageData *image.NRGBA, width, height int, threshold int, svgF *SVGfile) ContourS {
	seen := make([]bool, width*height)
	skipping := false
	contourCount := 0
	contours := make(ContourS, 0, 3)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			p := PointT{x, y}
			if getPix(imageData, width, height, p) < threshold {
				if !seen[x+y*width] && !skipping {
					contour, moreSeen := traceContour(imageData, width, height, threshold, p, svgF)
					contourCount += 1
					contours = append(contours, contour)
					// this could be a _lot_ more efficient
					for _, p := range moreSeen {
						seen[p.x+p.y*width] = true
					}
					if svgF != nil {
						svgF.plotContour(contour, width, height)
					}
				}
				skipping = true
			} else {
				skipping = false
			}
		}
	}
	// contours are really only returned for test cases
	// FIXME some contours may not have generated an SVG e.g. heightmap1
	//  could return a value from plotContour
	return contours
}

func parseArgs(args []string) OptsT {
	var opts OptsT
	pf := pflag.NewFlagSet("contours", pflag.ExitOnError)
	pf.Float64VarP(&opts.margin, "margin", "m", 15, "Minimum margin (in mm).")
	pf.StringVarP(&opts.paper, "paper", "p", "A4L", "Paper size and orientation.  A4L | A4P | A3L | A3P.")
	pf.IntSliceVarP(&opts.thresholds, "threshold", "t", []int{128}, "Threshold levels, each 0..255")
	pf.BoolVarP(&opts.frame, "frame", "f", false, "Draw a frame around the SVG image")
	pf.SortFlags = false
	if args == nil {
		pf.Parse(os.Args[1:]) // don't pass program name
	} else {
		pf.Parse(args) // args passed as a string (for testing)
	}
	if pf.NArg() < 1 {
		fmt.Println("No input file name given")
		os.Exit(1)
	}
	opts.infile = pf.Arg(0)
	return opts
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
	frameString := ""
	if opts.frame {
		frameString = "F"
	}
	optString := fmt.Sprintf("-hc-t%sm%g%s%s", intsToString(opts.thresholds), opts.margin, opts.paper, frameString)
	ext := filepath.Ext(opts.infile)
	svgFilename := strings.TrimSuffix(opts.infile, ext) + optString + ".svg"
	svgF.openStart(svgFilename, opts)
	for t, threshold := range opts.thresholds {
		svgF.layer(t + 1) // Axidraw layers start at 1, not 0   FIXME no, they don't
		contours := contourFinder(img, opts.width, opts.height, threshold, svgF)
		fmt.Printf("%d contours found at threshold %d\n", len(contours), threshold)
	}
	svgF.stopSave()
	return svgFilename
}

func main() {
	opts := parseArgs(nil)
	fmt.Printf("hcontours: processing '%s'\n", opts.infile)
	//fmt.Printf("options: %#v\n", opts)
	_ = createSVG(opts)
}
