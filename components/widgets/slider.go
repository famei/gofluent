package widgets

import (
	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// SliderHandle is the round handle drawn inside a Slider.
//
// The Python implementation animates the handle radius through a
// QPropertyAnimation on a custom "radius" property; that custom meta-object
// property is not expressible in Go, so the radius is driven by a frame-based
// ProgressAnimation instead.
type SliderHandle struct {
	*qt.QWidget
	radius           float64
	lightHandleColor *qt.QColor
	darkHandleColor  *qt.QColor
	radiusAni        *common.ProgressAnimation
	onPressed        func()
	onReleased       func()
}

// NewSliderHandle builds a slider handle.
func NewSliderHandle(parent *qt.QSlider) *SliderHandle {
	w := &SliderHandle{QWidget: qt.NewQWidget(parent.QWidget)}
	w.SetFixedSize2(22, 22)
	w.radius = 5
	w.lightHandleColor = qt.NewQColor()
	w.darkHandleColor = qt.NewQColor()
	w.installEvents()
	return w
}

// SetHandleColor sets the inner-circle color for light/dark themes.
func (w *SliderHandle) SetHandleColor(light, dark *qt.QColor) {
	w.lightHandleColor = cloneColor(light)
	w.darkHandleColor = cloneColor(dark)
	w.Update()
}

// OnPressed registers the press callback.
func (w *SliderHandle) OnPressed(f func()) { w.onPressed = f }

// OnReleased registers the release callback.
func (w *SliderHandle) OnReleased(f func()) { w.onReleased = f }

// startRadiusAni animates the handle radius toward target over 100ms (linear
// easing, matching Python's radiusAni).
func (w *SliderHandle) startRadiusAni(target float64) {
	w.stopRadiusAni()
	from := w.radius
	w.radiusAni = common.AnimateFloat(from, target, 100, nil, func(v float64) {
		w.radius = v
		w.Update()
	})
}

// stopRadiusAni interrupts and releases any in-flight radius animation.
func (w *SliderHandle) stopRadiusAni() {
	if w.radiusAni != nil {
		w.radiusAni.Stop()
		w.radiusAni.Delete()
		w.radiusAni = nil
	}
}

func (w *SliderHandle) installEvents() {
	w.OnEnterEvent(func(super func(e *qt.QEvent), e *qt.QEvent) {
		w.startRadiusAni(6.5)
	})
	w.OnLeaveEvent(func(super func(e *qt.QEvent), e *qt.QEvent) {
		w.startRadiusAni(5)
	})
	w.OnMousePressEvent(func(super func(e *qt.QMouseEvent), e *qt.QMouseEvent) {
		w.startRadiusAni(4)
		if w.onPressed != nil {
			w.onPressed()
		}
	})
	w.OnMouseReleaseEvent(func(super func(e *qt.QMouseEvent), e *qt.QMouseEvent) {
		w.startRadiusAni(6.5)
		if w.onReleased != nil {
			w.onReleased()
		}
	})
	w.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing)
		painter.SetPenWithStyle(qt.NoPen)

		isDark := common.IsDarkTheme()
		var fillBrush *qt.QBrush
		if isDark {
			painter.SetPen(qt.NewQColor11(0, 0, 0, 90))
			fillBrush = qt.NewQBrush3(qt.NewQColor3(69, 69, 69))
		} else {
			painter.SetPen(qt.NewQColor11(0, 0, 0, 25))
			fillBrush = qt.NewQBrush3(qt.NewQColor3(255, 255, 255))
		}
		painter.SetBrush(fillBrush)
		fillBrush.Delete()
		adjusted := w.Rect().Adjusted(1, 1, -1, -1) // GoGC-armed — do NOT Delete
		outer := qt.NewQRectF5(adjusted)
		defer outer.Delete()
		painter.DrawEllipse(outer)

		innerColor := common.AutoFallbackThemeColor(w.lightHandleColor, w.darkHandleColor)
		innerBrush := qt.NewQBrush3(innerColor)
		painter.SetBrush(innerBrush)
		innerBrush.Delete()
		painter.DrawEllipse3(qt.NewQPointF3(11, 11), w.radius, w.radius)
		painter.End()
	})
}

// Slider is a fluent-styled slider that jumps to the clicked position.
//
// Constructors
//   - NewSlider(parent *qt.QWidget)
//   - NewSliderOrientation(orientation qt.Orientation, parent *qt.QWidget)
type Slider struct {
	*qt.QSlider
	handle           *SliderHandle
	lightGrooveColor *qt.QColor
	darkGrooveColor  *qt.QColor
	onClicked        func(value int)
}

// NewSlider builds a horizontal slider.
func NewSlider(parent *qt.QWidget) *Slider {
	return newSliderBase(qt.Horizontal, parent)
}

// NewSliderOrientation builds a slider with the given orientation.
func NewSliderOrientation(orientation qt.Orientation, parent *qt.QWidget) *Slider {
	return newSliderBase(orientation, parent)
}

func newSliderBase(orientation qt.Orientation, parent *qt.QWidget) *Slider {
	var w *Slider
	if parent != nil {
		w = &Slider{QSlider: qt.NewQSlider4(orientation, parent)}
	} else {
		w = &Slider{QSlider: qt.NewQSlider3(orientation)}
	}
	w.handle = NewSliderHandle(w.QSlider)
	w.lightGrooveColor = qt.NewQColor()
	w.darkGrooveColor = qt.NewQColor()
	w.setOrientation(orientation)

	w.OnValueChanged(func(value int) {
		w.adjustHandlePos()
	})
	w.OnMousePressEvent(func(super func(e *qt.QMouseEvent), e *qt.QMouseEvent) {
		w.SetValue(w.posToValue(e.Pos()))
		if w.onClicked != nil {
			w.onClicked(w.Value())
		}
	})
	w.OnMouseMoveEvent(func(super func(e *qt.QMouseEvent), e *qt.QMouseEvent) {
		w.SetValue(w.posToValue(e.Pos()))
	})
	w.OnResizeEvent(func(super func(e *qt.QResizeEvent), e *qt.QResizeEvent) {
		w.adjustHandlePos()
	})
	w.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		w.paint()
	})
	w.adjustHandlePos()
	return w
}

// OnClicked registers the clicked callback.
func (w *Slider) OnClicked(f func(value int)) { w.onClicked = f }

// SetThemeColor sets the groove and handle color for light/dark themes.
func (w *Slider) SetThemeColor(light, dark *qt.QColor) {
	w.lightGrooveColor = cloneColor(light)
	w.darkGrooveColor = cloneColor(dark)
	w.handle.SetHandleColor(light, dark)
	w.Update()
}

// SetOrientation sets the orientation and adjusts the minimum size.
func (w *Slider) SetOrientation(orientation qt.Orientation) {
	w.setOrientation(orientation)
}

func (w *Slider) setOrientation(orientation qt.Orientation) {
	w.QAbstractSlider.SetOrientation(orientation)
	if orientation == qt.Horizontal {
		w.SetMinimumHeight(22)
	} else {
		w.SetMinimumWidth(22)
	}
}

func (w *Slider) grooveLength() int {
	var l int
	if w.Orientation() == qt.Horizontal {
		l = w.Width()
	} else {
		l = w.Height()
	}
	return l - w.handle.Width()
}

func (w *Slider) adjustHandlePos() {
	total := w.Maximum() - w.Minimum()
	if total < 1 {
		total = 1
	}
	delta := int(float64(w.Value()-w.Minimum()) / float64(total) * float64(w.grooveLength()))
	if w.Orientation() == qt.Vertical {
		w.handle.Move(0, delta)
	} else {
		w.handle.Move(delta, 0)
	}
}

func (w *Slider) posToValue(pos *qt.QPoint) int {
	// pos is GoGC-armed (QMouseEvent.Pos) — do NOT Delete
	pd := w.handle.Width() / 2
	gs := w.grooveLength()
	if gs < 1 {
		gs = 1
	}
	var v int
	if w.Orientation() == qt.Horizontal {
		v = pos.X()
	} else {
		v = pos.Y()
	}
	return int(float64(v-pd)/float64(gs)*float64(w.Maximum()-w.Minimum())) + w.Minimum()
}

func (w *Slider) paint() {
	painter := qt.NewQPainter2(w.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__Antialiasing)
	painter.SetPenWithStyle(qt.NoPen)
	var grooveBrush *qt.QBrush
	if common.IsDarkTheme() {
		grooveBrush = qt.NewQBrush3(qt.NewQColor11(255, 255, 255, 115))
	} else {
		grooveBrush = qt.NewQBrush3(qt.NewQColor11(0, 0, 0, 100))
	}
	painter.SetBrush(grooveBrush)
	grooveBrush.Delete()
	if w.Orientation() == qt.Horizontal {
		w.drawHorizonGroove(painter)
	} else {
		w.drawVerticalGroove(painter)
	}
	painter.End()
}

func (w *Slider) drawHorizonGroove(painter *qt.QPainter) {
	width := float64(w.Width())
	r := float64(w.handle.Width()) / 2
	groove := qt.NewQRectF4(r, r-2, width-r*2, 4)
	defer groove.Delete()
	painter.DrawRoundedRect(groove, 2, 2)

	if w.Maximum()-w.Minimum() == 0 {
		return
	}
	active := qt.NewQBrush3(common.AutoFallbackThemeColor(w.lightGrooveColor, w.darkGrooveColor))
	painter.SetBrush(active)
	active.Delete()
	aw := float64(w.Value()-w.Minimum()) / float64(w.Maximum()-w.Minimum()) * (width - r*2)
	filled := qt.NewQRectF4(r, r-2, aw, 4)
	defer filled.Delete()
	painter.DrawRoundedRect(filled, 2, 2)
}

func (w *Slider) drawVerticalGroove(painter *qt.QPainter) {
	height := float64(w.Height())
	r := float64(w.handle.Width()) / 2
	groove := qt.NewQRectF4(r-2, r, 4, height-2*r)
	defer groove.Delete()
	painter.DrawRoundedRect(groove, 2, 2)

	if w.Maximum()-w.Minimum() == 0 {
		return
	}
	active := qt.NewQBrush3(common.AutoFallbackThemeColor(w.lightGrooveColor, w.darkGrooveColor))
	painter.SetBrush(active)
	active.Delete()
	ah := float64(w.Value()-w.Minimum()) / float64(w.Maximum()-w.Minimum()) * (height - r*2)
	filled := qt.NewQRectF4(r-2, r, 4, ah)
	defer filled.Delete()
	painter.DrawRoundedRect(filled, 2, 2)
}

// ClickableSlider is a slider that jumps to the clicked position while keeping
// the native mouse handling.
type ClickableSlider struct {
	*qt.QSlider
	onClicked func(value int)
}

// NewClickableSlider builds a clickable slider.
func NewClickableSlider(parent *qt.QWidget) *ClickableSlider {
	w := &ClickableSlider{QSlider: qt.NewQSlider(parent)}
	w.OnMousePressEvent(func(super func(e *qt.QMouseEvent), e *qt.QMouseEvent) {
		super(e)
		var value int
		if w.Orientation() == qt.Horizontal {
			value = e.Pos().X() * w.Maximum() / w.Width()
		} else {
			value = (w.Height() - e.Pos().Y()) * w.Maximum() / w.Height()
		}
		w.SetValue(value)
		if w.onClicked != nil {
			w.onClicked(w.Value())
		}
	})
	return w
}

// OnClicked registers the clicked callback.
func (w *ClickableSlider) OnClicked(f func(value int)) { w.onClicked = f }
