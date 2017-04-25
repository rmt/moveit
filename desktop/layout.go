package desktop

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgb/xproto"
)

func (desk *Desktop) SmartPlacement(w xproto.Window, position string, horizpc int) {
	if w == 0 {
		fmt.Println("No focused window. Nothing to do.")
		return;
	}
	headnum := desk.GetHeadForWindow(w)
	headMinusStruts := desk.HeadsMinusStruts[headnum]
	extents := desk.GetFrameExtentsForWindow(w)

	headWidth, headHeight := int(headMinusStruts.Width()), int(headMinusStruts.Height())
	x := 0
	y := 0
	width := headWidth
	height := headHeight

	if position == "W" || position == "NW" || position == "SW" {
		// left side of screen
		width = headWidth * horizpc / 100
	} else if position == "E" || position == "NE" || position == "SE" {
		// right side of screen
		width = headWidth * (100 - horizpc) / 100
		x = headWidth - width
	}
	if position == "N" || position == "NE" || position == "NW" {
		height = headHeight / 2
	} else if position == "S" || position == "SE" || position == "SW" {
		height = headHeight / 2
		y = headHeight - height
	}
	if position == "C" {
		x = 100
		y = 60
		width = headWidth - x*2
		height = headHeight - y*2
	}
	if strings.HasPrefix(position, "B") {
		height = headHeight - 120
		width = headWidth - 200
		if strings.HasPrefix(position, "BS") {
			y = 120
		}
		if strings.HasSuffix(position, "E") {
			x = 200
		}
	}
	height = height - extents.Top - extents.Bottom
	width = width - extents.Left - extents.Right
	fmt.Printf("Calling desk.MoveResizeWindow(win, x=%d, y=%d, width=%d, height=%d)\n",
		x, y, width, height)
	desk.MoveResizeWindow(w, x, y, width, height)
	ewmh.RestackWindow(desk.X, w)
}

// SmartFocus will rotate focus for windows with positions between
// minX,minY and maxX,maxY
func (desk *Desktop) SmartFocusAt(activeWindow xproto.Window, minX, minY, maxX, maxY int) {
	fmt.Printf("SmartFocusAt(activeWindow=0x%x minX=%d, minY=%d, maxX=%d, maxY=%d)\n",
		activeWindow, minX, minY, maxX, maxY)

	var matchingWindows []xproto.Window = make([]xproto.Window, 0, 8)

	i := 0
	activeWinIndex := -1
	for win := range(desk.WindowsOnCurrentDesktop()) {
		geom := desk.GetGeometryForWindow(win)
		if minX >= geom.X() && minY >= geom.Y() && maxX <= (geom.X()+geom.Width()) && maxY <= (geom.Y()+geom.Height()) && desk.IsWindowVisible(win) {
			matchingWindows = append(matchingWindows, win)
			fmt.Printf("Matched window 0x%x\n", win)
			if(win == activeWindow) {
				activeWinIndex = i
			}
			i += 1
		}
	}
	if len(matchingWindows) == 0 {
		fmt.Println("No matching windows found.")
		return
	}

	// Matching windows should be in bottom-to-top stacking
	// order.  The goal is to focus the top window at a location
	// if it's not already focused, but iterate over all windows
	// at the same location if the location already has focus.
	var nextWinIndex int
	if activeWinIndex > -1 {
		nextWinIndex = (activeWinIndex+1) % len(matchingWindows)
	} else {
		nextWinIndex = len(matchingWindows)-1
	}
	fmt.Printf("Focusing window 0x%x (nextWinIndex==%d, activeWinIndex==%d)\n", matchingWindows[nextWinIndex], nextWinIndex, activeWinIndex)
	ewmh.RestackWindow(desk.X, matchingWindows[nextWinIndex])
	ewmh.ActiveWindowReq(desk.X, matchingWindows[nextWinIndex])
}

func (desk *Desktop) SmartFocus(activeWindow xproto.Window, position string) {
	var headnum int
	if activeWindow != 0 {
		headnum = desk.GetHeadForWindow(activeWindow)
	} else {
		headnum = desk.GetHeadForPointer()
	}
	headMinusStruts := desk.HeadsMinusStruts[headnum]

	var minX, minY, maxX, maxY int
	if strings.HasPrefix(position, "BS") || strings.HasPrefix(position, "S") {
		minY = headMinusStruts.Y() + headMinusStruts.Height()
		maxY = minY
	} else if strings.HasPrefix(position, "BN") || strings.HasPrefix(position, "N") {
		minY = headMinusStruts.Y()
		maxY = minY
	} else {
		minY = headMinusStruts.Y()
		maxY = minY + headMinusStruts.Height()
	}
	if strings.HasSuffix(position, "W") {
		minX = headMinusStruts.X()
		maxX = minX
	} else if strings.HasSuffix(position, "E") {
		minX = headMinusStruts.X() + headMinusStruts.Width()
		maxX = minX
	} else {
		minX = headMinusStruts.X()
		maxX = minX + headMinusStruts.Width()
	}
	if position == "C" {
		minX = headMinusStruts.X() + 100
		maxX = minX + headMinusStruts.Width() - 200
		minY = headMinusStruts.Y() + 60
		maxY = minY + headMinusStruts.Height() - 120
	}

	const ErrorMargin = 35
	desk.SmartFocusAt(activeWindow, minX+ErrorMargin, minY+ErrorMargin, maxX-ErrorMargin, maxY-ErrorMargin)
}
