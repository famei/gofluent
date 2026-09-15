package navigation

import (
	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	qt "github.com/mappu/miqt/qt"
)

// NavigationBarPushButton is a large bottom-navigation button (icon over text).
// The icon slides down/up when the selected state changes.
type NavigationBarPushButton struct {
	*NavigationPushButton
	iconOffset            float64
	selectedIcon          interface{}
	isSelectedTextVisible bool
	lightSelectedColor    *qt.QColor
	darkSelectedColor     *qt.QColor
	iconAni               *common.ProgressAnimation
}

// NewNavigationBarPushButton builds a navigation bar push button.
func NewNavigationBarPushButton(icon interface{}, text string, isSelectable bool, selectedIcon interface{}, parent *qt.QWidget) *NavigationBarPushButton {
	w := &NavigationBarPushButton{
		NavigationPushButton:  NewNavigationPushButton(icon, text, isSelectable, parent),
		selectedIcon:          selectedIcon,
		isSelectedTextVisible: true,
		lightSelectedColor:    qt.NewQColor(),
		darkSelectedColor:     qt.NewQColor(),
	}
	w.SetFixedSize2(64, 58)
	common.SetFont(w.QWidget, 11, int(qt.QFont__Normal))
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) { w.paint() })
	return w
}

// SetSelectedColor sets the light/dark selected colors.
func (w *NavigationBarPushButton) SetSelectedColor(light, dark interface{}) {
	if w.lightSelectedColor != nil {
		w.lightSelectedColor.Delete()
	}
	if w.darkSelectedColor != nil {
		w.darkSelectedColor.Delete()
	}
	w.lightSelectedColor = coerceColor(light)
	w.darkSelectedColor = coerceColor(dark)
	w.Update()
}

// SetSelectedIcon sets the icon shown when selected.
func (w *NavigationBarPushButton) SetSelectedIcon(icon interface{}) {
	w.selectedIcon = icon
	w.Update()
}

// SetSelectedTextVisible toggles the text label visibility in the selected
// state.
func (w *NavigationBarPushButton) SetSelectedTextVisible(isVisible bool) {
	w.isSelectedTextVisible = isVisible
	w.Update()
}

// IndicatorRect returns the indicator geometry for a bar button.
func (w *NavigationBarPushButton) IndicatorRect() *qt.QRectF {
	return qt.NewQRectF4(0, 16, 4, 24)
}

// SetSelected sets the selected state and slides the icon over 100ms.
func (w *NavigationBarPushButton) SetSelected(isSelected bool) {
	if isSelected == w.IsSelected {
		return
	}
	w.IsSelected = isSelected
	w.IsAboutSelected = false
	w.stopIconAni()

	target := 0.0
	if isSelected {
		target = 6
	}
	from := w.iconOffset
	curve := qt.NewQEasingCurve3(qt.QEasingCurve__OutQuad)
	w.iconAni = common.NewProgressAnimation(100, curve)
	curve.Delete()
	w.iconAni.OnProgress(func(t float64) {
		w.iconOffset = from + (target-from)*t
		w.Update()
	})
	w.iconAni.OnFinished(func() {
		w.iconOffset = target
		w.Update()
	})
	w.iconAni.Start()
}

func (w *NavigationBarPushButton) stopIconAni() {
	if w.iconAni != nil {
		w.iconAni.Stop()
		w.iconAni.Delete()
		w.iconAni = nil
	}
}

func (w *NavigationBarPushButton) paint() {
	painter := qt.NewQPainter2(w.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__Antialiasing | qt.QPainter__TextAntialiasing | qt.QPainter__SmoothPixmapTransform)
	painter.SetPenWithStyle(qt.NoPen)

	w.drawBackground(painter)
	w.drawIcon(painter)
	w.drawText(painter)
	painter.End()
}

func (w *NavigationBarPushButton) drawBackground(painter *qt.QPainter) {
	if w.IsSelected || w.IsAboutSelected {
		bg := qt.NewQColor11(255, 255, 255, 42)
		if !common.IsDarkTheme() {
			bg = qt.NewQColor3(255, 255, 255)
		}
		brush := qt.NewQBrush3(bg)
		bg.Delete()
		defer brush.Delete()
		painter.SetBrush(brush)
		rect := w.Rect()
		painter.DrawRoundedRect3(rect, 5, 5)

		if !w.IsAboutSelected {
			color := fallbackThemeColor(w.lightSelectedColor, w.darkSelectedColor)
			brush2 := qt.NewQBrush3(color)
			color.Delete()
			defer brush2.Delete()
			painter.SetBrush(brush2)
			if !w.IsPressed {
				painter.DrawRoundedRect2(0, 16, 4, 24, 2, 2)
			} else {
				painter.DrawRoundedRect2(0, 19, 4, 18, 2, 2)
			}
		}
	} else if w.IsPressed || w.IsEnter {
		c := 0
		if common.IsDarkTheme() {
			c = 255
		}
		alpha := 6
		if w.IsEnter {
			alpha = 9
		}
		bg := qt.NewQColor11(c, c, c, alpha)
		brush := qt.NewQBrush3(bg)
		bg.Delete()
		defer brush.Delete()
		painter.SetBrush(brush)
		rect := w.Rect()
		painter.DrawRoundedRect3(rect, 5, 5)

	}
}

func (w *NavigationBarPushButton) drawIcon(painter *qt.QPainter) {
	if (w.IsPressed || !w.IsEnter) && !(w.IsSelected || w.IsAboutSelected) {
		painter.SetOpacity(0.6)
	}
	if !w.IsEnabled() {
		painter.SetOpacity(0.4)
	}

	var rect *qt.QRectF
	if w.isSelectedTextVisible {
		rect = qt.NewQRectF4(22, 13, 20, 20)
	} else {
		rect = qt.NewQRectF4(22, 13+w.iconOffset, 20, 20)
	}
	defer rect.Delete()

	selectedIcon := w.icon
	if w.selectedIcon != nil {
		selectedIcon = w.selectedIcon
	}
	if w.IsSelected || w.IsAboutSelected {
		common.DrawIcon(selectedIcon, painter, rect)
	} else {
		common.DrawIcon(w.icon, painter, rect)
	}
}

func (w *NavigationBarPushButton) drawText(painter *qt.QPainter) {
	if w.IsSelected && !w.isSelectedTextVisible {
		return
	}
	if w.IsSelected || w.IsAboutSelected {
		color := fallbackThemeColor(w.lightSelectedColor, w.darkSelectedColor)
		painter.SetPen(color)
		color.Delete()
	} else {
		textColor := qt.NewQColor3(255, 255, 255)
		if !common.IsDarkTheme() {
			textColor = qt.NewQColor3(0, 0, 0)
		}
		painter.SetPen(textColor)
		textColor.Delete()
	}
	painter.SetFont(w.Font())
	rect := qt.NewQRect4(0, 32, w.Width(), 26)
	painter.DrawText6(rect, int(qt.AlignCenter), w.text)
	rect.Delete()
}

// ---------------------------------------------------------------------------
// NavigationBar

// NavigationBar is a vertical navigation bar for bottom navigation layouts.
type NavigationBar struct {
	*qt.QWidget
	indicator                   *NavigationIndicator
	isIndicatorAnimationEnabled bool
	isSelectedTextVisible       bool
	lightSelectedColor          *qt.QColor
	darkSelectedColor           *qt.QColor

	scrollArea   *widgets.ScrollArea
	scrollWidget *qt.QWidget

	vBoxLayout   *qt.QVBoxLayout
	topLayout    *qt.QVBoxLayout
	bottomLayout *qt.QVBoxLayout
	scrollLayout *qt.QVBoxLayout

	items           map[string]interface{}
	history         *common.Router
	currentRouteKey string
}

// NewNavigationBar builds a navigation bar.
func NewNavigationBar(parent *qt.QWidget) *NavigationBar {
	w := &NavigationBar{QWidget: qt.NewQWidget(parent)}
	w.indicator = NewNavigationIndicator(w.QWidget)
	w.isIndicatorAnimationEnabled = true
	w.isSelectedTextVisible = true
	w.lightSelectedColor = qt.NewQColor()
	w.darkSelectedColor = qt.NewQColor()

	w.scrollArea = widgets.NewScrollArea(w.QWidget)
	w.scrollWidget = qt.NewQWidget2()

	w.vBoxLayout = qt.NewQVBoxLayout(w.QWidget)
	w.topLayout = qt.NewQVBoxLayout2()
	w.bottomLayout = qt.NewQVBoxLayout2()
	w.scrollLayout = qt.NewQVBoxLayout(w.scrollWidget)

	w.items = map[string]interface{}{}
	w.history = common.RouterInstance

	w.initWidget()
	w.initLayout()
	w.indicator.OnAniFinished(func() { w.onIndicatorAniFinished() })
	return w
}

func (w *NavigationBar) initWidget() {
	w.Resize(48, w.Height())
	w.SetAttribute(qt.WA_StyledBackground)

	w.scrollArea.SetHorizontalScrollBarPolicy(qt.ScrollBarAlwaysOff)
	w.scrollArea.SetVerticalScrollBarPolicy(qt.ScrollBarAlwaysOff)
	w.scrollArea.HorizontalScrollBar().SetEnabled(false)
	w.scrollArea.SetWidget(w.scrollWidget)
	w.scrollArea.SetWidgetResizable(true)

	w.scrollWidget.SetObjectName("scrollWidget")
	common.FluentStyleSheet(common.FluentNavigationInterface).Apply(w.QWidget, common.ThemeAuto)
	common.FluentStyleSheet(common.FluentNavigationInterface).Apply(w.scrollWidget, common.ThemeAuto)
}

func (w *NavigationBar) initLayout() {
	w.vBoxLayout.SetContentsMargins(0, 5, 0, 5)
	w.topLayout.SetContentsMargins(4, 0, 4, 0)
	w.bottomLayout.SetContentsMargins(4, 0, 4, 0)
	w.scrollLayout.SetContentsMargins(4, 0, 4, 0)
	w.vBoxLayout.SetSpacing(4)
	w.topLayout.SetSpacing(4)
	w.bottomLayout.SetSpacing(4)
	w.scrollLayout.SetSpacing(4)

	w.vBoxLayout.AddLayout2(w.topLayout.QLayout, 0)
	w.vBoxLayout.AddWidget(w.scrollArea.QWidget)
	w.vBoxLayout.AddLayout2(w.bottomLayout.QLayout, 0)

	// Anchor the top/bottom groups (mirrors the Python setAlignment calls).
	w.vBoxLayout.SetAlignment2(w.topLayout.QLayout, qt.AlignTop)
	w.vBoxLayout.SetAlignment2(w.bottomLayout.QLayout, qt.AlignBottom)

	// miqt has no single-argument QLayout.setAlignment, so the whole-layout
	// `vBoxLayout.setAlignment(Qt.AlignTop)` / `scrollLayout.setAlignment(Qt.AlignTop)`
	// calls from the Python reference are applied through the embedded
	// QLayoutItem (the same trick used by window.FluentTitleBar). Without them
	// the leftover vertical space is spread between the top/scroll/bottom rows
	// instead of being pinned to the bottom, which stacks the buttons on top of
	// each other.
	w.vBoxLayout.QLayoutItem.SetAlignment(qt.AlignTop)
	w.scrollLayout.QLayoutItem.SetAlignment(qt.AlignTop)
}

// Widget returns the widget registered under routeKey (nil when missing).
func (w *NavigationBar) Widget(routeKey string) interface{} {
	return w.items[routeKey]
}

// AddItem appends a navigation bar item.
func (w *NavigationBar) AddItem(routeKey string, icon interface{}, text string, onClick func(bool), selectable bool, selectedIcon interface{}, position NavigationItemPosition) interface{} {
	return w.InsertItem(-1, routeKey, icon, text, onClick, selectable, selectedIcon, position)
}

// AddWidget appends a custom navigation widget.
func (w *NavigationBar) AddWidget(routeKey string, widget interface{}, onClick func(bool), position NavigationItemPosition) {
	w.InsertWidget(-1, routeKey, widget, onClick, position)
}

// InsertItem inserts a navigation bar item at index.
func (w *NavigationBar) InsertItem(index int, routeKey string, icon interface{}, text string, onClick func(bool), selectable bool, selectedIcon interface{}, position NavigationItemPosition) interface{} {
	if routeKey == "" {
		return nil
	}
	if _, ok := w.items[routeKey]; ok {
		return nil
	}
	item := NewNavigationBarPushButton(icon, text, selectable, selectedIcon, w.QWidget)
	item.SetSelectedColor(w.lightSelectedColor, w.darkSelectedColor)
	item.SetSelectedTextVisible(w.isSelectedTextVisible)
	w.InsertWidget(index, routeKey, item, onClick, position)
	return item
}

// InsertWidget inserts a custom navigation widget at index.
func (w *NavigationBar) InsertWidget(index int, routeKey string, widget interface{}, onClick func(bool), position NavigationItemPosition) {
	if routeKey == "" {
		return
	}
	if _, ok := w.items[routeKey]; ok {
		return
	}
	nw := navWidgetOf(widget)
	if nw == nil {
		return
	}
	w.registerWidget(routeKey, widget, nw, onClick)
	w.insertWidgetToLayout(index, widget, position)
}

func (w *NavigationBar) registerWidget(routeKey string, widget interface{}, nw *NavigationWidget, onClick func(bool)) {
	nw.routeKey = routeKey
	nw.OnClicked(func(v bool) { w.onWidgetClicked(nw, routeKey) })
	if onClick != nil {
		nw.OnClicked(onClick)
	}
	w.items[routeKey] = widget
}

func (w *NavigationBar) insertWidgetToLayout(index int, widget interface{}, position NavigationItemPosition) {
	nw := navWidgetOf(widget)
	if nw == nil {
		return
	}
	switch position {
	case NavigationItemPositionTop:
		nw.SetParent(w.QWidget)
		w.topLayout.InsertWidget3(index, nw.QWidget, 0, qt.AlignTop|qt.AlignHCenter)
	case NavigationItemPositionScroll:
		nw.SetParent(w.scrollWidget)
		w.scrollLayout.InsertWidget3(index, nw.QWidget, 0, qt.AlignTop|qt.AlignHCenter)
	default:
		nw.SetParent(w.QWidget)
		w.bottomLayout.InsertWidget3(index, nw.QWidget, 0, qt.AlignBottom|qt.AlignHCenter)
	}
	nw.Show()
}

// RemoveWidget removes a widget by route key.
func (w *NavigationBar) RemoveWidget(routeKey string) {
	widget, ok := w.items[routeKey]
	if !ok {
		return
	}
	delete(w.items, routeKey)
	if nw := navWidgetOf(widget); nw != nil {
		nw.DeleteLater()
	}
	w.history.Remove(routeKey)
}

func (w *NavigationBar) currentItem() *NavigationWidget {
	if w.currentRouteKey == "" {
		return nil
	}
	return navWidgetOf(w.items[w.currentRouteKey])
}

// SetCurrentItem selects the given route key, sliding the indicator from the
// previous item to the new item.
func (w *NavigationBar) SetCurrentItem(routeKey string) {
	if routeKey == "" {
		return
	}
	if _, ok := w.items[routeKey]; !ok || routeKey == w.currentRouteKey {
		return
	}

	w.stopIndicatorAnimation()
	prevItem := w.currentItem()
	w.currentRouteKey = routeKey

	if !w.isIndicatorAnimationEnabled || prevItem == nil {
		for k, widget := range w.items {
			if b, ok := widget.(*NavigationBarPushButton); ok {
				b.SetSelected(k == routeKey)
				continue
			}
			if nw := navWidgetOf(widget); nw != nil {
				nw.SetSelected(k == routeKey)
			}
		}
		return
	}

	newItem := w.currentItem()
	preRect := w.indicatorRectFor(prevItem)
	newRect := w.indicatorRectFor(newItem)

	prevItem.SetSelected(false)
	newItem.SetAboutSelected(true)
	w.indicator.Raise()
	w.indicator.SetIndicatorColor(newItem.lightIndicatorColor, newItem.darkIndicatorColor)
	w.indicator.StartAnimation(preRect, newRect, false)
	preRect.Delete()
	newRect.Delete()
}

// indicatorRectFor maps an item's indicator rect into bar coordinates.
func (w *NavigationBar) indicatorRectFor(item *NavigationWidget) *qt.QRectF {
	origin := qt.NewQPoint()
	mapped := item.MapTo(w.QWidget, origin) // GoGC-armed — do NOT Delete
	origin.Delete()

	rect := item.IndicatorRect()
	rect.Translate(float64(mapped.X()), float64(mapped.Y()))
	return rect
}

// SetFont sets the font on the bar and every bar button.
func (w *NavigationBar) SetFont(font *qt.QFont) {
	w.QWidget.SetFont(font)
	for _, b := range w.Buttons() {
		b.SetFont(font)
	}
}

// SetSelectedTextVisible toggles the selected text visibility on all buttons.
func (w *NavigationBar) SetSelectedTextVisible(isVisible bool) {
	if isVisible == w.isSelectedTextVisible {
		return
	}
	w.isSelectedTextVisible = isVisible
	for _, b := range w.Buttons() {
		b.SetSelectedTextVisible(isVisible)
	}
}

// IsSelectedTextVisible reports the selected text visibility flag.
func (w *NavigationBar) IsSelectedTextVisible() bool { return w.isSelectedTextVisible }

// SetSelectedColor sets the selected color of all bar buttons.
func (w *NavigationBar) SetSelectedColor(light, dark interface{}) {
	if w.lightSelectedColor != nil {
		w.lightSelectedColor.Delete()
	}
	if w.darkSelectedColor != nil {
		w.darkSelectedColor.Delete()
	}
	w.lightSelectedColor = coerceColor(light)
	w.darkSelectedColor = coerceColor(dark)
	for _, b := range w.Buttons() {
		b.SetSelectedColor(w.lightSelectedColor, w.darkSelectedColor)
	}
}

// Buttons returns the bar push buttons.
func (w *NavigationBar) Buttons() []*NavigationBarPushButton {
	var out []*NavigationBarPushButton
	for _, widget := range w.items {
		if b, ok := widget.(*NavigationBarPushButton); ok {
			out = append(out, b)
		}
	}
	return out
}

// IsIndicatorAnimationEnabled reports the indicator animation flag.
func (w *NavigationBar) IsIndicatorAnimationEnabled() bool { return w.isIndicatorAnimationEnabled }

// SetIndicatorAnimationEnabled sets the indicator animation flag.
func (w *NavigationBar) SetIndicatorAnimationEnabled(isEnabled bool) {
	w.isIndicatorAnimationEnabled = isEnabled
}

func (w *NavigationBar) onWidgetClicked(widget *NavigationWidget, routeKey string) {
	if widget.IsSelectable {
		w.SetCurrentItem(routeKey)
	}
}

func (w *NavigationBar) stopIndicatorAnimation() {
	if !w.isIndicatorAnimationEnabled {
		return
	}
	w.indicator.StopAnimation()
	w.onIndicatorAniFinished()
}

func (w *NavigationBar) onIndicatorAniFinished() {
	item := w.currentItem()
	if item == nil {
		return
	}
	item.SetAboutSelected(false)
	item.SetSelected(true)
	w.indicator.Hide()
}
