// svg.go
// Part of hcontours.go

package main

import (
	"fmt"
	"log"
	"os"
	"path"
)

type SVGfile struct {
	currentLayer    int
	file            *os.File
	filename        string
	pathCounter     int
	polygonCounter  int
	polylineCounter int
}

func (svg *SVGfile) write(s string) {
	//fmt.Printf("SVG.write: %s\n", s)
	fmt.Fprint(svg.file, s)
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
	svg.write(fmt.Sprintf("<path id=\"%d\" clip-path=\"url(#clip1)\" %s d=\"", svg.pathCounter, args)) // FIXME better name for clip1 ?
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

func (svg *SVGfile) openStart(filename string, opts OptsT) {
	svg.filename = filename
	svg.currentLayer = -1 // no layer open
	fh, err := os.Create(svg.filename)
	if err != nil {
		log.Fatalf("Unable to open SVG file %q - %s", svg.filename, err)
	}
	svg.file = fh
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

	transform := fmt.Sprintf("transform=\"translate(%g,%g) scale(%.4f)\"", translate.width, translate.height, scale)

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
		/* old: */ //clipString := fmt.Sprintf("<defs><clipPath id=\"clip1\" ><rect width=\"%v\" height=\"%v\" x=\"%v\" y=\"%v\" /></clipPath></defs>\n", opts.width, opts.height, 0, 0)
		//clipString := fmt.Sprintf("<defs><clipPath id=\"clip1\" ><rect width=\"%v\" height=\"%v\" x=\"%v\" y=\"%v\" /></clipPath></defs>\n", float64(opts.width)-2*opts.linewidth, float64(opts.height)-2*opts.linewidth, opts.linewidth, opts.linewidth)
		//clipString := fmt.Sprintf("<defs><clipPath id=\"clip1\" ><rect id=\"cliprect\" width=\"%v\" height=\"%v\" x=\"%v\" y=\"%v\" /></clipPath></defs>\n", float64(opts.width)-2*opts.linewidth, float64(opts.height)-2*opts.linewidth, opts.linewidth, opts.linewidth)
		clipString := fmt.Sprintf("<defs><clipPath id=\"clip1\" ><rect id=\"cliprect\" width=\"%.4f\" height=\"%.4f\" x=\"%.4f\" y=\"%.4f\" /></clipPath></defs>\n", float64(opts.width)-clippage*2, float64(opts.height)-clippage*2, clippage, clippage)
		svg.write(clipString)
	}

	//fmt.Printf("lw=%v  scale=%v   clippage=%v\n", opts.linewidth, scale, clippage)
	if opts.frame {
		svg.layer(0, "frame")
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
		frame := fmt.Sprintf("<rect id=\"frame\" width=\"%.4f\" height=\"%.4f\" x=\"%.4f\" y=\"%.4f\" stroke-width=\"%.4f\" />\n", w, h, x, y, fwdescaled)
		//fmt.Print(frame)
		svg.write(frame)
		// FIXME image is in frame layer -- OK?  if not, need to call endLayer, and have it set currentLayer to -1
		//   and if no frame, the image is not in any layer -- what will axidraw do with it?  maybe
	}
	if opts.image {
		// CHECK clip image same as plot?
		image := fmt.Sprintf("<image id=\"background\" href=\"%s\" width=\"%d\" height=\"%d\" clip-path=\"url(#clip1)\" />\n", path.Base(opts.infile), opts.width, opts.height)
		//fmt.Print(image)
		svg.write(image)
	}
}

func (svg *SVGfile) stopSave() {
	svg.endLayer()
	svg.write("</g>\n</svg>\n")
	svg.file.Close()
	fmt.Printf("Created SVG file %q\n", svg.filename)
}

func (svg *SVGfile) startLayer(l int, label string) {
	svg.write(fmt.Sprintf("<g inkscape:groupmode=\"layer\" inkscape:label=\"%d %s\" stroke=\"black\">\n", l, label))
	svg.currentLayer = l
}
func (svg *SVGfile) endLayer() {
	if svg.currentLayer >= 0 {
		svg.write("</g>\n") // end of stroke and layer group
	}
}
func (svg *SVGfile) layer(l int, label string) {
	if l == svg.currentLayer {
		// nothing to do
	} else {
		svg.endLayer()
		svg.startLayer(l, label)
	}
}
