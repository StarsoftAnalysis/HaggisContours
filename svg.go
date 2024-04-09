// svg.go

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
	"log"
	"math"
	"os"
	"path"
	"strconv"
	"strings"
)

type SVGfile struct {
	currentLayer    int
	file            *os.File
	filename        string
	pathCounter     int
	polygonCounter  int
	polylineCounter int
	thresholds      []int    // [0] is the background, so other indexes are bumped up by 1
	colours         []string //			SVGColourM // indexed by threshold
}

func (svg *SVGfile) write(s string) {
	fmt.Fprint(svg.file, s)
}

func (svg *SVGfile) writeComment(s string) {
	svg.write("<!-- " + s + " -->\n")
}

// Not used:
func (svg *SVGfile) line(fromX, fromY, toX, toY float64) {
	// Write a line path; coordinates are ... scaling is done in svg.openStart
	svg.write(fmt.Sprintf("<path id=\"%d\" d=\"M %6.3f,%6.3f L %6.3f,%6.3f\" />\n", svg.pathCounter, fromX, fromY, toX, toY))
	svg.pathCounter += 1
}

func (svg *SVGfile) polygon(contour ContourT, args string) {
	// Single polygon -- assume the contour is closed
	// e.g.  <polygon points="100,100 150,25 150,75 200,0" fill="none" stroke="black" />
	//svg.write(fmt.Sprintf("<!-- contour: %v -->\n", contour))
	//fmt.Printf("polygon: %v\n", contour)
	svg.write(fmt.Sprintf("<polygon id=\"%d\" %s points=\"", svg.polygonCounter, args))
	svg.polygonCounter += 1
	for _, p := range contour {
		svg.write(fmt.Sprintf("%.2f,%.2f ", p.x, p.y))
	}
	//svg.write(fmt.Sprint("\" fill=\"none\" stroke=\"black\" stroke-width=\"0.1mm\" />\n"))
	svg.write(fmt.Sprint("\" />\n"))
}

// Find the intercept between the line through p1 and p2 and the vertical line at x
func interceptX(p1, p2 Point64T, x float64) Point64T {
	m := (p2.y - p1.y) / (p2.x - p1.x)
	c := p1.y - m*p1.x
	y := m*x + c
	//fmt.Printf("iX: p1=%v p2=%v x=%v m=%v c=%v y=%v\n", p1, p2, x, m, c, y)
	return Point64T{x, y}
}

// Find the intercept between the line through p1 and p2 and the horizontal line at y
func interceptY(p1, p2 Point64T, y float64) Point64T {
	m := (p2.y - p1.y) / (p2.x - p1.x)
	c := p1.y - m*p1.x
	x := (y - c) / m
	return Point64T{x, y}
}

// Find the point on the edge of the image where the line
// from p1 to p2 crosses the edge.
// Assumes p1 is without the image, p2 is within it.
// NOTE to match with offImage() below, the edge is actually
// 1 pixel in.
func edgePoint(outPoint, inPoint Point64T, width, height int) Point64T {
	if outPoint.x < 0 {
		outPoint = interceptX(inPoint, outPoint, 0)
	}
	if outPoint.x > float64(width) {
		outPoint = interceptX(inPoint, outPoint, float64(width))
	}
	if outPoint.y < 0 {
		outPoint = interceptY(inPoint, outPoint, 0)
	}
	if outPoint.y > float64(height) {
		outPoint = interceptY(inPoint, outPoint, float64(height))
	}
	return outPoint
}

// 'off the image' includes contours around shapes that hit the edge.
// Because values have already been increased by 0.5 (in PointWeightedAvg()),
// choose anything here that's within 1 pixel of the edge.
// FIXME move this -- it's not an SVG thing
// FIXME limit is now 0.0
func offImage(p Point64T, width, height int) bool {
	const limit = 0.0 //1.0
	if p.x < limit || p.y < limit || p.x > float64(width)-limit || p.y > float64(height)-limit {
		return true
	}
	return false
}

// Given a contour (a slice of coordinates), make them into a polyline
func (svg *SVGfile) polyline(contour ContourT) {
	//fmt.Printf("polyline: %v\n", contour)
	svg.write(fmt.Sprintf("<polyline id=\"%d\" points=\"", svg.polylineCounter))
	svg.polylineCounter += 1
	for _, p := range contour {
		svg.write(fmt.Sprintf("%.2f,%.2f ", p.x, p.y))
	}
	svg.write(fmt.Sprint("\" />\n"))
}

// Polygon, or polyline if not closed
func (svg *SVGfile) polyshape(contour ContourT) {
	ccontour := contour.Compress()
	if ccontour[0].Equal(ccontour[len(ccontour)-1]) {
		svg.polygon(ccontour, "")
	} else {
		svg.polyline(ccontour)
	}
}

// Plot a contour onto the SVG file: as a polygon unless it goes off the
// edge of the image, in which case it becomes one or more polylines.
func (svg *SVGfile) plotContour(contour ContourT, width, height int) {
	lineOpen := false
	//fmt.Printf("plotC: contour=%v\n", contour)
	var subContour ContourT // may not be the whole contour
	for i, p := range contour {
		if offImage(p, width, height) {
			//fmt.Printf("plotC: offImage at %v  lineOpen=%v\n", p, lineOpen)
			if lineOpen {
				// stop the line - end right at the edge(s)
				edgeP := edgePoint(p, contour[i-1], width, height)
				subContour = append(subContour, edgeP)
				//fmt.Printf("plotC: stopping c-1=%v  p=%v  w=%v  h=%v  edgeP=%v subC=%v\n", contour[i-1], p, width, height, edgeP, subContour)
				svg.polyshape(subContour)
				subContour = nil
				lineOpen = false
			} else {
				//fmt.Printf("plotC: skipping %v\n", p)
				// line already closed -- skip the point
				// But wait!  what if we've gone over a corner?  FIXME TODO
				// Edge case (literally) -- line that starts and ends off-image -- see test11.png
			}
		} else {
			//fmt.Printf("plotC: on Image at %v  lineOpen=%v\n", p, lineOpen)
			if !lineOpen {
				// start a new line
				subContour = make(ContourT, 0, 10)
				if i > 0 {
					// Not the first point -- we've come back from off-image, so start on the edge
					edgeP := edgePoint(contour[i-1], p, width, height)
					//fmt.Printf("plotC: starting at edgeP %v\n", edgeP)
					subContour = append(subContour, edgeP)
				} else {
					//fmt.Printf("plotC: starting on image\n")
				}
				lineOpen = true
			}
			//fmt.Printf("plotC: adding %v\n", p)
			subContour = append(subContour, p)
		}
	}
	if lineOpen {
		// stop the line
		//fmt.Printf("plotC: final close\n")
		svg.polyshape(subContour)
		subContour = nil
	}
}

func (svg *SVGfile) closedPathStart(args string) {
	svg.write(fmt.Sprintf("<path id=\"%d\" clip-path=\"url(#clip1)\" %s d=\"", svg.pathCounter, args))
	svg.pathCounter += 1
}

func (svg *SVGfile) closedPathStop() {
	svg.write(fmt.Sprint("\" />\n"))
}

// Write one contour's worth of points to an already started path.
// e.g. M 10,20 L 20,20, L 20,10 Z
func (svg *SVGfile) closedPathLoop(contour ContourT, args string) {
	cmd := "M"
	for _, p := range contour {
		svg.write(fmt.Sprintf("%s %.2f,%.2f ", cmd, p.x, p.y))
		cmd = "L"
	}
	svg.write("Z ")
}

// Alternative strategy to plot a contour, using clipping instead of broken paths.
// This will allow filling, but won't work with AxiDraw.
func (svg *SVGfile) plotContourClip(contour ContourT, width, height int) {
	const args = "clip-path=\"url(#clip1)\""
	ccontour := contour.Compress()
	svg.closedPathLoop(ccontour, args)
}

func calcSizes(image RectangleT, margin float64, paper RectangleT, framewidth float64) (RectangleT, float64) {
	//g := fmt.Sprintf("<g transform=\"translate(%g,%g) scale(%g)\" stroke=\"black\" stroke-width=\"1\" stroke-linecap=\"round\" stroke-linejoin=\"round\" fill=\"none\">\n",
	printWidth := paper.width - 2*margin - 2*framewidth
	printHeight := paper.height - 2*margin - 2*framewidth
	imageAspect := float64(image.width) / float64(image.height)
	printAspect := printWidth / printHeight
	//fmt.Printf("print %g x %g  image %g x %g   pA %g   iA  %g\n", printWidth, printHeight, image.width, image.height, printAspect, imageAspect)
	var scale float64
	var translate RectangleT
	if imageAspect > printAspect {
		scale = printWidth / float64(image.width)
		//fmt.Println("scaling width")
		translate.width = margin + framewidth
		translate.height = (paper.height - float64(image.height)*scale) / 2
	} else {
		scale = printHeight / float64(image.height)
		//fmt.Println("scaling height")
		translate.width = (paper.width - float64(image.width)*scale) / 2
		translate.height = margin + framewidth
	}
	//fmt.Printf("translate = %g,%g  scale=%g\n", translate.width, translate.height, scale)
	return translate, scale
}

// Parse the colour string, e.g. "00ff00" or "123456,abcdef,ff7700" or "222222-eeeeee"
// into a slice of such values.
// The input has already been validated by regexp, so no error checking done here.
// Assumes svg.thresholds has already be set up.
func (svg *SVGfile) setColours(colourString string) {
	if colourString == "" {
		return
	}

	colourString = strings.ToLower(colourString)
	if len(colourString) == 6 {
		// Single colour -- treat as two (one for contour, one for background)
		svg.colours = []string{colourString, colourString}
		return
	}

	if colourString[6:7] == "," {
		// List of colours
		svg.colours = strings.Split(colourString, ",")
		return
	}

	// Range of colours  123456-abcdef
	svg.colours = make([]string, len(svg.thresholds))
	hex0 := colourString[:6]
	hex1 := colourString[7:]

	// first contour gets first colour
	svg.colours[0] = hex0

	tcount := len(svg.thresholds) // including 1 for the background
	if tcount > 2 {
		r0, _ := strconv.ParseInt(colourString[0:2], 16, 0)
		g0, _ := strconv.ParseInt(colourString[2:4], 16, 0)
		b0, _ := strconv.ParseInt(colourString[4:6], 16, 0)
		r1, _ := strconv.ParseInt(colourString[7:9], 16, 0)
		g1, _ := strconv.ParseInt(colourString[9:11], 16, 0)
		b1, _ := strconv.ParseInt(colourString[11:13], 16, 0)
		rStep := float64(r1-r0) / float64(tcount-1)
		gStep := float64(g1-g0) / float64(tcount-1)
		bStep := float64(b1-b0) / float64(tcount-1)
		//fmt.Printf("input=%s  t's=%v  %02x %02x %02x   %02x %02x %02x   step: %v %v %v\n", colourString, svg.thresholds, r0, g0, b0, r1, g1, b1, rStep, gStep, bStep)
		for i := 1; i < tcount; i++ {
			r := r0 + int64(math.Round(float64(i)*rStep))
			g := g0 + int64(math.Round(float64(i)*gStep))
			b := b0 + int64(math.Round(float64(i)*bStep))
			svg.colours[i] = fmt.Sprintf("%02x%02x%02x", r, g, b)
		}
	}

	// background counts as last threshold
	svg.colours[len(svg.colours)-1] = hex1
	//fmt.Printf("setColours: %#v\n", svg.colours)
}

func (svg *SVGfile) open(filename string) {
	svg.filename = filename
	fh, err := os.Create(svg.filename)
	if err != nil {
		log.Fatalf("Unable to open SVG file %q - %s", svg.filename, err)
	}
	svg.file = fh
	svg.write("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n") // needed so that next line can be a comment
}

func (svg *SVGfile) start(opts OptsT) (scale float64) {
	svg.currentLayer = -1                                 // no layer open
	svg.thresholds = append([]int{0}, opts.thresholds...) // the background counts as threshold 0
	svg.setColours(opts.colours)
	// write the wrapper SVG with  background colour first
	viewbox := fmt.Sprintf("viewBox=\"0 0 %g %g\"", opts.paperSize.width, opts.paperSize.height)
	// Set background via style rather than filling an oversized rect (which upsets Axidraw)
	// (The style seems to be ignored by gThumb)
	bg := fmt.Sprintf("style=\"background-color:%s\"", "white")
	xmlns := "xmlns=\"http://www.w3.org/2000/svg\" xmlns:inkscape=\"http://www.inkscape.org/namespaces/inkscape\""
	svgElement := fmt.Sprintf("<svg width=\"%gmm\" height=\"%gmm\" %s %s %s encoding=\"UTF-8\" >\n",
		opts.paperSize.width, opts.paperSize.height, viewbox, bg, xmlns)
	svg.write(svgElement)

	// Debug only: show paper limits
	if opts.debug {
		paperBox := fmt.Sprintf("<rect id=\"papersize\" width=\"%g\" height=\"%g\" stroke=\"blue\" stroke-dasharray=\"4\" fill=\"none\"/>\n", opts.paperSize.width, opts.paperSize.height)
		svg.write(paperBox)
	}

	translate, scale := calcSizes(RectangleT{float64(opts.width), float64(opts.height)}, opts.margin, opts.paperSize, opts.framewidth)

	// Debug only: show plot limits
	if opts.debug {
		plotBox := fmt.Sprintf("<rect id=\"plotsize\" width=\"%g\" height=\"%g\" x=\"%g\" y=\"%g\" stroke=\"green\" stroke-dasharray=\"3\" fill=\"none\"/>\n",
			float64(opts.width)*scale, float64(opts.height)*scale, translate.width, translate.height)
		svg.write(plotBox)
	}

	transform := fmt.Sprintf("transform=\"translate(%.4f,%.4f) scale(%.4f)\"", translate.width, translate.height, scale)

	// Main group -- scaled to fit paper
	// stroke-width is 'descaled' to result in what the user asked for
	g := fmt.Sprintf("<g stroke=\"black\" stroke-width=\"%.4f\" stroke-linecap=\"round\" stroke-linejoin=\"round\" fill=\"none\" %s>\n", opts.linewidth/scale, transform)
	svg.write(g)

	// Clippage is the amount to be taken off the edge of the image to hide the off-image
	// parts of contour polygons (while still allowing them to be filled with colour).
	// Ideally, clippage would be used in calcSizes, but that goes circular.
	// It's only an issue with very wide contour lines.
	clippage := 0.0
	if opts.clip {
		clippage = opts.linewidth / 2 / scale
	}

	if opts.clip { // inside the transformed group
		clipString := fmt.Sprintf("<defs><clipPath id=\"clip1\" ><rect id=\"cliprect\" width=\"%.4f\" height=\"%.4f\" x=\"%.4f\" y=\"%.4f\" /></clipPath></defs>\n", float64(opts.width)-clippage*2, float64(opts.height)-clippage*2, clippage, clippage)
		svg.write(clipString)
	}

	// Background layer for various reasons -- for clip because might have fill colours
	svg.layer(0, "background", len(svg.thresholds)-1)

	if opts.image {
		// CHECK clip image same as plot?
		imageString := fmt.Sprintf("<image id=\"background\" href=\"%s\" width=\"%d\" height=\"%d\" clip-path=\"url(#clip1)\" />\n", path.Base(opts.infile), opts.width, opts.height)
		//fmt.Print(imageString)
		svg.write(imageString)
	}

	// If colouring, need a background rect to be filled by the first colour
	if len(svg.colours) > 0 {
		rect := fmt.Sprintf("<rect id=\"plotsize\" width=\"%g\" height=\"%g\" stroke=\"none\" />\n",
			float64(opts.width), float64(opts.height))
		svg.write(rect)
	}

	//fmt.Printf("lw=%v  scale=%v   clippage=%v\n", opts.linewidth, scale, clippage)
	if opts.framewidth > 0.0 {
		// stroke-width is 'descaled' to result in what the user asked for:
		fwdescaled := opts.framewidth / scale
		w := float64(opts.width) + fwdescaled
		h := float64(opts.height) + fwdescaled
		// frame is outside the image, so shifted up and left a bit:
		x := -fwdescaled / 2
		y := -fwdescaled / 2
		if opts.clip {
			// adjust frame size and position to fit clipped image
			w -= 2 * clippage
			h -= 2 * clippage
			x += clippage
			y += clippage
		}
		//frame := fmt.Sprintf("<rect id=\"frame\" width=\"%d\" height=\"%d\" stroke-width=\"%.4f\" />\n", opts.width, opts.height, opts.framewidth/scale)
		frameString := fmt.Sprintf("<rect id=\"frame\" width=\"%.4f\" height=\"%.4f\" x=\"%.4f\" y=\"%.4f\" stroke-width=\"%.4f\" />\n", w, h, x, y, fwdescaled)
		//fmt.Print(frameString)
		svg.write(frameString)
	}

	return scale
}

func (svg *SVGfile) stopSave() {
	svg.endLayer()
	svg.write("</g>\n</svg>\n")
	svg.file.Close()
	fmt.Printf("Created SVG file %q\n", svg.filename)
}

func (svg *SVGfile) startLayer(l int, label string, colourIdx int) {
	fill := ""
	if len(svg.colours) > 0 {
		//fmt.Printf("svg.sL: contour fill: l=%d  svg.colours[%d]=%v\n", l, colourIdx, svg.colours[colourIdx%len(svg.colours)])
		fill = fmt.Sprintf("fill=\"#%s\"", svg.colours[colourIdx%len(svg.colours)])
	}
	svg.write(fmt.Sprintf("<g inkscape:groupmode=\"layer\" inkscape:label=\"%d %s\" stroke=\"black\" %s >\n", svg.thresholds[l], label, fill))
	svg.currentLayer = l
}
func (svg *SVGfile) endLayer() {
	if svg.currentLayer >= 0 {
		svg.write("</g>\n") // end of stroke and layer group
	}
	svg.currentLayer = -1
}
func (svg *SVGfile) layer(l int, label string, colourIdx int) {
	//fmt.Printf("svg.l: cL=%d l=%d label=%s\n", svg.currentLayer, l, label)
	if l == svg.currentLayer {
		// nothing to do
	} else {
		svg.endLayer()
		svg.startLayer(l, label, colourIdx)
	}
}
