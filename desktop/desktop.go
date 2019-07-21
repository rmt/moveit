/*
Package desktop implements routines to discover and manipulate windows
on the current X desktop.

The primary struct is Desktop, which is returned by NewDesktop()

There are two special functions called SmartPlacement and SmartFocus that
will move/resize & focus windows according to a string placement, where
these are:
  * NE, N, NW, W, SW, S, SE, E (North-East, North, ...)
  * C (Center)
  * BNE, BNW, BSW, BSE (big NE, ...) [SmartPlacement only]
*/
package desktop

import (
	"fmt"
	"log"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xinerama"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/BurntSushi/xgbutil/xrect"
)

const PinnedDesktopNumber = 4294967295

type Desktop struct {
	X					*xgbutil.XUtil
	Heads				xinerama.Heads
	HeadsMinusStruts	xinerama.Heads
}

func NewDesktop() *Desktop {
	conn, err := xgbutil.NewConn()
	if err != nil {
		log.Fatal(err)
	}

	// determine geometry of root window (ie. the desktop)
	root := xwindow.New(conn, conn.RootWin())
	rootgeom, err := root.Geometry()
	if err != nil {
		log.Fatalf("NewDesktop(): Getting root window geometry: %#v", err)
	}

	fmt.Printf("rootGeometry: X=%d-%d, Y=%d-%d\n", rootgeom.X(), rootgeom.Width(), rootgeom.Y(), rootgeom.Height())

	// get head geometry, & then with struts
	heads := getHeads(conn, rootgeom)
	headsMinusStruts := getHeads(conn, rootgeom) // TODO: copy instead

	// apply struts of top level windows against headsMinusStruts,
	// modifying it in-place.
	clients, err := ewmh.ClientListGet(conn)
	if err != nil {
		log.Fatalf("NewDesktop(): ewmh.ClientListGet: %#v", err)
	}
	for _, clientid := range clients {
		strut, err := ewmh.WmStrutPartialGet(conn, clientid)
		if err != nil { // no struts for this client
			continue
		}

		// Apply the struts to headsMinusStruts, modifying it
		xrect.ApplyStrut(headsMinusStruts,
			uint(rootgeom.Width()), uint(rootgeom.Height()),
			strut.Left, strut.Right, strut.Top, strut.Bottom,
			strut.LeftStartY, strut.LeftEndY,
			strut.RightStartY, strut.RightEndY,
			strut.TopStartX, strut.TopEndX,
			strut.BottomStartX, strut.BottomEndX,
		)
	}
	return &Desktop{
		X:					conn,
		Heads:				heads,
		HeadsMinusStruts:	headsMinusStruts,
	}
}

// return a list of Heads (monitors), falling back to the rootgeom
func getHeads(xu *xgbutil.XUtil, rootgeom xrect.Rect) xinerama.Heads {
	var heads xinerama.Heads
	if xu.ExtInitialized("XINERAMA") {
		var err error
		heads, err = xinerama.PhysicalHeads(xu)
		if err != nil {
			log.Fatalf("getHeads: xinerama.PhysicalHeads: %#v", err)
		}
	} else {
		heads = xinerama.Heads{rootgeom}
	}
	return heads
}


func (desk *Desktop) GetCurrentDesktop() uint {
	desktop, err := ewmh.CurrentDesktopGet(desk.X)
	if err != nil {
		log.Fatalf("GetCurrentDesktop(): %#v", err)
	}
	return desktop
}

func (desk *Desktop) IsWindowVisible(win xproto.Window) bool {
	state, err := icccm.WmStateGet(desk.X, win)
	if err != nil {
		log.Fatalf("IsWindowVisible(%d): %#v", win, err)
	}
	return state.State == icccm.StateNormal || state.State == icccm.StateZoomed
}

func (desk *Desktop) GetGeometryForWindow(win xproto.Window) xrect.Rect {
	dgeom, err := xwindow.New(desk.X, win).DecorGeometry()
	if err != nil {
		log.Fatalf("GetGeometryForWindow(%d): %#v", win, err)
	}
	return dgeom
}

func (desk *Desktop) GetFrameExtentsForWindow(win xproto.Window) *ewmh.FrameExtents {
	extents, err := ewmh.FrameExtentsGet(desk.X, win)
	if err != nil {
		log.Fatalf("GetFrameExtentsForWindow(%d): %#v", win, err)
	}
	return extents
}

func (desk *Desktop) GetHeadForPointer() int {
	rootwin := desk.X.RootWin()
	reply, err := xproto.QueryPointer(desk.X.Conn(), rootwin).Reply()
	if err != nil {
		log.Fatalf("GetHeadForPointer() xproto.QueryPointer: %#v", err)
	}
	for i, head := range(desk.Heads) {
		if int(reply.RootX) < head.X() { continue }
		if int(reply.RootX) >= head.X()+head.Width() { continue }
		if int(reply.RootY) < head.Y() { continue }
		if int(reply.RootY) >= head.Y()+head.Height() { continue }
		return i
	}
	log.Fatalf("HeadForPointer(): Couldn't determine head for pointer coordinates (%d,%d)\n",
		reply.RootX, reply.RootY)
	return 0
}

func (desk *Desktop) GetHeadForWindow(win xproto.Window) int {
	dgeom, err := xwindow.New(desk.X, win).DecorGeometry()
	if err != nil {
		log.Fatalf("GetHeadForWindow(%d): %#v", win, err)
	}
	for i, head := range(desk.Heads) {
		if dgeom.X() >= head.X() && dgeom.X() < (head.X()+head.Width()) && dgeom.Y() >= head.Y() && dgeom.Y() < (head.Y()+head.Height()) {
			return i
		}
	}
	return 0 // if it's off the screen somewhere, return 0
}

func (desk *Desktop) GetActiveWindow() xproto.Window {
	w, err := ewmh.ActiveWindowGet(desk.X)
	if err != nil {
		log.Fatalf("GetActiveWindow(): %#v", err)
	}
	return w
}

// move a window relative to the current head
func (desk *Desktop) MoveResizeWindow(win xproto.Window, x, y, width, height int) {
	fmt.Printf("MoveResizeWindow(0x%x, %d, %d, %d, %d)\n", win, x, y, width, height)
	headnr := desk.GetHeadForWindow(win)
	head := desk.Heads[headnr]
	x += head.X()
	y += head.Y()
	ewmh.MoveresizeWindow(desk.X, win, x, y, width, height)
	ewmh.RestackWindow(desk.X, win)
}

func (desk *Desktop) WindowsOnCurrentDesktop() chan xproto.Window {
	c := make(chan xproto.Window)
	go func() {
		desktop := desk.GetCurrentDesktop()
		windows, err := ewmh.ClientListStackingGet(desk.X)
		//windows, err := ewmh.ClientListGet(desk.X)
		if err != nil {
			close(c)
			log.Fatalf("WindowsOnCurrentDesktop(): %#v", err)
		}
		for _, win := range windows {
			windesktop, err := ewmh.WmDesktopGet(desk.X, win)
			if err == nil && (windesktop == desktop || windesktop == PinnedDesktopNumber) {
				c <- win
			}
		}
		close(c)
	}()
	return c
}

func (desk *Desktop) PrintHeadGeometry() {
	for i, head := range desk.Heads {
		fmt.Printf("\tHead              #%d : %s\n", i, head)
	}
	for i, head := range desk.HeadsMinusStruts {
		fmt.Printf("\tHead minus struts #%d: %s\n", i, head)
	}
}

func (desk *Desktop) PrintWindowSummary(win xproto.Window) {
	name, err := ewmh.WmNameGet(desk.X, win)
	if err != nil || len(name) == 0 {
		name = "N/A"
	}
	dgeom, err := xwindow.New(desk.X, win).DecorGeometry()
	if err != nil {
		log.Fatal(err)
	}
	head := desk.GetHeadForWindow(win)
	fmt.Printf("Window 0x%x: %s\n", win, name)
	fmt.Printf("\tGeometry: %s (head #%d)\n", dgeom, head)
}

func (desk *Desktop) PrintWindowsOnCurrentDesktop() {
	fmt.Printf("Desktop #%d\n", desk.GetCurrentDesktop())
	for win := range desk.WindowsOnCurrentDesktop() {
		desk.PrintWindowSummary(win)
	}
}
