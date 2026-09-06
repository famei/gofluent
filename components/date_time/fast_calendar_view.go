package date_time

import (
	qt "github.com/mappu/miqt/qt"
)

// FastCalendarView is the "Pro" calendar view shown inside a Flyout. In the
// Python port it derives from FlyoutViewBase and paints a rounded background;
// in Go it reuses CalendarView (the Flyout provides the rounded frame/shadow).
type FastCalendarView struct {
	*CalendarView
}

// NewFastCalendarView builds a fast calendar view.
func NewFastCalendarView(parent *qt.QWidget) *FastCalendarView {
	return &FastCalendarView{CalendarView: NewCalendarView(parent)}
}
