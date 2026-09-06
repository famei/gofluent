package widgets

import (
	"github.com/famei/gofluent/common"
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
	opacityAni      *qt.QPropertyAnimation
	shadowEffect    *qt.QGraphicsDropShadowEffect
	anchorWidget    *qt.QWidget
	anchorPosition  ToolTipPosition
}

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

	w.opacityAni = qt.NewQPropertyAnimation2(w.QObject, []byte("windowOpacity"))
	w.opacityAni.SetDuration(150)

	w.shadowEffect = qt.NewQGraphicsDropShadowEffect2(w.QObject)
	w.shadowEffect.SetBlurRadius(25)
	shadowColor := qt.NewQColor11(0, 0, 0, 50)
	w.shadowEffect.SetColor(shadowColor)
	shadowColor.Delete()
	w.shadowEffect.SetOffset2(0, 5)
	w.container.SetGraphicsEffect(w.shadowEffect.QGraphicsEffect)

	w.timer.SetSingleShot(true)
	w.timer.OnTimeout(func() {
		w.Hide()
	})

	w.SetAttribute(qt.WA_TransparentForMouseEvents)
	w.SetAttribute(qt.WA_TranslucentBackground)
	w.SetAttribute(qt.WA_ShowWithoutActivating)
	w.SetWindowFlags(qt.Tool | qt.FramelessWindowHint | qt.WindowStaysOnTopHint | qt.NoDropShadowWindowHint)
	w.setQss()

	w.OnShowEvent(func(super func(e *qt.QShowEvent), e *qt.QShowEvent) {
		start := qt.NewQVariant12(0)
		w.opacityAni.SetStartValue(start)
		start.Delete()
		end := qt.NewQVariant12(1)
		w.opacityAni.SetEndValue(end)
		end.Delete()
		w.opacityAni.Start()
		w.timer.Stop()
		if w.duration > 0 {
			w.timer.Start(w.duration + w.opacityAni.Duration())
		}
		super(e)
		// Re-anchor after the show event so the position is computed from the
		// tooltip's finalized size (the layout/QSS may not be fully polished
		// before the first show), keeping it anchored to its target widget.
		if w.anchorWidget != nil {
			pos := toolTipPosition(w, w.anchorWidget, w.anchorPosition)
			w.Move(pos.X(), pos.Y())
			pos.Delete()
		}
	})
	w.OnHideEvent(func(super func(e *qt.QHideEvent), e *qt.QHideEvent) {
		w.timer.Stop()
		super(e)
	})
	return w
}

func (w *ToolTip) createContainer() *qt.QFrame {
	return qt.NewQFrame(w.QWidget)
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
	pos := toolTipPosition(w, widget, position)
	w.Move(pos.X(), pos.Y())
	pos.Delete()
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

	screen := common.GetCurrentScreenGeometry(false)
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
		case qt.QEvent__Hide, qt.QEvent__Leave:
			w.hideToolTip()
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
