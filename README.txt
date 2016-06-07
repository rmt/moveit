Move It
-------

MoveIt is a command line driven manual desktop layout manager for EWMH
compatible Window managers such as OpenBox.

Rationale
---------

OpenBox just works...with everything.  But every tiling window manager that
I've tried has _some_ kind of quirks, and some apps misbehave.

Features
--------

* Xinerama aware
* Virtual Desktop aware
* Send windows to a specific position on the screen (according to layout)
* Focus on windows at a specific position
  - cycle through windows at some position
* Swap windows at positions
  - eg. swap a large centered window with one of the 4 background windows
* Multiple layouts
  - 4 tiled windows, one (optional) overlapping centered window
    - each window could be in one of N, NW, W, SW, S, SE, E, NE, or C for center
  - 4 overlapping windows (each corner is visible)
* Auto-layout .. move & resize windows so they fit the layout best.
