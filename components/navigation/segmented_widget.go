package navigation

import (
	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	qt "github.com/mappu/miqt/qt"
)

// SegmentedItem is a pivot item with a smaller (14px) font.
type SegmentedItem struct{ *PivotItem }

// NewSegmentedItem builds a segmented item.
func NewSegmentedItem(text string, parent *qt.QWidget) *SegmentedItem {
	w := &SegmentedItem{PivotItem: NewPivotItem(text, parent)}
	w.SetObjectName("segmentedItem")
	pivotItemStyleSheet{}.Apply(w.QWidget, common.ThemeAuto)
	common.SetFont(w.QWidget, 14, int(qt.QFont__Normal))
	return w
}

// SegmentedToolItem is an icon-only segmented item based on ToolButton.
type SegmentedToolItem struct {
	*widgets.ToolButton
	IsSelected bool

	itemClickedSig boolSignal
}

// NewSegmentedToolItem builds a segmented tool item.
func NewSegmentedToolItem(icon interface{}, parent *qt.QWidget) *SegmentedToolItem {
	w := &SegmentedToolItem{ToolButton: widgets.NewToolButtonIcon(icon, parent)}
	w.IsSelected = false
	w.SetProperty("isSelected", qt.NewQVariant11(false))
	w.SetObjectName("segmentedToolItem")
	w.SetFixedSize2(38, 33)
	pivotItemStyleSheet{}.Apply(w.QWidget, common.ThemeAuto)
	w.OnClicked(func() { w.itemClickedSig.emit(true) })
	return w
}

// OnItemClicked registers the itemClicked listener.
func (w *SegmentedToolItem) OnItemClicked(f func(bool)) { w.itemClickedSig.connect(f) }

// SetSelected sets the selected state.
func (w *SegmentedToolItem) SetSelected(isSelected bool) {
	if w.IsSelected == isSelected {
		return
	}
	w.IsSelected = isSelected
	w.SetProperty("isSelected", qt.NewQVariant11(isSelected))
	w.SetStyle(qt.QApplication_Style())
	w.Update()
}

// SegmentedToggleToolItem is a checkable segmented tool item.
type SegmentedToggleToolItem struct {
	*widgets.TransparentToolButton
	IsSelected bool

	itemClickedSig boolSignal
}

// NewSegmentedToggleToolItem builds a checkable segmented tool item.
func NewSegmentedToggleToolItem(icon interface{}, parent *qt.QWidget) *SegmentedToggleToolItem {
	w := &SegmentedToggleToolItem{TransparentToolButton: widgets.NewTransparentToolButtonIcon(icon, parent)}
	w.IsSelected = false
	w.SetFixedSize2(50, 32)
	w.OnClicked(func() { w.itemClickedSig.emit(true) })
	// When selected, the icon must reverse its theme so it renders white on the
	// accent-colored indicator capsule (the Go analogue of
	// SegmentedToggleToolItem._drawIcon in the reference, which picks
	// Theme.DARK in light mode / Theme.LIGHT in dark mode).
	w.SetDrawIconFunc(func(icon interface{}, painter *qt.QPainter, rect *qt.QRectF) {
		if w.IsSelected {
			if fi, ok := icon.(common.FluentIconBase); ok {
				theme := common.ThemeLight
				if !common.IsDarkTheme() {
					theme = common.ThemeDark
				}
				fi.Render(painter, rect, theme)
				return
			}
		}
		common.DrawIcon(icon, painter, rect)
	})
	return w
}

// OnItemClicked registers the itemClicked listener.
func (w *SegmentedToggleToolItem) OnItemClicked(f func(bool)) { w.itemClickedSig.connect(f) }

// SetSelected sets the selected/checked state.
func (w *SegmentedToggleToolItem) SetSelected(isSelected bool) {
	if w.IsSelected == isSelected {
		return
	}
	w.IsSelected = isSelected
	w.SetChecked(isSelected)
}

// SegmentedWidget is a pivot whose selected background and indicator slide
// horizontally. slideValue is animated point-to-point on selection changes.
type SegmentedWidget struct {
	*Pivot
	slideValue float64
	slideAni   *common.ProgressAnimation
}

// NewSegmentedWidget builds a segmented widget.
func NewSegmentedWidget(parent *qt.QWidget) *SegmentedWidget {
	w := &SegmentedWidget{Pivot: NewPivot(parent)}
	w.SetAttribute(qt.WA_StyledBackground)
	w.SetObjectName("segmentedWidget")
	pivotItemStyleSheet{}.Apply(w.QWidget, common.ThemeAuto)
	// Redirect the item-clicked path to the segmented SetCurrentItem (which
	// drives slideValue), since Go does not dispatch through the embedded Pivot.
	w.setCurrentItemFunc = w.SetCurrentItem

	// Re-register events so the segmented-specific indicator math runs.
	w.OnShowEvent(func(super func(event *qt.QShowEvent), event *qt.QShowEvent) {
		super(event)
		w.adjustIndicatorPos()
	})
	w.OnResizeEvent(func(super func(event *qt.QResizeEvent), event *qt.QResizeEvent) {
		super(event)
		w.adjustIndicatorPos()
	})
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) { w.paint(super, event) })
	return w
}

// InsertItem inserts a segmented item at index.
func (w *SegmentedWidget) InsertItem(index int, routeKey, text string, onClick func(bool), icon interface{}) interface{} {
	if routeKey == "" {
		return nil
	}
	if _, ok := w.items[routeKey]; ok {
		return nil
	}
	item := NewSegmentedItem(text, w.QWidget)
	if icon != nil {
		item.SetIcon(icon)
	}
	w.InsertWidget(index, routeKey, item, onClick)
	return item
}

// AddItem appends a segmented item. It must be declared on SegmentedWidget
// rather than relying on the promoted Pivot.AddItem: Go method dispatch is
// static, so the promoted Pivot.AddItem would call Pivot.InsertItem and build
// an 18px PivotItem instead of a 14px SegmentedItem. In the reference the
// virtual addItem → insertItem chain dispatches to SegmentedWidget.insertItem
// (segmented_widget.py), which this override reproduces.
func (w *SegmentedWidget) AddItem(routeKey, text string, onClick func(bool), icon interface{}) interface{} {
	return w.InsertItem(-1, routeKey, text, onClick, icon)
}

// SetCurrentItem selects the given route key and animates slideValue to it.
func (w *SegmentedWidget) SetCurrentItem(routeKey string) {
	if routeKey == "" {
		return
	}
	if _, ok := w.items[routeKey]; !ok || routeKey == w.currentRouteKey {
		return
	}
	w.adjustIndicatorPos() // stop + snap slideValue to the current (old) item
	w.currentRouteKey = routeKey
	w.startSlideToCurrentItem()
	for k, item := range w.items {
		if behavior, ok := item.(pivotItemBehavior); ok {
			behavior.SetSelected(k == routeKey)
		}
	}
	w.currentItemChangedSig.emit(routeKey)
}

func (w *SegmentedWidget) adjustIndicatorPos() {
	w.stopSlideAni()
	if item, ok := w.CurrentItem().(pivotGeometry); ok {
		w.slideValue = float64(item.X())
	}
}

func (w *SegmentedWidget) startSlideToCurrentItem() {
	w.stopSlideAni()

	var target float64
	if item, ok := w.CurrentItem().(pivotGeometry); ok {
		target = float64(item.X())
	} else {
		return
	}

	from := w.slideValue
	w.slideAni = common.NewFluentProgressAnimation(common.FluentAnimationTypePointToPoint, common.FluentAnimationSpeedFast)
	w.slideAni.OnProgress(func(t float64) {
		w.slideValue = from + (target-from)*t
		w.Update()
	})
	w.slideAni.OnFinished(func() {
		w.slideValue = target
		w.Update()
	})
	w.slideAni.Start()
}

func (w *SegmentedWidget) stopSlideAni() {
	if w.slideAni != nil {
		w.slideAni.Stop()
		w.slideAni.Delete()
		w.slideAni = nil
	}
}

// CurrentIndicatorGeometry returns the current item's x offset.
func (w *SegmentedWidget) CurrentIndicatorGeometry() float64 {
	if item, ok := w.CurrentItem().(pivotGeometry); ok {
		return float64(item.X())
	}
	return 0
}

func (w *SegmentedWidget) paint(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
	super(event)
	if w.CurrentItem() == nil {
		return
	}
	painter := qt.NewQPainter2(w.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__Antialiasing)

	var borderColor *qt.QColor
	var bg *qt.QColor
	if common.IsDarkTheme() {
		borderColor = qt.NewQColor11(255, 255, 255, 14)
		bg = qt.NewQColor11(255, 255, 255, 15)
	} else {
		borderColor = qt.NewQColor11(0, 0, 0, 19)
		bg = qt.NewQColor11(255, 255, 255, 179)
	}
	painter.SetPen(borderColor)
	borderColor.Delete()
	brush := qt.NewQBrush3(bg)
	bg.Delete()
	defer brush.Delete()
	painter.SetBrush(brush)

	if item, ok := w.CurrentItem().(pivotGeometry); ok {
		// The selected capsule is positioned by slideValue (the selected item's
		// x). Build an explicit rect so the item's own x offset is not
		// double-counted (the Python port translates item.rect(), which starts
		// at (0,0), by slideValue).
		rect := qt.NewQRect4(int(w.slideValue)+1, 1, item.Width()-2, item.Height()-2)
		painter.DrawRoundedRect3(rect, 5, 5)
		rect.Delete()
	}

	painter.SetPenWithStyle(qt.NoPen)
	color := fallbackThemeColor(w.lightIndicatorColor, w.darkIndicatorColor)
	brush = qt.NewQBrush3(color)
	color.Delete()
	defer brush.Delete()
	painter.SetBrush(brush)

	x := w.slideValue
	if item, ok := w.CurrentItem().(pivotGeometry); ok {
		x = float64(item.Width()/2-8) + w.slideValue
	}
	indicatorRect := qt.NewQRectF4(x, float64(w.Height())-3.5, 16, 3)
	painter.DrawRoundedRect(indicatorRect, 1.5, 1.5)
	indicatorRect.Delete()
	painter.End()
}

// SegmentedToolWidget is a segmented widget whose items are icon-only tool
// buttons.
type SegmentedToolWidget struct {
	*SegmentedWidget
	createItemFunc func(icon interface{}) interface{}
}

// NewSegmentedToolWidget builds a segmented tool widget.
func NewSegmentedToolWidget(parent *qt.QWidget) *SegmentedToolWidget {
	w := &SegmentedToolWidget{SegmentedWidget: NewSegmentedWidget(parent)}
	w.SetAttribute(qt.WA_StyledBackground)
	w.SetObjectName("segmentedToolWidget")
	pivotItemStyleSheet{}.Apply(w.QWidget, common.ThemeAuto)
	w.createItemFunc = func(icon interface{}) interface{} { return NewSegmentedToolItem(icon, w.QWidget) }
	return w
}

// AddItem appends an icon item.
func (w *SegmentedToolWidget) AddItem(routeKey string, icon interface{}, onClick func(bool)) interface{} {
	return w.InsertItem(-1, routeKey, icon, onClick)
}

// InsertItem inserts an icon item at index.
func (w *SegmentedToolWidget) InsertItem(index int, routeKey string, icon interface{}, onClick func(bool)) interface{} {
	if routeKey == "" {
		return nil
	}
	if _, ok := w.items[routeKey]; ok {
		return nil
	}
	item := w.createItemFunc(icon)
	w.InsertWidget(index, routeKey, item, onClick)
	return item
}

// SegmentedToggleToolWidget is a segmented tool widget with checkable items.
type SegmentedToggleToolWidget struct{ *SegmentedToolWidget }

// NewSegmentedToggleToolWidget builds a checkable segmented tool widget.
func NewSegmentedToggleToolWidget(parent *qt.QWidget) *SegmentedToggleToolWidget {
	w := &SegmentedToggleToolWidget{SegmentedToolWidget: NewSegmentedToolWidget(parent)}
	w.createItemFunc = func(icon interface{}) interface{} { return NewSegmentedToggleToolItem(icon, w.QWidget) }
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) { w.paint(super, event) })
	return w
}

func (w *SegmentedToggleToolWidget) paint(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
	super(event)
	if w.CurrentItem() == nil {
		return
	}
	painter := qt.NewQPainter2(w.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__Antialiasing)
	painter.SetPenWithStyle(qt.NoPen)
	color := fallbackThemeColor(w.lightIndicatorColor, w.darkIndicatorColor)
	brush := qt.NewQBrush3(color)
	color.Delete()
	defer brush.Delete()
	painter.SetBrush(brush)

	item, ok := w.CurrentItem().(pivotGeometry)
	if ok {
		rect := qt.NewQRectF4(w.slideValue, 0, float64(item.Width()), float64(item.Height()))
		painter.DrawRoundedRect(rect, 4, 4)
		rect.Delete()
	}
	painter.End()
}
