package navigation

import (
	"math"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	qt "github.com/mappu/miqt/qt"
)

// BreadcrumbWidget is the base clickable widget for breadcrumb items.
type BreadcrumbWidget struct {
	*qt.QWidget
	isHover   bool
	isPressed bool

	clickedSig voidSignal
}

func newBreadcrumbWidget(parent *qt.QWidget) *BreadcrumbWidget {
	w := &BreadcrumbWidget{QWidget: qt.NewQWidget(parent)}
	w.OnMousePressEvent(func(super func(event *qt.QMouseEvent), event *qt.QMouseEvent) {
		w.isPressed = true
		w.Update()
	})
	w.OnMouseReleaseEvent(func(super func(event *qt.QMouseEvent), event *qt.QMouseEvent) {
		w.isPressed = false
		w.Update()
		w.clickedSig.emit()
	})
	w.OnEnterEvent(func(super func(event *qt.QEvent), event *qt.QEvent) {
		w.isHover = true
		w.Update()
	})
	w.OnLeaveEvent(func(super func(event *qt.QEvent), event *qt.QEvent) {
		w.isHover = false
		w.Update()
	})
	return w
}

// OnClicked registers the clicked listener.
func (w *BreadcrumbWidget) OnClicked(f func()) { w.clickedSig.connect(f) }

// ElideButton is the "…" button shown when breadcrumb items overflow.
type ElideButton struct{ *BreadcrumbWidget }

// NewElideButton builds an elide button.
func NewElideButton(parent *qt.QWidget) *ElideButton {
	w := &ElideButton{BreadcrumbWidget: newBreadcrumbWidget(parent)}
	w.SetFixedSize2(16, 16)
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) { w.paint() })
	return w
}

// ClearState resets the hover state (the QHoverEvent sendEvent from the Python
// version is simplified to a direct attribute update).
func (w *ElideButton) ClearState() {
	w.SetAttribute2(qt.WA_UnderMouse, false)
	w.isHover = false
	w.Update()
}

func (w *ElideButton) paint() {
	painter := qt.NewQPainter2(w.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__Antialiasing)
	painter.SetPenWithStyle(qt.NoPen)

	if w.isPressed {
		painter.SetOpacity(0.5)
	} else if !w.isHover {
		painter.SetOpacity(0.61)
	}
	rect := qt.NewQRectF5(w.Rect())
	common.More.Render(painter, rect, common.ThemeAuto)
	rect.Delete()
	painter.End()
}

// BreadcrumbItem is a single breadcrumb label.
type BreadcrumbItem struct {
	*BreadcrumbWidget
	text       string
	routeKey   string
	index      int
	spacing    int
	isSelected bool
}

// NewBreadcrumbItem builds a breadcrumb item.
func NewBreadcrumbItem(routeKey, text string, index int, parent *qt.QWidget) *BreadcrumbItem {
	w := &BreadcrumbItem{
		BreadcrumbWidget: newBreadcrumbWidget(parent),
		text:             text,
		routeKey:         routeKey,
		index:            index,
		spacing:          5,
	}
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) { w.paint() })
	w.SetText(text)
	return w
}

// RouteKey returns the item route key.
func (w *BreadcrumbItem) RouteKey() string { return w.routeKey }

// Text returns the item text.
func (w *BreadcrumbItem) Text() string { return w.text }

// SetText sets the item text and recomputes its size.
func (w *BreadcrumbItem) SetText(text string) {
	w.text = text
	fm := w.FontMetrics()
	bound := fm.BoundingRectWithText(text)
	width := bound.Width() + int(math.Ceil(float64(w.Font().PixelSize())/10.0))
	if !w.IsRoot() {
		width += w.spacing * 2
	}
	w.SetFixedWidth(width)
	w.SetFixedHeight(bound.Height())
	w.Update()
}

// IsRoot reports whether this is the first (root) item.
func (w *BreadcrumbItem) IsRoot() bool { return w.index == 0 }

// SetSelected sets the selected state.
func (w *BreadcrumbItem) SetSelected(isSelected bool) {
	w.isSelected = isSelected
	w.Update()
}

// SetFont applies the font and recomputes the text size.
func (w *BreadcrumbItem) SetFont(font *qt.QFont) {
	w.QWidget.SetFont(font)
	w.SetText(w.text)
}

// SetSpacing sets the separator spacing and recomputes the text size.
func (w *BreadcrumbItem) SetSpacing(spacing int) {
	w.spacing = spacing
	w.SetText(w.text)
}

func (w *BreadcrumbItem) paint() {
	painter := qt.NewQPainter2(w.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__TextAntialiasing | qt.QPainter__Antialiasing)
	painter.SetPenWithStyle(qt.NoPen)

	sw := w.spacing * 2
	if !w.IsRoot() {
		iw := float64(w.Font().PixelSize()) / 14.0 * 8.0
		rect := qt.NewQRectF4((float64(sw)-iw)/2, (float64(w.Height())-iw)/2+1, iw, iw)
		painter.SetOpacity(0.61)
		common.ChevronRightMed.Render(painter, rect, common.ThemeAuto)
		rect.Delete()
	}

	if w.isPressed {
		alpha := 0.45
		if common.IsDarkTheme() {
			alpha = 0.54
		}
		if w.isSelected {
			painter.SetOpacity(1)
		} else {
			painter.SetOpacity(alpha)
		}
	} else if w.isSelected || w.isHover {
		painter.SetOpacity(1)
	} else {
		if common.IsDarkTheme() {
			painter.SetOpacity(0.79)
		} else {
			painter.SetOpacity(0.61)
		}
	}

	painter.SetFont(w.Font())
	textColor := qt.NewQColor3(255, 255, 255)
	if !common.IsDarkTheme() {
		textColor = qt.NewQColor3(0, 0, 0)
	}
	painter.SetPen(textColor)
	textColor.Delete()

	if w.IsRoot() {
		rect := w.Rect()
		painter.DrawText6(rect, int(qt.AlignVCenter|qt.AlignLeft), w.text)

	} else {
		rect := qt.NewQRectF4(float64(sw), 0, float64(w.Width()-sw), float64(w.Height()))
		painter.DrawText5(rect, int(qt.AlignVCenter|qt.AlignLeft), w.text)
		rect.Delete()
	}
	painter.End()
}

// BreadcrumbBar is a horizontal breadcrumb navigation bar.
type BreadcrumbBar struct {
	*qt.QWidget
	itemMap      map[string]*BreadcrumbItem
	items        []*BreadcrumbItem
	hiddenItems  []*BreadcrumbItem
	spacing      int
	currentIndex int
	elideButton  *ElideButton

	currentItemChangedSig  strSignal
	currentIndexChangedSig intSignal
}

// NewBreadcrumbBar builds a breadcrumb bar.
func NewBreadcrumbBar(parent *qt.QWidget) *BreadcrumbBar {
	w := &BreadcrumbBar{QWidget: qt.NewQWidget(parent)}
	w.itemMap = map[string]*BreadcrumbItem{}
	w.spacing = 10
	w.currentIndex = -1
	w.elideButton = NewElideButton(w.QWidget)
	common.SetFont(w.QWidget, 14, int(qt.QFont__Normal))
	w.SetAttribute(qt.WA_TranslucentBackground)
	w.elideButton.Hide()
	w.elideButton.OnClicked(func() { w.showHiddenItemsMenu() })

	// Re-flow the items (and elide overflow) whenever the bar is shown or resized.
	w.OnShowEvent(func(super func(event *qt.QShowEvent), event *qt.QShowEvent) {
		super(event)
		w.updateGeometry()
	})
	w.OnResizeEvent(func(super func(event *qt.QResizeEvent), event *qt.QResizeEvent) {
		super(event)
		w.updateGeometry()
	})
	return w
}

// OnCurrentItemChanged registers the currentItemChanged listener.
func (w *BreadcrumbBar) OnCurrentItemChanged(f func(string)) { w.currentItemChangedSig.connect(f) }

// OnCurrentIndexChanged registers the currentIndexChanged listener.
func (w *BreadcrumbBar) OnCurrentIndexChanged(f func(int)) { w.currentIndexChangedSig.connect(f) }

// AddItem appends a breadcrumb item.
func (w *BreadcrumbBar) AddItem(routeKey, text string) {
	if routeKey == "" {
		return
	}
	if _, ok := w.itemMap[routeKey]; ok {
		return
	}
	item := NewBreadcrumbItem(routeKey, text, len(w.items), w.QWidget)
	item.SetFont(w.Font())
	item.SetSpacing(w.spacing)
	item.OnClicked(func() { w.SetCurrentItem(routeKey) })

	w.itemMap[routeKey] = item
	w.items = append(w.items, item)

	maxH := 0
	for _, it := range w.items {
		if it.Height() > maxH {
			maxH = it.Height()
		}
	}
	w.SetFixedHeight(maxH)
	w.SetCurrentItem(routeKey)
	w.updateGeometry()
}

// SetCurrentIndex selects the item at index and removes trailing items.
func (w *BreadcrumbBar) SetCurrentIndex(index int) {
	if index < 0 || index >= len(w.items) || index == w.currentIndex {
		return
	}
	if w.currentIndex >= 0 && w.currentIndex < len(w.items) {
		w.items[w.currentIndex].SetSelected(false)
	}
	w.currentIndex = index
	w.items[index].SetSelected(true)

	for len(w.items) > index+1 {
		item := w.items[len(w.items)-1]
		w.items = w.items[:len(w.items)-1]
		delete(w.itemMap, item.routeKey)
		item.DeleteLater()
	}
	w.updateGeometry()
	w.currentIndexChangedSig.emit(index)
	w.currentItemChangedSig.emit(w.items[index].routeKey)
}

// SetCurrentItem selects the item with the given route key.
func (w *BreadcrumbBar) SetCurrentItem(routeKey string) {
	if item, ok := w.itemMap[routeKey]; ok {
		for i, it := range w.items {
			if it == item {
				w.SetCurrentIndex(i)
				return
			}
		}
	}
}

// SetItemText sets the text of an item.
func (w *BreadcrumbBar) SetItemText(routeKey, text string) {
	if item := w.itemMap[routeKey]; item != nil {
		item.SetText(text)
	}
}

// Item returns the item with the given route key.
func (w *BreadcrumbBar) Item(routeKey string) *BreadcrumbItem { return w.itemMap[routeKey] }

// ItemAt returns the item at index.
func (w *BreadcrumbBar) ItemAt(index int) *BreadcrumbItem {
	if index >= 0 && index < len(w.items) {
		return w.items[index]
	}
	return nil
}

// CurrentIndex returns the current index.
func (w *BreadcrumbBar) CurrentIndex() int { return w.currentIndex }

// CurrentItem returns the current item.
func (w *BreadcrumbBar) CurrentItem() *BreadcrumbItem {
	if w.currentIndex >= 0 && w.currentIndex < len(w.items) {
		return w.items[w.currentIndex]
	}
	return nil
}

// Clear removes all items.
func (w *BreadcrumbBar) Clear() {
	for len(w.items) > 0 {
		item := w.items[len(w.items)-1]
		w.items = w.items[:len(w.items)-1]
		delete(w.itemMap, item.routeKey)
		item.DeleteLater()
	}
	w.elideButton.Hide()
	w.currentIndex = -1
}

// PopItem removes the trailing item.
func (w *BreadcrumbBar) PopItem() {
	if len(w.items) == 0 {
		return
	}
	if w.Count() >= 2 {
		w.SetCurrentIndex(w.currentIndex - 1)
	} else {
		w.Clear()
	}
}

// Count returns the number of items.
func (w *BreadcrumbBar) Count() int { return len(w.items) }

func (w *BreadcrumbBar) updateGeometry() {
	if len(w.items) == 0 {
		return
	}
	w.elideButton.Hide()
	w.hiddenItems = append([]*BreadcrumbItem(nil), w.items[:len(w.items)-1]...)

	var visible []*qt.QWidget
	if !w.isElideVisible() {
		w.hiddenItems = nil
		visible = make([]*qt.QWidget, 0, len(w.items))
		for _, it := range w.items {
			visible = append(visible, it.QWidget)
		}
	} else {
		visible = []*qt.QWidget{w.elideButton.QWidget, w.items[len(w.items)-1].QWidget}
		total := 0
		for _, v := range visible {
			total += v.Width()
		}
		for i := len(w.items) - 2; i >= 0; i-- {
			item := w.items[i]
			total += item.Width()
			if total > w.Width() {
				break
			}
			visible = append(visible[:1], append([]*qt.QWidget{item.QWidget}, visible[1:]...)...)
			w.hiddenItems = removeHiddenItem(w.hiddenItems, item)
		}
	}

	for _, h := range w.hiddenItems {
		h.Hide()
	}
	pos := 0
	for _, v := range visible {
		v.Move(pos, (w.Height()-v.Height())/2)
		v.Show()
		pos += v.Width()
	}
}

func (w *BreadcrumbBar) isElideVisible() bool {
	total := 0
	for _, it := range w.items {
		total += it.Width()
	}
	return total > w.Width()
}

// SetFont sets the font and resizes the elide button and items.
func (w *BreadcrumbBar) SetFont(font *qt.QFont) {
	w.QWidget.SetFont(font)
	// Use float arithmetic so fonts smaller than 14px do not collapse the
	// elide button to 0 (mirrors Python's `int(pixelSize / 14 * 16)`).
	s := int(float64(font.PixelSize()) / 14.0 * 16.0)
	w.elideButton.SetFixedSize2(s, s)
	for _, item := range w.items {
		item.SetFont(font)
	}
}

// Spacing returns the item spacing.
func (w *BreadcrumbBar) Spacing() int { return w.spacing }

// SetSpacing sets the item spacing.
func (w *BreadcrumbBar) SetSpacing(spacing int) {
	if spacing == w.spacing {
		return
	}
	w.spacing = spacing
	for _, item := range w.items {
		item.SetSpacing(spacing)
	}
}

func (w *BreadcrumbBar) showHiddenItemsMenu() {
	w.elideButton.ClearState()
	menu := widgets.NewRoundMenu("", w.QWidget)
	menu.SetItemHeight(32)

	for _, item := range w.hiddenItems {
		route := item.routeKey
		action := menu.AddActionText(item.text)
		action.OnTriggered(func() { w.SetCurrentItem(route) })
	}

	// Position the menu below or above the bar depending on available space.
	pd := w.MapToGlobal(qt.NewQPoint2(0, w.Height()))
	pu := w.MapToGlobal(qt.NewQPoint2(0, 0))
	hd := menu.View().HeightForAnimation(pd, widgets.MenuAnimationDropDown)
	hu := menu.View().HeightForAnimation(pu, widgets.MenuAnimationPullUp)
	if hd >= hu {
		menu.View().AdjustSizeAt(pd, widgets.MenuAnimationDropDown)
		menu.Exec(pd, widgets.MenuAnimationDropDown)
	} else {
		menu.View().AdjustSizeAt(pu, widgets.MenuAnimationPullUp)
		menu.Exec(pu, widgets.MenuAnimationPullUp)
	}
}

func removeHiddenItem(list []*BreadcrumbItem, item *BreadcrumbItem) []*BreadcrumbItem {
	for i, it := range list {
		if it == item {
			return append(list[:i], list[i+1:]...)
		}
	}
	return list
}
