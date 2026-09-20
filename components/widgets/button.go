package widgets

import (
	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// ---------------------------------------------------------------------------
// PushButton
//
// A fluent-styled push button. The icon is painted manually (the underlying
// QAbstractButton icon stays null), so the QSS `hasIcon` property controls the
// text padding exactly like the Python port.
//
// Constructors
//   - NewPushButton(parent *qt.QWidget)
//   - NewPushButtonText(text string, parent *qt.QWidget)
//   - NewPushButtonIcon(icon interface{}, text string, parent *qt.QWidget)
type PushButton struct {
	*qt.QPushButton
	isPressed      bool
	isHover        bool
	hasIcon        bool
	iconSrc        interface{}
	drawIconFunc   func(icon interface{}, painter *qt.QPainter, rect *qt.QRectF)
	paintExtraFunc func(painter *qt.QPainter)
	onPressedFunc  func()
	onReleasedFunc func()
}

// newPushButtonBase performs the shared construction for every push-button
// variant.
func newPushButtonBase(parent *qt.QWidget) *PushButton {
	w := &PushButton{QPushButton: qt.NewQPushButton(parent)}
	common.FluentStyleSheet(common.FluentButton).Apply(w.QWidget, common.ThemeAuto)
	w.SetIconSize(qt.NewQSize2(16, 16))
	w.hasIcon = false
	w.iconSrc = qt.NewQIcon()
	// Mirror Python's setIcon(None): publish the hasIcon=false dynamic property
	// so the QSS `PushButton[hasIcon=false]` padding rule matches.
	w.SetProperty("hasIcon", qt.NewQVariant11(false))
	common.SetFont(w.QWidget, 14, 400)
	w.drawIconFunc = w.drawIconBase
	w.installEvents()
	return w
}

// NewPushButton builds an empty push button.
func NewPushButton(parent *qt.QWidget) *PushButton {
	return newPushButtonBase(parent)
}

// NewPushButtonText builds a push button with text.
func NewPushButtonText(text string, parent *qt.QWidget) *PushButton {
	w := NewPushButton(parent)
	w.SetText(text)
	return w
}

// NewPushButtonIcon builds a push button with icon and text.
func NewPushButtonIcon(icon interface{}, text string, parent *qt.QWidget) *PushButton {
	w := NewPushButtonText(text, parent)
	w.SetIcon(icon)
	return w
}

// SetIcon sets the icon source (nil, string, *qt.QIcon or FluentIconBase).
func (w *PushButton) SetIcon(icon interface{}) {
	if icon == nil {
		w.hasIcon = false
	} else if ic, ok := icon.(*qt.QIcon); ok && ic.IsNull() {
		w.hasIcon = false
	} else {
		w.hasIcon = true
	}
	w.SetProperty("hasIcon", qt.NewQVariant11(w.hasIcon))
	w.SetStyle(qt.QApplication_Style())
	if icon == nil {
		w.iconSrc = qt.NewQIcon()
	} else {
		w.iconSrc = icon
	}
	w.Update()
}

// Icon returns the current icon as a *qt.QIcon.
func (w *PushButton) Icon() *qt.QIcon {
	return common.ToQIcon(w.iconSrc)
}

// GetIcon is an alias of Icon for API parity with the fluent icon widget.
func (w *PushButton) GetIcon() *qt.QIcon { return w.Icon() }

func (w *PushButton) installEvents() {
	w.OnMousePressEvent(func(super func(e *qt.QMouseEvent), e *qt.QMouseEvent) {
		w.isPressed = true
		super(e)
		if w.onPressedFunc != nil {
			w.onPressedFunc()
		}
	})
	w.OnMouseReleaseEvent(func(super func(e *qt.QMouseEvent), e *qt.QMouseEvent) {
		w.isPressed = false
		super(e)
		if w.onReleasedFunc != nil {
			w.onReleasedFunc()
		}
	})
	w.OnEnterEvent(func(super func(e *qt.QEvent), e *qt.QEvent) {
		w.isHover = true
		w.Update()
	})
	w.OnLeaveEvent(func(super func(e *qt.QEvent), e *qt.QEvent) {
		w.isHover = false
		w.Update()
	})
	w.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		super(e)
		w.paintIcon()
		w.paintExtra()
	})
}

func (w *PushButton) paintIcon() {
	if !w.hasIcon {
		return
	}
	painter := qt.NewQPainter2(w.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__Antialiasing | qt.QPainter__SmoothPixmapTransform)

	if !w.IsEnabled() {
		painter.SetOpacity(0.3628)
	} else if w.isPressed {
		painter.SetOpacity(0.786)
	}

	sz := w.IconSize() // GoGC-armed — do NOT Delete
	iw, ih := sz.Width(), sz.Height()
	y := (w.Height() - ih) / 2
	mh := w.MinimumSizeHint() // GoGC-armed — do NOT Delete
	mw := mh.Width()
	x := 12
	if mw > 0 {
		x = 12 + (w.Width()-mw)/2
	}
	if w.IsRightToLeft() {
		x = w.Width() - iw - x
	}
	rect := qt.NewQRectF4(float64(x), float64(y), float64(iw), float64(ih))
	defer rect.Delete()
	w.drawIconFunc(w.iconSrc, painter, rect)
	painter.End()
}

func (w *PushButton) paintExtra() {
	if w.paintExtraFunc == nil {
		return
	}
	painter := qt.NewQPainter2(w.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__Antialiasing)
	w.paintExtraFunc(painter)
	painter.End()
}

func (w *PushButton) drawIconBase(icon interface{}, painter *qt.QPainter, rect *qt.QRectF) {
	renderFluentIcon(icon, painter, rect, common.ThemeAuto)
}

func (w *PushButton) drawIconPrimary(icon interface{}, painter *qt.QPainter, rect *qt.QRectF) {
	if fi, ok := icon.(common.FluentIconBase); ok {
		if w.IsEnabled() {
			fi.Render(painter, rect, reversedTheme())
		} else {
			if common.IsDarkTheme() {
				painter.SetOpacity(0.786)
			} else {
				painter.SetOpacity(0.9)
			}
			fi.Render(painter, rect, common.ThemeDark)
		}
		return
	}
	renderFluentIcon(icon, painter, rect, common.ThemeAuto)
}

// ---------------------------------------------------------------------------
// PrimaryPushButton

// PrimaryPushButton is a primary-color push button.
//
// Constructors
//   - NewPrimaryPushButton(parent *qt.QWidget)
//   - NewPrimaryPushButtonText(text string, parent *qt.QWidget)
//   - NewPrimaryPushButtonIcon(icon interface{}, text string, parent *qt.QWidget)
type PrimaryPushButton struct{ *PushButton }

// NewPrimaryPushButton builds an empty primary push button.
func NewPrimaryPushButton(parent *qt.QWidget) *PrimaryPushButton {
	w := &PrimaryPushButton{PushButton: newPushButtonBase(parent)}
	w.SetObjectName("primaryPushButton")
	w.drawIconFunc = w.drawIconPrimary
	return w
}

// NewPrimaryPushButtonText builds a primary push button with text.
func NewPrimaryPushButtonText(text string, parent *qt.QWidget) *PrimaryPushButton {
	w := NewPrimaryPushButton(parent)
	w.SetText(text)
	return w
}

// NewPrimaryPushButtonIcon builds a primary push button with icon and text.
func NewPrimaryPushButtonIcon(icon interface{}, text string, parent *qt.QWidget) *PrimaryPushButton {
	w := NewPrimaryPushButtonText(text, parent)
	w.SetIcon(icon)
	return w
}

// ---------------------------------------------------------------------------
// TransparentPushButton

// TransparentPushButton is a transparent-background push button.
//
// Constructors
//   - NewTransparentPushButton(parent *qt.QWidget)
//   - NewTransparentPushButtonText(text string, parent *qt.QWidget)
//   - NewTransparentPushButtonIcon(icon interface{}, text string, parent *qt.QWidget)
type TransparentPushButton struct{ *PushButton }

// NewTransparentPushButton builds an empty transparent push button.
func NewTransparentPushButton(parent *qt.QWidget) *TransparentPushButton {
	w := &TransparentPushButton{PushButton: newPushButtonBase(parent)}
	w.SetObjectName("transparentPushButton")
	return w
}

// NewTransparentPushButtonText builds a transparent push button with text.
func NewTransparentPushButtonText(text string, parent *qt.QWidget) *TransparentPushButton {
	w := NewTransparentPushButton(parent)
	w.SetText(text)
	return w
}

// NewTransparentPushButtonIcon builds a transparent push button with icon and text.
func NewTransparentPushButtonIcon(icon interface{}, text string, parent *qt.QWidget) *TransparentPushButton {
	w := NewTransparentPushButtonText(text, parent)
	w.SetIcon(icon)
	return w
}

// ---------------------------------------------------------------------------
// ToggleButton

// ToggleButton is a checkable toggle push button.
//
// Constructors
//   - NewToggleButton(parent *qt.QWidget)
//   - NewToggleButtonText(text string, parent *qt.QWidget)
//   - NewToggleButtonIcon(icon interface{}, text string, parent *qt.QWidget)
type ToggleButton struct{ *PushButton }

// NewToggleButton builds an empty toggle button.
func NewToggleButton(parent *qt.QWidget) *ToggleButton {
	w := &ToggleButton{PushButton: newPushButtonBase(parent)}
	w.SetObjectName("toggleButton")
	w.SetCheckable(true)
	w.SetChecked(false)
	w.drawIconFunc = w.drawIconToggle
	return w
}

// NewToggleButtonText builds a toggle button with text.
func NewToggleButtonText(text string, parent *qt.QWidget) *ToggleButton {
	w := NewToggleButton(parent)
	w.SetText(text)
	return w
}

// NewToggleButtonIcon builds a toggle button with icon and text.
func NewToggleButtonIcon(icon interface{}, text string, parent *qt.QWidget) *ToggleButton {
	w := NewToggleButtonText(text, parent)
	w.SetIcon(icon)
	return w
}

func (w *ToggleButton) drawIconToggle(icon interface{}, painter *qt.QPainter, rect *qt.QRectF) {
	if !w.IsChecked() {
		renderFluentIcon(icon, painter, rect, common.ThemeAuto)
		return
	}
	w.PushButton.drawIconPrimary(icon, painter, rect)
}

// ---------------------------------------------------------------------------
// TransparentTogglePushButton

// TransparentTogglePushButton is a transparent toggle push button.
//
// Constructors
//   - NewTransparentTogglePushButton(parent *qt.QWidget)
//   - NewTransparentTogglePushButtonText(text string, parent *qt.QWidget)
//   - NewTransparentTogglePushButtonIcon(icon interface{}, text string, parent *qt.QWidget)
type TransparentTogglePushButton struct{ *ToggleButton }

// NewTransparentTogglePushButton builds an empty transparent toggle button.
func NewTransparentTogglePushButton(parent *qt.QWidget) *TransparentTogglePushButton {
	w := &TransparentTogglePushButton{ToggleButton: NewToggleButton(parent)}
	w.SetObjectName("transparentTogglePushButton")
	return w
}

// NewTransparentTogglePushButtonText builds a transparent toggle button with text.
func NewTransparentTogglePushButtonText(text string, parent *qt.QWidget) *TransparentTogglePushButton {
	w := NewTransparentTogglePushButton(parent)
	w.SetText(text)
	return w
}

// NewTransparentTogglePushButtonIcon builds a transparent toggle button with icon and text.
func NewTransparentTogglePushButtonIcon(icon interface{}, text string, parent *qt.QWidget) *TransparentTogglePushButton {
	w := NewTransparentTogglePushButtonText(text, parent)
	w.SetIcon(icon)
	return w
}

// ---------------------------------------------------------------------------
// HyperlinkButton

// HyperlinkButton is a push button that opens a URL when clicked.
//
// Constructors
//   - NewHyperlinkButton(parent *qt.QWidget)
//   - NewHyperlinkButtonURL(url, text string, parent *qt.QWidget)
//   - NewHyperlinkButtonIcon(icon interface{}, url, text string, parent *qt.QWidget)
type HyperlinkButton struct {
	*PushButton
	url *qt.QUrl
}

// NewHyperlinkButton builds a hyperlink button.
func NewHyperlinkButton(parent *qt.QWidget) *HyperlinkButton {
	w := &HyperlinkButton{PushButton: newPushButtonBase(parent), url: qt.NewQUrl()}
	w.SetObjectName("hyperlinkButton")
	w.SetCursor(qt.NewQCursor2(qt.PointingHandCursor))
	w.drawIconFunc = w.drawIconHyperlink
	w.OnClicked(func() {
		if w.url.IsValid() {
			qt.QDesktopServices_OpenUrl(w.url)
		}
	})
	return w
}

// NewHyperlinkButtonURL builds a hyperlink button with url and text.
func NewHyperlinkButtonURL(url, text string, parent *qt.QWidget) *HyperlinkButton {
	w := NewHyperlinkButton(parent)
	w.SetText(text)
	w.SetUrl(url)
	return w
}

// NewHyperlinkButtonIcon builds a hyperlink button with icon, url and text.
func NewHyperlinkButtonIcon(icon interface{}, url, text string, parent *qt.QWidget) *HyperlinkButton {
	w := NewHyperlinkButtonURL(url, text, parent)
	w.SetIcon(icon)
	return w
}

// Url returns the target URL.
func (w *HyperlinkButton) Url() *qt.QUrl { return w.url }

// SetUrl sets the target URL from a string.
func (w *HyperlinkButton) SetUrl(url string) { w.url = qt.NewQUrl3(url) }

func (w *HyperlinkButton) drawIconHyperlink(icon interface{}, painter *qt.QPainter, rect *qt.QRectF) {
	if !w.IsEnabled() {
		if common.IsDarkTheme() {
			painter.SetOpacity(0.3628)
		} else {
			painter.SetOpacity(0.36)
		}
	}
	renderFluentIcon(icon, painter, rect, common.ThemeAuto)
}

// ---------------------------------------------------------------------------
// RadioButton

// RadioButton is a fluent-styled radio button with a custom-painted indicator.
//
// Constructors
//   - NewRadioButton(parent *qt.QWidget)
//   - NewRadioButtonText(text string, parent *qt.QWidget)
type RadioButton struct {
	*qt.QRadioButton
	lightTextColor      *qt.QColor
	darkTextColor       *qt.QColor
	lightIndicatorColor *qt.QColor
	darkIndicatorColor  *qt.QColor
	indicatorX          int
	indicatorY          int
	isHover             bool
}

// NewRadioButton builds a radio button.
func NewRadioButton(parent *qt.QWidget) *RadioButton {
	w := &RadioButton{QRadioButton: qt.NewQRadioButton(parent)}
	w.lightTextColor = qt.NewQColor3(0, 0, 0)
	w.darkTextColor = qt.NewQColor3(255, 255, 255)
	w.lightIndicatorColor = qt.NewQColor()
	w.darkIndicatorColor = qt.NewQColor()
	w.indicatorX = 11
	w.indicatorY = 12
	// Mirror CheckBox: set the 14px Fluent font explicitly instead of relying
	// solely on the `font: 14px` rule in button.qss (which only reaches
	// QWidget::font() after the widget is polished). The custom _drawText uses
	// w.Font(), so this guarantees correct text size on the first paint.
	common.SetFont(w.QWidget, 14, int(qt.QFont__Normal))
	w.SetObjectName("radioButton")
	common.FluentStyleSheet(common.FluentButton).Apply(w.QWidget, common.ThemeAuto)
	w.SetAttribute2(qt.WA_MacShowFocusRect, false)
	w.installEvents()
	return w
}

// NewRadioButtonText builds a radio button with text.
func NewRadioButtonText(text string, parent *qt.QWidget) *RadioButton {
	w := NewRadioButton(parent)
	w.SetText(text)
	return w
}

// SetTextColor sets the light/dark text colors.
func (w *RadioButton) SetTextColor(light, dark *qt.QColor) {
	w.SetLightTextColor(light)
	w.SetDarkTextColor(dark)
}

// SetLightTextColor sets the light text color.
func (w *RadioButton) SetLightTextColor(color *qt.QColor) {
	w.lightTextColor = cloneColor(color)
	w.Update()
}

// SetDarkTextColor sets the dark text color.
func (w *RadioButton) SetDarkTextColor(color *qt.QColor) {
	w.darkTextColor = cloneColor(color)
	w.Update()
}

// LightTextColor returns the light text color.
func (w *RadioButton) LightTextColor() *qt.QColor { return w.lightTextColor }

// DarkTextColor returns the dark text color.
func (w *RadioButton) DarkTextColor() *qt.QColor { return w.darkTextColor }

// SetIndicatorColor sets the checked indicator color for light/dark themes.
func (w *RadioButton) SetIndicatorColor(light, dark *qt.QColor) {
	w.lightIndicatorColor = cloneColor(light)
	w.darkIndicatorColor = cloneColor(dark)
	w.Update()
}

func (w *RadioButton) installEvents() {
	w.OnEnterEvent(func(super func(e *qt.QEvent), e *qt.QEvent) {
		w.isHover = true
		w.Update()
	})
	w.OnLeaveEvent(func(super func(e *qt.QEvent), e *qt.QEvent) {
		w.isHover = false
		w.Update()
	})
	w.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing | qt.QPainter__TextAntialiasing)
		w.drawIndicator(painter)
		w.drawText(painter)
		painter.End()
	})
}

func (w *RadioButton) textColor() *qt.QColor {
	if common.IsDarkTheme() {
		return w.darkTextColor
	}
	return w.lightTextColor
}

func (w *RadioButton) drawText(painter *qt.QPainter) {
	if !w.IsEnabled() {
		painter.SetOpacity(0.36)
	}
	font := w.Font() // borrowed reference — do NOT Delete
	painter.SetFont(font)
	painter.SetPen(w.textColor())
	rect := qt.NewQRect4(29, 0, w.Width(), w.Height())
	defer rect.Delete()
	painter.DrawText6(rect, int(qt.AlignVCenter), w.Text())
}

func (w *RadioButton) drawIndicator(painter *qt.QPainter) {
	dark := common.IsDarkTheme()
	var borderColor, filledColor *qt.QColor

	if w.IsChecked() {
		if w.IsEnabled() {
			borderColor = common.AutoFallbackThemeColor(w.lightIndicatorColor, w.darkIndicatorColor)
		} else if dark {
			borderColor = qt.NewQColor11(255, 255, 255, 40)
		} else {
			borderColor = qt.NewQColor11(0, 0, 0, 55)
		}
		if dark {
			filledColor = qt.NewQColor3(0, 0, 0)
		} else {
			filledColor = qt.NewQColor3(255, 255, 255)
		}
		thickness := 5
		if w.isHover && !w.IsDown() {
			thickness = 4
		}
		w.drawCircle(painter, w.indicatorX, w.indicatorY, 10, thickness, borderColor, filledColor)
		return
	}

	// unchecked
	if w.IsEnabled() {
		if !w.IsDown() {
			if dark {
				borderColor = qt.NewQColor11(255, 255, 255, 153)
			} else {
				borderColor = qt.NewQColor11(0, 0, 0, 153)
			}
		} else if dark {
			borderColor = qt.NewQColor11(255, 255, 255, 40)
		} else {
			borderColor = qt.NewQColor11(0, 0, 0, 55)
		}
		if w.IsDown() {
			if dark {
				filledColor = qt.NewQColor3(0, 0, 0)
			} else {
				filledColor = qt.NewQColor3(255, 255, 255)
			}
		} else if w.isHover {
			if dark {
				filledColor = qt.NewQColor11(255, 255, 255, 11)
			} else {
				filledColor = qt.NewQColor11(0, 0, 0, 15)
			}
		} else if dark {
			filledColor = qt.NewQColor11(0, 0, 0, 26)
		} else {
			filledColor = qt.NewQColor11(0, 0, 0, 6)
		}
	} else {
		filledColor = qt.NewQColor11(0, 0, 0, 0)
		if dark {
			borderColor = qt.NewQColor11(255, 255, 255, 40)
		} else {
			borderColor = qt.NewQColor11(0, 0, 0, 55)
		}
	}

	w.drawCircle(painter, w.indicatorX, w.indicatorY, 10, 1, borderColor, filledColor)

	if w.IsEnabled() && w.IsDown() {
		if dark {
			borderColor = qt.NewQColor11(255, 255, 255, 40)
		} else {
			borderColor = qt.NewQColor11(0, 0, 0, 24)
		}
		transparent := qt.NewQColor11(0, 0, 0, 0)
		w.drawCircle(painter, w.indicatorX, w.indicatorY, 9, 4, borderColor, transparent)
	}
}

func (w *RadioButton) drawCircle(painter *qt.QPainter, cx, cy, radius, thickness int, borderColor, filledColor *qt.QColor) {
	path := qt.NewQPainterPath()
	defer path.Delete()
	path.SetFillRule(qt.WindingFill)

	outer := qt.NewQRectF4(float64(cx-radius), float64(cy-radius), float64(2*radius), float64(2*radius))
	defer outer.Delete()
	path.AddEllipse(outer)

	ir := radius - thickness
	inner := qt.NewQRectF4(float64(cx-ir), float64(cy-ir), float64(2*ir), float64(2*ir))
	defer inner.Delete()
	innerPath := qt.NewQPainterPath()
	defer innerPath.Delete()
	innerPath.AddEllipse(inner)

	ring := path.Subtracted(innerPath) // GoGC-armed — do NOT Delete

	painter.SetPenWithStyle(qt.NoPen)
	ringBrush := qt.NewQBrush3(borderColor)
	painter.FillPath(ring, ringBrush)
	ringBrush.Delete()
	innerBrush := qt.NewQBrush3(filledColor)
	painter.FillPath(innerPath, innerBrush)
	innerBrush.Delete()
}

// ---------------------------------------------------------------------------
// ToolButton

// ToolButton is a fluent-styled tool button with a manually painted icon.
//
// Constructors
//   - NewToolButton(parent *qt.QWidget)
//   - NewToolButtonIcon(icon interface{}, parent *qt.QWidget)
type ToolButton struct {
	*qt.QToolButton
	isPressed      bool
	isHover        bool
	iconSrc        interface{}
	drawIconFunc   func(icon interface{}, painter *qt.QPainter, rect *qt.QRectF)
	paintExtraFunc func(painter *qt.QPainter)
	onPressedFunc  func()
	onReleasedFunc func()
}

// newToolButtonBase performs the shared construction for tool-button variants.
func newToolButtonBase(parent *qt.QWidget) *ToolButton {
	w := &ToolButton{QToolButton: qt.NewQToolButton(parent)}
	common.FluentStyleSheet(common.FluentButton).Apply(w.QWidget, common.ThemeAuto)
	w.SetIconSize(qt.NewQSize2(16, 16))
	w.iconSrc = qt.NewQIcon()
	common.SetFont(w.QWidget, 14, 400)
	w.drawIconFunc = w.drawIconBase
	w.installEvents()
	return w
}

// NewToolButton builds an empty tool button.
func NewToolButton(parent *qt.QWidget) *ToolButton {
	return newToolButtonBase(parent)
}

// NewToolButtonIcon builds a tool button with an icon.
func NewToolButtonIcon(icon interface{}, parent *qt.QWidget) *ToolButton {
	w := NewToolButton(parent)
	w.SetIcon(icon)
	return w
}

// SetIcon sets the icon source.
func (w *ToolButton) SetIcon(icon interface{}) {
	if icon == nil {
		w.iconSrc = qt.NewQIcon()
	} else {
		w.iconSrc = icon
	}
	w.Update()
}

// Icon returns the current icon as a *qt.QIcon.
func (w *ToolButton) Icon() *qt.QIcon { return common.ToQIcon(w.iconSrc) }

// SetDrawIconFunc overrides the per-item icon drawing routine (the Go analogue
// of overriding _drawIcon in a Python ToolButton subclass). It lets callers in
// other packages, e.g. SegmentedToggleToolItem, recolor the icon based on the
// item state without exposing the unexported drawIconFunc field.
func (w *ToolButton) SetDrawIconFunc(f func(icon interface{}, painter *qt.QPainter, rect *qt.QRectF)) {
	w.drawIconFunc = f
}

func (w *ToolButton) installEvents() {
	w.OnMousePressEvent(func(super func(e *qt.QMouseEvent), e *qt.QMouseEvent) {
		w.isPressed = true
		super(e)
		if w.onPressedFunc != nil {
			w.onPressedFunc()
		}
	})
	w.OnMouseReleaseEvent(func(super func(e *qt.QMouseEvent), e *qt.QMouseEvent) {
		w.isPressed = false
		super(e)
		if w.onReleasedFunc != nil {
			w.onReleasedFunc()
		}
	})
	w.OnEnterEvent(func(super func(e *qt.QEvent), e *qt.QEvent) {
		w.isHover = true
		w.Update()
	})
	w.OnLeaveEvent(func(super func(e *qt.QEvent), e *qt.QEvent) {
		w.isHover = false
		w.Update()
	})
	w.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		super(e)
		w.paintIcon()
		w.paintExtra()
	})
}

func (w *ToolButton) paintIcon() {
	if w.iconSrc == nil {
		return
	}
	painter := qt.NewQPainter2(w.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__Antialiasing | qt.QPainter__SmoothPixmapTransform)

	if !w.IsEnabled() {
		painter.SetOpacity(0.43)
	} else if w.isPressed {
		painter.SetOpacity(0.63)
	}

	sz := w.IconSize()

	iw, ih := sz.Width(), sz.Height()
	y := (w.Height() - ih) / 2
	x := (w.Width() - iw) / 2
	rect := qt.NewQRectF4(float64(x), float64(y), float64(iw), float64(ih))
	defer rect.Delete()
	w.drawIconFunc(w.iconSrc, painter, rect)
	painter.End()
}

func (w *ToolButton) paintExtra() {
	if w.paintExtraFunc == nil {
		return
	}
	painter := qt.NewQPainter2(w.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__Antialiasing)
	w.paintExtraFunc(painter)
	painter.End()
}

func (w *ToolButton) drawIconBase(icon interface{}, painter *qt.QPainter, rect *qt.QRectF) {
	renderFluentIcon(icon, painter, rect, common.ThemeAuto)
}

func (w *ToolButton) drawIconPrimary(icon interface{}, painter *qt.QPainter, rect *qt.QRectF) {
	if fi, ok := icon.(common.FluentIconBase); ok {
		if w.IsEnabled() {
			fi.Render(painter, rect, reversedTheme())
		} else {
			if common.IsDarkTheme() {
				painter.SetOpacity(0.786)
			} else {
				painter.SetOpacity(0.9)
			}
			fi.Render(painter, rect, common.ThemeDark)
		}
		return
	}
	if ic, ok := icon.(*common.Icon); ok {
		if w.IsEnabled() {
			ic.FluentIcon.Render(painter, rect, reversedTheme())
		} else {
			if common.IsDarkTheme() {
				painter.SetOpacity(0.786)
			} else {
				painter.SetOpacity(0.9)
			}
			ic.FluentIcon.Render(painter, rect, common.ThemeDark)
		}
		return
	}
	renderFluentIcon(icon, painter, rect, common.ThemeAuto)
}

// ---------------------------------------------------------------------------
// TransparentToolButton

// TransparentToolButton is a transparent-background tool button.
//
// Constructors
//   - NewTransparentToolButton(parent *qt.QWidget)
//   - NewTransparentToolButtonIcon(icon interface{}, parent *qt.QWidget)
type TransparentToolButton struct{ *ToolButton }

// NewTransparentToolButton builds an empty transparent tool button.
func NewTransparentToolButton(parent *qt.QWidget) *TransparentToolButton {
	w := &TransparentToolButton{ToolButton: newToolButtonBase(parent)}
	w.SetObjectName("transparentToolButton")
	return w
}

// NewTransparentToolButtonIcon builds a transparent tool button with an icon.
func NewTransparentToolButtonIcon(icon interface{}, parent *qt.QWidget) *TransparentToolButton {
	w := NewTransparentToolButton(parent)
	w.SetIcon(icon)
	return w
}

// ---------------------------------------------------------------------------
// PrimaryToolButton

// PrimaryToolButton is a primary-color tool button.
//
// Constructors
//   - NewPrimaryToolButton(parent *qt.QWidget)
//   - NewPrimaryToolButtonIcon(icon interface{}, parent *qt.QWidget)
type PrimaryToolButton struct{ *ToolButton }

// NewPrimaryToolButton builds an empty primary tool button.
func NewPrimaryToolButton(parent *qt.QWidget) *PrimaryToolButton {
	w := &PrimaryToolButton{ToolButton: newToolButtonBase(parent)}
	w.SetObjectName("primaryToolButton")
	w.drawIconFunc = w.drawIconPrimary
	return w
}

// NewPrimaryToolButtonIcon builds a primary tool button with an icon.
func NewPrimaryToolButtonIcon(icon interface{}, parent *qt.QWidget) *PrimaryToolButton {
	w := NewPrimaryToolButton(parent)
	w.SetIcon(icon)
	return w
}

// ---------------------------------------------------------------------------
// ToggleToolButton

// ToggleToolButton is a checkable toggle tool button.
//
// Constructors
//   - NewToggleToolButton(parent *qt.QWidget)
//   - NewToggleToolButtonIcon(icon interface{}, parent *qt.QWidget)
type ToggleToolButton struct{ *ToolButton }

// NewToggleToolButton builds an empty toggle tool button.
func NewToggleToolButton(parent *qt.QWidget) *ToggleToolButton {
	w := &ToggleToolButton{ToolButton: newToolButtonBase(parent)}
	w.SetObjectName("toggleToolButton")
	w.SetCheckable(true)
	w.SetChecked(false)
	w.drawIconFunc = w.drawIconToggle
	return w
}

// NewToggleToolButtonIcon builds a toggle tool button with an icon.
func NewToggleToolButtonIcon(icon interface{}, parent *qt.QWidget) *ToggleToolButton {
	w := NewToggleToolButton(parent)
	w.SetIcon(icon)
	return w
}

func (w *ToggleToolButton) drawIconToggle(icon interface{}, painter *qt.QPainter, rect *qt.QRectF) {
	if !w.IsChecked() {
		renderFluentIcon(icon, painter, rect, common.ThemeAuto)
		return
	}
	w.ToolButton.drawIconPrimary(icon, painter, rect)
}

// ---------------------------------------------------------------------------
// TransparentToggleToolButton

// TransparentToggleToolButton is a transparent toggle tool button.
//
// Constructors
//   - NewTransparentToggleToolButton(parent *qt.QWidget)
//   - NewTransparentToggleToolButtonIcon(icon interface{}, parent *qt.QWidget)
type TransparentToggleToolButton struct{ *ToggleToolButton }

// NewTransparentToggleToolButton builds an empty transparent toggle tool button.
func NewTransparentToggleToolButton(parent *qt.QWidget) *TransparentToggleToolButton {
	w := &TransparentToggleToolButton{ToggleToolButton: NewToggleToolButton(parent)}
	w.SetObjectName("transparentToggleToolButton")
	return w
}

// NewTransparentToggleToolButtonIcon builds a transparent toggle tool button with an icon.
func NewTransparentToggleToolButtonIcon(icon interface{}, parent *qt.QWidget) *TransparentToggleToolButton {
	w := NewTransparentToggleToolButton(parent)
	w.SetIcon(icon)
	return w
}

// ---------------------------------------------------------------------------
// Drop-down buttons

// translateYAnimation ports Python's TranslateYAnimation: it bounces the
// drop-down arrow down on press (OutQuad) and springs it back on release
// (OutElastic) by driving a custom float offset through ProgressAnimation.
type translateYAnimation struct {
	owner     *qt.QWidget
	maxOffset float64
	y         float64
	ani       *common.ProgressAnimation
}

func newTranslateYAnimation(owner *qt.QWidget, offset float64) *translateYAnimation {
	return &translateYAnimation{owner: owner, maxOffset: offset}
}

func (a *translateYAnimation) setY(y float64) {
	a.y = y
	a.owner.Update()
}

func (a *translateYAnimation) press() {
	a.run(150, qt.QEasingCurve__OutQuad, a.maxOffset)
}

func (a *translateYAnimation) release() {
	a.run(500, qt.QEasingCurve__OutElastic, 0)
}

func (a *translateYAnimation) run(duration int, easing qt.QEasingCurve__Type, to float64) {
	if a.ani != nil {
		a.ani.Stop()
		a.ani.Delete()
		a.ani = nil
	}
	from := a.y
	curve := qt.NewQEasingCurve3(easing)
	a.ani = common.NewProgressAnimation(duration, curve)
	curve.Delete()
	a.ani.OnProgress(func(t float64) {
		a.setY(from + (to-from)*t)
	})
	a.ani.Start()
}

func drawDropDownArrow(btn *qt.QWidget, isEnabled, isHover, isPressed, primary bool, arrowOffset float64, painter *qt.QPainter) {
	if !isEnabled {
		if common.IsDarkTheme() {
			painter.SetOpacity(0.43)
		} else {
			painter.SetOpacity(0.6)
		}
	} else if isHover {
		painter.SetOpacity(0.8)
	} else if isPressed {
		painter.SetOpacity(0.7)
	}

	arrowY := float64(btn.Height())/2 - 5 + arrowOffset
	rect := qt.NewQRectF4(float64(btn.Width())-22, arrowY, 10, 10)
	defer rect.Delete()

	if primary {
		theme := common.ThemeLight
		if !(isEnabled && common.IsDarkTheme()) {
			theme = common.ThemeDark
		}
		common.ChevronDown.Render(painter, rect, theme)
		return
	}
	if common.IsDarkTheme() {
		common.ChevronDown.Render(painter, rect, common.ThemeDark)
	} else {
		renderFluentIconWithFill(common.ChevronDown, painter, rect, "#646464")
	}
}

// DropDownPushButton is a push button with a drop-down arrow and menu.
//
// Constructors
//   - NewDropDownPushButton(parent *qt.QWidget)
//   - NewDropDownPushButtonText(text string, parent *qt.QWidget)
//   - NewDropDownPushButtonIcon(icon interface{}, text string, parent *qt.QWidget)
type DropDownPushButton struct {
	*PushButton
	menu     *RoundMenu
	arrowAni *translateYAnimation
}

func newDropDownPushButtonBase(parent *qt.QWidget) *DropDownPushButton {
	w := &DropDownPushButton{PushButton: newPushButtonBase(parent)}
	w.SetObjectName("dropDownPushButton")
	w.arrowAni = newTranslateYAnimation(w.QWidget, 2)
	w.paintExtraFunc = w.paintArrow
	w.onPressedFunc = w.arrowAni.press
	w.onReleasedFunc = func() {
		w.arrowAni.release()
		w.showMenu()
	}
	return w
}

// NewDropDownPushButton builds an empty drop-down push button.
func NewDropDownPushButton(parent *qt.QWidget) *DropDownPushButton {
	return newDropDownPushButtonBase(parent)
}

// NewDropDownPushButtonText builds a drop-down push button with text.
func NewDropDownPushButtonText(text string, parent *qt.QWidget) *DropDownPushButton {
	w := NewDropDownPushButton(parent)
	w.SetText(text)
	return w
}

// NewDropDownPushButtonIcon builds a drop-down push button with icon and text.
func NewDropDownPushButtonIcon(icon interface{}, text string, parent *qt.QWidget) *DropDownPushButton {
	w := NewDropDownPushButtonText(text, parent)
	w.SetIcon(icon)
	return w
}

// SetMenu sets the menu shown when the arrow is clicked.
func (w *DropDownPushButton) SetMenu(menu *RoundMenu) { w.menu = menu }

// Menu returns the associated menu.
func (w *DropDownPushButton) Menu() *RoundMenu { return w.menu }

func (w *DropDownPushButton) paintArrow(painter *qt.QPainter) {
	drawDropDownArrow(w.QWidget, w.IsEnabled(), w.isHover, w.isPressed, false, w.arrowAni.y, painter)
}

func (w *DropDownPushButton) showMenu() {
	if w.menu == nil {
		return
	}
	w.menu.ExecAtCentered(w.QWidget)
}

// TransparentDropDownPushButton is a transparent drop-down push button.
//
// Constructors
//   - NewTransparentDropDownPushButton(parent *qt.QWidget)
//   - NewTransparentDropDownPushButtonText(text string, parent *qt.QWidget)
//   - NewTransparentDropDownPushButtonIcon(icon interface{}, text string, parent *qt.QWidget)
type TransparentDropDownPushButton struct{ *DropDownPushButton }

// NewTransparentDropDownPushButton builds an empty transparent drop-down push button.
func NewTransparentDropDownPushButton(parent *qt.QWidget) *TransparentDropDownPushButton {
	w := &TransparentDropDownPushButton{DropDownPushButton: newDropDownPushButtonBase(parent)}
	w.SetObjectName("transparentDropDownPushButton")
	return w
}

// NewTransparentDropDownPushButtonText builds a transparent drop-down push button with text.
func NewTransparentDropDownPushButtonText(text string, parent *qt.QWidget) *TransparentDropDownPushButton {
	w := NewTransparentDropDownPushButton(parent)
	w.SetText(text)
	return w
}

// NewTransparentDropDownPushButtonIcon builds a transparent drop-down push button with icon and text.
func NewTransparentDropDownPushButtonIcon(icon interface{}, text string, parent *qt.QWidget) *TransparentDropDownPushButton {
	w := NewTransparentDropDownPushButtonText(text, parent)
	w.SetIcon(icon)
	return w
}

// PrimaryDropDownPushButton is a primary-color drop-down push button.
//
// Constructors
//   - NewPrimaryDropDownPushButton(parent *qt.QWidget)
//   - NewPrimaryDropDownPushButtonText(text string, parent *qt.QWidget)
//   - NewPrimaryDropDownPushButtonIcon(icon interface{}, text string, parent *qt.QWidget)
type PrimaryDropDownPushButton struct{ *DropDownPushButton }

func newPrimaryDropDownPushButtonBase(parent *qt.QWidget) *PrimaryDropDownPushButton {
	w := &PrimaryDropDownPushButton{DropDownPushButton: newDropDownPushButtonBase(parent)}
	w.SetObjectName("primaryDropDownPushButton")
	w.drawIconFunc = w.PushButton.drawIconPrimary
	w.paintExtraFunc = w.paintArrow
	return w
}

// NewPrimaryDropDownPushButton builds an empty primary drop-down push button.
func NewPrimaryDropDownPushButton(parent *qt.QWidget) *PrimaryDropDownPushButton {
	return newPrimaryDropDownPushButtonBase(parent)
}

// NewPrimaryDropDownPushButtonText builds a primary drop-down push button with text.
func NewPrimaryDropDownPushButtonText(text string, parent *qt.QWidget) *PrimaryDropDownPushButton {
	w := NewPrimaryDropDownPushButton(parent)
	w.SetText(text)
	return w
}

// NewPrimaryDropDownPushButtonIcon builds a primary drop-down push button with icon and text.
func NewPrimaryDropDownPushButtonIcon(icon interface{}, text string, parent *qt.QWidget) *PrimaryDropDownPushButton {
	w := NewPrimaryDropDownPushButtonText(text, parent)
	w.SetIcon(icon)
	return w
}

func (w *PrimaryDropDownPushButton) paintArrow(painter *qt.QPainter) {
	drawDropDownArrow(w.QWidget, w.IsEnabled(), w.isHover, w.isPressed, true, w.arrowAni.y, painter)
}

// DropDownToolButton is a tool button with a drop-down arrow and menu.
//
// Constructors
//   - NewDropDownToolButton(parent *qt.QWidget)
//   - NewDropDownToolButtonIcon(icon interface{}, parent *qt.QWidget)
type DropDownToolButton struct {
	*ToolButton
	menu     *RoundMenu
	arrowAni *translateYAnimation
}

func newDropDownToolButtonBase(parent *qt.QWidget) *DropDownToolButton {
	w := &DropDownToolButton{ToolButton: newToolButtonBase(parent)}
	w.SetObjectName("dropDownToolButton")
	w.arrowAni = newTranslateYAnimation(w.QWidget, 2)
	w.drawIconFunc = func(icon interface{}, painter *qt.QPainter, rect *qt.QRectF) {
		rect.MoveLeft(12)
		w.ToolButton.drawIconBase(icon, painter, rect)
	}
	w.paintExtraFunc = w.paintArrow
	w.onPressedFunc = w.arrowAni.press
	w.onReleasedFunc = func() {
		w.arrowAni.release()
		w.showMenu()
	}
	return w
}

// NewDropDownToolButton builds an empty drop-down tool button.
func NewDropDownToolButton(parent *qt.QWidget) *DropDownToolButton {
	return newDropDownToolButtonBase(parent)
}

// NewDropDownToolButtonIcon builds a drop-down tool button with an icon.
func NewDropDownToolButtonIcon(icon interface{}, parent *qt.QWidget) *DropDownToolButton {
	w := NewDropDownToolButton(parent)
	w.SetIcon(icon)
	return w
}

// SetMenu sets the menu shown when the arrow is clicked.
func (w *DropDownToolButton) SetMenu(menu *RoundMenu) { w.menu = menu }

// Menu returns the associated menu.
func (w *DropDownToolButton) Menu() *RoundMenu { return w.menu }

func (w *DropDownToolButton) paintArrow(painter *qt.QPainter) {
	drawDropDownArrow(w.QWidget, w.IsEnabled(), w.isHover, w.isPressed, false, w.arrowAni.y, painter)
}

func (w *DropDownToolButton) showMenu() {
	if w.menu == nil {
		return
	}
	w.menu.ExecAtCentered(w.QWidget)
}

// TransparentDropDownToolButton is a transparent drop-down tool button.
//
// Constructors
//   - NewTransparentDropDownToolButton(parent *qt.QWidget)
//   - NewTransparentDropDownToolButtonIcon(icon interface{}, parent *qt.QWidget)
type TransparentDropDownToolButton struct{ *DropDownToolButton }

// NewTransparentDropDownToolButton builds an empty transparent drop-down tool button.
func NewTransparentDropDownToolButton(parent *qt.QWidget) *TransparentDropDownToolButton {
	w := &TransparentDropDownToolButton{DropDownToolButton: newDropDownToolButtonBase(parent)}
	w.SetObjectName("transparentDropDownToolButton")
	return w
}

// NewTransparentDropDownToolButtonIcon builds a transparent drop-down tool button with an icon.
func NewTransparentDropDownToolButtonIcon(icon interface{}, parent *qt.QWidget) *TransparentDropDownToolButton {
	w := NewTransparentDropDownToolButton(parent)
	w.SetIcon(icon)
	return w
}

// PrimaryDropDownToolButton is a primary-color drop-down tool button.
//
// Constructors
//   - NewPrimaryDropDownToolButton(parent *qt.QWidget)
//   - NewPrimaryDropDownToolButtonIcon(icon interface{}, parent *qt.QWidget)
type PrimaryDropDownToolButton struct{ *DropDownToolButton }

func newPrimaryDropDownToolButtonBase(parent *qt.QWidget) *PrimaryDropDownToolButton {
	w := &PrimaryDropDownToolButton{DropDownToolButton: newDropDownToolButtonBase(parent)}
	w.SetObjectName("primaryDropDownToolButton")
	w.drawIconFunc = func(icon interface{}, painter *qt.QPainter, rect *qt.QRectF) {
		rect.MoveLeft(12)
		w.ToolButton.drawIconPrimary(icon, painter, rect)
	}
	w.paintExtraFunc = w.paintArrow
	return w
}

// NewPrimaryDropDownToolButton builds an empty primary drop-down tool button.
func NewPrimaryDropDownToolButton(parent *qt.QWidget) *PrimaryDropDownToolButton {
	return newPrimaryDropDownToolButtonBase(parent)
}

// NewPrimaryDropDownToolButtonIcon builds a primary drop-down tool button with an icon.
func NewPrimaryDropDownToolButtonIcon(icon interface{}, parent *qt.QWidget) *PrimaryDropDownToolButton {
	w := NewPrimaryDropDownToolButton(parent)
	w.SetIcon(icon)
	return w
}

func (w *PrimaryDropDownToolButton) paintArrow(painter *qt.QPainter) {
	drawDropDownArrow(w.QWidget, w.IsEnabled(), w.isHover, w.isPressed, true, w.arrowAni.y, painter)
}

// ---------------------------------------------------------------------------
// Split buttons

// newSplitDropButton builds the drop-down arrow button on the right side of a
// split button.
func newSplitDropButton(parent *qt.QWidget, primary bool) *ToolButton {
	b := NewToolButton(parent)
	b.SetIcon(common.ChevronDown)
	b.SetIconSize(qt.NewQSize2(10, 10))
	// Mirror Python SplitDropButton._postInit: horizontal Minimum keeps the drop
	// button at its sizeHint width (8 left padding + 10 icon + 9 right padding +
	// 1 right border = 28px) while vertical Expanding makes it fill the main
	// button's height inside the split layout. SetFixedWidth only pinned the
	// width, leaving the drop button shorter than the main button.
	b.SetSizePolicy2(qt.QSizePolicy__Minimum, qt.QSizePolicy__Expanding)
	if primary {
		b.SetObjectName("primarySplitDropButton")
		b.drawIconFunc = b.drawIconPrimary
	} else {
		b.SetObjectName("splitDropButton")
	}
	return b
}

// SplitPushButton is a push button composed of a main button and a drop-down
// arrow button.
//
// Constructors
//   - NewSplitPushButton(parent *qt.QWidget)
//   - NewSplitPushButtonText(text string, parent *qt.QWidget)
//   - NewSplitPushButtonIcon(icon interface{}, text string, parent *qt.QWidget)
type SplitPushButton struct {
	*qt.QWidget
	Button            *PushButton
	DropButton        *ToolButton
	hBox              *qt.QHBoxLayout
	flyout            interface{}
	onClicked         func()
	onDropDownClicked func()
}

func newSplitPushButtonBase(parent *qt.QWidget, primary bool) *SplitPushButton {
	w := &SplitPushButton{QWidget: qt.NewQWidget(parent)}
	w.SetAttribute(qt.WA_TranslucentBackground)
	w.SetSizePolicy2(qt.QSizePolicy__Fixed, qt.QSizePolicy__Fixed)

	w.hBox = qt.NewQHBoxLayout(w.QWidget)
	w.hBox.SetSpacing(0)
	w.hBox.SetContentsMargins(0, 0, 0, 0)

	w.DropButton = newSplitDropButton(w.QWidget, primary)
	if primary {
		w.Button = NewPrimaryPushButton(w.QWidget).PushButton
		w.Button.SetObjectName("primarySplitPushButton")
	} else {
		w.Button = NewPushButton(w.QWidget)
		w.Button.SetObjectName("splitPushButton")
	}
	w.Button.OnClicked(func() {
		if w.onClicked != nil {
			w.onClicked()
		}
	})
	w.DropButton.OnClicked(func() {
		if w.onDropDownClicked != nil {
			w.onDropDownClicked()
		}
		w.showFlyout()
	})

	w.hBox.AddWidget3(w.Button.QWidget, 1, qt.AlignLeft)
	w.hBox.AddWidget(w.DropButton.QWidget)
	return w
}

// NewSplitPushButton builds an empty split push button.
func NewSplitPushButton(parent *qt.QWidget) *SplitPushButton {
	return newSplitPushButtonBase(parent, false)
}

// NewSplitPushButtonText builds a split push button with text.
func NewSplitPushButtonText(text string, parent *qt.QWidget) *SplitPushButton {
	w := NewSplitPushButton(parent)
	w.SetText(text)
	return w
}

// NewSplitPushButtonIcon builds a split push button with icon and text.
func NewSplitPushButtonIcon(icon interface{}, text string, parent *qt.QWidget) *SplitPushButton {
	w := NewSplitPushButtonText(text, parent)
	w.SetIcon(icon)
	return w
}

// OnClicked registers the main-button click callback.
func (w *SplitPushButton) OnClicked(f func()) { w.onClicked = f }

// OnDropDownClicked registers the drop-down button click callback.
func (w *SplitPushButton) OnDropDownClicked(f func()) { w.onDropDownClicked = f }

// SetFlyout sets the widget/menu shown when the drop-down button is clicked.
func (w *SplitPushButton) SetFlyout(flyout interface{}) { w.flyout = flyout }

// SetText sets the main button text.
func (w *SplitPushButton) SetText(text string) {
	w.Button.SetText(text)
	w.AdjustSize()
}

// Text returns the main button text.
func (w *SplitPushButton) Text() string { return w.Button.Text() }

// SetIcon sets the main button icon.
func (w *SplitPushButton) SetIcon(icon interface{}) { w.Button.SetIcon(icon) }

// Icon returns the main button icon.
func (w *SplitPushButton) Icon() *qt.QIcon { return w.Button.Icon() }

// SetIconSize sets the main button icon size.
func (w *SplitPushButton) SetIconSize(size *qt.QSize) { w.Button.SetIconSize(size) }

// SetDropIcon sets the drop-down button icon.
func (w *SplitPushButton) SetDropIcon(icon interface{}) { w.DropButton.SetIcon(icon) }

// SetDropIconSize sets the drop-down button icon size.
func (w *SplitPushButton) SetDropIconSize(size *qt.QSize) { w.DropButton.SetIconSize(size) }

func (w *SplitPushButton) showFlyout() {
	if w.flyout == nil {
		return
	}
	if menu, ok := w.flyout.(*RoundMenu); ok {
		menu.ExecAtCentered(w.QWidget)
	}
}

// PrimarySplitPushButton is a primary-color split push button.
//
// Constructors
//   - NewPrimarySplitPushButton(parent *qt.QWidget)
//   - NewPrimarySplitPushButtonText(text string, parent *qt.QWidget)
//   - NewPrimarySplitPushButtonIcon(icon interface{}, text string, parent *qt.QWidget)
type PrimarySplitPushButton struct{ *SplitPushButton }

// NewPrimarySplitPushButton builds an empty primary split push button.
func NewPrimarySplitPushButton(parent *qt.QWidget) *PrimarySplitPushButton {
	return &PrimarySplitPushButton{SplitPushButton: newSplitPushButtonBase(parent, true)}
}

// NewPrimarySplitPushButtonText builds a primary split push button with text.
func NewPrimarySplitPushButtonText(text string, parent *qt.QWidget) *PrimarySplitPushButton {
	w := NewPrimarySplitPushButton(parent)
	w.SetText(text)
	return w
}

// NewPrimarySplitPushButtonIcon builds a primary split push button with icon and text.
func NewPrimarySplitPushButtonIcon(icon interface{}, text string, parent *qt.QWidget) *PrimarySplitPushButton {
	w := NewPrimarySplitPushButtonText(text, parent)
	w.SetIcon(icon)
	return w
}

// SplitToolButton is a tool button composed of a main tool button and a
// drop-down arrow button.
//
// Constructors
//   - NewSplitToolButton(parent *qt.QWidget)
//   - NewSplitToolButtonIcon(icon interface{}, parent *qt.QWidget)
type SplitToolButton struct {
	*qt.QWidget
	button            *ToolButton
	dropButton        *ToolButton
	hBox              *qt.QHBoxLayout
	flyout            interface{}
	onClicked         func()
	onDropDownClicked func()
}

func newSplitToolButtonBase(parent *qt.QWidget, primary bool) *SplitToolButton {
	w := &SplitToolButton{QWidget: qt.NewQWidget(parent)}
	w.SetAttribute(qt.WA_TranslucentBackground)
	w.SetSizePolicy2(qt.QSizePolicy__Fixed, qt.QSizePolicy__Fixed)

	w.hBox = qt.NewQHBoxLayout(w.QWidget)
	w.hBox.SetSpacing(0)
	w.hBox.SetContentsMargins(0, 0, 0, 0)

	w.dropButton = newSplitDropButton(w.QWidget, primary)
	if primary {
		w.button = NewPrimaryToolButton(w.QWidget).ToolButton
		w.button.SetObjectName("primarySplitToolButton")
	} else {
		w.button = NewToolButton(w.QWidget)
		w.button.SetObjectName("splitToolButton")
	}
	w.button.OnClicked(func() {
		if w.onClicked != nil {
			w.onClicked()
		}
	})
	w.dropButton.OnClicked(func() {
		if w.onDropDownClicked != nil {
			w.onDropDownClicked()
		}
		w.showFlyout()
	})

	w.hBox.AddWidget3(w.button.QWidget, 1, qt.AlignLeft)
	w.hBox.AddWidget(w.dropButton.QWidget)
	return w
}

// NewSplitToolButton builds an empty split tool button.
func NewSplitToolButton(parent *qt.QWidget) *SplitToolButton {
	return newSplitToolButtonBase(parent, false)
}

// NewSplitToolButtonIcon builds a split tool button with an icon.
func NewSplitToolButtonIcon(icon interface{}, parent *qt.QWidget) *SplitToolButton {
	w := NewSplitToolButton(parent)
	w.SetIcon(icon)
	return w
}

// OnClicked registers the main-button click callback.
func (w *SplitToolButton) OnClicked(f func()) { w.onClicked = f }

// OnDropDownClicked registers the drop-down button click callback.
func (w *SplitToolButton) OnDropDownClicked(f func()) { w.onDropDownClicked = f }

// SetFlyout sets the widget/menu shown when the drop-down button is clicked.
func (w *SplitToolButton) SetFlyout(flyout interface{}) { w.flyout = flyout }

// SetIcon sets the main button icon.
func (w *SplitToolButton) SetIcon(icon interface{}) { w.button.SetIcon(icon) }

// Icon returns the main button icon.
func (w *SplitToolButton) Icon() *qt.QIcon { return w.button.Icon() }

// SetIconSize sets the main button icon size.
func (w *SplitToolButton) SetIconSize(size *qt.QSize) { w.button.SetIconSize(size) }

// SetDropIcon sets the drop-down button icon.
func (w *SplitToolButton) SetDropIcon(icon interface{}) { w.dropButton.SetIcon(icon) }

// SetDropIconSize sets the drop-down button icon size.
func (w *SplitToolButton) SetDropIconSize(size *qt.QSize) { w.dropButton.SetIconSize(size) }

func (w *SplitToolButton) showFlyout() {
	if w.flyout == nil {
		return
	}
	if menu, ok := w.flyout.(*RoundMenu); ok {
		menu.ExecAtCentered(w.QWidget)
	}
}

// PrimarySplitToolButton is a primary-color split tool button.
//
// Constructors
//   - NewPrimarySplitToolButton(parent *qt.QWidget)
//   - NewPrimarySplitToolButtonIcon(icon interface{}, parent *qt.QWidget)
type PrimarySplitToolButton struct{ *SplitToolButton }

// NewPrimarySplitToolButton builds an empty primary split tool button.
func NewPrimarySplitToolButton(parent *qt.QWidget) *PrimarySplitToolButton {
	return &PrimarySplitToolButton{SplitToolButton: newSplitToolButtonBase(parent, true)}
}

// NewPrimarySplitToolButtonIcon builds a primary split tool button with an icon.
func NewPrimarySplitToolButtonIcon(icon interface{}, parent *qt.QWidget) *PrimarySplitToolButton {
	w := NewPrimarySplitToolButton(parent)
	w.SetIcon(icon)
	return w
}

// ---------------------------------------------------------------------------
// Pill buttons

// paintPillBackground draws the rounded pill background shared by PillPushButton
// and PillToolButton.
func paintPillBackground(w *qt.QWidget, checked, enabled, pressed, hover bool) {
	painter := qt.NewQPainter2(w.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__Antialiasing)

	isDark := common.IsDarkTheme()
	var rect *qt.QRect
	var borderColor, bgColor *qt.QColor
	ownsRect := false

	if !checked {
		base := w.Rect()
		rect = qt.NewQRect4(base.X()+1, base.Y()+1, base.Width()-2, base.Height()-2)
		ownsRect = true

		if isDark {
			borderColor = qt.NewQColor11(255, 255, 255, 18)
		} else {
			borderColor = qt.NewQColor11(0, 0, 0, 15)
		}
		if !enabled {
			if isDark {
				bgColor = qt.NewQColor11(255, 255, 255, 11)
			} else {
				bgColor = qt.NewQColor11(249, 249, 249, 75)
			}
		} else if pressed || hover {
			if isDark {
				bgColor = qt.NewQColor11(255, 255, 255, 21)
			} else {
				bgColor = qt.NewQColor11(249, 249, 249, 128)
			}
		} else if isDark {
			bgColor = qt.NewQColor11(255, 255, 255, 15)
		} else {
			bgColor = qt.NewQColor11(243, 243, 243, 194)
		}
	} else {
		if !enabled {
			if isDark {
				bgColor = qt.NewQColor11(255, 255, 255, 40)
			} else {
				bgColor = qt.NewQColor11(0, 0, 0, 55)
			}
		} else if pressed {
			if isDark {
				bgColor = common.ThemeColorDark2.Color()
			} else {
				bgColor = common.ThemeColorLight3.Color()
			}
		} else if hover {
			if isDark {
				bgColor = common.ThemeColorDark1.Color()
			} else {
				bgColor = common.ThemeColorLight1.Color()
			}
		} else {
			bgColor = common.ThemeColorPrimary.Color()
		}
		borderColor = qt.NewQColor11(0, 0, 0, 0)
		rect = w.Rect() // GoGC-armed — do NOT Delete
	}
	if ownsRect {
		defer rect.Delete()
	}
	defer borderColor.Delete()
	defer bgColor.Delete()

	painter.SetPen(borderColor)
	brush := qt.NewQBrush3(bgColor)
	painter.SetBrush(brush)
	brush.Delete()

	r := float64(rect.Height()) / 2
	painter.DrawRoundedRect3(rect, r, r)
	painter.End()
}

// PillPushButton is a pill-shaped toggle push button.
//
// Constructors
//   - NewPillPushButton(parent *qt.QWidget)
//   - NewPillPushButtonText(text string, parent *qt.QWidget)
//   - NewPillPushButtonIcon(icon interface{}, text string, parent *qt.QWidget)
type PillPushButton struct{ *ToggleButton }

// NewPillPushButton builds an empty pill push button.
func NewPillPushButton(parent *qt.QWidget) *PillPushButton {
	w := &PillPushButton{ToggleButton: NewToggleButton(parent)}
	w.SetObjectName("pillPushButton")
	w.installPillPaint()
	return w
}

// NewPillPushButtonText builds a pill push button with text.
func NewPillPushButtonText(text string, parent *qt.QWidget) *PillPushButton {
	w := NewPillPushButton(parent)
	w.SetText(text)
	return w
}

// NewPillPushButtonIcon builds a pill push button with icon and text.
func NewPillPushButtonIcon(icon interface{}, text string, parent *qt.QWidget) *PillPushButton {
	w := NewPillPushButtonText(text, parent)
	w.SetIcon(icon)
	return w
}

func (w *PillPushButton) installPillPaint() {
	w.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		paintPillBackground(w.QWidget, w.IsChecked(), w.IsEnabled(), w.isPressed, w.isHover)
		super(e)
		w.paintIcon()
	})
}

// PillToolButton is a pill-shaped toggle tool button.
//
// Constructors
//   - NewPillToolButton(parent *qt.QWidget)
//   - NewPillToolButtonIcon(icon interface{}, parent *qt.QWidget)
type PillToolButton struct{ *ToggleToolButton }

// NewPillToolButton builds an empty pill tool button.
func NewPillToolButton(parent *qt.QWidget) *PillToolButton {
	w := &PillToolButton{ToggleToolButton: NewToggleToolButton(parent)}
	w.SetObjectName("pillToolButton")
	w.installPillPaint()
	return w
}

// NewPillToolButtonIcon builds a pill tool button with an icon.
func NewPillToolButtonIcon(icon interface{}, parent *qt.QWidget) *PillToolButton {
	w := NewPillToolButton(parent)
	w.SetIcon(icon)
	return w
}

func (w *PillToolButton) installPillPaint() {
	w.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		paintPillBackground(w.QWidget, w.IsChecked(), w.IsEnabled(), w.isPressed, w.isHover)
		super(e)
		w.paintIcon()
	})
}
