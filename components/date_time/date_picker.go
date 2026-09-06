package date_time

import (
	"fmt"
	"strconv"
	"strings"

	qt "github.com/mappu/miqt/qt"
)

// rangeSlice builds []interface{} from start (inclusive) to end (exclusive).
func rangeSlice(start, end int) []interface{} {
	out := make([]interface{}, 0, maxInt(0, end-start))
	for i := start; i < end; i++ {
		out = append(out, i)
	}
	return out
}

func anyNonEmpty(values []string) bool {
	for _, v := range values {
		if v != "" {
			return true
		}
	}
	return false
}

func cloneDate(d *qt.QDate) *qt.QDate {
	if d == nil || !d.IsValid() {
		return qt.NewQDate()
	}
	return qt.NewQDate2(d.Year(), d.Month(), d.Day())
}

// MonthFormatter encodes a month number to its English name and back.
type MonthFormatter struct {
	months []string
}

// NewMonthFormatter builds a month formatter.
func NewMonthFormatter() *MonthFormatter {
	return &MonthFormatter{months: []string{
		"January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December",
	}}
}

func (f *MonthFormatter) Encode(value interface{}) string {
	m := asInt(value)
	if 1 <= m && m <= 12 {
		return f.months[m-1]
	}
	return fmt.Sprint(value)
}

func (f *MonthFormatter) Decode(value string) interface{} {
	for i, m := range f.months {
		if m == value {
			return i + 1
		}
	}
	n, _ := strconv.Atoi(value)
	return n
}

// DatePickerBase is the base class of the date pickers.
type DatePickerBase struct {
	*PickerBase
	date           *qt.QDate
	calendar       *qt.QCalendar
	yearFormatter  PickerColumnFormatter
	monthFormatter PickerColumnFormatter
	dayFormatter   PickerColumnFormatter

	OnDateChanged func(*qt.QDate)
}

// NewDatePickerBase builds a date picker base.
func NewDatePickerBase(parent *qt.QWidget) *DatePickerBase {
	return &DatePickerBase{
		PickerBase: NewPickerBase(parent),
		date:       qt.NewQDate(),
		calendar:   qt.NewQCalendar(),
	}
}

// Date returns the current date.
func (b *DatePickerBase) Date() *qt.QDate { return b.date }

// SetDate sets the current date (overridden by DatePicker).
func (b *DatePickerBase) SetDate(date *qt.QDate) {}

// SetYearFormatter sets the year column formatter.
func (b *DatePickerBase) SetYearFormatter(f PickerColumnFormatter) { b.yearFormatter = f }

// SetMonthFormatter sets the month column formatter.
func (b *DatePickerBase) SetMonthFormatter(f PickerColumnFormatter) { b.monthFormatter = f }

// SetDayFormatter sets the day column formatter.
func (b *DatePickerBase) SetDayFormatter(f PickerColumnFormatter) { b.dayFormatter = f }

// YearFormatter returns the year column formatter.
func (b *DatePickerBase) YearFormatter() PickerColumnFormatter {
	if b.yearFormatter != nil {
		return b.yearFormatter
	}
	return DigitFormatter{}
}

// DayFormatter returns the day column formatter.
func (b *DatePickerBase) DayFormatter() PickerColumnFormatter {
	if b.dayFormatter != nil {
		return b.dayFormatter
	}
	return DigitFormatter{}
}

// MonthFormatter returns the month column formatter.
func (b *DatePickerBase) MonthFormatter() PickerColumnFormatter {
	if b.monthFormatter != nil {
		return b.monthFormatter
	}
	return NewMonthFormatter()
}

// Reset clears the date and resets every column.
func (b *DatePickerBase) Reset() {
	b.date = qt.NewQDate()
	b.PickerBase.Reset()
}

// Date picker formats.
const (
	DatePickerMMDDYYYY = iota
	DatePickerYYYYMMDD
)

// DatePicker is the fluent date picker.
type DatePicker struct {
	*DatePickerBase
	monthIndex int
	dayIndex   int
	yearIndex  int
	dateFormat int
	isTight    bool
	monthLabel string
	yearLabel  string
	dayLabel   string
}

// NewDatePicker builds a date picker.
func NewDatePicker(parent *qt.QWidget, format int, isMonthTight bool) *DatePicker {
	d := &DatePicker{
		DatePickerBase: NewDatePickerBase(parent),
		monthLabel:     "month",
		yearLabel:      "year",
		dayLabel:       "day",
		isTight:        isMonthTight,
	}
	d.panelInitValueFunc = d.panelInitialValue
	d.confirmValuesFunc = d.confirmValues
	d.columnChangedFunc = d.columnChanged
	d.SetDateFormat(format)
	return d
}

// SetDateFormat rebuilds the columns for the given format.
func (d *DatePicker) SetDateFormat(format int) {
	d.ClearColumns()
	cur := qt.QDate_CurrentDate()
	y := cur.Year()
	d.dateFormat = format

	if format == DatePickerMMDDYYYY {
		d.monthIndex = 0
		d.dayIndex = 1
		d.yearIndex = 2

		d.AddColumn(d.monthLabel, rangeSlice(1, 13), 80, int(qt.AlignLeft), d.MonthFormatter())
		d.AddColumn(d.dayLabel, rangeSlice(1, 32), 80, int(qt.AlignCenter), d.DayFormatter())
		d.AddColumn(d.yearLabel, rangeSlice(y-100, y+101), 80, int(qt.AlignCenter), d.YearFormatter())
	} else {
		d.yearIndex = 0
		d.monthIndex = 1
		d.dayIndex = 2

		d.AddColumn(d.yearLabel, rangeSlice(y-100, y+101), 80, int(qt.AlignCenter), d.YearFormatter())
		d.AddColumn(d.monthLabel, rangeSlice(1, 13), 80, int(qt.AlignCenter), d.MonthFormatter())
		d.AddColumn(d.dayLabel, rangeSlice(1, 32), 80, int(qt.AlignCenter), d.DayFormatter())
	}

	d.SetColumnWidth(d.monthIndex, d.monthColumnWidth())
}

// PanelInitialValue returns the current date values (or today's).
func (d *DatePicker) PanelInitialValue() []string { return d.panelInitialValue() }

func (d *DatePicker) panelInitialValue() []string {
	if anyNonEmpty(d.Value()) {
		return d.Value()
	}
	date := qt.QDate_CurrentDate()
	y := d.EncodeValue(d.yearIndex, date.Year())
	m := d.EncodeValue(d.monthIndex, date.Month())
	day := d.EncodeValue(d.dayIndex, date.Day())
	if d.dateFormat == DatePickerYYYYMMDD {
		return []string{y, m, day}
	}
	return []string{m, day, y}
}

// SetMonthTight toggles the tight month column.
func (d *DatePicker) SetMonthTight(isTight bool) {
	if d.isTight == isTight {
		return
	}
	d.isTight = isTight
	d.SetColumnWidth(d.monthIndex, d.monthColumnWidth())
}

func (d *DatePicker) monthColumnWidth() int {
	fm := d.FontMetrics()
	wm := 0
	for _, it := range d.columns[d.monthIndex].Items() {
		wm = maxInt(wm, fm.Width(it))
	}
	wm += 20

	if d.monthLabel == "month" {
		return wm + 49
	}
	if d.isTight {
		return maxInt(80, wm)
	}
	return wm + 49
}

// OnColumnValueChanged refreshes the day column when month/year change.
func (d *DatePicker) OnColumnValueChanged(panel *PickerPanel, index int, value string) {
	d.columnChanged(panel, index, value)
}

func (d *DatePicker) columnChanged(panel *PickerPanel, index int, value string) {
	if index == d.dayIndex {
		return
	}
	month := asInt(d.DecodeValue(d.monthIndex, panel.ColumnValue(d.monthIndex)))
	year := asInt(d.DecodeValue(d.yearIndex, panel.ColumnValue(d.yearIndex)))
	days := d.calendar.DaysInMonth2(month, year)

	c := panel.Column(d.dayIndex)
	if c == nil {
		return
	}
	day := ""
	if item := c.CurrentItem(); item != nil {
		day = item.Text()
	}

	d.SetColumnItems(d.dayIndex, rangeSlice(1, days+1))
	c.SetItems(d.columns[d.dayIndex].Items())
	c.SetSelectedItem(day)
}

func (d *DatePicker) confirmValues(values []string) {
	year := asInt(d.DecodeValue(d.yearIndex, values[d.yearIndex]))
	month := asInt(d.DecodeValue(d.monthIndex, values[d.monthIndex]))
	day := asInt(d.DecodeValue(d.dayIndex, values[d.dayIndex]))

	date := qt.NewQDate2(year, month, day)
	defer date.Delete()
	od := d.date
	d.SetDate(date)

	if od == nil || od.Year() != year || od.Month() != month || od.Day() != day {
		if d.OnDateChanged != nil {
			d.OnDateChanged(cloneDate(date))
		}
	}
}

// SetDate sets the selected date.
func (d *DatePicker) SetDate(date *qt.QDate) {
	if date == nil || !date.IsValid() || date.IsNull() {
		return
	}
	d.date = cloneDate(date)
	d.SetColumnValue(d.monthIndex, date.Month())
	d.SetColumnValue(d.dayIndex, date.Day())
	d.SetColumnValue(d.yearIndex, date.Year())
	d.SetColumnItems(d.dayIndex, rangeSlice(1, date.DaysInMonth()+1))
}

// ZhFormatter is the base Chinese date formatter.
type ZhFormatter struct {
	suffix string
}

func (f *ZhFormatter) Encode(value interface{}) string { return fmt.Sprint(value) + f.suffix }
func (f *ZhFormatter) Decode(value string) interface{} {
	n, _ := strconv.Atoi(strings.TrimSuffix(value, f.suffix))
	return n
}

// ZhYearFormatter formats a year with the 年 suffix.
type ZhYearFormatter struct{ ZhFormatter }

// NewZhYearFormatter builds a Chinese year formatter.
func NewZhYearFormatter() *ZhYearFormatter {
	return &ZhYearFormatter{ZhFormatter: ZhFormatter{suffix: "年"}}
}

// ZhMonthFormatter formats a month with the 月 suffix.
type ZhMonthFormatter struct{ ZhFormatter }

// NewZhMonthFormatter builds a Chinese month formatter.
func NewZhMonthFormatter() *ZhMonthFormatter {
	return &ZhMonthFormatter{ZhFormatter: ZhFormatter{suffix: "月"}}
}

// ZhDayFormatter formats a day with the 日 suffix.
type ZhDayFormatter struct{ ZhFormatter }

// NewZhDayFormatter builds a Chinese day formatter.
func NewZhDayFormatter() *ZhDayFormatter {
	return &ZhDayFormatter{ZhFormatter: ZhFormatter{suffix: "日"}}
}

// ZhDatePicker is the Chinese date picker (YYYY年MM月DD日).
type ZhDatePicker struct {
	*DatePicker
}

// NewZhDatePicker builds a Chinese date picker.
func NewZhDatePicker(parent *qt.QWidget) *ZhDatePicker {
	d := &ZhDatePicker{DatePicker: NewDatePicker(parent, DatePickerYYYYMMDD, true)}
	d.monthLabel = "月"
	d.yearLabel = "年"
	d.dayLabel = "日"
	d.SetDayFormatter(NewZhDayFormatter())
	d.SetYearFormatter(NewZhYearFormatter())
	d.SetMonthFormatter(NewZhMonthFormatter())
	d.SetDateFormat(DatePickerYYYYMMDD)
	return d
}
