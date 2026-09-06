package widgets

import (
	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// SpinIcon enumerates the two spin box arrow glyphs. The glyphs live under
// images/spin_box/ and are rendered through readEmbeddedImage.
type SpinIcon string

const (
	// SpinIconUp is the increment (up) arrow.
	SpinIconUp SpinIcon = "Up"
	// SpinIconDown is the decrement (down) arrow.
	SpinIconDown SpinIcon = "Down"
)

// render draws the spin icon SVG into painter at rect.
func (s SpinIcon) render(painter *qt.QPainter, rect *qt.QRectF) {
	color := common.GetIconColor(common.ThemeAuto, false)
	b := readEmbeddedImage("spin_box/" + string(s) + "_" + color + ".svg")
	drawSvgBytes(b, painter, rect)
}

// SpinButton is a tool button that renders a SpinIcon (used by inline spin
// boxes for the up/down arrows).
type SpinButton struct {
	*qt.QToolButton
	icon      SpinIcon
	isPressed bool
}

// NewSpinButton builds a spin button.
func NewSpinButton(icon SpinIcon, parent *qt.QWidget) *SpinButton {
	w := &SpinButton{QToolButton: qt.NewQToolButton(parent), icon: icon}
	w.SetObjectName("spinButton")
	w.SetFixedSize2(31, 23)
	size := qt.NewQSize2(10, 10)
	w.SetIconSize(size)
	size.Delete()
	common.FluentStyleSheet(common.FluentSpinBox).Apply(w.QWidget, common.ThemeAuto)

	w.OnMousePressEvent(func(super func(event *qt.QMouseEvent), event *qt.QMouseEvent) {
		w.isPressed = true
		super(event)
	})
	w.OnMouseReleaseEvent(func(super func(event *qt.QMouseEvent), event *qt.QMouseEvent) {
		w.isPressed = false
		super(event)
	})
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
		super(event)
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing | qt.QPainter__SmoothPixmapTransform)

		if !w.IsEnabled() {
			painter.SetOpacity(0.36)
		} else if w.isPressed {
			painter.SetOpacity(0.7)
		}

		// Match Python's SpinButton.paintEvent: render at a fixed 11x11 rect
		// centred in the 31x23 button (the iconSize is only used for the
		// internal button layout, not for the manual glyph draw).
		rect := qt.NewQRectF4(10, 6.5, 11, 11)
		defer rect.Delete()
		w.icon.render(painter, rect)
		painter.End()
	})
	return w
}

// CompactSpinButton is a compact tool button that renders both spin arrows and
// steps up/down depending on which half is clicked.
type CompactSpinButton struct {
	*qt.QToolButton
	onUp   func()
	onDown func()
}

// NewCompactSpinButton builds a compact spin button.
func NewCompactSpinButton(parent *qt.QWidget) *CompactSpinButton {
	w := &CompactSpinButton{QToolButton: qt.NewQToolButton(parent)}
	w.SetFixedSize2(26, 33)
	w.SetCursor(qt.NewQCursor2(qt.IBeamCursor))

	w.OnMouseReleaseEvent(func(super func(event *qt.QMouseEvent), event *qt.QMouseEvent) {
		super(event)
		if event.Y() < w.Height()/2 {
			if w.onUp != nil {
				w.onUp()
			}
		} else if w.onDown != nil {
			w.onDown()
		}
	})
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing)

		x := float64(w.Width()-10) / 2
		s := float64(9)
		upRect := qt.NewQRectF4(x, float64(w.Height())/2-s+1, s, s)
		defer upRect.Delete()
		downRect := qt.NewQRectF4(x, float64(w.Height())/2, s, s)
		defer downRect.Delete()

		SpinIconUp.render(painter, upRect)
		SpinIconDown.render(painter, downRect)
		painter.End()
	})
	return w
}

// SpinFlyoutView is the pop-up view with two big step buttons. In this port it
// is a standalone widget (the Flyout host is simplified away).
type SpinFlyoutView struct {
	*qt.QWidget
	upButton   *SpinButton
	downButton *SpinButton
	vBoxLayout *qt.QVBoxLayout
}

// NewSpinFlyoutView builds a spin flyout view.
func NewSpinFlyoutView(parent *qt.QWidget) *SpinFlyoutView {
	w := &SpinFlyoutView{QWidget: qt.NewQWidget(parent)}
	w.upButton = NewSpinButton(SpinIconUp, w.QWidget)
	w.downButton = NewSpinButton(SpinIconDown, w.QWidget)
	w.vBoxLayout = qt.NewQVBoxLayout(w.QWidget)

	w.upButton.SetFixedSize2(36, 36)
	w.downButton.SetFixedSize2(36, 36)
	size := qt.NewQSize2(13, 13)
	w.upButton.SetIconSize(size)
	w.downButton.SetIconSize(size)
	size.Delete()

	w.vBoxLayout.SetContentsMargins(6, 6, 6, 6)
	w.vBoxLayout.SetSpacing(0)
	w.vBoxLayout.AddWidget(w.upButton.QWidget)
	w.vBoxLayout.AddWidget(w.downButton.QWidget)

	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing)

		var bg, border *qt.QColor
		if common.IsDarkTheme() {
			bg = qt.NewQColor3(46, 46, 46)
			border = qt.NewQColor11(0, 0, 0, 51)
		} else {
			bg = qt.NewQColor3(249, 249, 249)
			border = qt.NewQColor11(0, 0, 0, 15)
		}
		defer bg.Delete()
		defer border.Delete()

		brush := qt.NewQBrush3(bg)
		defer brush.Delete()
		painter.SetBrush(brush)
		painter.SetPen(border)

		rect := qt.NewQRectF4(1, 1, float64(w.Width()-2), float64(w.Height()-2))
		defer rect.Delete()
		painter.DrawRoundedRect(rect, 8, 8)
		painter.End()
	})
	return w
}

// spinBoxBase holds the shared error/border state and paints the focused
// underline for every spin box flavor.
type spinBoxBase struct {
	abstract                *qt.QAbstractSpinBox
	isError                 bool
	lightFocusedBorderColor *qt.QColor
	darkFocusedBorderColor  *qt.QColor
}

func newSpinBoxBase(abstract *qt.QAbstractSpinBox) spinBoxBase {
	abstract.SetProperty("transparent", qt.NewQVariant11(true))
	common.FluentStyleSheet(common.FluentSpinBox).Apply(abstract.QWidget, common.ThemeAuto)
	abstract.SetButtonSymbols(qt.QAbstractSpinBox__NoButtons)
	abstract.SetFixedHeight(33)
	common.SetFont(abstract.QWidget, 14, 400)
	abstract.SetAttribute2(qt.WA_MacShowFocusRect, false)
	return spinBoxBase{
		abstract:                abstract,
		lightFocusedBorderColor: qt.NewQColor(),
		darkFocusedBorderColor:  qt.NewQColor(),
	}
}

// IsError reports whether the spin box is in error status.
func (b *spinBoxBase) IsError() bool { return b.isError }

// SetError sets the error status.
func (b *spinBoxBase) SetError(isError bool) {
	if isError == b.isError {
		return
	}
	b.isError = isError
	b.abstract.Update()
}

// SetCustomFocusedBorderColor sets the focused border color for light/dark mode.
func (b *spinBoxBase) SetCustomFocusedBorderColor(light, dark *qt.QColor) {
	b.lightFocusedBorderColor = cloneColor(light)
	b.darkFocusedBorderColor = cloneColor(dark)
	b.abstract.Update()
}

func (b *spinBoxBase) focusedBorderColor() *qt.QColor {
	if b.isError {
		return common.CriticalForeground.Color(common.ThemeAuto)
	}
	var c *qt.QColor
	if common.IsDarkTheme() {
		c = b.darkFocusedBorderColor
	} else {
		c = b.lightFocusedBorderColor
	}
	if c != nil && c.IsValid() {
		return cloneColor(c)
	}
	return common.ThemeColorPrimary.Color()
}

func (b *spinBoxBase) installBorderPaint() {
	// b.abstract is a promoted (non-directly-constructed) QAbstractSpinBox, so
	// its virtual OnPaintEvent cannot be overridden. Install an event filter
	// instead and draw the focus border on the paint event.
	filter := qt.NewQObject2(b.abstract.QObject)
	b.abstract.InstallEventFilter(filter)
	filter.OnEventFilter(func(super func(watched *qt.QObject, event *qt.QEvent) bool, watched *qt.QObject, event *qt.QEvent) bool {
		if event.Type() == qt.QEvent__Paint {
			b.drawBorderBottom()
		}
		return super(watched, event)
	})
}

func (b *spinBoxBase) drawBorderBottom() {
	if !b.abstract.HasFocus() {
		return
	}
	painter := qt.NewQPainter2(b.abstract.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__Antialiasing)
	painter.SetPenWithStyle(qt.NoPen)

	wd, ht := b.abstract.Width(), b.abstract.Height()
	path := qt.NewQPainterPath()
	defer path.Delete()
	rect := qt.NewQRectF4(0, float64(ht-10), float64(wd), 10)
	defer rect.Delete()
	path.AddRoundedRect(rect, 5, 5)

	cutPath := qt.NewQPainterPath()
	defer cutPath.Delete()
	cut := qt.NewQRectF4(0, float64(ht-10), float64(wd), 8)
	defer cut.Delete()
	cutPath.AddRect(cut)

	sub := path.Subtracted(cutPath) // GoGC-armed — do NOT Delete

	color := b.focusedBorderColor()
	defer color.Delete()
	brush := qt.NewQBrush3(color)
	defer brush.Delete()
	painter.FillPath(sub, brush)
	painter.End()
}

// inlineSpinUI is the inline spin box UI (two SpinButtons embedded on the right).
type inlineSpinUI struct {
	spinBoxBase
	upButton   *SpinButton
	downButton *SpinButton
	hBoxLayout *qt.QHBoxLayout
}

func newInlineSpinUI(abstract *qt.QAbstractSpinBox) *inlineSpinUI {
	u := &inlineSpinUI{spinBoxBase: newSpinBoxBase(abstract)}
	u.upButton = NewSpinButton(SpinIconUp, abstract.QWidget)
	u.downButton = NewSpinButton(SpinIconDown, abstract.QWidget)
	u.hBoxLayout = qt.NewQHBoxLayout(abstract.QWidget)
	u.hBoxLayout.SetContentsMargins(0, 4, 4, 4)
	u.hBoxLayout.SetSpacing(5)
	u.hBoxLayout.AddWidget3(u.upButton.QWidget, 0, qt.AlignRight|qt.AlignVCenter)
	u.hBoxLayout.AddWidget3(u.downButton.QWidget, 0, qt.AlignRight|qt.AlignVCenter)
	// Port of spin_box.py's `hBoxLayout.setAlignment(Qt.AlignRight |
	// Qt.AlignVCenter)`: right-aligns the button group against the spinbox's
	// right edge so the up/down buttons sit flush beside the reserved right
	// padding instead of leaving a gap on the right. The alignment lives on
	// QLayoutItem, so it is reached through the promoted embedded field (the
	// same-name QLayout.SetAlignment overload takes a child widget).
	u.hBoxLayout.QLayoutItem.SetAlignment(qt.AlignRight | qt.AlignVCenter)

	u.upButton.OnClicked(abstract.StepUp)
	u.downButton.OnClicked(abstract.StepDown)
	u.installBorderPaint()
	return u
}

// SetSymbolVisible toggles the spin symbols.
func (u *inlineSpinUI) SetSymbolVisible(isVisible bool) {
	u.abstract.SetProperty("symbolVisible", qt.NewQVariant11(isVisible))
	u.abstract.SetStyle(qt.QApplication_Style())
	u.upButton.SetVisible(isVisible)
	u.downButton.SetVisible(isVisible)
}

// compactSpinUI is the compact spin box UI (one CompactSpinButton on the right).
type compactSpinUI struct {
	spinBoxBase
	compactSpinButton *CompactSpinButton
	hBoxLayout        *qt.QHBoxLayout
}

func newCompactSpinUI(abstract *qt.QAbstractSpinBox) *compactSpinUI {
	u := &compactSpinUI{spinBoxBase: newSpinBoxBase(abstract)}
	u.compactSpinButton = NewCompactSpinButton(abstract.QWidget)
	u.hBoxLayout = qt.NewQHBoxLayout(abstract.QWidget)
	u.hBoxLayout.SetContentsMargins(0, 0, 0, 0)
	u.hBoxLayout.AddWidget3(u.compactSpinButton.QWidget, 0, qt.AlignRight|qt.AlignVCenter)
	u.hBoxLayout.QLayoutItem.SetAlignment(qt.AlignRight | qt.AlignVCenter)

	u.compactSpinButton.onUp = abstract.StepUp
	u.compactSpinButton.onDown = abstract.StepDown
	u.installBorderPaint()
	return u
}

// SetSymbolVisible toggles the spin symbol.
func (u *compactSpinUI) SetSymbolVisible(isVisible bool) {
	u.abstract.SetProperty("symbolVisible", qt.NewQVariant11(isVisible))
	u.abstract.SetStyle(qt.QApplication_Style())
	u.compactSpinButton.SetVisible(isVisible)
}

// SpinBox is a fluent spin box with inline step buttons.
type SpinBox struct {
	*qt.QSpinBox
	*inlineSpinUI
}

// NewSpinBox builds a spin box.
func NewSpinBox(parent *qt.QWidget) *SpinBox {
	q := qt.NewQSpinBox(parent)
	return &SpinBox{QSpinBox: q, inlineSpinUI: newInlineSpinUI(q.QAbstractSpinBox)}
}

// SetAccelerated toggles auto-repeat acceleration.
func (w *SpinBox) SetAccelerated(on bool) {
	w.QSpinBox.SetAccelerated(on)
	w.upButton.SetAutoRepeat(on)
	w.downButton.SetAutoRepeat(on)
}

// SetReadOnly sets the read-only status and hides the step buttons.
func (w *SpinBox) SetReadOnly(isReadOnly bool) {
	w.QSpinBox.SetReadOnly(isReadOnly)
	w.SetSymbolVisible(!isReadOnly)
}

// DoubleSpinBox is a fluent double spin box with inline step buttons.
type DoubleSpinBox struct {
	*qt.QDoubleSpinBox
	*inlineSpinUI
}

// NewDoubleSpinBox builds a double spin box.
func NewDoubleSpinBox(parent *qt.QWidget) *DoubleSpinBox {
	q := qt.NewQDoubleSpinBox(parent)
	return &DoubleSpinBox{QDoubleSpinBox: q, inlineSpinUI: newInlineSpinUI(q.QAbstractSpinBox)}
}

// SetAccelerated toggles auto-repeat acceleration.
func (w *DoubleSpinBox) SetAccelerated(on bool) {
	w.QDoubleSpinBox.SetAccelerated(on)
	w.upButton.SetAutoRepeat(on)
	w.downButton.SetAutoRepeat(on)
}

// SetReadOnly sets the read-only status and hides the step buttons.
func (w *DoubleSpinBox) SetReadOnly(isReadOnly bool) {
	w.QDoubleSpinBox.SetReadOnly(isReadOnly)
	w.SetSymbolVisible(!isReadOnly)
}

// DateEdit is a fluent date edit with inline step buttons.
type DateEdit struct {
	*qt.QDateEdit
	*inlineSpinUI
}

// NewDateEdit builds a date edit.
func NewDateEdit(parent *qt.QWidget) *DateEdit {
	q := qt.NewQDateEdit(parent)
	return &DateEdit{QDateEdit: q, inlineSpinUI: newInlineSpinUI(q.QAbstractSpinBox)}
}

// SetAccelerated toggles auto-repeat acceleration.
func (w *DateEdit) SetAccelerated(on bool) {
	w.QDateEdit.SetAccelerated(on)
	w.upButton.SetAutoRepeat(on)
	w.downButton.SetAutoRepeat(on)
}

// SetReadOnly sets the read-only status and hides the step buttons.
func (w *DateEdit) SetReadOnly(isReadOnly bool) {
	w.QDateEdit.SetReadOnly(isReadOnly)
	w.SetSymbolVisible(!isReadOnly)
}

// DateTimeEdit is a fluent date-time edit with inline step buttons.
type DateTimeEdit struct {
	*qt.QDateTimeEdit
	*inlineSpinUI
}

// NewDateTimeEdit builds a date-time edit.
func NewDateTimeEdit(parent *qt.QWidget) *DateTimeEdit {
	q := qt.NewQDateTimeEdit(parent)
	return &DateTimeEdit{QDateTimeEdit: q, inlineSpinUI: newInlineSpinUI(q.QAbstractSpinBox)}
}

// SetAccelerated toggles auto-repeat acceleration.
func (w *DateTimeEdit) SetAccelerated(on bool) {
	w.QDateTimeEdit.SetAccelerated(on)
	w.upButton.SetAutoRepeat(on)
	w.downButton.SetAutoRepeat(on)
}

// SetReadOnly sets the read-only status and hides the step buttons.
func (w *DateTimeEdit) SetReadOnly(isReadOnly bool) {
	w.QDateTimeEdit.SetReadOnly(isReadOnly)
	w.SetSymbolVisible(!isReadOnly)
}

// TimeEdit is a fluent time edit with inline step buttons.
type TimeEdit struct {
	*qt.QTimeEdit
	*inlineSpinUI
}

// NewTimeEdit builds a time edit.
func NewTimeEdit(parent *qt.QWidget) *TimeEdit {
	q := qt.NewQTimeEdit(parent)
	return &TimeEdit{QTimeEdit: q, inlineSpinUI: newInlineSpinUI(q.QAbstractSpinBox)}
}

// SetAccelerated toggles auto-repeat acceleration.
func (w *TimeEdit) SetAccelerated(on bool) {
	w.QTimeEdit.SetAccelerated(on)
	w.upButton.SetAutoRepeat(on)
	w.downButton.SetAutoRepeat(on)
}

// SetReadOnly sets the read-only status and hides the step buttons.
func (w *TimeEdit) SetReadOnly(isReadOnly bool) {
	w.QTimeEdit.SetReadOnly(isReadOnly)
	w.SetSymbolVisible(!isReadOnly)
}

// CompactSpinBox is a fluent compact spin box.
type CompactSpinBox struct {
	*qt.QSpinBox
	*compactSpinUI
}

// NewCompactSpinBox builds a compact spin box.
func NewCompactSpinBox(parent *qt.QWidget) *CompactSpinBox {
	q := qt.NewQSpinBox(parent)
	q.SetObjectName("compactSpinBox")
	return &CompactSpinBox{QSpinBox: q, compactSpinUI: newCompactSpinUI(q.QAbstractSpinBox)}
}

// SetAccelerated toggles auto-repeat acceleration.
func (w *CompactSpinBox) SetAccelerated(on bool) { w.QSpinBox.SetAccelerated(on) }

// SetReadOnly sets the read-only status and hides the step symbol.
func (w *CompactSpinBox) SetReadOnly(isReadOnly bool) {
	w.QSpinBox.SetReadOnly(isReadOnly)
	w.SetSymbolVisible(!isReadOnly)
}

// CompactDoubleSpinBox is a fluent compact double spin box.
type CompactDoubleSpinBox struct {
	*qt.QDoubleSpinBox
	*compactSpinUI
}

// NewCompactDoubleSpinBox builds a compact double spin box.
func NewCompactDoubleSpinBox(parent *qt.QWidget) *CompactDoubleSpinBox {
	q := qt.NewQDoubleSpinBox(parent)
	q.SetObjectName("compactDoubleSpinBox")
	return &CompactDoubleSpinBox{QDoubleSpinBox: q, compactSpinUI: newCompactSpinUI(q.QAbstractSpinBox)}
}

// SetAccelerated toggles auto-repeat acceleration.
func (w *CompactDoubleSpinBox) SetAccelerated(on bool) { w.QDoubleSpinBox.SetAccelerated(on) }

// SetReadOnly sets the read-only status and hides the step symbol.
func (w *CompactDoubleSpinBox) SetReadOnly(isReadOnly bool) {
	w.QDoubleSpinBox.SetReadOnly(isReadOnly)
	w.SetSymbolVisible(!isReadOnly)
}

// CompactDateEdit is a fluent compact date edit.
type CompactDateEdit struct {
	*qt.QDateEdit
	*compactSpinUI
}

// NewCompactDateEdit builds a compact date edit.
func NewCompactDateEdit(parent *qt.QWidget) *CompactDateEdit {
	q := qt.NewQDateEdit(parent)
	q.SetObjectName("compactDateEdit")
	return &CompactDateEdit{QDateEdit: q, compactSpinUI: newCompactSpinUI(q.QAbstractSpinBox)}
}

// SetAccelerated toggles auto-repeat acceleration.
func (w *CompactDateEdit) SetAccelerated(on bool) { w.QDateEdit.SetAccelerated(on) }

// SetReadOnly sets the read-only status and hides the step symbol.
func (w *CompactDateEdit) SetReadOnly(isReadOnly bool) {
	w.QDateEdit.SetReadOnly(isReadOnly)
	w.SetSymbolVisible(!isReadOnly)
}

// CompactDateTimeEdit is a fluent compact date-time edit.
type CompactDateTimeEdit struct {
	*qt.QDateTimeEdit
	*compactSpinUI
}

// NewCompactDateTimeEdit builds a compact date-time edit.
func NewCompactDateTimeEdit(parent *qt.QWidget) *CompactDateTimeEdit {
	q := qt.NewQDateTimeEdit(parent)
	q.SetObjectName("compactDateTimeEdit")
	return &CompactDateTimeEdit{QDateTimeEdit: q, compactSpinUI: newCompactSpinUI(q.QAbstractSpinBox)}
}

// SetAccelerated toggles auto-repeat acceleration.
func (w *CompactDateTimeEdit) SetAccelerated(on bool) { w.QDateTimeEdit.SetAccelerated(on) }

// SetReadOnly sets the read-only status and hides the step symbol.
func (w *CompactDateTimeEdit) SetReadOnly(isReadOnly bool) {
	w.QDateTimeEdit.SetReadOnly(isReadOnly)
	w.SetSymbolVisible(!isReadOnly)
}

// CompactTimeEdit is a fluent compact time edit.
type CompactTimeEdit struct {
	*qt.QTimeEdit
	*compactSpinUI
}

// NewCompactTimeEdit builds a compact time edit.
func NewCompactTimeEdit(parent *qt.QWidget) *CompactTimeEdit {
	q := qt.NewQTimeEdit(parent)
	q.SetObjectName("compactTimeEdit")
	return &CompactTimeEdit{QTimeEdit: q, compactSpinUI: newCompactSpinUI(q.QAbstractSpinBox)}
}

// SetAccelerated toggles auto-repeat acceleration.
func (w *CompactTimeEdit) SetAccelerated(on bool) { w.QTimeEdit.SetAccelerated(on) }

// SetReadOnly sets the read-only status and hides the step symbol.
func (w *CompactTimeEdit) SetReadOnly(isReadOnly bool) {
	w.QTimeEdit.SetReadOnly(isReadOnly)
	w.SetSymbolVisible(!isReadOnly)
}
