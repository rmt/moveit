package desktop

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xrect"
)

func (desk *Desktop) SmartPlacement(w xproto.Window, position string, horizpc int) {
	if w == 0 {
		fmt.Println("No focused window. Nothing to do.")
		return
	}
	headnum := desk.GetHeadForWindow(w)
	headMinusStruts := desk.HeadsMinusStruts[headnum]
	extents := desk.GetFrameExtentsForWindow(w)

	headWidth, headHeight := int(headMinusStruts.Width()), int(headMinusStruts.Height())
	x := 0
	y := 0
	width := headWidth
	height := headHeight

	// horizontal
	if position == "W" || position == "NW" || position == "SW" {
		// left side of screen
		width = headWidth * horizpc / 100
	} else if position == "E" || position == "NE" || position == "SE" {
		// right side of screen
		width = headWidth * (100 - horizpc) / 100
		x = headWidth - width
	}

	// vertical
	if position == "N" || position == "NE" || position == "NW" {
		height = headHeight / 2
	} else if position == "S" || position == "SE" || position == "SW" {
		height = headHeight / 2
		y = headHeight - height
	}

	// center
	if position == "C" {
		width = headWidth / 10 * 7
		height = headHeight / 10 * 7
		x = (headWidth - width) / 2
		y = (headHeight - height) / 2
	}

	// big windows
	if strings.HasPrefix(position, "B") {
		width = headWidth / 10 * 7
		height = headHeight / 10 * 7
		if strings.HasPrefix(position, "BS") {
			y = headHeight - height
		}
		if strings.HasSuffix(position, "E") {
			x = headWidth - width
		}
	}
	height = height - extents.Top - extents.Bottom
	width = width - extents.Left - extents.Right
	fmt.Printf("Calling desk.MoveResizeWindow(win, x=%d, y=%d, width=%d, height=%d)\n",
		x, y, width, height)
	desk.MoveResizeWindow(w, x, y, width, height)
	ewmh.RestackWindow(desk.X, w)
}

func (desk *Desktop) SmartFocus(activeWindow xproto.Window, position string, horizpc int) {
	var headnum int
	if activeWindow != 0 {
		headnum = desk.GetHeadForWindow(activeWindow)
	} else {
		headnum = desk.GetHeadForPointer()
	}
	head := desk.HeadsMinusStruts[headnum]
	headWidth, headHeight := int(head.Width()), int(head.Height())
	centerX := head.X() + headWidth*horizpc/100
	centerY := head.Y() + headHeight/2

	var minX, minY, maxX, maxY int
	if strings.HasPrefix(position, "BS") || strings.HasPrefix(position, "S") {
		minY = centerY
		maxY = head.Y() + head.Height()
	} else if strings.HasPrefix(position, "BN") || strings.HasPrefix(position, "N") {
		minY = head.Y()
		maxY = centerY
	} else {
		minY = head.Y()
		maxY = minY + head.Height()
	}

	if strings.HasSuffix(position, "W") {
		minX = head.X()
		maxX = centerX
	} else if strings.HasSuffix(position, "E") {
		minX = centerX
		maxX = head.X() + head.Width()
	} else {
		minX = head.X()
		maxX = minX + head.Width()
	}

	var newWin xproto.Window = 0

	if position == "O" { // other monitor
		if len(desk.Heads) <= 1 {
			return
		}
		newHead := desk.HeadsMinusStruts[(headnum+1)%len(desk.HeadsMinusStruts)]
		headRect := xrect.New(newHead.X(), newHead.Y(), newHead.X()+newHead.Width(), newHead.Y()+newHead.Height())
		fmt.Println(headRect)
		newWin = desk.NextMatchingWindow(activeWindow, func(r xrect.Rect) bool {
			x := RectMostlyInRect(r, headRect)
			if x {
				fmt.Printf("Matched rect: %v\n", r)
			}
			return x
		})
	} else if position == "C" {
		// only focus centered windows
		minX = head.X() + (head.Width() / 20 * 5)
		maxX = head.X() + (head.Width() / 20 * 15)
		minY = head.Y() + (head.Height() / 20 * 5)
		maxY = head.Y() + (head.Height() / 20 * 15)
		centerRect := xrect.New(minX, minY, maxX-minX, maxY-minY)
		newWin = desk.NextMatchingWindow(activeWindow, func(r xrect.Rect) bool {
			return RectMostlyInRect(r, centerRect)
		})
	} else {
		cmpRect := xrect.New(minX, minY, maxX-minX, maxY-minY)
		newWin = desk.NextMatchingWindow(activeWindow, func(r xrect.Rect) bool {
			// RectMostlyInRect will match smaller windows
			// EachAxisMostlyOverlaps will match bigger windows
			return RectMostlyInRect(r, cmpRect) || EachAxisMostlyOverlaps(r, cmpRect)
		})
	}
	if newWin != 0 {
		ewmh.RestackWindow(desk.X, newWin)
		ewmh.ActiveWindowReq(desk.X, newWin)
	}
}
