package widgets

import (
	"fmt"
	"math"

	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// TabCloseButtonDisplayMode enumerates when the tab close button is shown.
type TabCloseButtonDisplayMode int

const (
	// TabCloseButtonDisplayAlways always shows the close button.
	TabCloseButtonDisplayAlways TabCloseButtonDisplayMode = iota
	// TabCloseButtonDisplayOnHover shows the close button on hover.
	TabCloseButtonDisplayOnHover
	// TabCloseButtonDisplayNever never shows the close button.
	TabCloseButtonDisplayNever
)

// TabToolButton is a transparent tool button used for the tab close/add
// buttons (port of tab_view.py's TabToolButton): 32x24 with the icon recolored
// to the tab-bar gray (#484848 light / #eaeaea dark).
type TabToolButton struct{ *TransparentToolButton }

func newTabToolButton(icon interface{}, parent *qt.QWidget) *TabToolButton {
	w := &TabToolButton{TransparentToolButton: NewTransparentToolButton(parent)}
	w.SetFixedSize2(32, 24)
	w.SetIconSize(qt.NewQSize2(12, 12))
	w.SetIcon(icon)
	w.drawIconFunc = w.drawTabIcon
	return w
}

func (w *TabToolButton) drawTabIcon(icon interface{}, painter *qt.QPainter, rect *qt.QRectF) {
	fill := "#484848"
	if common.IsDarkTheme() {
		fill = "#eaeaea"
	}
	if fi, ok := icon.(common.FluentIcon); ok {
		renderFluentIconWithFill(fi, painter, rect, fill)
		return
	}
	renderFluentIcon(icon, painter, rect, common.ThemeAuto)
}

// TabCloseButton is the close button of a TabItem (a TabToolButton carrying the
// Fluent Close icon at 10x10).
type TabCloseButton struct{ *TabToolButton }

// NewTabCloseButton builds a tab close button.
func NewTabCloseButton(parent *qt.QWidget) *TabCloseButton {
	w := &TabCloseButton{TabToolButton: newTabToolButton(common.Cancel, parent)}
	w.SetCursor(qt.NewQCursor2(qt.PointingHandCursor))
	w.SetIconSize(qt.NewQSize2(10, 10))
	return w
}

// TabItem is a single tab button with a close button and a selected state.
type TabItem struct {
	*qt.QPushButton
	icon                         interface{}
	routeKey                     string
	textColor                    *qt.QColor
	lightSelectedBackgroundColor *qt.QColor
	darkSelectedBackgroundColor  *qt.QColor
	closeButton                  *TabCloseButton
	shadowEffect                 *qt.QGraphicsDropShadowEffect
	isSelected                   bool
	isShadowEnabled              bool
	borderRadius                 int
	closeButtonDisplayMode       TabCloseButtonDisplayMode
	isHover                      bool
	isPressed                    bool
	onClosed                     func()
	onDoubleClicked              func()

	slideAni        *common.ProgressAnimation
	onSlideFinished func()
	onDragPress     func(x, y int)
	onDragMove      func(x, y int)
	onDragRelease   func(x, y int)
}

// NewTabItem builds a tab item.
func NewTabItem(text string, parent *qt.QWidget, icon interface{}) *TabItem {
	w := &TabItem{
		QPushButton:                  qt.NewQPushButton(parent),
		icon:                         icon,
		borderRadius:                 5,
		isShadowEnabled:              true,
		closeButtonDisplayMode:       TabCloseButtonDisplayAlways,
		lightSelectedBackgroundColor: qt.NewQColor3(249, 249, 249),
		darkSelectedBackgroundColor:  qt.NewQColor3(40, 40, 40),
	}
	w.SetText(text)
	if icon != nil {
		w.SetIcon(common.ToQIcon(icon))
	}
	common.SetFont(w.QWidget, 12, 400)
	w.SetFixedHeight(36)
	w.SetMaximumWidth(240)
	w.SetMinimumWidth(64)
	w.SetAttribute(qt.WA_LayoutUsesWidgetRect)

	// The tab prefers its maximum width so tabs stretch to fill the bar instead
	// of collapsing to their minimum (port of tab_view.py's TabItem.sizeHint).
	w.OnSizeHint(func(super func() *qt.QSize) *qt.QSize {
		sz := qt.NewQSize2(w.MaximumWidth(), 36)
		sz.GoGC()
		return sz
	})

	w.closeButton = NewTabCloseButton(w.QWidget)
	w.closeButton.OnClicked(func() {
		if w.onClosed != nil {
			w.onClosed()
		}
	})

	w.shadowEffect = qt.NewQGraphicsDropShadowEffect2(w.QObject)
	w.shadowEffect.SetBlurRadius(5)
	w.shadowEffect.SetOffset2(0, 1)
	w.SetGraphicsEffect(w.shadowEffect.QGraphicsEffect)

	w.SetSelected(false)
	w.installEvents()
	return w
}

func (w *TabItem) installEvents() {
	w.OnResizeEvent(func(super func(event *qt.QResizeEvent), event *qt.QResizeEvent) {
		super(event)
		w.closeButton.Move(w.Width()-6-w.closeButton.Width(), w.Height()/2-w.closeButton.Height()/2)
	})
	w.OnEnterEvent(func(super func(event *qt.QEvent), event *qt.QEvent) {
		super(event)
		w.isHover = true
		if w.closeButtonDisplayMode == TabCloseButtonDisplayOnHover {
			w.closeButton.Show()
		}
	})
	w.OnLeaveEvent(func(super func(event *qt.QEvent), event *qt.QEvent) {
		super(event)
		w.isHover = false
		if w.closeButtonDisplayMode == TabCloseButtonDisplayOnHover && !w.isSelected {
			w.closeButton.Hide()
		}
	})
	w.OnMousePressEvent(func(super func(event *qt.QMouseEvent), event *qt.QMouseEvent) {
		super(event)
		w.isPressed = true
		if w.onDragPress != nil && event.Button() == qt.LeftButton {
			p := event.Pos()
			w.onDragPress(w.X()+p.X(), w.Y()+p.Y())
		}
	})
	w.OnMouseMoveEvent(func(super func(event *qt.QMouseEvent), event *qt.QMouseEvent) {
		super(event)
		if w.onDragMove != nil {
			p := event.Pos()
			w.onDragMove(w.X()+p.X(), w.Y()+p.Y())
		}
	})
	w.OnMouseReleaseEvent(func(super func(event *qt.QMouseEvent), event *qt.QMouseEvent) {
		super(event)
		w.isPressed = false
		if w.onDragRelease != nil {
			p := event.Pos()
			w.onDragRelease(w.X()+p.X(), w.Y()+p.Y())
		}
	})
	w.OnMouseDoubleClickEvent(func(super func(event *qt.QMouseEvent), event *qt.QMouseEvent) {
		if event.Button() == qt.LeftButton && w.onDoubleClicked != nil {
			w.onDoubleClicked()
		}
		super(event)
	})
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing)

		if w.isSelected {
			w.drawSelectedBackground(painter)
		} else {
			w.drawNotSelectedBackground(painter)
		}

		if !w.isSelected {
			if common.IsDarkTheme() {
				painter.SetOpacity(0.79)
			} else {
				painter.SetOpacity(0.61)
			}
		}
		if w.icon != nil {
			rect := qt.NewQRectF4(10, 10, 16, 16)
			defer rect.Delete()
			common.DrawIcon(w.icon, painter, rect)
		}
		w.drawText(painter)
		painter.End()
	})
}

// SetRouteKey sets the route key of the tab.
func (w *TabItem) SetRouteKey(key string) { w.routeKey = key }

// RouteKey returns the route key of the tab.
func (w *TabItem) RouteKey() string { return w.routeKey }

// SetBorderRadius sets the background corner radius.
func (w *TabItem) SetBorderRadius(radius int) {
	w.borderRadius = radius
	w.Update()
}

// SetSelected sets the selected state and repaints.
func (w *TabItem) SetSelected(isSelected bool) {
	w.isSelected = isSelected
	alpha := 0
	if w.isSelected && w.isShadowEnabled {
		alpha = 50
	}
	color := qt.NewQColor11(0, 0, 0, alpha)
	defer color.Delete()
	w.shadowEffect.SetColor(color)
	w.Update()
	if isSelected {
		w.Raise()
	}
	if w.closeButtonDisplayMode == TabCloseButtonDisplayOnHover {
		w.closeButton.SetVisible(isSelected)
	}
}

// SetShadowEnabled toggles the tab shadow.
func (w *TabItem) SetShadowEnabled(isEnabled bool) {
	if isEnabled == w.isShadowEnabled {
		return
	}
	w.isShadowEnabled = isEnabled
	alpha := 0
	if w.isSelected && isEnabled {
		alpha = 50
	}
	color := qt.NewQColor11(0, 0, 0, alpha)
	defer color.Delete()
	w.shadowEffect.SetColor(color)
}

// SetCloseButtonDisplayMode sets the close button display mode.
func (w *TabItem) SetCloseButtonDisplayMode(mode TabCloseButtonDisplayMode) {
	if mode == w.closeButtonDisplayMode {
		return
	}
	w.closeButtonDisplayMode = mode
	switch mode {
	case TabCloseButtonDisplayNever:
		w.closeButton.Hide()
	case TabCloseButtonDisplayAlways:
		w.closeButton.Show()
	default:
		w.closeButton.SetVisible(w.isHover || w.isSelected)
	}
}

// slideTo animates the tab's x position to x over duration ms (port of
// TabItem.slideTo). A duration <= 0 falls back to the Python default (250ms).
func (w *TabItem) slideTo(x int, duration int) {
	w.stopSlideAni()
	if duration <= 0 {
		duration = 250
	}
	from := w.X()
	y := w.Y()
	curve := qt.NewQEasingCurve3(qt.QEasingCurve__InOutQuad)
	w.slideAni = common.NewProgressAnimation(duration, curve)
	curve.Delete()
	w.slideAni.OnProgress(func(t float64) {
		w.Move(from+int(float64(x-from)*t+0.5), y)
	})
	w.slideAni.OnFinished(func() {
		w.Move(x, y)
		if w.onSlideFinished != nil {
			w.onSlideFinished()
		}
	})
	w.slideAni.Start()
}

// stopSlideAni cancels an in-flight slide and releases the animation.
func (w *TabItem) stopSlideAni() {
	if w.slideAni != nil {
		w.slideAni.Stop()
		w.slideAni.Delete()
		w.slideAni = nil
	}
}

// SetTextColor sets the text color.
func (w *TabItem) SetTextColor(color *qt.QColor) {
	w.textColor = cloneColor(color)
	w.Update()
}

// SetSelectedBackgroundColor sets the light/dark selected background colors.
func (w *TabItem) SetSelectedBackgroundColor(light, dark *qt.QColor) {
	w.lightSelectedBackgroundColor = cloneColor(light)
	w.darkSelectedBackgroundColor = cloneColor(dark)
	w.Update()
}

// SetOnClosed registers the callback for the close button.
func (w *TabItem) SetOnClosed(f func()) { w.onClosed = f }

// SetOnDoubleClicked registers the callback for a double click.
func (w *TabItem) SetOnDoubleClicked(f func()) { w.onDoubleClicked = f }

func (w *TabItem) drawSelectedBackground(painter *qt.QPainter) {
	wd := float64(w.Width())
	h := float64(w.Height())
	r := float64(w.borderRadius)
	d := 2 * r
	isDark := common.IsDarkTheme()

	// top border
	path := qt.NewQPainterPath()
	path.ArcMoveTo2(1, h-d-1, d, d, 225)
	path.ArcTo2(1, h-d-1, d, d, 225, -45)
	path.LineTo2(1, r)
	path.ArcTo2(1, 1, d, d, -180, -90)
	path.LineTo2(wd-r, 1)
	path.ArcTo2(wd-d-1, 1, d, d, 90, -90)
	path.LineTo2(wd-1, h-r)
	path.ArcTo2(wd-d-1, h-d-1, d, d, 0, -45)

	var topBorder *qt.QColor
	if isDark {
		switch {
		case w.isPressed:
			topBorder = qt.NewQColor11(255, 255, 255, 18)
		case w.isHover:
			topBorder = qt.NewQColor11(255, 255, 255, 13)
		default:
			topBorder = qt.NewQColor11(0, 0, 0, 20)
		}
	} else {
		topBorder = qt.NewQColor11(0, 0, 0, 16)
	}
	defer topBorder.Delete()
	topPen := qt.NewQPen3(topBorder)
	defer topPen.Delete()
	painter.StrokePath(path, topPen)
	path.Delete()

	// bottom border
	path = qt.NewQPainterPath()
	path.ArcMoveTo2(1, h-d-1, d, d, 225)
	path.ArcTo2(1, h-d-1, d, d, 225, 45)
	path.LineTo2(wd-r-1, h-1)
	path.ArcTo2(wd-d-1, h-d-1, d, d, 270, 45)

	var bottomBorder *qt.QColor
	if !isDark {
		bottomBorder = qt.NewQColor11(0, 0, 0, 63)
	} else {
		bottomBorder = cloneColor(topBorder)
	}
	defer bottomBorder.Delete()
	bottomPen := qt.NewQPen3(bottomBorder)
	defer bottomPen.Delete()
	painter.StrokePath(path, bottomPen)
	path.Delete()

	// background
	painter.SetPenWithStyle(qt.NoPen)
	r2 := w.Rect()
	rect := r2.Adjusted(1, 1, -1, -1) // GoGC-armed — do NOT Delete

	bg := w.lightSelectedBackgroundColor
	if isDark {
		bg = w.darkSelectedBackgroundColor
	}
	brush := qt.NewQBrush3(bg)
	defer brush.Delete()
	painter.SetBrush(brush)
	painter.DrawRoundedRect3(rect, r, r)
}

func (w *TabItem) drawNotSelectedBackground(painter *qt.QPainter) {
	if !(w.isPressed || w.isHover) {
		return
	}
	var color *qt.QColor
	if common.IsDarkTheme() {
		if w.isPressed {
			color = qt.NewQColor11(255, 255, 255, 12)
		} else {
			color = qt.NewQColor11(255, 255, 255, 15)
		}
	} else {
		if w.isPressed {
			color = qt.NewQColor11(0, 0, 0, 7)
		} else {
			color = qt.NewQColor11(0, 0, 0, 10)
		}
	}
	defer color.Delete()
	brush := qt.NewQBrush3(color)
	defer brush.Delete()
	painter.SetBrush(brush)
	painter.SetPenWithStyle(qt.NoPen)
	r := w.Rect()
	rect := r.Adjusted(1, 1, -1, -1) // GoGC-armed — do NOT Delete

	painter.DrawRoundedRect3(rect, float64(w.borderRadius), float64(w.borderRadius))
}

func (w *TabItem) drawText(painter *qt.QPainter) {
	fm := w.FontMetrics() // GoGC-armed — do NOT Delete
	tw := fm.Width(w.Text())

	var dw int
	if w.Icon().IsNull() {
		if w.closeButton.IsVisible() {
			dw = 47
		} else {
			dw = 20
		}
	} else {
		if w.closeButton.IsVisible() {
			dw = 70
		} else {
			dw = 45
		}
	}
	x := 10
	if !w.Icon().IsNull() {
		x = 33
	}
	r := qt.NewQRect4(x, 0, w.Width()-dw, w.Height())
	defer r.Delete()

	owned := false
	var color *qt.QColor
	if w.textColor != nil {
		color = w.textColor
	} else {
		owned = true
		if common.IsDarkTheme() {
			color = qt.NewQColor3(255, 255, 255)
		} else {
			color = qt.NewQColor3(0, 0, 0)
		}
	}
	if owned {
		defer color.Delete()
	}

	pen := qt.NewQPen()
	defer pen.Delete()
	rw := r.Width()
	if tw > rw {
		// Fade the overflowing text out before it reaches the close button
		// (port of tab_view.py's TabItem._drawText linear-gradient).
		gradient := qt.NewQLinearGradient3(float64(r.X()), 0, float64(tw+r.X()), 0)
		defer gradient.Delete()
		transparent := qt.NewQColor11(0, 0, 0, 0)
		defer transparent.Delete()
		gradient.SetColorAt(0, color)
		gradient.SetColorAt(math.Max(0, float64(rw-10)/float64(tw)), color)
		gradient.SetColorAt(math.Max(0, float64(rw)/float64(tw)), transparent)
		gradient.SetColorAt(1, transparent)
		brush := qt.NewQBrush10(gradient.QGradient)
		defer brush.Delete()
		pen.SetBrush(brush)
	} else {
		pen.SetColor(color)
	}

	painter.SetPenWithPen(pen)
	painter.SetFont(w.Font())
	painter.DrawText6(r, int(qt.AlignVCenter|qt.AlignLeft), w.Text())
}

// TabBar is a scrollable tab bar that hosts a list of TabItem widgets.
type TabBar struct {
	*SingleDirectionScrollArea
	items                        []*TabItem
	itemMap                      map[string]*TabItem
	currentIndex                 int
	isMovable                    bool
	isScrollable                 bool
	isTabShadowEnabled           bool
	tabMaxWidth                  int
	tabMinWidth                  int
	lightSelectedBackgroundColor *qt.QColor
	darkSelectedBackgroundColor  *qt.QColor
	closeButtonDisplayMode       TabCloseButtonDisplayMode
	view                         *qt.QWidget
	hBoxLayout                   *qt.QHBoxLayout
	itemLayout                   *qt.QHBoxLayout
	widgetLayout                 *qt.QHBoxLayout
	addButton                    *TabToolButton
	dragPosX                     int
	dragPosY                     int
	isDraging                    bool
	onCurrentChanged             func(int)
	onTabBarClicked              func(int)
	onTabBarDoubleClicked        func(int)
	onTabCloseRequested          func(int)
	onTabAddRequested            func()
	onTabMoved                   func(int, int)
}

// NewTabBar builds a tab bar.
func NewTabBar(parent *qt.QWidget) *TabBar {
	w := &TabBar{
		SingleDirectionScrollArea:    NewSingleDirectionScrollArea(parent, qt.Horizontal),
		itemMap:                      make(map[string]*TabItem),
		currentIndex:                 -1,
		isTabShadowEnabled:           true,
		tabMaxWidth:                  240,
		tabMinWidth:                  64,
		lightSelectedBackgroundColor: qt.NewQColor3(249, 249, 249),
		darkSelectedBackgroundColor:  qt.NewQColor3(40, 40, 40),
		closeButtonDisplayMode:       TabCloseButtonDisplayAlways,
	}
	w.view = qt.NewQWidget(w.QWidget)
	w.hBoxLayout = qt.NewQHBoxLayout(w.view)
	w.itemLayout = qt.NewQHBoxLayout2()
	w.widgetLayout = qt.NewQHBoxLayout2()
	w.initWidget()
	return w
}

func (w *TabBar) initWidget() {
	w.SetFixedHeight(46)
	w.SetWidget(w.view)
	w.SetWidgetResizable(true)
	w.SetHorizontalScrollBarPolicy(qt.ScrollBarAlwaysOff)
	w.SetVerticalScrollBarPolicy(qt.ScrollBarAlwaysOff)

	w.hBoxLayout.SetSizeConstraint(qt.QLayout__SetMaximumSize)

	w.addButton = newTabToolButton(common.Add, w.QWidget)
	w.addButton.OnClicked(func() {
		if w.onTabAddRequested != nil {
			w.onTabAddRequested()
		}
	})
	// The add button is interactive; exclude it from the FramelessWindow
	// title-bar drag area so it receives its click (mirrors the tab items).
	w.addButton.SetProperty("isTitleBarButton", qt.NewQVariant11(true))

	w.view.SetObjectName("view")
	common.FluentStyleSheet(common.FluentTabView).Apply(w.QWidget, common.ThemeAuto)
	common.FluentStyleSheet(common.FluentTabView).Apply(w.view, common.ThemeAuto)

	// tab_view.qss uses a `TabBar` type selector, which only matches a C++
	// subclass named TabBar. The Go port is a plain QScrollArea, so that rule
	// never applies and the bar renders with an opaque background. Force the
	// scroll area (and its viewport) transparent directly.
	w.EnableTransparentBackground()

	w.initLayout()
	w.installSeparatorPaint()
}

func (w *TabBar) initLayout() {
	w.itemLayout.SetContentsMargins(5, 5, 5, 5)
	w.widgetLayout.SetContentsMargins(0, 0, 0, 0)
	w.hBoxLayout.SetContentsMargins(0, 0, 0, 0)
	w.itemLayout.SetSizeConstraint(qt.QLayout__SetMinAndMaxSize)
	w.hBoxLayout.SetSpacing(0)
	w.itemLayout.SetSpacing(0)

	w.hBoxLayout.AddLayout(w.itemLayout.QLayout)
	w.hBoxLayout.AddSpacing(3)
	w.widgetLayout.AddWidget3(w.addButton.QWidget, 0, qt.AlignLeft)
	w.hBoxLayout.AddLayout(w.widgetLayout.QLayout)
	w.hBoxLayout.AddStretchWithStretch(1)
}

func (w *TabBar) installSeparatorPaint() {
	w.view.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
		super(event)
		painter := qt.NewQPainter2(w.view.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing)
		var color *qt.QColor
		if common.IsDarkTheme() {
			color = qt.NewQColor11(255, 255, 255, 21)
		} else {
			color = qt.NewQColor11(0, 0, 0, 15)
		}
		defer color.Delete()
		painter.SetPen(color)
		for i, item := range w.items {
			canDraw := !(item.isHover || item.isSelected)
			if i < len(w.items)-1 {
				next := w.items[i+1]
				if next.isHover || next.isSelected {
					canDraw = false
				}
			}
			if canDraw {
				geom := item.Geometry() // borrowed — do NOT Delete
				x := geom.Right()
				y := w.Height()/2 - 8
				painter.DrawLine2(x, y, x, y+16)
			}
		}
		painter.End()
	})
}

// SetAddButtonVisible toggles the add button.
func (w *TabBar) SetAddButtonVisible(isVisible bool) {
	w.addButton.SetVisible(isVisible)
}

// AddTab adds a tab at the end.
func (w *TabBar) AddTab(routeKey, text string, icon interface{}, onClick func()) *TabItem {
	return w.InsertTab(-1, routeKey, text, icon, onClick)
}

// InsertTab inserts a tab at the given index.
func (w *TabBar) InsertTab(index int, routeKey, text string, icon interface{}, onClick func()) *TabItem {
	if _, ok := w.itemMap[routeKey]; ok {
		panic("The route key `" + routeKey + "` is duplicated.")
	}
	if index == -1 {
		index = len(w.items)
	}
	if index <= w.currentIndex && w.currentIndex >= 0 {
		w.currentIndex++
	}

	item := NewTabItem(text, w.view, icon)
	item.SetRouteKey(routeKey)

	// A tab item is interactive (clickable/closeable/draggable-to-reorder), so
	// exclude it from the FramelessWindow title-bar drag area. This mirrors the
	// reference CustomTitleBar.canDrag, which returns False for the tab region
	// only — the TabBar's empty area (and the rest of the title bar) stays
	// draggable.
	item.SetProperty("isTitleBarButton", qt.NewQVariant11(true))

	wd := w.tabMinWidth
	if w.isScrollable {
		wd = w.tabMaxWidth
	}
	item.SetMinimumWidth(wd)
	item.SetMaximumWidth(w.tabMaxWidth)
	item.SetShadowEnabled(w.isTabShadowEnabled)
	item.SetCloseButtonDisplayMode(w.closeButtonDisplayMode)
	item.SetSelectedBackgroundColor(w.lightSelectedBackgroundColor, w.darkSelectedBackgroundColor)

	item.OnPressed(func() {
		w.onItemPressed(item)
		if onClick != nil {
			onClick()
		}
	})
	item.SetOnDoubleClicked(func() {
		if w.onTabBarDoubleClicked != nil {
			w.onTabBarDoubleClicked(w.indexOfItem(item))
		}
	})
	item.SetOnClosed(func() {
		if w.onTabCloseRequested != nil {
			w.onTabCloseRequested(w.indexOfItem(item))
		}
	})

	// Wire drag reorder (port of TabBar._onDragMove / _swapItem / _adjustLayout).
	item.onDragPress = w.onItemDragPress
	item.onDragMove = w.onItemDragMove
	item.onDragRelease = w.onItemDragRelease

	w.itemLayout.InsertWidget2(index, item.QWidget, 1)
	w.items = append(w.items[:index], append([]*TabItem{item}, w.items[index:]...)...)
	w.itemMap[routeKey] = item

	if len(w.items) == 1 {
		w.SetCurrentIndex(0)
	}
	return item
}

// RemoveTab removes the tab at the given index.
func (w *TabBar) RemoveTab(index int) {
	if !(0 <= index && index < len(w.items)) {
		return
	}
	if index < w.currentIndex {
		w.currentIndex--
	} else if index == w.currentIndex {
		switch {
		case w.currentIndex > 0:
			w.SetCurrentIndex(w.currentIndex - 1)
			if w.onCurrentChanged != nil {
				w.onCurrentChanged(w.currentIndex)
			}
		case len(w.items) == 1:
			w.currentIndex = -1
		default:
			w.SetCurrentIndex(1)
			w.currentIndex = 0
			if w.onCurrentChanged != nil {
				w.onCurrentChanged(0)
			}
		}
	}

	item := w.items[index]
	w.items = append(w.items[:index], w.items[index+1:]...)
	delete(w.itemMap, item.RouteKey())
	w.itemLayout.RemoveWidget(item.QWidget)
	common.RouterInstance.Remove(item.RouteKey())
	item.DeleteLater()
	w.Update()
}

// RemoveTabByKey removes the tab with the given route key.
func (w *TabBar) RemoveTabByKey(routeKey string) {
	item, ok := w.itemMap[routeKey]
	if !ok {
		return
	}
	w.RemoveTab(w.indexOfItem(item))
}

// SetCurrentIndex sets the current tab index.
func (w *TabBar) SetCurrentIndex(index int) {
	if index == w.currentIndex {
		return
	}
	if w.currentIndex >= 0 && w.currentIndex < len(w.items) {
		w.items[w.currentIndex].SetSelected(false)
	}
	w.currentIndex = index
	w.items[index].SetSelected(true)
}

// SetCurrentTab sets the current tab by route key.
func (w *TabBar) SetCurrentTab(routeKey string) {
	item, ok := w.itemMap[routeKey]
	if !ok {
		return
	}
	w.SetCurrentIndex(w.indexOfItem(item))
}

// CurrentIndex returns the current tab index.
func (w *TabBar) CurrentIndex() int { return w.currentIndex }

// CurrentTab returns the current tab item.
func (w *TabBar) CurrentTab() *TabItem { return w.TabItem(w.currentIndex) }

func (w *TabBar) indexOfItem(item *TabItem) int {
	for i, it := range w.items {
		if it == item {
			return i
		}
	}
	return -1
}

func (w *TabBar) onItemPressed(item *TabItem) {
	for _, it := range w.items {
		it.SetSelected(it == item)
	}
	index := w.indexOfItem(item)
	if w.onTabBarClicked != nil {
		w.onTabBarClicked(index)
	}
	if index != w.currentIndex {
		w.SetCurrentIndex(index)
		if w.onCurrentChanged != nil {
			w.onCurrentChanged(index)
		}
	}
}

// tabRectX returns the logical x position of the tab at index (sum of the
// widths of every preceding tab), matching TabBar.tabRect in the Python port.
func (w *TabBar) tabRectX(index int) int {
	x := 0
	for i := 0; i < index; i++ {
		x += w.items[i].Width()
	}
	return x
}

func (w *TabBar) onItemDragPress(x, y int) {
	if !w.isMovable || !w.itemLayout.Geometry().Contains2(x, y) {
		return
	}
	w.dragPosX = x
	w.dragPosY = y
}

func (w *TabBar) onItemDragMove(x, y int) {
	if !w.isMovable || w.Count() <= 1 || !w.itemLayout.Geometry().Contains2(x, y) {
		return
	}
	if w.currentIndex < 0 || w.currentIndex >= len(w.items) {
		return
	}

	index := w.currentIndex
	item := w.items[index]
	dx := x - w.dragPosX
	w.dragPosX = x
	w.dragPosY = y

	// First tab can't move left; last tab can't move right.
	if index == 0 && dx < 0 && item.X() <= 0 {
		return
	}
	if index == w.Count()-1 && dx > 0 {
		sh := w.itemLayout.SizeHint() // GoGC-armed — do NOT Delete
		if item.Geometry().Right() >= sh.Width() {
			return
		}
	}

	item.Move(item.X()+dx, item.Y())
	w.isDraging = true

	// Move the left sibling item to the right.
	if dx < 0 && index > 0 {
		siblingIndex := index - 1
		center := w.items[siblingIndex].Geometry().Center() // GoGC-armed — do NOT Delete
		if item.X() < center.X() {
			w.swapItem(siblingIndex)
		}
	} else if dx > 0 && index < w.Count()-1 {
		// Move the right sibling item to the left.
		siblingIndex := index + 1
		center := w.items[siblingIndex].Geometry().Center() // GoGC-armed — do NOT Delete
		if item.Geometry().Right() > center.X() {
			w.swapItem(siblingIndex)
		}
	}
}

func (w *TabBar) onItemDragRelease(x, y int) {
	if !w.isMovable || !w.isDraging {
		return
	}
	w.isDraging = false
	if w.currentIndex < 0 || w.currentIndex >= len(w.items) {
		return
	}

	item := w.items[w.currentIndex]
	target := w.tabRectX(w.currentIndex)
	d := item.X() - target
	if d < 0 {
		d = -d
	}
	duration := d * 250 / item.Width()
	// Only the release slide settles the layout (swap slides stay async, exactly
	// like the Python port which only connects finished on mouse release).
	item.onSlideFinished = w.adjustLayout
	item.slideTo(target, duration)
}

// swapItem swaps the current item with the item at index and slides the swapped
// item back into place (port of TabBar._swapItem).
func (w *TabBar) swapItem(index int) {
	if index < 0 || index >= len(w.items) {
		return
	}
	oldIndex := w.currentIndex
	if oldIndex < 0 || oldIndex >= len(w.items) {
		return
	}

	swappedItem := w.items[index]
	x := w.tabRectX(oldIndex)
	w.items[oldIndex], w.items[index] = w.items[index], w.items[oldIndex]
	w.currentIndex = index
	swappedItem.slideTo(x, 250)
	if w.onTabMoved != nil {
		w.onTabMoved(oldIndex, index)
	}
}

// adjustLayout re-adds every tab to the item layout in list order, settling the
// layout after a drag completes (port of TabBar._adjustLayout).
func (w *TabBar) adjustLayout() {
	for _, item := range w.items {
		w.itemLayout.RemoveWidget(item.QWidget)
	}
	for _, item := range w.items {
		w.itemLayout.AddWidget(item.QWidget)
	}
	for _, item := range w.items {
		item.onSlideFinished = nil
	}
}

// SetCloseButtonDisplayMode sets the close button display mode for all tabs.
func (w *TabBar) SetCloseButtonDisplayMode(mode TabCloseButtonDisplayMode) {
	if mode == w.closeButtonDisplayMode {
		return
	}
	w.closeButtonDisplayMode = mode
	for _, item := range w.items {
		item.SetCloseButtonDisplayMode(mode)
	}
}

// TabItem returns the tab item at the given index (nil when out of range).
func (w *TabBar) TabItem(index int) *TabItem {
	if index < 0 || index >= len(w.items) {
		return nil
	}
	return w.items[index]
}

// Tab returns the tab item with the given route key.
func (w *TabBar) Tab(routeKey string) *TabItem { return w.itemMap[routeKey] }

// TabRegion returns the bounding rect of all tabs.
func (w *TabBar) TabRegion() *qt.QRect { return w.itemLayout.Geometry() }

// TabRect returns the rect of the tab at the given index.
func (w *TabBar) TabRect(index int) *qt.QRect {
	item := w.TabItem(index)
	if item == nil {
		return nil
	}
	return item.Geometry()
}

// TabData returns the "data" property of the tab at the index.
func (w *TabBar) TabData(index int) *qt.QVariant {
	item := w.TabItem(index)
	if item == nil {
		return nil
	}
	return item.Property("data")
}

// SetTabData sets the "data" property of the tab at the index.
func (w *TabBar) SetTabData(index int, data *qt.QVariant) {
	item := w.TabItem(index)
	if item == nil {
		return
	}
	item.SetProperty("data", data)
}

// TabText returns the text of the tab at the index.
func (w *TabBar) TabText(index int) string {
	item := w.TabItem(index)
	if item == nil {
		return ""
	}
	return item.Text()
}

// TabIcon returns the icon of the tab at the index.
func (w *TabBar) TabIcon(index int) *qt.QIcon {
	item := w.TabItem(index)
	if item == nil {
		return nil
	}
	return item.Icon()
}

// TabToolTip returns the tool tip of the tab at the index.
func (w *TabBar) TabToolTip(index int) string {
	item := w.TabItem(index)
	if item == nil {
		return ""
	}
	return item.ToolTip()
}

// IsTabEnabled reports whether the tab at the index is enabled.
func (w *TabBar) IsTabEnabled(index int) bool {
	item := w.TabItem(index)
	if item == nil {
		return false
	}
	return item.IsEnabled()
}

// SetTabEnabled enables/disables the tab at the index.
func (w *TabBar) SetTabEnabled(index int, isEnabled bool) {
	item := w.TabItem(index)
	if item == nil {
		return
	}
	item.SetEnabled(isEnabled)
}

// SetTabsClosable sets whether tabs are closable.
func (w *TabBar) SetTabsClosable(isClosable bool) {
	if isClosable {
		w.SetCloseButtonDisplayMode(TabCloseButtonDisplayAlways)
	} else {
		w.SetCloseButtonDisplayMode(TabCloseButtonDisplayNever)
	}
}

// TabsClosable reports whether tabs are closable.
func (w *TabBar) TabsClosable() bool {
	return w.closeButtonDisplayMode != TabCloseButtonDisplayNever
}

// SetTabIcon sets the icon of the tab at the index.
func (w *TabBar) SetTabIcon(index int, icon interface{}) {
	item := w.TabItem(index)
	if item == nil {
		return
	}
	item.SetIcon(common.ToQIcon(icon))
}

// SetTabText sets the text of the tab at the index.
func (w *TabBar) SetTabText(index int, text string) {
	item := w.TabItem(index)
	if item == nil {
		return
	}
	item.SetText(text)
}

// IsTabVisible reports whether the tab at the index is visible.
func (w *TabBar) IsTabVisible(index int) bool {
	item := w.TabItem(index)
	if item == nil {
		return false
	}
	return item.IsVisible()
}

// SetTabVisible sets the visibility of the tab at the index.
func (w *TabBar) SetTabVisible(index int, isVisible bool) {
	item := w.TabItem(index)
	if item == nil {
		return
	}
	item.SetVisible(isVisible)
	if isVisible && w.currentIndex < 0 {
		w.SetCurrentIndex(0)
	} else if !isVisible {
		switch {
		case w.currentIndex > 0:
			w.SetCurrentIndex(w.currentIndex - 1)
			if w.onCurrentChanged != nil {
				w.onCurrentChanged(w.currentIndex)
			}
		case len(w.items) == 1:
			w.currentIndex = -1
		default:
			w.SetCurrentIndex(1)
			w.currentIndex = 0
			if w.onCurrentChanged != nil {
				w.onCurrentChanged(0)
			}
		}
	}
}

// SetTabTextColor sets the text color of the tab at the index.
func (w *TabBar) SetTabTextColor(index int, color *qt.QColor) {
	item := w.TabItem(index)
	if item == nil {
		return
	}
	item.SetTextColor(color)
}

// SetTabToolTip sets the tool tip of the tab at the index.
func (w *TabBar) SetTabToolTip(index int, toolTip string) {
	item := w.TabItem(index)
	if item == nil {
		return
	}
	item.SetToolTip(toolTip)
}

// SetTabSelectedBackgroundColor sets the selected background color of all tabs.
func (w *TabBar) SetTabSelectedBackgroundColor(light, dark *qt.QColor) {
	w.lightSelectedBackgroundColor = cloneColor(light)
	w.darkSelectedBackgroundColor = cloneColor(dark)
	for _, item := range w.items {
		item.SetSelectedBackgroundColor(light, dark)
	}
}

// SetTabShadowEnabled toggles the shadow of all tabs.
func (w *TabBar) SetTabShadowEnabled(isEnabled bool) {
	if isEnabled == w.isTabShadowEnabled {
		return
	}
	w.isTabShadowEnabled = isEnabled
	for _, item := range w.items {
		item.SetShadowEnabled(isEnabled)
	}
}

// IsTabShadowEnabled reports whether tab shadows are enabled.
func (w *TabBar) IsTabShadowEnabled() bool { return w.isTabShadowEnabled }

// SetMovable toggles tab dragging.
func (w *TabBar) SetMovable(movable bool) { w.isMovable = movable }

// IsMovable reports whether tabs are movable.
func (w *TabBar) IsMovable() bool { return w.isMovable }

// SetScrollable toggles whether tabs use their maximum width.
func (w *TabBar) SetScrollable(scrollable bool) {
	w.isScrollable = scrollable
	wd := w.tabMinWidth
	if scrollable {
		wd = w.tabMaxWidth
	}
	for _, item := range w.items {
		item.SetMinimumWidth(wd)
	}
}

// SetTabMaximumWidth sets the maximum width of tabs.
func (w *TabBar) SetTabMaximumWidth(width int) {
	if width == w.tabMaxWidth {
		return
	}
	w.tabMaxWidth = width
	for _, item := range w.items {
		item.SetMaximumWidth(width)
	}
}

// SetTabMinimumWidth sets the minimum width of tabs.
func (w *TabBar) SetTabMinimumWidth(width int) {
	if width == w.tabMinWidth {
		return
	}
	w.tabMinWidth = width
	if !w.isScrollable {
		for _, item := range w.items {
			item.SetMinimumWidth(width)
		}
	}
}

// TabMaximumWidth returns the maximum tab width.
func (w *TabBar) TabMaximumWidth() int { return w.tabMaxWidth }

// TabMinimumWidth returns the minimum tab width.
func (w *TabBar) TabMinimumWidth() int { return w.tabMinWidth }

// IsScrollable reports whether tabs are scrollable.
func (w *TabBar) IsScrollable() bool { return w.isScrollable }

// Count returns the number of tabs.
func (w *TabBar) Count() int { return len(w.items) }

// Clear removes all tabs.
func (w *TabBar) Clear() {
	for w.Count() > 0 {
		w.RemoveTab(w.Count() - 1)
	}
}

// OnCurrentChanged registers the current-changed callback.
func (w *TabBar) OnCurrentChanged(f func(int)) { w.onCurrentChanged = f }

// OnTabBarClicked registers the tab-clicked callback.
func (w *TabBar) OnTabBarClicked(f func(int)) { w.onTabBarClicked = f }

// OnTabBarDoubleClicked registers the tab-double-clicked callback.
func (w *TabBar) OnTabBarDoubleClicked(f func(int)) { w.onTabBarDoubleClicked = f }

// OnTabCloseRequested registers the tab-close-requested callback.
func (w *TabBar) OnTabCloseRequested(f func(int)) { w.onTabCloseRequested = f }

// OnTabAddRequested registers the tab-add-requested callback.
func (w *TabBar) OnTabAddRequested(f func()) { w.onTabAddRequested = f }

// OnTabMoved registers the tab-moved callback.
func (w *TabBar) OnTabMoved(f func(int, int)) { w.onTabMoved = f }

// TabWidget is a tab bar paired with a stacked widget.
type TabWidget struct {
	*qt.QWidget
	tabBar                *TabBar
	stackedWidget         *qt.QStackedWidget
	vBoxLayout            *qt.QVBoxLayout
	onCurrentChanged      func(int)
	onTabBarClicked       func(int)
	onTabCloseRequested   func(int)
	onTabAddRequested     func()
	onTabBarDoubleClicked func(int)
}

// NewTabWidget builds a tab widget.
func NewTabWidget(parent *qt.QWidget) *TabWidget {
	w := &TabWidget{QWidget: qt.NewQWidget(parent)}
	w.tabBar = NewTabBar(w.QWidget)
	w.stackedWidget = qt.NewQStackedWidget(w.QWidget)
	w.vBoxLayout = qt.NewQVBoxLayout(w.QWidget)
	w.vBoxLayout.AddWidget3(w.tabBar.QWidget, 0, qt.AlignTop)
	w.vBoxLayout.AddWidget2(w.stackedWidget.QWidget, 1)
	w.vBoxLayout.SetSpacing(1)
	w.vBoxLayout.SetContentsMargins(0, 0, 0, 0)
	w.connectTabBarSignals()
	return w
}

func (w *TabWidget) connectTabBarSignals() {
	w.tabBar.OnTabCloseRequested(func(index int) {
		if w.onTabCloseRequested != nil {
			w.onTabCloseRequested(index)
		}
	})
	w.tabBar.OnTabBarClicked(func(index int) {
		if w.onTabBarClicked != nil {
			w.onTabBarClicked(index)
		}
	})
	w.tabBar.OnTabBarDoubleClicked(func(index int) {
		if w.onTabBarDoubleClicked != nil {
			w.onTabBarDoubleClicked(index)
		}
	})
	w.tabBar.OnTabAddRequested(func() {
		if w.onTabAddRequested != nil {
			w.onTabAddRequested()
		}
	})
	w.tabBar.OnCurrentChanged(func(index int) {
		w.stackedWidget.SetCurrentIndex(index)
		if w.onCurrentChanged != nil {
			w.onCurrentChanged(index)
		}
	})
	w.tabBar.OnTabMoved(func(from int, to int) {
		widget := w.stackedWidget.Widget(from)
		w.stackedWidget.RemoveWidget(widget)
		w.stackedWidget.InsertWidget(to, widget)
		w.stackedWidget.SetCurrentIndex(to)
	})
}

// AddPage adds a tab with the given page and returns its index.
func (w *TabWidget) AddPage(widget *qt.QWidget, label string, icon interface{}, routeKey string) int {
	return w.InsertTab(-1, widget, label, icon, routeKey)
}

// AddTab adds a tab with the given page and returns its index.
func (w *TabWidget) AddTab(widget *qt.QWidget, label string, icon interface{}, routeKey string) int {
	return w.InsertTab(-1, widget, label, icon, routeKey)
}

// InsertTab inserts a tab with the given page at the index.
func (w *TabWidget) InsertTab(index int, widget *qt.QWidget, label string, icon interface{}, routeKey string) int {
	if w.stackedWidget.IndexOf(widget) >= 0 {
		return -1
	}
	if routeKey == "" {
		routeKey = newTabRouteKey()
	}
	widget.SetProperty("routeKey", qt.NewQVariant14(routeKey))
	w.tabBar.InsertTab(index, routeKey, label, icon, nil)
	w.stackedWidget.InsertWidget(index, widget)
	return w.stackedWidget.IndexOf(widget)
}

// RemoveTab removes the tab at the index (the page widget is not deleted).
func (w *TabWidget) RemoveTab(index int) {
	if !(0 <= index && index < w.stackedWidget.Count()) {
		return
	}
	w.stackedWidget.RemoveWidget(w.stackedWidget.Widget(index))
	w.tabBar.RemoveTab(index)
}

// Clear removes all pages (without deleting them).
func (w *TabWidget) Clear() {
	for w.stackedWidget.Count() > 0 {
		w.stackedWidget.RemoveWidget(w.stackedWidget.Widget(0))
	}
	w.tabBar.Clear()
}

// Widget returns the page at the index.
func (w *TabWidget) Widget(index int) *qt.QWidget { return w.stackedWidget.Widget(index) }

// CurrentWidget returns the currently displayed page.
func (w *TabWidget) CurrentWidget() *qt.QWidget { return w.stackedWidget.CurrentWidget() }

// CurrentIndex returns the index of the current page.
func (w *TabWidget) CurrentIndex() int { return w.stackedWidget.CurrentIndex() }

// SetTabBar replaces the tab bar.
func (w *TabWidget) SetTabBar(tabBar *TabBar) {
	if tabBar == w.tabBar {
		return
	}
	if w.tabBar != nil {
		w.vBoxLayout.RemoveWidget(w.tabBar.QWidget)
		w.tabBar.DeleteLater()
		w.tabBar.Hide()
	}
	w.tabBar = tabBar
	w.vBoxLayout.InsertWidget(0, w.tabBar.QWidget)
	w.connectTabBarSignals()
}

// IsMovable reports whether tabs can be moved.
func (w *TabWidget) IsMovable() bool { return w.tabBar.IsMovable() }

// SetMovable toggles tab moving.
func (w *TabWidget) SetMovable(movable bool) { w.tabBar.SetMovable(movable) }

// IsTabEnabled reports whether the tab is enabled.
func (w *TabWidget) IsTabEnabled(index int) bool { return w.tabBar.IsTabEnabled(index) }

// SetTabEnabled enables/disables the tab.
func (w *TabWidget) SetTabEnabled(index int, isEnabled bool) {
	w.tabBar.SetTabEnabled(index, isEnabled)
}

// IsTabVisible reports whether the tab is visible.
func (w *TabWidget) IsTabVisible(index int) bool { return w.tabBar.IsTabVisible(index) }

// SetTabVisible sets the tab visibility.
func (w *TabWidget) SetTabVisible(index int, isVisible bool) {
	w.tabBar.SetTabVisible(index, isVisible)
}

// TabText returns the tab text.
func (w *TabWidget) TabText(index int) string { return w.tabBar.TabText(index) }

// TabIcon returns the tab icon.
func (w *TabWidget) TabIcon(index int) *qt.QIcon { return w.tabBar.TabIcon(index) }

// TabToolTip returns the tab tool tip.
func (w *TabWidget) TabToolTip(index int) string { return w.tabBar.TabToolTip(index) }

// SetTabsClosable sets whether tabs are closable.
func (w *TabWidget) SetTabsClosable(closable bool) { w.tabBar.SetTabsClosable(closable) }

// TabsClosable reports whether tabs are closable.
func (w *TabWidget) TabsClosable() bool { return w.tabBar.TabsClosable() }

// SetTabIcon sets the tab icon.
func (w *TabWidget) SetTabIcon(index int, icon interface{}) { w.tabBar.SetTabIcon(index, icon) }

// SetTabText sets the tab text.
func (w *TabWidget) SetTabText(index int, text string) { w.tabBar.SetTabText(index, text) }

// SetTabToolTip sets the tab tool tip.
func (w *TabWidget) SetTabToolTip(index int, tip string) { w.tabBar.SetTabToolTip(index, tip) }

// SetTabTextColor sets the tab text color.
func (w *TabWidget) SetTabTextColor(index int, color *qt.QColor) {
	w.tabBar.SetTabTextColor(index, color)
}

// SetTabSelectedBackgroundColor sets the selected background color of all tabs.
func (w *TabWidget) SetTabSelectedBackgroundColor(light, dark *qt.QColor) {
	w.tabBar.SetTabSelectedBackgroundColor(light, dark)
}

// SetTabShadowEnabled toggles tab shadows.
func (w *TabWidget) SetTabShadowEnabled(enabled bool) { w.tabBar.SetTabShadowEnabled(enabled) }

// SetScrollable toggles scrollable tabs.
func (w *TabWidget) SetScrollable(scrollable bool) { w.tabBar.SetScrollable(scrollable) }

// IsScrollable reports whether tabs are scrollable.
func (w *TabWidget) IsScrollable() bool { return w.tabBar.IsScrollable() }

// SetTabMaximumWidth sets the maximum tab width.
func (w *TabWidget) SetTabMaximumWidth(width int) { w.tabBar.SetTabMaximumWidth(width) }

// SetTabMinimumWidth sets the minimum tab width.
func (w *TabWidget) SetTabMinimumWidth(width int) { w.tabBar.SetTabMinimumWidth(width) }

// TabMaximumWidth returns the maximum tab width.
func (w *TabWidget) TabMaximumWidth() int { return w.tabBar.TabMaximumWidth() }

// TabMinimumWidth returns the minimum tab width.
func (w *TabWidget) TabMinimumWidth() int { return w.tabBar.TabMinimumWidth() }

// TabData returns the tab data.
func (w *TabWidget) TabData(index int) *qt.QVariant { return w.tabBar.TabData(index) }

// SetTabData sets the tab data.
func (w *TabWidget) SetTabData(index int, data *qt.QVariant) { w.tabBar.SetTabData(index, data) }

// Count returns the number of tabs.
func (w *TabWidget) Count() int { return w.stackedWidget.Count() }

// SetCurrentIndex sets the current tab index.
func (w *TabWidget) SetCurrentIndex(index int) {
	w.tabBar.SetCurrentIndex(index)
	w.stackedWidget.SetCurrentIndex(index)
}

// SetCurrentWidget sets the current tab to the one containing the widget.
func (w *TabWidget) SetCurrentWidget(widget *qt.QWidget) {
	index := w.stackedWidget.IndexOf(widget)
	if index != -1 {
		w.SetCurrentIndex(index)
	}
}

// SetCloseButtonDisplayMode sets the close button display mode.
func (w *TabWidget) SetCloseButtonDisplayMode(mode TabCloseButtonDisplayMode) {
	w.tabBar.SetCloseButtonDisplayMode(mode)
}

// OnCurrentChanged registers the current-changed callback.
func (w *TabWidget) OnCurrentChanged(f func(int)) { w.onCurrentChanged = f }

// OnTabBarClicked registers the tab-clicked callback.
func (w *TabWidget) OnTabBarClicked(f func(int)) { w.onTabBarClicked = f }

// OnTabCloseRequested registers the tab-close-requested callback.
func (w *TabWidget) OnTabCloseRequested(f func(int)) { w.onTabCloseRequested = f }

// OnTabAddRequested registers the tab-add-requested callback.
func (w *TabWidget) OnTabAddRequested(f func()) { w.onTabAddRequested = f }

// OnTabBarDoubleClicked registers the tab-double-clicked callback.
func (w *TabWidget) OnTabBarDoubleClicked(f func(int)) { w.onTabBarDoubleClicked = f }

var tabRouteKeyCounter uint64

func newTabRouteKey() string {
	tabRouteKeyCounter++
	return fmt.Sprintf("%x", tabRouteKeyCounter)
}
