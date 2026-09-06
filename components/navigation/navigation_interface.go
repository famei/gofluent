package navigation

import (
	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// NavigationInterface wraps a NavigationPanel and exposes a convenience API for
// adding items, separators, headers and a user card.
type NavigationInterface struct {
	*qt.QWidget
	panel *NavigationPanel

	displayModeChangedSig modeSignal
}

// NewNavigationInterface builds a navigation interface.
func NewNavigationInterface(parent *qt.QWidget, showMenuButton, showReturnButton, collapsible bool) *NavigationInterface {
	w := &NavigationInterface{QWidget: qt.NewQWidget(parent)}
	w.panel = NewNavigationPanel(w.QWidget, false)
	w.panel.SetMenuButtonVisible(showMenuButton && collapsible)
	w.panel.SetReturnButtonVisible(showReturnButton)
	w.panel.SetCollapsible(collapsible)
	w.panel.OnDisplayModeChanged(func(m NavigationDisplayMode) { w.displayModeChangedSig.emit(m) })

	w.Resize(48, w.Height())
	w.SetMinimumWidth(48)
	w.SetAttribute(qt.WA_TranslucentBackground)

	w.OnResizeEvent(func(super func(event *qt.QResizeEvent), event *qt.QResizeEvent) {
		super(event)
		// NOTE: QResizeEvent.Size()/OldSize() return references into the event
		// object (miqt const_cast), NOT copies — they must never be Delete()d.
		old := event.OldSize()
		size := event.Size()
		if old.Height() != w.Height() {
			w.panel.SetFixedHeight(w.Height())
		}
		// Sync the interface width to the panel in non-menu display modes
		// (approximates NavigationInterface.eventFilter without downcasting
		// QEvent to QResizeEvent).
		if w.panel.displayMode != NavigationDisplayModeMenu && old.Width() != size.Width() {
			w.SetFixedWidth(size.Width())
		}
	})
	return w
}

// OnDisplayModeChanged registers a displayModeChanged listener.
func (w *NavigationInterface) OnDisplayModeChanged(f func(NavigationDisplayMode)) {
	w.displayModeChangedSig.connect(f)
}

// AddItem appends a navigation item.
func (w *NavigationInterface) AddItem(routeKey string, icon interface{}, text string, onClick func(bool), selectable bool, position NavigationItemPosition, tooltip string, parentRouteKey string) interface{} {
	return w.InsertItem(-1, routeKey, icon, text, onClick, selectable, position, tooltip, parentRouteKey)
}

// AddWidget appends a custom navigation widget.
func (w *NavigationInterface) AddWidget(routeKey string, widget interface{}, onClick func(bool), position NavigationItemPosition, tooltip string, parentRouteKey string) {
	w.InsertWidget(-1, routeKey, widget, onClick, position, tooltip, parentRouteKey)
}

// InsertItem inserts a navigation item at index.
func (w *NavigationInterface) InsertItem(index int, routeKey string, icon interface{}, text string, onClick func(bool), selectable bool, position NavigationItemPosition, tooltip string, parentRouteKey string) interface{} {
	item := w.panel.InsertItem(index, routeKey, icon, text, onClick, selectable, position, tooltip, parentRouteKey)
	w.SetMinimumHeight(w.panel.LayoutMinHeight())
	return item
}

// InsertWidget inserts a custom navigation widget at index.
func (w *NavigationInterface) InsertWidget(index int, routeKey string, widget interface{}, onClick func(bool), position NavigationItemPosition, tooltip string, parentRouteKey string) {
	w.panel.InsertWidget(index, routeKey, widget, onClick, position, tooltip, parentRouteKey)
	w.SetMinimumHeight(w.panel.LayoutMinHeight())
}

// AddSeparator appends a separator.
func (w *NavigationInterface) AddSeparator(position NavigationItemPosition) {
	w.InsertSeparator(-1, position)
}

// AddItemHeader appends an item header.
func (w *NavigationInterface) AddItemHeader(text string, position NavigationItemPosition) *NavigationItemHeader {
	return w.panel.AddItemHeader(text, position)
}

// InsertItemHeader inserts an item header at index.
func (w *NavigationInterface) InsertItemHeader(index int, text string, position NavigationItemPosition) *NavigationItemHeader {
	return w.panel.InsertItemHeader(index, text, position)
}

// AddUserCard adds a user card to the panel.
func (w *NavigationInterface) AddUserCard(routeKey string, avatar interface{}, title, subtitle string, onClick func(bool), position NavigationItemPosition, aboveMenuButton bool) *NavigationUserCard {
	card := NewNavigationUserCard(w.QWidget)
	if avatar != nil {
		if _, ok := avatar.(common.FluentIconBase); ok {
			card.SetAvatarIcon(avatar)
		} else {
			card.SetAvatar(avatar)
		}
	}
	card.SetTitle(title)
	card.SetSubtitle(subtitle)

	index := -1
	if aboveMenuButton && position == NavigationItemPositionTop {
		for i := 0; i < w.panel.topLayout.Count(); i++ {
			item := w.panel.topLayout.ItemAt(i)
			if item != nil && item.Widget() != nil && item.Widget().UnsafePointer() == w.panel.menuButton.QWidget.UnsafePointer() {
				index = i
				break
			}
		}
	}
	if index >= 0 {
		w.panel.InsertWidget(index, routeKey, card, onClick, position, "", "")
	} else {
		w.AddWidget(routeKey, card, onClick, position, "", "")
	}
	return card
}

// InsertSeparator inserts a separator at index.
func (w *NavigationInterface) InsertSeparator(index int, position NavigationItemPosition) {
	w.panel.InsertSeparator(index, position)
	w.SetMinimumHeight(w.panel.LayoutMinHeight())
}

// RemoveWidget removes a widget by route key.
func (w *NavigationInterface) RemoveWidget(routeKey string) { w.panel.RemoveWidget(routeKey) }

// SetCurrentItem selects the given route key.
func (w *NavigationInterface) SetCurrentItem(routeKey string) { w.panel.SetCurrentItem(routeKey) }

// Expand expands the navigation panel.
func (w *NavigationInterface) Expand(useAni bool) { w.panel.Expand(useAni) }

// Toggle toggles the navigation panel.
func (w *NavigationInterface) Toggle() { w.panel.Toggle() }

// SetExpandWidth sets the expanded width.
func (w *NavigationInterface) SetExpandWidth(width int) { w.panel.SetExpandWidth(width) }

// SetMinimumExpandWidth sets the minimum window width for expansion.
func (w *NavigationInterface) SetMinimumExpandWidth(width int) { w.panel.SetMinimumExpandWidth(width) }

// SetMenuButtonVisible toggles the menu button.
func (w *NavigationInterface) SetMenuButtonVisible(isVisible bool) {
	w.panel.SetMenuButtonVisible(isVisible)
}

// SetReturnButtonVisible toggles the return button.
func (w *NavigationInterface) SetReturnButtonVisible(isVisible bool) {
	w.panel.SetReturnButtonVisible(isVisible)
}

// SetCollapsible toggles collapsibility.
func (w *NavigationInterface) SetCollapsible(collapsible bool) { w.panel.SetCollapsible(collapsible) }

// IsAcrylicEnabled reports the acrylic flag.
func (w *NavigationInterface) IsAcrylicEnabled() bool { return w.panel.IsAcrylicEnabled() }

// SetAcrylicEnabled sets the acrylic flag (no-op, see MIGRATION_GUIDE §10).
func (w *NavigationInterface) SetAcrylicEnabled(isEnabled bool) { w.panel.SetAcrylicEnabled(isEnabled) }

// IsIndicatorAnimationEnabled reports the indicator animation flag.
func (w *NavigationInterface) IsIndicatorAnimationEnabled() bool {
	return w.panel.IsIndicatorAnimationEnabled()
}

// SetIndicatorAnimationEnabled sets the indicator animation flag.
func (w *NavigationInterface) SetIndicatorAnimationEnabled(isEnabled bool) {
	w.panel.SetIndicatorAnimationEnabled(isEnabled)
}

// IsUpdateIndicatorPosOnCollapseFinished reports the collapse update flag.
func (w *NavigationInterface) IsUpdateIndicatorPosOnCollapseFinished() bool {
	return w.panel.IsUpdateIndicatorPosOnCollapseFinished()
}

// SetUpdateIndicatorPosOnCollapseFinished sets the collapse update flag.
func (w *NavigationInterface) SetUpdateIndicatorPosOnCollapseFinished(update bool) {
	w.panel.SetUpdateIndicatorPosOnCollapseFinished(update)
}

// Widget returns the widget registered under routeKey.
func (w *NavigationInterface) Widget(routeKey string) interface{} { return w.panel.Widget(routeKey) }
