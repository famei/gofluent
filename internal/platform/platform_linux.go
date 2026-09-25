//go:build linux

package platform

import (
	qt "github.com/mappu/miqt/qt"
)

// This file holds the APIs that only the Linux build implements: shaping a window
// through Qt, which is what X11 needs because it has neither DWM nor a window manager
// that rounds a frameless window for us.

func init() {
	// Install the Linux implementation of the platform API (see platform_other.go for the
	// default no-op the other systems keep).
	setWindowRoundedMask = linuxSetWindowRoundedMask
}

// linuxSetWindowRoundedMask rounds the corners of a top-level window with a widget mask.
// The mask clips the window (and everything in it), so the corners are cut out of the
// window itself; a QRegion is a 1 bit mask, so the arc is a staircase rather than an
// anti-aliased curve. Call it again after the window is resized: the mask is built for
// the current size.
func linuxSetWindowRoundedMask(w *qt.QWidget, radius int) bool {
	if w == nil {
		return false
	}
	if radius <= 0 {
		w.ClearMask()
		return true
	}
	w.SetMaskWithMask(RoundedWindowRegion(w.Width(), w.Height(), radius))
	return true
}
