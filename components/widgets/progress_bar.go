package widgets

import (
	"strconv"
	"strings"

	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// ProgressBar is a fluent-styled progress bar. The animated "val" custom
// property is applied directly (the QPropertyAnimation on a custom property is
// not expressible in Go).
//
// Constructors
//   - NewProgressBar(parent *qt.QWidget, useAni bool)
type ProgressBar struct {
	*qt.QProgressBar
	val                  float64
	useAni               bool
	lightBackgroundColor *qt.QColor
	darkBackgroundColor  *qt.QColor
	lightBarColor        *qt.QColor
	darkBarColor         *qt.QColor
	isPaused             bool
	isError              bool
}

// NewProgressBar builds a progress bar.
func NewProgressBar(parent *qt.QWidget, useAni bool) *ProgressBar {
	w := &ProgressBar{QProgressBar: qt.NewQProgressBar(parent)}
	w.SetFixedHeight(4)
	w.useAni = useAni
	w.lightBackgroundColor = qt.NewQColor11(0, 0, 0, 155)
	w.darkBackgroundColor = qt.NewQColor11(255, 255, 255, 155)
	w.lightBarColor = qt.NewQColor()
	w.darkBarColor = qt.NewQColor()
	w.OnValueChanged(func(value int) {
		w.val = float64(value)
	})
	w.SetValue(0)
	w.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		w.paint()
	})
	return w
}

// Val returns the current animated value.
func (w *ProgressBar) Val() float64 { return w.val }

// SetVal sets the current animated value.
func (w *ProgressBar) SetVal(v float64) {
	w.val = v
	w.Update()
}

// IsUseAni reports whether animation is enabled.
func (w *ProgressBar) IsUseAni() bool { return w.useAni }

// SetUseAni enables or disables animation.
func (w *ProgressBar) SetUseAni(isUse bool) { w.useAni = isUse }

// LightBarColor returns the light-theme bar color.
func (w *ProgressBar) LightBarColor() *qt.QColor {
	if w.lightBarColor.IsValid() {
		return w.lightBarColor
	}
	return common.ThemeColorPrimary.Color()
}

// DarkBarColor returns the dark-theme bar color.
func (w *ProgressBar) DarkBarColor() *qt.QColor {
	if w.darkBarColor.IsValid() {
		return w.darkBarColor
	}
	return common.ThemeColorPrimary.Color()
}

// SetCustomBarColor sets the light/dark bar colors.
func (w *ProgressBar) SetCustomBarColor(light, dark *qt.QColor) {
	w.lightBarColor = cloneColor(light)
	w.darkBarColor = cloneColor(dark)
	w.Update()
}

// SetCustomBackgroundColor sets the light/dark background colors.
func (w *ProgressBar) SetCustomBackgroundColor(light, dark *qt.QColor) {
	w.lightBackgroundColor = cloneColor(light)
	w.darkBackgroundColor = cloneColor(dark)
	w.Update()
}

// Resume clears the paused/error state.
func (w *ProgressBar) Resume() {
	w.isPaused = false
	w.isError = false
	w.Update()
}

// Pause sets the paused state.
func (w *ProgressBar) Pause() {
	w.isPaused = true
	w.Update()
}

// SetPaused sets the paused state.
func (w *ProgressBar) SetPaused(isPaused bool) {
	w.isPaused = isPaused
	w.Update()
}

// IsPaused reports the paused state.
func (w *ProgressBar) IsPaused() bool { return w.isPaused }

// Error sets the error state.
func (w *ProgressBar) Error() {
	w.isError = true
	w.Update()
}

// SetError sets the error state.
func (w *ProgressBar) SetError(isError bool) {
	w.isError = isError
	if isError {
		w.Error()
	} else {
		w.Resume()
	}
}

// IsError reports the error state.
func (w *ProgressBar) IsError() bool { return w.isError }

// BarColor returns the theme/state-appropriate bar color.
func (w *ProgressBar) BarColor() *qt.QColor {
	if w.IsPaused() {
		if common.IsDarkTheme() {
			return qt.NewQColor3(252, 225, 0)
		}
		return qt.NewQColor3(157, 93, 0)
	}
	if w.IsError() {
		if common.IsDarkTheme() {
			return qt.NewQColor3(255, 153, 164)
		}
		return qt.NewQColor3(196, 43, 28)
	}
	if common.IsDarkTheme() {
		return w.DarkBarColor()
	}
	return w.LightBarColor()
}

// ValText renders the progress format string with the current value.
func (w *ProgressBar) ValText() string {
	if w.Maximum() <= w.Minimum() {
		return ""
	}
	total := w.Maximum() - w.Minimum()
	result := w.Format()
	result = strings.ReplaceAll(result, "%m", strconv.Itoa(total))
	result = strings.ReplaceAll(result, "%v", strconv.Itoa(int(w.val)))
	if total == 0 {
		return strings.ReplaceAll(result, "%p", "100")
	}
	progress := int((w.val - float64(w.Minimum())) * 100 / float64(total))
	return strings.ReplaceAll(result, "%p", strconv.Itoa(progress))
}

func (w *ProgressBar) paint() {
	painter := qt.NewQPainter2(w.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__Antialiasing)

	var bc *qt.QColor
	if common.IsDarkTheme() {
		bc = w.darkBackgroundColor
	} else {
		bc = w.lightBackgroundColor
	}
	painter.SetPen(bc)
	y := w.Height() / 2
	painter.DrawLine2(0, y, w.Width(), y)

	if w.Minimum() >= w.Maximum() {
		painter.End()
		return
	}

	painter.SetPenWithStyle(qt.NoPen)
	bar := w.BarColor() // borrowed (may return the stored field color) — do NOT Delete
	brush := qt.NewQBrush3(bar)
	painter.SetBrush(brush)
	brush.Delete()

	width := int(w.val / float64(w.Maximum()-w.Minimum()) * float64(w.Width()))
	r := float64(w.Height()) / 2
	painter.DrawRoundedRect2(0, 0, width, w.Height(), r, r)
	painter.End()
}

// IndeterminateProgressBar is an animated indeterminate progress bar.
//
// The Python version drives two QPropertyAnimations on custom properties; the
// Go port approximates the sweep with a QTimer.
//
// Constructors
//   - NewIndeterminateProgressBar(parent *qt.QWidget, start bool)
type IndeterminateProgressBar struct {
	*qt.QProgressBar
	shortPos      float64
	longPos       float64
	lightBarColor *qt.QColor
	darkBarColor  *qt.QColor
	isError       bool
	isPaused      bool
	elapsed       int
	timer         *qt.QTimer
}

const indeterminateTimerInterval = 16

// NewIndeterminateProgressBar builds an indeterminate progress bar.
func NewIndeterminateProgressBar(parent *qt.QWidget, start bool) *IndeterminateProgressBar {
	w := &IndeterminateProgressBar{QProgressBar: qt.NewQProgressBar(parent)}
	w.lightBarColor = qt.NewQColor()
	w.darkBarColor = qt.NewQColor()
	w.SetFixedHeight(4)

	w.timer = qt.NewQTimer2(w.QObject)
	w.timer.SetInterval(indeterminateTimerInterval)
	w.timer.OnTimeout(func() {
		if !w.isPaused {
			w.elapsed += indeterminateTimerInterval
			w.advance()
			w.Update()
		}
	})

	w.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		w.paint()
	})

	if start {
		w.Start()
	}
	return w
}

func (w *IndeterminateProgressBar) advance() {
	// Short bar: 833ms cycle, sweeping 0 -> 1.45.
	w.shortPos = float64(w.elapsed%833) / 833.0 * 1.45

	// Long bar: 785ms pause then a 1167ms sweep 0 -> 1.75.
	cycle := w.elapsed % 1952
	if cycle >= 785 {
		w.longPos = float64(cycle-785) / 1167.0 * 1.75
	} else {
		w.longPos = 0
	}
}

// LightBarColor returns the light-theme bar color.
func (w *IndeterminateProgressBar) LightBarColor() *qt.QColor {
	if w.lightBarColor.IsValid() {
		return w.lightBarColor
	}
	return common.ThemeColorPrimary.Color()
}

// DarkBarColor returns the dark-theme bar color.
func (w *IndeterminateProgressBar) DarkBarColor() *qt.QColor {
	if w.darkBarColor.IsValid() {
		return w.darkBarColor
	}
	return common.ThemeColorPrimary.Color()
}

// SetCustomBarColor sets the light/dark bar colors.
func (w *IndeterminateProgressBar) SetCustomBarColor(light, dark *qt.QColor) {
	w.lightBarColor = cloneColor(light)
	w.darkBarColor = cloneColor(dark)
	w.Update()
}

// Start starts the animation.
func (w *IndeterminateProgressBar) Start() {
	w.shortPos = 0
	w.longPos = 0
	w.elapsed = 0
	w.isPaused = false
	w.timer.Start(indeterminateTimerInterval)
	w.Update()
}

// Stop stops the animation.
func (w *IndeterminateProgressBar) Stop() {
	w.timer.Stop()
	w.shortPos = 0
	w.longPos = 0
	w.Update()
}

// IsStarted reports whether the animation is running.
func (w *IndeterminateProgressBar) IsStarted() bool { return w.timer.IsActive() }

// Pause pauses the animation.
func (w *IndeterminateProgressBar) Pause() {
	w.isPaused = true
	w.timer.Stop()
	w.Update()
}

// Resume resumes the animation.
func (w *IndeterminateProgressBar) Resume() {
	w.isPaused = false
	w.timer.Start(indeterminateTimerInterval)
	w.Update()
}

// SetPaused sets the paused state.
func (w *IndeterminateProgressBar) SetPaused(isPaused bool) {
	w.isPaused = isPaused
	w.Update()
}

// IsPaused reports the paused state.
func (w *IndeterminateProgressBar) IsPaused() bool { return w.isPaused }

// Error stops the animation and shows the error color.
func (w *IndeterminateProgressBar) Error() {
	w.isError = true
	w.timer.Stop()
	w.Update()
}

// SetError sets the error state.
func (w *IndeterminateProgressBar) SetError(isError bool) {
	w.isError = isError
	if isError {
		w.Error()
	} else {
		w.Start()
	}
}

// IsError reports the error state.
func (w *IndeterminateProgressBar) IsError() bool { return w.isError }

// BarColor returns the theme/state-appropriate bar color.
func (w *IndeterminateProgressBar) BarColor() *qt.QColor {
	if w.IsError() {
		if common.IsDarkTheme() {
			return qt.NewQColor3(255, 153, 164)
		}
		return qt.NewQColor3(196, 43, 28)
	}
	if w.IsPaused() {
		if common.IsDarkTheme() {
			return qt.NewQColor3(252, 225, 0)
		}
		return qt.NewQColor3(157, 93, 0)
	}
	if common.IsDarkTheme() {
		return w.DarkBarColor()
	}
	return w.LightBarColor()
}

func (w *IndeterminateProgressBar) paint() {
	painter := qt.NewQPainter2(w.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__Antialiasing)
	painter.SetPenWithStyle(qt.NoPen)

	bar := w.BarColor() // borrowed (may return the stored field color) — do NOT Delete
	brush := qt.NewQBrush3(bar)
	painter.SetBrush(brush)
	brush.Delete()

	r := float64(w.Height()) / 2

	x := int((w.shortPos - 0.4) * float64(w.Width()))
	width := int(0.4 * float64(w.Width()))
	painter.DrawRoundedRect2(x, 0, width, w.Height(), r, r)

	x = int((w.longPos - 0.6) * float64(w.Width()))
	width = int(0.6 * float64(w.Width()))
	painter.DrawRoundedRect2(x, 0, width, w.Height(), r, r)
	painter.End()
}
