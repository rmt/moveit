package main

import (
	"fmt"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xwindow"
	"log"
	"os"
)

func CurrentDesktop(xu *xgbutil.XUtil) uint {
	desktop, err := ewmh.CurrentDesktopGet(xu)
	if err != nil {
		log.Fatal(err)
	}
	return desktop
}

func WindowsOnCurrentDesktop(xu *xgbutil.XUtil) chan xproto.Window {
	c := make(chan xproto.Window)
	go func() {
		desktop := CurrentDesktop(xu)
		windows, err := ewmh.ClientListGet(xu)
		if err != nil {
			close(c)
			log.Fatal(err)
		}
		for _, win := range windows {
			windesktop, err := ewmh.WmDesktopGet(xu, win)
			if err != nil && windesktop == desktop {
				c <- win
			}
		}
		close(c)
	}()
	return c
}

//func FindBiggestWindowAt(xu *xgbutil.XUtil, x int, y int) {
//}

func print_windows_on_current_desktop(xu *xgbutil.XUtil) {
	desktop := CurrentDesktop(xu)
	windows, err := ewmh.ClientListGet(xu)
	if err != nil {
		log.Fatal(err)
	}
	for _, win := range windows {
		windesktop, err := ewmh.WmDesktopGet(xu, win)
		if err != nil {
			log.Fatal(err)
		}
		if windesktop == desktop {
			name, err := ewmh.WmNameGet(xu, win)
			if err != nil || len(name) == 0 {
				name = "N/A"
			}
			dgeom, err := xwindow.New(xu, win).DecorGeometry()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%s (0x%x)\n", name, win)
			fmt.Printf("\tGeometry: %s\n", dgeom)
		}
	}
}

func smart_placement(xu *xgbutil.XUtil, w xproto.Window, position string, horizpc int) {
	// viewports define the starting position of desktops
	viewports, err := ewmh.DesktopViewportGet(xu)
	if err != nil {
		log.Fatal(err)
	}
	// workareas define the usable area of a desktop
	workareas, err := ewmh.WorkareaGet(xu)
	if err != nil {
		log.Fatal(err)
	}
	window_desktop, err := ewmh.WmDesktopGet(xu, w)
	if err != nil {
		log.Fatal(err)
	}
	workarea := workareas[window_desktop]
	viewport := viewports[window_desktop]
	swidth, sheight := int(workarea.Width), int(workarea.Height)
	extents, err := ewmh.FrameExtentsGet(xu, w)
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
	x += viewport.X + workarea.X
	y += viewport.Y + workarea.Y
	// adjust for window extents
	width -= (extents.Left + extents.Right)
	height -= (extents.Top + extents.Bottom)
	fmt.Printf("MoveResizeWindow(id=0x%x, x=%d, y=%d, width=%d, height=%d)\n",
		w, x, y, width, height)
	ewmh.MoveresizeWindow(xu, w, x, y, width, height)
	ewmh.RestackWindow(xu, w)
}

func main() {
	fmt.Println("Move & resize a window")
	conn, err := xgbutil.NewConn()
	if err != nil {
		log.Fatal(err)
	}
	w, err := ewmh.ActiveWindowGet(conn)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Active Window: %d\n", w)
	//screen := w.GetScreen()
	//fmt.Printf("  On Screen #: %d\n", screen.number)
	//geom := w.GetGeometry()
	//fmt.Printf("  X: %d, Y: %d, Width: %d, Height: %d\n",
	//	geom.X, geom.Y, geom.Width, geom.Height)
	if len(os.Args) > 1 {
		command := os.Args[1]
		if command == "list" {
			print_windows_on_current_desktop(conn)
		} else {
			smart_placement(conn, w, command, 60)
		}
	}
}
