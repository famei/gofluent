package settings

import (
	"math"

	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// ExpandButton is the arrow button that toggles an ExpandSettingCard. The
// rotation animation on the custom "angle" pyqtProperty is driven by a
// frame-based ProgressAnimation (the custom meta-object property is not
// expressible in Go).
type ExpandButton struct {
	*qt.QAbstractButton
	angle         float64
	isHover       bool
	isPressed     bool
	onClickedHook func()
	rotateAni     *common.ProgressAnimation
}

// NewExpandButton builds an expand arrow button.
func NewExpandButton(parent *qt.QWidget) *ExpandButton {
	b := &ExpandButton{QAbstractButton: qt.NewQAbstractButton(parent)}
	b.SetFixedSize2(30, 30)
	b.OnClicked(func() {
		if b.angle < 180 {
			b.SetExpand(true)
		} else {
			b.SetExpand(false)
		}
		if b.onClickedHook != nil {
			b.onClickedHook()
		}
	})
	b.OnEnterEvent(func(super func(e *qt.QEvent), e *qt.QEvent) {
		super(e)
		b.SetHover(true)
	})
	b.OnLeaveEvent(func(super func(e *qt.QEvent), e *qt.QEvent) {
		super(e)
		b.SetHover(false)
	})
	b.OnMousePressEvent(func(super func(e *qt.QMouseEvent), e *qt.QMouseEvent) {
		super(e)
		b.SetPressed(true)
	})
	b.OnMouseReleaseEvent(func(super func(e *qt.QMouseEvent), e *qt.QMouseEvent) {
		super(e)
		b.SetPressed(false)
	})
	b.OnPaintEvent(func(e *qt.QPaintEvent) {
		b.paint()
	})
	return b
}

// SetClickedHook registers an extra listener invoked after the arrow rotates.
func (b *ExpandButton) SetClickedHook(f func()) { b.onClickedHook = f }

// SetHover updates the hover state and repaints.
func (b *ExpandButton) SetHover(isHover bool) {
	b.isHover = isHover
	b.Update()
}

// SetPressed updates the pressed state and repaints.
func (b *ExpandButton) SetPressed(isPressed bool) {
	b.isPressed = isPressed
	b.Update()
}

// SetExpand sets the arrow angle (180 expanded, 0 collapsed), animating over
// 200ms with the default linear easing (mirrors Python's rotateAni).
func (b *ExpandButton) SetExpand(isExpand bool) {
	target := 0.0
	if isExpand {
		target = 180
	}
	b.stopRotateAni()
	from := b.angle
	b.rotateAni = common.AnimateFloat(from, target, 200, nil, func(v float64) {
		b.angle = v
		b.Update()
	})
}

// stopRotateAni interrupts and releases any in-flight rotation animation.
func (b *ExpandButton) stopRotateAni() {
	if b.rotateAni != nil {
		b.rotateAni.Stop()
		b.rotateAni.Delete()
		b.rotateAni = nil
	}
}

// Angle returns the current arrow angle.
func (b *ExpandButton) Angle() float64 { return b.angle }

func (b *ExpandButton) paint() {
	painter := qt.NewQPainter2(b.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__Antialiasing | qt.QPainter__SmoothPixmapTransform)
	painter.SetPenWithStyle(qt.NoPen)

	r := 255
	if !common.IsDarkTheme() {
		r = 0
	}
	color := qt.NewQColor11(0, 0, 0, 0)
	if b.IsEnabled() {
		if b.isPressed {
			color = qt.NewQColor11(r, r, r, 10)
		} else if b.isHover {
			color = qt.NewQColor11(r, r, r, 14)
		}
	} else {
		painter.SetOpacity(0.36)
	}
	brush := qt.NewQBrush3(color)
	painter.SetBrush(brush)
	brush.Delete()
	painter.DrawRoundedRect3(b.Rect(), 4, 4)

	painter.Translate2(float64(b.Width())/2, float64(b.Height())/2)
	painter.Rotate(b.angle)
	rect := qt.NewQRectF4(-5, -5, 9.6, 9.6)
	defer rect.Delete()
	common.ArrowDown.Render(painter, rect, common.ThemeAuto)
	painter.End()
}

// SpaceWidget is a 1px transparent spacer used at the bottom of an expand card.
type SpaceWidget struct {
	*qt.QWidget
}

// NewSpaceWidget builds a space widget.
func NewSpaceWidget(parent *qt.QWidget) *SpaceWidget {
	w := &SpaceWidget{QWidget: qt.NewQWidget(parent)}
	w.SetAttribute(qt.WA_TranslucentBackground)
	w.SetFixedHeight(1)
	return w
}

// HeaderSettingCard is a setting card with an expand button on the right. It
// acts as the header of an ExpandSettingCard.
type HeaderSettingCard struct {
	*SettingCard
	expandButton *ExpandButton
	expandState  func() bool
}

// NewHeaderSettingCard builds a header setting card.
func NewHeaderSettingCard(icon interface{}, title, content string, parent *qt.QWidget) *HeaderSettingCard {
	card := &HeaderSettingCard{SettingCard: NewSettingCard(icon, title, content, parent)}
	card.expandButton = NewExpandButton(card.QWidget)

	card.hBoxLayout.AddWidget3(card.expandButton.QWidget, 0, qt.AlignRight)
	card.hBoxLayout.AddSpacing(8)

	card.titleLabel.SetObjectName("titleLabel")
	card.InstallEventFilter(card.QObject)
	card.OnEventFilter(func(super func(watched *qt.QObject, event *qt.QEvent) bool, watched *qt.QObject, event *qt.QEvent) bool {
		if watched.UnsafePointer() == card.UnsafePointer() {
			// The Python port additionally checks e.button() == Qt.LeftButton;
			// miqt's eventFilter delivers a *qt.QEvent, so the button check is
			// omitted (right/middle clicks also toggle).
			switch event.Type() {
			case qt.QEvent__Enter:
				card.expandButton.SetHover(true)
			case qt.QEvent__Leave:
				card.expandButton.SetHover(false)
			case qt.QEvent__MouseButtonPress:
				card.expandButton.SetPressed(true)
			case qt.QEvent__MouseButtonRelease:
				card.expandButton.SetPressed(false)
				card.expandButton.Click()
			}
		}
		return super(watched, event)
	})
	card.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		super(e)
		card.paint()
	})
	return card
}

// ExpandButton exposes the header expand button.
func (c *HeaderSettingCard) ExpandButton() *ExpandButton { return c.expandButton }

// AddWidget inserts a widget before the expand button in the header.
func (c *HeaderSettingCard) AddWidget(widget *qt.QWidget) {
	n := c.hBoxLayout.Count()
	if n > 0 {
		if item := c.hBoxLayout.ItemAt(n - 1); item != nil {
			c.hBoxLayout.RemoveItem(item)
		}
	}
	c.hBoxLayout.AddWidget3(widget, 0, qt.AlignRight)
	c.hBoxLayout.AddSpacing(19)
	c.hBoxLayout.AddWidget3(c.expandButton.QWidget, 0, qt.AlignRight)
	c.hBoxLayout.AddSpacing(8)
}

func (c *HeaderSettingCard) paint() {
	painter := qt.NewQPainter2(c.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__Antialiasing)
	painter.SetPenWithStyle(qt.NoPen)

	var color *qt.QColor
	if common.IsDarkTheme() {
		color = qt.NewQColor11(255, 255, 255, 13)
	} else {
		color = qt.NewQColor11(255, 255, 255, 170)
	}
	brush := qt.NewQBrush3(color)
	painter.SetBrush(brush)
	brush.Delete()

	path := qt.NewQPainterPath()
	defer path.Delete()
	path.SetFillRule(qt.WindingFill)
	rect := qt.NewQRectF5(c.Rect())
	defer rect.Delete()
	path.AddRoundedRect(rect, 10, 10)

	if c.expandState != nil && c.expandState() {
		path.AddRect2(0, float64(c.Height()-8), float64(c.Width()), 8)
	}
	painter.DrawPath(path.Simplified())
	painter.End()
}

// ExpandBorderWidget draws the border that outlines an expand card.
type ExpandBorderWidget struct {
	*qt.QWidget
	card *HeaderSettingCard
}

// NewExpandBorderWidget builds a border widget tracking a header card.
func NewExpandBorderWidget(parent *qt.QWidget, card *HeaderSettingCard) *ExpandBorderWidget {
	w := &ExpandBorderWidget{QWidget: qt.NewQWidget(parent), card: card}
	w.SetAttribute(qt.WA_TransparentForMouseEvents)
	parent.InstallEventFilter(w.QObject)
	w.OnEventFilter(func(super func(watched *qt.QObject, event *qt.QEvent) bool, watched *qt.QObject, event *qt.QEvent) bool {
		if watched.UnsafePointer() == parent.UnsafePointer() && event.Type() == qt.QEvent__Resize {
			w.Resize(parent.Width(), parent.Height())
		}
		return super(watched, event)
	})
	w.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		super(e)
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing)
		painter.SetBrushWithStyle(qt.NoBrush)

		var pen *qt.QColor
		if common.IsDarkTheme() {
			pen = qt.NewQColor11(0, 0, 0, 50)
		} else {
			pen = qt.NewQColor11(0, 0, 0, 19)
		}
		painter.SetPen(pen)
		adjusted := w.Rect().Adjusted(1, 1, -1, -1) // GoGC-armed — do NOT Delete
		painter.DrawRoundedRect3(adjusted, 6, 6)

		ch := w.card.Height()
		if ch < w.Height() {
			painter.DrawLine2(1, ch, w.Width()-1, ch)
		}
		painter.End()
	})
	return w
}

// ExpandSettingCard is an expandable setting card backed by a QScrollArea. The
// expand/collapse height change is animated with a frame-based
// ProgressAnimation (OutQuad, 200ms) mirroring Python's scroll-value animation.
type ExpandSettingCard struct {
	*qt.QScrollArea
	isExpand     bool
	card         *HeaderSettingCard
	scrollWidget *qt.QFrame
	view         *qt.QFrame
	scrollLayout *qt.QVBoxLayout
	viewLayout   *qt.QVBoxLayout
	spaceWidget  *SpaceWidget
	borderWidget *ExpandBorderWidget
	expandAni    *common.ProgressAnimation

	adjustViewSizeFn func()
}

// NewExpandSettingCard builds an expandable setting card.
func NewExpandSettingCard(icon interface{}, title, content string, parent *qt.QWidget) *ExpandSettingCard {
	e := &ExpandSettingCard{QScrollArea: qt.NewQScrollArea(parent)}
	e.scrollWidget = qt.NewQFrame(e.QWidget)
	e.view = qt.NewQFrame(e.scrollWidget.QWidget)
	e.card = NewHeaderSettingCard(icon, title, content, e.QWidget)
	e.scrollLayout = qt.NewQVBoxLayout(e.scrollWidget.QWidget)
	e.viewLayout = qt.NewQVBoxLayout(e.view.QWidget)
	e.spaceWidget = NewSpaceWidget(e.scrollWidget.QWidget)
	e.borderWidget = NewExpandBorderWidget(e.QWidget, e.card)
	e.initWidget()
	return e
}

func (e *ExpandSettingCard) initWidget() {
	e.SetWidget(e.scrollWidget.QWidget)
	e.SetWidgetResizable(true)
	e.SetFixedHeight(e.card.Height())
	e.SetViewportMargins(0, e.card.Height(), 0, 0)
	e.SetVerticalScrollBarPolicy(qt.ScrollBarAlwaysOff)
	e.SetHorizontalScrollBarPolicy(qt.ScrollBarAlwaysOff)

	e.scrollLayout.SetContentsMargins(0, 0, 0, 0)
	e.scrollLayout.SetSpacing(0)
	e.scrollLayout.AddWidget(e.view.QWidget)
	e.scrollLayout.AddWidget(e.spaceWidget.QWidget)

	e.view.SetObjectName("view")
	e.scrollWidget.SetObjectName("scrollWidget")
	e.SetProperty("isExpand", qt.NewQVariant11(false))

	common.FluentStyleSheet(common.FluentExpandSettingCard).Apply(e.card.QFrame.QWidget, common.ThemeAuto)
	common.FluentStyleSheet(common.FluentExpandSettingCard).Apply(e.QWidget, common.ThemeAuto)

	e.card.expandButton.SetClickedHook(e.ToggleExpand)
	e.card.expandState = func() bool { return e.isExpand }

	// Mirror Python ExpandSettingCard.wheelEvent: pass (leave the event ignored)
	// so Qt propagates it to the parent scroll area. Accepting it here swallowed
	// the wheel and froze page scrolling whenever the pointer was over the card.
	e.OnWheelEvent(func(super func(event *qt.QWheelEvent), event *qt.QWheelEvent) {
		event.SetAccepted(false)
	})
	e.OnResizeEvent(func(super func(event *qt.QResizeEvent), event *qt.QResizeEvent) {
		super(event)
		e.card.Resize(e.Width(), e.card.Height())
		e.scrollWidget.Resize(e.Width(), e.scrollWidget.Height())
	})
}

// Card exposes the header card.
func (e *ExpandSettingCard) Card() *HeaderSettingCard { return e.card }

// View exposes the scrollable view frame.
func (e *ExpandSettingCard) View() *qt.QFrame { return e.view }

// ViewLayout exposes the view layout.
func (e *ExpandSettingCard) ViewLayout() *qt.QVBoxLayout { return e.viewLayout }

// SpaceWidget exposes the bottom spacer.
func (e *ExpandSettingCard) SpaceWidget() *SpaceWidget { return e.spaceWidget }

// IsExpand reports the expand state.
func (e *ExpandSettingCard) IsExpand() bool { return e.isExpand }

// AddWidget adds a trailing widget to the header card.
func (e *ExpandSettingCard) AddWidget(widget *qt.QWidget) {
	e.card.AddWidget(widget)
	e.adjustViewSize()
}

// SetExpand sets the expand state.
func (e *ExpandSettingCard) SetExpand(isExpand bool) {
	if e.isExpand == isExpand {
		return
	}

	e.adjustViewSize()
	e.isExpand = isExpand
	e.SetProperty("isExpand", qt.NewQVariant11(isExpand))
	e.SetStyle(qt.QApplication_Style())
	e.card.expandButton.SetExpand(isExpand)

	// Animate the vertical scrollbar value and derive the fixed height from it,
	// exactly like Python (expand_setting_card.py setExpand + _onExpandValueChanged):
	// the content is revealed/hidden by scrolling, and the height is
	// cardH + vh - scrollValue clamped to cardH.
	e.stopExpandAni()
	sb := e.VerticalScrollBar()

	var from, to float64
	if isExpand {
		vh := e.viewLayout.SizeHint().Height()
		sb.SetValue(vh)
		from = float64(vh)
		to = 0
	} else {
		from = 0
		to = float64(sb.Maximum())
	}

	if from == to {
		// No scroll range to animate (content fits the viewport); snap to the
		// final height instead of starting a degenerate animation.
		e.syncExpandedHeight()
		return
	}

	curve := qt.NewQEasingCurve3(qt.QEasingCurve__OutQuad)
	e.expandAni = common.AnimateFloat(from, to, 200, curve, func(v float64) {
		iv := int(math.Round(v))
		sb.SetValue(iv)
		e.syncExpandedHeight()
		// Repaint the parent (the card/group hosting this expand card) so the
		// sibling widgets repaint cleanly while the card shrinks on collapse.
		if p := e.ParentWidget(); p != nil {
			p.Repaint()
		}
	})
	curve.Delete()
}

// syncExpandedHeight derives the card height from the current scrollbar value,
// mirroring Python's _onExpandValueChanged: fixedHeight = cardH + viewH -
// scrollValue, clamped to cardH.
func (e *ExpandSettingCard) syncExpandedHeight() {
	vh := e.viewLayout.SizeHint().Height()
	cardH := e.card.Height()
	h := cardH + vh - e.VerticalScrollBar().Value()
	if h < cardH {
		h = cardH
	}
	e.SetFixedHeight(h)
}

// stopExpandAni interrupts and releases any in-flight expand animation.
func (e *ExpandSettingCard) stopExpandAni() {
	if e.expandAni != nil {
		e.expandAni.Stop()
		e.expandAni.Delete()
		e.expandAni = nil
	}
}

// ToggleExpand flips the expand state.
func (e *ExpandSettingCard) ToggleExpand() {
	e.SetExpand(!e.isExpand)
}

// SetValue stores the config value (overridden by subclasses).
func (e *ExpandSettingCard) SetValue(value interface{}) {}

func (e *ExpandSettingCard) adjustViewSize() {
	if e.adjustViewSizeFn != nil {
		e.adjustViewSizeFn()
		return
	}
	h := e.viewLayout.SizeHint().Height()
	e.spaceWidget.SetFixedHeight(h)
	if e.isExpand {
		e.SetFixedHeight(e.card.Height() + h)
	}
}

// GroupSeparator draws a hairline separator between group widgets.
type GroupSeparator struct {
	*qt.QWidget
}

// NewGroupSeparator builds a group separator.
func NewGroupSeparator(parent *qt.QWidget) *GroupSeparator {
	s := &GroupSeparator{QWidget: qt.NewQWidget(parent)}
	s.SetFixedHeight(3)
	s.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		super(e)
		painter := qt.NewQPainter2(s.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing)
		var pen *qt.QColor
		if common.IsDarkTheme() {
			pen = qt.NewQColor11(0, 0, 0, 50)
		} else {
			pen = qt.NewQColor11(0, 0, 0, 19)
		}
		painter.SetPen(pen)
		painter.DrawLine2(0, 1, s.Width(), 1)
		painter.End()
	})
	return s
}

// GroupWidget is a single row (icon + title + content + trailing widget).
type GroupWidget struct {
	*qt.QWidget
	iconWidget   *SettingIconWidget
	titleLabel   *qt.QLabel
	contentLabel *qt.QLabel
	widget       *qt.QWidget
	hBoxLayout   *qt.QHBoxLayout
	vBoxLayout   *qt.QVBoxLayout
}

// NewGroupWidget builds a group row.
func NewGroupWidget(icon interface{}, title, content string, widget *qt.QWidget, stretch int, parent *qt.QWidget) *GroupWidget {
	g := &GroupWidget{QWidget: qt.NewQWidget(parent), widget: widget}
	g.iconWidget = NewSettingIconWidget(icon, g.QWidget)
	g.titleLabel = qt.NewQLabel5(title, g.QWidget)
	g.contentLabel = qt.NewQLabel5(content, g.QWidget)
	g.hBoxLayout = qt.NewQHBoxLayout(g.QWidget)
	g.vBoxLayout = qt.NewQVBoxLayout2()

	if content == "" {
		g.contentLabel.Hide()
	}
	g.iconWidget.SetFixedSize2(16, 16)
	g.SetMinimumHeight(60)
	g.SetIcon(icon)

	g.hBoxLayout.SetSpacing(16)
	g.hBoxLayout.SetContentsMargins(48, 12, 48, 12)
	g.vBoxLayout.SetSpacing(0)
	g.vBoxLayout.SetContentsMargins(0, 0, 0, 0)

	g.hBoxLayout.AddWidget3(g.iconWidget.QWidget, 0, qt.AlignLeft)
	g.hBoxLayout.AddLayout(g.vBoxLayout.QLayout)
	g.vBoxLayout.AddWidget3(g.titleLabel.QWidget, 0, qt.AlignLeft)
	g.vBoxLayout.AddWidget3(g.contentLabel.QWidget, 0, qt.AlignLeft)
	g.hBoxLayout.AddStretchWithStretch(1)
	g.hBoxLayout.AddWidget2(widget, stretch)

	g.titleLabel.SetObjectName("titleLabel")
	g.contentLabel.SetObjectName("contentLabel")
	return g
}

// SetTitle sets the row title.
func (g *GroupWidget) SetTitle(title string) { g.titleLabel.SetText(title) }

// SetContent sets the row content.
func (g *GroupWidget) SetContent(content string) {
	g.contentLabel.SetText(content)
	g.contentLabel.SetVisible(content != "")
}

// SetIconSize sets the icon fixed size.
func (g *GroupWidget) SetIconSize(width, height int) {
	g.iconWidget.SetFixedSize2(width, height)
}

// SetIcon sets the icon and hides it when the icon is null.
func (g *GroupWidget) SetIcon(icon interface{}) {
	g.iconWidget.SetIcon(icon)
	g.iconWidget.SetHidden(g.iconWidget.GetIcon().IsNull())
}

// Widget exposes the trailing widget.
func (g *GroupWidget) Widget() *qt.QWidget { return g.widget }

// ExpandGroupSettingCard is an expand card that stacks group widgets.
type ExpandGroupSettingCard struct {
	*ExpandSettingCard
	widgets []*qt.QWidget
}

// NewExpandGroupSettingCard builds an expand group setting card.
func NewExpandGroupSettingCard(icon interface{}, title, content string, parent *qt.QWidget) *ExpandGroupSettingCard {
	c := &ExpandGroupSettingCard{ExpandSettingCard: NewExpandSettingCard(icon, title, content, parent)}
	c.viewLayout.SetContentsMargins(0, 0, 0, 0)
	c.viewLayout.SetSpacing(0)
	c.adjustViewSizeFn = c.adjustGroupViewSize
	return c
}

// AddGroupWidget appends a widget to the group (with a separator between rows).
func (c *ExpandGroupSettingCard) AddGroupWidget(widget *qt.QWidget) {
	if c.viewLayout.Count() >= 1 {
		c.viewLayout.AddWidget(NewGroupSeparator(c.view.QWidget).QWidget)
	}
	widget.SetParent(c.view.QWidget)
	c.widgets = append(c.widgets, widget)
	c.viewLayout.AddWidget(widget)
	c.adjustViewSize()
}

// AddGroup builds a GroupWidget row and appends it.
func (c *ExpandGroupSettingCard) AddGroup(icon interface{}, title, content string, widget *qt.QWidget, stretch int) *GroupWidget {
	group := NewGroupWidget(icon, title, content, widget, stretch, nil)
	c.AddGroupWidget(group.QWidget)
	return group
}

// RemoveGroupWidget removes a widget from the group.
func (c *ExpandGroupSettingCard) RemoveGroupWidget(widget *qt.QWidget) {
	index := widgetIndex(c.widgets, widget)
	if index < 0 {
		return
	}
	layoutIndex := c.viewLayout.IndexOf(widget)
	c.viewLayout.RemoveWidget(widget)
	c.widgets = append(c.widgets[:index], c.widgets[index+1:]...)

	if len(c.widgets) == 0 {
		c.adjustViewSize()
		return
	}

	if layoutIndex >= 1 {
		if item := c.viewLayout.ItemAt(layoutIndex - 1); item != nil {
			if sep := item.Widget(); sep != nil {
				sep.DeleteLater()
				c.viewLayout.RemoveWidget(sep)
			}
		}
	} else if index == 0 {
		if item := c.viewLayout.ItemAt(0); item != nil {
			if sep := item.Widget(); sep != nil {
				sep.DeleteLater()
				c.viewLayout.RemoveWidget(sep)
			}
		}
	}

	c.adjustViewSize()
}

// Widgets returns the group widgets.
func (c *ExpandGroupSettingCard) Widgets() []*qt.QWidget { return c.widgets }

func (c *ExpandGroupSettingCard) adjustGroupViewSize() {
	h := 0
	for _, w := range c.widgets {
		size := w.SizeHint()
		h += size.Height() + 3

	}
	c.spaceWidget.SetFixedHeight(h)
	if c.isExpand {
		c.SetFixedHeight(c.card.Height() + h)
	}
}

// SimpleExpandGroupSettingCard sizes the view from the layout size hint instead
// of summing individual widget heights.
type SimpleExpandGroupSettingCard struct {
	*ExpandGroupSettingCard
}

// NewSimpleExpandGroupSettingCard builds a simple expand group setting card.
func NewSimpleExpandGroupSettingCard(icon interface{}, title, content string, parent *qt.QWidget) *SimpleExpandGroupSettingCard {
	c := &SimpleExpandGroupSettingCard{ExpandGroupSettingCard: NewExpandGroupSettingCard(icon, title, content, parent)}
	c.adjustViewSizeFn = c.adjustSimpleViewSize
	return c
}

func (c *SimpleExpandGroupSettingCard) adjustSimpleViewSize() {
	h := c.viewLayout.SizeHint().Height()
	c.spaceWidget.SetFixedHeight(h)
	if c.isExpand {
		c.SetFixedHeight(c.card.Height() + h)
	}
}

func widgetIndex(list []*qt.QWidget, w *qt.QWidget) int {
	for i, x := range list {
		if x.UnsafePointer() == w.UnsafePointer() {
			return i
		}
	}
	return -1
}
