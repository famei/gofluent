package common

import (
	"math"
	"time"

	qt "github.com/mappu/miqt/qt"
)

// SmoothMode enumerates the interpolation modes of the smooth scroller.
type SmoothMode int

const (
	SmoothModeNoSmooth SmoothMode = iota
	SmoothModeConstant
	SmoothModeLinear
	SmoothModeQuadrati
	SmoothModeCosine
)

// SmoothScroll scrolls a QAbstractScrollArea smoothly. Attach it by delegating
// the area's OnWheelEvent: if HandleWheelEvent returns false, call the base
// implementation.
type SmoothScroll struct {
	widget               *qt.QAbstractScrollArea
	orient               qt.Orientation
	dynamicEngineEnabled bool
	widthThreshold       float64
	smoothMode           SmoothMode

	fixedStepScrollEngine *FixedStepSmoothScrollEngine
	adaptiveScrollEngine  *AdaptiveSmoothScrollEngine
}

// NewSmoothScroll builds a smooth scroller for a scroll area.
func NewSmoothScroll(widget *qt.QAbstractScrollArea, orient qt.Orientation) *SmoothScroll {
	s := &SmoothScroll{
		widget:               widget,
		orient:               orient,
		dynamicEngineEnabled: true,
		widthThreshold:       2560,
		smoothMode:           SmoothModeLinear,
	}
	s.fixedStepScrollEngine = NewFixedStepSmoothScrollEngine(widget, orient)
	s.adaptiveScrollEngine = NewAdaptiveSmoothScrollEngine(widget, orient)
	return s
}

// SetDynamicEngineEnabled toggles the HiDPI adaptive engine.
func (s *SmoothScroll) SetDynamicEngineEnabled(enabled bool) {
	s.dynamicEngineEnabled = enabled
}

// SetSmoothMode sets the interpolation mode on both engines.
func (s *SmoothScroll) SetSmoothMode(mode SmoothMode) {
	s.smoothMode = mode
	s.fixedStepScrollEngine.SetSmoothMode(mode)
	s.adaptiveScrollEngine.SetSmoothMode(mode)
}

// HandleWheelEvent processes a wheel event; it returns false when the event
// should be forwarded to the base implementation.
func (s *SmoothScroll) HandleWheelEvent(e *qt.QWheelEvent) bool {
	ad := e.AngleDelta()
	delta := ad.Y()
	if delta == 0 {
		delta = ad.X()
	}

	if s.smoothMode == SmoothModeNoSmooth || intAbs(delta)%120 != 0 {
		return false
	}

	engine := s.chooseScrollEngine()
	engine.wheelEvent(delta)
	return true
}

func (s *SmoothScroll) chooseScrollEngine() smoothScrollEngine {
	if s.dynamicEngineEnabled && float64(s.widget.Width())*s.widget.DevicePixelRatioF() > s.widthThreshold {
		return s.adaptiveScrollEngine
	}
	return s.fixedStepScrollEngine
}

type smoothScrollEngine interface {
	SetSmoothMode(mode SmoothMode)
	wheelEvent(delta int)
}

// smoothScrollEngineBase holds the shared scrolling state. Time is tracked with
// Go's monotonic clock rather than QDateTime/QElapsedTimer.
type smoothScrollEngineBase struct {
	widget   *qt.QAbstractScrollArea
	orient   qt.Orientation
	fps      int
	duration float64

	stepsTotal   float64
	stepRatio    float64
	acceleration float64

	scrollStamps   []int64
	stepsLeftQueue [][2]float64 // {delta, stepsLeft|duration}
	smoothMode     SmoothMode

	// totalDeltaFn dispatches to the concrete engine's totalDelta. Go has no
	// virtual dispatch, so smoothMove (defined on the base) would otherwise
	// always call the base totalDelta() which returns 0 and never scrolls.
	totalDeltaFn    func() float64
	smoothMoveTimer *qt.QTimer
}

func newSmoothScrollEngineBase(widget *qt.QAbstractScrollArea, orient qt.Orientation) *smoothScrollEngineBase {
	b := &smoothScrollEngineBase{
		widget:          widget,
		orient:          orient,
		fps:             60,
		duration:        400,
		stepRatio:       1.5,
		acceleration:    1,
		smoothMode:      SmoothModeLinear,
		smoothMoveTimer: qt.NewQTimer2(widget.QObject),
	}
	b.totalDeltaFn = b.totalDelta // base default (0); overridden by concrete engines
	b.smoothMoveTimer.OnTimeout(b.smoothMove)
	return b
}

func (b *smoothScrollEngineBase) SetSmoothMode(mode SmoothMode) { b.smoothMode = mode }

// smoothMove drains one interpolation step.
func (b *smoothScrollEngineBase) smoothMove() {
	if len(b.stepsLeftQueue) == 0 {
		return
	}
	totalDelta := b.totalDeltaFn()
	b.sendScrollEventToScrollBar(totalDelta)

	if len(b.stepsLeftQueue) == 0 {
		b.smoothMoveTimer.Stop()
	}
}

func (b *smoothScrollEngineBase) totalDelta() float64 { return 0 }

func (b *smoothScrollEngineBase) sendScrollEventToScrollBar(totalDelta float64) {
	var bar *qt.QScrollBar
	if b.orient == qt.Vertical {
		bar = b.widget.VerticalScrollBar()
	} else {
		bar = b.widget.HorizontalScrollBar()
	}
	if bar == nil {
		return
	}
	// Match Python SmoothScrollEngineBase.sendScrollEventToScrollBar: it forwards
	// a wheel event whose angleDelta equals totalDelta to the scroll bar, and
	// QAbstractSlider::wheelEvent then does setValue(value - angleDelta). Setting
	// the value directly must therefore SUBTRACT totalDelta. The previous "+"
	// inverted the direction, so a wheel-down delta (negative) pushed the value
	// below minimum and got clamped to zero — the scroll appeared dead.
	bar.SetValue(bar.Value() - int(math.Round(totalDelta)))
}

func (b *smoothScrollEngineBase) pushStamp() int64 {
	now := time.Now().UnixMilli()
	b.scrollStamps = append(b.scrollStamps, now)
	for len(b.scrollStamps) > 0 && now-b.scrollStamps[0] > 500 {
		b.scrollStamps = b.scrollStamps[1:]
	}
	return now
}

// FixedStepSmoothScrollEngine scrolls by a fixed step count per event.
type FixedStepSmoothScrollEngine struct {
	*smoothScrollEngineBase
}

// NewFixedStepSmoothScrollEngine builds the fixed-step engine.
func NewFixedStepSmoothScrollEngine(widget *qt.QAbstractScrollArea, orient qt.Orientation) *FixedStepSmoothScrollEngine {
	e := &FixedStepSmoothScrollEngine{smoothScrollEngineBase: newSmoothScrollEngineBase(widget, orient)}
	e.totalDeltaFn = e.totalDelta
	return e
}

func (e *FixedStepSmoothScrollEngine) wheelEvent(delta int) {
	e.pushStamp()

	accelerationRatio := math.Min(float64(len(e.scrollStamps))/15, 1)
	e.stepsTotal = float64(e.fps) * e.duration / 1000

	d := float64(delta) * e.stepRatio
	if e.acceleration > 0 {
		d += d * e.acceleration * accelerationRatio
	}

	e.stepsLeftQueue = append(e.stepsLeftQueue, [2]float64{d, e.stepsTotal})
	e.smoothMoveTimer.Start(int(1000 / e.fps))
}

func (e *FixedStepSmoothScrollEngine) totalDelta() float64 {
	total := 0.0
	for i := range e.stepsLeftQueue {
		total += e.subDelta(e.stepsLeftQueue[i][0], e.stepsLeftQueue[i][1])
		e.stepsLeftQueue[i][1]--
	}
	for len(e.stepsLeftQueue) > 0 && e.stepsLeftQueue[0][1] == 0 {
		e.stepsLeftQueue = e.stepsLeftQueue[1:]
	}
	return total
}

func (e *FixedStepSmoothScrollEngine) subDelta(delta, stepsLeft float64) float64 {
	m := e.stepsTotal / 2
	x := math.Abs(e.stepsTotal - stepsLeft - m)

	switch e.smoothMode {
	case SmoothModeNoSmooth:
		return 0
	case SmoothModeConstant:
		return delta / e.stepsTotal
	case SmoothModeLinear:
		return 2 * delta / e.stepsTotal * (m - x) / m
	case SmoothModeQuadrati:
		return 3.0 / 4.0 / m * (1 - x*x/m/m) * delta
	case SmoothModeCosine:
		return (math.Cos(x*math.Pi/m) + 1) / (2 * m) * delta
	default:
		return 0
	}
}

// AdaptiveSmoothScrollEngine is a time-based engine for HiDPI displays.
type AdaptiveSmoothScrollEngine struct {
	*smoothScrollEngineBase
	lastTick     time.Time
	maxQueueSize int
	minDuration  float64
}

// NewAdaptiveSmoothScrollEngine builds the adaptive engine.
func NewAdaptiveSmoothScrollEngine(widget *qt.QAbstractScrollArea, orient qt.Orientation) *AdaptiveSmoothScrollEngine {
	e := &AdaptiveSmoothScrollEngine{
		smoothScrollEngineBase: newSmoothScrollEngineBase(widget, orient),
		maxQueueSize:           3,
		minDuration:            120,
	}
	e.totalDeltaFn = e.totalDelta
	return e
}

func (e *AdaptiveSmoothScrollEngine) wheelEvent(delta int) {
	e.pushStamp()
	accelerationRatio := math.Min(float64(len(e.scrollStamps))/15, 1)

	queuePressure := len(e.stepsLeftQueue)
	pressureRatio := math.Min(float64(queuePressure)/float64(e.maxQueueSize), 1)

	effectiveDuration := e.duration * (1 - 0.6*pressureRatio)
	effectiveDuration = math.Max(e.minDuration, effectiveDuration)

	d := float64(delta) * e.stepRatio
	if e.acceleration > 0 {
		d += d * e.acceleration * accelerationRatio
	}

	if len(e.stepsLeftQueue) >= e.maxQueueSize {
		last := &e.stepsLeftQueue[len(e.stepsLeftQueue)-1]
		last[0] += d
		last[1] = math.Max(last[1], effectiveDuration)
	} else {
		e.stepsLeftQueue = append(e.stepsLeftQueue, [2]float64{d, effectiveDuration})
	}

	e.lastTick = time.Now()
	e.smoothMoveTimer.Start(int(1000 / e.fps))
}

func (e *AdaptiveSmoothScrollEngine) totalDelta() float64 {
	now := time.Now()
	dt := float64(now.Sub(e.lastTick).Milliseconds())
	e.lastTick = now
	if dt < 0 {
		dt = 0
	}

	total := 0.0
	for i := range e.stepsLeftQueue {
		remainingDelta := e.stepsLeftQueue[i][0]
		remainingTime := e.stepsLeftQueue[i][1]
		if remainingTime <= 0 {
			continue
		}
		consumeTime := math.Min(dt, remainingTime)
		ratio := consumeTime / remainingTime
		subDelta := e.subDelta(remainingDelta, ratio)

		e.stepsLeftQueue[i][0] -= subDelta
		e.stepsLeftQueue[i][1] -= consumeTime
		total += subDelta
	}

	for len(e.stepsLeftQueue) > 0 && e.stepsLeftQueue[0][1] <= 0 {
		e.stepsLeftQueue = e.stepsLeftQueue[1:]
	}
	return total
}

func (e *AdaptiveSmoothScrollEngine) subDelta(delta, ratio float64) float64 {
	switch e.smoothMode {
	case SmoothModeConstant, SmoothModeLinear:
		return delta * ratio
	case SmoothModeQuadrati:
		return delta * (1 - math.Pow(1-ratio, 2))
	case SmoothModeCosine:
		return delta * (1 - math.Cos(ratio*math.Pi)) / 2
	default:
		return delta * ratio
	}
}

func intAbs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
