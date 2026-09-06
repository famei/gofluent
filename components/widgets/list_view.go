package widgets

import (
	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// ListItemDelegate is the fluent item delegate for list views. It reuses the
// table delegate structure with the list-specific background / indicator layout.
type ListItemDelegate struct {
	*TableItemDelegate
}

// NewListItemDelegate builds a list item delegate.
func NewListItemDelegate(view *qt.QAbstractItemView) *ListItemDelegate {
	d := NewTableItemDelegate(view)
	d.listStyle = true
	return &ListItemDelegate{TableItemDelegate: d}
}

func setupListView(view *qt.QListView, delegate *ListItemDelegate) {
	common.FluentStyleSheet(common.FluentListView).Apply(view.QWidget, common.ThemeAuto)
	view.SetItemDelegate(delegate.QAbstractItemDelegate)
	view.SetMouseTracking(true)
	installItemTracking(view.QAbstractItemView, delegate.TableItemDelegate)
}

// ListWidget is a fluent styled list widget.
type ListWidget struct {
	*qt.QListWidget
	delegate                *ListItemDelegate
	scrollDelegate          *SmoothScrollDelegate
	isSelectRightClickedRow bool
}

// NewListWidget builds a list widget.
func NewListWidget(parent *qt.QWidget) *ListWidget {
	w := &ListWidget{QListWidget: qt.NewQListWidget(parent)}
	w.delegate = NewListItemDelegate(w.QAbstractItemView)
	setupListView(w.QListView, w.delegate)
	w.scrollDelegate = NewSmoothScrollDelegate(w.QAbstractItemView.QAbstractScrollArea, false)
	return w
}

// SetCurrentItem sets the current item.
func (w *ListWidget) SetCurrentItem(item *qt.QListWidgetItem) {
	w.QListWidget.SetCurrentItem(item)
}

// SetCurrentRow sets the current row.
func (w *ListWidget) SetCurrentRow(row int) {
	w.QListWidget.SetCurrentRow(row)
}

// SetCheckedColor sets the checked indicator color.
func (w *ListWidget) SetCheckedColor(light, dark *qt.QColor) {
	w.delegate.SetCheckedColor(light, dark)
}

// IsSelectRightClickedRow reports whether right-click selects the row.
func (w *ListWidget) IsSelectRightClickedRow() bool { return w.isSelectRightClickedRow }

// SetSelectRightClickedRow toggles right-click row selection.
func (w *ListWidget) SetSelectRightClickedRow(isSelect bool) { w.isSelectRightClickedRow = isSelect }

// ListView is a fluent styled list view.
type ListView struct {
	*qt.QListView
	delegate                *ListItemDelegate
	scrollDelegate          *SmoothScrollDelegate
	isSelectRightClickedRow bool
}

// NewListView builds a list view.
func NewListView(parent *qt.QWidget) *ListView {
	w := &ListView{QListView: qt.NewQListView(parent)}
	w.delegate = NewListItemDelegate(w.QAbstractItemView)
	setupListView(w.QListView, w.delegate)
	w.scrollDelegate = NewSmoothScrollDelegate(w.QAbstractItemView.QAbstractScrollArea, false)
	return w
}

// SetCheckedColor sets the checked indicator color.
func (w *ListView) SetCheckedColor(light, dark *qt.QColor) {
	w.delegate.SetCheckedColor(light, dark)
}

// IsSelectRightClickedRow reports whether right-click selects the row.
func (w *ListView) IsSelectRightClickedRow() bool { return w.isSelectRightClickedRow }

// SetSelectRightClickedRow toggles right-click row selection.
func (w *ListView) SetSelectRightClickedRow(isSelect bool) { w.isSelectRightClickedRow = isSelect }
