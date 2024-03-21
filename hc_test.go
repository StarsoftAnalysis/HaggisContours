package main

import (
	"fmt"
	"testing"
)

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
			{0.498, 0}, {1, -0.502}, {1.502, 0}, {2, 0.498}, {2.502, 1}, {2.502, 2}, {2, 2.502}, {1.502, 3}, {1, 3.502}, {0.498, 3}, {0, 2.502}, {-0.502, 2}, {-0.502, 1}, {0, 0.498},
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
			t.Errorf("Wrong result for %s/%v (wanted=%v  got %v)\n", td.infile, td.start, td.contour, got)
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
		{"tests/test1.png", ContourS{{{0.498, 1.000}, {1.000, 0.498}, {2.000, 0.498}, {3.000, 0.498}, {3.502, 1.000}, {3.000, 1.502}, {2.502, 2.000}, {2.000, 2.502}, {1.502, 3.000}, {1.000, 3.502}, {0.498, 3.000}, {0.498, 2.000}}}, 1},
		{"tests/test2.png", ContourS{{{-0.502, 0.000}, {0.000, -0.502}, {1.000, -0.502}, {1.502, 0.000}, {1.000, 0.502}, {0.502, 1.000}, {0.000, 1.502}, {-0.502, 1.000}}, {{2.498, 2.000}, {3.000, 1.498}, {3.502, 2.000}, {3.502, 3.000}, {3.000, 3.502}, {2.000, 3.502}, {1.498, 3.000}, {2.000, 2.498}}}, 2},
		{"tests/test3.png", ContourS{{{0.498, 0.000}, {1.000, -0.502}, {2.000, -0.502}, {2.502, 0.000}, {2.502, 1.000}, {2.000, 1.502}, {1.502, 2.000}, {1.000, 2.502}, {0.000, 2.502}, {-0.502, 2.000}, {-0.502, 1.000}, {0.000, 0.498}}, {{3.498, 0.000}, {4.000, -0.502}, {5.000, -0.502}, {6.000, -0.502}, {7.000, -0.502}, {7.502, 0.000}, {7.502, 1.000}, {7.502, 2.000}, {7.502, 3.000}, {7.502, 4.000}, {7.502, 5.000}, {7.502, 6.000}, {7.502, 7.000}, {7.000, 7.502}, {6.000, 7.502}, {5.000, 7.502}, {4.000, 7.502}, {3.000, 7.502}, {2.000, 7.502}, {1.000, 7.502}, {0.000, 7.502}, {-0.502, 7.000}, {-0.502, 6.000}, {-0.502, 5.000}, {-0.502, 4.000}, {0.000, 3.498}, {1.000, 3.498}, {2.000, 3.498}, {2.498, 3.000}, {3.000, 2.498}, {3.498, 2.000}, {3.498, 1.000}}, {{4.498, 4.000}, {5.000, 3.498}, {5.502, 4.000}, {5.502, 5.000}, {5.000, 5.502}, {4.000, 5.502}, {3.498, 5.000}, {4.000, 4.498}}}, 3},
		//	{"tests/test4.png", ContourS{{1, 8, 14, 19, 12, 6}, {4, 5, 11, 10}, {22}}, 3},
		//	{"tests/test5.png", ContourS{{9, 10, 11, 12, 13, 14, 22, 30, 38, 46, 54, 53, 52, 51, 50, 49, 41, 33, 25, 17}, {29, 20, 19, 26, 34, 43, 44, 37}}, 2},
		//	{"tests/test6.png", ContourS{{9, 10, 11, 12, 13, 14, 22, 30, 38, 46, 54, 53, 52, 51, 50, 49, 41, 33, 25, 17}}, 1},
		//	// These two have non-closed thin lines -- the contour loops back to close itself:
		//	{"tests/test7.png", ContourS{{6, 7, 8, 13, 18, 17, 21, 27, 28, 27, 26, 21, 16, 17, 13, 7}}, 1},
		//	{"tests/test8.png", ContourS{{10, 11, 12, 13, 14, 15, 16, 15, 14, 13, 12, 11}, {28, 37, 46, 55, 64, 55, 46, 37}, {30, 40, 50, 60, 70, 60, 50, 40}}, 3},
		//	{"tests/example.png", nil, 10},
		//	{"tests/bottom.png", nil, 156},
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

/*
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
		{"tests/test3.png", []int{128}, 15, "A4L", "<svg width=\"297mm\" height=\"210mm\" viewBox=\"0 0 297 210\" style=\"background-color:white\" xmlns=\"http://www.w3.org/2000/svg\" xmlns:inkscape=\"http://www.inkscape.org/namespaces/inkscape\" encoding=\"UTF-8\" >\n<g stroke=\"black\" stroke-width=\"0.1mm\" stroke-linecap=\"round\" stroke-linejoin=\"round\" fill=\"none\" transform=\"translate(58.5,15) scale(2.000)\">\n<g inkscape:groupmode=\"layer\" inkscape:label=\"1\" stroke=\"rgb(0%, 0%, 0%)\">\n<polygon points=\"1,0 2,0 2,1 1,2 0,2 0,1 \" />\n<polygon points=\"4,0 7,0 7,7 0,7 0,4 2,4 4,2 4,1 \" />\n<polygon points=\"5,4 5,5 4,5 \" />\n</g>\n</g>\n</svg>\n"},
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
			t.Errorf("Wrong result for %s (wanted '%s'  got '%s')\n", td.infile, td.wanted, got)
		}
	}
}

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
