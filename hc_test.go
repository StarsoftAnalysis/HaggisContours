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

func TestTraceContour(t *testing.T) {
	fmt.Println("TestTraceContour")

	type testdataT struct {
		infile  string
		contour ContourT
		start   PointT
	}
	testdata := []testdataT{ // Don't forget that no compression happens for these cases
		{"tests/test0.png", ContourT{
			{0.998, 1.500}, {1.500, 0.998}, {2.500, 0.998}, {3.002, 1.500}, {3.002, 2.500}, {2.500, 3.002}, {1.500, 3.002}, {0.998, 2.500}, {0.998, 1.500},
		}, PointT{1, 1}},
		{"tests/test1.png", ContourT{
			{0.998, 1.500}, {1.500, 0.998}, {2.500, 0.998}, {3.500, 0.998}, {4.002, 1.500}, {3.500, 2.002}, {3.002, 2.500}, {2.500, 3.002}, {2.002, 3.500}, {1.500, 4.002}, {0.998, 3.500}, {0.998, 2.500}, {0.998, 1.500},
		}, PointT{1, 1}},
		{"tests/test4.png", ContourT{
			{0.998, 0.500}, {1.500, -0.001}, {2.002, 0.500}, {2.500, 0.998}, {3.002, 1.500}, {3.002, 2.500}, {2.500, 3.002}, {2.002, 3.500}, {1.500, 4.001}, {0.998, 3.500}, {0.500, 3.002}, {-0.001, 2.500}, {-0.001, 1.500}, {0.500, 0.998}, {0.998, 0.500},
		}, PointT{1, 0}},
	}
	for _, td := range testdata {
		fmt.Printf("\t%s\n", td.infile)
		img, width, height, err := loadImage(td.infile)
		if err != nil {
			t.Errorf("Input file %s not found\n", td.infile)
		}
		//		got, _ := traceContour(img, width, height, 128, td.start, nil)
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

func TestCreateSVG(t *testing.T) {
	fmt.Println("TestCreateSVG")
	type testdataT struct {
		infile     string
		thresholds []int
		margin     float64
		paper      string
		wanted     string
	}
	testdata := []testdataT{ // Compression is done for SVG contours
		{"tests/test3.png", []int{128}, 15, "A4L",
			"<svg width=\"297mm\" height=\"210mm\" viewBox=\"0 0 297 210\" style=\"background-color:white\" xmlns=\"http://www.w3.org/2000/svg\" xmlns:inkscape=\"http://www.inkscape.org/namespaces/inkscape\" encoding=\"UTF-8\" >\n<g stroke=\"black\" stroke-width=\"0.00mm\" stroke-linecap=\"round\" stroke-linejoin=\"round\" fill=\"none\" transform=\"translate(58.5,15) scale(8.000)\">\n<g inkscape:groupmode=\"layer\" inkscape:label=\"1\" stroke=\"rgb(0%, 0%, 0%)\">\n<polyline vector-effect=\"non-scaling-stroke\" points=\"1.00,0.50 1.50,0.00 \" />\n<polyline vector-effect=\"non-scaling-stroke\" points=\"2.50,0.00 3.00,0.50 3.00,1.50 1.50,3.00 0.50,3.00 0.00,2.50 \" />\n<polyline vector-effect=\"non-scaling-stroke\" points=\"0.00,1.50 1.00,0.50 \" />\n<polyline vector-effect=\"non-scaling-stroke\" points=\"4.00,0.50 4.50,0.00 \" />\n<polyline vector-effect=\"non-scaling-stroke\" points=\"0.00,4.50 0.50,4.00 2.50,4.00 4.00,2.50 4.00,0.50 \" />\n<polygon vector-effect=\"non-scaling-stroke\"  points=\"5.00,4.50 5.50,4.00 6.00,4.50 6.00,5.50 5.50,6.00 4.50,6.00 4.00,5.50 \" />\n</g>\n</g>\n</svg>\n",
		},
	}
	for _, td := range testdata {
		fmt.Printf("\t%s\n", td.infile)
		opts := OptsT{infile: td.infile, thresholds: td.thresholds, margin: td.margin, paper: td.paper}
		parsePaperSize(&opts)
		svgFilename := createSVG(opts)
		// read back the output
		bytes, err := os.ReadFile(svgFilename)
		if err != nil {
			t.Errorf("Can't read in the SVG file: %s", err)
		}
		got := string(bytes)
		if got != td.wanted {
			t.Errorf("Wrong result for %s\n\twanted '%s'\n\t   got '%s')\n", td.infile, td.wanted, got)
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
