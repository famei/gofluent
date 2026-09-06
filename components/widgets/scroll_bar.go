package widgets

import (
	"math"

	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// ArrowButton is the small arrow button rendered at both ends of a fluent
// scroll bar groove.
type ArrowButton struct {
	*qt.QToolButton
	icon       common.FluentIcon
	lightColor *qt.QColor
	darkColor  *qt.QColor
	opacity    float64
}

// NewArrowButton builds an arrow button from a FluentIcon.
func NewArrowButton(icon common.FluentIcon, parent *qt.QWidget) *ArrowButton {
	w := &ArrowButton{
		QToolButton: qt.NewQToolButton(parent),
		icon:        icon,
		lightColor:  qt.NewQColor11(0, 0, 0, 114),
		darkColor:   qt.NewQColor11(255, 255, 255, 139),
		opacity:     1,
	}
	w.SetFixedSize2(10, 10)
	w.installPaintEvent()
	return w
}

// SetOpacity sets the arrow opacity (0..1).
func (w *ArrowButton) SetOpacity(opacity float64) {
	w.opacity = opacity
	w.Update()
}

// SetLightColor sets the light-theme arrow color.
func (w *ArrowButton) SetLightColor(color *qt.QColor) {
	w.lightColor = cloneColor(color)
	w.Update()
}

// SetDarkColor sets the dark-theme arrow color.
func (w *ArrowButton) SetDarkColor(color *qt.QColor) {
	w.darkColor = cloneColor(color)
	w.Update()
}

func (w *ArrowButton) installPaintEvent() {
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
		// Do NOT call super(event): the native QToolButton chrome (border /
		// background / bevel) would be painted under the Fluent icon and make
		// the arrow buttons look non-fluent. Python's ArrowButton.paintEvent
		// only paints the icon.
		_ = super
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing)

		var color *qt.QColor
		if common.IsDarkTheme() {
			color = w.darkColor
		} else {
			color = w.lightColor
		}
		painter.SetOpacity(w.opacity * float64(color.Alpha()) / 255)

		s := 8.0
		if w.IsDown() {
			s = 7.0
		}
		x := (float64(w.Width()) - s) / 2
		rect := qt.NewQRectF4(x, x, s, s)
		defer rect.Delete()
		renderFluentIconWithFill(w.icon, painter, rect, color.Name())
		painter.End()
	})
}

// ScrollBarGroove is the rounded groove background of a fluent scroll bar.
type ScrollBarGroove struct {
	*qt.QWidget
	orientation          qt.Orientation
	opacity              float64
	lightBackgroundColor *qt.QColor
	darkBackgroundColor  *qt.QColor
	upButton             *ArrowButton
	downButton           *ArrowButton
	opacityAni           *common.ProgressAnimation
	onOpacityChanged     func(float64)
}

// NewScrollBarGroove builds a scroll bar groove for the given orientation.
func NewScrollBarGroove(orient qt.Orientation, parent *qt.QWidget) *ScrollBarGroove {
	w := &ScrollBarGroove{
		QWidget:              qt.NewQWidget(parent),
		orientation:          orient,
		opacity:              1,
		lightBackgroundColor: qt.NewQColor11(252, 252, 252, 217),
		darkBackgroundColor:  qt.NewQColor11(44, 44, 44, 245),
	}
	if orient == qt.Vertical {
		w.SetFixedWidth(12)
		w.upButton = NewArrowButton(common.CareUpSolid, w.QWidget)
		w.downButton = NewArrowButton(common.CareDownSolid, w.QWidget)
		layout := qt.NewQVBoxLayout(w.QWidget)
		layout.AddWidget3(w.upButton.QWidget, 0, qt.AlignHCenter)
		layout.AddStretchWithStretch(1)
		layout.AddWidget3(w.downButton.QWidget, 0, qt.AlignHCenter)
		layout.SetContentsMargins(0, 3, 0, 3)
	} else {
		w.SetFixedHeight(12)
		w.upButton = NewArrowButton(common.CareLeftSolid, w.QWidget)
		w.downButton = NewArrowButton(common.CareRightSolid, w.QWidget)
		layout := qt.NewQHBoxLayout(w.QWidget)
		layout.AddWidget3(w.upButton.QWidget, 0, qt.AlignVCenter)
		layout.AddStretchWithStretch(1)
		layout.AddWidget3(w.downButton.QWidget, 0, qt.AlignVCenter)
		layout.SetContentsMargins(3, 0, 3, 0)
	}
	w.installPaintEvent()
	w.SetOpacity(0)
	return w
}

// SetLightBackgroundColor sets the light-theme groove color.
func (w *ScrollBarGroove) SetLightBackgroundColor(color *qt.QColor) {
	w.lightBackgroundColor = cloneColor(color)
	w.Update()
}

// SetDarkBackgroundColor sets the dark-theme groove color.
func (w *ScrollBarGroove) SetDarkBackgroundColor(color *qt.QColor) {
	w.darkBackgroundColor = cloneColor(color)
	w.Update()
}

// FadeIn animates the groove opacity to 1 over 150ms.
func (w *ScrollBarGroove) FadeIn() { w.animateOpacity(1) }

// FadeOut animates the groove opacity to 0 over 150ms.
func (w *ScrollBarGroove) FadeOut() { w.animateOpacity(0) }

func (w *ScrollBarGroove) animateOpacity(to float64) {
	if w.opacityAni != nil {
		w.opacityAni.Stop()
		w.opacityAni.Delete()
		w.opacityAni = nil
	}
	from := w.opacity
	w.opacityAni = common.NewProgressAnimation(150, nil)
	w.opacityAni.OnProgress(func(t float64) {
		w.SetOpacity(from + (to-from)*t)
	})
	w.opacityAni.Start()
}

// SetOpacity sets the groove opacity and propagates it to the arrow buttons.
func (w *ScrollBarGroove) SetOpacity(opacity float64) {
	w.opacity = opacity
	w.upButton.SetOpacity(opacity)
	w.downButton.SetOpacity(opacity)
	w.Update()
	if w.onOpacityChanged != nil {
		w.onOpacityChanged(opacity)
	}
}

func (w *ScrollBarGroove) backgroundForTheme() *qt.QColor {
	if common.IsDarkTheme() {
		return w.darkBackgroundColor
	}
	return w.lightBackgroundColor
}

func (w *ScrollBarGroove) installPaintEvent() {
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
		super(event)
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing)
		painter.SetOpacity(w.opacity)
		painter.SetPenWithStyle(qt.NoPen)

		brush := qt.NewQBrush3(w.backgroundForTheme())
		defer brush.Delete()
		painter.SetBrush(brush)
		rect := w.Rect()

		painter.DrawRoundedRect3(rect, 6, 6)
		painter.End()
	})
}

// ScrollBarHandle is the draggable thumb of a fluent scroll bar.
type ScrollBarHandle struct {
	*qt.QWidget
	orientation qt.Orientation
	opacity     float64
	lightColor  *qt.QColor
	darkColor   *qt.QColor
	opacityAni  *common.ProgressAnimation
}

// NewScrollBarHandle builds a scroll bar handle for the given orientation.
func NewScrollBarHandle(orient qt.Orientation, parent *qt.QWidget) *ScrollBarHandle {
	w := &ScrollBarHandle{
		QWidget:     qt.NewQWidget(parent),
		orientation: orient,
		opacity:     1,
		lightColor:  qt.NewQColor11(0, 0, 0, 114),
		darkColor:   qt.NewQColor11(255, 255, 255, 139),
	}
	if orient == qt.Vertical {
		w.SetFixedWidth(3)
	} else {
		w.SetFixedHeight(3)
	}
	w.installPaintEvent()
	return w
}

// SetLightColor sets the light-theme handle color.
func (w *ScrollBarHandle) SetLightColor(color *qt.QColor) {
	w.lightColor = cloneColor(color)
	w.Update()
}

// SetDarkColor sets the dark-theme handle color.
func (w *ScrollBarHandle) SetDarkColor(color *qt.QColor) {
	w.darkColor = cloneColor(color)
	w.Update()
}

// FadeIn animates the handle opacity to 1 over 150ms.
func (w *ScrollBarHandle) FadeIn() { w.animateOpacity(1) }

// FadeOut animates the handle opacity to 0 over 150ms.
func (w *ScrollBarHandle) FadeOut() { w.animateOpacity(0) }

func (w *ScrollBarHandle) animateOpacity(to float64) {
	if w.opacityAni != nil {
		w.opacityAni.Stop()
		w.opacityAni.Delete()
		w.opacityAni = nil
	}
	from := w.opacity
	w.opacityAni = common.NewProgressAnimation(150, nil)
	w.opacityAni.OnProgress(func(t float64) {
		w.SetOpacity(from + (to-from)*t)
	})
	w.opacityAni.Start()
}

// SetOpacity sets the handle opacity.
func (w *ScrollBarHandle) SetOpacity(opacity float64) {
	w.opacity = opacity
	w.Update()
}

func (w *ScrollBarHandle) colorForTheme() *qt.QColor {
	if common.IsDarkTheme() {
		return w.darkColor
	}
	return w.lightColor
}

func (w *ScrollBarHandle) installPaintEvent() {
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
		super(event)
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing)
		painter.SetPenWithStyle(qt.NoPen)

		var r float64
		if w.orientation == qt.Vertical {
			r = float64(w.Width()) / 2
		} else {
			r = float64(w.Height()) / 2
		}
		painter.SetOpacity(w.opacity)
		brush := qt.NewQBrush3(w.colorForTheme())
		defer brush.Delete()
		painter.SetBrush(brush)
		rect := w.Rect()

		painter.DrawRoundedRect3(rect, r, r)
		painter.End()
	})
}

// ScrollBarHandleDisplayMode enumerates how the scroll bar handle is shown.
type ScrollBarHandleDisplayMode int

const (
	// ScrollBarHandleDisplayAlways always shows the handle.
	ScrollBarHandleDisplayAlways ScrollBarHandleDisplayMode = iota
	// ScrollBarHandleDisplayOnHover only shows the handle on hover.
	ScrollBarHandleDisplayOnHover
)

// ScrollBar is a fluent styled custom scroll bar. It is a plain QWidget that
// is decoupled from the QAbstractScrollArea native scroll bars while keeping a
// two-way value/range sync with the partner native scroll bar.
type ScrollBar struct {
	*qt.QWidget
	groove            *ScrollBarGroove
	handle            *ScrollBarHandle
	area              *qt.QAbstractScrollArea
	partnerBar        *qt.QScrollBar
	orientation       qt.Orientation
	singleStep        int
	pageStep          int
	padding           int
	minimum           int
	maximum           int
	value             int
	isPressed         bool
	isEnter           bool
	isExpanded        bool
	pressedX          int
	pressedY          int
	isForceHidden     bool
	handleDisplayMode ScrollBarHandleDisplayMode
	expandTimer       *qt.QTimer
	collapseTimer     *qt.QTimer
	onValueChanged    func(int)
	onRangeChanged    func(int, int)
	onSliderPressed   func()
	onSliderReleased  func()
	onSliderMoved     func()
}

// NewScrollBar builds a fluent scroll bar for the given orientation that is
// laid out over the edges of parent.
func NewScrollBar(orient qt.Orientation, parent *qt.QAbstractScrollArea) *ScrollBar {
	w := &ScrollBar{
		QWidget:           qt.NewQWidget(parent.QWidget),
		orientation:       orient,
		singleStep:        1,
		pageStep:          50,
		padding:           14,
		handleDisplayMode: ScrollBarHandleDisplayAlways,
		area:              parent,
	}
	w.groove = NewScrollBarGroove(orient, w.QWidget)
	w.handle = NewScrollBarHandle(orient, w.QWidget)
	w.groove.onOpacityChanged = w.onGrooveOpacityChanged
	w.expandTimer = qt.NewQTimer2(w.QObject)
	w.expandTimer.SetSingleShot(true)
	w.expandTimer.SetInterval(200)
	w.expandTimer.OnTimeout(w.Expand)
	w.collapseTimer = qt.NewQTimer2(w.QObject)
	w.collapseTimer.SetSingleShot(true)
	w.collapseTimer.SetInterval(200)
	w.collapseTimer.OnTimeout(w.Collapse)

	if orient == qt.Vertical {
		w.partnerBar = parent.VerticalScrollBar()
		parent.SetVerticalScrollBarPolicy(qt.ScrollBarAlwaysOff)
	} else {
		w.partnerBar = parent.HorizontalScrollBar()
		parent.SetHorizontalScrollBarPolicy(qt.ScrollBarAlwaysOff)
	}

	w.initWidget(parent)
	return w
}

func (w *ScrollBar) initWidget(parent *qt.QAbstractScrollArea) {
	w.groove.upButton.OnClicked(w.onPageUp)
	w.groove.downButton.OnClicked(w.onPageDown)
	w.onValueChanged = func(value int) {
		if w.partnerBar != nil {
			w.partnerBar.SetValue(value)
		}
	}
	if w.partnerBar != nil {
		w.partnerBar.OnValueChanged(func(value int) { w.setVal(value) })
		w.partnerBar.OnRangeChanged(func(min int, max int) { w.SetRange(min, max) })
	}

	parent.InstallEventFilter(w.QObject)
	w.OnEventFilter(func(super func(watched *qt.QObject, event *qt.QEvent) bool, watched *qt.QObject, event *qt.QEvent) bool {
		if event.Type() == qt.QEvent__Resize {
			sz := parent.Size()

			w.adjustPos(sz)
		}
		return super(watched, event)
	})

	w.SetRange(w.partnerBar.Minimum(), w.partnerBar.Maximum())
	w.SetVisible(w.maximum > 0 && !w.isForceHidden)
	sz := parent.Size()

	w.adjustPos(sz)

	w.installEvents()
}

func (w *ScrollBar) installEvents() {
	w.OnResizeEvent(func(super func(event *qt.QResizeEvent), event *qt.QResizeEvent) {
		super(event)
		w.groove.Resize(w.Width(), w.Height())
	})
	w.OnEnterEvent(func(super func(event *qt.QEvent), event *qt.QEvent) {
		super(event)
		w.isEnter = true
		w.expandTimer.Start(200)
	})
	w.OnLeaveEvent(func(super func(event *qt.QEvent), event *qt.QEvent) {
		super(event)
		w.isEnter = false
		w.collapseTimer.Start(200)
	})
	w.OnMousePressEvent(func(super func(event *qt.QMouseEvent), event *qt.QMouseEvent) {
		super(event)
		w.isPressed = true
		w.pressedX = event.X()
		w.pressedY = event.Y()

		child := w.ChildAt(event.X(), event.Y())
		isOnHandle := child != nil && child.UnsafePointer() == w.handle.QWidget.UnsafePointer()
		if isOnHandle || !w.isSlideRegion(event.X(), event.Y()) {
			return
		}

		var value int
		geom := w.handle.Geometry()
		if w.orientation == qt.Vertical {
			if event.Y() > geom.Bottom() {
				value = event.Y() - w.handle.Height() - w.padding
			} else {
				value = event.Y() - w.padding
			}
		} else {
			if event.X() > geom.Right() {
				value = event.X() - w.handle.Width() - w.padding
			} else {
				value = event.X() - w.padding
			}
		}
		// geom is borrowed (QWidget.Geometry const_cast) — do NOT Delete

		w.setVal(int(float64(value) / float64(max(w.slideLength(), 1)) * float64(w.maximum)))
		if w.onSliderPressed != nil {
			w.onSliderPressed()
		}
	})
	w.OnMouseReleaseEvent(func(super func(event *qt.QMouseEvent), event *qt.QMouseEvent) {
		super(event)
		w.isPressed = false
		if w.onSliderReleased != nil {
			w.onSliderReleased()
		}
	})
	w.OnMouseMoveEvent(func(super func(event *qt.QMouseEvent), event *qt.QMouseEvent) {
		super(event)
		var dv int
		if w.orientation == qt.Vertical {
			dv = event.Y() - w.pressedY
		} else {
			dv = event.X() - w.pressedX
		}
		dv = int(float64(dv) / float64(max(w.slideLength(), 1)) * float64(w.maximum-w.minimum))
		w.setVal(w.value + dv)
		w.pressedX = event.X()
		w.pressedY = event.Y()
		if w.onSliderMoved != nil {
			w.onSliderMoved()
		}
	})
	w.OnWheelEvent(func(super func(event *qt.QWheelEvent), event *qt.QWheelEvent) {
		if w.area != nil && w.area.Viewport() != nil {
			qt.QCoreApplication_SendEvent(w.area.Viewport().QObject, event.QEvent)
		}
	})
}

// onGrooveOpacityChanged stretches the handle as the groove fades in/out
// (port of Python's _onOpacityAniValueChanged): the handle grows from 3px to
// 6px in the cross axis as the groove expands.
func (w *ScrollBar) onGrooveOpacityChanged(opacity float64) {
	if w.orientation == qt.Vertical {
		w.handle.SetFixedWidth(int(3 + opacity*3))
	} else {
		w.handle.SetFixedHeight(int(3 + opacity*3))
	}
	w.adjustHandlePos()
}

func (w *ScrollBar) onPageUp() {
	w.setVal(w.value - w.pageStep)
}

func (w *ScrollBar) onPageDown() {
	w.setVal(w.value + w.pageStep)
}

// Value returns the current value.
func (w *ScrollBar) Value() int { return w.value }

// Minimum returns the minimum value.
func (w *ScrollBar) Minimum() int { return w.minimum }

// Maximum returns the maximum value.
func (w *ScrollBar) Maximum() int { return w.maximum }

// Orientation returns the scroll bar orientation.
func (w *ScrollBar) Orientation() qt.Orientation { return w.orientation }

// PageStep returns the page step.
func (w *ScrollBar) PageStep() int { return w.pageStep }

// SingleStep returns the single step.
func (w *ScrollBar) SingleStep() int { return w.singleStep }

// IsSliderDown reports whether the handle is currently pressed.
func (w *ScrollBar) IsSliderDown() bool { return w.isPressed }

// SetValue sets the scroll value.
func (w *ScrollBar) SetValue(value int) { w.setVal(value) }

func (w *ScrollBar) setVal(value int) {
	if value == w.value {
		return
	}
	if value < w.minimum {
		value = w.minimum
	}
	if value > w.maximum {
		value = w.maximum
	}
	w.value = value
	if w.onValueChanged != nil {
		w.onValueChanged(value)
	}
	w.adjustHandlePos()
}

// SetMinimum sets the minimum value.
func (w *ScrollBar) SetMinimum(min int) {
	if min == w.minimum {
		return
	}
	w.minimum = min
	if w.onRangeChanged != nil {
		w.onRangeChanged(min, w.maximum)
	}
}

// SetMaximum sets the maximum value.
func (w *ScrollBar) SetMaximum(max int) {
	if max == w.maximum {
		return
	}
	w.maximum = max
	if w.onRangeChanged != nil {
		w.onRangeChanged(w.minimum, max)
	}
}

// SetRange sets the value range.
func (w *ScrollBar) SetRange(min int, max int) {
	if min > max || (min == w.minimum && max == w.maximum) {
		return
	}
	w.SetMinimum(min)
	w.SetMaximum(max)
	w.adjustHandleSize()
	w.adjustHandlePos()
	w.SetVisible(max > 0 && !w.isForceHidden)
	if w.onRangeChanged != nil {
		w.onRangeChanged(min, max)
	}
}

// SetPageStep sets the page step.
func (w *ScrollBar) SetPageStep(step int) {
	if step >= 1 {
		w.pageStep = step
	}
}

// SetSingleStep sets the single step.
func (w *ScrollBar) SetSingleStep(step int) {
	if step >= 1 {
		w.singleStep = step
	}
}

// SetSliderDown updates the pressed state and emits the matching signal.
func (w *ScrollBar) SetSliderDown(isDown bool) {
	w.isPressed = true
	if isDown {
		if w.onSliderPressed != nil {
			w.onSliderPressed()
		}
	} else {
		if w.onSliderReleased != nil {
			w.onSliderReleased()
		}
	}
}

// SetHandleColor sets the handle color for light/dark themes.
func (w *ScrollBar) SetHandleColor(light, dark *qt.QColor) {
	w.handle.SetLightColor(light)
	w.handle.SetDarkColor(dark)
}

// SetArrowColor sets the arrow color for light/dark themes.
func (w *ScrollBar) SetArrowColor(light, dark *qt.QColor) {
	w.groove.upButton.SetLightColor(light)
	w.groove.upButton.SetDarkColor(dark)
	w.groove.downButton.SetLightColor(light)
	w.groove.downButton.SetDarkColor(dark)
}

// SetGrooveColor sets the groove color for light/dark themes.
func (w *ScrollBar) SetGrooveColor(light, dark *qt.QColor) {
	w.groove.SetLightBackgroundColor(light)
	w.groove.SetDarkBackgroundColor(dark)
}

// SetHandleDisplayMode sets the handle display mode.
func (w *ScrollBar) SetHandleDisplayMode(mode ScrollBarHandleDisplayMode) {
	if mode == w.handleDisplayMode {
		return
	}
	w.handleDisplayMode = mode
	if mode == ScrollBarHandleDisplayOnHover && !w.isEnter {
		w.handle.FadeOut()
	} else if mode == ScrollBarHandleDisplayAlways {
		w.handle.FadeIn()
	}
}

// Expand expands the scroll bar (shows the groove).
func (w *ScrollBar) Expand() {
	if w.isExpanded || !w.isEnter {
		return
	}
	w.isExpanded = true
	w.groove.FadeIn()
	w.handle.FadeIn()
}

// Collapse collapses the scroll bar (hides the groove).
func (w *ScrollBar) Collapse() {
	if !w.isExpanded || w.isEnter {
		return
	}
	w.isExpanded = false
	w.groove.FadeOut()
	if w.handleDisplayMode == ScrollBarHandleDisplayOnHover {
		w.handle.FadeOut()
	}
}

// SetForceHidden forces the scroll bar to be hidden regardless of its range.
func (w *ScrollBar) SetForceHidden(isHidden bool) {
	w.isForceHidden = isHidden
	w.SetVisible(w.maximum > 0 && !isHidden)
}

// OnValueChanged registers a callback for value changes.
func (w *ScrollBar) OnValueChanged(f func(int)) { w.onValueChanged = f }

// OnRangeChanged registers a callback for range changes.
func (w *ScrollBar) OnRangeChanged(f func(int, int)) { w.onRangeChanged = f }

// OnSliderPressed registers a callback for handle presses.
func (w *ScrollBar) OnSliderPressed(f func()) { w.onSliderPressed = f }

// OnSliderReleased registers a callback for handle releases.
func (w *ScrollBar) OnSliderReleased(f func()) { w.onSliderReleased = f }

// OnSliderMoved registers a callback for handle movement.
func (w *ScrollBar) OnSliderMoved(f func()) { w.onSliderMoved = f }

func (w *ScrollBar) adjustPos(size *qt.QSize) {
	if w.orientation == qt.Vertical {
		w.Resize(12, size.Height()-2)
		w.Move(size.Width()-13, 1)
	} else {
		w.Resize(size.Width()-2, 12)
		w.Move(1, size.Height()-13)
	}
}

func (w *ScrollBar) adjustHandleSize() {
	if w.orientation == qt.Vertical {
		total := w.maximum - w.minimum + w.area.Height()
		s := int(float64(w.grooveLength()*w.area.Height()) / float64(max(total, 1)))
		w.handle.SetFixedHeight(max(30, s))
	} else {
		total := w.maximum - w.minimum + w.area.Width()
		s := int(float64(w.grooveLength()*w.area.Width()) / float64(max(total, 1)))
		w.handle.SetFixedWidth(max(30, s))
	}
}

func (w *ScrollBar) adjustHandlePos() {
	total := max(w.maximum-w.minimum, 1)
	delta := int(float64(w.value) / float64(total) * float64(w.slideLength()))
	if w.orientation == qt.Vertical {
		x := w.Width() - w.handle.Width() - 3
		w.handle.Move(x, w.padding+delta)
	} else {
		y := w.Height() - w.handle.Height() - 3
		w.handle.Move(w.padding+delta, y)
	}
}

func (w *ScrollBar) grooveLength() int {
	if w.orientation == qt.Vertical {
		return w.Height() - 2*w.padding
	}
	return w.Width() - 2*w.padding
}

func (w *ScrollBar) slideLength() int {
	if w.orientation == qt.Vertical {
		return w.grooveLength() - w.handle.Height()
	}
	return w.grooveLength() - w.handle.Width()
}

func (w *ScrollBar) isSlideRegion(x, y int) bool {
	if w.orientation == qt.Vertical {
		return w.padding <= y && y <= w.Height()-w.padding
	}
	return w.padding <= x && x <= w.Width()-w.padding
}

// SmoothScrollBar is a scroll bar whose value changes are animated with an
// OutCubic progress driver (port of Python's QPropertyAnimation on "val").
type SmoothScrollBar struct {
	*ScrollBar
	duration  int
	scrollVal int
	ani       *common.ProgressAnimation
}

// NewSmoothScrollBar builds a smooth scroll bar for the given orientation.
func NewSmoothScrollBar(orient qt.Orientation, parent *qt.QAbstractScrollArea) *SmoothScrollBar {
	b := &SmoothScrollBar{ScrollBar: NewScrollBar(orient, parent), duration: 500}
	b.scrollVal = b.Value()
	return b
}

// SetValue animates the scroll value from the current value to value.
func (b *SmoothScrollBar) SetValue(value int) {
	if value == b.Value() {
		return
	}
	if b.ani != nil {
		b.ani.Stop()
		b.ani.Delete()
		b.ani = nil
	}

	dv := value - b.Value()
	if dv < 0 {
		dv = -dv
	}
	duration := b.duration
	if dv < 50 {
		duration = int(b.duration * dv / 70)
	}
	if duration < 30 {
		duration = 30
	}

	from := b.Value()
	curve := qt.NewQEasingCurve3(qt.QEasingCurve__OutCubic)
	b.ani = common.NewProgressAnimation(duration, curve)
	curve.Delete()
	b.ani.OnProgress(func(t float64) {
		b.setVal(from + int(math.Round(float64(value-from)*t)))
	})
	b.ani.Start()
}

// ScrollValue scrolls by the given distance.
func (b *SmoothScrollBar) ScrollValue(value int) {
	b.scrollVal += value
	if b.scrollVal < b.Minimum() {
		b.scrollVal = b.Minimum()
	}
	if b.scrollVal > b.Maximum() {
		b.scrollVal = b.Maximum()
	}
	b.SetValue(b.scrollVal)
}

// ScrollTo scrolls to the given position.
func (b *SmoothScrollBar) ScrollTo(value int) {
	b.scrollVal = value
	if b.scrollVal < b.Minimum() {
		b.scrollVal = b.Minimum()
	}
	if b.scrollVal > b.Maximum() {
		b.scrollVal = b.Maximum()
	}
	b.SetValue(b.scrollVal)
}

// ResetValue resets the internal scroll position.
func (b *SmoothScrollBar) ResetValue(value int) { b.scrollVal = value }

// SetScrollAnimation stores the scroll animation duration (no-op without
// animation, kept for API parity).
func (b *SmoothScrollBar) SetScrollAnimation(duration int) { b.duration = duration }

// SmoothScrollDelegate wraps a QAbstractScrollArea: it hides the native scroll
// bars, creates fluent SmoothScrollBars over them and forwards wheel events.
type SmoothScrollDelegate struct {
	*qt.QObject
	area                   *qt.QAbstractScrollArea
	useAni                 bool
	vScrollBar             *SmoothScrollBar
	hScrollBar             *SmoothScrollBar
	verticalSmoothScroll   *common.SmoothScroll
	horizontalSmoothScroll *common.SmoothScroll
	verticalSmoothMode     common.SmoothMode
	horizontalSmoothMode   common.SmoothMode
}

// NewSmoothScrollDelegate builds a smooth scroll delegate for parent.
func NewSmoothScrollDelegate(parent *qt.QAbstractScrollArea, useAni bool) *SmoothScrollDelegate {
	d := &SmoothScrollDelegate{
		QObject: qt.NewQObject2(parent.QObject),
		area:    parent,
		useAni:  useAni,
	}
	d.vScrollBar = NewSmoothScrollBar(qt.Vertical, parent)
	d.hScrollBar = NewSmoothScrollBar(qt.Horizontal, parent)
	d.verticalSmoothScroll = common.NewSmoothScroll(parent, qt.Vertical)
	d.horizontalSmoothScroll = common.NewSmoothScroll(parent, qt.Horizontal)

	// QAbstractItemView defaults to ScrollPerItem, where the native scroll bar
	// range is measured in items rather than pixels; adding a wheel delta of
	// ±120 to that range jumps straight to the top/bottom. Switch to pixel-based
	// scrolling like the Python SmoothScrollDelegate before installing the
	// wheel filter.
	if parent.QObject.Inherits("QAbstractItemView") {
		view := qt.UnsafeNewQAbstractItemView(parent.UnsafePointer())
		view.SetVerticalScrollMode(qt.QAbstractItemView__ScrollPerPixel)
		view.SetHorizontalScrollMode(qt.QAbstractItemView__ScrollPerPixel)
		if parent.QObject.Inherits("QListView") {
			// Keep the native horizontal scroll bar hidden (a 0-height QSS),
			// matching the reference; the fluent horizontal overlay bar remains.
			view.SetHorizontalScrollBarPolicy(qt.ScrollBarAlwaysOn)
			view.HorizontalScrollBar().SetStyleSheet("QScrollBar:horizontal{height: 0px}")
		}
	}

	// The viewport is owned by Qt (not a directly-constructed miqt type), so its
	// virtual methods cannot be overridden. Instead install `d` as an event
	// filter and translate the wheel QEvent into a QWheelEvent here.
	if viewport := parent.Viewport(); viewport != nil {
		viewport.InstallEventFilter(d.QObject)
	}
	d.OnEventFilter(func(super func(watched *qt.QObject, event *qt.QEvent) bool, watched *qt.QObject, event *qt.QEvent) bool {
		if event.Type() == qt.QEvent__Wheel {
			return d.handleWheel(qt.UnsafeNewQWheelEvent(event.UnsafePointer()))
		}
		return super(watched, event)
	})
	return d
}

// VerticalSmoothScrollBar returns the vertical fluent scroll bar.
func (d *SmoothScrollDelegate) VerticalSmoothScrollBar() *SmoothScrollBar { return d.vScrollBar }

// HorizontalSmoothScrollBar returns the horizontal fluent scroll bar.
func (d *SmoothScrollDelegate) HorizontalSmoothScrollBar() *SmoothScrollBar { return d.hScrollBar }

// SetSmoothMode stores the smooth scroll mode for the given orientation and
// applies it to the corresponding smooth-scroll engine.
func (d *SmoothScrollDelegate) SetSmoothMode(mode common.SmoothMode, orient qt.Orientation) {
	if orient == qt.Vertical {
		d.verticalSmoothMode = mode
		d.verticalSmoothScroll.SetSmoothMode(mode)
	} else {
		d.horizontalSmoothMode = mode
		d.horizontalSmoothScroll.SetSmoothMode(mode)
	}
}

// handleWheel forwards a viewport wheel event to the correct engine. It returns
// false when the event should be propagated to the native handler (at the end
// of the range, or for touchpad deltas the smooth engine declines).
func (d *SmoothScrollDelegate) handleWheel(event *qt.QWheelEvent) bool {
	ad := event.AngleDelta()
	dy := ad.Y()
	dx := ad.X()

	if dy != 0 {
		if (dy < 0 && d.vScrollBar.Value() == d.vScrollBar.Maximum()) ||
			(dy > 0 && d.vScrollBar.Value() == d.vScrollBar.Minimum()) {
			return false
		}
		if d.useAni {
			d.vScrollBar.ScrollValue(-dy)
			event.SetAccepted(true)
			return true
		}
		if d.verticalSmoothScroll.HandleWheelEvent(event) {
			event.SetAccepted(true)
			return true
		}
		return false
	}

	if dx != 0 {
		if (dx < 0 && d.hScrollBar.Value() == d.hScrollBar.Maximum()) ||
			(dx > 0 && d.hScrollBar.Value() == d.hScrollBar.Minimum()) {
			return false
		}
		if d.useAni {
			d.hScrollBar.ScrollValue(-dx)
			event.SetAccepted(true)
			return true
		}
		if d.horizontalSmoothScroll.HandleWheelEvent(event) {
			event.SetAccepted(true)
			return true
		}
		return false
	}

	return false
}
