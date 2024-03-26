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
		{PointT{0, -1}, PointT{0, 0}, 200, 20, 80, 2, 2, Point64T{0.5, -0.5}},
		{PointT{2, 0}, PointT{1, 0}, 200, 20, 80, 2, 2, Point64T{2.5, 0.5}},
		{PointT{1, 2}, PointT{1, 1}, 200, 20, 80, 2, 2, Point64T{1.5, 2.5}},
		{PointT{-1, 1}, PointT{0, 1}, 200, 20, 80, 2, 2, Point64T{-0.5, 1.5}},
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
	testdata := []testdataT{
		{"tests/test0.png", ContourT{
			{0.498, 1}, {1, 0.498}, {2, 0.498}, {2.502, 1}, {2.502, 2}, {2, 2.502}, {1, 2.502}, {0.498, 2},
		}, PointT{1, 1}},
		{"tests/test1.png", ContourT{
			{0.498, 1}, {1, 0.498}, {2, 0.498}, {3, 0.498}, {3.502, 1}, {3, 1.502}, {2.502, 2}, {2, 2.502}, {1.502, 3}, {1, 3.502}, {0.498, 3}, {0.498, 2},
		}, PointT{1, 1}},
		{"tests/test4.png", ContourT{
			{0.998, 1.500}, {1.500, 0.998}, {2.500, 0.998}, {3.002, 1.500}, {3.002, 2.500}, {2.500, 3.002}, {1.500, 3.002}, {0.998, 2.500}, {0.998, 1.500},
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
	testdata := []testdataT{
		{"tests/test0.png", ContourS{{{0.498, 1.000}, {1.000, 0.498}, {2.000, 0.498}, {2.502, 1.000}, {2.502, 2.000}, {2.000, 2.502}, {1.000, 2.502}, {0.498, 2.000}}}, 1},
		{"tests/test1.png", ContourS{{{0.498, 1.000}, {1.000, 0.498}, {3.000, 0.498}, {3.502, 1.000}, {1.000, 3.502}, {0.498, 3.000}, {0.498, 2.000}}}, 1},
		{"tests/test2.png", ContourS{{{-0.502, 0.000}, {0.000, -0.502}, {1.000, -0.502}, {1.502, 0.000}, {0.000, 1.502}, {-0.502, 1.000}}, {{2.498, 2.000}, {3.000, 1.498}, {3.502, 2.000}, {3.502, 3.000}, {3.000, 3.502}, {2.000, 3.502}, {1.498, 3.000}, {2.000, 2.498}}}, 2},
		{"tests/test3.png", ContourS{{{0.498, 0.000}, {1.000, -0.502}, {2.000, -0.502}, {2.502, 0.000}, {2.502, 1.000}, {1.000, 2.502}, {0.000, 2.502}, {-0.502, 2.000}, {-0.502, 1.000}, {0.000, 0.498}}, {{3.498, 0.000}, {4.000, -0.502}, {7.000, -0.502}, {7.502, 0.000}, {7.502, 7.000}, {7.000, 7.502}, {0.000, 7.502}, {-0.502, 7.000}, {-0.502, 4.000}, {0.000, 3.498}, {2.000, 3.498}, {3.498, 2.000}, {3.498, 1.000}}, {{4.498, 4.000}, {5.000, 3.498}, {5.502, 4.000}, {5.502, 5.000}, {5.000, 5.502}, {4.000, 5.502}, {3.498, 5.000}, {4.000, 4.498}}}, 3},
		{"tests/test4.png", ContourS{{{0.498, 0.000}, {1.000, -0.502}, {2.502, 1.000}, {2.502, 2.000}, {1.000, 3.502}, {-0.502, 2.000}, {-0.502, 1.000}, {0.000, 0.498}}, {{3.498, 0.000}, {4.000, -0.502}, {5.000, -0.502}, {5.502, 0.000}, {5.502, 1.000}, {5.000, 1.502}, {4.000, 1.502}, {3.498, 1.000}}, {{3.498, 3.000}, {4.000, 2.498}, {4.502, 3.000}, {4.000, 3.502}}}, 3},
		{"tests/test5.png", ContourS{{{0.498, 1.000}, {1.000, 0.498}, {6.000, 0.498}, {6.502, 1.000}, {6.502, 6.000}, {6.000, 6.502}, {1.000, 6.502}, {0.498, 6.000}, {0.498, 2.000}}, {{4.498, 3.000}, {4.000, 2.502}, {3.000, 2.502}, {2.502, 3.000}, {2.502, 4.000}, {3.000, 4.498}, {4.000, 4.498}, {4.498, 4.000}}}, 2},
		{"tests/test6.png", ContourS{{{0.498, 1.000}, {1.000, 0.498}, {6.000, 0.498}, {6.502, 1.000}, {6.502, 6.000}, {6.000, 6.502}, {1.000, 6.502}, {0.498, 6.000}, {0.498, 2.000}}}, 1},
		// These two have non-closed thin lines -- the contour loops back to close itself:
		{"tests/test7.png", ContourS{{{0.498, 1.000}, {1.000, 0.498}, {3.000, 0.498}, {3.502, 1.000}, {3.502, 3.000}, {3.000, 3.502}, {2.000, 3.502}, {1.502, 4.000}, {2.000, 4.498}, {3.000, 4.498}, {3.502, 5.000}, {3.000, 5.502}, {1.000, 5.502}, {0.498, 5.000}, {0.498, 3.000}, {1.000, 2.498}, {2.000, 2.498}, {2.498, 2.000}, {2.000, 1.502}, {1.000, 1.502}}}, 1},
		{"tests/test8.png", ContourS{{{0.498, 1.000}, {1.000, 0.498}, {7.000, 0.498}, {7.502, 1.000}, {7.000, 1.502}, {1.000, 1.502}}, {{0.498, 3.000}, {1.000, 2.498}, {1.502, 3.000}, {1.502, 7.000}, {1.000, 7.502}, {0.498, 7.000}, {0.498, 4.000}}, {{2.498, 3.000}, {3.000, 2.498}, {7.502, 7.000}, {7.000, 7.502}, {3.000, 3.502}}}, 3},
		{"tests/example.png", nil, 10},
		{"tests/bottom.png", nil, 155},
	}
	for _, td := range testdata {
		fmt.Printf("\t%s\n", td.infile)
		img, width, height, err := loadImage(td.infile)
		if err != nil {
			t.Errorf("Input file %s not found\n", td.infile)
		}
		got := contourFinder(img, width, height, 128, nil)
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
	testdata := []testdataT{
		{"tests/test3.png", []int{128}, 15, "A4L", "<svg width=\"297mm\" height=\"210mm\" viewBox=\"0 0 297 210\" style=\"background-color:white\" xmlns=\"http://www.w3.org/2000/svg\" xmlns:inkscape=\"http://www.inkscape.org/namespaces/inkscape\" encoding=\"UTF-8\" >\n<g stroke=\"black\" stroke-width=\"0.1mm\" stroke-linecap=\"round\" stroke-linejoin=\"round\" fill=\"none\" transform=\"translate(58.5,15) scale(2.000)\">\n<g inkscape:groupmode=\"layer\" inkscape:label=\"1\" stroke=\"rgb(0%, 0%, 0%)\">\n<polygon points=\"0.50,0.00 1.00,-0.50 2.00,-0.50 2.50,0.00 2.50,1.00 1.00,2.50 0.00,2.50 -0.50,2.00 -0.50,1.00 0.00,0.50 \" />\n<polygon points=\"3.50,0.00 4.00,-0.50 7.00,-0.50 7.50,0.00 7.50,7.00 7.00,7.50 0.00,7.50 -0.50,7.00 -0.50,4.00 0.00,3.50 2.00,3.50 3.50,2.00 3.50,1.00 \" />\n<polygon points=\"4.50,4.00 5.00,3.50 5.50,4.00 5.50,5.00 5.00,5.50 4.00,5.50 3.50,5.00 4.00,4.50 \" />\n</g>\n</g>\n</svg>\n"},
	}
	for _, td := range testdata {
		fmt.Printf("\t%s\n", td.infile)
		opts := OptsT{infile: td.infile, thresholds: td.thresholds, margin: td.margin, paper: td.paper}
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

/*
func TestCompressMoves(t *testing.T) {
	fmt.Println("TestCompressMoves")
	type testdataT struct {
		orig   []int
		width  int
		wanted []int
	}
	testdata := []testdataT{
		{[]int{1, 2, 10, 17, 16, 8}, 8, []int{1, 2, 10, 17, 16, 8}}, // no compression
		{[]int{4, 5, 6, 7, 15, 23, 31, 39, 47, 55, 63, 62, 61, 60, 59, 58, 57, 56, 48, 40, 32, 33, 34, 27, 20, 12}, 8, []int{4, 7, 63, 56, 32, 34, 20, 12}},
	}
	for i, td := range testdata {
		fmt.Printf("\t%d\n", i)
		got := compressMoves(td.orig, td.width)
		if !reflect.DeepEqual(got, td.wanted) {
			t.Errorf("Wrong result for test %d (wanted %v  got %v)\n", i, td.wanted, got)
		}
	}
}
*/
