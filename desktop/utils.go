package desktop

import (
	"fmt"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/xrect"
)

func min(a, b int) int { if a < b { return a } else { return b } }
func max(a, b int) int { if a > b { return a } else { return b } }

func lineOverlap(min1, max1, min2, max2 int) int {
	return max(0, min(max1, max2) - max(min1, min2))
}

func AreaOverlap(a xrect.Rect, b xrect.Rect) int {
	xOverlap := lineOverlap(a.X(), a.X()+a.Width(), b.X(), b.X()+b.Width())
	yOverlap := lineOverlap(a.Y(), a.Y()+a.Height(), b.Y(), b.Y()+b.Height())
	return xOverlap * yOverlap
}

func EachAxisMostlyOverlaps(a xrect.Rect, b xrect.Rect) bool {
	x1, y1, w1, h1 := xrect.RectPieces(a)
	x2, y2, w2, h2 := xrect.RectPieces(b)
	iw := min(x1+w1-1, x2+w2-1) - max(x1, x2) + 1
	ih := min(y1+h1-1, y2+h2-1) - max(y1, y2) + 1
	fmt.Printf("%v %v: iw=%d ih=%d\n", a, b, iw, ih)
	return iw > w2*3/4 && ih > h2*3/4
}

// is >=50% of rect "a" inside of rect "b"
func RectMostlyInRect(a xrect.Rect, b xrect.Rect) bool {
	aArea := a.Height() * a.Width()
	bArea := b.Height() * b.Width()
	overlapArea := xrect.IntersectArea(a, b)
	return overlapArea == bArea || (overlapArea*1000 > aArea*505)
}

// is rect "a" inside of rect "b"
func RectInRect(a xrect.Rect, b xrect.Rect) bool {
	result := a.X() >= b.X() && (a.X()+a.X()+a.Width() <= b.X()+b.Width()) && a.Y() >= b.Y() && (a.Y()+a.Y()+a.Width() <= b.Y()+b.Width())
	fmt.Printf("RectInRect(%v, %v): %v\n", a, b, result)
	return result
}

type RectMatch func(geom xrect.Rect) bool

func (desk *Desktop) NextMatchingWindow(activeWindow xproto.Window, matches RectMatch) xproto.Window {
	matchingWindows := make([]xproto.Window, 0, 8)
	i := 0
	activeWinIndex := -1

	for win := range(desk.WindowsOnCurrentDesktop()) {
		geom := desk.GetGeometryForWindow(win)
		if matches(geom) {
			fmt.Printf("%d: Win(id=0x%x: x=%d,y=%d,w=%d,h=%d))\n", i, win, geom.X(), geom.Y(), geom.Width(), geom.Height())
			matchingWindows = append(matchingWindows, win)
			if win == activeWindow {
				activeWinIndex = i
			}
			i++
		} else {
			fmt.Printf("-: Win(id=0x%x: x=%d,y=%d,w=%d,h=%d))\n", win, geom.X(), geom.Y(), geom.Width(), geom.Height())
		}
	}
	if len(matchingWindows) == 0 {
		return 0
	}
	var nextWinIndex int
	if activeWinIndex > -1 {
		nextWinIndex = (activeWinIndex+1) % len(matchingWindows)
	} else {
		nextWinIndex = len(matchingWindows)-1
	}
	return matchingWindows[nextWinIndex]
}


