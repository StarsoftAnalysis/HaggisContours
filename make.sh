#!/usr/bin/env bash

# Update examples etc.

./hcontours examples/Heightmap.png -t 64,128,192 --paper 200x200 --margin 0 --frame
./hcontours examples/beach.png -t 32,64,96,128,160,192,224 --paper A4L --image --linewidth 0.3

inkscape "examples/beach-hc-t32,64,96,128,160,192,224m15pA4LI.svg" -o "examples/beach-hc-t32,64,96,128,160,192,224m15pA4LI.png"
