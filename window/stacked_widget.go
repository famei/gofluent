package window

import (
	"github.com/famei/gofluent/components/widgets"
	qt "github.com/mappu/miqt/qt"
)

// StackedWidget is a frame that hosts a PopUpAniStackedWidget. The pop-up
// page-switch animation is dropped by the underlying widgets package, so pages
// switch synchronously; the animation-enabled flag and the currentChanged
// callback are preserved for API compatibility.
type StackedWidget struct {
	*qt.QFrame
	view               *widgets.PopUpAniStackedWidget
	hBoxLayout         *qt.QHBoxLayout
	isAnimationEnabled bool
}

// NewStackedWidget builds a window stacked widget.
func NewStackedWidget(parent *qt.QWidget) *StackedWidget {
	w := &StackedWidget{QFrame: qt.NewQFrame(parent)}
	// The fluent_window.qss `StackedWidget { background-color: rgba(255,255,255,0.5) }`
	// frosted overlay (and dark-mode tint) is matched via this objectName (the
	// selectorTranslation map rewrites StackedWidget -> QFrame#stackedWidget).
	w.SetObjectName("stackedWidget")
	w.hBoxLayout = qt.NewQHBoxLayout(w.QWidget)
	w.view = widgets.NewPopUpAniStackedWidget(w.QWidget)
	w.isAnimationEnabled = true

	w.hBoxLayout.SetContentsMargins(0, 0, 0, 0)
	w.hBoxLayout.AddWidget(w.view.QWidget)
	w.SetAttribute(qt.WA_StyledBackground)
	return w
}

// OnCurrentChanged registers the page-switch callback.
func (w *StackedWidget) OnCurrentChanged(f func(index int)) {
	w.view.QStackedWidget.OnCurrentChanged(f)
}

// IsAnimationEnabled reports whether the (dropped) pop-up animation is enabled.
func (w *StackedWidget) IsAnimationEnabled() bool { return w.isAnimationEnabled }

// SetAnimationEnabled toggles the (dropped) pop-up animation.
func (w *StackedWidget) SetAnimationEnabled(isEnabled bool) {
	w.isAnimationEnabled = isEnabled
	w.view.SetAnimationEnabled(isEnabled)
}

// AddWidget adds a page (deltaY defaults to 76 like the Python port).
func (w *StackedWidget) AddWidget(widget *qt.QWidget) {
	w.view.AddWidget(widget, 0, 76)
}

// RemoveWidget removes a page.
func (w *StackedWidget) RemoveWidget(widget *qt.QWidget) {
	w.view.RemoveWidget(widget)
}

// Widget returns the page at index.
func (w *StackedWidget) Widget(index int) *qt.QWidget { return w.view.Widget(index) }

// SetCurrentWidget switches to the given page. The pop-out animation and the
// QAbstractScrollArea scrollbar reset are dropped in the Go port.
func (w *StackedWidget) SetCurrentWidget(widget *qt.QWidget, popOut bool) {
	w.view.SetCurrentWidget(widget)
}

// SetCurrentIndex switches to the page at index (popOut accepted for API
// compatibility).
func (w *StackedWidget) SetCurrentIndex(index int, popOut bool) {
	w.view.SetCurrentIndex(index)
}

// CurrentIndex returns the current page index.
func (w *StackedWidget) CurrentIndex() int { return w.view.CurrentIndex() }

// CurrentWidget returns the current page.
func (w *StackedWidget) CurrentWidget() *qt.QWidget { return w.view.CurrentWidget() }

// IndexOf returns the index of widget.
func (w *StackedWidget) IndexOf(widget *qt.QWidget) int { return w.view.IndexOf(widget) }

// Count returns the number of pages.
func (w *StackedWidget) Count() int { return w.view.Count() }

// QStackedWidget exposes the underlying QStackedWidget for the router.
func (w *StackedWidget) QStackedWidget() *qt.QStackedWidget { return w.view.QStackedWidget }
