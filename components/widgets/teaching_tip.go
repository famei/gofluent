package widgets

import (
	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/internal/platform"
	qt "github.com/mappu/miqt/qt"
)

// TeachingTipTailPosition enumerates where the teaching tip tail points.
type TeachingTipTailPosition int

const (
	TeachingTipTailTop TeachingTipTailPosition = iota
	TeachingTipTailBottom
	TeachingTipTailLeft
	TeachingTipTailRight
	TeachingTipTailTopLeft
	TeachingTipTailTopRight
	TeachingTipTailBottomLeft
	TeachingTipTailBottomRight
	TeachingTipTailLeftTop
	TeachingTipTailLeftBottom
	TeachingTipTailRightTop
	TeachingTipTailRightBottom
	TeachingTipTailNone
)

// teachingTipImagePosition enumerates where the image is placed inside a tip.
type teachingTipImagePosition int

const (
	teachingTipImageTop teachingTipImagePosition = iota
	teachingTipImageBottom
	teachingTipImageLeft
	teachingTipImageRight
)

// tipTailManager inlines the TeachingTipManager logic for a tail position.
type tipTailManager struct {
	tailPosition TeachingTipTailPosition
}

func newTipTailManager(position TeachingTipTailPosition) *tipTailManager {
	return &tipTailManager{tailPosition: position}
}

// layoutMargins returns the (left, top, right, bottom) margins of the bubble.
func (m *tipTailManager) layoutMargins() (int, int, int, int) {
	switch m.tailPosition {
	case TeachingTipTailTop, TeachingTipTailTopLeft, TeachingTipTailTopRight:
		return 0, 8, 0, 0
	case TeachingTipTailBottom, TeachingTipTailBottomLeft, TeachingTipTailBottomRight:
		return 0, 0, 0, 8
	case TeachingTipTailLeft, TeachingTipTailLeftTop, TeachingTipTailLeftBottom:
		return 8, 0, 0, 0
	case TeachingTipTailRight, TeachingTipTailRightTop, TeachingTipTailRightBottom:
		return 0, 0, 8, 0
	default:
		return 0, 0, 0, 0
	}
}

func (m *tipTailManager) imagePosition() teachingTipImagePosition {
	switch m.tailPosition {
	case TeachingTipTailTop, TeachingTipTailTopLeft, TeachingTipTailTopRight,
		TeachingTipTailLeftTop, TeachingTipTailRightTop:
		return teachingTipImageBottom
	case TeachingTipTailLeft:
		return teachingTipImageRight
	case TeachingTipTailRight:
		return teachingTipImageLeft
	default:
		return teachingTipImageTop
	}
}

// draw paints the rounded bubble with its triangular tail.
func (m *tipTailManager) draw(tip *TeachTipBubble, painter *qt.QPainter) {
	w := float64(tip.Width())
	h := float64(tip.Height())

	path := qt.NewQPainterPath()
	defer path.Delete()

	switch m.tailPosition {
	case TeachingTipTailTop:
		path.AddRoundedRect2(1, 8, w-2, h-9, 8, 8)
		path.MoveTo2(w/2-7, 8)
		path.LineTo2(w/2, 1)
		path.LineTo2(w/2+7, 8)
		path.CloseSubpath()
	case TeachingTipTailTopLeft:
		path.AddRoundedRect2(1, 8, w-2, h-9, 8, 8)
		path.MoveTo2(20, 8)
		path.LineTo2(27, 1)
		path.LineTo2(34, 8)
		path.CloseSubpath()
	case TeachingTipTailTopRight:
		path.AddRoundedRect2(1, 8, w-2, h-9, 8, 8)
		path.MoveTo2(w-20, 8)
		path.LineTo2(w-27, 1)
		path.LineTo2(w-34, 8)
		path.CloseSubpath()
	case TeachingTipTailBottom:
		path.AddRoundedRect2(1, 1, w-2, h-9, 8, 8)
		path.MoveTo2(w/2-7, h-8)
		path.LineTo2(w/2, h-1)
		path.LineTo2(w/2+7, h-8)
		path.CloseSubpath()
	case TeachingTipTailBottomLeft:
		path.AddRoundedRect2(1, 1, w-2, h-9, 8, 8)
		path.MoveTo2(20, h-8)
		path.LineTo2(27, h-1)
		path.LineTo2(34, h-8)
		path.CloseSubpath()
	case TeachingTipTailBottomRight:
		path.AddRoundedRect2(1, 1, w-2, h-9, 8, 8)
		path.MoveTo2(w-20, h-8)
		path.LineTo2(w-27, h-1)
		path.LineTo2(w-34, h-8)
		path.CloseSubpath()
	case TeachingTipTailLeft:
		path.AddRoundedRect2(8, 1, w-10, h-2, 8, 8)
		path.MoveTo2(8, h/2-7)
		path.LineTo2(1, h/2)
		path.LineTo2(8, h/2+7)
		path.CloseSubpath()
	case TeachingTipTailLeftTop:
		path.AddRoundedRect2(8, 1, w-10, h-2, 8, 8)
		path.MoveTo2(8, 10)
		path.LineTo2(1, 17)
		path.LineTo2(8, 24)
		path.CloseSubpath()
	case TeachingTipTailLeftBottom:
		path.AddRoundedRect2(9, 1, w-10, h-2, 8, 8)
		path.MoveTo2(9, h-10)
		path.LineTo2(1, h-17)
		path.LineTo2(9, h-24)
		path.CloseSubpath()
	case TeachingTipTailRight:
		path.AddRoundedRect2(1, 1, w-9, h-2, 8, 8)
		path.MoveTo2(w-8, h/2-7)
		path.LineTo2(w-1, h/2)
		path.LineTo2(w-8, h/2+7)
		path.CloseSubpath()
	case TeachingTipTailRightTop:
		path.AddRoundedRect2(1, 1, w-9, h-2, 8, 8)
		path.MoveTo2(w-8, 10)
		path.LineTo2(w-1, 17)
		path.LineTo2(w-8, 24)
		path.CloseSubpath()
	case TeachingTipTailRightBottom:
		path.AddRoundedRect2(1, 1, w-9, h-2, 8, 8)
		path.MoveTo2(w-8, h-10)
		path.LineTo2(w-1, h-17)
		path.LineTo2(w-8, h-24)
		path.CloseSubpath()
	default: // TeachingTipTailNone
		path.AddRoundedRect2(1, 1, w-2, h-2, 8, 8)
	}

	simplified := path.Simplified() // GoGC-armed — do NOT Delete
	painter.DrawPath(simplified)
}

// pos computes the global position of the tip relative to its target.
func (m *tipTailManager) pos(tip *TeachingTip) *qt.QPoint {
	target := tip.target
	sh := tip.SizeHint()
	tw := sh.Width()
	th := sh.Height()
	margins := tip.Layout().ContentsMargins()

	viewSh := tip.View().SizeHint()

	vh := viewSh.Height()

	switch m.tailPosition {
	case TeachingTipTailTop:
		p := qt.NewQPoint2(0, target.Height())
		pos := target.MapToGlobal(p) // GoGC-armed — do NOT Delete
		p.Delete()
		return qt.NewQPoint2(pos.X()+target.Width()/2-tw/2, pos.Y()-margins.Top())
	case TeachingTipTailBottom:
		p := qt.NewQPoint2(0, 0)
		pos := target.MapToGlobal(p) // GoGC-armed — do NOT Delete
		p.Delete()
		return qt.NewQPoint2(pos.X()+target.Width()/2-tw/2, pos.Y()-th+margins.Bottom())
	case TeachingTipTailLeft:
		p := qt.NewQPoint2(target.Width(), 0)
		pos := target.MapToGlobal(p) // GoGC-armed — do NOT Delete
		p.Delete()
		return qt.NewQPoint2(pos.X()-margins.Left(), pos.Y()-vh/2+target.Height()/2-margins.Top())
	case TeachingTipTailRight:
		p := qt.NewQPoint2(0, 0)
		pos := target.MapToGlobal(p) // GoGC-armed — do NOT Delete
		p.Delete()
		return qt.NewQPoint2(pos.X()-tw+margins.Right(), pos.Y()-vh/2+target.Height()/2-margins.Top())
	case TeachingTipTailTopLeft:
		p := qt.NewQPoint2(0, target.Height())
		pos := target.MapToGlobal(p) // GoGC-armed — do NOT Delete
		p.Delete()
		return qt.NewQPoint2(pos.X()-margins.Left(), pos.Y()-margins.Top())
	case TeachingTipTailTopRight:
		p := qt.NewQPoint2(target.Width(), target.Height())
		pos := target.MapToGlobal(p) // GoGC-armed — do NOT Delete
		p.Delete()
		return qt.NewQPoint2(pos.X()-tw+margins.Left(), pos.Y()-margins.Top())
	case TeachingTipTailBottomLeft:
		p := qt.NewQPoint2(0, 0)
		pos := target.MapToGlobal(p) // GoGC-armed — do NOT Delete
		p.Delete()
		return qt.NewQPoint2(pos.X()-margins.Left(), pos.Y()-th+margins.Bottom())
	case TeachingTipTailBottomRight:
		p := qt.NewQPoint2(target.Width(), 0)
		pos := target.MapToGlobal(p) // GoGC-armed — do NOT Delete
		p.Delete()
		return qt.NewQPoint2(pos.X()-tw+margins.Left(), pos.Y()-th+margins.Bottom())
	case TeachingTipTailLeftTop:
		p := qt.NewQPoint2(target.Width(), 0)
		pos := target.MapToGlobal(p) // GoGC-armed — do NOT Delete
		p.Delete()
		return qt.NewQPoint2(pos.X()-margins.Left(), pos.Y()-margins.Top())
	case TeachingTipTailLeftBottom:
		p := qt.NewQPoint2(target.Width(), target.Height())
		pos := target.MapToGlobal(p) // GoGC-armed — do NOT Delete
		p.Delete()
		return qt.NewQPoint2(pos.X()-margins.Left(), pos.Y()-th+margins.Bottom())
	case TeachingTipTailRightTop:
		p := qt.NewQPoint2(0, 0)
		pos := target.MapToGlobal(p) // GoGC-armed — do NOT Delete
		p.Delete()
		return qt.NewQPoint2(pos.X()-tw+margins.Right(), pos.Y()-margins.Top())
	case TeachingTipTailRightBottom:
		p := qt.NewQPoint2(0, target.Height())
		pos := target.MapToGlobal(p) // GoGC-armed — do NOT Delete
		p.Delete()
		return qt.NewQPoint2(pos.X()-tw+margins.Right(), pos.Y()-th+margins.Bottom())
	default: // TeachingTipTailNone
		return qt.NewQPoint2(tip.X(), tip.Y())
	}
}

// position clamps the raw position to the current screen.
func (m *tipTailManager) position(tip *TeachingTip) *qt.QPoint {
	p := m.pos(tip)
	defer p.Delete()
	g := common.GetCurrentScreenGeometry(false)
	if g == nil {
		return qt.NewQPoint2(p.X(), p.Y())
	}
	// g is GoGC-armed (QScreen.Geometry) — do NOT Delete

	x := p.X()
	if x < g.Left() {
		x = g.Left()
	}
	if x > g.Right()-tip.Width()-4 {
		x = g.Right() - tip.Width() - 4
	}
	y := p.Y()
	if y < g.Top() {
		y = g.Top()
	}
	if y > g.Bottom()-tip.Height()-4 {
		y = g.Bottom() - tip.Height() - 4
	}
	return qt.NewQPoint2(x, y)
}

// TeachingTipView is a flyout view whose image layout follows the tail.
type TeachingTipView struct {
	*FlyoutView
	manager    *tipTailManager
	hBoxLayout *qt.QHBoxLayout
}

// NewTeachingTipView builds a teaching tip view.
func NewTeachingTipView(title, content string, icon, image interface{}, isClosable bool, tailPosition TeachingTipTailPosition, parent *qt.QWidget) *TeachingTipView {
	w := &TeachingTipView{
		manager:    newTipTailManager(tailPosition),
		hBoxLayout: qt.NewQHBoxLayout2(),
	}
	w.hBoxLayout.SetContentsMargins(0, 0, 0, 0)
	w.FlyoutView = NewFlyoutView(title, content, icon, image, isClosable, parent)
	w.FlyoutView.SetDrawBackground(false)
	w.applyImageLayout()
	return w
}

func (w *TeachingTipView) applyImageLayout() {
	pos := w.manager.imagePosition()
	w.imageLabel.SetHidden(w.imageLabel.IsNull())

	switch pos {
	case teachingTipImageBottom:
		w.vBoxLayout.RemoveWidget(w.imageLabel.QWidget)
		w.imageLabel.SetBorderRadius(0, 0, 8, 8)
		w.vBoxLayout.AddWidget(w.imageLabel.QWidget)
	default: // top / left / right fall back to the top placement
		w.imageLabel.SetBorderRadius(8, 8, 0, 0)
	}

	if pos == teachingTipImageTop || pos == teachingTipImageBottom {
		sh := w.vBoxLayout.SizeHint()
		w.imageLabel.ScaledToWidth(sh.Width() - 2)
	} else {
		sh := w.vBoxLayout.SizeHint()
		w.imageLabel.ScaledToHeight(sh.Height() - 2)
	}
}

// TeachTipBubble is the rounded bubble that hosts a teaching tip view.
type TeachTipBubble struct {
	*qt.QWidget
	manager    *tipTailManager
	hBoxLayout *qt.QHBoxLayout
	view       *FlyoutViewBase
}

// NewTeachTipBubble builds a bubble for the given view and tail position.
func NewTeachTipBubble(view *FlyoutViewBase, tailPosition TeachingTipTailPosition, parent *qt.QWidget) *TeachTipBubble {
	w := &TeachTipBubble{
		QWidget: qt.NewQWidget(parent),
		manager: newTipTailManager(tailPosition),
		view:    view,
	}
	w.hBoxLayout = qt.NewQHBoxLayout(w.QWidget)
	l, t, r, b := w.manager.layoutMargins()
	w.hBoxLayout.SetContentsMargins(l, t, r, b)
	w.hBoxLayout.AddWidget(view.QWidget)
	w.installPaintEvent()
	return w
}

// SetView replaces the hosted view.
func (w *TeachTipBubble) SetView(view *FlyoutViewBase) {
	w.hBoxLayout.RemoveWidget(w.view.QWidget)
	w.view.DeleteLater()
	w.view = view
	w.hBoxLayout.AddWidget(view.QWidget)
}

func (w *TeachTipBubble) installPaintEvent() {
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
		super(event)
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing)

		var bg, border *qt.QColor
		if common.IsDarkTheme() {
			bg = qt.NewQColor3(40, 40, 40)
			border = qt.NewQColor3(23, 23, 23)
		} else {
			bg = qt.NewQColor3(248, 248, 248)
			border = qt.NewQColor11(0, 0, 0, 17)
		}
		defer bg.Delete()
		defer border.Delete()
		brush := qt.NewQBrush3(bg)
		defer brush.Delete()
		painter.SetBrush(brush)
		painter.SetPen(border)

		w.manager.draw(w, painter)
		painter.End()
	})
}

// TeachingTip is a pop-up teaching tip anchored to a target widget.
type TeachingTip struct {
	*qt.QWidget
	target          *qt.QWidget
	duration        int
	isDeleteOnClose bool
	manager         *tipTailManager
	hBoxLayout      *qt.QHBoxLayout
	bubble          *TeachTipBubble
	shadowEffect    *qt.QGraphicsDropShadowEffect
	fadeOutTimer    *qt.QTimer
	onClosed        func()
}

// NewTeachingTip builds a teaching tip.
func NewTeachingTip(view *FlyoutViewBase, target *qt.QWidget, duration int, tailPosition TeachingTipTailPosition, parent *qt.QWidget, isDeleteOnClose bool) *TeachingTip {
	w := &TeachingTip{
		QWidget:         qt.NewQWidget(parent),
		target:          target,
		duration:        duration,
		isDeleteOnClose: isDeleteOnClose,
		manager:         newTipTailManager(tailPosition),
	}
	w.hBoxLayout = qt.NewQHBoxLayout(w.QWidget)
	w.bubble = NewTeachTipBubble(view, tailPosition, w.QWidget)
	w.hBoxLayout.SetContentsMargins(15, 8, 15, 20)
	w.hBoxLayout.AddWidget(w.bubble.QWidget)
	w.SetShadowEffect(35, 0, 8)

	w.SetAttribute(qt.WA_TranslucentBackground)
	// A tip must never take the focus from the window it explains (the Qt::Tool window it
	// replaces on X11 does not either).
	w.SetAttribute(qt.WA_ShowWithoutActivating)
	w.SetWindowFlags(platform.FloatingWindowFlags())

	w.fadeOutTimer = qt.NewQTimer2(w.QObject)
	w.fadeOutTimer.SetSingleShot(true)
	w.fadeOutTimer.OnTimeout(w.fadeOut)
	w.installEvents()
	return w
}

// SetShadowEffect adds a drop shadow behind the bubble.
func (w *TeachingTip) SetShadowEffect(blurRadius float64, dx, dy float64) {
	alpha := 30
	if common.IsDarkTheme() {
		alpha = 80
	}
	color := qt.NewQColor11(0, 0, 0, alpha)
	defer color.Delete()
	w.shadowEffect = qt.NewQGraphicsDropShadowEffect2(w.bubble.QObject)
	w.shadowEffect.SetBlurRadius(blurRadius)
	w.shadowEffect.SetOffset2(dx, dy)
	w.shadowEffect.SetColor(color)
	w.bubble.SetGraphicsEffect(nil)
	w.bubble.SetGraphicsEffect(w.shadowEffect.QGraphicsEffect)
}

// View returns the hosted flyout view.
func (w *TeachingTip) View() *FlyoutViewBase { return w.bubble.view }

// SetView replaces the hosted view.
func (w *TeachingTip) SetView(view *FlyoutViewBase) { w.bubble.SetView(view) }

// OnClosed registers a callback emitted when the tip closes.
func (w *TeachingTip) OnClosed(f func()) { w.onClosed = f }

// adjustPosition moves the tip so that its content sits where the target widget asks for (see
// common.MoveWindowContentTo: move() would place the window frame there instead, which a
// window manager that decorates the window turns into an offset).
func (w *TeachingTip) adjustPosition() {
	pos := w.manager.position(w)
	common.MoveWindowContentTo(w.QWidget, pos.X(), pos.Y())
	pos.Delete()
}

// fadeIn fades the bubble in over 167ms.
//
// The animation is a graphics effect on the bubble, not the window opacity the Python
// original animates: windowOpacity only works where the compositor honours
// _NET_WM_WINDOW_OPACITY (WSLg's Weston does not, so the tip appeared out of nowhere), and
// a graphics effect on a top level window is never repainted on X11.
func (w *TeachingTip) fadeIn() {
	// The bubble carries the drop shadow as its graphics effect and Qt gives a widget only
	// one effect, so the shadow is installed again when the fade ends.
	common.FadeInThen(w.bubble.QWidget, 167, qt.NewQEasingCurve3(qt.QEasingCurve__InSine),
		func() { w.SetShadowEffect(35, 0, 8) })
}

// fadeOut fades the bubble out over 167ms and then closes the tip.
func (w *TeachingTip) fadeOut() {
	common.FadeOutThen(w.bubble.QWidget, 167, qt.NewQEasingCurve3(qt.QEasingCurve__InSine),
		func() { w.Close() })
}

func (w *TeachingTip) installEvents() {
	w.OnShowEvent(func(super func(event *qt.QShowEvent), event *qt.QShowEvent) {
		super(event)
		if w.duration >= 0 {
			w.fadeOutTimer.Start(w.duration)
		}
		w.adjustPosition()
		w.AdjustSize()
		// Place it again once the window is mapped: a position set before the map can be
		// overridden by the window manager (see common.ApplyAfterMap), which would leave
		// the tip next to - but not on - its target.
		common.ApplyAfterMap(w.QWidget, w.adjustPosition)
		w.fadeIn()
	})
	w.OnCloseEvent(func(super func(event *qt.QCloseEvent), event *qt.QCloseEvent) {
		if w.isDeleteOnClose {
			w.DeleteLater()
		}
		super(event)
		if w.onClosed != nil {
			w.onClosed()
		}
	})
}

// TeachingTipMake creates and shows a teaching tip from a view.
func TeachingTipMake(view *FlyoutViewBase, target *qt.QWidget, duration int, tailPosition TeachingTipTailPosition, parent *qt.QWidget, isDeleteOnClose bool) *TeachingTip {
	w := NewTeachingTip(view, target, duration, tailPosition, parent, isDeleteOnClose)
	w.Show()
	return w
}

// TeachingTipCreate creates and shows a teaching tip using the default view.
func TeachingTipCreate(target *qt.QWidget, title, content string, icon, image interface{}, isClosable bool, duration int, tailPosition TeachingTipTailPosition, parent *qt.QWidget, isDeleteOnClose bool) *TeachingTip {
	view := NewTeachingTipView(title, content, icon, image, isClosable, tailPosition, nil)
	w := TeachingTipMake(view.FlyoutViewBase, target, duration, tailPosition, parent, isDeleteOnClose)
	view.SetOnClosed(func() { w.Close() })
	return w
}

// PopupTeachingTip is a teaching tip shown as a popup window.
type PopupTeachingTip struct {
	*TeachingTip
}

// NewPopupTeachingTip builds a popup teaching tip.
func NewPopupTeachingTip(view *FlyoutViewBase, target *qt.QWidget, duration int, tailPosition TeachingTipTailPosition, parent *qt.QWidget, isDeleteOnClose bool) *PopupTeachingTip {
	w := &PopupTeachingTip{TeachingTip: NewTeachingTip(view, target, duration, tailPosition, parent, isDeleteOnClose)}
	w.SetWindowFlags(qt.Popup | qt.FramelessWindowHint | qt.NoDropShadowWindowHint)
	return w
}
