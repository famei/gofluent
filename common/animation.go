package common

import (
	"fmt"
	"math"

	qt "github.com/mappu/miqt/qt"
)

// FluentAnimationSpeed enumerates the animation speed presets.
type FluentAnimationSpeed int

const (
	FluentAnimationSpeedFast FluentAnimationSpeed = iota
	FluentAnimationSpeedMedium
	FluentAnimationSpeedSlow
)

// FluentAnimationType enumerates the built-in animation types.
type FluentAnimationType int

const (
	FluentAnimationTypeFastInvoke FluentAnimationType = iota
	FluentAnimationTypeStrongInvoke
	FluentAnimationTypeFastDismiss
	FluentAnimationTypeSoftDismiss
	FluentAnimationTypePointToPoint
	FluentAnimationTypeFadeInOut
)

// FluentAnimationProperty enumerates the animatable property names.
type FluentAnimationProperty string

const (
	FluentAnimationPropertyPosition FluentAnimationProperty = "position"
	FluentAnimationPropertyScale    FluentAnimationProperty = "scale"
	FluentAnimationPropertyAngle    FluentAnimationProperty = "angle"
	FluentAnimationPropertyOpacity  FluentAnimationProperty = "opacity"
)

// qwidgetSizeMax is Qt's QWIDGETSIZE_MAX, used to lift a maximum size constraint
// after an expand animation completes (mirrors Python's setMaximumHeight).
const qwidgetSizeMax = 16777215

// FluentAnimation collects the timing parameters of the built-in animations.
// The Python implementation subclasses QPropertyAnimation and animates custom
// meta-object properties, which Go cannot register without extra C++; the
// reusable curve/duration tables are preserved here and the generic
// ProgressAnimation below is used to drive those custom properties manually.
type FluentAnimation struct{}

// CreateBezierCurve builds a QEasingCurve with the given cubic bezier control
// points (endpoint is fixed at (1,1)).
func CreateBezierCurve(x1, y1, x2, y2 float64) *qt.QEasingCurve {
	curve := qt.NewQEasingCurve3(qt.QEasingCurve__BezierSpline)
	c1 := qt.NewQPointF3(x1, y1)
	defer c1.Delete()
	c2 := qt.NewQPointF3(x2, y2)
	defer c2.Delete()
	end := qt.NewQPointF3(1, 1)
	defer end.Delete()
	curve.AddCubicBezierSegment(c1, c2, end)
	return curve
}

// FluentAnimationCurve returns the bezier control points for a type (the
// `curve()` classmethod of each registered animation).
func FluentAnimationCurve(t FluentAnimationType) (x1, y1, x2, y2 float64) {
	switch t {
	case FluentAnimationTypeFastInvoke, FluentAnimationTypeFastDismiss, FluentAnimationTypePointToPoint:
		if t == FluentAnimationTypePointToPoint {
			return 0.55, 0.55, 0, 1
		}
		return 0, 0, 0, 1
	case FluentAnimationTypeStrongInvoke:
		return 0.13, 1.62, 0, 0.92
	case FluentAnimationTypeSoftDismiss:
		return 1, 0, 1, 1
	default:
		return 0, 0, 1, 1
	}
}

// FluentAnimationDuration returns the duration in ms for a type and speed
// (the `speedToDuration` method of each registered animation).
func FluentAnimationDuration(t FluentAnimationType, speed FluentAnimationSpeed) int {
	switch t {
	case FluentAnimationTypeFastInvoke, FluentAnimationTypeFastDismiss, FluentAnimationTypePointToPoint:
		switch speed {
		case FluentAnimationSpeedFast:
			return 187
		case FluentAnimationSpeedMedium:
			return 333
		default:
			return 500
		}
	case FluentAnimationTypeStrongInvoke:
		return 667
	case FluentAnimationTypeSoftDismiss:
		return 167
	case FluentAnimationTypeFadeInOut:
		return 83
	default:
		return 100
	}
}

// ---------------------------------------------------------------------------
// ProgressAnimation — generic custom-property interpolation driver

// ProgressAnimation drives an arbitrary custom-property animation by exposing a
// normalized 0..1 progress (already easing-curved). It wraps a QVariantAnimation
// with start=0 end=1 and re-emits OnValueChanged as OnProgress(float64).
//
// Fluent's custom meta-object properties (angle / slideValue / backgroundColor /
// length / position / scale ...) cannot be registered from Go, so they are all
// animated by lerping inside OnProgress. QColor is intentionally excluded from
// QVariantAnimation's native interpolation because QVariant has no ToColor()
// readback in miqt — use AnimateColor below instead.
type ProgressAnimation struct {
	ani *qt.QVariantAnimation

	onProgress func(t float64)
	onFinished func()
}

// NewProgressAnimation builds an animation with the given duration(ms) and
// easing curve. The curve is copied by QVariantAnimation.SetEasingCurve, so the
// caller keeps ownership of curve and may Delete it immediately after this call.
// Pass nil to use the default linear curve.
func NewProgressAnimation(duration int, curve *qt.QEasingCurve) *ProgressAnimation {
	a := &ProgressAnimation{ani: qt.NewQVariantAnimation()}

	start := qt.NewQVariant12(0.0)
	a.ani.SetStartValue(start)
	start.Delete()
	end := qt.NewQVariant12(1.0)
	a.ani.SetEndValue(end)
	end.Delete()

	a.ani.SetDuration(duration)
	if curve != nil {
		a.ani.SetEasingCurve(curve)
	}

	a.ani.OnValueChanged(func(v *qt.QVariant) {
		if a.onProgress != nil {
			a.onProgress(v.ToDouble())
		}
	})
	a.ani.OnFinished(func() {
		if a.onFinished != nil {
			a.onFinished()
		}
	})
	return a
}

// NewFluentProgressAnimation builds a ProgressAnimation from Fluent type/speed.
func NewFluentProgressAnimation(t FluentAnimationType, speed FluentAnimationSpeed) *ProgressAnimation {
	x1, y1, x2, y2 := FluentAnimationCurve(t)
	curve := CreateBezierCurve(x1, y1, x2, y2)
	defer curve.Delete() // SetEasingCurve copies; the temporary is freed here.
	return NewProgressAnimation(FluentAnimationDuration(t, speed), curve)
}

// OnProgress registers the per-frame callback (t in [0,1], easing applied).
func (a *ProgressAnimation) OnProgress(f func(t float64)) { a.onProgress = f }

// OnFinished registers the completion callback (fires on natural completion
// only; Stop does not emit it).
func (a *ProgressAnimation) OnFinished(f func()) { a.onFinished = f }

// Start stops any in-flight run and starts a new one.
func (a *ProgressAnimation) Start() {
	if a.ani == nil {
		return
	}
	a.ani.Stop()
	a.ani.Start()
}

// Stop halts the animation, keeping the current interpolated value.
func (a *ProgressAnimation) Stop() {
	if a.ani != nil {
		a.ani.Stop()
	}
}

// Running reports whether the underlying QVariantAnimation is running.
func (a *ProgressAnimation) Running() bool {
	return a.ani != nil && a.ani.State() == qt.QAbstractAnimation__Running
}

// Delete stops and releases the underlying C++ QVariantAnimation. It is safe to
// call more than once. After Delete the animation must not be reused.
func (a *ProgressAnimation) Delete() {
	if a == nil || a.ani == nil {
		return
	}
	a.ani.Stop()
	a.ani.Delete()
	a.ani = nil
}

// AnimateFloat interpolates between two scalar values and calls apply with the
// eased value each frame. It is the generic "custom property" wrapper: pass a
// setter (or SetProperty) in apply.
func AnimateFloat(from, to float64, duration int, curve *qt.QEasingCurve, apply func(float64)) *ProgressAnimation {
	a := NewProgressAnimation(duration, curve)
	a.OnProgress(func(t float64) {
		apply(from + (to-from)*t)
	})
	a.Start()
	return a
}

// ---------------------------------------------------------------------------
// Fade (opacity)

// FadeIn fades a widget in from fully transparent to fully opaque using a
// QGraphicsOpacityEffect + QPropertyAnimation("opacity"). The effect and
// animation are parented to w, so their C++ memory follows w's lifetime; the
// returned animation must not be Deleted while parented. Repeated calls attach
// a fresh effect (the previous one stays parented but inactive until w dies).
func FadeIn(w *qt.QWidget, duration int, curve *qt.QEasingCurve) *qt.QPropertyAnimation {
	return animateOpacity(w, 0.0, 1.0, duration, curve)
}

// FadeOut fades a widget out from fully opaque to fully transparent.
func FadeOut(w *qt.QWidget, duration int, curve *qt.QEasingCurve) *qt.QPropertyAnimation {
	return animateOpacity(w, 1.0, 0.0, duration, curve)
}

// FadeInThen fades a widget in and runs done once the animation has finished. Use it for a
// child widget that carries a drop shadow: Qt gives a widget only one graphics effect, so
// the opacity effect replaces the shadow and done has to install it again. A child widget
// repaints normally, which is what makes this fade visible where fading a top level window
// is not (see FadeWindowIn).
func FadeInThen(w *qt.QWidget, duration int, curve *qt.QEasingCurve, done func()) *qt.QPropertyAnimation {
	return animateOpacityThen(w, 0.0, 1.0, duration, curve, done)
}

// FadeOutThen fades a widget out and runs done once the animation has finished.
func FadeOutThen(w *qt.QWidget, duration int, curve *qt.QEasingCurve, done func()) *qt.QPropertyAnimation {
	return animateOpacityThen(w, 1.0, 0.0, duration, curve, done)
}

func animateOpacity(w *qt.QWidget, from, to float64, duration int, curve *qt.QEasingCurve) *qt.QPropertyAnimation {
	return animateOpacityThen(w, from, to, duration, curve, nil)
}

func animateOpacityThen(w *qt.QWidget, from, to float64, duration int, curve *qt.QEasingCurve, done func()) *qt.QPropertyAnimation {
	effect := qt.NewQGraphicsOpacityEffect2(w.QObject)
	w.SetGraphicsEffect(effect.QGraphicsEffect)

	ani := qt.NewQPropertyAnimation4(effect.QObject, []byte("opacity"), w.QObject)
	ani.SetDuration(duration)
	if curve != nil {
		ani.SetEasingCurve(curve)
	}

	start := qt.NewQVariant12(from)
	ani.SetStartValue(start)
	start.Delete()
	end := qt.NewQVariant12(to)
	ani.SetEndValue(end)
	end.Delete()

	if done != nil {
		ani.OnFinished(func() {
			w.SetGraphicsEffect(nil)
			done()
		})
	}

	ani.Start()
	return ani
}

// FadeWindowIn fades a top-level window in via its windowOpacity property
// (0 -> 1). Prefer FadeIn/FadeOut for non-window child widgets.
func FadeWindowIn(w *qt.QWidget, duration int, curve *qt.QEasingCurve) *qt.QPropertyAnimation {
	return animateWindowOpacity(w, 0.0, 1.0, duration, curve)
}

// FadeWindowOut fades a top-level window out via windowOpacity (1 -> 0).
func FadeWindowOut(w *qt.QWidget, duration int, curve *qt.QEasingCurve) *qt.QPropertyAnimation {
	return animateWindowOpacity(w, 1.0, 0.0, duration, curve)
}

// ApplyAfterMap runs fn once the widget has been mapped, i.e. on the next event loop
// iteration after it was shown. A position set before the window is mapped is not final: a
// window manager may place the window itself, so anything that anchors a popup re-applies
// its geometry here (see also widgets.ToolTip's move guard, which corrects a window manager
// that moves the window after the map).
func ApplyAfterMap(w *qt.QWidget, fn func()) {
	if w == nil || fn == nil {
		return
	}
	timer := qt.NewQTimer2(w.QObject)
	timer.SetSingleShot(true)
	timer.OnTimeout(func() {
		fn()
		timer.DeleteLater()
	})
	timer.Start(0)
}

func animateWindowOpacity(w *qt.QWidget, from, to float64, duration int, curve *qt.QEasingCurve) *qt.QPropertyAnimation {
	ani := qt.NewQPropertyAnimation4(w.QObject, []byte("windowOpacity"), w.QObject)
	ani.SetDuration(duration)
	if curve != nil {
		ani.SetEasingCurve(curve)
	}

	start := qt.NewQVariant12(from)
	ani.SetStartValue(start)
	start.Delete()
	end := qt.NewQVariant12(to)
	ani.SetEndValue(end)
	end.Delete()

	ani.Start()
	return ani
}

// ---------------------------------------------------------------------------
// Slide (pos / geometry)

// SlidePos animates a widget's pos property from one point to another.
func SlidePos(w *qt.QWidget, from, to *qt.QPoint, duration int, curve *qt.QEasingCurve) *qt.QPropertyAnimation {
	ani := qt.NewQPropertyAnimation4(w.QObject, []byte("pos"), w.QObject)
	ani.SetDuration(duration)
	if curve != nil {
		ani.SetEasingCurve(curve)
	}

	start := qt.NewQVariant27(from)
	ani.SetStartValue(start)
	start.Delete()
	end := qt.NewQVariant27(to)
	ani.SetEndValue(end)
	end.Delete()

	ani.Start()
	return ani
}

// SlideGeometry animates a widget's geometry property from one rect to another
// (move + resize).
func SlideGeometry(w *qt.QWidget, from, to *qt.QRect, duration int, curve *qt.QEasingCurve) *qt.QPropertyAnimation {
	ani := qt.NewQPropertyAnimation4(w.QObject, []byte("geometry"), w.QObject)
	ani.SetDuration(duration)
	if curve != nil {
		ani.SetEasingCurve(curve)
	}

	start := qt.NewQVariant31(from)
	ani.SetStartValue(start)
	start.Delete()
	end := qt.NewQVariant31(to)
	ani.SetEndValue(end)
	end.Delete()

	ani.Start()
	return ani
}

// ---------------------------------------------------------------------------
// Expand / Collapse (maximumHeight)

// AnimateMaximumHeight animates a widget's maximumHeight property between two
// pixel heights. The returned animation is parented to w.
func AnimateMaximumHeight(w *qt.QWidget, from, to int, duration int, curve *qt.QEasingCurve) *qt.QPropertyAnimation {
	ani := qt.NewQPropertyAnimation4(w.QObject, []byte("maximumHeight"), w.QObject)
	ani.SetDuration(duration)
	if curve != nil {
		ani.SetEasingCurve(curve)
	}

	start := qt.NewQVariant7(from)
	ani.SetStartValue(start)
	start.Delete()
	end := qt.NewQVariant7(to)
	ani.SetEndValue(end)
	end.Delete()

	ani.Start()
	return ani
}

// Expand animates a widget's maximumHeight from its current rendered height up
// to targetHeight, then lifts the maximum constraint (QWIDGETSIZE_MAX) so the
// layout is no longer clipped.
func Expand(w *qt.QWidget, targetHeight, duration int, curve *qt.QEasingCurve) *qt.QPropertyAnimation {
	ani := AnimateMaximumHeight(w, w.Height(), targetHeight, duration, curve)
	ani.OnFinished(func() {
		w.SetMaximumHeight(qwidgetSizeMax)
	})
	return ani
}

// Collapse animates a widget's maximumHeight from its current rendered height
// down to zero.
func Collapse(w *qt.QWidget, duration int, curve *qt.QEasingCurve) *qt.QPropertyAnimation {
	return AnimateMaximumHeight(w, w.Height(), 0, duration, curve)
}

// ---------------------------------------------------------------------------
// Color interpolation

func lerpInt(a, b int, t float64) int {
	return a + int(math.Round(float64(b-a)*t))
}

// AnimateColor lerps the RGBA channels of two colors and calls apply with a
// fresh QColor each frame. Colors are read once at construction time, so from
// and to stay owned by the caller and may be deleted immediately after this
// call returns. The returned ProgressAnimation is not parented: the caller must
// hold it and call Delete when finished (or when no longer needed).
func AnimateColor(from, to *qt.QColor, duration int, curve *qt.QEasingCurve, apply func(*qt.QColor)) *ProgressAnimation {
	fr, fg, fb, fa := from.Red(), from.Green(), from.Blue(), from.Alpha()
	tr, tg, tb, ta := to.Red(), to.Green(), to.Blue(), to.Alpha()

	a := NewProgressAnimation(duration, curve)
	a.OnProgress(func(t float64) {
		c := qt.NewQColor11(
			lerpInt(fr, tr, t),
			lerpInt(fg, tg, t),
			lerpInt(fb, tb, t),
			lerpInt(fa, ta, t),
		)
		apply(c)
		c.Delete()
	})
	a.Start()
	return a
}

// AnimateBackgroundColor transitions a widget's stylesheet background-color
// between two colors. It overwrites the widget's existing style sheet each
// frame, so it is intended for widgets whose background is driven by QSS.
func AnimateBackgroundColor(w *qt.QWidget, from, to *qt.QColor, duration int, curve *qt.QEasingCurve) *ProgressAnimation {
	return AnimateColor(from, to, duration, curve, func(c *qt.QColor) {
		w.SetStyleSheet(fmt.Sprintf("background-color: rgba(%d, %d, %d, %d);", c.Red(), c.Green(), c.Blue(), c.Alpha()))
	})
}

// ---------------------------------------------------------------------------
// Parallel grouping

// ParallelGroup creates a QParallelAnimationGroup holding the given animations
// (without starting it). When parent is non-nil the group follows parent's
// lifetime and must not be Deleted; when parent is nil the caller owns the
// group and must Delete it after use. Call Start on the returned group to run.
func ParallelGroup(parent *qt.QObject, animations ...*qt.QAbstractAnimation) *qt.QParallelAnimationGroup {
	var group *qt.QParallelAnimationGroup
	if parent != nil {
		group = qt.NewQParallelAnimationGroup2(parent)
	} else {
		group = qt.NewQParallelAnimationGroup()
	}
	for _, a := range animations {
		if a != nil {
			group.AddAnimation(a)
		}
	}
	return group
}

// ---------------------------------------------------------------------------
// DropShadowAnimation

// DropShadowAnimation attaches a QGraphicsDropShadowEffect to a widget and
// animates its color on hover. The effect is parented to the widget; the
// running color transition is held in a struct field and stopped/cleaned when a
// new transition starts.
type DropShadowAnimation struct {
	parent      *qt.QWidget
	normalColor *qt.QColor
	hoverColor  *qt.QColor
	offsetX     float64
	offsetY     float64
	blurRadius  float64

	shadowEffect *qt.QGraphicsDropShadowEffect
	ani          *ProgressAnimation
}

// NewDropShadowAnimation builds a drop shadow animation for parent.
func NewDropShadowAnimation(parent *qt.QWidget, normalColor, hoverColor *qt.QColor) *DropShadowAnimation {
	a := &DropShadowAnimation{
		parent:      parent,
		normalColor: normalColor,
		hoverColor:  hoverColor,
		blurRadius:  38,
	}
	a.applyEffect(a.normalColor)
	// NOTE: hover enter/leave are NOT registered here. The only caller
	// (widgets.NewElevatedCardWidget) re-registers OnEnterEvent/OnLeaveEvent on
	// its own directly-constructed QFrame; registering them here on the passed
	// embedded *QWidget would panic with "miqt: can only override virtual
	// methods for directly constructed types".
	return a
}

// SetBlurRadius sets the shadow blur radius.
func (a *DropShadowAnimation) SetBlurRadius(radius float64) { a.blurRadius = radius }

// SetOffset sets the shadow offset.
func (a *DropShadowAnimation) SetOffset(dx, dy float64) {
	a.offsetX, a.offsetY = dx, dy
}

// SetNormalColor sets the resting shadow color.
func (a *DropShadowAnimation) SetNormalColor(color *qt.QColor) { a.normalColor = color }

// SetHoverColor sets the hover shadow color.
func (a *DropShadowAnimation) SetHoverColor(color *qt.QColor) { a.hoverColor = color }

// SetHover transitions the shadow color toward the hover or normal color over
// 150ms. It re-applies the current offset/blur settings (so late SetOffset /
// SetBlurRadius calls take effect) and interrupts any in-flight transition.
func (a *DropShadowAnimation) SetHover(hover bool) {
	a.ensureEffect()
	a.shadowEffect.SetOffset2(a.offsetX, a.offsetY)
	a.shadowEffect.SetBlurRadius(a.blurRadius)

	a.stopAnimation()

	from := qt.NewQColor9(a.shadowEffect.Color())
	var to *qt.QColor
	if hover {
		to = a.hoverColor
	} else {
		to = a.normalColor
	}

	a.ani = AnimateColor(from, to, 150, qt.NewQEasingCurve3(qt.QEasingCurve__OutCubic), func(c *qt.QColor) {
		a.shadowEffect.SetColor(c)
	})
	from.Delete()
}

func (a *DropShadowAnimation) applyEffect(color *qt.QColor) {
	a.ensureEffect()
	a.shadowEffect.SetOffset2(a.offsetX, a.offsetY)
	a.shadowEffect.SetBlurRadius(a.blurRadius)
	a.shadowEffect.SetColor(color)
}

func (a *DropShadowAnimation) ensureEffect() {
	if a.shadowEffect == nil {
		a.shadowEffect = qt.NewQGraphicsDropShadowEffect2(a.parent.QObject)
		a.parent.SetGraphicsEffect(a.shadowEffect.QGraphicsEffect)
	}
}

func (a *DropShadowAnimation) stopAnimation() {
	if a.ani != nil {
		a.ani.Stop()
		a.ani.Delete()
		a.ani = nil
	}
}

// ShadowEffect returns the attached effect (nil until first applied).
func (a *DropShadowAnimation) ShadowEffect() *qt.QGraphicsDropShadowEffect {
	return a.shadowEffect
}

// ---------------------------------------------------------------------------
// ScaleSlideAnimation geometry math

// ScaleSlideGeometry holds the indicator geometry used by the WinUI
// squash-and-stretch slide animation (pivot/segmented indicators). It ports the
// geometry/pos/length property logic of ScaleSlideAnimation.
type ScaleSlideGeometry struct {
	horizontal bool
	geometry   *qt.QRectF
}

// NewScaleSlideGeometry builds the geometry holder.
func NewScaleSlideGeometry(horizontal bool) *ScaleSlideGeometry {
	var g *qt.QRectF
	if horizontal {
		g = qt.NewQRectF4(0, 0, 16, 3)
	} else {
		g = qt.NewQRectF4(0, 0, 3, 16)
	}
	return &ScaleSlideGeometry{horizontal: horizontal, geometry: g}
}

// IsHorizontal reports the orientation.
func (s *ScaleSlideGeometry) IsHorizontal() bool { return s.horizontal }

// Geometry returns the current geometry.
func (s *ScaleSlideGeometry) Geometry() *qt.QRectF { return s.geometry }

// SetGeometry replaces the geometry.
func (s *ScaleSlideGeometry) SetGeometry(rect *qt.QRectF) {
	if s.geometry != nil {
		s.geometry.Delete()
	}
	s.geometry = qt.NewQRectF6(rect)
}

// Pos returns the top-left corner.
func (s *ScaleSlideGeometry) Pos() *qt.QPointF { return s.geometry.TopLeft() }

// SetPos moves the top-left corner.
func (s *ScaleSlideGeometry) SetPos(pos *qt.QPointF) {
	s.geometry.MoveTopLeft(pos)
}

// Length returns the width (horizontal) or height (vertical).
func (s *ScaleSlideGeometry) Length() float64 {
	if s.horizontal {
		return s.geometry.Width()
	}
	return s.geometry.Height()
}

// SetLength sets the width (horizontal) or height (vertical).
func (s *ScaleSlideGeometry) SetLength(length float64) {
	if s.horizontal {
		s.geometry.SetWidth(length)
	} else {
		s.geometry.SetHeight(length)
	}
}

// MoveLeft moves the geometry's left edge to x.
func (s *ScaleSlideGeometry) MoveLeft(x float64) {
	s.geometry.MoveLeft(x)
}
