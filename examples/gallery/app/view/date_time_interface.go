package view

import (
	"github.com/famei/gofluent/components/date_time"
	gallerycommon "github.com/famei/gofluent/examples/gallery/app/common"
	qt "github.com/mappu/miqt/qt"
)

// DateTimeInterface is the "Date & time" gallery page (port of
// app/view/date_time_interface.py).
type DateTimeInterface struct {
	*GalleryInterface
}

// NewDateTimeInterface builds the date time interface.
func NewDateTimeInterface(parent *qt.QWidget) *DateTimeInterface {
	t := gallerycommon.NewTranslator()
	i := &DateTimeInterface{GalleryInterface: NewGalleryInterface(t.DateTime, "github.com/famei/gofluent/components/date_time", parent)}
	i.SetObjectName("dateTimeInterface")

	tr := func(s string) string { return gallerycommon.Tr("DateTimeInterface", s) }

	const src = "date_time/calendar_picker/main.go"

	i.AddExampleCard(tr("A simple CalendarPicker"), date_time.NewCalendarPicker(i.QWidget).QWidget, "components/date_time/calendar_picker.go", codeCalendarPicker, 0)
	i.AddExampleCard(tr("A fast CalendarPicker"), date_time.NewFastCalendarPicker(i.QWidget).QWidget, "components/date_time/calendar_picker.go", codeFastCalendarPicker, 0)

	w := date_time.NewCalendarPicker(i.QWidget)
	w.SetDateFormat("dd/MM/yyyy")
	i.AddExampleCard(tr("A CalendarPicker in another format"), w.QWidget, "components/date_time/calendar_picker.go", codeCalendarPickerFormat, 0)

	const tsrc = "date_time/time_picker/main.go"

	i.AddExampleCard(tr("A simple DatePicker"), date_time.NewDatePicker(i.QWidget, date_time.DatePickerYYYYMMDD, false).QWidget, "components/date_time/date_picker.go", codeDatePicker, 0)
	i.AddExampleCard(tr("A DatePicker in another format"), date_time.NewZhDatePicker(i.QWidget).QWidget, "components/date_time/date_picker.go", codeZhDatePicker, 0)
	i.AddExampleCard(tr("A simple TimePicker"), date_time.NewAMTimePicker(i.QWidget, false).QWidget, "components/date_time/time_picker.go", codeAMTimePicker, 0)
	i.AddExampleCard(tr("A TimePicker using a 24-hour clock"), date_time.NewTimePicker(i.QWidget, false).QWidget, "components/date_time/time_picker.go", codeTimePicker, 0)
	i.AddExampleCard(tr("A TimePicker with seconds column"), date_time.NewTimePicker(i.QWidget, true).QWidget, "components/date_time/time_picker.go", codeTimePickerSeconds, 0)
	return i
}
