package date_time

import (
	"fmt"
	"strconv"

	qt "github.com/mappu/miqt/qt"
)

func cloneTime(t *qt.QTime) *qt.QTime {
	if t == nil || !t.IsValid() {
		return qt.NewQTime()
	}
	return qt.NewQTime4(t.Hour(), t.Minute(), t.Second())
}

// MiniuteFormatter zero-pads minute/second values (the spelling matches the
// original Python class name).
type MiniuteFormatter struct{}

func (MiniuteFormatter) Encode(value interface{}) string {
	return fmt.Sprintf("%02d", asInt(value))
}
func (MiniuteFormatter) Decode(value string) interface{} {
	n, _ := strconv.Atoi(value)
	return n
}

// AMHourFormatter renders a 24-hour value as a 12-hour value (0/12 -> 12).
type AMHourFormatter struct{}

func (AMHourFormatter) Encode(value interface{}) string {
	h := asInt(value)
	if h == 0 || h == 12 {
		return "12"
	}
	return strconv.Itoa(h % 12)
}
func (AMHourFormatter) Decode(value string) interface{} {
	n, _ := strconv.Atoi(value)
	return n
}

// AMPMFormatter renders an hour value as "AM" or "PM".
type AMPMFormatter struct {
	am string
	pm string
}

// NewAMPMFormatter builds an AM/PM formatter.
func NewAMPMFormatter() *AMPMFormatter { return &AMPMFormatter{am: "AM", pm: "PM"} }

func (f *AMPMFormatter) Encode(value interface{}) string {
	s := fmt.Sprint(value)
	if _, err := strconv.Atoi(s); err != nil {
		return s
	}
	if asInt(value) < 12 {
		return f.am
	}
	return f.pm
}
func (f *AMPMFormatter) Decode(value string) interface{} { return value }

// TimePickerBase is the base class of the time pickers.
type TimePickerBase struct {
	*PickerBase
	time            *qt.QTime
	isSecondVisible bool

	OnTimeChanged func(*qt.QTime)
}

// NewTimePickerBase builds a time picker base.
func NewTimePickerBase(parent *qt.QWidget, showSeconds bool) *TimePickerBase {
	return &TimePickerBase{
		PickerBase:      NewPickerBase(parent),
		time:            qt.NewQTime(),
		isSecondVisible: showSeconds,
	}
}

// Time returns the current time.
func (b *TimePickerBase) Time() *qt.QTime { return b.time }

// SetTime sets the current time (overridden by the concrete pickers).
func (b *TimePickerBase) SetTime(time *qt.QTime) {}

// IsSecondVisible reports whether the seconds column is visible.
func (b *TimePickerBase) IsSecondVisible() bool { return b.isSecondVisible }

// SetSecondVisible toggles the seconds column (overridden by the concrete
// pickers).
func (b *TimePickerBase) SetSecondVisible(isVisible bool) {}

// Reset clears the time and resets every column.
func (b *TimePickerBase) Reset() {
	b.time = qt.NewQTime()
	b.PickerBase.Reset()
}

// TimePicker is the 24-hour fluent time picker.
type TimePicker struct {
	*TimePickerBase
}

// NewTimePicker builds a 24-hour time picker.
func NewTimePicker(parent *qt.QWidget, showSeconds bool) *TimePicker {
	t := &TimePicker{TimePickerBase: NewTimePickerBase(parent, showSeconds)}
	w := 80
	if !showSeconds {
		w = 120
	}
	t.AddColumn("hour", rangeSlice(0, 24), w, int(qt.AlignCenter), DigitFormatter{})
	t.AddColumn("minute", rangeSlice(0, 60), w, int(qt.AlignCenter), MiniuteFormatter{})
	t.AddColumn("second", rangeSlice(0, 60), w, int(qt.AlignCenter), MiniuteFormatter{})
	t.SetColumnVisible(2, showSeconds)

	t.panelInitValueFunc = t.panelInitialValue
	t.confirmValuesFunc = t.confirmValues
	return t
}

// SetTime sets the selected time.
func (t *TimePicker) SetTime(time *qt.QTime) {
	if time == nil || !time.IsValid() || time.IsNull() {
		return
	}
	t.time = cloneTime(time)
	t.SetColumnValue(0, time.Hour())
	t.SetColumnValue(1, time.Minute())
	t.SetColumnValue(2, time.Second())
}

// SetSecondVisible toggles the seconds column.
func (t *TimePicker) SetSecondVisible(isVisible bool) {
	t.isSecondVisible = isVisible
	t.SetColumnVisible(2, isVisible)

	w := 80
	if !isVisible {
		w = 120
	}
	for _, button := range t.columns {
		button.SetFixedWidth(w)
	}
}

func (t *TimePicker) confirmValues(values []string) {
	t.PickerBase.confirmValues(values)

	h := asInt(t.DecodeValue(0, values[0]))
	m := asInt(t.DecodeValue(1, values[1]))
	s := 0
	if len(values) == 3 {
		s = asInt(t.DecodeValue(2, values[2]))
	}

	time := qt.NewQTime4(h, m, s)
	defer time.Delete()
	ot := t.time
	t.SetTime(time)

	if ot == nil || ot.Hour() != h || ot.Minute() != m || ot.Second() != s {
		if t.OnTimeChanged != nil {
			t.OnTimeChanged(cloneTime(time))
		}
	}
}

func (t *TimePicker) panelInitialValue() []string {
	if anyNonEmpty(t.Value()) {
		return t.Value()
	}
	time := qt.QTime_CurrentTime()
	h := t.EncodeValue(0, time.Hour())
	m := t.EncodeValue(1, time.Minute())
	s := t.EncodeValue(2, time.Second())
	if t.isSecondVisible {
		return []string{h, m, s}
	}
	return []string{h, m}
}

// AMTimePicker is the 12-hour (AM/PM) fluent time picker.
type AMTimePicker struct {
	*TimePickerBase
}

// NewAMTimePicker builds an AM/PM time picker.
func NewAMTimePicker(parent *qt.QWidget, showSeconds bool) *AMTimePicker {
	t := &AMTimePicker{TimePickerBase: NewTimePickerBase(parent, showSeconds)}
	t.AddColumn("hour", rangeSlice(1, 13), 80, int(qt.AlignCenter), AMHourFormatter{})
	t.AddColumn("minute", rangeSlice(0, 60), 80, int(qt.AlignCenter), MiniuteFormatter{})
	t.AddColumn("second", rangeSlice(0, 60), 80, int(qt.AlignCenter), MiniuteFormatter{})
	t.SetColumnVisible(2, showSeconds)
	t.AddColumn("AM", toStringSlice([]string{"AM", "PM"}), 80, int(qt.AlignCenter), NewAMPMFormatter())

	t.panelInitValueFunc = t.panelInitialValue
	t.confirmValuesFunc = t.confirmValues
	return t
}

// SetSecondVisible toggles the seconds column.
func (t *AMTimePicker) SetSecondVisible(isVisible bool) {
	t.isSecondVisible = isVisible
	t.SetColumnVisible(2, isVisible)
}

// SetTime sets the selected time.
func (t *AMTimePicker) SetTime(time *qt.QTime) {
	if time == nil || !time.IsValid() || time.IsNull() {
		return
	}
	t.time = cloneTime(time)
	t.SetColumnValue(0, time.Hour())
	t.SetColumnValue(1, time.Minute())
	t.SetColumnValue(2, time.Second())
	t.SetColumnValue(3, time.Hour())
}

func (t *AMTimePicker) confirmValues(values []string) {
	t.PickerBase.confirmValues(values)

	var h, m, s int
	var p string
	if len(values) == 3 {
		h = asInt(t.DecodeValue(0, values[0]))
		m = asInt(t.DecodeValue(1, values[1]))
		p = values[2]
	} else {
		h = asInt(t.DecodeValue(0, values[0]))
		m = asInt(t.DecodeValue(1, values[1]))
		s = asInt(t.DecodeValue(2, values[2]))
		p = values[3]
	}

	if p == "AM" {
		if h == 12 {
			h = 0
		}
	} else if p == "PM" {
		if h != 12 {
			h += 12
		}
	}

	time := qt.NewQTime4(h, m, s)
	defer time.Delete()
	ot := t.time
	t.SetTime(time)

	if ot == nil || ot.Hour() != h || ot.Minute() != m || ot.Second() != s {
		if t.OnTimeChanged != nil {
			t.OnTimeChanged(cloneTime(time))
		}
	}
}

func (t *AMTimePicker) panelInitialValue() []string {
	if anyNonEmpty(t.Value()) {
		return t.Value()
	}
	time := qt.QTime_CurrentTime()
	h := t.EncodeValue(0, time.Hour())
	m := t.EncodeValue(1, time.Minute())
	s := t.EncodeValue(2, time.Second())
	p := t.EncodeValue(3, time.Hour())
	if t.isSecondVisible {
		return []string{h, m, s, p}
	}
	return []string{h, m, p}
}
