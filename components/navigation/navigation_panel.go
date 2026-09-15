package navigation

import (
	"fmt"
	"strings"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	qt "github.com/mappu/miqt/qt"
)

// NavigationDisplayMode enumerates how the navigation panel is presented.
type NavigationDisplayMode int

const (
	NavigationDisplayModeMinimal NavigationDisplayMode = iota
	NavigationDisplayModeCompact
	NavigationDisplayModeExpand
	NavigationDisplayModeMenu
)

// NavigationItemPosition enumerates where a navigation item is inserted.
type NavigationItemPosition int

const (
	NavigationItemPositionTop NavigationItemPosition = iota
	NavigationItemPositionScroll
	NavigationItemPositionBottom
)

// RouteKeyError is the error reported for an unknown route key. Go returns nil
// from the widget accessors instead of raising, but the type is preserved for
// API parity.
type RouteKeyError struct{ Key string }

// Error implements the error interface.
func (e *RouteKeyError) Error() string { return fmt.Sprintf("`%s` is illegal.", e.Key) }

// NavigationItem records a registered route key and its widget.
type NavigationItem struct {
	RouteKey       string
	ParentRouteKey string
	Widget         interface{}
}

// navigationPanelStyleSheet adapts the shared navigation_interface QSS for the
// plain-QFrame Go port: the `NavigationPanel` type selector (which only matches
// a C++ subclass) is rewritten to `QFrame#NavigationPanel`, so the widget's
// object name must be "NavigationPanel".
type navigationPanelStyleSheet struct{}

// Path delegates to the built-in navigation interface stylesheet.
func (navigationPanelStyleSheet) Path(theme common.Theme) string {
	return common.FluentNavigationInterface.Path(theme)
}

// Content returns the navigation interface QSS with the selector rewritten.
func (navigationPanelStyleSheet) Content(theme common.Theme) string {
	return strings.ReplaceAll(
		common.FluentNavigationInterface.Content(theme),
		"NavigationPanel",
		"QFrame#NavigationPanel",
	)
}

// Apply registers and applies the adapted stylesheet.
func (s navigationPanelStyleSheet) Apply(widget *qt.QWidget, theme common.Theme) {
	common.SetStyleSheet(widget, s, theme)
}

// applyNavigationPanelQSS applies the adapted navigation interface stylesheet.
func applyNavigationPanelQSS(widget *qt.QWidget, theme common.Theme) {
	navigationPanelStyleSheet{}.Apply(widget, theme)
}

// NavigationPanel is the collapsible navigation surface (QFrame based).
type NavigationPanel struct {
	*qt.QFrame

	parentWidget                           *qt.QWidget
	isMenuButtonVisible                    bool
	isReturnButtonVisible                  bool
	isCollapsible                          bool
	isAcrylicEnabled                       bool
	isIndicatorAnimationEnabled            bool
	isUpdateIndicatorPosOnCollapseFinished bool

	indicator    *NavigationIndicator
	scrollArea   *widgets.ScrollArea
	scrollWidget *qt.QWidget

	menuButton   *NavigationToolButton
	returnButton *NavigationToolButton

	vBoxLayout   *qt.QVBoxLayout
	topLayout    *NavigationItemLayout
	bottomLayout *NavigationItemLayout
	scrollLayout *NavigationItemLayout

	items           map[string]*NavigationItem
	history         *common.Router
	currentRouteKey string

	expandWidth        int
	minimumExpandWidth int
	isMinimalEnabled   bool
	displayMode        NavigationDisplayMode
	expanding          bool

	expandAni *common.ProgressAnimation

	separators []*NavigationSeparator
	allWidgets []interface{}

	displayModeChangedSig modeSignal
}

// NewNavigationPanel builds a navigation panel.
func NewNavigationPanel(parent *qt.QWidget, isMinimalEnabled bool) *NavigationPanel {
	p := &NavigationPanel{QFrame: qt.NewQFrame(parent)}
	p.parentWidget = parent
	p.isMenuButtonVisible = true
	p.isCollapsible = true
	p.isIndicatorAnimationEnabled = true

	p.indicator = NewNavigationIndicator(p.QWidget)
	p.scrollArea = widgets.NewScrollArea(p.QWidget)
	p.scrollWidget = qt.NewQWidget2()

	p.menuButton = NewNavigationToolButton(common.GlobalNavButton, p.QWidget)
	p.returnButton = NewNavigationToolButton(common.Back, p.QWidget)

	p.vBoxLayout = qt.NewQVBoxLayout(p.QWidget)
	p.topLayout = newNavigationItemLayout(nil)
	p.bottomLayout = newNavigationItemLayout(nil)
	p.scrollLayout = newNavigationItemLayout(p.scrollWidget)

	p.items = map[string]*NavigationItem{}
	p.history = common.RouterInstance

	p.expandWidth = 322
	p.minimumExpandWidth = 1008
	p.isMinimalEnabled = isMinimalEnabled
	if isMinimalEnabled {
		p.displayMode = NavigationDisplayModeMinimal
	} else {
		p.displayMode = NavigationDisplayModeCompact
	}

	p.initWidget()
	p.initLayout()
	return p
}

// OnDisplayModeChanged registers a displayModeChanged listener.
func (p *NavigationPanel) OnDisplayModeChanged(f func(NavigationDisplayMode)) {
	p.displayModeChangedSig.connect(f)
}

func (p *NavigationPanel) initWidget() {
	p.Resize(48, p.Height())
	p.SetAttribute(qt.WA_StyledBackground)

	p.returnButton.Hide()
	p.returnButton.SetDisabled(true)

	p.scrollArea.SetHorizontalScrollBarPolicy(qt.ScrollBarAlwaysOff)
	p.scrollArea.HorizontalScrollBar().SetEnabled(false)
	p.scrollArea.SetWidget(p.scrollWidget)
	p.scrollArea.SetWidgetResizable(true)
	// The vertical fluent scroll bar handle only appears on hover (matches the
	// reference `scrollDelagate.vScrollBar.setHandleDisplayMode(ON_HOVER)`).
	p.scrollArea.ScrollDelegate().VerticalSmoothScrollBar().SetHandleDisplayMode(widgets.ScrollBarHandleDisplayOnHover)

	p.menuButton.OnClicked(func(v bool) { p.Toggle() })
	p.history.OnEmptyChanged(func(empty bool) { p.returnButton.SetDisabled(empty) })
	p.returnButton.OnClicked(func(v bool) { p.history.Pop() })
	p.indicator.OnAniFinished(func() { p.onIndicatorAniFinished() })

	_ = widgets.NewToolTipFilter(p.returnButton.QWidget, 1000, widgets.ToolTipPositionTop)
	p.returnButton.SetToolTip("Back")
	_ = widgets.NewToolTipFilter(p.menuButton.QWidget, 1000, widgets.ToolTipPositionTop)
	p.menuButton.SetToolTip("Open Navigation")

	p.SetProperty("menu", qt.NewQVariant11(false))
	p.scrollWidget.SetObjectName("scrollWidget")

	// The shared QSS uses a `NavigationPanel` type selector, which only matches a
	// C++ QFrame subclass named NavigationPanel. The Go port uses a plain QFrame,
	// so the selector is rewritten to `QFrame#NavigationPanel` and the object name
	// is set accordingly (fixes the transparent/missing panel background).
	p.SetObjectName("NavigationPanel")
	applyNavigationPanelQSS(p.QWidget, common.ThemeAuto)
	common.FluentStyleSheet(common.FluentNavigationInterface).Apply(p.scrollWidget, common.ThemeAuto)
}

func (p *NavigationPanel) initLayout() {
	p.vBoxLayout.SetContentsMargins(0, 5, 0, 5)
	p.topLayout.SetContentsMargins(4, 0, 4, 0)
	p.bottomLayout.SetContentsMargins(4, 0, 4, 0)
	p.scrollLayout.SetContentsMargins(4, 0, 4, 0)
	p.vBoxLayout.SetSpacing(4)
	p.topLayout.SetSpacing(4)
	p.bottomLayout.SetSpacing(4)
	p.scrollLayout.SetSpacing(4)

	// miqt's QLayout does not expose the whole-layout setAlignment(Qt.AlignTop)
	// that the Python reference uses (scrollLayout.setAlignment(Qt.AlignTop)) to
	// keep navigation items packed at the top. Add a trailing stretch that
	// absorbs the extra vertical space instead of distributing it between items
	// (which made the gap between buttons grow when the window was expanded).
	p.scrollLayout.AddStretch()

	p.vBoxLayout.AddLayout2(p.topLayout.QLayout, 0)
	p.vBoxLayout.AddWidget2(p.scrollArea.QWidget, 1)
	p.vBoxLayout.AddLayout2(p.bottomLayout.QLayout, 0)

	// Keep the top/bottom item groups anchored (mirrors the Python
	// setAlignment calls; QLayout has no whole-layout alignment binding).
	p.vBoxLayout.SetAlignment2(p.topLayout.QLayout, qt.AlignTop)
	p.vBoxLayout.SetAlignment2(p.bottomLayout.QLayout, qt.AlignBottom)

	// The Python reference also calls vBoxLayout.setAlignment(Qt.AlignTop),
	// which keeps the whole panel content flush against the top edge. miqt has
	// no single-argument setAlignment, so apply it through the embedded
	// QLayoutItem (same trick as window.FluentTitleBar).
	p.vBoxLayout.QLayoutItem.SetAlignment(qt.AlignTop)

	p.topLayout.AddWidget3(p.returnButton.QWidget, 0, qt.AlignTop)
	p.topLayout.AddWidget3(p.menuButton.QWidget, 0, qt.AlignTop)

	// Keep the sliding indicator above the menu/return buttons and the scroll
	// area so the selection animation is never painted over (z-order fix).
	p.indicator.Raise()
}

// IsIndicatorAnimationEnabled reports whether the indicator slide is enabled.
func (p *NavigationPanel) IsIndicatorAnimationEnabled() bool { return p.isIndicatorAnimationEnabled }

// SetIndicatorAnimationEnabled toggles the (degraded) indicator animation.
func (p *NavigationPanel) SetIndicatorAnimationEnabled(isEnabled bool) {
	p.isIndicatorAnimationEnabled = isEnabled
}

// IsUpdateIndicatorPosOnCollapseFinished reports the collapse update flag.
func (p *NavigationPanel) IsUpdateIndicatorPosOnCollapseFinished() bool {
	return p.isUpdateIndicatorPosOnCollapseFinished
}

// SetUpdateIndicatorPosOnCollapseFinished sets the collapse update flag.
func (p *NavigationPanel) SetUpdateIndicatorPosOnCollapseFinished(update bool) {
	p.isUpdateIndicatorPosOnCollapseFinished = update
}

// Widget returns the widget registered under routeKey (nil when missing).
func (p *NavigationPanel) Widget(routeKey string) interface{} {
	if item, ok := p.items[routeKey]; ok {
		return item.Widget
	}
	return nil
}

// AddItem appends a navigation tree item.
func (p *NavigationPanel) AddItem(routeKey string, icon interface{}, text string, onClick func(bool), selectable bool, position NavigationItemPosition, tooltip string, parentRouteKey string) interface{} {
	return p.InsertItem(-1, routeKey, icon, text, onClick, selectable, position, tooltip, parentRouteKey)
}

// AddWidget appends a custom navigation widget.
func (p *NavigationPanel) AddWidget(routeKey string, widget interface{}, onClick func(bool), position NavigationItemPosition, tooltip string, parentRouteKey string) {
	p.InsertWidget(-1, routeKey, widget, onClick, position, tooltip, parentRouteKey)
}

// InsertItem inserts a navigation tree item at index.
func (p *NavigationPanel) InsertItem(index int, routeKey string, icon interface{}, text string, onClick func(bool), selectable bool, position NavigationItemPosition, tooltip string, parentRouteKey string) interface{} {
	if routeKey == "" {
		return nil
	}
	if _, ok := p.items[routeKey]; ok {
		return nil
	}
	w := NewNavigationTreeWidget(icon, text, selectable, p.QWidget)
	p.InsertWidget(index, routeKey, w, onClick, position, tooltip, parentRouteKey)
	return w
}

// InsertWidget inserts a custom navigation widget at index.
func (p *NavigationPanel) InsertWidget(index int, routeKey string, widget interface{}, onClick func(bool), position NavigationItemPosition, tooltip string, parentRouteKey string) {
	if routeKey == "" {
		return
	}
	if _, ok := p.items[routeKey]; ok {
		return
	}
	nw := navWidgetOf(widget)
	if nw == nil {
		return
	}
	p.registerWidget(routeKey, parentRouteKey, widget, nw, onClick, tooltip)
	if parentRouteKey != "" {
		if parent, ok := p.items[parentRouteKey]; ok {
			if tw, ok := parent.Widget.(*NavigationTreeWidget); ok {
				if c, ok := widget.(*NavigationTreeWidget); ok {
					tw.InsertChild(index, c)
					return
				}
			}
		}
	}
	p.insertWidgetToLayout(index, widget, position)
}

// AddSeparator appends a separator.
func (p *NavigationPanel) AddSeparator(position NavigationItemPosition) {
	p.InsertSeparator(-1, position)
}

// InsertSeparator inserts a separator at index.
func (p *NavigationPanel) InsertSeparator(index int, position NavigationItemPosition) {
	separator := NewNavigationSeparator(p.QWidget)
	p.separators = append(p.separators, separator)
	p.allWidgets = append(p.allWidgets, separator)
	p.insertWidgetToLayout(index, separator, position)
}

// AddItemHeader appends an item header.
func (p *NavigationPanel) AddItemHeader(text string, position NavigationItemPosition) *NavigationItemHeader {
	return p.InsertItemHeader(-1, text, position)
}

// InsertItemHeader inserts an item header at index.
func (p *NavigationPanel) InsertItemHeader(index int, text string, position NavigationItemPosition) *NavigationItemHeader {
	header := NewNavigationItemHeader(text, p.QWidget)
	p.allWidgets = append(p.allWidgets, header)
	p.insertWidgetToLayout(index, header, position)

	isCompacted := p.displayMode != NavigationDisplayModeExpand && p.displayMode != NavigationDisplayModeMenu
	header.SetCompacted(isCompacted)
	return header
}

func (p *NavigationPanel) registerWidget(routeKey, parentRouteKey string, widget interface{}, nw *NavigationWidget, onClick func(bool), tooltip string) {
	nw.routeKey = routeKey
	nw.parentRouteKey = parentRouteKey
	nw.OnClicked(func(v bool) { p.onWidgetClicked(widget) })
	if onClick != nil {
		nw.OnClicked(onClick)
	}
	p.items[routeKey] = &NavigationItem{RouteKey: routeKey, ParentRouteKey: parentRouteKey, Widget: widget}
	p.allWidgets = append(p.allWidgets, widget)

	if p.displayMode == NavigationDisplayModeExpand || p.displayMode == NavigationDisplayModeMenu {
		setNavigationCompacted(widget, false)
	}
	if tooltip != "" {
		nw.SetToolTip(tooltip)
		_ = widgets.NewToolTipFilter(nw.QWidget, 1000, widgets.ToolTipPositionTop)
	}
}

func (p *NavigationPanel) insertWidgetToLayout(index int, widget interface{}, position NavigationItemPosition) {
	nw := navWidgetOf(widget)
	if nw == nil {
		return
	}
	switch position {
	case NavigationItemPositionTop:
		nw.SetParent(p.QWidget)
		p.topLayout.InsertWidget3(index, nw.QWidget, 0, qt.AlignTop)
	case NavigationItemPositionScroll:
		nw.SetParent(p.scrollWidget)
		idx := index
		// The scroll layout ends with a trailing stretch (see initLayout);
		// insert before it so navigation items stay packed at the top.
		if idx < 0 || idx >= p.scrollLayout.Count() {
			idx = p.scrollLayout.Count() - 1
		}
		p.scrollLayout.InsertWidget3(idx, nw.QWidget, 0, qt.AlignTop)
	default:
		nw.SetParent(p.QWidget)
		p.bottomLayout.InsertWidget3(index, nw.QWidget, 0, qt.AlignBottom)
	}
	nw.Show()

	// Newly added items stack above the indicator by default; raise it back so
	// the selection slide stays visible over every navigation item.
	p.indicator.Raise()
}

// RemoveWidget removes a widget by route key.
func (p *NavigationPanel) RemoveWidget(routeKey string) {
	item, ok := p.items[routeKey]
	if !ok {
		return
	}
	if p.currentRouteKey == routeKey {
		p.currentRouteKey = ""
	}
	delete(p.items, routeKey)

	if item.ParentRouteKey != "" {
		if parent, ok := p.items[item.ParentRouteKey]; ok {
			if tw, ok := parent.Widget.(*NavigationTreeWidget); ok {
				if c, ok := item.Widget.(*NavigationTreeWidget); ok {
					tw.RemoveChild(c)
				}
			}
		}
	}

	if tw, ok := item.Widget.(*NavigationTreeWidget); ok {
		for _, child := range tw.treeChildren {
			p.removeTreeChild(child)
		}
	}

	if nw := navWidgetOf(item.Widget); nw != nil {
		nw.DeleteLater()
	}
	p.history.Remove(routeKey)
}

func (p *NavigationPanel) removeTreeChild(child *NavigationTreeWidget) {
	if child.routeKey != "" {
		delete(p.items, child.routeKey)
		p.history.Remove(child.routeKey)
	}
	for _, c := range child.treeChildren {
		p.removeTreeChild(c)
	}
	child.DeleteLater()
}

// SetMenuButtonVisible toggles the menu button visibility.
func (p *NavigationPanel) SetMenuButtonVisible(isVisible bool) {
	p.isMenuButtonVisible = isVisible
	p.menuButton.SetVisible(isVisible)
}

// SetReturnButtonVisible toggles the return button visibility.
func (p *NavigationPanel) SetReturnButtonVisible(isVisible bool) {
	p.isReturnButtonVisible = isVisible
	p.returnButton.SetVisible(isVisible)
}

// SetCollapsible toggles whether the panel can collapse.
func (p *NavigationPanel) SetCollapsible(on bool) {
	p.isCollapsible = on
	if !on && p.displayMode != NavigationDisplayModeExpand {
		p.Expand(false)
	}
}

// SetExpandWidth sets the expanded width.
func (p *NavigationPanel) SetExpandWidth(width int) {
	if width <= 42 {
		return
	}
	p.expandWidth = width
	navigationWidgetExpandWidth = width - 10
}

// SetMinimumExpandWidth sets the minimum window width that allows expansion.
func (p *NavigationPanel) SetMinimumExpandWidth(width int) { p.minimumExpandWidth = width }

// SetAcrylicEnabled is retained for API compatibility but is a no-op: the
// acrylic/Mica effect was removed from the port.
func (p *NavigationPanel) SetAcrylicEnabled(isEnabled bool) {
	p.isAcrylicEnabled = isEnabled
}

// IsAcrylicEnabled reports the acrylic flag.
func (p *NavigationPanel) IsAcrylicEnabled() bool { return p.isAcrylicEnabled }

// Expand expands the navigation panel, animating its width (150ms OutQuad) when
// useAni is true.
func (p *NavigationPanel) Expand(useAni bool) {
	p.stopIndicatorAnimation()
	p.setWidgetCompacted(false)
	p.restoreTreeExpandState(useAni)
	p.expanding = true
	p.menuButton.SetToolTip("Close Navigation")

	expandWidth := p.minimumExpandWidth + p.expandWidth - 322
	if (p.Window().Width() >= expandWidth && !p.isMinimalEnabled) || !p.isCollapsible {
		p.displayMode = NavigationDisplayModeExpand
	} else {
		p.SetProperty("menu", qt.NewQVariant11(true))
		p.SetStyle(qt.QApplication_Style())
		p.displayMode = NavigationDisplayModeMenu
		if !p.parentWidget.IsWindow() {
			pos := p.ParentWidget().Pos()
			p.SetParent(p.Window())
			p.MoveWithQPoint(pos)

		}
		p.Show()
	}

	p.displayModeChangedSig.emit(p.displayMode)
	if useAni {
		p.animateWidth(p.expandWidth)
	} else {
		p.Resize(p.expandWidth, p.Height())
		if p.displayMode == NavigationDisplayModeExpand {
			p.syncParentWidth(p.expandWidth)
		}
		p.onExpandAniFinished()
	}
}

// Collapse collapses the navigation panel, animating its width back to 48px.
func (p *NavigationPanel) Collapse() {
	if p.currentRouteKey != "" {
		if item, ok := p.items[p.currentRouteKey]; ok && item.ParentRouteKey != "" {
			p.stopIndicatorAnimation()
		}
	}

	for _, item := range p.items {
		if tw, ok := item.Widget.(*NavigationTreeWidget); ok && tw.IsRoot() {
			tw.SaveExpandState()
			tw.SetExpanded(false, false)
		}
	}

	p.expanding = false
	p.menuButton.SetToolTip("Open Navigation")
	p.animateWidth(48)
}

// animateWidth animates the panel width (and the parent interface width, except
// in menu mode where the panel is a floating overlay). On completion it invokes
// onExpandAniFinished.
func (p *NavigationPanel) animateWidth(targetWidth int) {
	p.stopExpandAni()

	from := p.Width()
	syncParent := p.displayMode != NavigationDisplayModeMenu
	curve := qt.NewQEasingCurve3(qt.QEasingCurve__OutQuad)
	p.expandAni = common.NewProgressAnimation(150, curve)
	curve.Delete()

	p.expandAni.OnProgress(func(t float64) {
		w := from + int(float64(targetWidth-from)*t+0.5)
		p.Resize(w, p.Height())
		if syncParent {
			p.syncParentWidth(w)
		}
	})
	p.expandAni.OnFinished(func() {
		p.Resize(targetWidth, p.Height())
		if syncParent {
			p.syncParentWidth(targetWidth)
		}
		p.onExpandAniFinished()
	})
	p.expandAni.Start()
}

func (p *NavigationPanel) stopExpandAni() {
	if p.expandAni != nil {
		p.expandAni.Stop()
		p.expandAni.Delete()
		p.expandAni = nil
	}
}

func (p *NavigationPanel) stopIndicatorAnimation() {
	p.indicator.StopAnimation()
	p.onIndicatorAniFinished()
}

func (p *NavigationPanel) restoreTreeExpandState(useAni bool) {
	for _, item := range p.items {
		if tw, ok := item.Widget.(*NavigationTreeWidget); ok && tw.IsRoot() {
			tw.RestoreExpandState(useAni)
		}
	}
}

// Toggle expands or collapses the panel.
func (p *NavigationPanel) Toggle() {
	if p.displayMode == NavigationDisplayModeCompact || p.displayMode == NavigationDisplayModeMinimal {
		p.Expand(true)
	} else {
		p.Collapse()
	}
}

// SetCurrentItem selects the given route key, sliding the indicator from the
// previous item to the new item (squash-and-stretch).
func (p *NavigationPanel) SetCurrentItem(routeKey string) {
	if routeKey == "" {
		return
	}
	if _, ok := p.items[routeKey]; !ok || routeKey == p.currentRouteKey {
		return
	}

	prevItem := p.currentItem()
	p.currentRouteKey = routeKey
	newItem := p.currentItem()
	newIndicatorItem := p.findIndicatorItem(newItem)
	prevIndicatorItem := p.findIndicatorItem(prevItem)

	if !(p.isIndicatorAnimationEnabled && prevItem != nil && prevIndicatorItem != nil && newIndicatorItem != nil) {
		for k, item := range p.items {
			if nw := navWidgetOf(item.Widget); nw != nil {
				nw.SetSelected(k == routeKey)
			}
		}
		return
	}

	preRect := p.indicatorRectFor(prevIndicatorItem)
	newRect := p.indicatorRectFor(newIndicatorItem)

	prevItem.SetSelected(false)
	prevIndicatorItem.SetSelected(false)
	newIndicatorItem.SetAboutSelected(true)

	p.indicator.SetIndicatorColor(newItem.lightIndicatorColor, newItem.darkIndicatorColor)
	p.indicator.StartAnimation(preRect, newRect, false)
	preRect.Delete()
	newRect.Delete()
}

func (p *NavigationPanel) currentItem() *NavigationWidget {
	if p.currentRouteKey == "" {
		return nil
	}
	if item, ok := p.items[p.currentRouteKey]; ok {
		return navWidgetOf(item.Widget)
	}
	return nil
}

// findIndicatorItem walks up the tree chain to the nearest visible navigation
// widget, so the indicator rests on a collapsed root while its child is hidden.
func (p *NavigationPanel) findIndicatorItem(item *NavigationWidget) *NavigationWidget {
	cur := item
	for cur != nil && !cur.IsVisible() && cur.TreeParent != nil {
		cur = cur.TreeParent.NavigationWidget
	}
	return cur
}

// indicatorRectFor maps an item's indicator rect into panel coordinates.
func (p *NavigationPanel) indicatorRectFor(item *NavigationWidget) *qt.QRectF {
	origin := qt.NewQPoint()
	mapped := item.MapTo(p.QWidget, origin) // GoGC-armed — do NOT Delete
	origin.Delete()

	rect := item.IndicatorRect()
	rect.Translate(float64(mapped.X()), float64(mapped.Y()))
	return rect
}

func (p *NavigationPanel) onIndicatorAniFinished() {
	item := p.currentItem()
	if item == nil {
		return
	}
	item.SetSelected(true)
	if ii := p.findIndicatorItem(item); ii != nil {
		ii.SetAboutSelected(false)
	}
	p.indicator.Hide()
}

func (p *NavigationPanel) onWidgetClicked(widget interface{}) {
	nw := navWidgetOf(widget)
	if nw == nil {
		return
	}
	if !nw.IsSelectable {
		p.showFlyoutNavigationMenu(widget)
		return
	}
	p.SetCurrentItem(nw.routeKey)

	isLeaf := true
	if tw, ok := widget.(*NavigationTreeWidget); ok {
		isLeaf = tw.IsLeaf()
	}
	if p.displayMode == NavigationDisplayModeMenu && isLeaf {
		p.Collapse()
	} else if p.isCollapsed() {
		p.showFlyoutNavigationMenu(widget)
	}
}

func (p *NavigationPanel) showFlyoutNavigationMenu(widget interface{}) {
	if !p.isCollapsed() {
		return
	}
	tw, ok := widget.(*NavigationTreeWidget)
	if !ok {
		return
	}
	if !tw.IsRoot() || tw.IsLeaf() {
		return
	}

	view := widgets.NewFlyoutViewBase(nil)
	layout := qt.NewQHBoxLayout(view.QWidget)
	view.SetLayout(layout.QLayout)
	menu := NewNavigationFlyoutMenu(tw, view.QWidget)
	layout.SetContentsMargins(0, 0, 0, 0)
	layout.AddWidget(menu.QWidget)

	// The flyout slide + menu resize animation are simplified to a native
	// right-aligned popup (MIGRATION_GUIDE §10/§12).
	_ = widgets.FlyoutMake(view, tw.QWidget, p.Window(), widgets.FlyoutAnimationSlideRight, true)
}

// IsCollapsed reports whether the panel is in compact mode.
func (p *NavigationPanel) IsCollapsed() bool { return p.displayMode == NavigationDisplayModeCompact }

func (p *NavigationPanel) isCollapsed() bool { return p.IsCollapsed() }

func (p *NavigationPanel) onExpandAniFinished() {
	if !p.expanding {
		if p.isMinimalEnabled {
			p.displayMode = NavigationDisplayModeMinimal
		} else {
			p.displayMode = NavigationDisplayModeCompact
		}
		p.displayModeChangedSig.emit(p.displayMode)
	}

	switch p.displayMode {
	case NavigationDisplayModeMinimal:
		p.Hide()
		p.SetProperty("menu", qt.NewQVariant11(false))
		p.SetStyle(qt.QApplication_Style())
		p.syncParentWidth(48)
	case NavigationDisplayModeCompact:
		p.SetProperty("menu", qt.NewQVariant11(false))
		p.SetStyle(qt.QApplication_Style())
		p.setWidgetCompacted(true)
		if p.isUpdateIndicatorPosOnCollapseFinished {
			p.stopIndicatorAnimation()
		}
		if !p.parentWidget.IsWindow() {
			p.SetParent(p.parentWidget)
			p.Move(0, 0)
			p.Show()
		}
		p.syncParentWidth(48)
	}
}

func (p *NavigationPanel) syncParentWidth(width int) {
	if p.parentWidget != nil && !p.parentWidget.IsWindow() {
		p.parentWidget.SetFixedWidth(width)
	}
}

func (p *NavigationPanel) setWidgetCompacted(isCompacted bool) {
	for _, w := range p.allWidgets {
		setNavigationCompacted(w, isCompacted)
	}
}

// LayoutMinHeight returns the minimum height needed to lay out the panel.
func (p *NavigationPanel) LayoutMinHeight() int {
	th := p.topLayout.MinimumSize()    // GoGC-armed — do NOT Delete
	bh := p.bottomLayout.MinimumSize() // GoGC-armed — do NOT Delete
	sh := 0
	for _, s := range p.separators {
		sh += s.Height()
	}
	spacing := p.topLayout.Count()*p.topLayout.Spacing() + p.bottomLayout.Count()*p.bottomLayout.Spacing()
	return 36 + th.Height() + bh.Height() + sh + spacing
}

// ---------------------------------------------------------------------------
// NavigationItemLayout

// NavigationItemLayout re-anchors separator widgets to x=0 when the layout
// geometry changes (mirroring NavigationItemLayout.setGeometry).
type NavigationItemLayout struct{ *qt.QVBoxLayout }

func newNavigationItemLayout(parent *qt.QWidget) *NavigationItemLayout {
	var layout *NavigationItemLayout
	if parent == nil {
		layout = &NavigationItemLayout{QVBoxLayout: qt.NewQVBoxLayout2()}
	} else {
		layout = &NavigationItemLayout{QVBoxLayout: qt.NewQVBoxLayout(parent)}
	}
	layout.OnSetGeometry(func(super func(geometry *qt.QRect), geometry *qt.QRect) {
		super(geometry)
		for i := 0; i < layout.Count(); i++ {
			item := layout.ItemAt(i)
			if item == nil {
				continue
			}
			w := item.Widget()
			if w != nil && navigationSeparatorPtrs[w.UnsafePointer()] {
				geo := item.Geometry()
				w.SetGeometry(0, geo.Y(), geo.Width(), geo.Height())
			}
		}
	})
	return layout
}
