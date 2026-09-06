package navigation

import (
	"math"

	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// slideRectAnimator drives a squash-and-stretch (same-level) or cross-fade
// (level-change) transition of a QRectF, reproducing the WinUI 3 indicator
// slide used by QFluentWidgets' ScaleSlideAnimation.
//
// The rect passed to newSlideRectAnimator is mutated in place every frame; it
// is owned by the caller and must outlive the animation. onFrame (optional) is
// invoked after each mutation so the owner can repaint, and onDone (optional)
// fires once the transition completes naturally.
type slideRectAnimator struct {
	rect       *qt.QRectF
	horizontal bool
	onFrame    func()
	onDone     func()

	seg1 *common.ProgressAnimation
	seg2 *common.ProgressAnimation
}

func newSlideRectAnimator(rect *qt.QRectF, horizontal bool, onFrame, onDone func()) *slideRectAnimator {
	return &slideRectAnimator{rect: rect, horizontal: horizontal, onFrame: onFrame, onDone: onDone}
}

// start stops any in-flight transition and slides rect toward to.
func (a *slideRectAnimator) start(to *qt.QRectF, useCrossFade bool) {
	a.stop()

	sameLevel := false
	var from, toPos, dim float64
	if a.horizontal {
		sameLevel = math.Abs(a.rect.Y()-to.Y()) < 1
		from = a.rect.X()
		toPos = to.X()
		dim = a.rect.Width()
	} else {
		sameLevel = math.Abs(a.rect.X()-to.X()) < 1
		from = a.rect.Y()
		toPos = to.Y()
		dim = a.rect.Height()
	}

	if sameLevel && !useCrossFade {
		a.startSlide(from, toPos, dim, to)
	} else {
		a.startCrossFade(to)
	}
}

// startSlide implements the two-segment squash-and-stretch: the indicator first
// stretches to cover the distance between the two items, then travels and
// shrinks back to its resting length.
func (a *slideRectAnimator) startSlide(from, to, dim float64, end *qt.QRectF) {
	ex, ey, ew, eh := end.X(), end.Y(), end.Width(), end.Height()

	dist := math.Abs(to - from)
	mid := dist + dim
	forward := to > from

	// The Python ScaleSlideAnimation animates two independent properties:
	// `pos` uses moveTopLeft (translates the whole rect, preserving size) and
	// `length` uses setWidth/setHeight (keeps the left/top edge fixed). Using
	// SetX/SetY here would instead keep the opposite (right/bottom) edge fixed,
	// so during phase 2 the tail edge would slide backwards and the bar would
	// collapse at twice the intended rate (the reported "尾段反向扩展" glitch).
	// MoveLeft/MoveTop translate the rect without changing its size, matching
	// moveTopLeft on the animated axis.
	var setPos, setLen func(float64)
	if a.horizontal {
		setPos = func(v float64) { a.rect.MoveLeft(v) }
		setLen = func(v float64) { a.rect.SetWidth(v) }
	} else {
		setPos = func(v float64) { a.rect.MoveTop(v) }
		setLen = func(v float64) { a.rect.SetHeight(v) }
	}

	emit := func() {
		if a.onFrame != nil {
			a.onFrame()
		}
	}

	c1 := common.CreateBezierCurve(0.9, 0.1, 1, 0.2)
	a.seg1 = common.NewProgressAnimation(200, c1)
	c1.Delete()
	c2 := common.CreateBezierCurve(0.1, 0.9, 0.2, 1.0)
	a.seg2 = common.NewProgressAnimation(400, c2)
	c2.Delete()

	a.seg1.OnProgress(func(t float64) {
		setLen(dim + (mid-dim)*t)
		if forward {
			setPos(from)
		} else {
			setPos(from + (to-from)*t)
		}
		emit()
	})
	a.seg2.OnProgress(func(t float64) {
		setLen(mid + (dim-mid)*t)
		if forward {
			setPos(from + (to-from)*t)
		} else {
			setPos(to)
		}
		emit()
	})
	a.seg2.OnFinished(func() {
		a.rect.SetRect(ex, ey, ew, eh)
		emit()
		if a.onDone != nil {
			a.onDone()
		}
	})
	a.seg1.OnFinished(func() { a.seg2.Start() })
	a.seg1.Start()
}

// startCrossFade grows the indicator from the appropriate edge of the target
// rect (OutQuint, 600ms) when the two items are not on the same level.
func (a *slideRectAnimator) startCrossFade(end *qt.QRectF) {
	ex, ey, ew, eh := end.X(), end.Y(), end.Width(), end.Height()

	var startX, startY, startW, startH float64
	if a.horizontal {
		startY, startH = ey, eh
		if ex > a.rect.X() {
			startX = ex
		} else {
			startX = ex + ew
		}
		startW = 0
	} else {
		startX, startW = ex, ew
		if ey > a.rect.Y() {
			startY = ey
		} else {
			startY = ey + eh
		}
		startH = 0
	}

	a.rect.SetRect(startX, startY, startW, startH)
	if a.onFrame != nil {
		a.onFrame()
	}

	curve := qt.NewQEasingCurve3(qt.QEasingCurve__OutQuint)
	a.seg1 = common.NewProgressAnimation(600, curve)
	curve.Delete()

	a.seg1.OnProgress(func(t float64) {
		if a.horizontal {
			a.rect.SetX(startX + (ex-startX)*t)
			a.rect.SetWidth(ew * t)
		} else {
			a.rect.SetY(startY + (ey-startY)*t)
			a.rect.SetHeight(eh * t)
		}
		if a.onFrame != nil {
			a.onFrame()
		}
	})
	a.seg1.OnFinished(func() {
		a.rect.SetRect(ex, ey, ew, eh)
		if a.onFrame != nil {
			a.onFrame()
		}
		if a.onDone != nil {
			a.onDone()
		}
	})
	a.seg1.Start()
}

// stop cancels the transition and releases the underlying animations.
func (a *slideRectAnimator) stop() {
	if a.seg1 != nil {
		a.seg1.Stop()
		a.seg1.Delete()
		a.seg1 = nil
	}
	if a.seg2 != nil {
		a.seg2.Stop()
		a.seg2.Delete()
		a.seg2 = nil
	}
}
