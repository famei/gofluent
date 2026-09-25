package common

import (
	qt "github.com/mappu/miqt/qt"
)

// MoveWindowContentTo moves a top level window so that its content - not its frame - sits at
// x,y.
//
// QWidget.move() positions the window *including its frame*, which is invisible on Windows
// (a frameless window has no frame there) but not on X11: a window manager may add frame
// extents even to a frameless window (WSLg's XWM reports 38/59), and the popup then appears
// that far away from the widget it is anchored to.
func MoveWindowContentTo(w *qt.QWidget, x, y int) {
	if w == nil {
		return
	}
	frame := w.FrameGeometry() // GoGC-armed — do NOT Delete
	content := w.Geometry()    // GoGC-armed — do NOT Delete
	w.Move(x-(content.X()-frame.X()), y-(content.Y()-frame.Y()))
}

// GetCurrentScreen returns the screen under the mouse cursor, or nil.
func GetCurrentScreen() *qt.QScreen {
	cursorPos := qt.QCursor_Pos() // GoGC-armed — do NOT Delete

	for _, s := range qt.QGuiApplication_Screens() {
		g := s.Geometry() // GoGC-armed — do NOT Delete
		contains := g.ContainsWithQPoint(cursorPos)
		if contains {
			return s
		}
	}
	return nil
}

// GetCurrentScreenGeometry returns the geometry of the current screen. When
// available is true the usable (taskbar-excluded) geometry is returned.
func GetCurrentScreenGeometry(available bool) *qt.QRect {
	screen := GetCurrentScreen()
	if screen == nil {
		screen = qt.QGuiApplication_PrimaryScreen()
	}
	if screen == nil {
		return qt.NewQRect4(0, 0, 1920, 1080)
	}
	if available {
		return screen.AvailableGeometry()
	}
	return screen.Geometry()
}

// GetWidgetScreenGeometry returns the geometry of the screen a widget is on, falling
// back to the screen under the cursor and then to the primary screen. A popup anchored
// to a widget has to be clamped to the widget's own screen: with several monitors the
// cursor can sit on another one (while the pointer travels to the popup, or when the
// popup is shown programmatically), and clamping there would drag the popup onto the
// wrong monitor.
func GetWidgetScreenGeometry(w *qt.QWidget, available bool) *qt.QRect {
	if w != nil {
		if screen := w.Screen(); screen != nil {
			if available {
				return screen.AvailableGeometry()
			}
			return screen.Geometry()
		}
	}
	return GetCurrentScreenGeometry(available)
}
