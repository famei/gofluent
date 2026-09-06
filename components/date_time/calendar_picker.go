package date_time

import (
	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// CalendarPicker is a push button that opens a CalendarView pop-up and shows
// the selected date.
type CalendarPicker struct {
	*qt.QPushButton
	date           *qt.QDate
	dateFormat     string
	isResetEnabled bool

	OnDateChanged func(*qt.QDate)
}

// NewCalendarPicker builds a calendar picker.
func NewCalendarPicker(parent *qt.QWidget) *CalendarPicker {
	w := &CalendarPicker{
		QPushButton: qt.NewQPushButton(parent),
		date:        qt.NewQDate(),
		dateFormat:  "yyyy-MM-dd",
	}
	// The object name makes the calendar_picker.qss `CalendarPicker` selector
	// (rewritten to QPushButton#calendarPicker by common.RenderQss) match this
	// button, so the full Fluent look — 14px font, rounded border, right padding
	// for the calendar icon, hover/pressed/disabled states and the hasDate text
	// colour — comes from the QSS instead of a hand-maintained copy.
	w.SetObjectName("calendarPicker")
	w.SetText(qt.QCoreApplication_Translate("CalendarPicker", "Pick a date"))
	common.FluentStyleSheet(common.FluentCalendarPicker).Apply(w.QWidget, common.ThemeAuto)
	w.OnClicked(w.showCalendarView)
	w.installPaintEvent()
	return w
}

// Date returns the selected date.
func (w *CalendarPicker) Date() *qt.QDate { return w.date }

// SetDate sets the selected date.
func (w *CalendarPicker) SetDate(date *qt.QDate) { w.onDateChanged(date) }

// Reset clears the selected date.
func (w *CalendarPicker) Reset() {
	w.date = qt.NewQDate()
	w.SetText(qt.QCoreApplication_Translate("CalendarPicker", "Pick a date"))
	w.SetProperty("hasDate", qt.NewQVariant11(false))
	w.SetStyle(qt.QApplication_Style())
	w.Update()
}

// DateFormat returns the date format string.
func (w *CalendarPicker) DateFormat() string { return w.dateFormat }

// SetDateFormat sets the date format string ("" resets to the placeholder).
func (w *CalendarPicker) SetDateFormat(format string) {
	w.dateFormat = format
	if w.date.IsValid() {
		w.SetText(w.dateText())
	}
}

// IsResetEnabled reports whether the reset button is enabled.
func (w *CalendarPicker) IsResetEnabled() bool { return w.isResetEnabled }

// SetResetEnabled toggles the reset button.
func (w *CalendarPicker) SetResetEnabled(isEnabled bool) { w.isResetEnabled = isEnabled }

func (w *CalendarPicker) dateText() string {
	if w.dateFormat != "" {
		return w.date.ToStringWithFormat(w.dateFormat)
	}
	return w.date.ToString6(qt.SystemLocaleDate)
}

func (w *CalendarPicker) showCalendarView() {
	view := NewCalendarView(w.Window())
	view.SetResetEnabled(w.isResetEnabled)
	view.OnResetted = func() { w.Reset() }
	view.OnDateChanged = w.onDateChanged

	if w.date.IsValid() {
		view.SetDate(w.date)
	}

	sh := view.SizeHint()
	x := w.Width()/2 - sh.Width()/2
	p := qt.NewQPoint2(x, w.Height())
	pos := w.MapToGlobal(p)
	p.Delete()
	view.Exec(pos, true)
}

func (w *CalendarPicker) onDateChanged(date *qt.QDate) {
	if date == nil {
		return
	}
	w.date = cloneDate(date)
	w.SetText(w.dateText())
	w.SetProperty("hasDate", qt.NewQVariant11(true))
	w.SetStyle(qt.QApplication_Style())
	w.Update()

	if w.OnDateChanged != nil {
		w.OnDateChanged(cloneDate(date))
	}
}

func (w *CalendarPicker) installPaintEvent() {
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
		super(event)
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing)

		if !w.hasDate() {
			painter.SetOpacity(0.6)
		}
		const size = 12
		rect := qt.NewQRectF4(float64(w.Width()-23), float64(w.Height())/2-size/2, size, size)
		defer rect.Delete()
		common.Calendar.Render(painter, rect, common.ThemeAuto)
		painter.End()
	})
}

func (w *CalendarPicker) hasDate() bool {
	v := w.Property("hasDate")
	return v != nil && v.ToBool()
}

// FastCalendarPicker is the "Pro" calendar picker that shows the view inside a
// Flyout.
type FastCalendarPicker struct {
	*CalendarPicker
	flyoutAnimationType int
}

// NewFastCalendarPicker builds a fast calendar picker.
func NewFastCalendarPicker(parent *qt.QWidget) *FastCalendarPicker {
	return &FastCalendarPicker{
		CalendarPicker:      NewCalendarPicker(parent),
		flyoutAnimationType: 0,
	}
}

// SetFlyoutAnimationType stores the flyout animation type (kept for API
// compatibility; the flyout animation is simplified away).
func (w *FastCalendarPicker) SetFlyoutAnimationType(aniType int) { w.flyoutAnimationType = aniType }

func (w *FastCalendarPicker) showCalendarView() {
	view := NewFastCalendarView(w.Window())
	view.SetResetEnabled(w.isResetEnabled)
	view.OnResetted = func() { w.Reset() }
	view.OnDateChanged = w.onDateChanged

	if w.date.IsValid() {
		view.SetDate(w.date)
	}

	p := qt.NewQPoint2(0, w.Height())
	pos := w.MapToGlobal(p)
	p.Delete()
	view.Exec(pos, true)
}
