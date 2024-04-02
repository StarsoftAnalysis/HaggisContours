package main

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
		frame      bool
		image      bool
		clip       bool
		dev        bool
		linewidth  float64
		framewidth float64
	*/
	testdata := []testdataT{
		{OptsT{"file1.png", 100, 200, []int{44, 55}, -1, 15.0, "5x7", RectangleT{0, 0}, false, true, false, true, 1.0, 2.0}, "file1-hc-t44,55m15p5x7ID.svg"},
		{OptsT{"file1.png", 100, 200, []int{}, 3, 10.3, "200x300", RectangleT{0, 0}, true, false, true, false, 1.0, 2.0}, "file1-hc-T3m10.3p200x300FC.svg"},
	}
	for i, td := range testdata {
		filename := buildSVGfilename(td.opts)
		if filename != td.wanted {
			t.Errorf("(%d) Wrong filename: wanted '%s' got '%s'\n", i, td.wanted, filename)
		}
	}
}

func TestTraceContour(t *testing.T) {
	fmt.Println("TestTraceContour")
	type testdataT struct {
		infile  string
		contour ContourT
		start   PointT
	}
	testdata := []testdataT{ // Don't forget that no compression happens for these cases
		{"tests/test0.png", ContourT{
			{0.998, 1.500}, {1.500, 0.998}, {2.500, 0.998}, {3.002, 1.500}, {3.002, 2.500}, {2.500, 3.002}, {1.500, 3.002}, {0.998, 2.500},
			{0.998, 1.500},
		}, PointT{1, 1}},
		{"tests/test1.png", ContourT{
			{0.998, 1.500}, {1.500, 0.998}, {2.500, 0.998}, {3.500, 0.998}, {4.002, 1.500}, {3.500, 2.002}, {3.002, 2.500}, {2.500, 3.002},
			{2.002, 3.500}, {1.500, 4.002}, {0.998, 3.500}, {0.998, 2.500}, {0.998, 1.500},
		}, PointT{1, 1}},
		{"tests/test4.png", ContourT{
			{0.998, 0.500}, {1.500, -0.001}, {2.002, 0.500}, {2.500, 0.998}, {3.002, 1.500}, {3.002, 2.500}, {2.500, 3.002}, {2.002, 3.500},
			{1.500, 4.001}, {0.998, 3.500}, {0.500, 3.002}, {-0.001, 2.500}, {-0.001, 1.500}, {0.500, 0.998}, {0.998, 0.500},
		}, PointT{1, 0}},
	}
	for _, td := range testdata {
		fmt.Printf("\t%s\n", td.infile)
		img, width, height, err := loadImage(td.infile)
		if err != nil {
			t.Errorf("Input file %s not found\n", td.infile)
		}
		got, _ := traceContour(img, width, height, 128, td.start, nil)
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
	}
	testdata := []testdataT{ // Don't forget that no compression happens for these cases
		{"tests/test0.png", ContourS{
			{{0.998, 1.500}, {1.500, 0.998}, {2.500, 0.998}, {3.002, 1.500}, {3.002, 2.500}, {2.500, 3.002}, {1.500, 3.002}, {0.998, 2.500}, {0.998, 1.500}},
		}, 1},
		{"tests/test1.png", ContourS{
			{{0.998, 1.500}, {1.500, 0.998}, {2.500, 0.998}, {3.500, 0.998}, {4.002, 1.500}, {3.500, 2.002}, {3.002, 2.500}, {2.500, 3.002}, {2.002, 3.500},
				{1.500, 4.002}, {0.998, 3.500}, {0.998, 2.500}, {0.998, 1.500}},
		}, 1},
		{"tests/test2.png", ContourS{
			{{-0.001, 0.500}, {0.500, -0.001}, {1.500, -0.001}, {2.002, 0.500}, {1.500, 1.002}, {1.002, 1.500}, {0.500, 2.002}, {-0.001, 1.500}, {-0.001, 0.500}},
			{{2.998, 2.500}, {3.500, 1.998}, {4.001, 2.500}, {4.001, 3.500}, {3.500, 4.001}, {2.500, 4.001}, {1.998, 3.500}, {2.500, 2.998}, {2.998, 2.500}},
		}, 2},
		{"tests/test3.png", ContourS{
			{{0.998, 0.500}, {1.500, -0.001}, {2.500, -0.001}, {3.002, 0.500}, {3.002, 1.500}, {2.500, 2.002}, {2.002, 2.500}, {1.500, 3.002}, {0.500, 3.002},
				{-0.001, 2.500}, {-0.001, 1.500}, {0.500, 0.998}, {0.998, 0.500}},
			{{3.998, 0.500}, {4.500, -0.001}, {5.500, -0.001}, {6.500, -0.001}, {7.500, -0.001}, {8.001, 0.500}, {8.001, 1.500}, {8.001, 2.500}, {8.001, 3.500},
				{8.001, 4.500}, {8.001, 5.500}, {8.001, 6.500}, {8.001, 7.500}, {7.500, 8.001}, {6.500, 8.001}, {5.500, 8.001}, {4.500, 8.001}, {3.500, 8.001},
				{2.500, 8.001}, {1.500, 8.001}, {0.500, 8.001}, {-0.001, 7.500}, {-0.001, 6.500}, {-0.001, 5.500}, {-0.001, 4.500}, {0.500, 3.998}, {1.500, 3.998},
				{2.500, 3.998}, {2.998, 3.500}, {3.500, 2.998}, {3.998, 2.500}, {3.998, 1.500}, {3.998, 0.500}},
			{{4.998, 4.500}, {5.500, 3.998}, {6.002, 4.500}, {6.002, 5.500}, {5.500, 6.002}, {4.500, 6.002}, {3.998, 5.500}, {4.500, 4.998}, {4.998, 4.500}},
		}, 3},
		{"tests/test4.png", ContourS{
			{{0.998, 0.500}, {1.500, -0.001}, {2.002, 0.500}, {2.500, 0.998}, {3.002, 1.500}, {3.002, 2.500}, {2.500, 3.002}, {2.002, 3.500}, {1.500, 4.001},
				{0.998, 3.500}, {0.500, 3.002}, {-0.001, 2.500}, {-0.001, 1.500}, {0.500, 0.998}, {0.998, 0.500}},
			{{3.998, 0.500}, {4.500, -0.001}, {5.500, -0.001}, {6.001, 0.500}, {6.001, 1.500}, {5.500, 2.002}, {4.500, 2.002}, {3.998, 1.500}, {3.998, 0.500}},
			{{3.998, 3.500}, {4.500, 2.998}, {5.002, 3.500}, {4.500, 4.001}, {3.998, 3.500}},
		}, 3},
		{"tests/test5.png", ContourS{
			{{0.998, 1.500}, {1.500, 0.998}, {2.500, 0.998}, {3.500, 0.998}, {4.500, 0.998}, {5.500, 0.998}, {6.500, 0.998}, {7.002, 1.500}, {7.002, 2.500},
				{7.002, 3.500}, {7.002, 4.500}, {7.002, 5.500}, {7.002, 6.500}, {6.500, 7.002}, {5.500, 7.002}, {4.500, 7.002}, {3.500, 7.002}, {2.500, 7.002},
				{1.500, 7.002}, {0.998, 6.500}, {0.998, 5.500}, {0.998, 4.500}, {0.998, 3.500}, {0.998, 2.500}, {0.998, 1.500}},
			{{4.998, 3.500}, {4.500, 3.002}, {3.500, 3.002}, {3.002, 3.500}, {3.002, 4.500}, {3.500, 4.998}, {4.500, 4.998}, {4.998, 4.500}, {4.998, 3.500}},
		}, 2},
		{"tests/test6.png", ContourS{
			{{0.998, 1.500}, {1.500, 0.998}, {2.500, 0.998}, {3.500, 0.998}, {4.500, 0.998}, {5.500, 0.998}, {6.500, 0.998}, {7.002, 1.500}, {7.002, 2.500},
				{7.002, 3.500}, {7.002, 4.500}, {7.002, 5.500}, {7.002, 6.500}, {6.500, 7.002}, {5.500, 7.002}, {4.500, 7.002}, {3.500, 7.002}, {2.500, 7.002},
				{1.500, 7.002}, {0.998, 6.500}, {0.998, 5.500}, {0.998, 4.500}, {0.998, 3.500}, {0.998, 2.500}, {0.998, 1.500}},
		}, 1},
		// These two have non-closed thin lines -- the contour loops back to close itself:
		{"tests/test7.png", ContourS{
			{{0.998, 1.500}, {1.500, 0.998}, {2.500, 0.998}, {3.500, 0.998}, {4.002, 1.500}, {4.002, 2.500}, {4.002, 3.500}, {3.500, 4.002}, {2.500, 4.002},
				{2.002, 4.500}, {2.500, 4.998}, {3.500, 4.998}, {4.002, 5.500}, {3.500, 6.002}, {2.500, 6.002}, {1.500, 6.002}, {0.998, 5.500}, {0.998, 4.500},
				{0.998, 3.500}, {1.500, 2.998}, {2.500, 2.998}, {2.998, 2.500}, {2.500, 2.002}, {1.500, 2.002}, {0.998, 1.500}},
		}, 1},
		{"tests/test8.png", ContourS{
			{{0.998, 1.500}, {1.500, 0.998}, {2.500, 0.998}, {3.500, 0.998}, {4.500, 0.998}, {5.500, 0.998}, {6.500, 0.998}, {7.500, 0.998}, {8.002, 1.500},
				{7.500, 2.002}, {6.500, 2.002}, {5.500, 2.002}, {4.500, 2.002}, {3.500, 2.002}, {2.500, 2.002}, {1.500, 2.002}, {0.998, 1.500}},
			{{0.998, 3.500}, {1.500, 2.998}, {2.002, 3.500}, {2.002, 4.500}, {2.002, 5.500}, {2.002, 6.500}, {2.002, 7.500}, {1.500, 8.002}, {0.998, 7.500},
				{0.998, 6.500}, {0.998, 5.500}, {0.998, 4.500}, {0.998, 3.500}},
			{{2.998, 3.500}, {3.500, 2.998}, {4.002, 3.500}, {4.500, 3.998}, {5.002, 4.500}, {5.500, 4.998}, {6.002, 5.500}, {6.500, 5.998}, {7.002, 6.500},
				{7.500, 6.998}, {8.002, 7.500}, {7.500, 8.002}, {6.998, 7.500}, {6.500, 7.002}, {5.998, 6.500}, {5.500, 6.002}, {4.998, 5.500}, {4.500, 5.002},
				{3.998, 4.500}, {3.500, 4.002}, {2.998, 3.500}},
		}, 3},
		{"tests/example.png", nil, 10},
		{"tests/bottom.png", nil, 155},
	}
	for _, td := range testdata {
		fmt.Printf("\t%s\n", td.infile)
		img, width, height, err := loadImage(td.infile)
		if err != nil {
			t.Errorf("Input file %s not found\n", td.infile)
		}
		got := contourFinder(img, width, height, 128, false, nil)
		if len(got) != td.count {
			t.Errorf("Wrong result for %s (wanted length %v  got %v)\n", td.infile, td.count, len(got))
		}
		if td.contours != nil {
			equal := true
			for i := 0; equal && i < len(got); i++ {
				if !got[i].Equal(td.contours[i]) {
					equal = false
				}
			}
			if !equal {
				t.Errorf("Wrong result for %s\n\twanted %v\n\t   got %v\n", td.infile, td.contours, got)
			}
		}
	}
}

func TestCalcSizes(t *testing.T) {
	fmt.Println("TestCalcSizes")
	type testdataT struct {
		plot      RectangleT
		margin    float64
		paper     RectangleT
		translate RectangleT
		scale     float64
	}
	testdata := []testdataT{
		{RectangleT{400, 400}, 00, RectangleT{400, 400}, RectangleT{0, 0}, 1},
		{RectangleT{100, 200}, 50, RectangleT{400, 400}, RectangleT{125, 50}, 1.5},
		{RectangleT{600, 200}, 40, RectangleT{400, 400}, RectangleT{40.0000, 146.6667}, 0.5333},
		{RectangleT{600, 400}, 15, RectangleT{297, 210}, RectangleT{15, 16}, 0.445},
		{RectangleT{6, 4}, 15, RectangleT{297, 210}, RectangleT{15, 16}, 44.5},
	}
	for i, td := range testdata {
		translate, scale := calcSizes(td.plot, td.margin, td.paper)
		if !translate.Equal(td.translate) || !almostEqual(scale, td.scale, 0.001) {
			t.Errorf("(%d) Wrong result: wanted %v, %g   got %v, %g", i, td.translate, td.scale, translate, scale)
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
		frame      bool
		paper      string
		clip       bool
		wanted     string
	}
	testdata := []testdataT{ // Compression is done for SVG contours
		{"tests/test3.png", "tests/test3-hc-t128m15pA4L.svg", []int{128}, 15, true, "A4L", false,
			"<svg width=\"297mm\" height=\"210mm\" viewBox=\"0 0 297 210\" style=\"background-color:white\" xmlns=\"http://www.w3.org/2000/svg\" xmlns:inkscape=\"http://www.inkscape.org/namespaces/inkscape\" encoding=\"UTF-8\" >\n<g stroke=\"black\" stroke-width=\"0.0444\" stroke-linecap=\"round\" stroke-linejoin=\"round\" fill=\"none\" transform=\"translate(58.5,15) scale(22.5000)\">\n<g inkscape:groupmode=\"layer\" inkscape:label=\"128\" stroke=\"rgb(0%, 0%, 0%)\">\n<polyline points=\"1.00,0.50 1.50,0.00 \" />\n<polyline points=\"2.50,0.00 3.00,0.50 3.00,1.50 1.50,3.00 0.50,3.00 0.00,2.50 \" />\n<polyline points=\"0.00,1.50 1.00,0.50 \" />\n<polyline points=\"4.00,0.50 4.50,0.00 \" />\n<polyline points=\"0.00,4.50 0.50,4.00 2.50,4.00 4.00,2.50 4.00,0.50 \" />\n<polygon  points=\"5.00,4.50 5.50,4.00 6.00,4.50 6.00,5.50 5.50,6.00 4.50,6.00 4.00,5.50 5.00,4.50 \" />\n</g>\n</g>\n</svg>\n",
		},
		{"tests/test4.png", "tests/test4-hc-t100,200m15pA4PC.svg", []int{100, 200}, 15, false, "A4P", true,
			"<svg width=\"210mm\" height=\"297mm\" viewBox=\"0 0 210 297\" style=\"background-color:white\" xmlns=\"http://www.w3.org/2000/svg\" xmlns:inkscape=\"http://www.inkscape.org/namespaces/inkscape\" encoding=\"UTF-8\" >\n<g stroke=\"black\" stroke-width=\"0.0333\" stroke-linecap=\"round\" stroke-linejoin=\"round\" fill=\"none\" transform=\"translate(15,88.5) scale(30.0000)\">\n<clipPath id=\"clip1\" ><rect width=\"6\" height=\"4\" /></clipPath>\n<g inkscape:groupmode=\"layer\" inkscape:label=\"100\" stroke=\"rgb(0%, 0%, 0%)\">\n<path clip-path=\"url(#clip1)\"  d=\"M 1.11,0.50 L 1.50,-0.00 L 1.89,0.50 L 2.89,1.50 L 2.89,2.50 L 1.89,3.50 L 1.50,4.00 L 1.11,3.50 L 0.50,2.89 L -0.00,2.50 L -0.00,1.50 L 0.50,1.11 L 1.11,0.50 Z M 4.11,0.50 L 4.50,-0.00 L 5.50,-0.00 L 6.00,0.50 L 6.00,1.50 L 5.50,1.89 L 4.50,1.89 L 4.11,1.50 L 4.11,0.50 Z M 4.11,3.50 L 4.50,3.11 L 4.89,3.50 L 4.50,4.00 L 4.11,3.50 Z \" />\n</g>\n<g inkscape:groupmode=\"layer\" inkscape:label=\"200\" stroke=\"rgb(0%, 0%, 0%)\">\n<path clip-path=\"url(#clip1)\"  d=\"M 0.72,0.50 L 1.50,-0.00 L 2.28,0.50 L 3.28,1.50 L 3.28,2.50 L 2.28,3.50 L 1.50,4.00 L 0.72,3.50 L 0.50,3.28 L -0.00,2.50 L -0.00,1.50 L 0.50,0.72 L 0.72,0.50 Z M 3.72,0.50 L 4.50,-0.00 L 5.50,-0.00 L 6.00,0.50 L 6.00,1.50 L 5.50,2.28 L 4.50,2.28 L 3.72,1.50 L 3.72,0.50 Z M 3.72,3.50 L 4.50,2.72 L 5.28,3.50 L 4.50,4.00 L 3.72,3.50 Z \" />\n</g>\n</g>\n</svg>\n",
		},
	}
	for _, td := range testdata {
		fmt.Printf("\t%s\n", td.infile)
		opts := OptsT{infile: td.infile, thresholds: td.thresholds, tcount: -1, margin: td.margin, frame: td.frame, framewidth: 2, paper: td.paper, clip: td.clip, linewidth: 1}
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
