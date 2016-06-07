package desktop

import (
	"fmt"

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
	//x += headMinusStruts.X()
	//y += headMinusStruts.Y()
	height = height - extents.Top - extents.Bottom
	width = width - extents.Left - extents.Right
	fmt.Printf("Calling desk.MoveResizeWindow(win, x=%d, y=%d, width=%d, height=%d)\n",
		x, y, width, height)
	desk.MoveResizeWindow(w, x, y, width, height)
	ewmh.RestackWindow(desk.X, w)
}
