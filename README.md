# HaggisContours

A method of converting a heightmap (or similar input) to a set of contours, using the wild haggis algorithm.

![Go workflow](https://github.com/starsoftanalysis/haggiscontours/actions/workflows/go.yml/badge.svg)

This is a Go programme for extacting contours from a [heightmap](https://en.wikipedia.org/wiki/Heightmap), or from a similar data source such as an image.

It uses a simple algorithm, based on the behaviour of wild haggis on the mountains of Scotland.  As is [well known](https://wildhaggis.com/),
dextrous wild haggis have their left legs longer than their right legs, which makes it easy for them to run clockwise
around hillsides to escape predators.  Obviously the opposite applies to the sinistrous sub-species, which don't.

The code provided here takes an PNG file as input, and converts each pixel to a value between 0 (black) and 255 (white).  Threshold values
for contours can have any value from 0 to 255.  Output is in the form of a simple SVG file.

## Status

The algorithm came to me while working on a version of [Ben Foxall's Moore-Neighbourhood contour finder](https://github.com/benfoxall/contours).

In fact an important part of the method -- using an array of 'seen' flags to avoid reworking contours that have
been found already -- come's straight from Foxall's code.

## Usage

    $ hcontours thingy.png

will create a file called thingy-hc-t128m15A4L.svg.  The numbers in the output SVG file name indicate
the values used for the threshold, margin, and paper options -- in this case, the default values.

### Options

* `--threshold | -t <value[,...]>`
Specify one or more threshold values, separated by commas, each in the range 0..255.  These are the pixel
values that are used to find the contours.  Default `128`. Examples: `-t 99` `-threshold 32,64,96,128,160,192,224`

* `--margin | -m <width>`
Define the minimum width, in mm, of the margin around the created image.  Default 15.  Example: `-m 10`

* `--paper | -p <papersize>`
Choose the paper size to use.  One of A4L, A4P, A3L, or A3P.  Default A4L. Example: `-p A3L`

* `--frame | -f`
Use this option to draw a simple frame around the image.  Default false. Example: `-f`

## Examples

`hcontours beach.png -t 32,64,96,128,160,192,224` produces this:

<img alt="Photo of breakwaters on a beach" src="examples/beach.png" title="Input image" width=45%> <img alt="The same photo after processing, showing as the outlines of shapes" src="examples/beach-hc-t32,64,96,128,160,192,224m15A4L.svg" title="Created SVG image" width=45%>


`hcontours Heightmap.png -t 32,64,96,128,160,192,224` produces this:

<img alt="Sample heightmap, taken from Wikipedia, shown as a greyscale image" src="examples/Heightmap.png" title="Input image" width=45%> <img alt="The contours generated from the heightmap" src="examples/Heightmap-hc-t32,64,96,128,160,192,224m15A4L.svg" title="Created SVG image" width=45%>

