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
func getPix(imageData *image.NRGBA, width, height int, p PointT) int {
	if p.x < 0 || p.y < 0 || p.x >= width || p.y >= height {
		return white
	}
	pix := p.y*imageData.Stride + p.x*4
	//fmt.Printf("gP: p=%v  stride=%v  pix=%v\n", p, imageData.Stride, pix)
	return int(math.Round(0.299*float64(imageData.Pix[pix]) + 0.587*float64(imageData.Pix[pix+1]) + 0.114*float64(imageData.Pix[pix+2])))
}

// Calculate the angle from p1 to p2, in radians widdershins.
func relAngle(p1, p2 int, width int) float64 {
	pt1 := PointT{p1 % width, p1 / width}
	pt2 := PointT{p2 % width, p2 / width}
	return math.Atan2(float64(pt2.y-pt1.y), float64(pt2.x-pt1.x))
}

// Return true if the two angles are close enough
func sameAngle(a1, a2 float64) bool {
	return math.Abs(a1-a2) < 0.01
}

// Simplify a list of moves (between points) by combining consecutive moves
// in the same direction.
func compressMoves(moves []int, width int) []int {
	// moves is a slice of Pix indices.
	if len(moves) < 3 {
		return moves
	}
	var cmoves = make([]int, 0, len(moves)/2) // optimistic guess on the amount of compression
	p1 := moves[0]
	cmoves = append(cmoves, p1)
	i := 1
	p2 := moves[i]
	p3 := moves[i+1] // start the loop about here
	// calculate angle from one point to the next
	dir1 := relAngle(p1, p2, width)
	for i < len(moves)-1 {
		if p2 == p1 {
			// ignore non-moves
		} else {
			dir2 := relAngle(p2, p3, width)
			if sameAngle(dir1, dir2) {
				// do nothing: p1 and dir1 stay the same
			} else {
				// new direction -- add the point to the compressed array
				cmoves = append(cmoves, p2)
				p1 = p2
				dir1 = dir2
			}
		}
		i += 1
		p2 = p3
		if i+1 < len(moves) {
			p3 = moves[i+1]
		}
	}
	// need to add the last move
	cmoves = append(cmoves, moves[i])
	return cmoves
}

// Calculate the weighted average between points 'out' and 'in',
// based on where the threshold lies between the two pixel values.
// We expect the out pixel to have a higher value (i.e. be lighter)
// than the in pixel.
func pointWeightedAvg(img *image.NRGBA, out, in PointT, outPix, inPix int, threshold int) Point64T {
	//outpix := getPix(img, out)
	//inpix := getPix(img, in)
	if outPix == inPix {
		panic(fmt.Sprintf("pointWeightedAvg: points %v and %v shouldn't have the same pixel value (%v)", out, in, outPix))
	}
	proportion := float64(outPix-threshold) / float64(outPix-inPix)
	avgX := float64(out.x) + float64(in.x-out.x)*proportion
	avgY := float64(out.y) + float64(in.y-out.y)*proportion
	//fmt.Printf("pWA: out=%v %v  in=%v %v  t=%v  prop=%v  avg=%.2f,%.2f\n", out, outPix, in, inPix, threshold, proportion, avgX, avgY)
	return Point64T{avgX, avgY}
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
	contour = append(contour, pointWeightedAvg(imageData, out, in, outPix, inPix, threshold))
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
		// OPTION: add the point to the contour here to include the last point again.
		// Break if back at beginning
		//fmt.Printf("tE: break?  in=%v start=%v   dir=%v startDir=%v\n", in, start, direction, startDir)
		if in.Equal(start) && direction == startDir {
			break
		}
		// Add point to the contour
		contour = append(contour, pointWeightedAvg(imageData, out, in, outPix, inPix, threshold))
	}

	return contour, seen
}

func b2c(b bool) string {
	if b {
		return "t"
	}
	return "f"
}

func contourFinder(imageData *image.NRGBA, width, height int, threshold int, svgF *SVGfile) int {
	//var contours = make([]ContourT, 0, 10)
	seen := make([]bool, width*height)
	skipping := false
	contourCount := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			p := PointT{x, y}
			if getPix(imageData, width, height, p) < threshold {
				if !seen[x+y*width] && !skipping {
					contour, moreSeen := traceContour(imageData, width, height, threshold, p, svgF)
					contourCount += 1
					//contours = append(contours, contour)
					// this could be a _lot_ more efficient
					for _, p := range moreSeen {
						seen[p.x+p.y*width] = true
					}
					//fmt.Printf("eF: contour=%v  moreSeen=%v\n", contour, moreSeen)
					//for j := 0; j < height; j++ {
					//	for i := 0; i < width; i++ {
					//		fmt.Printf("%s", b2c(seen[i+j*width]))
					//	}
					//	fmt.Println()
					//}
					if svgF != nil {
						// Single polygon -- assume the contour is closed
						//ccontour := compressMoves(contour, width)
						svgF.polygon(contour, width)
					}
				}
				skipping = true
			} else {
				skipping = false
			}
		}
	}
	return contourCount
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
	optString := fmt.Sprintf("-contours-t%sm%g%s%s", intsToString(opts.thresholds), opts.margin, opts.paper, frameString)
	ext := filepath.Ext(opts.infile)
	svgFilename := strings.TrimSuffix(opts.infile, ext) + optString + ".svg"
	svgF.openStart(svgFilename, opts)
	for t, threshold := range opts.thresholds {
		svgF.layer(t + 1) // Axidraw layers start at 1, not 0
		contourCount := contourFinder(img, opts.width, opts.height, threshold, svgF)
		fmt.Printf("%d contours found at threshold %d\n", contourCount, threshold)
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
