package widgets

import (
	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// pageFade drives a short fade-in transition for a stacked-widget page. It
// installs a QGraphicsOpacityEffect on the target page and drives its opacity
// from 0 to 1 with a ProgressAnimation. The effect is parented to the target,
// so Qt owns it once installed and frees it when the effect is detached via
// SetGraphicsEffect(nil); the ProgressAnimation is owned by pageFade and freed
// when the next transition starts (or when stop is called). This avoids the
// double-free that a manual effect.Delete() would cause and avoids leaking an
// animation per switch.
type pageFade struct {
	ani    *common.ProgressAnimation
	target *qt.QWidget
	effect *qt.QGraphicsOpacityEffect
	onDone func()
}

// start fades target in from transparent. It cancels any in-flight fade first
// and fires onDone (when non-nil) once the fade completes.
func (f *pageFade) start(target *qt.QWidget, onDone func()) {
	f.stop()
	if target == nil {
		return
	}

	f.target = target
	f.onDone = onDone
	f.effect = qt.NewQGraphicsOpacityEffect2(target.QObject)
	target.SetGraphicsEffect(f.effect.QGraphicsEffect)
	f.effect.SetOpacity(0.0)

	curve := qt.NewQEasingCurve3(qt.QEasingCurve__OutCubic)
	f.ani = common.NewProgressAnimation(150, curve)
	curve.Delete()

	f.ani.OnProgress(func(t float64) {
		if f.effect != nil {
			f.effect.SetOpacity(t)
		}
	})
	f.ani.OnFinished(func() {
		f.finish()
		if f.onDone != nil {
			done := f.onDone
			f.onDone = nil
			done()
		}
	})
	f.ani.Start()
}

// finish completes the fade and detaches the effect. Qt's setGraphicsEffect
// takes ownership of the installed effect and deletes the previous one when a
// new value (including nil) is set, so no manual effect.Delete() is performed
// here — that would be a double free.
func (f *pageFade) finish() {
	if f.effect != nil {
		f.effect.SetOpacity(1.0)
	}
	if f.target != nil {
		f.target.SetGraphicsEffect(nil)
	}
	f.effect = nil
	f.target = nil
}

// stop halts the animation and detaches any lingering effect. The animation is
// freed here (it is owned by pageFade, not by Qt).
func (f *pageFade) stop() {
	f.onDone = nil
	if f.ani != nil {
		f.ani.Stop()
		f.ani.Delete()
		f.ani = nil
	}
	f.finish()
}

// OpacityAniStackedWidget is a stacked widget that fades between pages. Pages
// switch synchronously but the newly shown page fades in from transparent.
type OpacityAniStackedWidget struct {
	*qt.QStackedWidget
	fade pageFade
}

// NewOpacityAniStackedWidget builds an opacity-animated stacked widget.
func NewOpacityAniStackedWidget(parent *qt.QWidget) *OpacityAniStackedWidget {
	return &OpacityAniStackedWidget{QStackedWidget: qt.NewQStackedWidget(parent)}
}

// AddWidget adds a page.
func (w *OpacityAniStackedWidget) AddWidget(widget *qt.QWidget) {
	w.QStackedWidget.AddWidget(widget)
}

// SetCurrentIndex switches to the page at index, fading the new page in.
func (w *OpacityAniStackedWidget) SetCurrentIndex(index int) {
	if index < 0 || index >= w.Count() || index == w.CurrentIndex() {
		return
	}
	w.QStackedWidget.SetCurrentIndex(index)
	w.fade.start(w.Widget(index), nil)
}

// SetCurrentWidget switches to the given page.
func (w *OpacityAniStackedWidget) SetCurrentWidget(widget *qt.QWidget) {
	w.SetCurrentIndex(w.IndexOf(widget))
}

// PopUpAniStackedWidget is a stacked widget that fades the new page in on
// switch. The original WinUI pop-up (scale) transition is approximated by this
// fade; the aniStart/aniFinished callbacks are preserved.
type PopUpAniStackedWidget struct {
	*qt.QStackedWidget
	isAnimationEnabled bool
	onAniFinished      func()
	onAniStart         func()
	fade               pageFade
}

// NewPopUpAniStackedWidget builds a pop-up animated stacked widget.
func NewPopUpAniStackedWidget(parent *qt.QWidget) *PopUpAniStackedWidget {
	return &PopUpAniStackedWidget{QStackedWidget: qt.NewQStackedWidget(parent), isAnimationEnabled: true}
}

// OnAniFinished registers a callback fired after the page switch.
func (w *PopUpAniStackedWidget) OnAniFinished(f func()) { w.onAniFinished = f }

// OnAniStart registers a callback fired before the page switch.
func (w *PopUpAniStackedWidget) OnAniStart(f func()) { w.onAniStart = f }

// AddWidget adds a page (deltaX/deltaY are accepted for API compatibility but
// unused because the pop-up geometry animation is approximated by a fade).
func (w *PopUpAniStackedWidget) AddWidget(widget *qt.QWidget, deltaX, deltaY int) {
	w.QStackedWidget.AddWidget(widget)
}

// RemoveWidget removes a page.
func (w *PopUpAniStackedWidget) RemoveWidget(widget *qt.QWidget) {
	w.QStackedWidget.RemoveWidget(widget)
}

// SetAnimationEnabled toggles the fade animation.
func (w *PopUpAniStackedWidget) SetAnimationEnabled(isEnabled bool) {
	w.isAnimationEnabled = isEnabled
}

// SetCurrentIndex switches to the page at index, fading the new page in when
// the animation is enabled.
func (w *PopUpAniStackedWidget) SetCurrentIndex(index int) {
	if index < 0 || index >= w.Count() || index == w.CurrentIndex() {
		return
	}
	if !w.isAnimationEnabled {
		w.QStackedWidget.SetCurrentIndex(index)
		return
	}
	if w.onAniStart != nil {
		w.onAniStart()
	}
	w.QStackedWidget.SetCurrentIndex(index)
	w.fade.start(w.Widget(index), w.onAniFinished)
}

// SetCurrentWidget switches to the given page.
func (w *PopUpAniStackedWidget) SetCurrentWidget(widget *qt.QWidget) {
	w.SetCurrentIndex(w.IndexOf(widget))
}

// TransitionStackedWidget is the base class of the transition stacked widgets.
// The snapshot/scale/slide transitions are approximated by a fade-in of the
// newly shown page; pages switch synchronously while the API and callbacks are
// preserved.
type TransitionStackedWidget struct {
	*qt.QStackedWidget
	isAnimationEnabled bool
	onAniFinished      func()
	onAniStart         func()
	fade               pageFade
}

// NewTransitionStackedWidget builds a transition stacked widget.
func NewTransitionStackedWidget(parent *qt.QWidget) *TransitionStackedWidget {
	return &TransitionStackedWidget{QStackedWidget: qt.NewQStackedWidget(parent), isAnimationEnabled: true}
}

// OnAniFinished registers a callback fired after the page switch.
func (w *TransitionStackedWidget) OnAniFinished(f func()) { w.onAniFinished = f }

// OnAniStart registers a callback fired before the page switch.
func (w *TransitionStackedWidget) OnAniStart(f func()) { w.onAniStart = f }

// SetAnimationEnabled toggles the transition animation.
func (w *TransitionStackedWidget) SetAnimationEnabled(isEnabled bool) {
	w.isAnimationEnabled = isEnabled
}

// IsAnimationEnabled reports whether the transition animation is enabled.
func (w *TransitionStackedWidget) IsAnimationEnabled() bool { return w.isAnimationEnabled }

// AddWidget adds a page.
func (w *TransitionStackedWidget) AddWidget(widget *qt.QWidget) int {
	widget.SetAttribute(qt.WA_TranslucentBackground)
	return w.QStackedWidget.AddWidget(widget)
}

// InsertWidget inserts a page at index.
func (w *TransitionStackedWidget) InsertWidget(index int, widget *qt.QWidget) int {
	widget.SetAttribute(qt.WA_TranslucentBackground)
	return w.QStackedWidget.InsertWidget(index, widget)
}

// SetCurrentWidget switches to the given page (duration/isBack are accepted for
// API compatibility but unused; the fade uses a fixed 150ms duration).
func (w *TransitionStackedWidget) SetCurrentWidget(widget *qt.QWidget, duration int, isBack bool) {
	w.SetCurrentIndex(w.IndexOf(widget), duration, isBack)
}

// SetCurrentIndex switches to the page at index (duration/isBack are accepted
// for API compatibility but unused), fading the new page in when enabled.
func (w *TransitionStackedWidget) SetCurrentIndex(index int, duration int, isBack bool) {
	if index < 0 || index >= w.Count() || index == w.CurrentIndex() {
		return
	}
	if !w.isAnimationEnabled {
		w.QStackedWidget.SetCurrentIndex(index)
		return
	}
	if w.onAniStart != nil {
		w.onAniStart()
	}
	w.QStackedWidget.SetCurrentIndex(index)
	w.fade.start(w.Widget(index), w.onAniFinished)
}

// EntranceTransitionStackedWidget is a transition stacked widget whose entrance
// transition is approximated by a fade-in in the Go port.
type EntranceTransitionStackedWidget struct {
	*TransitionStackedWidget
}

// NewEntranceTransitionStackedWidget builds an entrance transition widget.
func NewEntranceTransitionStackedWidget(parent *qt.QWidget) *EntranceTransitionStackedWidget {
	return &EntranceTransitionStackedWidget{TransitionStackedWidget: NewTransitionStackedWidget(parent)}
}

// DrillInTransitionStackedWidget is a transition stacked widget whose drill-in
// transition is approximated by a fade-in in the Go port.
type DrillInTransitionStackedWidget struct {
	*TransitionStackedWidget
}

// NewDrillInTransitionStackedWidget builds a drill-in transition widget.
func NewDrillInTransitionStackedWidget(parent *qt.QWidget) *DrillInTransitionStackedWidget {
	return &DrillInTransitionStackedWidget{TransitionStackedWidget: NewTransitionStackedWidget(parent)}
}
