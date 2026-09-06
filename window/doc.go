// Package window ports qfluentwidgets/window to Go: the fluent window and
// title-bar hierarchy, the splash screen, and the pop-up stacked widget.
//
// The Python package builds on the external `qframelesswindow` library for the
// native frame, DWM/Mica backdrop and system title-bar buttons. Those native
// effects are not reachable through miqt, so the title bar is re-implemented
// here as a plain child widget and the Mica effect degrades to a solid
// light/dark background (see MIGRATION_GUIDE §10).
package window

import (
	qt "github.com/mappu/miqt/qt"
)

// cloneColor returns an independent copy of c so callers can mutate the result
// without affecting a shared QColor (the QColor copy constructor).
func cloneColor(c *qt.QColor) *qt.QColor {
	if c == nil {
		return qt.NewQColor()
	}
	return qt.NewQColor9(c)
}
