package widgets

import (
	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// ProgressRing is a circular progress indicator.
//
// Constructors
//   - NewProgressRing(parent *qt.QWidget, useAni bool)
type ProgressRing struct {
	*ProgressBar
	strokeWidth int
}

// NewProgressRing builds a progress ring.
func NewProgressRing(parent *qt.QWidget, useAni bool) *ProgressRing {
	w := &ProgressRing{ProgressBar: NewProgressBar(parent, useAni)}
	w.lightBackgroundColor = qt.NewQColor11(0, 0, 0, 34)
	w.darkBackgroundColor = qt.NewQColor11(255, 255, 255, 34)
	w.strokeWidth = 6
	w.SetTextVisible(false)
	w.SetFixedSize2(100, 100)
	common.SetFont(w.QWidget, 14, 400)
	w.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		w.paint()
	})
	return w
}

// StrokeWidth returns the ring stroke width.
func (w *ProgressRing) StrokeWidth() int { return w.strokeWidth }

// SetStrokeWidth sets the ring stroke width.
func (w *ProgressRing) SetStrokeWidth(width int) {
	w.strokeWidth = width
	w.Update()
}

func (w *ProgressRing) drawText(painter *qt.QPainter, text string) {
	font := w.Font() // borrowed reference — do NOT Delete
	painter.SetFont(font)
	if common.IsDarkTheme() {
		painter.SetPen(qt.NewQColor3(255, 255, 255))
	} else {
		painter.SetPen(qt.NewQColor3(0, 0, 0))
	}
	rect := w.Rect()

	painter.DrawText6(rect, int(qt.AlignCenter), text)
}

func (w *ProgressRing) paint() {
	painter := qt.NewQPainter2(w.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__Antialiasing)

	cw := float64(w.strokeWidth)
	side := w.Height()
	if w.Width() < side {
		side = w.Width()
	}
	wd := float64(side) - cw
	rc := qt.NewQRectF4(cw/2, float64(w.Height())/2-wd/2, wd, wd)
	defer rc.Delete()

	var bc *qt.QColor
	if common.IsDarkTheme() {
		bc = w.darkBackgroundColor
	} else {
		bc = w.lightBackgroundColor
	}
	bgBrush := qt.NewQBrush3(bc)
	pen := qt.NewQPen8(bgBrush, cw, qt.SolidLine, qt.RoundCap, qt.RoundJoin)
	bgBrush.Delete()
	defer pen.Delete()
	painter.SetPenWithPen(pen)
	painter.DrawArc(rc, 0, 360*16)

	if w.Maximum() <= w.Minimum() {
		painter.End()
		return
	}

	bar := w.BarColor() // borrowed (may return the stored field color) — do NOT Delete
	pen.SetColor(bar)
	painter.SetPenWithPen(pen)
	degree := int(w.val / float64(w.Maximum()-w.Minimum()) * 360)
	painter.DrawArc(rc, 90*16, -degree*16)

	if w.IsTextVisible() {
		w.drawText(painter, w.ValText())
	}
	painter.End()
}

// IndeterminateProgressRing is a spinning indeterminate progress ring. The
// start-angle/span-angle QPropertyAnimations are approximated with a QTimer.
//
// Constructors
//   - NewIndeterminateProgressRing(parent *qt.QWidget, start bool)
type IndeterminateProgressRing struct {
	*qt.QProgressBar
	lightBackgroundColor *qt.QColor
	darkBackgroundColor  *qt.QColor
	lightBarColor        *qt.QColor
	darkBarColor         *qt.QColor
	strokeWidth          int
	startAngle           int
	spanAngle            int
	elapsed              int
	timer                *qt.QTimer
}

// NewIndeterminateProgressRing builds an indeterminate progress ring.
func NewIndeterminateProgressRing(parent *qt.QWidget, start bool) *IndeterminateProgressRing {
	w := &IndeterminateProgressRing{QProgressBar: qt.NewQProgressBar(parent)}
	w.lightBackgroundColor = qt.NewQColor11(0, 0, 0, 0)
	w.darkBackgroundColor = qt.NewQColor11(255, 255, 255, 0)
	w.lightBarColor = qt.NewQColor()
	w.darkBarColor = qt.NewQColor()
	w.strokeWidth = 6
	w.SetFixedSize2(80, 80)

	w.timer = qt.NewQTimer2(w.QObject)
	w.timer.SetInterval(indeterminateTimerInterval)
	w.timer.OnTimeout(func() {
		w.elapsed += indeterminateTimerInterval
		w.advance()
		w.Update()
	})

	w.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		w.paint()
	})

	if start {
		w.Start()
	}
	return w
}

func (w *IndeterminateProgressRing) advance() {
	t := w.elapsed % 2000
	if t < 1000 {
		w.startAngle = 450 * t / 1000
		w.spanAngle = 180 * t / 1000
	} else {
		t2 := t - 1000
		w.startAngle = 450 + 630*t2/1000
		w.spanAngle = 180 - 180*t2/1000
	}
}

// StrokeWidth returns the ring stroke width.
func (w *IndeterminateProgressRing) StrokeWidth() int { return w.strokeWidth }

// SetStrokeWidth sets the ring stroke width.
func (w *IndeterminateProgressRing) SetStrokeWidth(width int) {
	w.strokeWidth = width
	w.Update()
}

// Start starts the spin animation.
func (w *IndeterminateProgressRing) Start() {
	w.startAngle = 0
	w.spanAngle = 0
	w.elapsed = 0
	w.timer.Start(indeterminateTimerInterval)
}

// Stop stops the spin animation.
func (w *IndeterminateProgressRing) Stop() {
	w.timer.Stop()
	w.startAngle = 0
	w.spanAngle = 0
	w.Update()
}

// LightBarColor returns the light-theme bar color.
func (w *IndeterminateProgressRing) LightBarColor() *qt.QColor {
	if w.lightBarColor.IsValid() {
		return w.lightBarColor
	}
	return common.ThemeColorPrimary.Color()
}

// DarkBarColor returns the dark-theme bar color.
func (w *IndeterminateProgressRing) DarkBarColor() *qt.QColor {
	if w.darkBarColor.IsValid() {
		return w.darkBarColor
	}
	return common.ThemeColorPrimary.Color()
}

// SetCustomBarColor sets the light/dark bar colors.
func (w *IndeterminateProgressRing) SetCustomBarColor(light, dark *qt.QColor) {
	w.lightBarColor = cloneColor(light)
	w.darkBarColor = cloneColor(dark)
	w.Update()
}

// SetCustomBackgroundColor sets the light/dark background colors.
func (w *IndeterminateProgressRing) SetCustomBackgroundColor(light, dark *qt.QColor) {
	w.lightBackgroundColor = cloneColor(light)
	w.darkBackgroundColor = cloneColor(dark)
	w.Update()
}

func (w *IndeterminateProgressRing) paint() {
	painter := qt.NewQPainter2(w.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__Antialiasing)

	cw := float64(w.strokeWidth)
	side := w.Height()
	if w.Width() < side {
		side = w.Width()
	}
	wd := float64(side) - cw
	rc := qt.NewQRectF4(cw/2, float64(w.Height())/2-wd/2, wd, wd)
	defer rc.Delete()

	var bc *qt.QColor
	if common.IsDarkTheme() {
		bc = w.darkBackgroundColor
	} else {
		bc = w.lightBackgroundColor
	}
	bgBrush := qt.NewQBrush3(bc)
	pen := qt.NewQPen8(bgBrush, cw, qt.SolidLine, qt.RoundCap, qt.RoundJoin)
	bgBrush.Delete()
	defer pen.Delete()
	painter.SetPenWithPen(pen)
	painter.DrawArc(rc, 0, 360*16)

	bar := w.LightBarColor() // borrowed (may return the stored field color) — do NOT Delete
	if common.IsDarkTheme() {
		bar = w.DarkBarColor()
	}
	pen.SetColor(bar)
	painter.SetPenWithPen(pen)

	startAngle := -w.startAngle + 180
	painter.DrawArc(rc, (startAngle%360)*16, -w.spanAngle*16)
	painter.End()
}
