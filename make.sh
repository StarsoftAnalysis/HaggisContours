#!/usr/bin/env bash

# This file is part of hcontours -- HarrisContours.
# Copyright (C) 2024 Chris Dennis, chris@starsoftanalysis.co.uk
#
# hcontours is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.

# Update examples etc.
go build
./hcontours examples/Heightmap.png -t 64,128,192 --paper 200x200 --margin 0 --framewidth 1.0
./hcontours examples/beach.png -t 32,64,96,128,160,192,224 --paper A4L --image --linewidth 0.3
./hcontours examples/beach.png --colours ff0000,777700,00ff00,00ffff,0000ff,770077 -T5

inkscape "examples/beach-hc-t32,64,96,128,160,192,224m15pA4LI.svg" -o "examples/beach-hc-t32,64,96,128,160,192,224m15pA4LI.png"
inkscape "examples/beach-hc-T5m15pA4LCff0000,777700,00ff00,00ffff,0000ff,770077.svg" -o "examples/beach-hc-T5m15pA4LCff0000,777700,00ff00,00ffff,0000ff,770077.png"
