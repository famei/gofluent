// Command fast_calendar_picker migrates examples/date_time/fast_calendar_picker/demo.py.
package main

import (
	"fmt"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/date_time"
	"github.com/famei/gofluent/examples/internal/demo"
	qt "github.com/mappu/miqt/qt"
)

// Demo is the fast calendar picker example window.
type Demo struct {
	*qt.QWidget
	picker *date_time.FastCalendarPicker
}

func newDemo() *Demo {
	d := &Demo{QWidget: qt.NewQWidget2()}
	// setTheme(Theme.DARK)
	d.SetStyleSheet("Demo{background: white}")

	d.picker = date_time.NewFastCalendarPicker(d.QWidget)
	d.picker.OnDateChanged = func(date *qt.QDate) {
		fmt.Println(date.ToString())
	}

	// enable reset button
	// d.picker.SetResetEnabled(true)

	// customize animation
	// d.picker.SetFlyoutAnimationType(...)

	// set date
	// d.picker.SetDate(qt.NewQDate2(2023, 5, 30))

	// customize date format
	// d.picker.SetDateFormat("yyyy-M-d")

	layout := qt.NewQHBoxLayout(d.QWidget)
	layout.AddWidget3(d.picker.QWidget, 0, qt.AlignCenter)
	d.Resize(500, 500)
	return d
}

func main() {
	demo.SetupHighDPI()
	app := demo.NewApp()

	// install translator
	translator := common.NewFluentTranslator(nil, app.QObject)
	qt.QCoreApplication_InstallTranslator(translator.QTranslator)

	w := newDemo()
	w.Show()
	demo.Exec()
}
