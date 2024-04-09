// types.go -- types and constants for hcontours.go

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
func (p1 Point64T) Equal(p2 Point64T) bool {
	// Points don't have to be precisely equal for our purposes
	return math.Abs(p1.x-p2.x) < 0.001 && math.Abs(p1.y-p2.y) < 0.001
}
func (p1 Point64T) RelAngle(p2 Point64T) float64 {
	// Calculate the angle from p1 to p2, in radians widdershins.
	return math.Atan2(float64(p2.y-p1.y), float64(p2.x-p1.x))
}
func (p1 Point64T) Distance(p2 Point64T) float64 {
	return float64(math.Hypot(float64(p1.x-p2.x), float64(p1.y-p2.y)))
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

type RectangleT struct {
	width  float64
	height float64
}

func (r RectangleT) String() string {
	return fmt.Sprintf("{%.4f, %.4f}", r.width, r.height)
}
func (r1 RectangleT) Equal(r2 RectangleT) bool {
	return math.Abs(r1.width-r2.width) < 0.001 && math.Abs(r1.height-r2.height) < 0.001
}

var paperSizes = map[string]RectangleT{
	"A4L": RectangleT{width: 297, height: 210},
	"A4P": RectangleT{width: 210, height: 297},
	"A3L": RectangleT{width: 420, height: 297},
	"A3P": RectangleT{width: 297, height: 420},
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
	paperSize  RectangleT
	image      bool
	clip       bool
	debug      bool
	linewidth  float64
	framewidth float64
	colours    string // two hex colours, e.g. "0033ff,0c4088"
}

func (o OptsT) String() string {
	return fmt.Sprintf("infile: \"%s\", width: %d, height: %d, thresholds: %v, tcount: %d, margin: %.2f, paper: \"%s\", paperSize: {%.2f, %.2f}, image: %t, clip: %t, debug: %t, linewidth: %.2f, framewidth: %.2f, colours: \"%s\"", o.infile, o.width, o.height, o.thresholds, o.tcount, o.margin, o.paper, o.paperSize.width, o.paperSize.height, o.image, o.clip, o.debug, o.linewidth, o.framewidth, o.colours)
}

const white = 0xff
const black = 0x00
