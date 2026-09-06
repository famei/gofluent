package navigation

import (
	"strings"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	qt "github.com/mappu/miqt/qt"
)

// pivotSelectorRewrites maps the PyQt-Fluent-Widgets class-name type selectors in
// pivot.qss to objectName selectors that match the native-widget Go port (the
// shared style_sheet.go selector table does not cover the pivot types). The
// order matters: longer names are rewritten first so "SegmentedToolItem" is not
// corrupted by the "SegmentedItem" rewrite.
//
// In the Python QSS, `SegmentedItem` is a QSS subclass of `PivotItem`, so the
// base `PivotItem { padding: 10px 12px; color/border/... }` rules also apply to
// segmented items. The objectName-based Go port has no such inheritance, so the
// `PivotItem` rules are duplicated for `#segmentedItem` as well. Compound
// selectors (`PivotItem[isSelected=...]`, `PivotItem[hasIcon=...]`) are
// rewritten before the bare `PivotItem` so the state/pseudo suffix stays
// attached to both object names.
var pivotSelectorRewrites = [][2]string{
	{"SegmentedToolWidget", "QWidget#segmentedToolWidget"},
	{"SegmentedToolItem", "QToolButton#segmentedToolItem"},
	{"SegmentedWidget", "QWidget#segmentedWidget"},
	{"SegmentedItem", "QPushButton#segmentedItem"},
	{"PivotItem[isSelected=true]:hover", "QPushButton#pivotItem[isSelected=true]:hover, QPushButton#segmentedItem[isSelected=true]:hover"},
	{"PivotItem[isSelected=true]:pressed", "QPushButton#pivotItem[isSelected=true]:pressed, QPushButton#segmentedItem[isSelected=true]:pressed"},
	{"PivotItem[isSelected=false]:pressed", "QPushButton#pivotItem[isSelected=false]:pressed, QPushButton#segmentedItem[isSelected=false]:pressed"},
	{"PivotItem[hasIcon=false]", "QPushButton#pivotItem[hasIcon=false], QPushButton#segmentedItem[hasIcon=false]"},
	{"PivotItem[hasIcon=true]", "QPushButton#pivotItem[hasIcon=true], QPushButton#segmentedItem[hasIcon=true]"},
	{"PivotItem", "QPushButton#pivotItem, QPushButton#segmentedItem"},
	{"Pivot", "QWidget#pivot"},
}

// pivotItemStyleSheet adapts the shared pivot QSS for the native-widget Go port:
// the Python class-name type selectors are rewritten to objectName selectors, so
// pivot items get their transparent background + pivot padding instead of the
// default button look (which would otherwise cover the sliding indicator).
type pivotItemStyleSheet struct{}

// Path delegates to the built-in pivot stylesheet.
func (pivotItemStyleSheet) Path(theme common.Theme) string {
	return common.FluentPivot.Path(theme)
}

// Content returns the pivot QSS with its class-name selectors rewritten.
func (pivotItemStyleSheet) Content(theme common.Theme) string {
	qss := common.FluentPivot.Content(theme)
	for _, r := range pivotSelectorRewrites {
		qss = strings.ReplaceAll(qss, r[0], r[1])
	}
	return qss
}

// Apply registers and applies the rewritten pivot stylesheet.
func (s pivotItemStyleSheet) Apply(widget *qt.QWidget, theme common.Theme) {
	common.SetStyleSheet(widget, s, theme)
}

// pivotItemBehavior is implemented by every pivot/segmented item type. It is
// the Go replacement for the Python duck-typed PivotItem hierarchy.
type pivotItemBehavior interface {
	SetSelected(bool)
	OnItemClicked(func(bool))
}

// pivotGeometry exposes the QWidget geometry methods needed to position the
// sliding indicator.
type pivotGeometry interface {
	Geometry() *qt.QRect
	X() int
	Width() int
	Height() int
}

// pivotFontItem is implemented by text pivot items (font sizing).
type pivotFontItem interface {
	Font() *qt.QFont
	SetFont(*qt.QFont)
	AdjustSize()
}

// pivotTextItem is implemented by text pivot items.
type pivotTextItem interface{ SetText(string) }

// pivotQWidget returns the embedded *qt.QWidget of any pivot item type.
func pivotQWidget(v interface{}) *qt.QWidget {
	switch w := v.(type) {
	case *PivotItem:
		return w.QWidget
	case *SegmentedItem:
		return w.QWidget
	case *SegmentedToolItem:
		return w.QWidget
	case *SegmentedToggleToolItem:
		return w.QWidget
	default:
		return nil
	}
}

// PivotItem is a selectable pivot (segmented) item based on PushButton.
type PivotItem struct {
	*widgets.PushButton
	IsSelected bool

	itemClickedSig boolSignal
}

// NewPivotItem builds a pivot item.
func NewPivotItem(text string, parent *qt.QWidget) *PivotItem {
	w := &PivotItem{PushButton: widgets.NewPushButtonText(text, parent)}
	w.IsSelected = false
	w.SetProperty("isSelected", qt.NewQVariant11(false))
	w.SetObjectName("pivotItem")
	w.OnClicked(func() { w.itemClickedSig.emit(true) })
	w.SetAttribute(qt.WA_LayoutUsesWidgetRect)
	pivotItemStyleSheet{}.Apply(w.QWidget, common.ThemeAuto)
	common.SetFont(w.QWidget, 18, int(qt.QFont__Normal))
	return w
}

// OnItemClicked registers the itemClicked listener.
func (w *PivotItem) OnItemClicked(f func(bool)) { w.itemClickedSig.connect(f) }

// SetSelected sets the selected state.
func (w *PivotItem) SetSelected(isSelected bool) {
	if w.IsSelected == isSelected {
		return
	}
	w.IsSelected = isSelected
	w.SetProperty("isSelected", qt.NewQVariant11(isSelected))
	w.SetStyle(qt.QApplication_Style())
	w.Update()
}

// Pivot is a horizontal pivoting navigation control with a sliding indicator.
// The indicator slide uses the WinUI squash-and-stretch transition.
type Pivot struct {
	*qt.QWidget
	items               map[string]interface{}
	currentRouteKey     string
	indicatorLength     int
	lightIndicatorColor *qt.QColor
	darkIndicatorColor  *qt.QColor
	hBoxLayout          *qt.QHBoxLayout
	indicatorRect       *qt.QRectF
	slide               *slideRectAnimator

	// setCurrentItemFunc lets subclasses (SegmentedWidget) redirect the
	// item-clicked path to their own SetCurrentItem, since Go method dispatch
	// is static (not virtual) through the embedded *Pivot.
	setCurrentItemFunc func(string)

	currentItemChangedSig strSignal
}

// NewPivot builds a pivot control.
func NewPivot(parent *qt.QWidget) *Pivot {
	w := &Pivot{QWidget: qt.NewQWidget(parent)}
	w.items = map[string]interface{}{}
	w.indicatorLength = 16
	w.lightIndicatorColor = qt.NewQColor()
	w.darkIndicatorColor = qt.NewQColor()
	w.hBoxLayout = qt.NewQHBoxLayout(w.QWidget)
	w.indicatorRect = qt.NewQRectF4(0, 0, 16, 3)

	common.FluentStyleSheet(common.FluentPivot).Apply(w.QWidget, common.ThemeAuto)
	w.hBoxLayout.SetSpacing(0)
	w.hBoxLayout.SetContentsMargins(0, 0, 0, 0)
	w.hBoxLayout.SetSizeConstraint(qt.QLayout__SetMinimumSize)
	w.SetSizePolicy2(qt.QSizePolicy__Minimum, qt.QSizePolicy__Minimum)

	w.OnShowEvent(func(super func(event *qt.QShowEvent), event *qt.QShowEvent) {
		super(event)
		w.adjustIndicatorPos()
	})
	w.OnResizeEvent(func(super func(event *qt.QResizeEvent), event *qt.QResizeEvent) {
		super(event)
		w.adjustIndicatorPos()
	})
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) { w.paint(super, event) })
	w.setCurrentItemFunc = w.SetCurrentItem
	return w
}

// OnCurrentItemChanged registers the currentItemChanged listener.
func (w *Pivot) OnCurrentItemChanged(f func(string)) { w.currentItemChangedSig.connect(f) }

// AddItem appends a pivot item.
func (w *Pivot) AddItem(routeKey, text string, onClick func(bool), icon interface{}) interface{} {
	return w.InsertItem(-1, routeKey, text, onClick, icon)
}

// AddWidget appends a custom pivot item.
func (w *Pivot) AddWidget(routeKey string, widget interface{}, onClick func(bool)) {
	w.InsertWidget(-1, routeKey, widget, onClick)
}

// InsertItem inserts a pivot item at index.
func (w *Pivot) InsertItem(index int, routeKey, text string, onClick func(bool), icon interface{}) interface{} {
	if routeKey == "" {
		return nil
	}
	if _, ok := w.items[routeKey]; ok {
		return nil
	}
	item := NewPivotItem(text, w.QWidget)
	if icon != nil {
		item.SetIcon(icon)
	}
	w.InsertWidget(index, routeKey, item, onClick)
	return item
}

// InsertWidget inserts a custom pivot item at index.
func (w *Pivot) InsertWidget(index int, routeKey string, widget interface{}, onClick func(bool)) {
	if routeKey == "" {
		return
	}
	if _, ok := w.items[routeKey]; ok {
		return
	}
	if behavior, ok := widget.(pivotItemBehavior); ok {
		behavior.OnItemClicked(func(v bool) { w.onItemClicked(widget, routeKey) })
		if onClick != nil {
			behavior.OnItemClicked(onClick)
		}
	}
	qw := pivotQWidget(widget)
	if qw == nil {
		return
	}
	w.items[routeKey] = widget
	// stretch=1 + no per-item alignment matches the reference
	// insertWidget(index, widget, 1). A zero alignment makes each item FILL its
	// cell so the segments share the available width equally (the reference's
	// layout-level setAlignment(Qt.AlignLeft) aligns the layout within its
	// parent, not the items). The previous AlignLeft here only aligned each
	// item within its cell, leaving content-sized items that differed in width;
	// that width jump between segments was what made the sliding capsule's
	// trailing edge appear to travel backwards for a few pixels.
	w.hBoxLayout.InsertWidget2(index, qw, 1)
}

// RemoveWidget removes a pivot item.
func (w *Pivot) RemoveWidget(routeKey string) {
	item, ok := w.items[routeKey]
	if !ok {
		return
	}
	delete(w.items, routeKey)
	if qw := pivotQWidget(item); qw != nil {
		w.hBoxLayout.RemoveWidget(qw)
		qw.DeleteLater()
	}
	common.RouterInstance.Remove(routeKey)
	if len(w.items) == 0 {
		w.currentRouteKey = ""
	}
}

// Clear removes every pivot item.
func (w *Pivot) Clear() {
	for k, item := range w.items {
		if qw := pivotQWidget(item); qw != nil {
			w.hBoxLayout.RemoveWidget(qw)
			qw.DeleteLater()
		}
		common.RouterInstance.Remove(k)
	}
	w.items = map[string]interface{}{}
	w.currentRouteKey = ""
}

// CurrentItem returns the current pivot item.
func (w *Pivot) CurrentItem() interface{} {
	if w.currentRouteKey == "" {
		return nil
	}
	return w.items[w.currentRouteKey]
}

// CurrentRouteKey returns the current route key.
func (w *Pivot) CurrentRouteKey() string { return w.currentRouteKey }

// SetCurrentItem selects the given route key and slides the indicator to it.
func (w *Pivot) SetCurrentItem(routeKey string) {
	if routeKey == "" {
		return
	}
	if _, ok := w.items[routeKey]; !ok || routeKey == w.currentRouteKey {
		return
	}
	w.adjustIndicatorPos() // stop + snap the indicator to the current (old) item
	w.currentRouteKey = routeKey
	w.startSlide(w.currentIndicatorGeometry())
	for k, item := range w.items {
		if behavior, ok := item.(pivotItemBehavior); ok {
			behavior.SetSelected(k == routeKey)
		}
	}
	w.currentItemChangedSig.emit(routeKey)
}

// SetIndicatorLength sets the indicator length.
func (w *Pivot) SetIndicatorLength(length int) {
	w.indicatorLength = length
	w.adjustIndicatorPos()
}

// IndicatorLength returns the indicator length.
func (w *Pivot) IndicatorLength() int { return w.indicatorLength }

// SetItemFontSize sets the font size of every text item.
func (w *Pivot) SetItemFontSize(size int) {
	for _, item := range w.items {
		if fi, ok := item.(pivotFontItem); ok {
			font := fi.Font() // borrowed reference — do NOT Delete
			font.SetPixelSize(size)
			fi.SetFont(font)
			fi.AdjustSize()
		}
	}
}

// SetItemText sets the text of an item.
func (w *Pivot) SetItemText(routeKey, text string) {
	if item, ok := w.items[routeKey].(pivotTextItem); ok {
		item.SetText(text)
	}
}

// SetIndicatorColor sets the light/dark indicator colors.
func (w *Pivot) SetIndicatorColor(light, dark interface{}) {
	if w.lightIndicatorColor != nil {
		w.lightIndicatorColor.Delete()
	}
	if w.darkIndicatorColor != nil {
		w.darkIndicatorColor.Delete()
	}
	w.lightIndicatorColor = coerceColor(light)
	w.darkIndicatorColor = coerceColor(dark)
	w.Update()
}

func (w *Pivot) onItemClicked(item interface{}, routeKey string) {
	if w.setCurrentItemFunc != nil {
		w.setCurrentItemFunc(routeKey)
		return
	}
	w.SetCurrentItem(routeKey)
}

func (w *Pivot) adjustIndicatorPos() {
	w.stopSlide()
	if w.CurrentItem() != nil {
		g := w.currentIndicatorGeometry()
		w.indicatorRect.SetRect(g.X(), g.Y(), g.Width(), g.Height())
		g.Delete()
	}
}

func (w *Pivot) startSlide(to *qt.QRectF) {
	w.slide = newSlideRectAnimator(w.indicatorRect, true, func() { w.Update() }, func() { w.slide = nil })
	w.slide.start(to, false)
	to.Delete()
}

func (w *Pivot) stopSlide() {
	if w.slide != nil {
		w.slide.stop()
		w.slide = nil
	}
}

func (w *Pivot) currentIndicatorGeometry() *qt.QRectF {
	item, ok := w.CurrentItem().(pivotGeometry)
	if !ok {
		return qt.NewQRectF4(0, float64(w.Height()-3), float64(w.indicatorLength), 3)
	}
	rect := item.Geometry() // borrowed (QWidget.Geometry const_cast) — do NOT Delete
	out := qt.NewQRectF4(float64(rect.X()-w.indicatorLength/2+rect.Width()/2), float64(w.Height()-3), float64(w.indicatorLength), 3)
	return out
}

func (w *Pivot) paint(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
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
	painter.DrawRoundedRect(w.indicatorRect, 1.5, 1.5)
	painter.End()
}
