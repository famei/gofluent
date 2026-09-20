package widgets

import (
	"math"
	"sort"

	"github.com/famei/gofluent/common"
	"github.com/mappu/miqt/qt"
)

// ---------------------------------------------------------------------------
// FileTable colors

// FileTableColors groups the row colors painted by FileTableItemDelegate for a
// single theme. The table itself paints a transparent background, so a color
// carrying an alpha channel blends with whatever is drawn behind the table —
// the dark palette relies on that for its accent tinted selection.
type FileTableColors struct {
	// Selected is the background of the selected rows while the table is in an
	// active window and holds the focus.
	Selected *qt.QColor
	// NonFocus is the background of the selected rows while the table is not
	// focused (Windows Explorer greys the selection in that case).
	NonFocus *qt.QColor
	// Hover is the background of the row under the mouse cursor.
	Hover *qt.QColor
	// Highlight is the background of the rows registered with SetHighlight.
	Highlight *qt.QColor
	// Border outlines every selected row.
	Border *qt.QColor
}

// LightFileTableColors returns the default light palette (the Windows Explorer
// light selection colors).
func LightFileTableColors() FileTableColors {
	return FileTableColors{
		Selected:  qt.NewQColor3(204, 232, 255),
		NonFocus:  qt.NewQColor3(217, 217, 217),
		Hover:     qt.NewQColor3(229, 243, 255),
		Highlight: qt.NewQColor3(229, 243, 255),
		Border:    qt.NewQColor3(0, 0, 0),
	}
}

// DarkFileTableColors returns the default dark palette. The selected row keeps
// the accent color at a low alpha, so it blends with the dark window behind the
// transparent table instead of glowing as a solid block.
func DarkFileTableColors() FileTableColors {
	return FileTableColors{
		Selected:  qt.NewQColor11(0, 120, 212, 100),
		NonFocus:  qt.NewQColor11(255, 255, 255, 30),
		Hover:     qt.NewQColor11(255, 255, 255, 21),
		Highlight: qt.NewQColor11(255, 255, 255, 30),
		Border:    qt.NewQColor11(255, 255, 255, 60),
	}
}

// cloneFileTableColors copies a palette; nil entries fall back to the defaults
// of the matching theme.
func cloneFileTableColors(colors FileTableColors, fallback FileTableColors) FileTableColors {
	clone := func(color, def *qt.QColor) *qt.QColor {
		if color == nil {
			return cloneColor(def)
		}
		return cloneColor(color)
	}
	return FileTableColors{
		Selected:  clone(colors.Selected, fallback.Selected),
		NonFocus:  clone(colors.NonFocus, fallback.NonFocus),
		Hover:     clone(colors.Hover, fallback.Hover),
		Highlight: clone(colors.Highlight, fallback.Highlight),
		Border:    clone(colors.Border, fallback.Border),
	}
}

// deleteFileTableColors releases the colors owned by a palette.
func deleteFileTableColors(colors FileTableColors) {
	for _, color := range []*qt.QColor{colors.Selected, colors.NonFocus, colors.Hover, colors.Highlight, colors.Border} {
		if color != nil {
			color.Delete()
		}
	}
}

// FileTableItemDelegate paints the rows of a FileTable: the icon column inset,
// the hover/highlight/selection backgrounds and the row separator.
type FileTableItemDelegate struct {
	*qt.QStyledItemDelegate
	parent *qt.QTableWidget

	borderPen *qt.QPen

	lightColors FileTableColors
	darkColors  FileTableColors

	MaxColumn  int
	hoverRow   int
	pressedRow int

	selectedRows map[int]struct{}
	highlight    map[int]struct{}
}

func NewFileTableItemDelegate(parent *qt.QTableWidget) *FileTableItemDelegate {
	a := &FileTableItemDelegate{
		parent:       parent,
		selectedRows: make(map[int]struct{}),
		highlight:    make(map[int]struct{}),
		lightColors:  LightFileTableColors(),
		darkColors:   DarkFileTableColors(),
		hoverRow:     -1,
		pressedRow:   -1,
	}
	a.Init_Gui(parent)
	return a
}

func (self *FileTableItemDelegate) Init_Gui(parent *qt.QTableWidget) {
	self.QStyledItemDelegate = qt.NewQStyledItemDelegate2(parent.QObject)
	self.borderPen = qt.NewQPen()
	self.borderPen.SetWidthF(0.6)
	self.borderPen.SetColor(self.lightColors.Border)
	self.OnPaint(self.Paint)
	self.OnDestroyed(func() {
		deleteFileTableColors(self.lightColors)
		deleteFileTableColors(self.darkColors)
	})
}

// ---------------------------------------------------------------------------
// Theme

// Colors returns the palette of the active theme (owned by the delegate).
func (self *FileTableItemDelegate) Colors() FileTableColors {
	if common.IsDarkTheme() {
		return self.darkColors
	}
	return self.lightColors
}

// LightColors returns the light palette (owned by the delegate).
func (self *FileTableItemDelegate) LightColors() FileTableColors { return self.lightColors }

// DarkColors returns the dark palette (owned by the delegate).
func (self *FileTableItemDelegate) DarkColors() FileTableColors { return self.darkColors }

// SetColors replaces both palettes; missing colors keep the defaults.
func (self *FileTableItemDelegate) SetColors(light, dark FileTableColors) {
	self.SetLightColors(light)
	self.SetDarkColors(dark)
}

// SetLightColors replaces the light palette; missing colors keep the defaults.
func (self *FileTableItemDelegate) SetLightColors(colors FileTableColors) {
	deleteFileTableColors(self.lightColors)
	self.lightColors = cloneFileTableColors(colors, LightFileTableColors())
	self.repaint()
}

// SetDarkColors replaces the dark palette; missing colors keep the defaults.
func (self *FileTableItemDelegate) SetDarkColors(colors FileTableColors) {
	deleteFileTableColors(self.darkColors)
	self.darkColors = cloneFileTableColors(colors, DarkFileTableColors())
	self.repaint()
}

// repaint repaints the viewport this delegate paints into.
func (self *FileTableItemDelegate) repaint() {
	if self.parent == nil {
		return
	}
	if vp := self.parent.Viewport(); vp != nil {
		vp.Update()
	}
}

func (self *FileTableItemDelegate) SetMaxColumn(i int) {
	self.MaxColumn = i
}

func (self *FileTableItemDelegate) SetSelectedRows(list map[int]struct{}) {
	if _, ok := list[self.pressedRow]; ok {
		self.pressedRow = -1
	}
	self.selectedRows = list
}

func (self *FileTableItemDelegate) SetNotSelectedRows() {
	self.selectedRows = make(map[int]struct{})
}

func (self *FileTableItemDelegate) SetHoverRow(row int) bool {
	if self.hoverRow == row {
		return false
	}
	self.hoverRow = row
	return true
}

func (self *FileTableItemDelegate) SetPressedRow(row int) {
	self.pressedRow = row
}

func (self *FileTableItemDelegate) Paint(super func(painter *qt.QPainter, option *qt.QStyleOptionViewItem, index *qt.QModelIndex), painter *qt.QPainter, option *qt.QStyleOptionViewItem, index *qt.QModelIndex) {
	colors := self.Colors()
	painter.Save()
	painter.SetPenWithStyle(qt.NoPen)
	painter.SetRenderHint2(qt.QPainter__Antialiasing, false) // 暂时关闭抗锯齿 过小的边框会导致颜色不正确
	painter.SetClipping(true)
	painter.SetClipRectWithQRect(option.Rect())
	//option.Rect().Adjust(0, 2, 0, -2)
	option.SetRect(*option.Rect().Adjusted(0, 2, 0, -2))
	row := index.Row()
	if _, ok := self.highlight[row]; ok {
		self.fillRow(painter, option.QStyleOption, index, colors.Highlight)
	}

	if self.hoverRow == row {
		self.fillRow(painter, option.QStyleOption, index, colors.Hover)
	}

	if _, ok := self.selectedRows[row]; ok {
		self.drawSelectedBackground(painter, option.QStyleOption, index, colors)
	}
	painter.Restore()
	super(painter, option, index)
}

// rowRect returns the rectangle of one cell. The first column keeps 16px free
// for the item icon and the inner edges are grown by one pixel, so the row
// separators of neighbouring cells join without a gap.
func (self *FileTableItemDelegate) rowRect(option *qt.QStyleOption, index *qt.QModelIndex) *qt.QRect {
	if index.Column() == 0 {
		return option.Rect().Adjusted(1+16, 0, 1, 0)
	} else if index.Column() == self.MaxColumn {
		return option.Rect().Adjusted(-1, 0, -1, 0)
	}
	return option.Rect().Adjusted(-1, 0, 1, 0)
}

// fillRow paints one cell of a row with a solid (possibly translucent) color.
func (self *FileTableItemDelegate) fillRow(painter *qt.QPainter, option *qt.QStyleOption, index *qt.QModelIndex, color *qt.QColor) {
	rect := self.rowRect(option, index)
	painter.FillRect6(rect, color)
	painter.DrawRectWithRect(rect)
}

// drawSelectedBackground paints the background of a selected cell: light blue
// while the file table owns the focus of an active window (Windows Explorer
// behavior), grey otherwise.
func (self *FileTableItemDelegate) drawSelectedBackground(painter *qt.QPainter, option *qt.QStyleOption, index *qt.QModelIndex, colors FileTableColors) {
	background := colors.NonFocus
	if self.isFocused() {
		background = colors.Selected
	}

	self.borderPen.SetColor(colors.Border)
	painter.SetPenWithPen(self.borderPen)
	rect := self.rowRect(option, index)
	painter.FillRect6(rect, background)
	painter.DrawRectWithRect(rect)
}

// isFocused reports whether the table owning this delegate is the focused
// widget of an active window.
func (self *FileTableItemDelegate) isFocused() bool {
	if self.parent == nil {
		return false
	}
	return self.parent.HasFocus() && self.parent.IsActiveWindow()
}

// SetHighlightColor sets the highlight color of both themes.
func (self *FileTableItemDelegate) SetHighlightColor(light, dark *qt.QColor) {
	if self.lightColors.Highlight != nil {
		self.lightColors.Highlight.Delete()
	}
	if self.darkColors.Highlight != nil {
		self.darkColors.Highlight.Delete()
	}
	self.lightColors.Highlight = cloneColor(light)
	self.darkColors.Highlight = cloneColor(dark)
	self.repaint()
}

func (self *FileTableItemDelegate) SetHighlight(highlight map[int]struct{}) {
	if highlight == nil {
		self.highlight = make(map[int]struct{})
		return
	}
	self.highlight = highlight
}

func (self *FileTableItemDelegate) GetHighlight() map[int]struct{} {
	return self.highlight
}

// ---------------------------------------------------------------------------
// FileTable

// Marquee (rubber band) constants.
const (
	// fileTableMarqueeMargin is the distance from the top/bottom viewport edge
	// at which a marquee drag starts scrolling the table.
	fileTableMarqueeMargin = 24
	// fileTableMarqueeScrollInterval is the auto scroll timer interval in ms.
	fileTableMarqueeScrollInterval = 30
	// fileTableMarqueeScrollMinStep is the slowest auto scroll speed, in pixels
	// per tick.
	fileTableMarqueeScrollMinStep = 3
	// fileTableMarqueeScrollMaxStep is the fastest auto scroll speed, in pixels
	// per tick.
	fileTableMarqueeScrollMaxStep = 18
	// fileTableMarqueeScrollDepth is the number of pixels the pointer has to
	// travel into the margin before the auto scroll speed increases by one pixel
	// per tick.
	fileTableMarqueeScrollDepth = 8
)

// FileTable is a Windows Explorer style file table: a QTableWidget whose rows
// are painted by FileTableItemDelegate.
//
// On top of the row painting it provides:
//
//   - live light/dark theme switching: the row palette, the item colors and the
//     stylesheet all follow common.SetTheme, so a theme change needs no widget
//     specific code;
//   - a marquee selection rubber band for drag selection. The band is a child of
//     the viewport (so Qt clips it to the visible area and it never covers the
//     headers) and its vertical edges are anchored to the row the drag started
//     on: scrolling the table grows the marquee over the rows that scroll past
//     instead of leaving it behind, and the selection follows. It can be
//     switched off at runtime with SetRubberBandEnabled, in which case the plain
//     Qt drag selection is used again. Dragging against the top or bottom edge
//     scrolls the table automatically, which is what makes marquee selection
//     usable in a directory with thousands of entries;
//   - Windows Explorer like drag semantics: dragging an unselected row (or the
//     empty area) draws the marquee, while dragging an already selected row is
//     reported to the handler installed with SetItemDragHandler, so the
//     application can implement the drag and drop itself. Without a handler the
//     press is left to Qt and the marquee takes the drag over.
type FileTable struct {
	*qt.QTableWidget
	TableItemDelegate *FileTableItemDelegate

	// scrollDelegate owns the fluent overlay scroll bars; the native scroll bars
	// stay hidden behind it, exactly like in TableWidget/TableView.
	scrollDelegate *SmoothScrollDelegate

	// viewportFilter receives the raw viewport mouse events that drive the
	// marquee. QAbstractItemView virtuals cannot be overridden safely here (see
	// installItemTracking in table_view.go) and the viewport is the widget that
	// actually receives the mouse events, so an event filter is the reliable
	// hook.
	viewportFilter *qt.QObject
	themeAlive     *widgetAlive

	rubberBand        *qt.QRubberBand
	rubberBandPen     *qt.QPen
	rubberBandEnabled bool
	autoScrollEnabled bool
	lightBandColor    *qt.QColor
	darkBandColor     *qt.QColor

	// marquee drag state, in viewport coordinates. The vertical anchor is
	// attached to the row it was pressed on (anchorRow/anchorOffset) instead of
	// the screen, so scrolling the table grows the marquee instead of leaving it
	// behind.
	marquee      bool
	dragOriginX  int
	dragOriginY  int
	dragX        int
	dragY        int
	anchorRow    int
	anchorOffset int
	dragAdditive bool
	dragBase     map[int]struct{}
	dragFirst    int
	dragLast     int
	dragRangeSet bool
	// pressConsumed records that the press was answered by the marquee, so the
	// matching release is swallowed as well and never reaches Qt as a click.
	pressConsumed bool
	// pressArmed records that a left press inside the viewport may turn into a
	// marquee. While it is set the moves are answered here (even below the drag
	// threshold), otherwise Qt extends the selection from its own drag anchor
	// and selects the rows between a previous marquee block and the clicked row.
	pressArmed bool

	// item drag state: a press on an already selected row arms a drag instead
	// of the marquee.
	itemDragHandler  ItemDragHandler
	itemDragArmed    bool
	itemDragActive   bool
	pendingClickRow  int
	pendingClickCtrl bool

	scrollTimer *qt.QTimer
	scrollStep  int
}

// ItemDragHandler is invoked when the user starts dragging an already selected
// row. rows holds the selected rows and x/y the cursor position in global
// (screen) coordinates, ready for QDrag.Exec. Registering a handler enables the
// Explorer like drag behavior: without one (and with QAbstractItemView drag
// support switched off) dragging a selected row falls back to the marquee.
type ItemDragHandler func(rows []int, x int, y int)

// SetItemDragHandler registers the drag handler (pass nil to disable it).
func (self *FileTable) SetItemDragHandler(handler ItemDragHandler) {
	self.itemDragHandler = handler
}

// GetItemDragHandler returns the registered drag handler (nil when disabled).
func (self *FileTable) GetItemDragHandler() ItemDragHandler { return self.itemDragHandler }

// fileTableLightQss is the light stylesheet of the file table (the header
// colors follow resources/qss/light/table_view.qss).
const fileTableLightQss = `
	QTableView {
		background: transparent;
		outline: none;
		border: none;
		selection-background-color: transparent;
		alternate-background-color: transparent;
	}

	QTableView::item {
		background: transparent;
		border: 0px;
		padding-left: 16px;
		/*padding-right: 16px;*/
		height: 20px;
		color: rgb(0, 0, 0);
	}

 	QTableWidget::item:selected {
        color: rgb(0, 0, 0);
		/*background: #CCE8FF;*/
		background: transparent;
    }

	QTableView::indicator {
		width: 18px;
		height: 18px;
		border-radius: 5px;
		border: none;
		background-color: transparent;
	}

	QHeaderView {
		background-color: transparent;
	}

	QHeaderView::section {
		background-color: transparent;
		color: rgb(96, 96, 96);
		padding-left: 5px;
		padding-right: 5px;
		border: 1px solid rgba(0, 0, 0, 15);
		font: 13px --FontFamilies;
	}

	QHeaderView::section:horizontal {
		border-left: none;
		height: 25px;
	}

	QHeaderView::section:horizontal:last {
		border-right: none;
	}

	QHeaderView::section:vertical {
		border-top: none;
	}

	QHeaderView::section:checked {
		background-color: transparent;
	}`

// fileTableDarkQss is the dark stylesheet of the file table.
const fileTableDarkQss = `
	QTableView {
		background: transparent;
		outline: none;
		border: none;
		selection-background-color: transparent;
		alternate-background-color: transparent;
	}

	QTableView::item {
		background: transparent;
		border: 0px;
		padding-left: 16px;
		/*padding-right: 16px;*/
		height: 20px;
		color: rgb(255, 255, 255);
	}

 	QTableWidget::item:selected {
        color: rgb(255, 255, 255);
		background: transparent;
    }

	QTableView::indicator {
		width: 18px;
		height: 18px;
		border-radius: 5px;
		border: none;
		background-color: transparent;
	}

	QHeaderView {
		background-color: transparent;
	}

	QHeaderView::section {
		background-color: transparent;
		color: rgb(203, 203, 203);
		padding-left: 5px;
		padding-right: 5px;
		border: 1px solid rgba(255, 255, 255, 21);
		font: 13px --FontFamilies;
	}

	QHeaderView::section:horizontal {
		border-left: none;
		height: 25px;
	}

	QHeaderView::section:horizontal:last {
		border-right: none;
	}

	QHeaderView::section:vertical {
		border-top: none;
	}

	QHeaderView::section:checked {
		background-color: transparent;
	}`

func NewFileTable(parent *qt.QWidget) *FileTable {
	a := &FileTable{}
	a.Init_Gui(parent)
	return a
}

func (self *FileTable) Init_Gui(parent *qt.QWidget) {
	self.QTableWidget = qt.NewQTableWidget(parent)
	self.SetShowGrid(false)
	self.SetMouseTracking(true) // 启用鼠标追踪
	self.SetWordWrap(false)
	self.ResizeColumnsToContents()
	self.SetEditTriggers(qt.QAbstractItemView__NoEditTriggers)
	self.SetContextMenuPolicy(qt.CustomContextMenu)
	self.SetSelectionBehavior(qt.QAbstractItemView__SelectRows)
	self.VerticalHeader().SetDefaultSectionSize(20)
	// A file listing has no row numbers.
	self.VerticalHeader().Hide()

	// The stylesheet follows the theme: common.SetCustomStyleSheet registers the
	// table, so common.SetTheme re-applies the matching QSS automatically.
	common.SetCustomStyleSheet(self.QWidget, fileTableLightQss, fileTableDarkQss)

	self.TableItemDelegate = NewFileTableItemDelegate(self.QTableWidget)
	self.SetItemDelegate(self.TableItemDelegate.QAbstractItemDelegate)
	// Fluent overlay scroll bars (the native ones are hidden by the delegate).
	self.scrollDelegate = NewSmoothScrollDelegate(self.QAbstractScrollArea, false)

	// The delegate paints the selected rows, so it has to follow every
	// selection change — including the programmatic ones (SelectRow,
	// SetRangeSelected, ClearSelection) that no mouse/key event reports.
	if sm := self.SelectionModel(); sm != nil {
		sm.OnSelectionChanged(func(selected, deselected *qt.QItemSelection) {
			self.updateSelectedRows()
		})
	}

	self.initRubberBand()
	self.initTheme()

	self.OnLeaveEvent(func(super func(event *qt.QEvent), event *qt.QEvent) {
		super(event)
		if self.TableItemDelegate.SetHoverRow(-1) {
			self.updateSelectedRows()
		}
	})
	self.OnEntered(func(index *qt.QModelIndex) {
		if self.TableItemDelegate.SetHoverRow(index.Row()) {
			self.updateSelectedRows()
		}
	}) //鼠标悬浮单元格
	self.OnPressed(func(index *qt.QModelIndex) {
		if self.SelectionMode() == qt.QAbstractItemView__NoSelection {
			return
		}
		self.TableItemDelegate.SetPressedRow(index.Row())
		self.updateSelectedRows()
	}) //鼠标单击单元格 在clicked之前

	self.OnResizeEvent(func(super func(event *qt.QResizeEvent), event *qt.QResizeEvent) {
		super(event)
		// Keep the marquee inside the viewport when the table is resized while
		// the user is still dragging.
		if self.marquee {
			self.refreshMarquee()
		}
		self.updateSelectedRows()
	})
	self.OnKeyPressEvent(func(super func(event *qt.QKeyEvent), event *qt.QKeyEvent) {
		super(event)
		self.updateSelectedRows()
	})

	self.OnSelectAll(func(super func()) {
		super()
		self.updateSelectedRows()
	})
	self.ClearSelection()

	self.OnFocusOutEvent(func(super func(event *qt.QFocusEvent), event *qt.QFocusEvent) {
		super(event)
		self.updateSelectedRows()
	})

	self.OnMouseReleaseEvent(func(super func(event *qt.QMouseEvent), event *qt.QMouseEvent) {
		defer super(event)
		self.updateSelectedRows()
	})

	// Scrolling moves the viewport content (and, without the band being a child
	// of the viewport, not the band): refresh the selection the fixed rectangle
	// covers. Both scrollbars are watched because horizontal scrolling shifts
	// the rows out from under the marquee as well.
	onScroll := func(_ int) {
		if self.marquee {
			self.refreshMarquee()
		}
		self.updateSelectedRows()
	}
	self.VerticalScrollBar().OnValueChanged(onScroll)
	self.HorizontalScrollBar().OnValueChanged(onScroll)
}

// initTheme repaints the rows when the application theme changes; the
// stylesheet itself is refreshed by the stylesheet manager.
func (self *FileTable) initTheme() {
	self.themeAlive = trackWidget(self.OnDestroyed)
	common.QConfigInstance.OnThemeChanged(func(common.Theme) {
		if !self.themeAlive.ok() {
			return
		}
		self.applyRubberBandStyle()
		if vp := self.Viewport(); vp != nil {
			vp.Update()
		}
		if hh := self.HorizontalHeader(); hh != nil {
			hh.QWidget.Update()
		}
		if vh := self.VerticalHeader(); vh != nil {
			vh.QWidget.Update()
		}
	})
}

// ---------------------------------------------------------------------------
// Marquee (rubber band) selection

func (self *FileTable) initRubberBand() {
	self.rubberBandEnabled = true
	self.autoScrollEnabled = true
	self.dragFirst, self.dragLast = -1, -1
	self.pendingClickRow = -1

	// The band is a child of the viewport, so Qt clips it to the viewport for
	// free: it can grow past the visible area while the table scrolls without
	// ever painting over the header or the scrollbars. Scrolling a viewport also
	// moves its child widgets, which is why refreshMarquee re-anchors the band
	// (see the paint event filter in filterViewportEvent).
	self.rubberBand = qt.NewQRubberBand2(qt.QRubberBand__Rectangle, self.Viewport())
	self.rubberBand.SetAttribute(qt.WA_TransparentForMouseEvents)
	// The band paints only its border and a translucent fill: without this the
	// widget would erase the rows behind it.
	self.rubberBand.SetAttribute(qt.WA_NoSystemBackground)
	self.rubberBand.Hide()

	// A visible marquee in both themes: the style default is a dotted frame that
	// is barely visible (and disappears on a dark background), so the band is
	// painted here instead (see paintRubberBand).
	self.rubberBandPen = qt.NewQPen()
	self.rubberBandPen.SetWidthF(1)
	self.lightBandColor = qt.NewQColor3(0, 120, 212)
	self.darkBandColor = qt.NewQColor3(76, 194, 255)
	self.applyRubberBandStyle()
	self.rubberBand.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
		self.paintRubberBand()
	})

	self.scrollTimer = qt.NewQTimer2(self.QObject)
	self.scrollTimer.SetInterval(fileTableMarqueeScrollInterval)
	self.scrollTimer.OnTimeout(func() { self.autoScrollTick() })

	self.viewportFilter = qt.NewQObject2(self.QObject)
	self.Viewport().InstallEventFilter(self.viewportFilter)
	self.viewportFilter.OnEventFilter(func(super func(watched *qt.QObject, event *qt.QEvent) bool, watched *qt.QObject, event *qt.QEvent) bool {
		return self.filterViewportEvent(super, watched, event)
	})
}

// SetRubberBandEnabled enables or disables the marquee selection. With the
// marquee disabled the viewport events are forwarded untouched, so the table
// falls back to the plain Qt drag selection.
func (self *FileTable) SetRubberBandEnabled(enabled bool) {
	if self.rubberBandEnabled == enabled {
		return
	}
	self.rubberBandEnabled = enabled
	if !enabled {
		self.endMarquee()
	}
}

// IsRubberBandEnabled reports whether the marquee selection is enabled.
func (self *FileTable) IsRubberBandEnabled() bool { return self.rubberBandEnabled }

// SetRubberBandAutoScrollEnabled enables or disables the automatic scrolling
// performed while a marquee drag is held against the top/bottom viewport edge.
func (self *FileTable) SetRubberBandAutoScrollEnabled(enabled bool) {
	self.autoScrollEnabled = enabled
	if !enabled {
		self.stopAutoScroll()
	}
}

// IsRubberBandAutoScrollEnabled reports whether marquee auto scrolling is on.
func (self *FileTable) IsRubberBandAutoScrollEnabled() bool { return self.autoScrollEnabled }

// SetRubberBandColors sets the marquee border color of the light and dark
// theme; the fill uses the same color at a low alpha. An invalid color restores
// the default.
func (self *FileTable) SetRubberBandColors(light, dark *qt.QColor) {
	if self.lightBandColor != nil {
		self.lightBandColor.Delete()
	}
	if self.darkBandColor != nil {
		self.darkBandColor.Delete()
	}
	self.lightBandColor = cloneColor(light)
	self.darkBandColor = cloneColor(dark)
	self.applyRubberBandStyle()
}

// applyRubberBandStyle applies the marquee colors of the active theme: the
// border uses the theme color, the fill the same color at a low alpha.
func (self *FileTable) applyRubberBandStyle() {
	if self.rubberBandPen == nil {
		return
	}
	color := self.lightBandColor
	if common.IsDarkTheme() {
		color = self.darkBandColor
	}
	if color == nil || !color.IsValid() {
		return
	}
	self.rubberBandPen.SetColor(color)
	if self.rubberBand != nil {
		self.rubberBand.Update()
	}
}

// paintRubberBand paints the marquee: a solid theme colored border around a
// translucent fill. QRubberBand's own painting uses the style's dotted frame,
// which is invisible on dark backgrounds and leaves gaps that look like a
// missing border.
//
// The frame is painted as four solid strips of whole device pixels on top of a
// translucent fill. A stroked logical one pixel pen lands on a half device pixel
// on 125%/150% screens, where the rasterizer rounds it now to one and now to two
// device pixels, which is what made the border look sometimes thick, sometimes
// thin and sometimes missing.
func (self *FileTable) paintRubberBand() {
	if self.rubberBand == nil {
		return
	}
	band := self.rubberBand
	width, height := float64(band.Width()), float64(band.Height())
	if width <= 0 || height <= 0 {
		return
	}

	painter := qt.NewQPainter2(band.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHint2(qt.QPainter__Antialiasing, false)
	painter.SetPenWithStyle(qt.NoPen)

	ratio := band.DevicePixelRatioF()
	if ratio <= 0 {
		ratio = 1
	}
	// Border width of whole device pixels, expressed in logical units.
	thickness := math.Max(1, math.Round(ratio)) / ratio

	frame := qt.NewQRectF4(0, 0, width, height)

	// The translucent fill goes first: it has to blend with the rows behind the
	// band, not with an opaque frame.
	fill := cloneColor(self.rubberBandPen.Color())
	fill.SetAlpha(38)
	painter.FillRect4(frame, fill)
	fill.Delete()

	// The frame: top, bottom, left and right strips.
	border := qt.NewQColor9(self.rubberBandPen.Color())
	frame.SetRect(0, 0, width, thickness)
	painter.FillRect4(frame, border)
	frame.SetRect(0, height-thickness, width, thickness)
	painter.FillRect4(frame, border)
	frame.SetRect(0, 0, thickness, height)
	painter.FillRect4(frame, border)
	frame.SetRect(width-thickness, 0, thickness, height)
	painter.FillRect4(frame, border)
	border.Delete()
	frame.Delete()
	painter.End()
}

// deviceStep returns the smallest number of logical pixels whose device size is
// a whole number of device pixels (2 on a 150% screen, 4 on 125%, 1 on 100% and
// 200%). Snapping the marquee to it puts its edges on the device pixel grid, so
// the one pixel frame is drawn with a constant thickness instead of being
// rounded to one device pixel on one side and two on the other.
func (self *FileTable) deviceStep() int {
	ratio := 1.0
	if viewport := self.Viewport(); viewport != nil {
		ratio = viewport.DevicePixelRatioF()
	}
	if ratio <= 0 {
		return 1
	}
	for step := 1; step <= 64; step++ {
		scaled := float64(step) * ratio
		if math.Abs(scaled-math.Round(scaled)) < 0.01 {
			return step
		}
	}
	return 1
}

// RubberBand returns the marquee widget (owned by the viewport).
func (self *FileTable) RubberBand() *qt.QRubberBand { return self.rubberBand }

// RubberBandRect returns the marquee rectangle in viewport coordinates; it is
// empty while no marquee drag is running. While the table is scrolled the
// rectangle can reach above/below the viewport (the band is clipped by the
// viewport, the rectangle is not). The caller owns the rectangle.
func (self *FileTable) RubberBandRect() *qt.QRect {
	if !self.marquee {
		return qt.NewQRect()
	}
	return self.marqueeRect()
}

// filterViewportEvent drives the marquee from the raw viewport mouse events.
func (self *FileTable) filterViewportEvent(super func(watched *qt.QObject, event *qt.QEvent) bool, watched *qt.QObject, event *qt.QEvent) bool {
	switch event.Type() {
	case qt.QEvent__MouseButtonPress:
		if !self.rubberBandEnabled {
			return super(watched, event)
		}
		e := qt.UnsafeNewQMouseEvent(event.UnsafePointer())
		if e.Button() != qt.LeftButton {
			return super(watched, event)
		}
		pos := e.Pos()
		self.beginMarquee(pos.X(), pos.Y(), self.additiveDrag(e.Modifiers()))
		if self.marqueePressOnItem(pos) {
			// Explorer semantics: dragging an already selected row drags the
			// selection (the application implements the drop through
			// SetItemDragHandler), dragging any other row draws the marquee.
			// Without a handler (and without Qt's own drag support) the press is
			// left to Qt, so the marquee keeps working everywhere.
			row := self.RowAt(pos.Y())
			itemDragAvailable := self.itemDragHandler != nil || self.DragEnabled()
			if row >= 0 && itemDragAvailable && self.isRowSelected(row) && e.Modifiers()&qt.ShiftModifier == 0 {
				self.armItemDrag(row, e.Modifiers()&qt.ControlModifier != 0)
				return true
			}
			// A press on an unselected row keeps Qt's normal row selection (and
			// its clicked signal); the marquee only takes the drag over once
			// the mouse travels further than the drag threshold.
			return super(watched, event)
		}
		// Qt would enter its own drag selecting state (and show its own rubber
		// band) for a press on an empty cell. The marquee replaces that, so the
		// press is answered here instead of forwarding it. An additive press
		// (Ctrl/Shift) keeps the current selection, like Windows Explorer.
		self.pressConsumed = true
		if !self.dragAdditive {
			self.ClearSelection()
		}
		self.SetFocus()
		self.updateSelectedRows()
		return true

	case qt.QEvent__MouseMove:
		if !self.rubberBandEnabled {
			return super(watched, event)
		}
		e := qt.UnsafeNewQMouseEvent(event.UnsafePointer())
		if e.Buttons()&qt.LeftButton == 0 {
			return super(watched, event)
		}
		if self.itemDragArmed || self.itemDragActive {
			pos, global := e.Pos(), e.GlobalPos()
			if self.updateItemDrag(pos.X(), pos.Y(), global.X(), global.Y(), self.additiveDrag(e.Modifiers())) {
				return true
			}
		}
		if self.pressArmed {
			// Also consume the moves below the drag threshold: Qt would turn
			// them into its own drag selection, extending the selection from the
			// pressed row (and thereby selecting everything between that row and
			// the previous marquee block).
			self.updateMarquee(e.Pos().X(), e.Pos().Y(), self.additiveDrag(e.Modifiers()))
			return true
		}
		if self.updateMarquee(e.Pos().X(), e.Pos().Y(), self.additiveDrag(e.Modifiers())) {
			// The marquee owns the drag: Qt must not extend the row selection
			// (nor start its own rubber band) while it is running.
			return true
		}
		return super(watched, event)

	case qt.QEvent__MouseButtonRelease:
		self.pressArmed = false
		if self.itemDragActive {
			self.itemDragActive = false
			return true
		}
		if self.itemDragArmed {
			self.applyPendingClick()
			return true
		}
		if self.marquee {
			self.endMarquee()
			return true
		}
		if self.pressConsumed {
			self.pressConsumed = false
			return true
		}
		return super(watched, event)

	case qt.QEvent__Hide:
		if self.marquee {
			self.endMarquee()
		}

	case qt.QEvent__Leave:
		// The implicit mouse grab keeps delivering the moves to the viewport
		// while the button is held, so a leave must not end the drag: the band
		// simply stops at the viewport edge.
		if self.marquee {
			self.refreshMarquee()
		}

	case qt.QEvent__Resize:
		if self.marquee {
			self.refreshMarquee()
		}

	case qt.QEvent__Paint:
		// Scrolling a viewport also moves its child widgets, which would drag
		// the band away from the drag rectangle. Re-anchoring it right before the
		// viewport paints guarantees the band is always drawn where the
		// rectangle is, whatever moved it in between.
		if self.marquee {
			self.refreshMarquee()
		}

	case qt.QEvent__Wheel:
		self.stopAutoScroll()
	}
	return super(watched, event)
}

// beginMarquee records the press position of a possible marquee drag and
// anchors it to the row under the cursor, so that a scroll (auto scroll or the
// wheel) grows the marquee instead of moving it away from its content.
func (self *FileTable) beginMarquee(x, y int, additive bool) {
	self.dragOriginX, self.dragOriginY = x, y
	self.dragX, self.dragY = x, y
	self.dragAdditive = additive
	self.dragBase = nil
	self.dragRangeSet = false
	self.dragFirst, self.dragLast = -1, -1
	self.marquee = false
	self.pressConsumed = false
	self.pressArmed = true
	self.anchorRow = self.RowAt(y)
	if self.anchorRow < 0 {
		// Above the first / below the last row: keep the anchor relative to the
		// edge row so it still follows the scrolling content.
		if self.RowCount() > 0 {
			self.anchorRow = self.RowCount() - 1
		}
	}
	if self.anchorRow >= 0 {
		self.anchorOffset = y - self.RowViewportPosition(self.anchorRow)
	} else {
		self.anchorOffset = y
	}
	self.stopAutoScroll()
}

// anchorY returns the current viewport y of the drag anchor. It is derived from
// the anchored row, so scrolling the content moves the anchor (and therefore
// grows or shrinks the marquee) with it.
func (self *FileTable) anchorY() int {
	if self.anchorRow < 0 {
		return self.anchorOffset
	}
	return self.RowViewportPosition(self.anchorRow) + self.anchorOffset
}

// updateMarquee advances a drag; it reports whether the marquee is drawing, in
// which case the event must not reach Qt.
func (self *FileTable) updateMarquee(x, y int, additive bool) bool {
	self.dragX, self.dragY = x, y
	if !self.marquee {
		dx, dy := x-self.dragOriginX, y-self.dragOriginY
		if dx < 0 {
			dx = -dx
		}
		if dy < 0 {
			dy = -dy
		}
		if dx+dy < qt.QApplication_StartDragDistance() {
			return false
		}
		// The drag passed the threshold: take the selection over, keeping the
		// selection of the moment the drag started when Ctrl/Shift is held.
		self.marquee = true
		self.dragAdditive = additive
		self.dragBase = self.selectedRowSet()
		self.dragRangeSet = false
		self.rubberBand.Show()
		self.rubberBand.Raise()
	}
	self.refreshMarquee()
	return true
}

// endMarquee finishes a marquee drag, keeping the selection it produced.
func (self *FileTable) endMarquee() {
	if self.marquee {
		self.marquee = false
		self.rubberBand.Hide()
	}
	self.stopAutoScroll()
	self.pressArmed = false
	self.pressConsumed = false
	self.dragBase = nil
	self.dragRangeSet = false
	self.updateSelectedRows()
}

// refreshMarquee repositions the band and refreshes the selection it covers.
func (self *FileTable) refreshMarquee() {
	if !self.marquee || self.rubberBand == nil {
		return
	}
	rect := self.marqueeRect() // viewport coordinates, caller owned
	self.applyMarqueeSelection(rect)
	self.rubberBand.SetGeometry(rect)
	rect.Delete()
	// Keep the band above the viewport content it overlays.
	self.rubberBand.Raise()
	self.updateAutoScroll()
}

// marqueeRect returns the drag rectangle in viewport coordinates: the
// horizontal edges follow the mouse (clamped to the viewport, so the band never
// covers the vertical header or the scrollbar), while the vertical edges follow
// the anchored row and the mouse — the rectangle therefore grows over the
// content that scrolls past it. The rectangle is snapped to the device pixel
// grid (see deviceStep) so its frame stays crisp, and the caller owns it.
func (self *FileTable) marqueeRect() *qt.QRect {
	x1, x2 := self.dragOriginX, self.dragX
	y1, y2 := self.anchorY(), self.dragY
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	if viewport := self.Viewport(); viewport != nil {
		if x1 < 0 {
			x1 = 0
		}
		if x2 > viewport.Width()-1 {
			x2 = viewport.Width() - 1
		}
	}

	step := self.deviceStep()
	snap := func(value int) int {
		return int(math.Round(float64(value)/float64(step))) * step
	}
	left, top := snap(x1), snap(y1)
	right, bottom := snap(x2+1), snap(y2+1)
	if right <= left {
		right = left + step
	}
	if bottom <= top {
		bottom = top + step
	}
	return qt.NewQRect4(left, top, right-left, bottom-top)
}

// applyMarqueeSelection selects every row touched by the marquee rectangle. The
// rows are recomputed only when the touched range actually changes, so a drag
// over a large directory costs one selection update per row instead of one per
// mouse move.
func (self *FileTable) applyMarqueeSelection(rect *qt.QRect) {
	first, last := self.marqueeRows(rect)
	if first < 0 || last < first {
		// The rectangle no longer touches any item (it moved into the blank
		// area right of the last column): drop the selection it produced.
		if self.dragRangeSet {
			self.dragRangeSet = false
			self.dragFirst, self.dragLast = -1, -1
			self.clearMarqueeSelection()
		}
		return
	}
	if self.dragRangeSet && self.dragFirst == first && self.dragLast == last {
		return
	}
	self.dragFirst, self.dragLast, self.dragRangeSet = first, last, true
	self.clearMarqueeSelection()
	self.selectRowRange(first, last)
	self.updateSelectedRows()
}

// clearMarqueeSelection drops the selection contributed by the marquee and
// restores the selection an additive (Ctrl/Shift) drag started with.
func (self *FileTable) clearMarqueeSelection() {
	self.ClearSelection()
	if self.dragAdditive {
		for row := range self.dragBase {
			self.selectRow(row)
		}
	}
}

// marqueeRows returns the first and the last row touched by the marquee
// rectangle. The rectangle can reach above or below the viewport while the
// table scrolls, so the rows hidden by the edges are walked back from the
// visible ones.
func (self *FileTable) marqueeRows(rect *qt.QRect) (int, int) {
	if self.RowCount() == 0 || !self.marqueeTouchesItems(rect) {
		return -1, -1
	}
	first := self.RowAt(rect.Top())
	if rect.Top() < 0 {
		row := self.RowAt(0)
		if row < 0 {
			row = 0
		}
		rowHeight := self.RowHeight(row)
		if rowHeight <= 0 {
			rowHeight = 1
		}
		hidden := (-rect.Top() + rowHeight - 1) / rowHeight
		first = row - hidden
	}
	last := self.RowAt(rect.Bottom())
	// RowAt returns -1 above the first and below the last row.
	if first < 0 {
		first = 0
	}
	if last < 0 {
		last = self.RowCount() - 1
	}
	if first >= self.RowCount() {
		first = self.RowCount() - 1
	}
	if last >= self.RowCount() {
		last = self.RowCount() - 1
	}
	if last < first {
		return -1, -1
	}
	return first, last
}

// marqueeTouchesItems reports whether the marquee rectangle overlaps the item
// area horizontally. Windows Explorer selects nothing while the marquee is
// drawn in the blank space right of the last column.
func (self *FileTable) marqueeTouchesItems(rect *qt.QRect) bool {
	last := self.ColumnCount() - 1
	if last < 0 {
		return false
	}
	right := self.ColumnViewportPosition(last) + self.ColumnWidth(last)
	return rect.Left() < right && rect.Right() >= 0
}

// selectRow selects one whole row.
func (self *FileTable) selectRow(row int) {
	self.selectRowRange(row, row)
}

// selectRowRange selects the rows in [first, last].
func (self *FileTable) selectRowRange(first, last int) {
	self.setRowRangeSelected(first, last, true)
}

// setRowRangeSelected selects or deselects the rows in [first, last].
func (self *FileTable) setRowRangeSelected(first, last int, selected bool) {
	if self.ColumnCount() == 0 || first < 0 || last < first {
		return
	}
	selectionRange := qt.NewQTableWidgetSelectionRange2(first, 0, last, self.ColumnCount()-1)
	self.SetRangeSelected(selectionRange, selected)
	selectionRange.Delete()
}

// selectedRowSet snapshots the currently selected rows.
func (self *FileTable) selectedRowSet() map[int]struct{} {
	rows := make(map[int]struct{})
	for _, index := range self.SelectedItems() {
		rows[index.Row()] = struct{}{}
	}
	return rows
}

// additiveDrag reports whether the modifiers make the marquee keep the current
// selection.
func (self *FileTable) additiveDrag(modifiers qt.KeyboardModifier) bool {
	return modifiers&(qt.ControlModifier|qt.ShiftModifier) != 0
}

// marqueePressOnItem reports whether the press landed on an item instead of the
// empty area of the viewport.
func (self *FileTable) marqueePressOnItem(pos *qt.QPoint) bool {
	if pos == nil {
		return false
	}
	index := self.IndexAt(pos)
	return index != nil && index.IsValid()
}

// ---------------------------------------------------------------------------
// Item drag (dragging the selected rows)

// isRowSelected reports whether row belongs to the current selection.
func (self *FileTable) isRowSelected(row int) bool {
	if model := self.SelectionModel(); model != nil {
		return model.IsRowSelected(row)
	}
	_, ok := self.selectedRowSet()[row]
	return ok
}

// selectedRows returns the selected rows in ascending order.
func (self *FileTable) selectedRows() []int {
	rows := make([]int, 0)
	for row := range self.selectedRowSet() {
		rows = append(rows, row)
	}
	sort.Ints(rows)
	return rows
}

// armItemDrag answers a press on an already selected row: the selection is kept
// (Windows Explorer does not collapse it before the mouse is released) and the
// press is remembered so a click without a drag can be completed on release.
func (self *FileTable) armItemDrag(row int, ctrl bool) {
	self.itemDragArmed = true
	self.itemDragActive = false
	self.pendingClickRow = row
	self.pendingClickCtrl = ctrl
	self.pressConsumed = true
	self.SetFocus()
}

// updateItemDrag reports whether the drag belongs to the item drag; the handler
// registered with SetItemDragHandler is called once the drag passes the
// threshold.
func (self *FileTable) updateItemDrag(x, y, globalX, globalY int, additive bool) bool {
	if self.itemDragActive {
		return true
	}
	if !self.itemDragArmed {
		return false
	}
	dx, dy := x-self.dragOriginX, y-self.dragOriginY
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	if dx+dy < qt.QApplication_StartDragDistance() {
		return true
	}
	self.itemDragArmed = false
	self.itemDragActive = true
	self.pendingClickRow = -1
	if self.itemDragHandler != nil {
		self.itemDragHandler(self.selectedRows(), globalX, globalY)
	}
	return true
}

// applyPendingClick completes a press on a selected row that was released
// without dragging: without Ctrl the selection collapses to that row, with Ctrl
// the row leaves the selection (Windows Explorer behavior).
func (self *FileTable) applyPendingClick() {
	row := self.pendingClickRow
	ctrl := self.pendingClickCtrl
	self.itemDragArmed = false
	self.itemDragActive = false
	self.pendingClickRow = -1
	self.pendingClickCtrl = false
	self.pressConsumed = false
	self.pressArmed = false
	if row < 0 {
		return
	}
	if ctrl {
		self.setRowRangeSelected(row, row, false)
	} else {
		self.ClearSelection()
		self.selectRow(row)
	}
	self.updateSelectedRows()
}

// ---------------------------------------------------------------------------
// Marquee auto scrolling

// updateAutoScroll starts or stops the auto scroll timer according to the
// current drag position.
func (self *FileTable) updateAutoScroll() {
	if !self.autoScrollEnabled || !self.marquee {
		self.stopAutoScroll()
		return
	}
	viewport := self.Viewport()
	bar := self.VerticalScrollBar()
	if viewport == nil || bar == nil || bar.Maximum() <= bar.Minimum() {
		self.stopAutoScroll()
		return
	}
	switch {
	case self.dragY < fileTableMarqueeMargin:
		self.scrollStep = -self.autoScrollStep()
	case self.dragY > viewport.Height()-fileTableMarqueeMargin:
		self.scrollStep = self.autoScrollStep()
	default:
		self.stopAutoScroll()
		return
	}
	if !self.scrollTimer.IsActive() {
		self.scrollTimer.Start2()
	}
}

// autoScrollStep returns how many scroll bar units one auto scroll tick moves:
// the speed grows with the distance the pointer travelled into the margin (or
// beyond the viewport), which is what makes dragging through a long directory
// usable. The fluent overlay scroll bars scroll per pixel while a plain view
// scrolls per item, so the speed is computed in pixels and converted.
func (self *FileTable) autoScrollStep() int {
	viewport := self.Viewport()
	if viewport == nil {
		return 1
	}
	depth := 0
	if self.dragY < fileTableMarqueeMargin {
		depth = fileTableMarqueeMargin - self.dragY
	} else {
		depth = self.dragY - (viewport.Height() - fileTableMarqueeMargin)
	}
	pixels := fileTableMarqueeScrollMinStep + depth/fileTableMarqueeScrollDepth
	if pixels > fileTableMarqueeScrollMaxStep {
		pixels = fileTableMarqueeScrollMaxStep
	}
	if pixels < 1 {
		pixels = 1
	}
	if self.VerticalScrollMode() == qt.QAbstractItemView__ScrollPerPixel {
		return pixels
	}

	row := self.RowAt(self.dragY)
	if row < 0 {
		row = self.RowAt(0)
	}
	rowHeight := 20
	if row >= 0 {
		if height := self.RowHeight(row); height > 0 {
			rowHeight = height
		}
	}
	items := pixels / rowHeight
	if items < 1 {
		items = 1
	}
	return items
}

func (self *FileTable) stopAutoScroll() {
	if self.scrollTimer != nil && self.scrollTimer.IsActive() {
		self.scrollTimer.Stop()
	}
}

// autoScrollTick scrolls the table by one step and refreshes the marquee: the
// band keeps its rectangle while the rows move underneath it.
func (self *FileTable) autoScrollTick() {
	if !self.marquee {
		self.stopAutoScroll()
		return
	}
	bar := self.VerticalScrollBar()
	if bar == nil {
		self.stopAutoScroll()
		return
	}
	value := bar.Value() + self.scrollStep
	if value < bar.Minimum() {
		value = bar.Minimum()
	}
	if value > bar.Maximum() {
		value = bar.Maximum()
	}
	if value == bar.Value() {
		self.stopAutoScroll()
		return
	}
	bar.SetValue(value)
	// The scroll moved the band widget along with the rows; re-anchor it (and
	// refresh the rows the rectangle currently covers) after the scroll.
	self.refreshMarquee()
	self.updateAutoScroll()
}

func (self *FileTable) updateSelectedRows() {
	nowData := make(map[int]struct{})
	for _, index := range self.SelectedItems() {
		nowData[index.Row()] = struct{}{}
	}
	self.TableItemDelegate.SetSelectedRows(nowData)
	self.QTableView.Viewport().Update()
}

func (self *FileTable) SetMaxColumn(i int) {
	self.TableItemDelegate.SetMaxColumn(i)
}

// SetHighlight marks rows with the highlight background (a search result, a
// drop target, ...). Pass nil to clear the highlight.
func (self *FileTable) SetHighlight(rows map[int]struct{}) {
	self.TableItemDelegate.SetHighlight(rows)
	if vp := self.Viewport(); vp != nil {
		vp.Update()
	}
}

// Highlight returns the highlighted rows (owned by the delegate).
func (self *FileTable) Highlight() map[int]struct{} {
	return self.TableItemDelegate.GetHighlight()
}

// SetHighlightColor sets the highlight color of both themes.
func (self *FileTable) SetHighlightColor(light, dark *qt.QColor) {
	self.TableItemDelegate.SetHighlightColor(light, dark)
}
