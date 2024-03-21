// svg.go
// Part of hcontours.go

package main

import (
	"fmt"
	"log"
	"os"
)

type paperSize struct {
	width  float64
	height float64
}

var paperSizes = map[string]paperSize{
	"A4L": paperSize{width: 297, height: 210},
	"A4P": paperSize{width: 210, height: 297},
	"A3L": paperSize{width: 420, height: 297},
	"A3P": paperSize{width: 297, height: 420},
}

type SVGfile struct {
	currentLayer int
	file         *os.File
	filename     string
}

func (svg *SVGfile) write(s string) {
	//fmt.Printf("SVG.write: %s\n", s)
	fmt.Fprint(svg.file, s)
}

func (svg *SVGfile) line(fromX, fromY, toX, toY float64) {
	// Write a line path; coordinates are ... scaling is done in svg.openStart
	svg.write(fmt.Sprintf("<path d=\"M %6.3f,%6.3f L %6.3f,%6.3f\" />\n", fromX, fromY, toX, toY))
}

func (svg *SVGfile) polygon(contour ContourT, width int) {
	// Single polygon -- assume the contour is closed
	// e.g.  <polygon points="100,100 150,25 150,75 200,0" fill="none" stroke="black" />
	//svg.write(fmt.Sprintf("<!-- contour: %v -->\n", contour))
	svg.write(fmt.Sprint("<polygon points=\""))
	for _, p := range contour {
		svg.write(fmt.Sprintf("%.2f,%.2f ", p.x, p.y))
	}
	//svg.write(fmt.Sprint("\" fill=\"none\" stroke=\"black\" stroke-width=\"0.1mm\" />\n"))
	svg.write(fmt.Sprint("\" />\n"))
}

func (svg *SVGfile) openStart(filename string, opts OptsT) {
	svg.filename = filename
	fh, err := os.Create(svg.filename)
	if err != nil {
		log.Fatalf("Unable to open SVG file %q - %s", svg.filename, err)
	}
	svg.file = fh
	// write the wrapper SVG with  background colour first
	viewbox := fmt.Sprintf("viewBox=\"0 0 %g %g\"", paperSizes[opts.paper].width, paperSizes[opts.paper].height)
	// Set background via style rather than filling an oversized rect (which upsets Axidraw)
	// (The style seems to be ignored by gThumb)
	bg := fmt.Sprintf("style=\"background-color:%s\"", "white")
	xmlns := "xmlns=\"http://www.w3.org/2000/svg\" xmlns:inkscape=\"http://www.inkscape.org/namespaces/inkscape\""
	svgAttribute := fmt.Sprintf("<svg width=\"%gmm\" height=\"%gmm\" %s %s %s encoding=\"UTF-8\" >\n",
		paperSizes[opts.paper].width, paperSizes[opts.paper].height, viewbox, bg, xmlns)
	svg.write(svgAttribute)

	// Apply translation and scale to whole plot: but don't magnify by more than a factor of 2
	//g := fmt.Sprintf("<g transform=\"translate(%g,%g) scale(%g)\" stroke=\"black\" stroke-width=\"1\" stroke-linecap=\"round\" stroke-linejoin=\"round\" fill=\"none\">\n",
	printWidth := paperSizes[opts.paper].width - 2*opts.margin
	printHeight := paperSizes[opts.paper].height - 2*opts.margin
	imageAspect := float64(opts.width) / float64(opts.height)
	printAspect := printWidth / printHeight
	//fmt.Printf("print %g x %g  img %d x %d   pA %g   iA  %g\n", printWidth, printHeight, opts.width, opts.height, printAspect, imageAspect)
	var scale, translateX, translateY float64
	if imageAspect > printAspect {
		scale = printWidth / float64(opts.width)
		//fmt.Println("scaling width")
		translateX = opts.margin
		translateY = (paperSizes[opts.paper].height - float64(opts.height)*scale) / 2
	} else {
		scale = printHeight / float64(opts.height)
		//fmt.Println("scaling height")
		translateX = (paperSizes[opts.paper].width - float64(opts.width)*scale) / 2
		translateY = opts.margin
	}
	scale = min(scale, 2.0)
	g := fmt.Sprintf("<g stroke=\"black\" stroke-width=\"0.1mm\" stroke-linecap=\"round\" stroke-linejoin=\"round\" fill=\"none\" transform=\"translate(%g,%g) scale(%.3f)\">\n",
		translateX, translateY, scale,
	)
	svg.write(g)
	if opts.frame {
		frame := fmt.Sprintf("<rect width=\"%d\" height=\"%d\" />", opts.width, opts.height)
		fmt.Println(frame)
		svg.write(frame)
	}
}

func (svg *SVGfile) stopSave() {
	svg.endLayer()
	svg.write("</g>\n</svg>\n")
	svg.file.Close()
	fmt.Printf("Created SVG file %q\n", svg.filename)
}

func (svg *SVGfile) startLayer(l int) {
	// l will not be 0 (that would plot white lines)
	//grey := 100 - int(math.Round(opts.penBandwidth*float64(l)*100.0))
	svg.write(fmt.Sprintf("<g inkscape:groupmode=\"layer\" inkscape:label=\"%d\" stroke=\"rgb(%d%%, %d%%, %d%%)\">\n", l, black, black, black))
	svg.currentLayer = l
}
func (svg *SVGfile) endLayer() {
	if svg.currentLayer > 0 {
		svg.write("</g>\n") // end of stroke and layer group
	}
}
func (svg *SVGfile) layer(l int) {
	if l == svg.currentLayer {
		// nothing to do
	} else {
		svg.endLayer()
		svg.startLayer(l)
	}
}
