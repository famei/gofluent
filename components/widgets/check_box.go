package widgets

import (
	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// CheckBoxIcon identifies the two indicator glyphs drawn inside the check box.
type CheckBoxIcon string

const (
	// CheckBoxIconAccept is the check mark.
	CheckBoxIconAccept CheckBoxIcon = "Accept"
	// CheckBoxIconPartialAccept is the partial (tri-state) mark.
	CheckBoxIconPartialAccept CheckBoxIcon = "PartialAccept"
)

// render draws the indicator SVG into painter at rect. The color is reversed
// (white on light theme, black on dark theme) because it sits on a colored
// indicator.
func (c CheckBoxIcon) render(painter *qt.QPainter, rect *qt.QRectF) {
	color := common.GetIconColor(common.ThemeAuto, true)
	b := readEmbeddedImage("check_box/" + string(c) + "_" + color + ".svg")
	drawSvgBytes(b, painter, rect)
}

// CheckBoxState enumerates the visual states of a check box.
type CheckBoxState int

const (
	CheckBoxStateNormal CheckBoxState = iota
	CheckBoxStateHover
	CheckBoxStatePressed
	CheckBoxStateChecked
	CheckBoxStateCheckedHover
	CheckBoxStateCheckedPressed
	CheckBoxStateDisabled
	CheckBoxStateCheckedDisabled
)

// CheckBox is a fluent-styled check box.
//
// Constructors
//   - NewCheckBox(parent *qt.QWidget)
//   - NewCheckBoxText(text string, parent *qt.QWidget)
type CheckBox struct {
	*qt.QCheckBox
	isPressed         bool
	isHover           bool
	lightCheckedColor *qt.QColor
	darkCheckedColor  *qt.QColor
	lightTextColor    *qt.QColor
	darkTextColor     *qt.QColor
}

// NewCheckBox builds a check box.
func NewCheckBox(parent *qt.QWidget) *CheckBox {
	w := &CheckBox{QCheckBox: qt.NewQCheckBox(parent)}
	common.SetFont(w.QWidget, 14, int(qt.QFont__Normal))
	common.FluentStyleSheet(common.FluentCheckBox).Apply(w.QWidget, common.ThemeAuto)
	w.lightCheckedColor = qt.NewQColor()
	w.darkCheckedColor = qt.NewQColor()
	w.lightTextColor = qt.NewQColor3(0, 0, 0)
	w.darkTextColor = qt.NewQColor3(255, 255, 255)
	w.installEvents()
	return w
}

// NewCheckBoxText builds a check box with text.
func NewCheckBoxText(text string, parent *qt.QWidget) *CheckBox {
	w := NewCheckBox(parent)
	w.SetText(text)
	return w
}

func (w *CheckBox) installEvents() {
	w.OnMousePressEvent(func(super func(event *qt.QMouseEvent), event *qt.QMouseEvent) {
		w.isPressed = true
		super(event)
	})
	w.OnMouseReleaseEvent(func(super func(event *qt.QMouseEvent), event *qt.QMouseEvent) {
		w.isPressed = false
		super(event)
	})
	w.OnEnterEvent(func(super func(event *qt.QEvent), event *qt.QEvent) {
		super(event)
		w.isHover = true
		w.Update()
	})
	w.OnLeaveEvent(func(super func(event *qt.QEvent), event *qt.QEvent) {
		super(event)
		w.isHover = false
		w.Update()
	})
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
		super(event)
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing)

		// Match Python: use the style's real indicator sub-element rect instead
		// of a hard-coded (0,0,19,19) box, so the indicator follows the QSS
		// indicator size and vertical offset.
		opt := qt.NewQStyleOptionButton()
		defer opt.Delete()
		opt.InitFrom(w.QWidget)
		ind := w.Style().SubElementRect(qt.QStyle__SE_CheckBoxIndicator, opt.QStyleOption, w.QWidget) // GoGC-armed — do NOT Delete
		rect := qt.NewQRectF5(ind)
		defer rect.Delete()
		border := w.borderColor() // borrowed (may return the stored field color) — do NOT Delete
		bg := w.backgroundColor() // borrowed (may return the stored field color) — do NOT Delete
		bgBrush := qt.NewQBrush3(bg)
		defer bgBrush.Delete()
		painter.SetPen(border)
		painter.SetBrush(bgBrush)
		painter.DrawRoundedRect(rect, 4.5, 4.5)

		if !w.IsEnabled() {
			painter.SetOpacity(0.8)
		}

		switch w.CheckState() {
		case qt.Checked:
			CheckBoxIconAccept.render(painter, rect)
		case qt.PartiallyChecked:
			CheckBoxIconPartialAccept.render(painter, rect)
		}
		painter.End()
	})
}

// SetCheckedColor sets the indicator color in checked status.
func (w *CheckBox) SetCheckedColor(light, dark *qt.QColor) {
	w.lightCheckedColor = cloneColor(light)
	w.darkCheckedColor = cloneColor(dark)
	w.Update()
}

// SetTextColor sets the text color for light/dark themes.
func (w *CheckBox) SetTextColor(light, dark *qt.QColor) {
	w.lightTextColor = cloneColor(light)
	w.darkTextColor = cloneColor(dark)
	common.SetCustomStyleSheet(w.QWidget,
		"CheckBox{color:"+w.lightTextColor.NameWithFormat(qt.QColor__HexArgb)+"}",
		"CheckBox{color:"+w.darkTextColor.NameWithFormat(qt.QColor__HexArgb)+"}")
}

// LightTextColor returns the light text color.
func (w *CheckBox) LightTextColor() *qt.QColor { return w.lightTextColor }

// DarkTextColor returns the dark text color.
func (w *CheckBox) DarkTextColor() *qt.QColor { return w.darkTextColor }

func (w *CheckBox) state() CheckBoxState {
	if !w.IsEnabled() {
		if w.IsChecked() {
			return CheckBoxStateCheckedDisabled
		}
		return CheckBoxStateDisabled
	}
	if w.IsChecked() {
		if w.isPressed {
			return CheckBoxStateCheckedPressed
		}
		if w.isHover {
			return CheckBoxStateCheckedHover
		}
		return CheckBoxStateChecked
	}
	if w.isPressed {
		return CheckBoxStatePressed
	}
	if w.isHover {
		return CheckBoxStateHover
	}
	return CheckBoxStateNormal
}

// checkedColor returns the theme-appropriate checked indicator color.
func (w *CheckBox) checkedColor() *qt.QColor {
	if common.IsDarkTheme() {
		return w.darkCheckedColor
	}
	return w.lightCheckedColor
}

func (w *CheckBox) borderColor() *qt.QColor {
	dark := common.IsDarkTheme()
	if dark {
		switch w.state() {
		case CheckBoxStateNormal, CheckBoxStateHover:
			return qt.NewQColor11(255, 255, 255, 141)
		case CheckBoxStatePressed:
			return qt.NewQColor11(255, 255, 255, 40)
		case CheckBoxStateChecked:
			return common.FallbackThemeColor(w.checkedColor())
		case CheckBoxStateCheckedHover:
			return common.ValidColor(w.checkedColor(), common.ThemeColorDark1.Color())
		case CheckBoxStateCheckedPressed:
			return common.ValidColor(w.checkedColor(), common.ThemeColorDark2.Color())
		case CheckBoxStateDisabled:
			return qt.NewQColor11(255, 255, 255, 41)
		default:
			return qt.NewQColor11(0, 0, 0, 0)
		}
	}
	switch w.state() {
	case CheckBoxStateNormal:
		return qt.NewQColor11(0, 0, 0, 122)
	case CheckBoxStateHover:
		return qt.NewQColor11(0, 0, 0, 143)
	case CheckBoxStatePressed:
		return qt.NewQColor11(0, 0, 0, 69)
	case CheckBoxStateChecked:
		return common.FallbackThemeColor(w.checkedColor())
	case CheckBoxStateCheckedHover:
		return common.ValidColor(w.checkedColor(), common.ThemeColorLight1.Color())
	case CheckBoxStateCheckedPressed:
		return common.ValidColor(w.checkedColor(), common.ThemeColorLight2.Color())
	case CheckBoxStateDisabled:
		return qt.NewQColor11(0, 0, 0, 56)
	default:
		return qt.NewQColor11(0, 0, 0, 0)
	}
}

func (w *CheckBox) backgroundColor() *qt.QColor {
	dark := common.IsDarkTheme()
	if dark {
		switch w.state() {
		case CheckBoxStateNormal:
			return qt.NewQColor11(0, 0, 0, 26)
		case CheckBoxStateHover:
			return qt.NewQColor11(255, 255, 255, 11)
		case CheckBoxStatePressed:
			return qt.NewQColor11(255, 255, 255, 18)
		case CheckBoxStateChecked:
			return common.FallbackThemeColor(w.checkedColor())
		case CheckBoxStateCheckedHover:
			return common.ValidColor(w.checkedColor(), common.ThemeColorDark1.Color())
		case CheckBoxStateCheckedPressed:
			return common.ValidColor(w.checkedColor(), common.ThemeColorDark2.Color())
		case CheckBoxStateDisabled:
			return qt.NewQColor11(0, 0, 0, 0)
		default:
			return qt.NewQColor11(255, 255, 255, 41)
		}
	}
	switch w.state() {
	case CheckBoxStateNormal:
		return qt.NewQColor11(0, 0, 0, 6)
	case CheckBoxStateHover:
		return qt.NewQColor11(0, 0, 0, 13)
	case CheckBoxStatePressed:
		return qt.NewQColor11(0, 0, 0, 31)
	case CheckBoxStateChecked:
		return common.FallbackThemeColor(w.checkedColor())
	case CheckBoxStateCheckedHover:
		return common.ValidColor(w.checkedColor(), common.ThemeColorLight1.Color())
	case CheckBoxStateCheckedPressed:
		return common.ValidColor(w.checkedColor(), common.ThemeColorLight2.Color())
	case CheckBoxStateDisabled:
		return qt.NewQColor11(0, 0, 0, 0)
	default:
		return qt.NewQColor11(0, 0, 0, 56)
	}
}

// cloneColor returns a fresh copy of color, or a null color when nil. The copy
// constructor preserves the alpha channel (Name() returns "#RRGGBB" and would
// silently drop it).
func cloneColor(c *qt.QColor) *qt.QColor {
	if c == nil {
		return qt.NewQColor()
	}
	return qt.NewQColor9(c)
}
