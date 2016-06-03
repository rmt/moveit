package mydesk

import (
	"fmt"
	"log"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xinerama"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/BurntSushi/xgbutil/xrect"
)

type MyDesk struct {
	X					*xgbutil.XUtil
	Heads				xinerama.Heads
	HeadsMinusStruts	xinerama.Heads
}

// determine the head configuration for X
func getHeads(xu *xgbutil.XUtil, rootgeom xrect.Rect) xinerama.Heads {
	var heads xinerama.Heads
	if xu.ExtInitialized("XINERAMA") {
		var err error
		heads, err = xinerama.PhysicalHeads(xu)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		heads = xinerama.Heads{rootgeom}
	}
	return heads
}

func NewMyDesk() *MyDesk {
	conn, err := xgbutil.NewConn()
	if err != nil {
		log.Fatal(err)
	}

	// determine geometry of root window (ie. the desktop)
	root := xwindow.New(conn, conn.RootWin())
	rootgeom, err := root.Geometry()
	if err != nil {
		log.Fatal(err)
	}

	// get head geometry, & then with struts
	heads := getHeads(conn, rootgeom)
	headsMinusStruts := getHeads(conn, rootgeom) // TODO: copy instead

	/*
	 *  apply struts of top level windows against headsMinusStruts,
	 *  modifying it in-place.
	 */
	clients, err := ewmh.ClientListGet(conn)
	if err != nil {
		log.Fatal(err)
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
			strut.BottomStartX, strut.BottomEndX)
	}
	return &MyDesk{
		X:					conn,
		Heads:				heads,
		HeadsMinusStruts:	headsMinusStruts}
}

func (m *MyDesk) CurrentDesktop() uint {
	desktop, err := ewmh.CurrentDesktopGet(m.X)
	if err != nil {
		log.Fatal(err)
	}
	return desktop
}

func (m *MyDesk) HeadForWindow(win xproto.Window) int {
	dgeom, err := xwindow.New(m.X, win).DecorGeometry()
	if err != nil {
		log.Fatal(err)
	}
	for i, head := range(m.Heads) {
		if dgeom.X() >= head.X() && dgeom.X() < (head.X()+head.Width()) && dgeom.Y() >= head.Y() && dgeom.Y() < (head.Y()+head.Height()) {
			return i
		}
	}
	return 0 // if it's off the screen somewhere, return 0
}

func (m *MyDesk) ActiveWindow() xproto.Window {
	w, err := ewmh.ActiveWindowGet(m.X)
	if err != nil {
		log.Fatal(err)
	}
	return w
}

// move a window relative to its monitor
func (m *MyDesk) MoveResizeWindow(win xproto.Window, x, y, width, height int) {
	fmt.Printf("MoveResizeWindow(0x%x, %d, %d, %d, %d)\n", win, x, y, width, height)
}

// return channels
func (m *MyDesk) WindowsOnCurrentDesktop() chan xproto.Window {
	c := make(chan xproto.Window)
	go func() {
		desktop := m.CurrentDesktop()
		windows, err := ewmh.ClientListGet(m.X)
		if err != nil {
			close(c)
			log.Fatal(err)
		}
		for _, win := range windows {
			windesktop, err := ewmh.WmDesktopGet(m.X, win)
			if err == nil && windesktop == desktop {
				c <- win
			}
		}
		close(c)
	}()
	return c
}

func (m *MyDesk) PrintHeadGeometry() {
	for i, head := range m.Heads {
		fmt.Printf("\tHead              #%d : %s\n", i, head)
	}
	for i, head := range m.HeadsMinusStruts {
		fmt.Printf("\tHead minus struts #%d: %s\n", i, head)
	}
}

func (m *MyDesk) PrintWindowSummary(win xproto.Window) {
	name, err := ewmh.WmNameGet(m.X, win)
	if err != nil || len(name) == 0 {
		name = "N/A"
	}
	dgeom, err := xwindow.New(m.X, win).DecorGeometry()
	if err != nil {
		log.Fatal(err)
	}
	head := m.HeadForWindow(win)
	fmt.Printf("Window 0x%x: %s\n", win, name)
	fmt.Printf("\tGeometry: %s (head #%d)\n", dgeom, head)
}

func (m *MyDesk) PrintWindowsOnCurrentDesktop() {
	fmt.Printf("Desktop #%d\n", m.CurrentDesktop())
	for win := range m.WindowsOnCurrentDesktop() {
		m.PrintWindowSummary(win)
	}
}


