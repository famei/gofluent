package widgets

import (
	"fmt"
	"math"

	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// TableItemDelegate is the fluent item delegate for table views. It paints the
// rounded row highlight and the left accent indicator exactly like the Python
// TableItemDelegate (hover / pressed / selected / alternate states), then lets
// the native delegate draw the item text and decoration on top.
type TableItemDelegate struct {
	*qt.QStyledItemDelegate
	view              *qt.QAbstractItemView
	lightCheckedColor *qt.QColor
	darkCheckedColor  *qt.QColor
	hoverRow          int
	pressedRow        int
	selectedRows      map[int]struct{}
	margin            int
	// listStyle switches to the ListItemDelegate layout (single-column rounded
	// background and a left-edge indicator at x=0).
	listStyle bool
}

// NewTableItemDelegate builds a table item delegate.
func NewTableItemDelegate(view *qt.QAbstractItemView) *TableItemDelegate {
	d := &TableItemDelegate{
		QStyledItemDelegate: qt.NewQStyledItemDelegate2(view.QObject),
		view:                view,
		lightCheckedColor:   qt.NewQColor(),
		darkCheckedColor:    qt.NewQColor(),
		hoverRow:            -1,
		pressedRow:          -1,
		selectedRows:        map[int]struct{}{},
		margin:              2,
	}
	d.installPaint()
	d.installEditor()
	return d
}

// SetHoverRow sets the hovered row.
func (d *TableItemDelegate) SetHoverRow(row int) { d.hoverRow = row }

// SetPressedRow sets the pressed row.
func (d *TableItemDelegate) SetPressedRow(row int) { d.pressedRow = row }

// SetSelectedRows records the currently selected rows and clears the pressed row
// when the pressed row has just become selected (mirrors Python setSelectedRows).
func (d *TableItemDelegate) SetSelectedRows(indexes []qt.QModelIndex) {
	d.selectedRows = make(map[int]struct{}, len(indexes))
	for i := range indexes {
		row := indexes[i].Row()
		d.selectedRows[row] = struct{}{}
		if row == d.pressedRow {
			d.pressedRow = -1
		}
	}
}

// SetCheckedColor sets the indicator color in checked status.
func (d *TableItemDelegate) SetCheckedColor(light, dark *qt.QColor) {
	d.lightCheckedColor = cloneColor(light)
	d.darkCheckedColor = cloneColor(dark)
	if d.view != nil {
		if vp := d.view.Viewport(); vp != nil {
			vp.Update()
		}
	}
}

// LightCheckedColor returns the light checked color.
func (d *TableItemDelegate) LightCheckedColor() *qt.QColor { return d.lightCheckedColor }

// DarkCheckedColor returns the dark checked color.
func (d *TableItemDelegate) DarkCheckedColor() *qt.QColor { return d.darkCheckedColor }

// installPaint overrides the delegate paint virtual so that the Fluent hover /
// pressed / selected highlight is drawn underneath the native text rendering.
func (d *TableItemDelegate) installPaint() {
	d.OnPaint(func(super func(painter *qt.QPainter, option *qt.QStyleOptionViewItem, index *qt.QModelIndex), painter *qt.QPainter, option *qt.QStyleOptionViewItem, index *qt.QModelIndex) {
		painter.Save()
		painter.SetPenWithStyle(qt.NoPen)
		painter.SetRenderHints(qt.QPainter__Antialiasing)

		orig := option.Rect() // GoGC-armed value copy — do NOT Delete
		painter.SetClipping(true)
		painter.SetClipRectWithQRect(orig)

		adj := orig.Adjusted(0, d.margin, 0, -d.margin) // GoGC-armed — do NOT Delete
		option.SetRect(*adj)

		d.paintHighlight(painter, option, index)

		painter.Restore()
		super(painter, option, index)
	})

	d.OnInitStyleOption(func(super func(option *qt.QStyleOptionViewItem, index *qt.QModelIndex), option *qt.QStyleOptionViewItem, index *qt.QModelIndex) {
		super(option, index)
		initItemViewOptionText(option)
	})

	d.OnSizeHint(func(super func(option *qt.QStyleOptionViewItem, index *qt.QModelIndex) *qt.QSize, option *qt.QStyleOptionViewItem, index *qt.QModelIndex) *qt.QSize {
		size := super(option, index)
		size.SetHeight(size.Height() + 2*d.margin)
		return size
	})
}

// installEditor overrides createEditor / updateEditorGeometry so that double
// clicking a cell opens a Fluent LineEdit with a clear button instead of the
// native plain QLineEdit (the Go port of TableItemDelegate.createEditor).
func (d *TableItemDelegate) installEditor() {
	d.OnCreateEditor(func(super func(parent *qt.QWidget, option *qt.QStyleOptionViewItem, index *qt.QModelIndex) *qt.QWidget, parent *qt.QWidget, option *qt.QStyleOptionViewItem, index *qt.QModelIndex) *qt.QWidget {
		lineEdit := NewLineEdit(parent)
		lineEdit.SetProperty("transparent", qt.NewQVariant11(false))
		lineEdit.SetStyle(qt.QApplication_Style())
		lineEdit.SetText(option.Text())
		lineEdit.SetClearButtonEnabled(true)
		return lineEdit.QWidget
	})

	d.OnUpdateEditorGeometry(func(super func(editor *qt.QWidget, option *qt.QStyleOptionViewItem, index *qt.QModelIndex), editor *qt.QWidget, option *qt.QStyleOptionViewItem, index *qt.QModelIndex) {
		rect := option.Rect() // GoGC-armed — do NOT Delete
		y := rect.Y() + (rect.Height()-editor.Height())/2
		x := rect.X()
		if x < 8 {
			x = 8
		}
		w := rect.Width()
		if index.Column() == 0 {
			w -= 8
		}
		editor.SetGeometry(x, y, w, rect.Height())
	})
}

// paintHighlight computes the Fluent alpha and draws the row background plus the
// selection indicator and (when the item carries a check state) the check box.
func (d *TableItemDelegate) paintHighlight(painter *qt.QPainter, option *qt.QStyleOptionViewItem, index *qt.QModelIndex) {
	row := index.Row()
	isHover := d.hoverRow == row
	isPressed := d.pressedRow == row
	_, isSelected := d.selectedRows[row]
	isAlternate := row%2 == 0 && d.view.AlternatingRowColors()

	c := 255
	if !common.IsDarkTheme() {
		c = 0
	}
	alpha := 0

	if !isSelected {
		switch {
		case isPressed:
			alpha = 6
			if common.IsDarkTheme() {
				alpha = 9
			}
		case isHover:
			alpha = 12
		case isAlternate:
			alpha = 5
		}
	} else {
		switch {
		case isPressed:
			alpha = 9
			if common.IsDarkTheme() {
				alpha = 15
			}
		case isHover:
			alpha = 25
		default:
			alpha = 17
		}
	}

	color := qt.NewQColor11(c, c, c, alpha)
	brush := qt.NewQBrush3(color)
	painter.SetBrush(brush)
	brush.Delete()
	color.Delete()

	d.drawBackground(painter, option, index)

	if isSelected && index.Column() == 0 && d.view.HorizontalScrollBar().Value() == 0 {
		d.drawIndicator(painter, option, index)
	}

	if _, ok := checkStateOf(index); ok {
		d.drawCheckBox(painter, option, index)
	}
}

// drawCheckBox draws the Fluent check box of an item that carries a
// CheckStateRole value (the Go port of TableItemDelegate._drawCheckBox).
func (d *TableItemDelegate) drawCheckBox(painter *qt.QPainter, option *qt.QStyleOptionViewItem, index *qt.QModelIndex) {
	state, ok := checkStateOf(index)
	if !ok {
		return
	}

	painter.Save()
	defer painter.Restore()

	r := 4.5
	rect := option.Rect() // GoGC-armed — do NOT Delete
	rf := qt.NewQRectF4(float64(rect.X())+15, float64(rect.Y()+rect.Height()/2)-9.5, 19, 19)
	defer rf.Delete()

	if state == qt.Unchecked {
		var bg, border *qt.QColor
		if common.IsDarkTheme() {
			bg = qt.NewQColor11(0, 0, 0, 26)
			border = qt.NewQColor11(255, 255, 255, 142)
		} else {
			bg = qt.NewQColor11(0, 0, 0, 6)
			border = qt.NewQColor11(0, 0, 0, 122)
		}
		defer bg.Delete()
		defer border.Delete()

		brush := qt.NewQBrush3(bg)
		painter.SetBrush(brush)
		brush.Delete()
		painter.SetPen(border)
		painter.DrawRoundedRect(rf, r, r)
		return
	}

	indColor := common.AutoFallbackThemeColor(d.lightCheckedColor, d.darkCheckedColor)
	brush := qt.NewQBrush3(indColor)
	painter.SetBrush(brush)
	brush.Delete()
	painter.SetPen(indColor)
	painter.DrawRoundedRect(rf, r, r)

	if state == qt.Checked {
		CheckBoxIconAccept.render(painter, rf)
	} else {
		CheckBoxIconPartialAccept.render(painter, rf)
	}
}

// checkStateOf returns the item's CheckStateRole value, reporting false when the
// item has no check state at all (an invalid QVariant).
func checkStateOf(index *qt.QModelIndex) (qt.CheckState, bool) {
	v := index.DataWithRole(int(qt.CheckStateRole))
	if v == nil || v.IsNull() {
		return qt.Unchecked, false
	}
	return qt.CheckState(v.ToInt()), true
}

// drawBackground draws the rounded row background, taking the column position
// into account so adjacent cells join into a single rounded strip.
func (d *TableItemDelegate) drawBackground(painter *qt.QPainter, option *qt.QStyleOptionViewItem, index *qt.QModelIndex) {
	rect := option.Rect() // GoGC-armed — do NOT Delete

	if d.listStyle {
		painter.DrawRoundedRect3(rect, 5, 5)
		return
	}

	r := 5
	switch index.Column() {
	case 0:
		adjusted := rect.Adjusted(4, 0, r+1, 0)
		painter.DrawRoundedRect3(adjusted, float64(r), float64(r))
	case d.lastColumn(index):
		adjusted := rect.Adjusted(-r-1, 0, -4, 0)
		painter.DrawRoundedRect3(adjusted, float64(r), float64(r))
	default:
		adjusted := rect.Adjusted(-1, 0, 1, 0)
		painter.DrawRectWithRect(adjusted)
	}
}

// drawIndicator draws the 3px accent bar at the left of a selected row.
func (d *TableItemDelegate) drawIndicator(painter *qt.QPainter, option *qt.QStyleOptionViewItem, index *qt.QModelIndex) {
	rect := option.Rect() // GoGC-armed — do NOT Delete
	y := rect.Y()
	h := rect.Height()

	ph := math.Round(0.257 * float64(h))
	if d.pressedRow == index.Row() {
		ph = math.Round(0.35 * float64(h))
	}

	indColor := common.AutoFallbackThemeColor(d.lightCheckedColor, d.darkCheckedColor)
	indBrush := qt.NewQBrush3(indColor)
	painter.SetBrush(indBrush)
	indBrush.Delete()

	x := 4.0
	if d.listStyle {
		x = 0.0
	}
	rf := qt.NewQRectF4(x, ph+float64(y), 3, float64(h)-2*ph)
	defer rf.Delete()
	painter.DrawRoundedRect(rf, 1.5, 1.5)
}

// initItemViewOptionText sets the item font and theme text color, mirroring the
// Python initStyleOption implementation (white text in dark theme, black in
// light theme).
func initItemViewOptionText(option *qt.QStyleOptionViewItem) {
	font := common.GetFont(13, int(qt.QFont__Normal))
	option.SetFont(*font)
	font.Delete()

	c := 255
	if !common.IsDarkTheme() {
		c = 0
	}
	textColor := qt.NewQColor11(c, c, c, 255)
	palette := option.Palette() // GoGC-armed value copy — do NOT Delete
	palette.SetColor2(qt.QPalette__Text, textColor)
	palette.SetColor2(qt.QPalette__HighlightedText, textColor)
	option.SetPalette(*palette)
	textColor.Delete()
}

// lastColumn returns the index of the last column of the row.
func (d *TableItemDelegate) lastColumn(index *qt.QModelIndex) int {
	if m := index.Model(); m != nil {
		return m.ColumnCount(index.Parent()) - 1
	}
	return 0
}

// installItemTracking wires the hover / pressed / selected row tracking signals
// so the delegate can paint the transient Fluent states.
func installItemTracking(view *qt.QAbstractItemView, delegate *TableItemDelegate) {
	view.OnEntered(func(index *qt.QModelIndex) {
		delegate.SetHoverRow(index.Row())
		if vp := view.Viewport(); vp != nil {
			vp.Update()
		}
	})
	view.OnPressed(func(index *qt.QModelIndex) {
		if view.SelectionMode() == qt.QAbstractItemView__NoSelection {
			return
		}
		delegate.SetPressedRow(index.Row())
		if vp := view.Viewport(); vp != nil {
			vp.Update()
		}
	})
	if sm := view.SelectionModel(); sm != nil {
		sm.OnSelectionChanged(func(selected *qt.QItemSelection, deselected *qt.QItemSelection) {
			delegate.SetSelectedRows(sm.SelectedIndexes())
			if vp := view.Viewport(); vp != nil {
				vp.Update()
			}
		})
	}
	// OnLeaveEvent / OnMouseReleaseEvent are virtual overrides and would panic on
	// a non-directly-constructed QAbstractItemView; install an event filter on
	// the viewport instead to reset the transient hover/pressed state.
	filter := qt.NewQObject2(view.QObject)
	if vp := view.Viewport(); vp != nil {
		vp.InstallEventFilter(filter)
	}
	filter.OnEventFilter(func(super func(watched *qt.QObject, event *qt.QEvent) bool, watched *qt.QObject, event *qt.QEvent) bool {
		switch event.Type() {
		case qt.QEvent__Leave:
			delegate.SetHoverRow(-1)
			if vp := view.Viewport(); vp != nil {
				vp.Update()
			}
		case qt.QEvent__MouseButtonRelease:
			delegate.SetPressedRow(-1)
			if sm := view.SelectionModel(); sm != nil {
				delegate.SetSelectedRows(sm.SelectedIndexes())
			}
			if vp := view.Viewport(); vp != nil {
				vp.Update()
			}
		}
		return super(watched, event)
	})
}

func setupTableView(view *qt.QTableView, delegate *TableItemDelegate) {
	common.FluentStyleSheet(common.FluentTableView).Apply(view.QWidget, common.ThemeAuto)
	view.SetShowGrid(false)
	view.SetMouseTracking(true)
	view.SetAlternatingRowColors(true)
	view.SetItemDelegate(delegate.QAbstractItemDelegate)
	view.SetSelectionBehavior(qt.QAbstractItemView__SelectRows)
	view.HorizontalHeader().SetHighlightSections(false)
	view.VerticalHeader().SetHighlightSections(false)
	view.VerticalHeader().SetDefaultSectionSize(38)
	view.VerticalHeader().OnSectionClicked(func(row int) { view.SelectRow(row) })
	installItemTracking(view.QAbstractItemView, delegate)
}

// TableWidget is a fluent styled table widget.
type TableWidget struct {
	*qt.QTableWidget
	delegate                *TableItemDelegate
	scrollDelegate          *SmoothScrollDelegate
	isSelectRightClickedRow bool
}

// NewTableWidget builds a table widget.
func NewTableWidget(parent *qt.QWidget) *TableWidget {
	w := &TableWidget{QTableWidget: qt.NewQTableWidget(parent)}
	w.delegate = NewTableItemDelegate(w.QAbstractItemView)
	setupTableView(w.QTableView, w.delegate)
	w.scrollDelegate = NewSmoothScrollDelegate(w.QAbstractItemView.QAbstractScrollArea, false)
	return w
}

// SetBorderVisible sets the border visibility.
func (w *TableWidget) SetBorderVisible(isVisible bool) {
	w.SetProperty("isBorderVisible", qt.NewQVariant11(isVisible))
	w.SetStyle(qt.QApplication_Style())
}

// SetBorderRadius sets the border radius via a custom stylesheet.
func (w *TableWidget) SetBorderRadius(radius int) {
	qss := fmt.Sprintf("QTableView{border-radius: %dpx}", radius)
	common.SetCustomStyleSheet(w.QWidget, qss, qss)
}

// SetCheckedColor sets the checked indicator color.
func (w *TableWidget) SetCheckedColor(light, dark *qt.QColor) {
	w.delegate.SetCheckedColor(light, dark)
}

// SetCurrentCell sets the current cell.
func (w *TableWidget) SetCurrentCell(row, column int) {
	w.SetCurrentItem(w.Item(row, column))
}

// IsSelectRightClickedRow reports whether right-click selects the row.
func (w *TableWidget) IsSelectRightClickedRow() bool { return w.isSelectRightClickedRow }

// SetSelectRightClickedRow toggles right-click row selection.
func (w *TableWidget) SetSelectRightClickedRow(isSelect bool) { w.isSelectRightClickedRow = isSelect }

// TableView is a fluent styled table view.
type TableView struct {
	*qt.QTableView
	delegate                *TableItemDelegate
	scrollDelegate          *SmoothScrollDelegate
	isSelectRightClickedRow bool
}

// NewTableView builds a table view.
func NewTableView(parent *qt.QWidget) *TableView {
	w := &TableView{QTableView: qt.NewQTableView(parent)}
	w.delegate = NewTableItemDelegate(w.QAbstractItemView)
	setupTableView(w.QTableView, w.delegate)
	w.scrollDelegate = NewSmoothScrollDelegate(w.QAbstractItemView.QAbstractScrollArea, false)
	return w
}

// SetBorderVisible sets the border visibility.
func (w *TableView) SetBorderVisible(isVisible bool) {
	w.SetProperty("isBorderVisible", qt.NewQVariant11(isVisible))
	w.SetStyle(qt.QApplication_Style())
}

// SetBorderRadius sets the border radius via a custom stylesheet.
func (w *TableView) SetBorderRadius(radius int) {
	qss := fmt.Sprintf("QTableView{border-radius: %dpx}", radius)
	common.SetCustomStyleSheet(w.QWidget, qss, qss)
}

// SetCheckedColor sets the checked indicator color.
func (w *TableView) SetCheckedColor(light, dark *qt.QColor) {
	w.delegate.SetCheckedColor(light, dark)
}

// IsSelectRightClickedRow reports whether right-click selects the row.
func (w *TableView) IsSelectRightClickedRow() bool { return w.isSelectRightClickedRow }

// SetSelectRightClickedRow toggles right-click row selection.
func (w *TableView) SetSelectRightClickedRow(isSelect bool) { w.isSelectRightClickedRow = isSelect }
