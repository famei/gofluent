package widgets

import (
	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// strokePathWithColor strokes path using a pen built from color.
func strokePathWithColor(painter *qt.QPainter, path *qt.QPainterPath, color *qt.QColor) {
	pen := qt.NewQPen3(color)
	painter.StrokePath(path, pen)
	pen.Delete()
}

// CardWidget is a rounded card with a theme/hover/pressed background. The
// Python BackgroundAnimationWidget color animation is simplified to direct
// color application (the QPropertyAnimation on a QColor property is not
// expressible in Go).
type CardWidget struct {
	*qt.QFrame
	isPressed        bool
	isHover          bool
	isClickEnabled   bool
	borderRadius     int
	onClicked        func()
	normalColorFunc  func() *qt.QColor
	hoverColorFunc   func() *qt.QColor
	pressedColorFunc func() *qt.QColor
}

// NewCardWidget builds a card widget.
func NewCardWidget(parent *qt.QWidget) *CardWidget {
	w := &CardWidget{QFrame: qt.NewQFrame(parent)}
	w.isClickEnabled = false
	w.borderRadius = 5
	w.normalColorFunc = w.normalBackgroundColor
	w.hoverColorFunc = w.hoverBackgroundColor
	w.pressedColorFunc = w.pressedBackgroundColor
	w.installEvents()
	return w
}

// OnClicked registers the click callback.
func (w *CardWidget) OnClicked(f func()) { w.onClicked = f }

// SetClickEnabled stores whether click feedback is enabled.
func (w *CardWidget) SetClickEnabled(isEnabled bool) {
	w.isClickEnabled = isEnabled
	w.Update()
}

// IsClickEnabled reports whether click feedback is enabled.
func (w *CardWidget) IsClickEnabled() bool { return w.isClickEnabled }

// BorderRadius returns the corner radius.
func (w *CardWidget) BorderRadius() int { return w.borderRadius }

// SetBorderRadius sets the corner radius.
func (w *CardWidget) SetBorderRadius(radius int) {
	w.borderRadius = radius
	w.Update()
}

func (w *CardWidget) normalBackgroundColor() *qt.QColor {
	if common.IsDarkTheme() {
		return qt.NewQColor11(255, 255, 255, 13)
	}
	return qt.NewQColor11(255, 255, 255, 170)
}

func (w *CardWidget) hoverBackgroundColor() *qt.QColor {
	if common.IsDarkTheme() {
		return qt.NewQColor11(255, 255, 255, 21)
	}
	return qt.NewQColor11(255, 255, 255, 64)
}

func (w *CardWidget) pressedBackgroundColor() *qt.QColor {
	if common.IsDarkTheme() {
		return qt.NewQColor11(255, 255, 255, 8)
	}
	return qt.NewQColor11(255, 255, 255, 64)
}

func (w *CardWidget) backgroundColor() *qt.QColor {
	if w.isPressed {
		return w.pressedColorFunc()
	}
	if w.isHover {
		return w.hoverColorFunc()
	}
	return w.normalColorFunc()
}

func (w *CardWidget) installEvents() {
	w.OnMousePressEvent(func(super func(e *qt.QMouseEvent), e *qt.QMouseEvent) {
		w.isPressed = true
		w.Update()
		super(e)
	})
	w.OnMouseReleaseEvent(func(super func(e *qt.QMouseEvent), e *qt.QMouseEvent) {
		w.isPressed = false
		w.Update()
		super(e)
		if w.onClicked != nil {
			w.onClicked()
		}
	})
	w.OnEnterEvent(func(super func(e *qt.QEvent), e *qt.QEvent) {
		w.isHover = true
		w.Update()
	})
	w.OnLeaveEvent(func(super func(e *qt.QEvent), e *qt.QEvent) {
		w.isHover = false
		w.Update()
	})
	w.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		w.paint()
	})
}

func (w *CardWidget) paint() {
	painter := qt.NewQPainter2(w.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__Antialiasing)

	fw, fh := float64(w.Width()), float64(w.Height())
	fr := float64(w.borderRadius)
	fd := 2 * fr
	isDark := common.IsDarkTheme()

	// top border
	top := qt.NewQPainterPath()
	defer top.Delete()
	top.ArcMoveTo2(1, fh-fd-1, fd, fd, 240)
	top.ArcTo2(1, fh-fd-1, fd, fd, 225, -60)
	top.LineTo2(1, fr)
	top.ArcTo2(1, 1, fd, fd, -180, -90)
	top.LineTo2(fw-fr, 1)
	top.ArcTo2(fw-fd-1, 1, fd, fd, 90, -90)
	top.LineTo2(fw-1, fh-fr)
	top.ArcTo2(fw-fd-1, fh-fd-1, fd, fd, 0, -60)

	var topBorder *qt.QColor
	if isDark {
		if w.isPressed {
			topBorder = qt.NewQColor11(255, 255, 255, 18)
		} else if w.isHover {
			topBorder = qt.NewQColor11(255, 255, 255, 13)
		} else {
			topBorder = qt.NewQColor11(0, 0, 0, 20)
		}
	} else {
		topBorder = qt.NewQColor11(0, 0, 0, 15)
	}
	defer topBorder.Delete()
	strokePathWithColor(painter, top, topBorder)

	// bottom border
	bottom := qt.NewQPainterPath()
	defer bottom.Delete()
	bottom.ArcMoveTo2(1, fh-fd-1, fd, fd, 240)
	bottom.ArcTo2(1, fh-fd-1, fd, fd, 240, 30)
	bottom.LineTo2(fw-fr-1, fh-1)
	bottom.ArcTo2(fw-fd-1, fh-fd-1, fd, fd, 270, 30)

	var bottomBorder *qt.QColor
	if !isDark && w.isHover && !w.isPressed {
		bottomBorder = qt.NewQColor11(0, 0, 0, 27)
	} else {
		// Do not alias topBorder: it is owned and freed by its own deferred
		// Delete(), so sharing the pointer would risk a double-free when the
		// hover branch later replaces it. Copy it instead.
		bottomBorder = cloneColor(topBorder)
	}
	defer bottomBorder.Delete()
	strokePathWithColor(painter, bottom, bottomBorder)

	// background
	painter.SetPenWithStyle(qt.NoPen)
	rect := w.Rect().Adjusted(1, 1, -1, -1) // GoGC-armed — do NOT Delete
	bg := w.backgroundColor()
	defer bg.Delete()
	brush := qt.NewQBrush3(bg)
	painter.SetBrush(brush)
	brush.Delete()
	painter.DrawRoundedRect3(rect, fr, fr)
	painter.End()
}

// SimpleCardWidget is a card with a flat background (hover/pressed equal the
// normal background).
type SimpleCardWidget struct{ *CardWidget }

// NewSimpleCardWidget builds a simple card widget.
func NewSimpleCardWidget(parent *qt.QWidget) *SimpleCardWidget {
	w := &SimpleCardWidget{CardWidget: NewCardWidget(parent)}
	w.hoverColorFunc = w.normalBackgroundColor
	w.pressedColorFunc = w.normalBackgroundColor
	w.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		w.paintSimple()
	})
	return w
}

func (w *SimpleCardWidget) paintSimple() {
	painter := qt.NewQPainter2(w.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__Antialiasing)

	bg := w.backgroundColor()
	defer bg.Delete()
	brush := qt.NewQBrush3(bg)
	painter.SetBrush(brush)
	brush.Delete()

	if common.IsDarkTheme() {
		painter.SetPen(qt.NewQColor11(0, 0, 0, 48))
	} else {
		painter.SetPen(qt.NewQColor11(0, 0, 0, 12))
	}
	r := float64(w.borderRadius)
	rect := w.Rect().Adjusted(1, 1, -1, -1) // GoGC-armed — do NOT Delete
	painter.DrawRoundedRect3(rect, r, r)
	painter.End()
}

// ElevatedCardWidget is a simple card with a drop shadow on hover. The Python
// position-lift animation is not ported.
type ElevatedCardWidget struct {
	*SimpleCardWidget
	shadowAni *common.DropShadowAnimation
}

// NewElevatedCardWidget builds an elevated card widget.
func NewElevatedCardWidget(parent *qt.QWidget) *ElevatedCardWidget {
	w := &ElevatedCardWidget{SimpleCardWidget: NewSimpleCardWidget(parent)}
	w.shadowAni = common.NewDropShadowAnimation(w.QWidget, qt.NewQColor11(0, 0, 0, 0), qt.NewQColor11(0, 0, 0, 20))
	w.shadowAni.SetOffset(0, 5)
	w.shadowAni.SetBlurRadius(38)
	w.shadowAni.SetHover(false)
	w.SetBorderRadius(8)
	w.hoverColorFunc = w.hoverElevatedColor
	w.pressedColorFunc = w.pressedElevatedColor

	// The drop-shadow animation registers its own enter/leave handlers; re-register
	// them so hover tracking and the shadow are both updated.
	w.OnEnterEvent(func(super func(e *qt.QEvent), e *qt.QEvent) {
		w.isHover = true
		w.Update()
		w.shadowAni.SetHover(true)
	})
	w.OnLeaveEvent(func(super func(e *qt.QEvent), e *qt.QEvent) {
		w.isHover = false
		w.Update()
		w.shadowAni.SetHover(false)
	})
	return w
}

func (w *ElevatedCardWidget) hoverElevatedColor() *qt.QColor {
	if common.IsDarkTheme() {
		return qt.NewQColor11(255, 255, 255, 16)
	}
	return qt.NewQColor3(255, 255, 255)
}

func (w *ElevatedCardWidget) pressedElevatedColor() *qt.QColor {
	if common.IsDarkTheme() {
		return qt.NewQColor11(255, 255, 255, 6)
	}
	return qt.NewQColor11(255, 255, 255, 118)
}

// CardSeparator is a thin line used inside cards.
type CardSeparator struct{ *qt.QWidget }

// NewCardSeparator builds a card separator.
func NewCardSeparator(parent *qt.QWidget) *CardSeparator {
	w := &CardSeparator{QWidget: qt.NewQWidget(parent)}
	w.SetFixedHeight(3)
	w.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing)
		if common.IsDarkTheme() {
			painter.SetPen(qt.NewQColor11(255, 255, 255, 46))
		} else {
			painter.SetPen(qt.NewQColor11(0, 0, 0, 12))
		}
		painter.DrawLine2(2, 1, w.Width()-2, 1)
		painter.End()
	})
	return w
}

// HeaderCardWidget is a card with a header label, separator and content view.
//
// Constructors
//   - NewHeaderCardWidget(parent *qt.QWidget)
//   - NewHeaderCardWidgetTitle(title string, parent *qt.QWidget)
type HeaderCardWidget struct {
	*SimpleCardWidget
	headerView   *qt.QWidget
	headerLabel  *qt.QLabel
	separator    *CardSeparator
	view         *qt.QWidget
	vBoxLayout   *qt.QVBoxLayout
	headerLayout *qt.QHBoxLayout
	viewLayout   *qt.QHBoxLayout
}

// NewHeaderCardWidget builds a header card widget.
func NewHeaderCardWidget(parent *qt.QWidget) *HeaderCardWidget {
	w := &HeaderCardWidget{SimpleCardWidget: NewSimpleCardWidget(parent)}
	w.headerView = qt.NewQWidget(w.QWidget)
	w.headerLabel = qt.NewQLabel(w.QWidget)
	w.separator = NewCardSeparator(w.QWidget)
	w.view = qt.NewQWidget(w.QWidget)

	w.vBoxLayout = qt.NewQVBoxLayout(w.QWidget)
	w.headerLayout = qt.NewQHBoxLayout(w.headerView)
	w.viewLayout = qt.NewQHBoxLayout(w.view)

	w.headerLayout.AddWidget(w.headerLabel.QWidget)
	w.headerLayout.SetContentsMargins(24, 0, 16, 0)
	w.headerView.SetFixedHeight(48)

	w.vBoxLayout.SetSpacing(0)
	w.vBoxLayout.SetContentsMargins(0, 0, 0, 0)
	w.vBoxLayout.AddWidget(w.headerView)
	w.vBoxLayout.AddWidget(w.separator.QWidget)
	w.vBoxLayout.AddWidget(w.view)

	w.viewLayout.SetContentsMargins(24, 24, 24, 24)
	common.SetFont(w.headerLabel.QWidget, 15, int(qt.QFont__DemiBold))

	w.view.SetObjectName("view")
	w.headerView.SetObjectName("headerView")
	w.headerLabel.SetObjectName("headerLabel")
	common.FluentStyleSheet(common.FluentCardWidget).Apply(w.QWidget, common.ThemeAuto)
	return w
}

// NewHeaderCardWidgetTitle builds a header card widget with a title.
func NewHeaderCardWidgetTitle(title string, parent *qt.QWidget) *HeaderCardWidget {
	w := NewHeaderCardWidget(parent)
	w.SetTitle(title)
	return w
}

// Title returns the header title.
func (w *HeaderCardWidget) Title() string { return w.headerLabel.Text() }

// SetTitle sets the header title.
func (w *HeaderCardWidget) SetTitle(title string) { w.headerLabel.SetText(title) }

// CardGroupWidget is a single row of a GroupHeaderCardWidget.
type CardGroupWidget struct {
	*qt.QWidget
	vBoxLayout   *qt.QVBoxLayout
	hBoxLayout   *qt.QHBoxLayout
	iconWidget   *IconWidget
	titleLabel   *BodyLabel
	contentLabel *CaptionLabel
	textLayout   *qt.QVBoxLayout
	separator    *CardSeparator
}

// NewCardGroupWidget builds a card group row.
func NewCardGroupWidget(icon interface{}, title, content string, parent *qt.QWidget) *CardGroupWidget {
	w := &CardGroupWidget{QWidget: qt.NewQWidget(parent)}
	w.vBoxLayout = qt.NewQVBoxLayout(w.QWidget)
	w.hBoxLayout = qt.NewQHBoxLayout2()
	w.iconWidget = NewIconWidgetIcon(icon, nil)
	w.titleLabel = NewBodyLabelText(title, nil)
	w.contentLabel = NewCaptionLabelText(content, nil)
	w.textLayout = qt.NewQVBoxLayout2()
	w.separator = NewCardSeparator(nil)
	w.initWidget()
	return w
}

func (w *CardGroupWidget) initWidget() {
	w.separator.Hide()
	w.iconWidget.SetFixedSize2(20, 20)
	light := qt.NewQColor3(96, 96, 96)
	dark := qt.NewQColor3(206, 206, 206)
	w.contentLabel.SetTextColor(light, dark)
	light.Delete()
	dark.Delete()

	w.vBoxLayout.SetSpacing(0)
	w.vBoxLayout.SetContentsMargins(0, 0, 0, 0)
	w.vBoxLayout.AddLayout(w.hBoxLayout.QLayout)
	w.vBoxLayout.AddWidget(w.separator.QWidget)

	w.textLayout.AddWidget(w.titleLabel.QWidget)
	w.textLayout.AddWidget(w.contentLabel.QWidget)
	w.hBoxLayout.AddWidget(w.iconWidget.QWidget)
	w.hBoxLayout.AddLayout(w.textLayout.QLayout)
	w.hBoxLayout.AddStretchWithStretch(1)

	w.hBoxLayout.SetSpacing(15)
	w.hBoxLayout.SetContentsMargins(24, 10, 24, 10)
	w.textLayout.SetContentsMargins(0, 0, 0, 0)
	w.textLayout.SetSpacing(0)
}

// Title returns the group title.
func (w *CardGroupWidget) Title() string { return w.titleLabel.Text() }

// SetTitle sets the group title.
func (w *CardGroupWidget) SetTitle(text string) { w.titleLabel.SetText(text) }

// Content returns the group content.
func (w *CardGroupWidget) Content() string { return w.contentLabel.Text() }

// SetContent sets the group content.
func (w *CardGroupWidget) SetContent(text string) { w.contentLabel.SetText(text) }

// Icon returns the group icon.
func (w *CardGroupWidget) Icon() *qt.QIcon { return w.iconWidget.GetIcon() }

// SetIcon sets the group icon.
func (w *CardGroupWidget) SetIcon(icon interface{}) { w.iconWidget.SetIcon(icon) }

// SetIconSize sets the group icon size.
func (w *CardGroupWidget) SetIconSize(size *qt.QSize) { w.iconWidget.SetFixedSize(size) }

// SetSeparatorVisible shows or hides the separator.
func (w *CardGroupWidget) SetSeparatorVisible(isVisible bool) { w.separator.SetVisible(isVisible) }

// IsSeparatorVisible reports whether the separator is visible.
func (w *CardGroupWidget) IsSeparatorVisible() bool { return w.separator.IsVisible() }

// AddWidget adds a widget to the group row.
func (w *CardGroupWidget) AddWidget(widget *qt.QWidget, stretch int) {
	w.hBoxLayout.AddWidget2(widget, stretch)
}

// GroupHeaderCardWidget is a header card that holds multiple card groups.
type GroupHeaderCardWidget struct {
	*HeaderCardWidget
	groupWidgets []*CardGroupWidget
	groupLayout  *qt.QVBoxLayout
}

// NewGroupHeaderCardWidget builds a group header card widget.
func NewGroupHeaderCardWidget(parent *qt.QWidget) *GroupHeaderCardWidget {
	w := &GroupHeaderCardWidget{HeaderCardWidget: NewHeaderCardWidget(parent)}
	w.groupWidgets = []*CardGroupWidget{}
	w.groupLayout = qt.NewQVBoxLayout2()
	w.groupLayout.SetSpacing(0)
	w.viewLayout.SetContentsMargins(0, 0, 0, 0)
	w.groupLayout.SetContentsMargins(0, 0, 0, 0)
	w.viewLayout.AddLayout(w.groupLayout.QLayout)
	return w
}

// AddGroup adds a new group row and returns it.
func (w *GroupHeaderCardWidget) AddGroup(icon interface{}, title, content string, widget *qt.QWidget, stretch int) *CardGroupWidget {
	group := NewCardGroupWidget(icon, title, content, w.QWidget)
	group.AddWidget(widget, stretch)
	if len(w.groupWidgets) > 0 {
		w.groupWidgets[len(w.groupWidgets)-1].SetSeparatorVisible(true)
	}
	w.groupLayout.AddWidget(group.QWidget)
	w.groupWidgets = append(w.groupWidgets, group)
	return group
}

// GroupCount returns the number of groups.
func (w *GroupHeaderCardWidget) GroupCount() int { return len(w.groupWidgets) }
