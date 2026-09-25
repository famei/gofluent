package widgets

import (
	"fmt"

	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// TreeItemDelegate is the fluent item delegate for tree views. It draws the
// subtle rounded hover / selected highlight and the selection indicator after
// the native text rendering, matching the Python TreeItemDelegate.
type TreeItemDelegate struct {
	*qt.QStyledItemDelegate
	view              *qt.QAbstractItemView
	lightCheckedColor *qt.QColor
	darkCheckedColor  *qt.QColor
}

// NewTreeItemDelegate builds a tree item delegate.
func NewTreeItemDelegate(view *qt.QAbstractItemView) *TreeItemDelegate {
	d := &TreeItemDelegate{
		QStyledItemDelegate: qt.NewQStyledItemDelegate2(view.QObject),
		view:                view,
		lightCheckedColor:   qt.NewQColor(),
		darkCheckedColor:    qt.NewQColor(),
	}
	d.installPaint()
	return d
}

// SetCheckedColor sets the indicator color in checked status.
func (d *TreeItemDelegate) SetCheckedColor(light, dark *qt.QColor) {
	d.lightCheckedColor = cloneColor(light)
	d.darkCheckedColor = cloneColor(dark)
	if d.view != nil {
		if vp := d.view.Viewport(); vp != nil {
			vp.Update()
		}
	}
}

// installPaint overrides the delegate paint virtual so the Fluent highlight is
// painted on top of the native item rendering.
func (d *TreeItemDelegate) installPaint() {
	d.OnPaint(func(super func(painter *qt.QPainter, option *qt.QStyleOptionViewItem, index *qt.QModelIndex), painter *qt.QPainter, option *qt.QStyleOptionViewItem, index *qt.QModelIndex) {
		painter.SetRenderHints(qt.QPainter__Antialiasing | qt.QPainter__TextAntialiasing)
		super(painter, option, index)

		if _, ok := checkStateOf(index); ok {
			d.drawCheckBox(painter, option, index)
		}

		if (option.State() & (qt.QStyle__State_Selected | qt.QStyle__State_MouseOver)) == 0 {
			return
		}

		painter.Save()
		painter.SetPenWithStyle(qt.NoPen)
		d.drawBackground(painter, option, index)
		if (option.State()&qt.QStyle__State_Selected) != 0 && d.view.HorizontalScrollBar().Value() == 0 {
			d.drawIndicator(painter, option, index)
		}
		painter.Restore()
	})

	d.OnInitStyleOption(func(super func(option *qt.QStyleOptionViewItem, index *qt.QModelIndex), option *qt.QStyleOptionViewItem, index *qt.QModelIndex) {
		super(option, index)
		initItemViewOptionText(option)
	})
}

// drawBackground draws the rounded hover/selected row highlight, joining
// adjacent columns into a single strip with rounded outer corners (the Go port
// of TreeItemDelegate._drawBackground).
func (d *TreeItemDelegate) drawBackground(painter *qt.QPainter, option *qt.QStyleOptionViewItem, index *qt.QModelIndex) {
	c := 255
	if !common.IsDarkTheme() {
		c = 0
	}
	color := qt.NewQColor11(c, c, c, 9)
	brush := qt.NewQBrush3(color)
	painter.SetBrush(brush)
	brush.Delete()
	color.Delete()

	r := option.Rect() // GoGC-armed — do NOT Delete
	column := index.Column()
	lastColumn := 0
	if m := index.Model(); m != nil {
		lastColumn = m.ColumnCount(index.Parent()) - 1
	}

	radius := 4.0
	rf := qt.NewQRectF4(float64(r.X()), float64(r.Y())+2, float64(r.Width()), float64(r.Height())-4)
	defer rf.Delete()
	if column == 0 {
		rf.SetX(4)
	}

	path := qt.NewQPainterPath()
	defer path.Delete()

	switch {
	case column == 0 && column == lastColumn:
		path.AddRoundedRect(rf, radius, radius)
	case column == 0:
		path.MoveTo2(rf.Right(), rf.Top())
		path.LineTo2(rf.Right(), rf.Bottom())
		path.LineTo2(rf.X()+radius, rf.Bottom())
		path.ArcTo2(rf.X(), rf.Bottom()-2*radius, 2*radius, 2*radius, 270, -90)
		path.LineTo2(rf.X(), rf.Top()+radius)
		path.ArcTo2(rf.X(), rf.Top(), 2*radius, 2*radius, 180, -90)
		path.CloseSubpath()
	case column == lastColumn:
		path.MoveTo2(rf.X(), rf.Top())
		path.LineTo2(rf.Right()-radius, rf.Top())
		path.ArcTo2(rf.Right()-2*radius, rf.Top(), 2*radius, 2*radius, 90, -90)
		path.LineTo2(rf.Right(), rf.Bottom()-radius)
		path.ArcTo2(rf.Right()-2*radius, rf.Bottom()-2*radius, 2*radius, 2*radius, 0, -90)
		path.LineTo2(rf.X(), rf.Bottom())
		path.CloseSubpath()
	default:
		path.AddRect(rf)
	}

	painter.DrawPath(path)
}

// drawCheckBox draws the Fluent check box of an item that carries a
// CheckStateRole value (the Go port of TreeItemDelegate._drawCheckBox).
func (d *TreeItemDelegate) drawCheckBox(painter *qt.QPainter, option *qt.QStyleOptionViewItem, index *qt.QModelIndex) {
	state, ok := checkStateOf(index)
	if !ok {
		return
	}

	painter.Save()
	defer painter.Restore()

	r := 4.5
	rect := option.Rect() // GoGC-armed — do NOT Delete
	rf := qt.NewQRectF4(float64(rect.X())+23, float64(rect.Y()+rect.Height()/2)-9, 19, 19)
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

// drawIndicator draws the accent bar at the left of a selected row.
func (d *TreeItemDelegate) drawIndicator(painter *qt.QPainter, option *qt.QStyleOptionViewItem, index *qt.QModelIndex) {
	r := option.Rect() // GoGC-armed — do NOT Delete
	h := r.Height() - 4

	indColor := common.AutoFallbackThemeColor(d.lightCheckedColor, d.darkCheckedColor)
	indBrush := qt.NewQBrush3(indColor)
	painter.SetBrush(indBrush)
	indBrush.Delete()

	rf := qt.NewQRectF4(4, 9+float64(r.Y()), 3, float64(h)-13)
	defer rf.Delete()
	painter.DrawRoundedRect(rf, 1.5, 1.5)
}

func setupTreeView(view *qt.QTreeView, delegate *TreeItemDelegate) {
	common.FluentStyleSheet(common.FluentTreeView).Apply(view.QWidget, common.ThemeAuto)
	view.Header().SetHighlightSections(false)
	view.Header().SetDefaultAlignment(qt.AlignCenter)
	view.SetItemDelegate(delegate.QAbstractItemDelegate)
	view.SetIconSize(qt.NewQSize2(16, 16))
	view.SetMouseTracking(true)
}

// treeBranchLevel counts the ancestors of index (0 for a top-level item).
func treeBranchLevel(index *qt.QModelIndex) int {
	level := 0
	p := index.Parent()
	for p != nil && p.IsValid() {
		level++
		p = p.Parent()
	}
	return level
}

// treeBranchSVG returns the SVG of the expander chevron of a branch: the open or
// the closed image of the active theme (the QSS branch images of the Python
// resource bundle are not available here, so they are drawn by hand).
func treeBranchSVG(view *qt.QTreeView, index *qt.QModelIndex) []byte {
	name := "tree_view/TreeViewOpen"
	if !view.IsExpanded(index) {
		name = "tree_view/TreeViewClose"
	}
	return readEmbeddedImage(name + "_" + common.GetIconColor(common.ThemeAuto, false) + ".svg")
}

// drawTreeBranchIndicator draws the Fluent chevron for a branch that has
// children. The QSS branch images of the Python resource bundle are not available
// here, so the TreeViewOpen/TreeViewClose SVGs are drawn directly (the Go port of
// the QSS branch indicator).
func drawTreeBranchIndicator(view *qt.QTreeView, painter *qt.QPainter, rect *qt.QRect, index *qt.QModelIndex) {
	model := index.Model()
	if model == nil || model.RowCount(index) <= 0 {
		return
	}

	svgBytes := treeBranchSVG(view, index)
	if len(svgBytes) == 0 {
		return
	}

	// Draw the 16px chevron where the reference draws it: drawBranches does
	// rect.moveLeft(15) and Qt centers the 16px branch image inside the
	// indentation-wide branch rect, so the glyph lands at
	// level*indentation + 15 + (indentation-16)/2. Vertically centered in the
	// branch rect.
	indent := view.Indentation()
	x := float64(treeBranchLevel(index)*indent + 15 + (indent-16)/2)
	y := float64(rect.Y()) + (float64(rect.Height())-16)/2
	rf := qt.NewQRectF4(x, y, 16, 16)
	defer rf.Delete()
	drawSvgBytes(svgBytes, painter, rf)
}

// handleTreeBranchClick toggles expand/collapse when the mouse press lands on
// the branch chevron strip (the Go port of TreeViewBase.viewportEvent). The hit
// area is (level*indentation+20, level*indentation+30), matching the reference.
func handleTreeBranchClick(view *qt.QTreeView, event *qt.QEvent) {
	if event.Type() != qt.QEvent__MouseButtonPress {
		return
	}
	me := qt.UnsafeNewQMouseEvent(event.UnsafePointer())
	pos := me.Pos()
	index := view.IndexAt(pos)
	if !index.IsValid() {
		return
	}

	indent := treeBranchLevel(index)*view.Indentation() + 20
	x := pos.X()
	if x <= indent || x >= indent+10 {
		return
	}
	if view.IsExpanded(index) {
		view.Collapse(index)
	} else {
		view.Expand(index)
	}
}

// TreeWidget is a fluent styled tree widget.
type TreeWidget struct {
	*qt.QTreeWidget
	delegate       *TreeItemDelegate
	scrollDelegate *SmoothScrollDelegate
}

// NewTreeWidget builds a tree widget.
func NewTreeWidget(parent *qt.QWidget) *TreeWidget {
	w := &TreeWidget{QTreeWidget: qt.NewQTreeWidget(parent)}
	w.delegate = NewTreeItemDelegate(w.QAbstractItemView)
	setupTreeView(w.QTreeView, w.delegate)
	// Fluent overlay scroll bars instead of the native ones.
	w.scrollDelegate = NewSmoothScrollDelegate(w.QAbstractScrollArea, false)
	// OnDrawBranches / OnViewportEvent are overridden on the directly
	// constructed QTreeWidget (not the embedded QTreeView base) to avoid the
	// miqt "directly constructed" panic.
	w.QTreeWidget.OnDrawBranches(func(super func(painter *qt.QPainter, rect *qt.QRect, index *qt.QModelIndex), painter *qt.QPainter, rect *qt.QRect, index *qt.QModelIndex) {
		super(painter, rect, index)
		drawTreeBranchIndicator(w.QTreeView, painter, rect, index)
	})
	w.QTreeWidget.OnViewportEvent(func(super func(event *qt.QEvent) bool, event *qt.QEvent) bool {
		if event.Type() == qt.QEvent__MouseButtonPress {
			handleTreeBranchClick(w.QTreeView, event)
		}
		return super(event)
	})
	return w
}

// SetBorderVisible sets the border visibility.
func (w *TreeWidget) SetBorderVisible(isVisible bool) {
	w.SetProperty("isBorderVisible", qt.NewQVariant11(isVisible))
	w.SetStyle(qt.QApplication_Style())
}

// SetBorderRadius sets the border radius via a custom stylesheet.
func (w *TreeWidget) SetBorderRadius(radius int) {
	qss := fmt.Sprintf("QTreeView{border-radius: %dpx}", radius)
	common.SetCustomStyleSheet(w.QWidget, qss, qss)
}

// SetCheckedColor sets the checked indicator color.
func (w *TreeWidget) SetCheckedColor(light, dark *qt.QColor) {
	w.delegate.SetCheckedColor(light, dark)
}

// TreeView is a fluent styled tree view.
type TreeView struct {
	*qt.QTreeView
	delegate       *TreeItemDelegate
	scrollDelegate *SmoothScrollDelegate
}

// NewTreeView builds a tree view.
func NewTreeView(parent *qt.QWidget) *TreeView {
	w := &TreeView{QTreeView: qt.NewQTreeView(parent)}
	w.delegate = NewTreeItemDelegate(w.QAbstractItemView)
	setupTreeView(w.QTreeView, w.delegate)
	// Fluent overlay scroll bars instead of the native ones.
	w.scrollDelegate = NewSmoothScrollDelegate(w.QAbstractScrollArea, false)
	w.QTreeView.OnDrawBranches(func(super func(painter *qt.QPainter, rect *qt.QRect, index *qt.QModelIndex), painter *qt.QPainter, rect *qt.QRect, index *qt.QModelIndex) {
		super(painter, rect, index)
		drawTreeBranchIndicator(w.QTreeView, painter, rect, index)
	})
	w.QTreeView.OnViewportEvent(func(super func(event *qt.QEvent) bool, event *qt.QEvent) bool {
		if event.Type() == qt.QEvent__MouseButtonPress {
			handleTreeBranchClick(w.QTreeView, event)
		}
		return super(event)
	})
	return w
}

// SetBorderVisible sets the border visibility.
func (w *TreeView) SetBorderVisible(isVisible bool) {
	w.SetProperty("isBorderVisible", qt.NewQVariant11(isVisible))
	w.SetStyle(qt.QApplication_Style())
}

// SetBorderRadius sets the border radius via a custom stylesheet.
func (w *TreeView) SetBorderRadius(radius int) {
	qss := fmt.Sprintf("QTreeView{border-radius: %dpx}", radius)
	common.SetCustomStyleSheet(w.QWidget, qss, qss)
}

// SetCheckedColor sets the checked indicator color.
func (w *TreeView) SetCheckedColor(light, dark *qt.QColor) {
	w.delegate.SetCheckedColor(light, dark)
}
