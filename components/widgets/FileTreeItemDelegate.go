package widgets

import (
	"math"
	"unsafe"

	"github.com/famei/gofluent/common"
	"github.com/mappu/miqt/qt"
	"github.com/mappu/miqt/qt/svg"
)

// ---------------------------------------------------------------------------
// FileTree icons

// FileTreeIcon identifies the glyph drawn in front of a FileTree item.
type FileTreeIcon int

const (
	// FileTreeIconFolder is the folder glyph (the default for items with
	// children).
	FileTreeIconFolder FileTreeIcon = iota
	// FileTreeIconDocument is the file glyph (the default for leaves).
	FileTreeIconDocument
	// FileTreeIconDrive is the drive glyph (the default for top level items).
	FileTreeIconDrive
)

// FileTreeIconProvider returns the fluent glyph of a model index; it is what a
// tree driven by SetModel uses instead of the icons the model carries.
type FileTreeIconProvider func(index *qt.QModelIndex) FileTreeIcon

// fileTreeIconRole is the item data role that stores an item's FileTreeIcon.
const fileTreeIconRole = int(qt.UserRole) + 1

// The geometry of a file tree row, measured against the Windows Explorer
// navigation pane.
const (
	// fileTreeIndentation is the horizontal step of one tree level.
	fileTreeIndentation = 19
	// fileTreeRowHeight is the default height of a row (the file table uses the
	// same height).
	fileTreeRowHeight = 20
	// fileTreeGlyphSize is the edge length of an item glyph.
	fileTreeGlyphSize = 16
	// fileTreeGlyphInset is the gap between the item rectangle and its glyph.
	fileTreeGlyphInset = 2
	// fileTreeChevronSize is the default size of the square the visible expander
	// arrow fits in.
	fileTreeChevronSize = 12
)

// hasBranch reports whether an item shows an expander chevron. A model loads its
// children lazily, so an item counts as a branch while it has children or still
// has to fetch them; a folder that is known to be empty is a leaf and shows no
// chevron.
func hasBranch(model *qt.QAbstractItemModel, index *qt.QModelIndex) bool {
	if model == nil || index == nil || !index.IsValid() {
		return false
	}
	if !model.HasChildren(index) {
		return false
	}
	return model.RowCount(index) > 0 || model.CanFetchMore(index)
}

// ---------------------------------------------------------------------------
// Branch chevron geometry

// svgInk is the normalized bounding box (0..1 of the view box) of the visible
// pixels of an SVG.
type svgInk struct{ X0, Y0, X1, Y1 float64 }

// width and height return the visible size of the ink box.
func (b svgInk) width() float64  { return b.X1 - b.X0 }
func (b svgInk) height() float64 { return b.Y1 - b.Y0 }

// svgInkCache holds the ink boxes of the branch images (four of them, one per
// theme and expanded state).
var svgInkCache = map[string]svgInk{}

// svgInkBox measures the visible part of an SVG icon by rendering it once. The
// TreeViewOpen/TreeViewClose images carry a wide transparent padding inside their
// view box (the glyph fills barely a third of it), so drawing them into a box of
// the requested size would give an arrow that is both much smaller than asked for
// and off center — the ink box is what lets the chevron be scaled and centered by
// what the user actually sees.
func svgInkBox(key string, svgBytes []byte) svgInk {
	if box, ok := svgInkCache[key]; ok {
		return box
	}

	const size = 64
	image := qt.NewQImage3(size, size, qt.QImage__Format_ARGB32)
	image.Fill(0) // transparent
	renderer := svg.NewQSvgRenderer3(svgBytes)
	painter := qt.NewQPainter2(image.QPaintDevice)
	renderer.Render2(painter, qt.NewQRectF4(0, 0, size, size))
	painter.End()
	painter.Delete()
	renderer.Delete()

	box := svgInk{X0: 1, Y0: 1}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if image.PixelColor(x, y).Alpha() == 0 {
				continue
			}
			fx, fy := float64(x)/size, float64(y)/size
			box.X0 = math.Min(box.X0, fx)
			box.Y0 = math.Min(box.Y0, fy)
			box.X1 = math.Max(box.X1, fx+1.0/size)
			box.Y1 = math.Max(box.Y1, fy+1.0/size)
		}
	}
	image.Delete()

	if box.X1 <= box.X0 || box.Y1 <= box.Y0 {
		return svgInk{} // nothing visible: fall back to the whole view box
	}
	svgInkCache[key] = box
	return box
}

// chevronInk returns the scale factor of a branch image and its ink box: the
// visible arrow is scaled so that its larger side is ChevronSize pixels. The open
// and the closed branch image have very different proportions (a wide flat "v"
// against a narrow tall ">"), so scaling them by height would draw an expanded
// arrow several times the size of a collapsed one; scaling by the larger side
// gives both the same visible size.
func (self *FileTree) chevronInk(svgBytes []byte, key string) (float64, svgInk) {
	ink := svgInkBox(key, svgBytes)
	extent := math.Max(ink.width(), ink.height())
	if extent <= 0 {
		return 0, ink
	}
	return float64(self.chevronSize) / extent, ink
}

// chevronCenterX returns the horizontal center of the indentation of one level,
// which is where the expander arrow of that level is centered.
func (self *FileTree) chevronCenterX(level int) float64 {
	return float64(level*self.Indentation()) + float64(self.Indentation())/2
}

// chevronCenterY returns the vertical center of the row the branch rectangle
// belongs to, which is where the expander arrow is centered.
func (self *FileTree) chevronCenterY(rect *qt.QRect) float64 {
	if rect == nil {
		return 0
	}
	return float64(rect.Y()) + float64(rect.Height())/2
}

// fluentIcon returns the fluent glyph of an icon kind.
func (i FileTreeIcon) fluentIcon() interface{} {
	switch i {
	case FileTreeIconDocument:
		return common.Document
	case FileTreeIconDrive:
		return common.HardDrive
	default:
		return common.Folder
	}
}

// DefaultFileTreeIcon picks the fluent glyph of a model index: a drive for a top
// level entry, a folder for anything that can have children and a document
// otherwise. A model flavor tree uses it when it is installed as the icon
// provider (see FileTree.SetIconProvider); by default the tree shows the icons
// the model provides.
func DefaultFileTreeIcon(index *qt.QModelIndex) FileTreeIcon {
	if index == nil || !index.IsValid() {
		return FileTreeIconDocument
	}
	parent := index.Parent()
	if parent == nil || !parent.IsValid() {
		return FileTreeIconDrive
	}
	if model := index.Model(); model != nil && model.HasChildren(index) {
		return FileTreeIconFolder
	}
	return FileTreeIconDocument
}

// ---------------------------------------------------------------------------
// FileTreeItemDelegate

// FileTreeItemDelegate paints the rows of a FileTree: the hover, highlight and
// selection backgrounds plus the outline of the selected row. It shares the
// FileTableColors palette with FileTableItemDelegate, so a file tree and a file
// table placed next to each other look like one control.
type FileTreeItemDelegate struct {
	*qt.QStyledItemDelegate
	view *qt.QTreeView

	borderPen *qt.QPen

	lightColors FileTableColors
	darkColors  FileTableColors

	// highlighted is keyed by the model index pointer: miqt hands out a new Go
	// wrapper for every call, so comparing item values would never match.
	highlighted map[unsafe.Pointer]struct{}

	// modelIcons draws the fluent glyph of every index instead of the icons the
	// model carries; it is on exactly while an icon provider is installed.
	modelIcons   bool
	iconProvider FileTreeIconProvider
	// emptyIcon is handed to the item option to hide the decoration a model
	// provides while keeping the space reserved for it.
	emptyIcon *qt.QIcon

	// rowHeight is the height of every row (see FileTree.SetRowHeight).
	rowHeight int
}

// NewFileTreeItemDelegate builds the delegate of a file tree.
func NewFileTreeItemDelegate(view *qt.QTreeView) *FileTreeItemDelegate {
	d := &FileTreeItemDelegate{
		view:        view,
		lightColors: LightFileTableColors(),
		darkColors:  DarkFileTableColors(),
		highlighted: make(map[unsafe.Pointer]struct{}),
		rowHeight:   fileTreeRowHeight,
	}
	d.Init_Gui(view)
	return d
}

// Init_Gui wires the delegate to its view.
func (d *FileTreeItemDelegate) Init_Gui(view *qt.QTreeView) {
	d.QStyledItemDelegate = qt.NewQStyledItemDelegate2(view.QObject)
	d.borderPen = qt.NewQPen()
	d.borderPen.SetWidthF(0.6)
	d.borderPen.SetColor(d.lightColors.Border)
	d.emptyIcon = qt.NewQIcon()
	d.OnPaint(d.Paint)
	d.OnInitStyleOption(d.InitStyleOption)
	d.OnSizeHint(d.SizeHint)
	d.OnDestroyed(func() {
		deleteFileTableColors(d.lightColors)
		deleteFileTableColors(d.darkColors)
		d.emptyIcon.Delete()
	})
}

// Colors returns the palette of the active theme (owned by the delegate).
func (d *FileTreeItemDelegate) Colors() FileTableColors {
	if common.IsDarkTheme() {
		return d.darkColors
	}
	return d.lightColors
}

// SetColors replaces both palettes; missing colors keep the defaults.
func (d *FileTreeItemDelegate) SetColors(light, dark FileTableColors) {
	d.SetLightColors(light)
	d.SetDarkColors(dark)
}

// SetLightColors replaces the light palette; missing colors keep the defaults.
func (d *FileTreeItemDelegate) SetLightColors(colors FileTableColors) {
	deleteFileTableColors(d.lightColors)
	d.lightColors = cloneFileTableColors(colors, LightFileTableColors())
	d.repaint()
}

// SetDarkColors replaces the dark palette; missing colors keep the defaults.
func (d *FileTreeItemDelegate) SetDarkColors(colors FileTableColors) {
	deleteFileTableColors(d.darkColors)
	d.darkColors = cloneFileTableColors(colors, DarkFileTableColors())
	d.repaint()
}

// SetHighlightColor sets the highlight color of both themes.
func (d *FileTreeItemDelegate) SetHighlightColor(light, dark *qt.QColor) {
	if d.lightColors.Highlight != nil {
		d.lightColors.Highlight.Delete()
	}
	if d.darkColors.Highlight != nil {
		d.darkColors.Highlight.Delete()
	}
	d.lightColors.Highlight = cloneColor(light)
	d.darkColors.Highlight = cloneColor(dark)
	d.repaint()
}

// SetHighlightIndexes marks the given model indexes with the highlight
// background (a search result, a drop target, ...); nil clears the highlight.
func (d *FileTreeItemDelegate) SetHighlightIndexes(indexes []*qt.QModelIndex) {
	d.highlighted = make(map[unsafe.Pointer]struct{}, len(indexes))
	for _, index := range indexes {
		if index != nil && index.IsValid() {
			d.highlighted[index.InternalPointer()] = struct{}{}
		}
	}
	d.repaint()
}

// SetIconProvider replaces the icons of a model flavor tree with the fluent
// glyphs the provider returns (nil restores the icons the model provides) and
// repaints the rows.
func (d *FileTreeItemDelegate) SetIconProvider(provider FileTreeIconProvider) {
	d.iconProvider = provider
	d.modelIcons = provider != nil
	d.repaint()
}

// SetRowHeight sets the height of every row and lays the rows out again (the
// expansion state is kept).
func (d *FileTreeItemDelegate) SetRowHeight(height int) {
	if height <= 0 || height == d.rowHeight {
		return
	}
	d.rowHeight = height
	if d.view != nil {
		d.view.DoItemsLayout()
	}
	d.repaint()
}

// repaint repaints the viewport this delegate paints into.
func (d *FileTreeItemDelegate) repaint() {
	if d.view == nil {
		return
	}
	if vp := d.view.Viewport(); vp != nil {
		vp.Update()
	}
}

// InitStyleOption hides the decoration a model carries while keeping the space
// it occupies: in the model flavor the delegate draws the fluent glyph of the
// index itself, in the very position the model icon would have been drawn in.
func (d *FileTreeItemDelegate) InitStyleOption(super func(option *qt.QStyleOptionViewItem, index *qt.QModelIndex), option *qt.QStyleOptionViewItem, index *qt.QModelIndex) {
	super(option, index)
	if !d.modelIcons {
		return
	}
	option.SetIcon(*d.emptyIcon)
	size := qt.NewQSize2(fileTreeGlyphSize, fileTreeGlyphSize)
	option.SetDecorationSize(*size)
}

// SizeHint gives every row the configured height (the width still comes from the
// base delegate).
func (d *FileTreeItemDelegate) SizeHint(super func(option *qt.QStyleOptionViewItem, index *qt.QModelIndex) *qt.QSize, option *qt.QStyleOptionViewItem, index *qt.QModelIndex) *qt.QSize {
	size := super(option, index) // GoGC-armed — do NOT Delete
	if d.rowHeight > 0 {
		size.SetHeight(d.rowHeight)
	}
	return size
}

// Paint draws the row background of one item and lets the base delegate draw the
// icon and the text on top of it.
func (d *FileTreeItemDelegate) Paint(super func(painter *qt.QPainter, option *qt.QStyleOptionViewItem, index *qt.QModelIndex), painter *qt.QPainter, option *qt.QStyleOptionViewItem, index *qt.QModelIndex) {
	colors := d.Colors()
	state := option.State()

	painter.Save()
	painter.SetPenWithStyle(qt.NoPen)
	painter.SetRenderHint2(qt.QPainter__Antialiasing, false) // 暂时关闭抗锯齿 过小的边框会导致颜色不正确
	painter.SetClipping(true)
	// The clip widens the item rectangle by one pixel, so the outline of a
	// selected row keeps its left edge (the file table outlines its selected
	// rows the same way).
	painter.SetClipRectWithQRect(option.Rect().Adjusted(-1, 0, 1, 0))
	// The same two pixel row gap as the file table, plus one free pixel at the
	// right so the outline is not hidden under the overlay scroll bar.
	option.SetRect(*option.Rect().Adjusted(0, 2, -1, -2))

	// The row starts behind the indentation, so Windows Explorer indents the
	// highlight with the level of the item.
	row := option.Rect() // GoGC-armed — do NOT Delete

	if _, ok := d.highlighted[index.InternalPointer()]; ok {
		painter.FillRect6(row, colors.Highlight)
	}
	if state&qt.QStyle__State_MouseOver != 0 {
		painter.FillRect6(row, colors.Hover)
	}
	if state&qt.QStyle__State_Selected != 0 {
		// Windows Explorer keeps the selection light blue while the tree owns
		// the focus and greys it otherwise.
		background := colors.NonFocus
		if d.isFocused() {
			background = colors.Selected
		}
		d.borderPen.SetColor(colors.Border)
		painter.SetPenWithPen(d.borderPen)
		painter.FillRect6(row, background)
		painter.DrawRectWithRect(row)
	}
	painter.Restore()

	super(painter, option, index)

	if d.modelIcons {
		d.drawIndexIcon(painter, option, index)
	}
}

// drawIndexIcon draws the fluent glyph of a model index where the model icon
// would have been drawn (the model icon itself is hidden by InitStyleOption).
func (d *FileTreeItemDelegate) drawIndexIcon(painter *qt.QPainter, option *qt.QStyleOptionViewItem, index *qt.QModelIndex) {
	if d.iconProvider == nil {
		return
	}
	glyph := d.iconProvider(index).fluentIcon()

	row := option.Rect() // GoGC-armed — do NOT Delete
	rect := qt.NewQRectF4(
		float64(row.X()+fileTreeGlyphInset),
		float64(row.Y())+(float64(row.Height())-fileTreeGlyphSize)/2,
		fileTreeGlyphSize, fileTreeGlyphSize)
	defer rect.Delete()
	renderFluentIcon(glyph, painter, rect, common.ThemeAuto)
}

// isFocused reports whether the tree owning this delegate is the focused widget
// of an active window.
func (d *FileTreeItemDelegate) isFocused() bool {
	if d.view == nil {
		return false
	}
	return d.view.HasFocus() && d.view.IsActiveWindow()
}

// ---------------------------------------------------------------------------
// FileTree

// fileTreeLightQss is the light stylesheet of the file tree; it mirrors the file
// table stylesheet (transparent rows, themed text color) and turns the native
// branch decoration off, because the chevron is drawn by drawChevron. The row
// height is not part of the stylesheet: the delegate gives every row the height
// set with FileTree.SetRowHeight.
const fileTreeLightQss = `
	QTreeView {
		background: transparent;
		outline: none;
		border: none;
		selection-background-color: transparent;
		alternate-background-color: transparent;
	}

	QTreeView::item {
		background: transparent;
		border: 0px;
		color: rgb(0, 0, 0);
	}

	QTreeView::item:selected {
		color: rgb(0, 0, 0);
		background: transparent;
	}

	QTreeView::branch {
		background: transparent;
		image: none;
		border-image: none;
	}

	QHeaderView {
		background-color: transparent;
	}`

// fileTreeDarkQss is the dark stylesheet of the file tree.
const fileTreeDarkQss = `
	QTreeView {
		background: transparent;
		outline: none;
		border: none;
		selection-background-color: transparent;
		alternate-background-color: transparent;
	}

	QTreeView::item {
		background: transparent;
		border: 0px;
		color: rgb(255, 255, 255);
	}

	QTreeView::item:selected {
		color: rgb(255, 255, 255);
		background: transparent;
	}

	QTreeView::branch {
		background: transparent;
		image: none;
		border-image: none;
	}

	QHeaderView {
		background-color: transparent;
	}`

// FileTree is a Windows Explorer style tree view: a single column QTreeView
// whose rows are painted by FileTreeItemDelegate, so it shares the FileTable
// look (row palette, hover, selection outline), the fluent overlay scroll bars
// and the live light/dark theme switching.
//
// It is filled either with items (AddItem / SetItemIcon, backed by an internal
// QStandardItemModel) or driven by an application model with SetModel — the
// Explorer example lists the local file system with QFileSystemModel:
//
//   - the chevron in the indentation expands and collapses the branch without
//     changing the selection (see consumeExpanderClick); Qt handles hover, the
//     selection and the double click expansion;
//   - an item shows a chevron while it has children or still has to load them
//     (see hasBranch), so an empty folder is drawn as a leaf;
//   - the icons come from the items or from the model; the item icons are
//     re-rendered on a theme switch, and an icon provider can replace the model
//     icons with theme aware fluent glyphs;
//   - the chevron itself is theme aware as well (treeBranchSVG picks the black or
//     the white branch image from the active theme).
type FileTree struct {
	*qt.QTreeView
	delegate       *FileTreeItemDelegate
	scrollDelegate *SmoothScrollDelegate
	themeAlive     *widgetAlive

	// items is the model behind the item flavor; it is replaced by SetModel.
	items *qt.QStandardItemModel

	// chevronSize is the edge length of the expander chevron.
	chevronSize int

	// viewportFilter observes the viewport mouse events (see
	// filterViewportEvent).
	viewportFilter *qt.QObject

	// modelMode records that the tree is driven by an application model instead
	// of its own items.
	modelMode bool

	// pressConsumed records that a press was answered by the expander chevron,
	// so the matching release is swallowed as well and never reaches Qt as a
	// click (which would select the item the chevron belongs to).
	pressConsumed bool
}

// NewFileTree builds a file tree.
func NewFileTree(parent *qt.QWidget) *FileTree {
	w := &FileTree{}
	w.Init_Gui(parent)
	return w
}

// Init_Gui builds the file tree.
func (self *FileTree) Init_Gui(parent *qt.QWidget) {
	self.QTreeView = qt.NewQTreeView(parent)
	self.chevronSize = fileTreeChevronSize
	// The item flavor is a plain QStandardItemModel: AddItem fills it and
	// SetModel replaces it.
	self.items = qt.NewQStandardItemModel3(self.QObject)
	self.items.SetColumnCount(1)
	self.QTreeView.SetModel(self.items.QAbstractItemModel)

	self.SetHeaderHidden(true)
	self.SetRootIsDecorated(true)
	self.SetIndentation(fileTreeIndentation) // Windows Explorer indents 19px per level
	self.SetUniformRowHeights(true)
	// The expand/collapse animation is off: Explorer toggles a branch instantly.
	self.SetAnimated(false)
	self.SetExpandsOnDoubleClick(true)
	self.SetSelectionMode(qt.QAbstractItemView__SingleSelection)
	self.SetMouseTracking(true) // 启用鼠标追踪
	self.SetIconSize(qt.NewQSize2(fileTreeGlyphSize, fileTreeGlyphSize))

	// The stylesheet follows the theme: common.SetCustomStyleSheet registers the
	// tree, so common.SetTheme re-applies the matching QSS automatically.
	common.SetCustomStyleSheet(self.QWidget, fileTreeLightQss, fileTreeDarkQss)

	self.delegate = NewFileTreeItemDelegate(self.QTreeView)
	self.SetItemDelegate(self.delegate.QAbstractItemDelegate)
	// Fluent overlay scroll bars (the native ones are hidden by the delegate).
	self.scrollDelegate = NewSmoothScrollDelegate(self.QAbstractScrollArea, false)

	self.initTheme()

	// The chevron of the fluent tree is drawn by hand (the QSS branch images are
	// not resolvable through the embedded resource system).
	self.QTreeView.OnDrawBranches(func(super func(painter *qt.QPainter, rect *qt.QRect, index *qt.QModelIndex), painter *qt.QPainter, rect *qt.QRect, index *qt.QModelIndex) {
		super(painter, rect, index)
		self.drawChevron(painter, rect, index)
	})

	// The viewport mouse events are observed with an event filter instead of
	// overriding viewportEvent: miqt's override forwards to
	// QAbstractScrollArea::viewportEvent, which does not hand the mouse events to
	// QAbstractItemView, so Qt's own selection handling would never run. The
	// filter returns false for everything it does not consume, which keeps the
	// normal click path intact.
	self.viewportFilter = qt.NewQObject2(self.QObject)
	self.Viewport().InstallEventFilter(self.viewportFilter)
	self.viewportFilter.OnEventFilter(func(super func(watched *qt.QObject, event *qt.QEvent) bool, watched *qt.QObject, event *qt.QEvent) bool {
		return self.filterViewportEvent(super, watched, event)
	})
}

// SetModel drives the tree from an application model instead of its items, which
// is how the Explorer example lists the local file system:
//
//	model := qt.NewQFileSystemModel2(tree.QObject)
//	model.SetRootPath(qt.QDir_RootPath())
//	tree.SetModel(model.QAbstractItemModel)
//
// The icons come from the model (a file system model brings the native file
// icons); install a FileTreeIconProvider with SetIconProvider to draw fluent
// glyphs instead. The columns behind the first one are hidden, because a file
// tree shows names only.
func (self *FileTree) SetModel(model *qt.QAbstractItemModel) {
	self.modelMode = true
	self.QTreeView.SetModel(model)

	self.SetHeaderHidden(true)
	self.SetRootIsDecorated(true)
	self.SetUniformRowHeights(true)
	self.SetAnimated(false)
	self.SetIndentation(fileTreeIndentation)

	if model != nil {
		// miqt's model accessors dereference the parent index, so the root is an
		// invalid index rather than nil.
		root := qt.NewQModelIndex()
		defer root.Delete()
		if columns := model.ColumnCount(root); columns > 1 {
			for column := 1; column < columns; column++ {
				self.SetColumnHidden(column, true)
			}
		}
	}
}

// IsModelMode reports whether the tree is driven by an application model.
func (self *FileTree) IsModelMode() bool { return self.modelMode }

// ItemModel returns the model behind the item flavor.
func (self *FileTree) ItemModel() *qt.QStandardItemModel { return self.items }

// SetIconProvider replaces the icons of a model flavor tree with the fluent
// glyphs the provider returns; nil (the default) keeps the icons the model
// provides.
func (self *FileTree) SetIconProvider(provider FileTreeIconProvider) {
	self.delegate.SetIconProvider(provider)
}

// SetRowHeight sets the height of every row in pixels; the default is 20, the
// height the FileTable uses.
func (self *FileTree) SetRowHeight(height int) {
	self.delegate.SetRowHeight(height)
}

// RowHeight returns the height of a row in pixels.
func (self *FileTree) RowHeight() int { return self.delegate.rowHeight }

// SetChevronSize sets the size of the square the visible expander arrow fits in,
// in pixels (the default is 12). The arrow is centered in the indentation of its
// level, so one that is wider than the indentation reaches into the column of its
// parent; pair a large chevron with a larger indentation.
func (self *FileTree) SetChevronSize(size int) {
	if size <= 0 || size == self.chevronSize {
		return
	}
	self.chevronSize = size
	if vp := self.Viewport(); vp != nil {
		vp.Update()
	}
}

// ChevronSize returns the size of the square the visible expander arrow fits in.
func (self *FileTree) ChevronSize() int { return self.chevronSize }

// initTheme re-renders the item glyphs and repaints the rows when the
// application theme changes; the stylesheet itself is refreshed by the
// stylesheet manager.
func (self *FileTree) initTheme() {
	self.themeAlive = trackWidget(self.OnDestroyed)
	common.QConfigInstance.OnThemeChanged(func(common.Theme) {
		if !self.themeAlive.ok() {
			return
		}
		self.RefreshItemIcons()
		if vp := self.Viewport(); vp != nil {
			vp.Update()
		}
	})
}

// SetColors replaces the row palette of both themes (see FileTableColors).
func (self *FileTree) SetColors(light, dark FileTableColors) {
	self.delegate.SetColors(light, dark)
}

// SetHighlightColor sets the highlight color of both themes.
func (self *FileTree) SetHighlightColor(light, dark *qt.QColor) {
	self.delegate.SetHighlightColor(light, dark)
}

// SetHighlight marks items with the highlight background; nil clears it.
func (self *FileTree) SetHighlight(items []*qt.QStandardItem) {
	indexes := make([]*qt.QModelIndex, 0, len(items))
	for _, item := range items {
		if item != nil {
			indexes = append(indexes, item.Index())
		}
	}
	self.delegate.SetHighlightIndexes(indexes)
}

// SetHighlightIndexes marks model indexes with the highlight background; nil
// clears it. It is the SetModel flavor of SetHighlight and expects the caller to
// re-apply the highlight when the model changes underneath it.
func (self *FileTree) SetHighlightIndexes(indexes []*qt.QModelIndex) {
	self.delegate.SetHighlightIndexes(indexes)
}

// SetItemIcon sets the glyph of an item.
func (self *FileTree) SetItemIcon(item *qt.QStandardItem, icon FileTreeIcon) {
	if item == nil {
		return
	}
	value := qt.NewQVariant7(int(icon))
	item.SetData(value, fileTreeIconRole)
	value.Delete()

	rendered := common.ToQIcon(icon.fluentIcon())
	item.SetIcon(rendered)
	rendered.Delete()
}

// ItemIcon returns the glyph of an item: the icon set with SetItemIcon, a folder
// when the item has children, a drive for a childless top level item and a
// document otherwise.
func (self *FileTree) ItemIcon(item *qt.QStandardItem) FileTreeIcon {
	if item == nil {
		return FileTreeIconFolder
	}
	if data := item.Data(fileTreeIconRole); data != nil && !data.IsNull() {
		return FileTreeIcon(data.ToInt())
	}
	if item.RowCount() > 0 {
		return FileTreeIconFolder
	}
	if item.Parent() == nil {
		return FileTreeIconDrive
	}
	return FileTreeIconDocument
}

// AddItem appends an item (a nil parent adds a top level item) with a text and a
// glyph, and returns it.
func (self *FileTree) AddItem(parent *qt.QStandardItem, text string, icon FileTreeIcon) *qt.QStandardItem {
	item := qt.NewQStandardItem2(text)
	item.SetEditable(false)
	if parent != nil {
		parent.AppendRow([]*qt.QStandardItem{item})
	} else {
		self.items.AppendRow([]*qt.QStandardItem{item})
	}
	self.SetItemIcon(item, icon)
	return item
}

// SetItemExpanded expands or collapses an item of the item flavor.
func (self *FileTree) SetItemExpanded(item *qt.QStandardItem, expanded bool) {
	if item == nil {
		return
	}
	if index := item.Index(); index != nil && index.IsValid() {
		self.SetExpanded(index, expanded)
	}
}

// IsItemExpanded reports whether an item of the item flavor is expanded.
func (self *FileTree) IsItemExpanded(item *qt.QStandardItem) bool {
	if item == nil {
		return false
	}
	index := item.Index()
	return index != nil && index.IsValid() && self.IsExpanded(index)
}

// RefreshItemIcons re-renders the glyph of every item. The glyphs are baked into
// a QIcon when they are set (the ink color is fixed at that moment), so they have
// to be rebuilt after a theme switch; the widget does that by itself, this method
// is only needed when the caller replaces item icons. A model flavor tree draws
// its glyphs per paint and needs no refresh.
func (self *FileTree) RefreshItemIcons() {
	if self.modelMode {
		return
	}
	for _, item := range self.allItems() {
		rendered := common.ToQIcon(self.ItemIcon(item).fluentIcon())
		item.SetIcon(rendered)
		rendered.Delete()
	}
}

// allItems flattens the item flavor tree.
func (self *FileTree) allItems() []*qt.QStandardItem {
	root := qt.NewQModelIndex()
	defer root.Delete()

	var items []*qt.QStandardItem
	var walk func(parent *qt.QStandardItem)
	walk = func(parent *qt.QStandardItem) {
		var count int
		var child func(i int) *qt.QStandardItem
		if parent == nil {
			count = self.items.RowCount(root)
			child = func(i int) *qt.QStandardItem { return self.items.Item(i) }
		} else {
			count = parent.RowCount()
			child = func(i int) *qt.QStandardItem { return parent.Child(i) }
		}
		for i := 0; i < count; i++ {
			item := child(i)
			if item == nil {
				continue
			}
			items = append(items, item)
			walk(item)
		}
	}
	walk(nil)
	return items
}

// chevronKey names the branch image of an item (one cache entry per theme and
// expanded state).
func chevronKey(view *qt.QTreeView, index *qt.QModelIndex) string {
	name := "tree_view/TreeViewOpen"
	if !view.IsExpanded(index) {
		name = "tree_view/TreeViewClose"
	}
	return name + "_" + common.GetIconColor(common.ThemeAuto, false)
}

// drawChevron draws the themed expander chevron of a branch: the visible arrow
// fits in a ChevronSize square, and it is centered in the row and in the
// indentation of its level (or nothing for a leaf).
func (self *FileTree) drawChevron(painter *qt.QPainter, rect *qt.QRect, index *qt.QModelIndex) {
	if index == nil || !index.IsValid() {
		return
	}
	if !hasBranch(index.Model(), index) {
		return
	}
	key := chevronKey(self.QTreeView, index)
	svgBytes := treeBranchSVG(self.QTreeView, index)
	if len(svgBytes) == 0 {
		return
	}

	scale, ink := self.chevronInk(svgBytes, key)
	if scale <= 0 {
		return
	}
	// Scale the whole view box so that the visible arrow lands centered on the
	// branch column.
	level := treeBranchLevel(index)
	x := self.chevronCenterX(level) - (ink.X0+ink.X1)/2*scale
	y := self.chevronCenterY(rect) - (ink.Y0+ink.Y1)/2*scale
	box := qt.NewQRectF4(x, y, scale, scale)
	defer box.Delete()
	drawSvgBytes(svgBytes, painter, box)
}

// filterViewportEvent answers the viewport mouse events of the expander chevron:
// clicking it toggles the item and must not change the selection, so the press
// and its release are consumed. Everything else is passed through untouched, so
// Qt keeps handling hover, selection and the double click expansion.
func (self *FileTree) filterViewportEvent(super func(watched *qt.QObject, event *qt.QEvent) bool, watched *qt.QObject, event *qt.QEvent) bool {
	switch event.Type() {
	case qt.QEvent__MouseButtonPress:
		if self.consumeExpanderClick(event) {
			return true
		}
	case qt.QEvent__MouseButtonRelease:
		if self.pressConsumed {
			self.pressConsumed = false
			return true
		}
	}
	return super(watched, event)
}

// consumeExpanderClick reports whether a viewport press hit the expander
// chevron; it toggles the item and consumes the event so the press never
// selects the item (Windows Explorer behavior).
func (self *FileTree) consumeExpanderClick(event *qt.QEvent) bool {
	me := qt.UnsafeNewQMouseEvent(event.UnsafePointer())
	pos := me.Pos()
	index := self.IndexAt(pos)
	if index == nil || !index.IsValid() || !self.isExpanderPos(pos, index) {
		return false
	}
	self.pressConsumed = true
	if self.IsExpanded(index) {
		self.Collapse(index)
	} else {
		self.Expand(index)
	}
	return true
}

// isExpanderPos reports whether a viewport position hits the chevron of the given
// item, matching the geometry drawn by drawChevron (plus three pixels of slack,
// so the arrow stays easy to hit).
func (self *FileTree) isExpanderPos(pos *qt.QPoint, index *qt.QModelIndex) bool {
	if pos == nil || !hasBranch(index.Model(), index) {
		return false
	}
	key := chevronKey(self.QTreeView, index)
	svgBytes := treeBranchSVG(self.QTreeView, index)
	if len(svgBytes) == 0 {
		return false
	}
	scale, ink := self.chevronInk(svgBytes, key)
	if scale <= 0 {
		return false
	}
	cx := self.chevronCenterX(treeBranchLevel(index))
	halfWidth := ink.width() * scale / 2
	const slack = 3.0
	x := float64(pos.X())
	return x >= cx-halfWidth-slack && x < cx+halfWidth+slack
}
