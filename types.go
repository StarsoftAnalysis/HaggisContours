// types.go -- types and constants for hcontours.go

package main

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

type ContourT []Point64T

// Options and derived things
type OptsT struct {
	infile     string
	width      int
	height     int
	thresholds []int
	margin     float64
	paper      string
}

const white = 0xff
const black = 0x00
