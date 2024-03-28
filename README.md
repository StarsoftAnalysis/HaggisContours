# HaggisContours

A method of converting a heightmap (or similar input) to a set of contours, using the wild haggis algorithm.

![Go workflow](https://github.com/starsoftanalysis/HaggisContours/actions/workflows/go.yml/badge.svg)

This is a Go programme for extacting contours from a [heightmap](https://en.wikipedia.org/wiki/Heightmap), or from a similar data source such as an image.

It uses a simple algorithm, based on the behaviour of wild haggis on the mountains of Scotland.  As is [well known](https://www.thehaggis.com/wild-haggis-all-about-haggis/),
dextrous wild haggis have their left legs longer than their right legs, which makes it easy for them to run clockwise
around hillsides to escape predators.  Obviously the opposite applies to the sinistrous sub-species, which don't.

The code provided here takes an PNG file as input, and converts each pixel to a value between 0 (black) and 255 (white).  Threshold values
for contours can have any value from 0 to 255.  Output is in the form of a simple SVG file.

## Status

The algorithm came to me while working on a version of [Ben Foxall's Moore-Neighbourhood contour finder](https://github.com/benfoxall/contours).

In fact an important part of the method -- using an array of 'seen' flags to avoid reworking contours that have
been found already -- comes straight from Foxall's code.

## Usage

    $ hcontours thingy.png

will create a file called thingy-hc-t128m15pA4L.svg.  The numbers in the output SVG file name indicate
the values used for the threshold, margin, and paper options -- in this case, the default values.

### Options

* `--threshold | -t <value[,...]>`
Specify one or more threshold values, separated by commas, each in the range 0..255.  These are the pixel
values that are used to find the contours.  Default `128`. Examples: `-t 99` `--threshold 32,64,96,128,160,192,224`

* `--margin | -m <width>`
Define the minimum width of the margin around the created image.  
The value is interpreted as millimetres if greater than 2, otherwise as inches.
Default 15 (mm).  Examples: `-m 10` (mm) `--margin 1.5` (inches)

* `--paper | -p <papersize>`
Choose the paper size to use, either one of the pre-defined sizes (A4L, A4P, A3L, or A3P),
or a custom size in the format `<width>x<height>`.  Width and height are interpreted as millimetres
if the value is greater than 30, otherwise as inches.  
Default A4L. Examples: `-p A3L` `--paper 200x300` (mm) `-p 7x5` (inches)

* `--frame | -f`
Draw a simple frame around the SVG image.  Default false. Example `-f`

* `--image | -i`
Use the original image as a background in the SVG image.  Default false. Example: `--image`

## Examples

`hcontours examples/beach.png -t 32,64,96,128,160,192,224 --paper A4L --image` produces this:

<img alt="Photo of breakwaters on a beach" src="examples/beach.png" title="Input image" width=45%>&nbsp;&nbsp;&nbsp;&nbsp;<img alt="The same photo after processing, showing as the outlines of shapes" src="examples/beach-hc-t32,64,96,128,160,192,224m15pA4LI.svg" title="Created SVG image" width=45%>


`hcontours examples/Heightmap.png -t 64,128,192 --paper 200x200 --margin 0 --frame` produces this:

<img alt="Sample heightmap, taken from Wikipedia, shown as a greyscale image" src="examples/Heightmap.png" title="Input image" width=45%>&nbsp;&nbsp;&nbsp;&nbsp;<img alt="The contours generated from the heightmap" src="examples/Heightmap-hc-t64,128,192m0p200x200F.svg" title="Created SVG image" width=45%>

