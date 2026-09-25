package platform

import (
	"math"

	qt "github.com/mappu/miqt/qt"
)

// RoundedWindowRegion returns a rounded-rectangle region for a widget mask: three
// bands plus one row per corner step. A QRegion is a 1 bit mask, so the arc is a
// staircase (the same result the Windows 10 window region produces). The rectangle is
// in the widget's logical coordinates.
func RoundedWindowRegion(width, height, radius int) *qt.QRegion {
	if radius*2 > width {
		radius = width / 2
	}
	if radius*2 > height {
		radius = height / 2
	}
	if radius <= 0 {
		return qt.NewQRegion2(0, 0, width, height)
	}

	// Everything but the four corner squares.
	region := qt.NewQRegion2(radius, 0, width-2*radius, radius)
	region = region.United(qt.NewQRegion2(0, radius, width, height-2*radius))
	region = region.United(qt.NewQRegion2(radius, height-radius, width-2*radius, radius))

	// One row per corner step: the inset follows the circle, x = r - sqrt(r² - dy²).
	for i := 0; i < radius; i++ {
		dy := float64(radius) - float64(i) - 0.5
		inset := int(math.Round(float64(radius) - math.Sqrt(float64(radius*radius)-dy*dy)))
		if inset < 0 {
			inset = 0
		}
		rowWidth := width - 2*inset
		if rowWidth <= 0 {
			continue
		}
		region = region.United(qt.NewQRegion2(inset, i, rowWidth, 1))
		region = region.United(qt.NewQRegion2(inset, height-1-i, rowWidth, 1))
	}
	return region
}

// IsX11 reports whether the application runs on an X11 server (the xcb platform
// plugin). Wayland sessions that run X11 clients through XWayland report xcb as well.
func IsX11() bool {
	return qt.QGuiApplication_PlatformName() == "xcb"
}

// FloatingWindowFlags returns the flags of a floating helper window - a tool tip, a
// teaching tip: frameless, never in the taskbar, never taking focus. Callers add what
// else they need (WindowStaysOnTopHint, NoDropShadowWindowHint).
//
// Qt::Tool is the natural flag and is used where the platform places tool windows. On
// X11 it becomes _NET_WM_WINDOW_TYPE_UTILITY, which some window managers never place at
// all: WSLg's Weston parks such a window off screen (at -32768) while still reporting it
// as viewable, so the window simply never appears. A transient dialog is placed
// everywhere and - together with ShowWithoutActivating and WindowStaysOnTopHint, which
// the callers set - behaves like the tool window it replaces. Override-redirect types
// (Qt::ToolTip, Qt::Popup) are not an alternative there: WSLg does not present those at
// all.
func FloatingWindowFlags() qt.WindowType {
	if IsX11() {
		return qt.Dialog | qt.FramelessWindowHint | qt.WindowStaysOnTopHint
	}
	return qt.Tool | qt.FramelessWindowHint
}
