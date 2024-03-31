// types.go -- types and constants for hcontours.go

package main

import (
	"fmt"
	"math"
	"strings"
)

// Direction is a number from 0 to 7 like this:
// 0  1  2
// 7  p  3
// 6  5  4
type DirectionT int

const approachDir = 3 // +v x direction, determined by the for x; for y logic in contourFinder()

func (dir *DirectionT) TurnLeft() {
	*dir = (*dir + 6) % 8
}
func (dir *DirectionT) TurnRight() {
	*dir = (*dir + 2) % 8
}

type PointT struct {
	x, y int
}

// Relative coordinates of neighbours in same order as Direction (see above)
var neighbourOffset = [8]PointT{
	{-1, -1}, {0, -1}, {1, -1}, {1, 0}, {1, 1}, {0, 1}, {-1, 1}, {-1, 0},
}

func (p1 PointT) Equal(p2 PointT) bool {
	return p1.x == p2.x && p1.y == p2.y
}
func (p1 PointT) Plus(p2 PointT) PointT {
	return PointT{p1.x + p2.x, p1.y + p2.y}
}
func (p PointT) Step(dir DirectionT) PointT {
	return p.Plus(neighbourOffset[dir])
}
func (p PointT) Backstep(dir DirectionT) PointT {
	return p.Plus(neighbourOffset[(dir+4)%8])
}

type Point64T struct {
	x, y float64
}

func (p Point64T) String() string {
	return fmt.Sprintf("{%.3f, %.3f}", p.x, p.y)
}

// Points don't have to be precisely equal for our purposes
func (p1 Point64T) Equal(p2 Point64T) bool {
	return math.Abs(p1.x-p2.x) < 0.001 && math.Abs(p1.y-p2.y) < 0.001
}

// Calculate the angle from p1 to p2, in radians widdershins.
func (p1 Point64T) RelAngle(p2 Point64T) float64 {
	return math.Atan2(float64(p2.y-p1.y), float64(p2.x-p1.x))
}

type ContourT []Point64T

func (c1 ContourT) Equal(c2 ContourT) bool {
	if len(c1) != len(c2) {
		return false
	}
	for i := range c1 {
		if !c1[i].Equal(c2[i]) {
			return false
		}
	}
	return true
}
func (c ContourT) String() string {
	s := make([]string, len(c))
	for i, p := range c {
		s[i] = fmt.Sprintf("%v", p)
	}
	return "{" + strings.Join(s, ", ") + "}"
}

// Simplify contour by combining consecutive moves in the same direction.
func (c ContourT) Compress() ContourT {
	lenc := len(c)
	if lenc < 3 {
		return c
	}
	var cc = make(ContourT, 0, lenc/2) // optimistic guess on the amount of compression
	p1 := c[0]
	cc = append(cc, p1)
	i := 1
	p2 := c[i]
	p3 := c[i+1]
	dir1 := p1.RelAngle(p2) // calculate angle from one point to the next
	for i < lenc-1 {
		if p2.Equal(p1) {
			// drop non-moves
		} else {
			dir2 := p2.RelAngle(p3)
			if sameAngle(dir1, dir2) {
				// do nothing: p1 and dir1 stay the same
			} else {
				// new direction -- add the point to the compressed array
				cc = append(cc, p2)
				p1 = p2
				dir1 = dir2
			}
		}
		i += 1
		p2 = p3
		if i+1 < lenc {
			p3 = c[i+1]
		}
	}
	cc = append(cc, c[i]) // don't forget the last point
	//fmt.Printf("cC: reduced len from %d to %d\n", lenc, len(cc))
	return cc
}

type ContourS []ContourT

func (cs ContourS) String() string {
	s := make([]string, len(cs))
	for i, c := range cs {
		s[i] = fmt.Sprintf("%v", c)
	}
	return "{" + strings.Join(s, ", ") + "}"
}

type PaperSizeT struct {
	width  float64
	height float64
}

var paperSizes = map[string]PaperSizeT{
	"A4L": PaperSizeT{width: 297, height: 210},
	"A4P": PaperSizeT{width: 210, height: 297},
	"A3L": PaperSizeT{width: 420, height: 297},
	"A3P": PaperSizeT{width: 297, height: 420},
}

// Options and derived things
type OptsT struct {
	infile     string
	width      int
	height     int
	thresholds []int
	tcount     int
	margin     float64
	paper      string
	paperSize  PaperSizeT
	frame      bool
	image      bool
	clip       bool
	dev        bool
	linewidth  float64
	framewidth float64
}

const white = 0xff
const black = 0x00
