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
	"math"
	"os"
	"testing"
)

func TestTypes(t *testing.T) {
	fmt.Println("TestTypes")
	type testdata1T struct {
		a, b Point64T
	}
	testdata1 := []testdata1T{
		{Point64T{1.0000, 1.0000}, Point64T{1.0000, 1.0001}},
		{Point64T{math.Pi, 2.9999}, Point64T{223.0 / 71.0, 3.0}},
	}
	for _, td := range testdata1 {
		if !td.a.Equal(td.b) {
			t.Errorf("Point64T's not equal: a=%v b=%v\n", td.a, td.b)
		}
	}
}

func TestPWA(t *testing.T) {
	fmt.Println("TestPWA")
	type testdataT struct {
		outPt, inPt              PointT
		outPix, inPix, threshold int
		width, height            int
		wanted                   Point64T
	}
	testdata := []testdataT{
		// On-image points
		{PointT{1, 0}, PointT{1, 1}, 200, 20, 80, 3, 3, Point64T{1.500, 1.167}},
		{PointT{2, 1}, PointT{1, 1}, 200, 20, 80, 3, 3, Point64T{1.833, 1.500}},
		{PointT{1, 2}, PointT{1, 1}, 200, 20, 80, 3, 3, Point64T{1.500, 1.833}},
		{PointT{0, 1}, PointT{1, 1}, 200, 20, 80, 3, 3, Point64T{1.167, 1.500}},
		// Edge points -- outPt is off the image
		{PointT{0, -1}, PointT{0, 0}, 200, 20, 80, 2, 2, Point64T{0.5, -0.001}},
		{PointT{2, 0}, PointT{1, 0}, 200, 20, 80, 2, 2, Point64T{2.001, 0.5}},
		{PointT{1, 2}, PointT{1, 1}, 200, 20, 80, 2, 2, Point64T{1.5, 2.001}},
		{PointT{-1, 1}, PointT{0, 1}, 200, 20, 80, 2, 2, Point64T{-0.001, 1.5}},
	}
	for i, td := range testdata {
		got := pointWeightedAvg(td.outPt, td.inPt, td.outPix, td.inPix, td.threshold, td.width, td.height)
		if !got.Equal(td.wanted) {
			t.Errorf("Wrong result for %d:  out %v %v  in %v %v  t %v  wanted %v  got %v\n", i, td.outPt, td.outPix, td.inPt, td.inPix, td.threshold, td.wanted, got)
		}
	}
}

// TODO colours tests
func TestFilename(t *testing.T) {
	fmt.Println("TestFilename")
	type testdataT struct {
		opts   OptsT
		wanted string
	}
	/*
		infile     string
		width      int
		height     int
		thresholds []int
		tcount     int
		margin     float64
		paper      string
		paperSize  RectangleT
		image      bool
		clip       bool
		dev        bool
		linewidth  float64
		framewidth float64
		colours    string // two hex colours, e.g. "0033ff,0c4088"
	*/
	testdata := []testdataT{
		{OptsT{"file1.png", 100, 200, []int{44, 55}, -1, 15.0, "5x7", RectangleT{0, 0}, true, false, true, 1.0, 0.0, ""},
			"file1-hc-t44,55m15p5x7I.svg"},
		{OptsT{"file1.png", 100, 200, []int{}, 3, 10.3, "200x300", RectangleT{0, 0}, false, true, false, 1.0, 2.0, ""},
			"file1-hc-T3m10.3p200x300F2C.svg"},
	}
	for i, td := range testdata {
		filename := buildSVGfilename(td.opts)
		if filename != td.wanted {
			t.Errorf("(%d) Wrong filename: wanted '%s' got '%s'\n", i, td.wanted, filename)
		}
	}
}

func TestSetColours(t *testing.T) {
	fmt.Println("TestSetColours")
	type testdataT struct {
		id           int
		colourString string
		tcount       int      // just need the number of thresholds
		colours      []string // expected result
	}
	testdata := []testdataT{
		{1, "abcdef", 1, []string{"abcdef", "abcdef"}},
		{2, "abcdef", 2, []string{"abcdef", "abcdef"}},
		{3, "abcdef,123456", 1, []string{"abcdef", "123456"}},
		{4, "abcdef,123456", 3, []string{"abcdef", "123456"}},
		{5, "111111-999999", 1, []string{"111111", "999999"}},
		{6, "111111-999999", 4, []string{"111111", "333333", "555555", "777777", "999999"}},
		{7, "FFFFFF-000000", 5, []string{"ffffff", "cccccc", "999999", "666666", "333333", "000000"}},
	}
	for _, td := range testdata {
		svg := new(SVGfile)
		svg.thresholds = make([]int, td.tcount+1) // plus 1 for the background
		svg.setColours(td.colourString)
		if !equalStringSlice(svg.colours, td.colours) {
			t.Errorf("Wrong result for test %d: %s / %d.  Wanted '%s'  got '%s'\n", td.id, td.colourString, td.tcount, td.colours, svg.colours)
		}
	}
}

func TestTraceContour(t *testing.T) {
	fmt.Println("TestTraceContour")
	type testdataT struct {
		infile  string
		contour ContourT
		start   PointT
		length  float64
	}
	testdata := []testdataT{ // Don't forget that no compression happens for these cases
		{"tests/test0.png", ContourT{
			{0.998, 1.500}, {1.500, 0.998}, {2.500, 0.998}, {3.002, 1.500}, {3.002, 2.500}, {2.500, 3.002}, {1.500, 3.002}, {0.998, 2.500},
			{0.998, 1.500},
		}, PointT{1, 1}, 6.840},
		{"tests/test1.png", ContourT{
			{0.998, 1.500}, {1.500, 0.998}, {2.500, 0.998}, {3.500, 0.998}, {4.002, 1.500}, {3.500, 2.002}, {3.002, 2.500}, {2.500, 3.002},
			{2.002, 3.500}, {1.500, 4.002}, {0.998, 3.500}, {0.998, 2.500}, {0.998, 1.500},
		}, PointT{1, 1}, 9.668},
		{"tests/test4.png", ContourT{
			{0.998, 0.500}, {1.500, -0.001}, {2.002, 0.500}, {2.500, 0.998}, {3.002, 1.500}, {3.002, 2.500}, {2.500, 3.002}, {2.002, 3.500},
			{1.500, 4.001}, {0.998, 3.500}, {0.500, 3.002}, {-0.001, 2.500}, {-0.001, 1.500}, {0.500, 0.998}, {0.998, 0.500},
		}, PointT{1, 0}, 10.492},
	}
	for _, td := range testdata {
		fmt.Printf("\t%s\n", td.infile)
		img, width, height, err := loadImage(td.infile)
		if err != nil {
			t.Errorf("Input file %s not found\n", td.infile)
		}
		got, _, length := traceContour(img, width, height, 128, td.start, nil)
		if !almostEqual(length, td.length, 0.001) {
			t.Errorf("Wrong result for %s (wanted length %.3f  got %.3f)\n", td.infile, td.length, length)
		}
		if !got.Equal(td.contour) {
			t.Errorf("Wrong result for %s start %v:\n\twanted=%v\n\t   got %v\n", td.infile, td.start, td.contour, got)
		}
	}
}

func TestContourFinder(t *testing.T) {
	fmt.Println("TestContourFinder")
	type testdataT struct {
		infile   string
		contours ContourS
		count    int
		length   float64
	}
	testdata := []testdataT{ // Don't forget that no compression happens for these cases
		{"tests/test0.png", ContourS{
			{{0.998, 1.500}, {1.500, 0.998}, {2.500, 0.998}, {3.002, 1.500}, {3.002, 2.500}, {2.500, 3.002}, {1.500, 3.002}, {0.998, 2.500}, {0.998, 1.500}},
		}, 1, 6.840},
		{"tests/test1.png", ContourS{
			{{0.998, 1.500}, {1.500, 0.998}, {2.500, 0.998}, {3.500, 0.998}, {4.002, 1.500}, {3.500, 2.002}, {3.002, 2.500}, {2.500, 3.002}, {2.002, 3.500},
				{1.500, 4.002}, {0.998, 3.500}, {0.998, 2.500}, {0.998, 1.500}},
		}, 1, 9.668},
		{"tests/test2.png", ContourS{
			{{-0.001, 0.500}, {0.500, -0.001}, {1.500, -0.001}, {2.002, 0.500}, {1.500, 1.002}, {1.002, 1.500}, {0.500, 2.002}, {-0.001, 1.500}, {-0.001, 0.500}},
			{{2.998, 2.500}, {3.500, 1.998}, {4.001, 2.500}, {4.001, 3.500}, {3.500, 4.001}, {2.500, 4.001}, {1.998, 3.500}, {2.500, 2.998}, {2.998, 2.500}},
		}, 2, 12.502},
		{"tests/test3.png", ContourS{
			{{0.998, 0.500}, {1.500, -0.001}, {2.500, -0.001}, {3.002, 0.500}, {3.002, 1.500}, {2.500, 2.002}, {2.002, 2.500}, {1.500, 3.002}, {0.500, 3.002},
				{-0.001, 2.500}, {-0.001, 1.500}, {0.500, 0.998}, {0.998, 0.500}},
			{{3.998, 0.500}, {4.500, -0.001}, {5.500, -0.001}, {6.500, -0.001}, {7.500, -0.001}, {8.001, 0.500}, {8.001, 1.500}, {8.001, 2.500}, {8.001, 3.500},
				{8.001, 4.500}, {8.001, 5.500}, {8.001, 6.500}, {8.001, 7.500}, {7.500, 8.001}, {6.500, 8.001}, {5.500, 8.001}, {4.500, 8.001}, {3.500, 8.001},
				{2.500, 8.001}, {1.500, 8.001}, {0.500, 8.001}, {-0.001, 7.500}, {-0.001, 6.500}, {-0.001, 5.500}, {-0.001, 4.500}, {0.500, 3.998}, {1.500, 3.998},
				{2.500, 3.998}, {2.998, 3.500}, {3.500, 2.998}, {3.998, 2.500}, {3.998, 1.500}, {3.998, 0.500}},
			{{4.998, 4.500}, {5.500, 3.998}, {6.002, 4.500}, {6.002, 5.500}, {5.500, 6.002}, {4.500, 6.002}, {3.998, 5.500}, {4.500, 4.998}, {4.998, 4.500}},
		}, 3, 45.581},
		{"tests/test4.png", ContourS{
			{{0.998, 0.500}, {1.500, -0.001}, {2.002, 0.500}, {2.500, 0.998}, {3.002, 1.500}, {3.002, 2.500}, {2.500, 3.002}, {2.002, 3.500}, {1.500, 4.001},
				{0.998, 3.500}, {0.500, 3.002}, {-0.001, 2.500}, {-0.001, 1.500}, {0.500, 0.998}, {0.998, 0.500}},
			{{3.998, 0.500}, {4.500, -0.001}, {5.500, -0.001}, {6.001, 0.500}, {6.001, 1.500}, {5.500, 2.002}, {4.500, 2.002}, {3.998, 1.500}, {3.998, 0.500}},
			{{3.998, 3.500}, {4.500, 2.998}, {5.002, 3.500}, {4.500, 4.001}, {3.998, 3.500}},
		}, 3, 20.167},
		{"tests/test5.png", ContourS{
			{{0.998, 1.500}, {1.500, 0.998}, {2.500, 0.998}, {3.500, 0.998}, {4.500, 0.998}, {5.500, 0.998}, {6.500, 0.998}, {7.002, 1.500}, {7.002, 2.500},
				{7.002, 3.500}, {7.002, 4.500}, {7.002, 5.500}, {7.002, 6.500}, {6.500, 7.002}, {5.500, 7.002}, {4.500, 7.002}, {3.500, 7.002}, {2.500, 7.002},
				{1.500, 7.002}, {0.998, 6.500}, {0.998, 5.500}, {0.998, 4.500}, {0.998, 3.500}, {0.998, 2.500}, {0.998, 1.500}},
			{{4.998, 3.500}, {4.500, 3.002}, {3.500, 3.002}, {3.002, 3.500}, {3.002, 4.500}, {3.500, 4.998}, {4.500, 4.998}, {4.998, 4.500}, {4.998, 3.500}},
		}, 2, 29.657},
		{"tests/test6.png", ContourS{
			{{0.998, 1.500}, {1.500, 0.998}, {2.500, 0.998}, {3.500, 0.998}, {4.500, 0.998}, {5.500, 0.998}, {6.500, 0.998}, {7.002, 1.500}, {7.002, 2.500},
				{7.002, 3.500}, {7.002, 4.500}, {7.002, 5.500}, {7.002, 6.500}, {6.500, 7.002}, {5.500, 7.002}, {4.500, 7.002}, {3.500, 7.002}, {2.500, 7.002},
				{1.500, 7.002}, {0.998, 6.500}, {0.998, 5.500}, {0.998, 4.500}, {0.998, 3.500}, {0.998, 2.500}, {0.998, 1.500}},
		}, 1, 22.840},
		// These two have non-closed thin lines -- the contour loops back to close itself:
		{"tests/test7.png", ContourS{
			{{0.998, 1.500}, {1.500, 0.998}, {2.500, 0.998}, {3.500, 0.998}, {4.002, 1.500}, {4.002, 2.500}, {4.002, 3.500}, {3.500, 4.002}, {2.500, 4.002},
				{2.002, 4.500}, {2.500, 4.998}, {3.500, 4.998}, {4.002, 5.500}, {3.500, 6.002}, {2.500, 6.002}, {1.500, 6.002}, {0.998, 5.500}, {0.998, 4.500},
				{0.998, 3.500}, {1.500, 2.998}, {2.500, 2.998}, {2.998, 2.500}, {2.500, 2.002}, {1.500, 2.002}, {0.998, 1.500}},
		}, 1, 20.496},
		{"tests/test8.png", ContourS{
			{{0.998, 1.500}, {1.500, 0.998}, {2.500, 0.998}, {3.500, 0.998}, {4.500, 0.998}, {5.500, 0.998}, {6.500, 0.998}, {7.500, 0.998}, {8.002, 1.500},
				{7.500, 2.002}, {6.500, 2.002}, {5.500, 2.002}, {4.500, 2.002}, {3.500, 2.002}, {2.500, 2.002}, {1.500, 2.002}, {0.998, 1.500}},
			{{0.998, 3.500}, {1.500, 2.998}, {2.002, 3.500}, {2.002, 4.500}, {2.002, 5.500}, {2.002, 6.500}, {2.002, 7.500}, {1.500, 8.002}, {0.998, 7.500},
				{0.998, 6.500}, {0.998, 5.500}, {0.998, 4.500}, {0.998, 3.500}},
			{{2.998, 3.500}, {3.500, 2.998}, {4.002, 3.500}, {4.500, 3.998}, {5.002, 4.500}, {5.500, 4.998}, {6.002, 5.500}, {6.500, 5.998}, {7.002, 6.500},
				{7.500, 6.998}, {8.002, 7.500}, {7.500, 8.002}, {6.998, 7.500}, {6.500, 7.002}, {5.998, 6.500}, {5.500, 6.002}, {4.998, 5.500}, {4.500, 5.002},
				{3.998, 4.500}, {3.500, 4.002}, {2.998, 3.500}},
		}, 3, 39.832},
		{"tests/example.png", nil, 10, 3663.063},
	}
	for _, td := range testdata {
		fmt.Printf("\t%s\n", td.infile)
		img, width, height, err := loadImage(td.infile)
		if err != nil {
			t.Errorf("Input file %s not found\n", td.infile)
		}
		contours, length := contourFinder(img, width, height, 128, false, nil)
		if len(contours) != td.count {
			t.Errorf("Wrong result for %s (wanted count %v  got %v)\n", td.infile, td.count, len(contours))
		}
		if !almostEqual(length, td.length, 0.001) {
			t.Errorf("Wrong result for %s (wanted length %.3f  got %.3f)\n", td.infile, td.length, length)
		}
		if td.contours != nil {
			equal := true
			for i := 0; equal && i < len(contours); i++ {
				if !contours[i].Equal(td.contours[i]) {
					equal = false
				}
			}
			if !equal {
				t.Errorf("Wrong result for %s\n\twanted %v\n\t   got %v\n", td.infile, td.contours, contours)
			}
		}
	}
}

func TestCalcSizes(t *testing.T) {
	fmt.Println("TestCalcSizes")
	type testdataT struct {
		image      RectangleT
		margin     float64
		paper      RectangleT
		framewidth float64
		translate  RectangleT
		scale      float64
	}
	testdata := []testdataT{
		{RectangleT{400, 400}, 00, RectangleT{400, 400}, 0, RectangleT{0, 0}, 1},
		{RectangleT{100, 200}, 50, RectangleT{400, 400}, 5, RectangleT{127.5000, 55.0000}, 1.45},
		{RectangleT{600, 200}, 40, RectangleT{400, 400}, 2, RectangleT{42.0000, 147.3333}, 0.5267},
		{RectangleT{600, 400}, 15, RectangleT{297, 210}, 4, RectangleT{19.5000, 19.0000}, 0.43},
		{RectangleT{6, 4}, 15, RectangleT{297, 210}, 0.5, RectangleT{15.5000, 16.3333}, 44.3333},
	}
	for i, td := range testdata {
		translate, scale := calcSizes(td.image, td.margin, td.paper, td.framewidth)
		if !translate.Equal(td.translate) || !almostEqual(scale, td.scale, 0.001) {
			t.Errorf("(%d) Wrong result with image=%v margin=%v paper=%v fwidth=%v:\n\twanted %v, %g   got %v, %g",
				i, td.image, td.margin, td.paper, td.framewidth, td.translate, td.scale, translate, scale)
		}
	}
}

func TestCreateSVG(t *testing.T) {
	fmt.Println("TestCreateSVG")
	type testdataT struct {
		infile     string
		outfile    string
		thresholds []int
		margin     float64
		framewidth float64
		paper      string
		clip       bool
		colours    string
		wanted     string
	}
	testdata := []testdataT{ // Compression is done for SVG contours
		{"tests/test3.png", "tests/test3-hc-t128m15pA4LF2.svg", []int{128}, 15, 2.0, "A4L", false, "",
			"<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<!-- tests/test3-hc-t128m15pA4LF2.svg, created by hcontours.test version 0.1.2 -->\n<!-- Options used: infile: \"tests/test3.png\", width: 8, height: 8, thresholds: [128], tcount: -1, margin: 15.00, paper: \"A4L\", paperSize: {297.00, 210.00}, image: false, clip: false, debug: false, linewidth: 1.00, framewidth: 2.00, colours: \"\" -->\n<svg width=\"297mm\" height=\"210mm\" viewBox=\"0 0 297 210\" style=\"background-color:white\" xmlns=\"http://www.w3.org/2000/svg\" xmlns:inkscape=\"http://www.inkscape.org/namespaces/inkscape\" encoding=\"UTF-8\" >\n<g stroke=\"black\" stroke-width=\"0.0455\" stroke-linecap=\"round\" stroke-linejoin=\"round\" fill=\"none\" transform=\"translate(60.5000,17.0000) scale(22.0000)\">\n<g inkscape:groupmode=\"layer\" inkscape:label=\"0 background\" stroke=\"black\"  >\n<rect id=\"frame\" width=\"8.0909\" height=\"8.0909\" x=\"-0.0455\" y=\"-0.0455\" stroke-width=\"0.0909\" />\n</g>\n<g inkscape:groupmode=\"layer\" inkscape:label=\"128 contour\" stroke=\"black\"  >\n<polyline id=\"0\" points=\"1.00,0.50 1.50,0.00 \" />\n<polyline id=\"1\" points=\"2.50,0.00 3.00,0.50 3.00,1.50 1.50,3.00 0.50,3.00 0.00,2.50 \" />\n<polyline id=\"2\" points=\"0.00,1.50 1.00,0.50 \" />\n<polyline id=\"3\" points=\"4.00,0.50 4.50,0.00 \" />\n<polyline id=\"4\" points=\"0.00,4.50 0.50,4.00 2.50,4.00 4.00,2.50 4.00,0.50 \" />\n<polygon id=\"0\"  points=\"5.00,4.50 5.50,4.00 6.00,4.50 6.00,5.50 5.50,6.00 4.50,6.00 4.00,5.50 5.00,4.50 \" />\n</g>\n<!-- 3 contours found at threshold 128, with length 1.00m -->\n<!-- Total contour length: 1.00m -->\n</g>\n</svg>\n",
		},
		{"tests/test4.png", "tests/test4-hc-t100,200m15pA4PC.svg", []int{100, 200}, 15, 0.0, "A4P", true, "",
			"<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<!-- tests/test4-hc-t100,200m15pA4PC.svg, created by hcontours.test version 0.1.2 -->\n<!-- Options used: infile: \"tests/test4.png\", width: 6, height: 4, thresholds: [100 200], tcount: -1, margin: 15.00, paper: \"A4P\", paperSize: {210.00, 297.00}, image: false, clip: true, debug: false, linewidth: 1.00, framewidth: 0.00, colours: \"\" -->\n<svg width=\"210mm\" height=\"297mm\" viewBox=\"0 0 210 297\" style=\"background-color:white\" xmlns=\"http://www.w3.org/2000/svg\" xmlns:inkscape=\"http://www.inkscape.org/namespaces/inkscape\" encoding=\"UTF-8\" >\n<g stroke=\"black\" stroke-width=\"0.0333\" stroke-linecap=\"round\" stroke-linejoin=\"round\" fill=\"none\" transform=\"translate(15.0000,88.5000) scale(30.0000)\">\n<defs><clipPath id=\"clip1\" ><rect id=\"cliprect\" width=\"5.9667\" height=\"3.9667\" x=\"0.0167\" y=\"0.0167\" /></clipPath></defs>\n<g inkscape:groupmode=\"layer\" inkscape:label=\"0 background\" stroke=\"black\"  >\n</g>\n<g inkscape:groupmode=\"layer\" inkscape:label=\"200 contour\" stroke=\"black\"  >\n<path id=\"0\" clip-path=\"url(#clip1)\"  d=\"M 0.72,0.50 L 1.50,-0.00 L 2.28,0.50 L 3.28,1.50 L 3.28,2.50 L 2.28,3.50 L 1.50,4.00 L 0.72,3.50 L 0.50,3.28 L -0.00,2.50 L -0.00,1.50 L 0.50,0.72 L 0.72,0.50 Z M 3.72,0.50 L 4.50,-0.00 L 5.50,-0.00 L 6.00,0.50 L 6.00,1.50 L 5.50,2.28 L 4.50,2.28 L 3.72,1.50 L 3.72,0.50 Z M 3.72,3.50 L 4.50,2.72 L 5.28,3.50 L 4.50,4.00 L 3.72,3.50 Z \" />\n</g>\n<g inkscape:groupmode=\"layer\" inkscape:label=\"100 contour\" stroke=\"black\"  >\n<path id=\"1\" clip-path=\"url(#clip1)\"  d=\"M 1.11,0.50 L 1.50,-0.00 L 1.89,0.50 L 2.89,1.50 L 2.89,2.50 L 1.89,3.50 L 1.50,4.00 L 1.11,3.50 L 0.50,2.89 L -0.00,2.50 L -0.00,1.50 L 0.50,1.11 L 1.11,0.50 Z M 4.11,0.50 L 4.50,-0.00 L 5.50,-0.00 L 6.00,0.50 L 6.00,1.50 L 5.50,1.89 L 4.50,1.89 L 4.11,1.50 L 4.11,0.50 Z M 4.11,3.50 L 4.50,3.11 L 4.89,3.50 L 4.50,4.00 L 4.11,3.50 Z \" />\n</g>\n<!-- 3 contours found at threshold 100, with length 0.58m -->\n<!-- 3 contours found at threshold 200, with length 0.68m -->\n<!-- Total contour length: 1.26m -->\n</g>\n</svg>\n",
		},
		{"tests/test7.png", "tests/test7-hc-t85,171m15pA4LCff7700-0077ff.svg", []int{85, 171}, 15, 0.0, "A4L", true, "ff7700-0077ff",
			"<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<!-- tests/test7-hc-t85,171m15pA4LCff7700-0077ff.svg, created by hcontours.test version 0.1.2 -->\n<!-- Options used: infile: \"tests/test7.png\", width: 5, height: 7, thresholds: [85 171], tcount: -1, margin: 15.00, paper: \"A4L\", paperSize: {297.00, 210.00}, image: false, clip: true, debug: false, linewidth: 1.00, framewidth: 0.00, colours: \"ff7700-0077ff\" -->\n<svg width=\"297mm\" height=\"210mm\" viewBox=\"0 0 297 210\" style=\"background-color:white\" xmlns=\"http://www.w3.org/2000/svg\" xmlns:inkscape=\"http://www.inkscape.org/namespaces/inkscape\" encoding=\"UTF-8\" >\n<g stroke=\"black\" stroke-width=\"0.0389\" stroke-linecap=\"round\" stroke-linejoin=\"round\" fill=\"none\" transform=\"translate(84.2143,15.0000) scale(25.7143)\">\n<defs><clipPath id=\"clip1\" ><rect id=\"cliprect\" width=\"4.9611\" height=\"6.9611\" x=\"0.0194\" y=\"0.0194\" /></clipPath></defs>\n<g inkscape:groupmode=\"layer\" inkscape:label=\"0 background\" stroke=\"black\" fill=\"#0077ff\" >\n<rect id=\"plotsize\" width=\"5\" height=\"7\" stroke=\"none\" />\n</g>\n<g inkscape:groupmode=\"layer\" inkscape:label=\"171 contour\" stroke=\"black\" fill=\"#7f7780\" >\n<path id=\"0\" clip-path=\"url(#clip1)\"  d=\"M 0.83,1.50 L 1.50,0.83 L 3.50,0.83 L 4.17,1.50 L 4.17,3.50 L 3.50,4.17 L 2.50,4.17 L 2.17,4.50 L 2.50,4.83 L 3.50,4.83 L 4.17,5.50 L 3.50,6.17 L 1.50,6.17 L 0.83,5.50 L 0.83,3.50 L 1.50,2.83 L 2.50,2.83 L 2.83,2.50 L 2.50,2.17 L 1.50,2.17 L 0.83,1.50 Z \" />\n</g>\n<g inkscape:groupmode=\"layer\" inkscape:label=\"85 contour\" stroke=\"black\" fill=\"#ff7700\" >\n<path id=\"1\" clip-path=\"url(#clip1)\"  d=\"M 1.17,1.50 L 1.50,1.17 L 3.50,1.17 L 3.83,1.50 L 3.83,3.50 L 3.50,3.83 L 2.50,3.83 L 1.83,4.50 L 2.50,5.17 L 3.50,5.17 L 3.83,5.50 L 3.50,5.83 L 1.50,5.83 L 1.17,5.50 L 1.17,3.50 L 1.50,3.17 L 2.50,3.17 L 3.17,2.50 L 2.50,1.83 L 1.50,1.83 L 1.17,1.50 Z \" />\n</g>\n<!-- 1 contours found at threshold 85, with length 0.50m -->\n<!-- 1 contours found at threshold 171, with length 0.55m -->\n<!-- Total contour length: 1.05m -->\n</g>\n</svg>\n",
		},
	}
	for _, td := range testdata {
		fmt.Printf("\t%s\n", td.infile)
		opts := OptsT{infile: td.infile, thresholds: td.thresholds, tcount: -1, margin: td.margin, framewidth: td.framewidth, paper: td.paper, clip: td.clip, linewidth: 1, colours: td.colours}
		parsePaperSize(&opts)
		svgFilename := createSVG(opts)
		if svgFilename != td.outfile {
			t.Errorf("Wrong filename for %s: wanted '%s' got '%s'\n", td.infile, td.outfile, svgFilename)
		}
		if td.wanted != "" {
			// read back the output
			bytes, err := os.ReadFile(svgFilename)
			if err != nil {
				t.Errorf("Can't read in the SVG file: %s", err)
			} else {
				got := string(bytes)
				if got != td.wanted {
					t.Errorf("Wrong result for %s\n\twanted '%s'\n\t   got '%s')\n", td.infile, td.wanted, got)
				}
			}
		}
	}
}

func TestCompress(t *testing.T) {
	fmt.Println("TestCompress")
	type testdataT struct {
		id     string
		orig   ContourT
		wanted ContourT
	}
	testdata := []testdataT{
		{"no compression",
			ContourT{{0.998, 1.500}, {1.500, 0.998}, {7.500, 0.998}, {8.002, 1.500}, {7.500, 2.002}, {1.500, 2.002}, {0.998, 1.500}},
			ContourT{{0.998, 1.500}, {1.500, 0.998}, {7.500, 0.998}, {8.002, 1.500}, {7.500, 2.002}, {1.500, 2.002}, {0.998, 1.500}},
		},
		{"lots of compression",
			ContourT{{0.998, 1.500}, {1.500, 0.998}, {2.500, 0.998}, {3.500, 0.998}, {4.500, 0.998}, {5.500, 0.998}, {6.500, 0.998}, {7.500, 0.998}, {8.002, 1.500},
				{7.500, 2.002}, {6.500, 2.002}, {5.500, 2.002}, {4.500, 2.002}, {3.500, 2.002}, {2.500, 2.002}, {1.500, 2.002}, {0.998, 1.500}},
			ContourT{{0.998, 1.500}, {1.500, 0.998}, {7.500, 0.998}, {8.002, 1.500}, {7.500, 2.002}, {1.500, 2.002}, {0.998, 1.500}},
		},
	}
	for _, td := range testdata {
		fmt.Printf("\t%s\n", td.id)
		got := td.orig.Compress()
		if !got.Equal(td.wanted) {
			t.Errorf("Wrong result for test '%s':\n\twanted: %v\n\t   got: %v\n", td.id, td.wanted, got)
		}
	}
}
