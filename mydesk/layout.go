package mydesk

import (
	"fmt"
	"log"

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
	desk *MyDesk
}

func (desk *MyDesk) SmartPlacement(w xproto.Window, position string, horizpc int) {
	// viewports define the starting position of desktops
	viewports, err := ewmh.DesktopViewportGet(desk.X)
	if err != nil {
		log.Fatal(err)
	}
	// workareas define the usable area of a desktop
	workareas, err := ewmh.WorkareaGet(desk.X)
	if err != nil {
		log.Fatal(err)
	}
	window_desktop, err := ewmh.WmDesktopGet(desk.X, w)
	if err != nil {
		log.Fatal(err)
	}
	workarea := workareas[window_desktop]
	viewport := viewports[window_desktop]
	swidth, sheight := int(workarea.Width), int(workarea.Height)
	extents, err := ewmh.FrameExtentsGet(desk.X, w)
	if err != nil {
		log.Fatal(err)
	}
	x := 0
	y := 0
	width := swidth
	height := sheight
	if position == "W" || position == "NW" || position == "SW" {
		// left side of screen
		width = swidth * horizpc / 100
	} else if position == "E" || position == "NE" || position == "SE" {
		// right side of screen
		width = swidth * (100 - horizpc) / 100
		x = swidth - width
	}
	if position == "N" || position == "NE" || position == "NW" {
		height = sheight / 2
	} else if position == "S" || position == "SE" || position == "SW" {
		height = sheight / 2
		y = sheight - height
	}
	// adjust for viewport offset & work area
	fmt.Printf("viewport.X==%d, viewport.Y==%d\n", viewport.X, viewport.Y)
	fmt.Printf("workarea.X==%d, workarea.Y==%d\n", workarea.X, workarea.Y)
	x += viewport.X + workarea.X
	y += viewport.Y + workarea.Y
	// adjust for window extents
	width -= (extents.Left + extents.Right)
	height -= (extents.Top + extents.Bottom)
	fmt.Printf("MoveResizeWindow(id=0x%x, x=%d, y=%d, width=%d, height=%d)\n",
		w, x, y, width, height)
	ewmh.MoveresizeWindow(desk.X, w, x, y, width, height)
	ewmh.RestackWindow(desk.X, w)
}



