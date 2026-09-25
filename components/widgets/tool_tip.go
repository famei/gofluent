package widgets

import (
	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/internal/platform"
	qt "github.com/mappu/miqt/qt"
)

// ToolTipPosition is the position of a tooltip relative to its target widget.
type ToolTipPosition int

const (
	ToolTipPositionTop ToolTipPosition = iota
	ToolTipPositionBottom
	ToolTipPositionLeft
	ToolTipPositionRight
	ToolTipPositionTopLeft
	ToolTipPositionTopRight
	ToolTipPositionBottomLeft
	ToolTipPositionBottomRight
)

// ItemViewToolTipType is the target item-view kind for an item-view tooltip.
type ItemViewToolTipType int

const (
	ItemViewToolTipList ItemViewToolTipType = iota
	ItemViewToolTipTable
)

// ToolTip is a fluent-styled tooltip window (QFrame based).
type ToolTip struct {
	*qt.QFrame
	text            string
	duration        int
	container       *qt.QFrame
	timer           *qt.QTimer
	containerLayout *qt.QHBoxLayout
	label           *qt.QLabel
	shadowEffect    *qt.QGraphicsDropShadowEffect
	anchorWidget    *qt.QWidget
	anchorPosition  ToolTipPosition

	// anchorFilter corrects the position when the window manager moves the window behind
	// our back after showing it (see installAnchorGuard).
	anchorFilter *qt.QObject
	// anchorCorrected limits that correction to one per show.
	anchorCorrected bool
}

// toolTipFadeDuration is the length of the fade-in of a tooltip, in milliseconds.
const toolTipFadeDuration = 150

// NewToolTip builds a tooltip.
func NewToolTip(text string, parent *qt.QWidget) *ToolTip {
	w := &ToolTip{QFrame: qt.NewQFrame(parent)}
	w.text = text
	w.duration = 1000

	w.container = w.createContainer()
	w.timer = qt.NewQTimer2(w.QObject)
	w.SetLayout(qt.NewQHBoxLayout2().QLayout)
	w.containerLayout = qt.NewQHBoxLayout(w.container.QWidget)
	w.label = qt.NewQLabel5(text, w.QWidget)

	w.Layout().SetContentsMargins(12, 8, 12, 12)
	w.Layout().AddWidget(w.container.QWidget)
	w.containerLayout.AddWidget(w.label.QWidget)
	w.containerLayout.SetContentsMargins(8, 6, 8, 6)

	w.installShadowEffect()

	w.timer.SetSingleShot(true)
	w.timer.OnTimeout(func() {
		w.Hide()
	})

	w.SetAttribute(qt.WA_TransparentForMouseEvents)
	w.SetAttribute(qt.WA_TranslucentBackground)
	w.SetAttribute(qt.WA_ShowWithoutActivating)
	w.SetWindowFlags(platform.FloatingWindowFlags() | qt.NoDropShadowWindowHint)
	w.setQss()

	w.OnShowEvent(func(super func(e *qt.QShowEvent), e *qt.QShowEvent) {
		// The fade is a graphics effect on the container instead of the window opacity the
		// Python original animates: windowOpacity needs a compositor that honours
		// _NET_WM_WINDOW_OPACITY (WSLg's Weston does not, which is why the tooltip used to
		// pop in without any animation), and a graphics effect on a top level window is
		// never repainted on X11. The container carries the drop shadow as its graphics
		// effect and Qt gives a widget only one, so the shadow comes back when the fade ends.
		common.FadeInThen(w.container.QWidget, toolTipFadeDuration, qt.NewQEasingCurve3(qt.QEasingCurve__InSine),
			func() { w.installShadowEffect() })
		w.timer.Stop()
		if w.duration > 0 {
			w.timer.Start(w.duration + toolTipFadeDuration)
		}
		super(e)
		// Re-anchor after the show event so the position is computed from the
		// tooltip's finalized size (the layout/QSS may not be fully polished
		// before the first show), keeping it anchored to its target widget ...
		w.anchorCorrected = false
		w.reanchor()
		// ... and again once the window is mapped: a position set before the map can be
		// overridden by the window manager (see common.ApplyAfterMap).
		common.ApplyAfterMap(w.QWidget, w.reanchor)
	})
	w.OnHideEvent(func(super func(e *qt.QHideEvent), e *qt.QHideEvent) {
		w.timer.Stop()
		super(e)
	})
	w.installAnchorGuard()
	return w
}

// installAnchorGuard re-applies the anchor when the window manager moves the tooltip behind
// our back right after showing it. WSLg's XWM (Weston) places a window that is mapped again
// one frame margin off the position Qt gave it and does not tell Qt, so a tooltip that is
// hovered over and over walks towards the top left by (-32,-32) per hover. The placement does
// arrive as a move event, which is the signal used here; one correction per show is enough,
// and it keeps the tooltip from fighting a window manager that moves the popup for a reason.
func (w *ToolTip) installAnchorGuard() {
	w.anchorFilter = qt.NewQObject2(w.QObject)
	w.InstallEventFilter(w.anchorFilter)
	w.anchorFilter.OnEventFilter(func(super func(watched *qt.QObject, event *qt.QEvent) bool, watched *qt.QObject, event *qt.QEvent) bool {
		if event.Type() == qt.QEvent__Move && !w.anchorCorrected && w.IsVisible() && w.anchorWidget != nil {
			w.anchorCorrected = true
			w.reanchor()
		}
		return super(watched, event)
	})
}

// reanchor moves the tooltip so that its content sits where the target widget asks for (see
// common.MoveWindowContentTo: move() would place the window frame there instead, which a
// window manager that decorates the window turns into an offset).
func (w *ToolTip) reanchor() {
	if w.anchorWidget == nil {
		return
	}
	pos := toolTipPosition(w, w.anchorWidget, w.anchorPosition)
	common.MoveWindowContentTo(w.QWidget, pos.X(), pos.Y())
	pos.Delete()
}

func (w *ToolTip) createContainer() *qt.QFrame {
	return qt.NewQFrame(w.QWidget)
}

// installShadowEffect puts the drop shadow on the container. The fade-in replaces it with
// an opacity effect (a widget has room for one graphics effect only) and QWidget deletes the
// effect it replaces, so a fresh shadow effect is built every time instead of reusing the
// previous one.
func (w *ToolTip) installShadowEffect() {
	w.shadowEffect = qt.NewQGraphicsDropShadowEffect2(w.QObject)
	w.shadowEffect.SetBlurRadius(25)
	shadowColor := qt.NewQColor11(0, 0, 0, 50)
	w.shadowEffect.SetColor(shadowColor)
	shadowColor.Delete()
	w.shadowEffect.SetOffset2(0, 5)
	w.container.SetGraphicsEffect(w.shadowEffect.QGraphicsEffect)
}

// Text returns the tooltip text.
func (w *ToolTip) Text() string { return w.text }

// SetText sets the tooltip text.
func (w *ToolTip) SetText(text string) {
	w.text = text
	w.label.SetText(text)
	w.container.AdjustSize()
	w.AdjustSize()
}

// Duration returns the auto-hide duration in milliseconds.
func (w *ToolTip) Duration() int { return w.duration }

// SetDuration sets the auto-hide duration; duration <= 0 disables auto-hide.
func (w *ToolTip) SetDuration(duration int) { w.duration = duration }

func (w *ToolTip) setQss() {
	w.container.SetObjectName("container")
	w.label.SetObjectName("contentLabel")
	common.FluentStyleSheet(common.FluentToolTip).Apply(w.QWidget, common.ThemeAuto)
	w.label.AdjustSize()
	w.AdjustSize()
}

// AdjustPos moves the tooltip relative to widget at position.
func (w *ToolTip) AdjustPos(widget *qt.QWidget, position ToolTipPosition) {
	w.anchorWidget = widget
	w.anchorPosition = position
	w.reanchor()
}

// toolTipPosition computes the screen-clamped position of the tooltip relative
// to the target widget (inline port of the Python ToolTipPositionManager set).
func toolTipPosition(tooltip *ToolTip, parent *qt.QWidget, position ToolTipPosition) *qt.QPoint {
	origin := qt.NewQPoint2(0, 0)
	base := parent.MapToGlobal(origin) // GoGC-armed — do NOT Delete
	origin.Delete()

	x, y := base.X(), base.Y()
	margins := tooltip.Layout().ContentsMargins()

	switch position {
	case ToolTipPositionTop:
		x += parent.Width()/2 - tooltip.Width()/2
		y -= tooltip.Height()
	case ToolTipPositionBottom:
		x += parent.Width()/2 - tooltip.Width()/2
		y += parent.Height()
	case ToolTipPositionLeft:
		x -= tooltip.Width()
		y += (parent.Height() - tooltip.Height()) / 2
	case ToolTipPositionRight:
		x += parent.Width()
		y += (parent.Height() - tooltip.Height()) / 2
	case ToolTipPositionTopRight:
		x += parent.Width() - tooltip.Width() + margins.Right()
		y -= tooltip.Height()
	case ToolTipPositionTopLeft:
		x -= margins.Left()
		y -= tooltip.Height()
	case ToolTipPositionBottomRight:
		x += parent.Width() - tooltip.Width() + margins.Right()
		y += parent.Height()
	case ToolTipPositionBottomLeft:
		x -= margins.Left()
		y += parent.Height()
	}

	screen := common.GetWidgetScreenGeometry(parent, false)
	if screen != nil {
		// screen is GoGC-armed (QScreen.Geometry) — do NOT Delete
		if x < screen.Left() {
			x = screen.Left()
		}
		if x > screen.Right()-tooltip.Width()-4 {
			x = screen.Right() - tooltip.Width() - 4
		}
		if y < screen.Top() {
			y = screen.Top()
		}
		if y > screen.Bottom()-tooltip.Height()-4 {
			y = screen.Bottom() - tooltip.Height() - 4
		}
	}
	return qt.NewQPoint2(x, y)
}

// ToolTipFilter installs a delayed tooltip on a widget.
type ToolTipFilter struct {
	*qt.QObject
	isEnter      bool
	tooltip      *ToolTip
	tooltipDelay int
	position     ToolTipPosition
	timer        *qt.QTimer
	parentWidget *qt.QWidget
}

// NewToolTipFilter builds a tooltip filter for parent.
func NewToolTipFilter(parent *qt.QWidget, showDelay int, position ToolTipPosition) *ToolTipFilter {
	w := &ToolTipFilter{QObject: qt.NewQObject2(parent.QObject)}
	w.isEnter = false
	w.tooltipDelay = showDelay
	w.position = position
	w.parentWidget = parent
	w.timer = qt.NewQTimer2(w.QObject)
	w.timer.SetSingleShot(true)
	w.timer.OnTimeout(func() {
		w.showToolTip()
	})

	parent.InstallEventFilter(w.QObject)
	w.OnEventFilter(func(super func(watched *qt.QObject, event *qt.QEvent) bool, watched *qt.QObject, event *qt.QEvent) bool {
		switch event.Type() {
		case qt.QEvent__ToolTip:
			return true
		case qt.QEvent__Hide:
			w.hideToolTip()
		case qt.QEvent__Leave:
			// A foreign window can sit on top of the cursor and steal the pointer - a
			// window manager frame drawn around the tooltip window, or WSLg's presentation
			// wrapper - and Qt then reports a Leave although the cursor never left the
			// widget. Hiding on that would kill the tooltip the moment it appears, so the
			// cursor position decides.
			if !w.cursorOverParent() {
				w.hideToolTip()
			}
		case qt.QEvent__Enter:
			w.isEnter = true
			if w.canShowToolTip() {
				if w.tooltip == nil {
					w.tooltip = w.createToolTip()
				}
				duration := -1
				if w.parentWidget.ToolTipDuration() > 0 {
					duration = w.parentWidget.ToolTipDuration()
				}
				w.tooltip.SetDuration(duration)
				w.timer.Start(w.tooltipDelay)
			}
		case qt.QEvent__MouseButtonPress:
			w.hideToolTip()
		}
		return super(watched, event)
	})
	return w
}

func (w *ToolTipFilter) createToolTip() *ToolTip {
	return NewToolTip(w.parentWidget.ToolTip(), w.parentWidget.Window())
}

// HideToolTip hides the tooltip and stops the show timer.
func (w *ToolTipFilter) HideToolTip() { w.hideToolTip() }

// cursorOverParent reports whether the mouse cursor is still inside the filtered widget,
// which is what decides whether a Leave event really means the pointer moved away.
func (w *ToolTipFilter) cursorOverParent() bool {
	if w.parentWidget == nil || !w.parentWidget.IsVisible() {
		return false
	}
	origin := qt.NewQPoint2(0, 0)
	topLeft := w.parentWidget.MapToGlobal(origin)
	origin.Delete()
	pos := qt.QCursor_Pos() // GoGC-armed — do NOT Delete
	return pos.X() >= topLeft.X() && pos.X() < topLeft.X()+w.parentWidget.Width() &&
		pos.Y() >= topLeft.Y() && pos.Y() < topLeft.Y()+w.parentWidget.Height()
}

func (w *ToolTipFilter) hideToolTip() {
	w.isEnter = false
	w.timer.Stop()
	if w.tooltip != nil {
		w.tooltip.Hide()
	}
}

func (w *ToolTipFilter) showToolTip() {
	if !w.isEnter {
		return
	}
	w.tooltip.SetText(w.parentWidget.ToolTip())
	w.tooltip.AdjustPos(w.parentWidget, w.position)
	w.tooltip.Show()
}

// SetToolTipDelay sets the show delay in milliseconds.
func (w *ToolTipFilter) SetToolTipDelay(delay int) { w.tooltipDelay = delay }

func (w *ToolTipFilter) canShowToolTip() bool {
	return w.parentWidget.IsWidgetType() && w.parentWidget.ToolTip() != "" && w.parentWidget.IsEnabled()
}
