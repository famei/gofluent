// Command time_picker migrates examples/date_time/time_picker/demo.py.
package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/famei/gofluent/components/date_time"
	"github.com/famei/gofluent/examples/internal/demo"
	qt "github.com/mappu/miqt/qt"
)

// SecondsFormatter renders a seconds value with a 秒 suffix.
type SecondsFormatter struct{}

func (SecondsFormatter) Encode(value interface{}) string {
	return fmt.Sprint(value) + "秒"
}

func (SecondsFormatter) Decode(value string) interface{} {
	n, _ := strconv.Atoi(strings.TrimSuffix(value, "秒"))
	return n
}

// Demo is the time picker example window.
type Demo struct {
	*qt.QWidget
	datePicker1 *date_time.DatePicker
	datePicker2 *date_time.ZhDatePicker
	timePicker1 *date_time.AMTimePicker
	timePicker2 *date_time.TimePicker
	timePicker3 *date_time.TimePicker
}

func newDemo() *Demo {
	d := &Demo{QWidget: qt.NewQWidget2()}
	d.SetStyleSheet("Demo{background: white}")
	// setTheme(Theme.DARK)
	// d.SetStyleSheet("Demo{background: rgb(32, 32, 32)}")

	vBoxLayout := qt.NewQVBoxLayout(d.QWidget)

	// create pickers with scroll button repeat enabled by default
	d.datePicker1 = date_time.NewDatePicker(d.QWidget, date_time.DatePickerMMDDYYYY, true)
	d.datePicker2 = date_time.NewZhDatePicker(d.QWidget)
	d.timePicker1 = date_time.NewAMTimePicker(d.QWidget, false)
	d.timePicker2 = date_time.NewTimePicker(d.QWidget, false)
	d.timePicker3 = date_time.NewTimePicker(d.QWidget, true)

	// enable reset button
	// d.datePicker1.SetResetEnabled(true)
	// d.timePicker1.SetResetEnabled(true)

	// disable scroll button repeat if needed
	// d.datePicker1.SetScrollButtonRepeatEnabled(false)
	// d.timePicker1.SetScrollButtonRepeatEnabled(false)

	// customize column format
	secondsFormatter := SecondsFormatter{}
	d.timePicker3.SetColumnFormatter(2, secondsFormatter)

	d.datePicker1.OnDateChanged = func(t *qt.QDate) { fmt.Println(t.ToString()) }
	d.datePicker2.OnDateChanged = func(t *qt.QDate) { fmt.Println(t.ToString()) }
	d.timePicker1.OnTimeChanged = func(t *qt.QTime) { fmt.Println(t.ToString()) }
	d.timePicker2.OnTimeChanged = func(t *qt.QTime) { fmt.Println(t.ToString()) }
	d.timePicker3.OnTimeChanged = func(t *qt.QTime) { fmt.Println(t.ToString()) }

	// set current date/time
	// d.datePicker1.SetDate(qt.QDate_CurrentDate())
	// d.timePicker1.SetTime(qt.NewQTime4(13, 15, 0))
	// d.timePicker2.SetTime(qt.NewQTime4(13, 15, 0))

	d.Resize(500, 500)
	vBoxLayout.AddWidget3(d.datePicker1.QWidget, 0, qt.AlignHCenter)
	vBoxLayout.AddWidget3(d.datePicker2.QWidget, 0, qt.AlignHCenter)
	vBoxLayout.AddWidget3(d.timePicker1.QWidget, 0, qt.AlignHCenter)
	vBoxLayout.AddWidget3(d.timePicker2.QWidget, 0, qt.AlignHCenter)
	vBoxLayout.AddWidget3(d.timePicker3.QWidget, 0, qt.AlignHCenter)

	return d
}

func main() {
	demo.Run(func() *qt.QWidget { return newDemo().QWidget })
}
