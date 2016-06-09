package main

import (
	"fmt"
	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xinerama"
	"github.com/BurntSushi/xgb/xproto"
	"log"
	"os"
)

type Display struct {
	conn            *xgb.Conn
	setupInfo       *xproto.SetupInfo
	xineramaScreens []Screen
}

type Screen struct {
	number int
	info   xinerama.ScreenInfo
}

func NewDisplay(conn *xgb.Conn) *Display {
	setupInfo := xproto.Setup(conn)
	err := xinerama.Init(conn)
	reply, err := xinerama.QueryScreens(conn).Reply()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Number of heads: %d\n", reply.Number)
	screens := make([]Screen, reply.Number, reply.Number)
	for i, info := range reply.ScreenInfo {
		fmt.Printf("%d :: X: %d, Y: %d, Width: %d, Height: %d\n",
			i, info.XOrg, info.YOrg, info.Width, info.Height)
		screens[i].number = i
		screens[i].info = info
	}

	return &Display{
		conn:            conn,
		setupInfo:       setupInfo,
		xineramaScreens: screens}
}

type Window struct {
	display *Display
	id      xproto.Window
}

func (w *Window) String() string {
	return fmt.Sprintf("Window(0x%x)", w.id)
}

func (d *Display) SendEvent(Propagate bool, EventMask uint32, Event string) error {
	root := d.setupInfo.DefaultScreen(d.conn).Root
	err := xproto.SendEventChecked(
		d.conn,
		Propagate,
		root,
		EventMask,
		Event).Check()
	return err
}

func (w *Window) GetGeometry() *xproto.GetGeometryReply {
	reply, err := xproto.GetGeometry(w.display.conn, xproto.Drawable(w.id)).Reply()
	if err != nil {
		log.Fatal(err)
	}
	return reply
}

// return the current screen/virtual head for the window
func (w *Window) GetScreen() Screen {
	dim := w.GetGeometry()

	for _, screen := range w.display.xineramaScreens {
		info := screen.info
		if int32(dim.X) >= int32(info.XOrg) &&
			int32(dim.X) < (int32(info.XOrg)+int32(info.Width)) &&
			int32(dim.Y) >= int32(info.YOrg) &&
			int32(dim.Y) < (int32(info.YOrg)+int32(info.Height)) {
			return screen
		}
	}
	log.Println("Returning default screen")
	return w.display.xineramaScreens[0]
}

// move a window, relative to its current screen
func (w *Window) Move(X uint32, Y uint32) {
	log.Printf("Move(%d, %d)\n", X, Y)
	screen := w.GetScreen()
	X = X + uint32(screen.info.XOrg)
	Y = Y + uint32(screen.info.YOrg)
	err := xproto.ConfigureWindowChecked(
		w.display.conn,
		w.id,
		xproto.ConfigWindowX|xproto.ConfigWindowY,
		[]uint32{X, Y}).Check()
	if err != nil {
		log.Fatal(err)
	}
}

func (d *Display) GetAtom(name string) xproto.Atom {
	reply, err := xproto.InternAtom(
		d.conn,
		true,
		uint16(len(name)),
		name).Reply()
	if err != nil {
		log.Fatal(err)
	}
	return reply.Atom
}

// ask the window manager to move a window
func (w *Window) NiceMoveResize(X uint32, Y uint32, width uint32, height uint32) {
	d := w.display
	clientmessage := xproto.ClientMessageEvent{Sequence: 0, // sequence
		Window: w.id,
		Type:   d.GetAtom("_NET_WINDOW_MOVERESIZE"),
		Format: 32, // not sure - copied from wmctrl.c
		Data:   xproto.ClientMessageDataUnionData32New([]uint32{1, X, Y, width, height})}
	d.SendEvent(
		true,
		xproto.EventMaskSubstructureRedirect|xproto.EventMaskSubstructureNotify,
		string(clientmessage.Bytes()))
}

// move & resize a window, relative to its current screen
func (w *Window) MoveResize(X uint32, Y uint32, width uint32, height uint32) {
	log.Printf("MoveResize(%d, %d, %d, %d)\n", X, Y, width, height)
	screen := w.GetScreen()
	X = X + uint32(screen.info.XOrg)
	Y = Y + uint32(screen.info.YOrg)
	err := xproto.ConfigureWindowChecked(
		w.display.conn,
		w.id,
		xproto.ConfigWindowX|xproto.ConfigWindowY|xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
		[]uint32{X, Y, width, height}).Check()
	if err != nil {
		log.Fatal(err)
	}
}

//

func (d *Display) GetActiveWindow() Window {
	// Get the atom id (i.e., intern an atom) of "_NET_ACTIVE_WINDOW".
	root := d.setupInfo.DefaultScreen(d.conn).Root
	aname := "_NET_ACTIVE_WINDOW"
	activeAtom, err := xproto.InternAtom(
		d.conn,
		true,
		uint16(len(aname)),
		aname).Reply()
	if err != nil {
		log.Fatal(err)
	}
	reply, err := xproto.GetProperty(
		d.conn,
		false,
		root,
		activeAtom.Atom,
		xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
	if err != nil {
		log.Fatal(err)
	}
	windowId := xproto.Window(xgb.Get32(reply.Value))
	w := Window{d, windowId}
	return w
}

//func (w *Window) GetFrameExtents() {
//	aname := "_NET_FRAME_EXTENTS"
//	activeAtom, err := xproto.InternAtom(
//		w.display.conn,
//		true,
//		uint16(len(aname)),
//		aname).Reply()
//	if err != nil {
//		log.Fatal(err)
//	}
//	reply, err := xproto.GetProperty(
//		w.display.conn,
//		false,
//		w.id,
//		xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
//}

// return the top-most window of the currently focused window
func (d *Display) OldGetActiveWindow() Window {
	focused_reply, err := xproto.GetInputFocus(d.conn).Reply()
	if err != nil {
		log.Fatal(err)
	}
	var current_window = focused_reply.Focus
	for {
		R, err := xproto.QueryTree(d.conn, current_window).Reply()
		if err != nil {
			log.Fatal(err)
		}
		if R.Parent == R.Root || current_window == R.Root {
			log.Printf("Parent is %x, Root is %x, Active is %x\n", R.Parent, R.Root, current_window)
			w := Window{d, current_window}
			return w
		} else {
			current_window = R.Parent
		}
	}
}

func smart_placement(s *Screen, w *Window, position string, horizpc uint32) {
	swidth := uint32(s.info.Width)
	sheight := uint32(s.info.Height) - 51 // FIXME: reserved for tint2
	x := uint32(0)
	y := uint32(0)
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
	w.NiceMoveResize(x, y, width, height)
}

func main() {
	conn, err := xgb.NewConn()
	if err != nil {
		fmt.Println(err)
		return
	}
	display := NewDisplay(conn)
	w := display.GetActiveWindow()
	fmt.Printf("Active Window: %d\n", w)
	screen := w.GetScreen()
	fmt.Printf("  On Screen #: %d\n", screen.number)
	geom := w.GetGeometry()
	fmt.Printf("  X: %d, Y: %d, Width: %d, Height: %d\n",
		geom.X, geom.Y, geom.Width, geom.Height)
	if len(os.Args) > 2 {
		cmd := os.Args[1]
		placement := os.Args[2]
		if cmd == "move" {
			smart_placement(&screen, &w, placement, 60)
		} else {
		}
	} else {
		fmt.Printf("Syntax: %s move|focus B?[NS][EW]|C\n", os.Args[0])
		fmt.Println("eg. move NE, move BSE (big south-east)")
	}
}
