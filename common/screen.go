package common

import (
	qt "github.com/mappu/miqt/qt"
)

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
