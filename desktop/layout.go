package desktop

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/xrect"
)

type LayoutClassifier interface {
	Classify(*xrect.Rect) string
}

type LayoutPositioner interface {
	Position(pos string) *xrect.Rect
	NextPosition(pos string) string
	PrevPosition(pos string) string
}

type Layout interface {
	LayoutClassifier
	LayoutPositioner
}

type DefaultLayout struct {
	desk *Desktop
}

func (desk *Desktop) SmartPlacement(w xproto.Window, position string, horizpc int) {
	headnum := desk.HeadForWindow(w)
	headMinusStruts := desk.HeadsMinusStruts[headnum]
	extents := desk.ExtentsForWindow(w)

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
	//x += headMinusStruts.X()
	//y += headMinusStruts.Y()
	height = height - extents.Top - extents.Bottom
	width = width - extents.Left - extents.Right
	fmt.Printf("Calling desk.MoveResizeWindow(win, x=%d, y=%d, width=%d, height=%d)\n",
		x, y, width, height)
	desk.MoveResizeWindow(w, x, y, width, height)
	ewmh.RestackWindow(desk.X, w)
}

// SmartFocus will rotate focus for windows with positions between
// minX,minY and maxX,maxY
func (desk *Desktop) SmartFocusAt(activeWindow xproto.Window, headnum, minX, minY, maxX, maxY int) {
	head := desk.Heads[headnum]
	fmt.Printf("SmartFocusAt(activeWindow=0x%x headnum=%d, minX=%d, minY=%d, maxX=%d, maxY=%d)\n",
		activeWindow, headnum, minX, minY, maxX, maxY)
	minX += head.X()
	minY += head.Y()
	maxX += head.X()
	maxY += head.Y()
	var matchingWindows []xproto.Window
	for win := range(desk.WindowsOnCurrentDesktop()) {
		desk.PrintWindowSummary(win)
		geom := desk.GeometryForWindow(win)
		if minX >= geom.X() && minY >= geom.Y() && maxX <= (geom.X()+geom.Width()) && maxY <= (geom.Y()+geom.Height()) {
			matchingWindows = append(matchingWindows, win)
			fmt.Printf("Matched window 0x%x\n", win)
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
	nextWinIndex := len(matchingWindows)-1
	if matchingWindows[nextWinIndex] == activeWindow {
		nextWinIndex = 0 // rotate
	}
	if nextWinIndex >= 0 {
		ewmh.RestackWindow(desk.X, matchingWindows[nextWinIndex])
		ewmh.ActiveWindowReq(desk.X, matchingWindows[nextWinIndex])
	}
}

func (desk *Desktop) SmartFocus(activeWindow xproto.Window, position string) {
	headnum := desk.HeadForWindow(activeWindow)
	headMinusStruts := desk.HeadsMinusStruts[headnum]

	minX, minY, maxX, maxY := 0, 0, 0, 0
	if strings.HasPrefix(position, "BS") || strings.HasPrefix(position, "S") {
		minY = headMinusStruts.Height() - 10
		maxY = minY
	} else {
		minY = headMinusStruts.Y() + 10
		maxY = minY
	}
	if strings.HasSuffix(position, "W") {
		minX = headMinusStruts.X() + 10
		maxX = minX
	} else {
		minX = headMinusStruts.Width() - 10
		maxX = minX
	}
	if position == "C" {
		minX = headMinusStruts.X() + 110
		maxX = headMinusStruts.Width() - 110
		minY = headMinusStruts.Y() + 70
		maxY = headMinusStruts.Height() - 70
	}
	desk.SmartFocusAt(activeWindow, headnum, minX, minY, maxX, maxY)
}
