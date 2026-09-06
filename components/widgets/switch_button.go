package widgets

import (
	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// SwitchIndicator is the toggle indicator of a SwitchButton. The slide
// animation on the custom "sliderX" property is driven by a frame-based
// ProgressAnimation (the custom meta-object property is not expressible in Go).
type SwitchIndicator struct {
	*ToolButton
	lightCheckedColor *qt.QColor
	darkCheckedColor  *qt.QColor
	sliderX           float64
	slideAni          *common.ProgressAnimation
	onCheckedChanged  func(bool)
}

// NewSwitchIndicator builds a switch indicator.
func NewSwitchIndicator(parent *qt.QWidget) *SwitchIndicator {
	w := &SwitchIndicator{ToolButton: NewToolButton(parent)}
	w.SetCheckable(true)
	w.SetFixedSize2(42, 22)
	w.lightCheckedColor = qt.NewQColor()
	w.darkCheckedColor = qt.NewQColor()
	w.sliderX = 5

	w.OnToggled(func(checked bool) {
		w.toggleSlider(checked)
	})
	w.OnMouseReleaseEvent(func(super func(e *qt.QMouseEvent), e *qt.QMouseEvent) {
		w.isPressed = false
		super(e)
		if w.onCheckedChanged != nil {
			w.onCheckedChanged(w.IsChecked())
		}
	})
	w.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		w.paint()
	})
	return w
}

// OnCheckedChanged registers the checked-changed callback.
func (w *SwitchIndicator) OnCheckedChanged(f func(bool)) { w.onCheckedChanged = f }

// Toggle flips the checked state.
func (w *SwitchIndicator) Toggle() { w.SetChecked(!w.IsChecked()) }

// SetDown sets the pressed state.
func (w *SwitchIndicator) SetDown(isDown bool) {
	w.isPressed = isDown
	w.ToolButton.SetDown(isDown)
}

// SetHover sets the hover state.
func (w *SwitchIndicator) SetHover(isHover bool) {
	w.isHover = isHover
	w.Update()
}

// SetCheckedColor sets the checked indicator color for light/dark themes.
func (w *SwitchIndicator) SetCheckedColor(light, dark *qt.QColor) {
	w.lightCheckedColor = cloneColor(light)
	w.darkCheckedColor = cloneColor(dark)
	w.Update()
}

func (w *SwitchIndicator) toggleSlider(checked bool) {
	target := 5.0
	if checked {
		target = 25
	}
	w.stopSliderAni()
	from := w.sliderX
	// Python animates the custom "sliderX" property over 120ms with the default
	// (linear) easing curve. Drive the same lerp manually each frame.
	w.slideAni = common.AnimateFloat(from, target, 120, nil, func(v float64) {
		w.sliderX = v
		w.Update()
	})
}

// stopSliderAni interrupts and releases any in-flight slider animation.
func (w *SwitchIndicator) stopSliderAni() {
	if w.slideAni != nil {
		w.slideAni.Stop()
		w.slideAni.Delete()
		w.slideAni = nil
	}
}

func (w *SwitchIndicator) paint() {
	painter := qt.NewQPainter2(w.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__Antialiasing)
	w.drawBackground(painter)
	w.drawCircle(painter)
	painter.End()
}

func (w *SwitchIndicator) drawBackground(painter *qt.QPainter) {
	r := float64(w.Height()) / 2
	border := w.borderColor() // borrowed (may return the stored field color) — do NOT Delete
	painter.SetPen(border)
	bg := w.backgroundColor() // borrowed (may return the stored field color) — do NOT Delete
	brush := qt.NewQBrush3(bg)
	painter.SetBrush(brush)
	brush.Delete()
	adjusted := w.Rect().Adjusted(1, 1, -1, -1) // GoGC-armed — do NOT Delete
	painter.DrawRoundedRect3(adjusted, r, r)
}

func (w *SwitchIndicator) drawCircle(painter *qt.QPainter) {
	painter.SetPenWithStyle(qt.NoPen)
	sc := w.sliderColor()
	defer sc.Delete()
	brush := qt.NewQBrush3(sc)
	painter.SetBrush(brush)
	brush.Delete()
	painter.DrawEllipse2(int(w.sliderX), 5, 12, 12)
}

func (w *SwitchIndicator) backgroundColor() *qt.QColor {
	isDark := common.IsDarkTheme()

	if w.IsChecked() {
		var color *qt.QColor
		if isDark {
			color = w.darkCheckedColor
		} else {
			color = w.lightCheckedColor
		}
		if !w.IsEnabled() {
			if isDark {
				return qt.NewQColor11(255, 255, 255, 41)
			}
			return qt.NewQColor11(0, 0, 0, 56)
		}
		if w.isPressed {
			return common.ValidColor(color, common.ThemeColorLight2.Color())
		}
		if w.isHover {
			return common.ValidColor(color, common.ThemeColorLight1.Color())
		}
		return common.FallbackThemeColor(color)
	}

	if !w.IsEnabled() {
		return qt.NewQColor11(0, 0, 0, 0)
	}
	if w.isPressed {
		if isDark {
			return qt.NewQColor11(255, 255, 255, 18)
		}
		return qt.NewQColor11(0, 0, 0, 23)
	}
	if w.isHover {
		if isDark {
			return qt.NewQColor11(255, 255, 255, 10)
		}
		return qt.NewQColor11(0, 0, 0, 15)
	}
	return qt.NewQColor11(0, 0, 0, 0)
}

func (w *SwitchIndicator) borderColor() *qt.QColor {
	isDark := common.IsDarkTheme()

	if w.IsChecked() {
		if w.IsEnabled() {
			return w.backgroundColor()
		}
		return qt.NewQColor11(0, 0, 0, 0)
	}

	if w.IsEnabled() {
		if isDark {
			return qt.NewQColor11(255, 255, 255, 153)
		}
		return qt.NewQColor11(0, 0, 0, 133)
	}
	if isDark {
		return qt.NewQColor11(255, 255, 255, 41)
	}
	return qt.NewQColor11(0, 0, 0, 56)
}

func (w *SwitchIndicator) sliderColor() *qt.QColor {
	isDark := common.IsDarkTheme()

	if w.IsChecked() {
		if w.IsEnabled() {
			if isDark {
				return qt.NewQColor3(0, 0, 0)
			}
			return qt.NewQColor3(255, 255, 255)
		}
		if isDark {
			return qt.NewQColor11(255, 255, 255, 77)
		}
		return qt.NewQColor3(255, 255, 255)
	}

	if w.IsEnabled() {
		if isDark {
			return qt.NewQColor11(255, 255, 255, 201)
		}
		return qt.NewQColor11(0, 0, 0, 156)
	}
	if isDark {
		return qt.NewQColor11(255, 255, 255, 96)
	}
	return qt.NewQColor11(0, 0, 0, 91)
}

// IndicatorPosition is the position of the switch indicator relative to text.
type IndicatorPosition int

const (
	// LEFT places the indicator on the left of the text.
	LEFT IndicatorPosition = iota
	// RIGHT places the indicator on the right of the text.
	RIGHT
)

// SwitchButton is a fluent-styled on/off switch with an optional text label.
//
// Constructors
//   - NewSwitchButton(parent *qt.QWidget, indicatorPos IndicatorPosition)
//   - NewSwitchButtonText(text string, parent *qt.QWidget, indicatorPos IndicatorPosition)
type SwitchButton struct {
	*qt.QWidget
	text             string
	offText          string
	onText           string
	spacing          int
	lightTextColor   *qt.QColor
	darkTextColor    *qt.QColor
	indicatorPos     IndicatorPosition
	hBox             *qt.QHBoxLayout
	indicator        *SwitchIndicator
	label            *qt.QLabel
	onCheckedChanged func(bool)
}

// NewSwitchButton builds a switch button.
func NewSwitchButton(parent *qt.QWidget, indicatorPos IndicatorPosition) *SwitchButton {
	w := &SwitchButton{QWidget: qt.NewQWidget(parent)}
	w.SetObjectName("switchButton")
	w.text = "Off"
	w.offText = "Off"
	w.onText = "On"
	w.spacing = 12
	w.lightTextColor = qt.NewQColor3(0, 0, 0)
	w.darkTextColor = qt.NewQColor3(255, 255, 255)
	w.indicatorPos = indicatorPos
	w.hBox = qt.NewQHBoxLayout(w.QWidget)
	w.indicator = NewSwitchIndicator(w.QWidget)
	w.label = qt.NewQLabel5(w.text, w.QWidget)
	w.initWidget()
	return w
}

// NewSwitchButtonText builds a switch button with custom off text.
func NewSwitchButtonText(text string, parent *qt.QWidget, indicatorPos IndicatorPosition) *SwitchButton {
	w := NewSwitchButton(parent, indicatorPos)
	w.offText = text
	w.SetText(text)
	return w
}

// OnCheckedChanged registers the checked-changed callback.
func (w *SwitchButton) OnCheckedChanged(f func(bool)) { w.onCheckedChanged = f }

func (w *SwitchButton) initWidget() {
	w.SetAttribute(qt.WA_StyledBackground)
	w.InstallEventFilter(w.QObject)
	w.SetFixedHeight(22)

	w.hBox.SetSpacing(w.spacing)
	w.hBox.SetContentsMargins(2, 0, 0, 0)

	if w.indicatorPos == LEFT {
		w.hBox.AddWidget(w.indicator.QWidget)
		w.hBox.AddWidget(w.label.QWidget)
	} else {
		w.hBox.AddWidget3(w.label.QWidget, 0, qt.AlignRight)
		w.hBox.AddWidget3(w.indicator.QWidget, 0, qt.AlignRight)
	}

	common.FluentStyleSheet(common.FluentSwitchButton).Apply(w.QWidget, common.ThemeAuto)
	common.FluentStyleSheet(common.FluentSwitchButton).Apply(w.label.QWidget, common.ThemeAuto)

	w.indicator.OnToggled(func(checked bool) {
		w.updateText()
		if w.onCheckedChanged != nil {
			w.onCheckedChanged(checked)
		}
	})
	w.OnEventFilter(func(super func(watched *qt.QObject, event *qt.QEvent) bool, watched *qt.QObject, event *qt.QEvent) bool {
		if watched.UnsafePointer() == w.UnsafePointer() && w.IsEnabled() {
			switch event.Type() {
			case qt.QEvent__MouseButtonPress:
				w.indicator.SetDown(true)
			case qt.QEvent__MouseButtonRelease:
				w.indicator.SetDown(false)
				w.indicator.Toggle()
			case qt.QEvent__Enter:
				w.indicator.SetHover(true)
			case qt.QEvent__Leave:
				w.indicator.SetHover(false)
			}
		}
		return super(watched, event)
	})
}

// IsChecked reports the checked state.
func (w *SwitchButton) IsChecked() bool { return w.indicator.IsChecked() }

// SetChecked sets the checked state.
func (w *SwitchButton) SetChecked(isChecked bool) {
	w.updateText()
	w.indicator.SetChecked(isChecked)
}

// ToggleChecked flips the checked state.
func (w *SwitchButton) ToggleChecked() {
	w.indicator.SetChecked(!w.indicator.IsChecked())
}

// Text returns the current text.
func (w *SwitchButton) Text() string { return w.text }

// SetText sets the text shown next to the indicator.
func (w *SwitchButton) SetText(text string) {
	w.text = text
	w.label.SetText(text)
	w.AdjustSize()
}

// OnText returns the on-state text.
func (w *SwitchButton) OnText() string { return w.onText }

// SetOnText sets the on-state text.
func (w *SwitchButton) SetOnText(text string) {
	w.onText = text
	w.updateText()
}

// OffText returns the off-state text.
func (w *SwitchButton) OffText() string { return w.offText }

// SetOffText sets the off-state text.
func (w *SwitchButton) SetOffText(text string) {
	w.offText = text
	w.updateText()
}

// Spacing returns the spacing between indicator and text.
func (w *SwitchButton) Spacing() int { return w.spacing }

// SetSpacing sets the spacing between indicator and text.
func (w *SwitchButton) SetSpacing(spacing int) {
	w.spacing = spacing
	w.hBox.SetSpacing(spacing)
	w.Update()
}

// SetTextColor sets the light/dark text colors.
func (w *SwitchButton) SetTextColor(light, dark *qt.QColor) {
	w.lightTextColor = cloneColor(light)
	w.darkTextColor = cloneColor(dark)
	common.SetCustomStyleSheet(w.label.QWidget,
		"SwitchButton>QLabel{color:"+w.lightTextColor.NameWithFormat(qt.QColor__HexArgb)+"}",
		"SwitchButton>QLabel{color:"+w.darkTextColor.NameWithFormat(qt.QColor__HexArgb)+"}")
}

// SetCheckedIndicatorColor sets the checked indicator color for light/dark themes.
func (w *SwitchButton) SetCheckedIndicatorColor(light, dark *qt.QColor) {
	w.indicator.SetCheckedColor(light, dark)
}

func (w *SwitchButton) updateText() {
	if w.IsChecked() {
		w.SetText(w.onText)
	} else {
		w.SetText(w.offText)
	}
}
