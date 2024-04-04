# MoveIt

MoveIt is a CLI manual desktop layout manager for EWMH compatible Window
managers such as OpenBox.  MoveIt is not a daemon.

It lets you move windows to the sides, corners, or center of the screen.

It lets you focus (and cycle through) windows at the sides, corners, or center
of the screen.

It adds features that I wish were in OpenBox, but aren't (or aren't quite as I
want them).


## Rationale

OpenBox just works...with everything.  But every tiling window manager that
I've tried has _some_ kind of quirks, and some apps misbehave.

I also wanted to learn how X works, as well as using it as a basis for learning
Go.


## Why not modify OpenBox?

This would be a reasonable approach for some of the features, but others are
really outside the scope of OpenBox.


## Does this work with Wayland?

No, it's X only.  Wayland hasn't yet standardized an API for this kind of
functionality, so features like this will have to be implemented in each
compositor for now.  LabWC is halfway there.


## Completed Features

* Xinerama/multi-head & multi-desktop aware
* Send windows to a specific position on the screen
* Focus on windows at a specific position
  - cycle through windows at some position
* Support Multiple layouts
  - 4 tiled windows, one (optional) overlapping centered window
    - each window could be in one of N, NW, W, SW, S, SE, E, NE, or C for center
  - 4 overlapping windows (each corner is visible)
* Switch to next monitor (focus based on most recently used, or if not possible, the largest unobscured window)


## Upcoming Features

* Auto-layout .. move & resize windows so they fit the layout best.
* Swap windows at positions - eg. swap a large centered window with one of the 4 background windows
* Cycle through windows on a specific head (forward and reverse, ~clockwise then center)
* Cycle through windows at the same position as the active window
* Swap all windows on Head X of desktop N with current windows on head X?  Better would be to logical head#s, which stay consistent, but default to desktop#
* Swap all windows on Head N for those on Head M
* Track logical position of each-windows, tie in to auto-layout


## Bugfixes
* If no window is focused, determine current head based on the mouse position
  - no longer crashes
  - TODO: check that it works on multi-head
* Unmaximize window before moving it


## OpenBox Configuration

Here is my OpenBox configuration, using the numeric keypad and the Shift & Windows keys to focus & move windows.

```xml
    <keybind key="S-W-KP_7">
      <action name="UnmaximizeFull"/>
      <action name="Execute">
        <command>moveit move NW</command>
      </action>
    </keybind>
    <keybind key="S-W-KP_8">
      <action name="UnmaximizeFull"/>
      <action name="Execute">
        <command>moveit move N</command>
      </action>
    </keybind>
    <keybind key="S-W-KP_9">
      <action name="UnmaximizeFull"/>
      <action name="Execute">
        <command>moveit move NE</command>
      </action>
    </keybind>
    <keybind key="S-W-KP_1">
      <action name="UnmaximizeFull"/>
      <action name="Execute">
        <command>moveit move SW</command>
      </action>
    </keybind>
    <keybind key="S-W-KP_2">
      <action name="UnmaximizeFull"/>
      <action name="Execute">
        <command>moveit move S</command>
      </action>
    </keybind>
    <keybind key="S-W-KP_3">
      <action name="UnmaximizeFull"/>
      <action name="Execute">
        <command>moveit move SE</command>
      </action>
    </keybind>
    <keybind key="S-W-KP_4">
      <action name="UnmaximizeFull"/>
      <action name="Execute">
        <command>moveit move W</command>
      </action>
    </keybind>
    <keybind key="S-W-KP_5">
      <action name="UnmaximizeFull"/>
      <action name="Execute">
        <command>moveit move C</command>
      </action>
      <action name="Raise"/>
      <action name="Focus"/>
    </keybind>
    <keybind key="C-W-KP_5">
      <action name="UnmaximizeFull"/>
      <action name="Execute">
        <command>moveit move C</command>
      </action>
      <action name="Raise"/>
      <action name="Focus"/>
    </keybind>
    <keybind key="C-W-KP_7">
      <action name="Execute">
        <command>moveit move BNW</command>
      </action>
    </keybind>
    <keybind key="C-W-KP_9">
      <action name="Execute">
        <command>moveit move BNE</command>
      </action>
    </keybind>
    <keybind key="C-W-KP_1">
      <action name="Execute">
        <command>moveit move BSW</command>
      </action>
    </keybind>
    <keybind key="C-W-KP_3">
      <action name="Execute">
        <command>moveit move BSE</command>
      </action>
    </keybind>
    <keybind key="S-W-KP_6">
      <action name="Execute">
        <command>moveit move E</command>
      </action>
    </keybind>
    <keybind key="W-KP_1">
      <action name="Execute">
        <command>moveit focus SW</command>
      </action>
    </keybind>
    <keybind key="W-KP_2">
      <action name="Execute">
        <command>moveit focus S</command>
      </action>
    </keybind>
    <keybind key="W-KP_3">
      <action name="Execute">
        <command>moveit focus SE</command>
      </action>
    </keybind>
    <keybind key="W-KP_4">
      <action name="Execute">
        <command>moveit focus W</command>
      </action>
    </keybind>
    <keybind key="W-KP_5">
      <action name="Execute">
        <command>moveit focus C</command>
      </action>
    </keybind>
    <keybind key="W-KP_6">
      <action name="Execute">
        <command>moveit focus E</command>
      </action>
    </keybind>
    <keybind key="W-KP_7">
      <action name="Execute">
        <command>moveit focus NW</command>
      </action>
    </keybind>
    <keybind key="W-KP_8">
      <action name="Execute">
        <command>moveit focus N</command>
      </action>
    </keybind>
    <keybind key="W-KP_9">
      <action name="Execute">
        <command>moveit focus NE</command>
      </action>
    </keybind>
```
