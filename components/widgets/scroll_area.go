package widgets

import (
	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// ScrollArea is a smooth scroll area backed by a SmoothScrollDelegate.
type ScrollArea struct {
	*qt.QScrollArea
	scrollDelegate *SmoothScrollDelegate
}

// NewScrollArea builds a smooth scroll area.
func NewScrollArea(parent *qt.QWidget) *ScrollArea {
	w := &ScrollArea{QScrollArea: qt.NewQScrollArea(parent)}
	w.scrollDelegate = NewSmoothScrollDelegate(w.QAbstractScrollArea, false)
	return w
}

// ScrollDelegate returns the smooth scroll delegate that owns the fluent
// overlay scroll bars (mirrors the Python `scrollDelagate` attribute).
func (w *ScrollArea) ScrollDelegate() *SmoothScrollDelegate { return w.scrollDelegate }

// SetVerticalScrollBarPolicy hides the native vertical scroll bar and mirrors
// the policy to the fluent vertical scroll bar, matching the Python
// SmoothScrollDelegate.setVerticalScrollBarPolicy monkey-patch.
func (w *ScrollArea) SetVerticalScrollBarPolicy(policy qt.ScrollBarPolicy) {
	w.QScrollArea.SetVerticalScrollBarPolicy(qt.ScrollBarAlwaysOff)
	w.scrollDelegate.VerticalSmoothScrollBar().SetForceHidden(policy == qt.ScrollBarAlwaysOff)
}

// SetHorizontalScrollBarPolicy hides the native horizontal scroll bar and
// mirrors the policy to the fluent horizontal scroll bar (which would otherwise
// stay visible at the bottom whenever the scroll widget has a horizontal
// range). Matches the Python SmoothScrollDelegate.setHorizontalScrollBarPolicy
// monkey-patch.
func (w *ScrollArea) SetHorizontalScrollBarPolicy(policy qt.ScrollBarPolicy) {
	w.QScrollArea.SetHorizontalScrollBarPolicy(qt.ScrollBarAlwaysOff)
	w.scrollDelegate.HorizontalSmoothScrollBar().SetForceHidden(policy == qt.ScrollBarAlwaysOff)
}

// SetSmoothMode sets the smooth scroll mode for the given orientation.
func (w *ScrollArea) SetSmoothMode(mode common.SmoothMode, orientation qt.Orientation) {
	w.scrollDelegate.SetSmoothMode(mode, orientation)
}

// EnableTransparentBackground makes the scroll area background transparent.
func (w *ScrollArea) EnableTransparentBackground() {
	w.SetStyleSheet("QScrollArea{border: none; background: transparent}")
	// The QScrollArea viewport is a separate child widget that auto-fills its
	// background, so it stays opaque even after the frame and inner widget are
	// transparent. Make it transparent too, otherwise the bar keeps a solid
	// background (the reported tab bar opacity bug).
	if vp := w.Viewport(); vp != nil {
		vp.SetAutoFillBackground(false)
		vp.SetStyleSheet("background: transparent")
	}
	if w.Widget() != nil {
		w.Widget().SetStyleSheet("QWidget{background: transparent}")
	}
}

// SingleDirectionScrollArea is a scroll area that only scrolls in one
// direction using a common.SmoothScroll engine.
type SingleDirectionScrollArea struct {
	*qt.QScrollArea
	orient       qt.Orientation
	smoothScroll *common.SmoothScroll
	vScrollBar   *SmoothScrollBar
	hScrollBar   *SmoothScrollBar
}

// NewSingleDirectionScrollArea builds a single direction scroll area.
func NewSingleDirectionScrollArea(parent *qt.QWidget, orient qt.Orientation) *SingleDirectionScrollArea {
	w := &SingleDirectionScrollArea{
		QScrollArea: qt.NewQScrollArea(parent),
		orient:      orient,
	}
	w.smoothScroll = common.NewSmoothScroll(w.QAbstractScrollArea, orient)
	w.vScrollBar = NewSmoothScrollBar(qt.Vertical, w.QAbstractScrollArea)
	w.hScrollBar = NewSmoothScrollBar(qt.Horizontal, w.QAbstractScrollArea)
	w.installWheelEvent()
	return w
}

// SetVerticalScrollBarPolicy keeps the native vertical scroll bar hidden and
// toggles the fluent scroll bar visibility.
func (w *SingleDirectionScrollArea) SetVerticalScrollBarPolicy(policy qt.ScrollBarPolicy) {
	w.QScrollArea.SetVerticalScrollBarPolicy(qt.ScrollBarAlwaysOff)
	w.vScrollBar.SetForceHidden(policy == qt.ScrollBarAlwaysOff)
}

// SetHorizontalScrollBarPolicy keeps the native horizontal scroll bar hidden
// and toggles the fluent scroll bar visibility.
func (w *SingleDirectionScrollArea) SetHorizontalScrollBarPolicy(policy qt.ScrollBarPolicy) {
	w.QScrollArea.SetHorizontalScrollBarPolicy(qt.ScrollBarAlwaysOff)
	w.hScrollBar.SetForceHidden(policy == qt.ScrollBarAlwaysOff)
}

// SetSmoothMode sets the smooth scroll mode.
func (w *SingleDirectionScrollArea) SetSmoothMode(mode common.SmoothMode) {
	w.smoothScroll.SetSmoothMode(mode)
}

func (w *SingleDirectionScrollArea) installWheelEvent() {
	w.OnWheelEvent(func(super func(event *qt.QWheelEvent), event *qt.QWheelEvent) {
		ad := event.AngleDelta()
		if ad.X() != 0 {
			return
		}
		w.smoothScroll.HandleWheelEvent(event)
		event.SetAccepted(true)
	})
}

// EnableTransparentBackground makes the scroll area background transparent.
func (w *SingleDirectionScrollArea) EnableTransparentBackground() {
	w.SetStyleSheet("QScrollArea{border: none; background: transparent}")
	// The QScrollArea viewport is a separate child widget that auto-fills its
	// background, so it stays opaque even after the frame and inner widget are
	// transparent. Make it transparent too, otherwise the bar keeps a solid
	// background (the reported tab bar opacity bug).
	if vp := w.Viewport(); vp != nil {
		vp.SetAutoFillBackground(false)
		vp.SetStyleSheet("background: transparent")
	}
	if w.Widget() != nil {
		w.Widget().SetStyleSheet("QWidget{background: transparent}")
	}
}

// SmoothScrollArea is a smooth scroll area that uses animated scroll bars
// (the animation itself is simplified to immediate value changes).
type SmoothScrollArea struct {
	*qt.QScrollArea
	delegate *SmoothScrollDelegate
}

// NewSmoothScrollArea builds a smooth scroll area.
func NewSmoothScrollArea(parent *qt.QWidget) *SmoothScrollArea {
	w := &SmoothScrollArea{QScrollArea: qt.NewQScrollArea(parent)}
	w.delegate = NewSmoothScrollDelegate(w.QAbstractScrollArea, true)
	return w
}

// SetVerticalScrollBarPolicy hides the native vertical scroll bar and mirrors
// the policy to the fluent vertical scroll bar.
func (w *SmoothScrollArea) SetVerticalScrollBarPolicy(policy qt.ScrollBarPolicy) {
	w.QScrollArea.SetVerticalScrollBarPolicy(qt.ScrollBarAlwaysOff)
	w.delegate.VerticalSmoothScrollBar().SetForceHidden(policy == qt.ScrollBarAlwaysOff)
}

// SetHorizontalScrollBarPolicy hides the native horizontal scroll bar and
// mirrors the policy to the fluent horizontal scroll bar.
func (w *SmoothScrollArea) SetHorizontalScrollBarPolicy(policy qt.ScrollBarPolicy) {
	w.QScrollArea.SetHorizontalScrollBarPolicy(qt.ScrollBarAlwaysOff)
	w.delegate.HorizontalSmoothScrollBar().SetForceHidden(policy == qt.ScrollBarAlwaysOff)
}

// SetScrollAnimation sets the scroll animation duration for the orientation.
func (w *SmoothScrollArea) SetScrollAnimation(orient qt.Orientation, duration int) {
	var bar *SmoothScrollBar
	if orient == qt.Horizontal {
		bar = w.delegate.HorizontalSmoothScrollBar()
	} else {
		bar = w.delegate.VerticalSmoothScrollBar()
	}
	bar.SetScrollAnimation(duration)
}

// EnableTransparentBackground makes the scroll area background transparent.
func (w *SmoothScrollArea) EnableTransparentBackground() {
	w.SetStyleSheet("QScrollArea{border: none; background: transparent}")
	// The QScrollArea viewport is a separate child widget that auto-fills its
	// background, so it stays opaque even after the frame and inner widget are
	// transparent. Make it transparent too, otherwise the bar keeps a solid
	// background (the reported tab bar opacity bug).
	if vp := w.Viewport(); vp != nil {
		vp.SetAutoFillBackground(false)
		vp.SetStyleSheet("background: transparent")
	}
	if w.Widget() != nil {
		w.Widget().SetStyleSheet("QWidget{background: transparent}")
	}
}
