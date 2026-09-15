package widgets

import (
	"unsafe"

	"github.com/famei/gofluent/common"

	qt "github.com/mappu/miqt/qt"
)

// MenuAnimationType enumerates the pop-up animation styles of a menu.
type MenuAnimationType int

const (
	MenuAnimationNone MenuAnimationType = iota
	MenuAnimationDropDown
	MenuAnimationPullUp
	MenuAnimationFadeInDropDown
	MenuAnimationFadeInPullUp
)

// MenuIndicatorType enumerates checkable menu item indicator styles.
type MenuIndicatorType int

const (
	MenuIndicatorCheck MenuIndicatorType = iota
	MenuIndicatorRadio
)

// menuItemKind classifies a row of the internal list view.
type menuItemKind int

const (
	menuItemAction menuItemKind = iota
	menuItemSubMenu
	menuItemSeparator
	menuItemWidget
)

// menuItemData holds the per-row metadata used by the item delegate and the
// click/enter handlers. It lives on the Go side (keyed by the QListWidgetItem)
// because miqt QVariant cannot round-trip a QAction/QObject pointer.
type menuItemData struct {
	kind    menuItemKind
	action  *qt.QAction
	submenu *RoundMenu
	widget  *qt.QWidget
}

// MenuItemDelegate is the fluent menu item delegate. It lets the native
// QStyledItemDelegate paint the QSS-styled background/icon/text, then paints the
// fluent extras on top: separators, shortcut keys, the sub-menu chevron and (for
// checkable menus) the check/radio indicator.
type MenuItemDelegate struct {
	*qt.QStyledItemDelegate
	menu           *RoundMenu
	checkIndicator bool
	indicatorType  MenuIndicatorType
}

// NewMenuItemDelegate builds the base (shortcut) menu item delegate. The
// delegate is parented to the menu's list view so its C++ object follows the
// view lifetime; the RoundMenu also keeps a Go reference in its delegate field.
func NewMenuItemDelegate(menu *RoundMenu) *MenuItemDelegate {
	d := &MenuItemDelegate{
		QStyledItemDelegate: qt.NewQStyledItemDelegate2(menu.view.QObject),
		menu:                menu,
	}
	d.installPaint()
	return d
}

// NewCheckableMenuItemDelegate builds the checkable menu item delegate.
func NewCheckableMenuItemDelegate(menu *RoundMenu, indicatorType MenuIndicatorType) *MenuItemDelegate {
	d := NewMenuItemDelegate(menu)
	d.checkIndicator = true
	d.indicatorType = indicatorType
	return d
}

func (d *MenuItemDelegate) installPaint() {
	d.OnPaint(func(super func(painter *qt.QPainter, option *qt.QStyleOptionViewItem, index *qt.QModelIndex), painter *qt.QPainter, option *qt.QStyleOptionViewItem, index *qt.QModelIndex) {
		super(painter, option, index)
		d.paintExtra(painter, option, index)
	})
}

func (d *MenuItemDelegate) paintExtra(painter *qt.QPainter, option *qt.QStyleOptionViewItem, index *qt.QModelIndex) {
	info := d.itemInfo(index)
	if info == nil {
		return
	}

	switch info.kind {
	case menuItemSeparator:
		d.drawSeparator(painter, option)
	case menuItemAction:
		d.drawShortcut(painter, option, info.action)
		if d.checkIndicator && info.action.IsChecked() {
			d.drawIndicator(painter, option)
		}
	case menuItemSubMenu:
		d.drawChevron(painter, option, index)
	}
}

func (d *MenuItemDelegate) itemInfo(index *qt.QModelIndex) *menuItemData {
	if index == nil || d.menu == nil || d.menu.itemData == nil {
		return nil
	}
	item := d.menu.view.ItemFromIndex(index)
	if item == nil {
		return nil
	}
	return d.menu.itemData[item.UnsafePointer()]
}

func (d *MenuItemDelegate) drawSeparator(painter *qt.QPainter, option *qt.QStyleOptionViewItem) {
	painter.Save()
	var c *qt.QColor
	if common.IsDarkTheme() {
		c = qt.NewQColor11(255, 255, 255, 25)
	} else {
		c = qt.NewQColor11(0, 0, 0, 25)
	}
	defer c.Delete()
	pen := qt.NewQPen3(c)
	defer pen.Delete()
	pen.SetCosmetic(true)
	painter.SetPenWithPen(pen)

	rect := option.Rect() // GoGC-armed — do NOT Delete
	painter.DrawLine2(0, rect.Y()+4, rect.Width()+12, rect.Y()+4)
	painter.Restore()
}

func (d *MenuItemDelegate) drawShortcut(painter *qt.QPainter, option *qt.QStyleOptionViewItem, action *qt.QAction) {
	if action == nil {
		return
	}
	shortcut := action.Shortcut()
	if shortcut == nil || shortcut.IsEmpty() {
		return
	}
	text := shortcut.ToStringWithFormat(qt.QKeySequence__NativeText)
	if text == "" {
		return
	}

	painter.Save()
	font := common.GetFont(12, 400)
	defer font.Delete()
	painter.SetFont(font)

	if option.State()&qt.QStyle__State_Enabled == 0 {
		if common.IsDarkTheme() {
			painter.SetOpacity(0.5)
		} else {
			painter.SetOpacity(0.6)
		}
	}

	var c *qt.QColor
	if common.IsDarkTheme() {
		c = qt.NewQColor11(255, 255, 255, 200)
	} else {
		c = qt.NewQColor11(0, 0, 0, 153)
	}
	defer c.Delete()
	painter.SetPen(c)

	rect := option.Rect() // GoGC-armed — do NOT Delete
	// Draw the shortcut right-aligned (20px from the item's right edge) and
	// vertically centered. menu.py centers using the text layout's tight
	// bounding rect height; the previous code used QFontMetrics::height()
	// (which includes leading) so the text was drawn toward the top. Using
	// drawText(rect, flags) reproduces the tight-bounds centering exactly.
	textRect := qt.NewQRect4(rect.X(), rect.Y(), rect.Width()-20, rect.Height())
	defer textRect.Delete()
	painter.DrawText6(textRect, int(qt.AlignRight)|int(qt.AlignVCenter), text)
	painter.Restore()
}

func (d *MenuItemDelegate) drawChevron(painter *qt.QPainter, option *qt.QStyleOptionViewItem, index *qt.QModelIndex) {
	painter.Save()
	painter.SetRenderHints(qt.QPainter__Antialiasing)
	rect := option.Rect() // GoGC-armed — do NOT Delete

	// The chevron follows the item's preferred width (the end of the text), not
	// the full viewport width. menu.py paints it inside SubMenuItemWidget, whose
	// width is the item's sizeHint (the text width), so the chevron sits right
	// after the text — not at the menu's right edge.
	w := rect.Width()
	if item := d.menu.view.ItemFromIndex(index); item != nil {
		if s := item.SizeHint(); s != nil { // GoGC-armed — do NOT Delete
			w = s.Width()
		}
	}
	rectF := qt.NewQRectF4(float64(rect.X()+w-15), float64(rect.Y()+rect.Height()/2-4), 9, 9)
	defer rectF.Delete()
	renderFluentIcon(common.ChevronRight, painter, rectF, common.ThemeAuto)
	painter.Restore()
}

func (d *MenuItemDelegate) drawIndicator(painter *qt.QPainter, option *qt.QStyleOptionViewItem) {
	painter.Save()
	painter.SetRenderHints(qt.QPainter__Antialiasing)
	rect := option.Rect() // GoGC-armed — do NOT Delete

	switch d.indicatorType {
	case MenuIndicatorRadio:
		const r = 5
		x := rect.X() + 22
		y := rect.Y() + rect.Height()/2 - r/2

		if option.State()&qt.QStyle__State_MouseOver != 0 {
			painter.SetOpacity(1.0)
		} else if common.IsDarkTheme() {
			painter.SetOpacity(0.75)
		} else {
			painter.SetOpacity(0.65)
		}

		painter.SetPenWithStyle(qt.NoPen)
		var brush *qt.QBrush
		if common.IsDarkTheme() {
			brush = qt.NewQBrush4(qt.White)
		} else {
			brush = qt.NewQBrush4(qt.Black)
		}
		defer brush.Delete()
		painter.SetBrush(brush)

		ellipse := qt.NewQRectF4(float64(x), float64(y), r, r)
		defer ellipse.Delete()
		painter.DrawEllipse(ellipse)

	case MenuIndicatorCheck:
		const s = 11
		x := rect.X() + 19
		y := rect.Y() + rect.Height()/2 - s/2

		if option.State()&qt.QStyle__State_MouseOver == 0 {
			painter.SetOpacity(0.75)
		}

		rectF := qt.NewQRectF4(float64(x), float64(y), s, s)
		defer rectF.Delete()
		renderFluentIcon(common.Accept, painter, rectF, common.ThemeAuto)
	}

	painter.Restore()
}

// MenuActionListWidget is the internal list view of a RoundMenu. It mirrors the
// Python MenuActionListWidget: each menu item is a QListWidgetItem sized by the
// fluent width rules, and the view computes its own fixed size from the item
// size hints plus viewport margins.
type MenuActionListWidget struct {
	*qt.QListWidget
	itemHeight      int
	maxVisibleItems int
}

// NewMenuActionListWidget builds the menu action list widget.
func NewMenuActionListWidget(parent *qt.QWidget) *MenuActionListWidget {
	w := &MenuActionListWidget{
		QListWidget:     qt.NewQListWidget(parent),
		itemHeight:      28,
		maxVisibleItems: -1,
	}
	w.SetViewportMargins(0, 6, 0, 6)
	w.SetIconSize(qt.NewQSize2(14, 14))
	w.SetVerticalScrollBarPolicy(qt.ScrollBarAlwaysOff)
	w.SetHorizontalScrollBarPolicy(qt.ScrollBarAlwaysOff)
	w.SetMouseTracking(true)
	// menu.py sets Qt.ElideNone so item text is never elided with "..." and the
	// fluent item width rules (computed against the 14px menu font) decide the
	// visible width. The view font is also pinned to the 14px fluent font so
	// FontMetrics() matches the QSS `MenuActionListWidget { font: 14px ... }`
	// rule used for painting; without it the size hints were measured against
	// the default app font and came out too narrow, truncating item text.
	w.SetTextElideMode(qt.ElideNone)
	common.SetFont(w.QWidget, 14, int(qt.QFont__Normal))
	return w
}

// SetItemHeight sets the item height.
func (w *MenuActionListWidget) SetItemHeight(height int) {
	if height == w.itemHeight {
		return
	}
	for i := 0; i < w.Count(); i++ {
		item := w.Item(i)
		if w.ItemWidget(item) == nil {
			s := item.SizeHint() // GoGC-armed — do NOT Delete
			item.SetSizeHint(qt.NewQSize2(s.Width(), height))
		}
	}
	w.itemHeight = height
	w.AdjustSize()
}

// SetMaxVisibleItems sets the maximum visible item count.
func (w *MenuActionListWidget) SetMaxVisibleItems(num int) {
	w.maxVisibleItems = num
	w.AdjustSize()
}

// MaxVisibleItems returns the maximum visible item count.
func (w *MenuActionListWidget) MaxVisibleItems() int { return w.maxVisibleItems }

// ItemsHeight returns the total height of all (visible) items plus margins.
func (w *MenuActionListWidget) ItemsHeight() int {
	n := w.Count()
	if w.maxVisibleItems >= 0 && w.maxVisibleItems < n {
		n = w.maxVisibleItems
	}
	h := 0
	for i := 0; i < n; i++ {
		h += w.Item(i).SizeHint().Height()
	}
	m := w.ViewportMargins() // GoGC-armed — do NOT Delete
	return h + m.Top() + m.Bottom()
}

// availableHeight returns the vertical space available for the popup at pos for
// aniType, or 0 when the screen geometry is unavailable.
func (w *MenuActionListWidget) availableHeight(pos *qt.QPoint, aniType MenuAnimationType) int {
	screen := common.GetCurrentScreenGeometry(false)
	if screen == nil {
		return 0
	}
	// screen is GoGC-armed (QScreen.Geometry) — do NOT Delete
	var available int
	switch aniType {
	case MenuAnimationPullUp, MenuAnimationFadeInPullUp:
		available = pos.Y() - screen.Top() - 28
	default:
		available = screen.Bottom() - pos.Y() - 10
	}
	if available < 1 {
		available = 1
	}
	return available
}

// HeightForAnimation returns the estimated popup height at pos for aniType. It is
// used to choose the animation direction (mirrors menu.py heightForAnimation), so
// it does not include the +3 buffer.
func (w *MenuActionListWidget) HeightForAnimation(pos *qt.QPoint, aniType MenuAnimationType) int {
	ih := w.ItemsHeight()
	if avail := w.availableHeight(pos, aniType); avail > 0 && ih > avail {
		return avail
	}
	return ih
}

// clampHeight clamps the view's current (AdjustSize-computed) height to the
// available vertical space. menu.py adjustSize(pos, aniType) clamps
// size.height()+3 (the full height including the +3 buffer), not the raw
// itemsHeight, so the popup keeps the buffer instead of clipping the last row.
func (w *MenuActionListWidget) clampHeight(pos *qt.QPoint, aniType MenuAnimationType) int {
	h := w.Height()
	if avail := w.availableHeight(pos, aniType); avail > 0 && h > avail {
		return avail
	}
	return h
}

// AdjustSizeAt recomputes the size and clamps the fixed height for the popup at
// pos/aniType (mirrors the BreadcrumbBar height clamp in the Python port).
func (w *MenuActionListWidget) AdjustSizeAt(pos *qt.QPoint, aniType MenuAnimationType) {
	w.AdjustSize()
	w.SetFixedHeight(w.clampHeight(pos, aniType))
}

// AdjustSize recomputes the fixed size from the item size hints, mirrors the
// Python MenuActionListWidget.adjustSize.
func (w *MenuActionListWidget) AdjustSize() {
	maxW, totalH := 1, 0
	for i := 0; i < w.Count(); i++ {
		s := w.Item(i).SizeHint() // GoGC-armed — do NOT Delete
		if s.Width() > maxW {
			maxW = s.Width()
		}
		totalH += s.Height()
	}
	if totalH < 1 {
		totalH = 1
	}

	m := w.ViewportMargins() // GoGC-armed — do NOT Delete
	totalW := maxW + m.Left() + m.Right() + 2
	// Clamp the width to the available view width (screen width - 100), mirroring
	// menu.py adjustSize: size.setWidth(max(min(w, size.width()), self.minimumWidth())).
	// This keeps an unusually wide menu inside the screen instead of running past
	// the right edge where its text would be clipped.
	if screen := common.GetCurrentScreenGeometry(false); screen != nil {
		// screen is GoGC-armed (QScreen.Geometry) — do NOT Delete
		availW := screen.Width() - 100
		if availW < 1 {
			availW = 1
		}
		if totalW > availW {
			totalW = availW
		}
	}
	if minW := w.MinimumWidth(); totalW < minW {
		totalW = minW
	}
	// The +3 mirrors menu.py adjustSize (size.setHeight(min(h, size.height()+3)):
	// it reserves a little extra room for the item margin so a 4-item combo
	// menu never overflows and shows an unwanted scrollbar.
	totalH += m.Top() + m.Bottom() + 3

	if w.maxVisibleItems > 0 && w.maxVisibleItems < w.Count() {
		totalH = w.maxVisibleItems*w.itemHeight + m.Top() + m.Bottom() + 3
	}

	w.SetFixedSize2(totalW, totalH)
}

// RoundMenu is the fluent round-corner menu. It keeps the native QMenu popup
// machinery (input grab / auto-dismiss) but renders its items through a
// QListWidget child (MenuActionListWidget), exactly like the Python port, so the
// fluent menu QSS applies to the items.
type RoundMenu struct {
	*qt.QMenu
	view       *MenuActionListWidget
	hBoxLayout *qt.QHBoxLayout
	delegate   *MenuItemDelegate

	actions  []*qt.QAction
	subMenus []*RoundMenu
	// itemData is keyed by the C++ QListWidgetItem pointer (item.UnsafePointer)
	// rather than the Go wrapper pointer: miqt constructs a fresh Go wrapper for
	// the same C++ item on every signal delivery and every Item/ItemFromIndex
	// call, so a *qt.QListWidgetItem map key would never match.
	itemData map[unsafe.Pointer]*menuItemData

	isSubMenu  bool
	parentMenu *RoundMenu
	menuItem   *qt.QListWidgetItem
	itemHeight int
	// extraWidth is appended to action item width hints by subclasses that
	// reserve extra horizontal space (CheckableMenu adds 26px for the check
	// indicator padding, mirroring menu.py CheckableMenu._adjustItemText).
	extraWidth int
	icon       *qt.QIcon
	title      string
	onClosed   func()

	timer                *qt.QTimer
	lastHoverItem        *qt.QListWidgetItem
	lastHoverSubMenuItem *qt.QListWidgetItem

	// anim is the in-flight pop-up animation. It is owned here and released in
	// stopAnim (called on close/clear and before a new popup), never inside the
	// animation's own finished signal (that would be a use-after-free).
	anim *common.ProgressAnimation

	// alive drops the action callbacks once the menu is destroyed: the actions it
	// renders are commonly unparented and outlive it (a combo box builds a fresh
	// menu per drop-down, which dies with the combo), and an action's changed
	// signal firing into a destroyed menu would crash (see widgetAlive).
	alive *widgetAlive
}

// NewRoundMenu builds a round menu.
func NewRoundMenu(title string, parent *qt.QWidget) *RoundMenu {
	m := &RoundMenu{QMenu: qt.NewQMenu4(title, parent), itemHeight: 28, title: title, itemData: map[unsafe.Pointer]*menuItemData{}}
	m.alive = trackWidget(m.OnDestroyed)
	m.SetTitle(title)
	m.SetWindowFlags(qt.Popup | qt.FramelessWindowHint | qt.NoDropShadowWindowHint)
	m.SetAttribute(qt.WA_TranslucentBackground)
	m.SetMouseTracking(true)

	m.view = NewMenuActionListWidget(m.QWidget)
	m.view.SetItemHeight(28)

	m.hBoxLayout = qt.NewQHBoxLayout(m.QWidget)
	m.hBoxLayout.SetContentsMargins(12, 8, 12, 20)
	m.hBoxLayout.AddWidget3(m.view.QWidget, 1, qt.AlignCenter)

	common.FluentStyleSheet(common.FluentMenu).Apply(m.QWidget, common.ThemeAuto)
	m.delegate = NewMenuItemDelegate(m)
	m.view.SetItemDelegate(m.delegate.QAbstractItemDelegate)
	m.setShadowEffect(30, 0, 8)

	m.timer = qt.NewQTimer2(m.QObject)
	m.timer.SetSingleShot(true)
	m.timer.SetInterval(400)
	m.timer.OnTimeout(m.onShowMenuTimeOut)

	// The QMenu's native frame/items are not painted; the QListWidget child
	// renders the fluent surface (mirrors Python RoundMenu.paintEvent -> pass).
	m.OnPaintEvent(func(super func(ev *qt.QPaintEvent), ev *qt.QPaintEvent) {})

	m.view.OnItemClicked(func(item *qt.QListWidgetItem) { m.onItemClicked(item) })
	m.view.OnItemEntered(func(item *qt.QListWidgetItem) { m.onItemEntered(item) })

	m.OnAboutToHide(func() {
		m.stopAnim()
		if m.onClosed != nil {
			m.onClosed()
		}
	})
	return m
}

// setShadowEffect adds a drop shadow behind the list view.
func (m *RoundMenu) setShadowEffect(blurRadius float64, dx, dy float64) {
	color := qt.NewQColor11(0, 0, 0, 30)
	defer color.Delete()
	effect := qt.NewQGraphicsDropShadowEffect2(m.view.QObject)
	effect.SetBlurRadius(blurRadius)
	effect.SetOffset2(dx, dy)
	effect.SetColor(color)
	m.view.SetGraphicsEffect(nil)
	m.view.SetGraphicsEffect(effect.QGraphicsEffect)
}

// stopAnim stops and releases any in-flight pop-up animation. It is safe to
// call from menu signals (aboutToHide / clear) and before a new popup, but must
// never be called from inside the animation's own finished callback.
func (m *RoundMenu) stopAnim() {
	if m.anim != nil {
		m.anim.Stop()
		m.anim.Delete()
		m.anim = nil
	}
}

// View returns the internal list view.
func (m *RoundMenu) View() *MenuActionListWidget { return m.view }

// SetTitle sets the menu title.
func (m *RoundMenu) SetTitle(title string) {
	m.title = title
	m.QMenu.SetTitle(title)
}

// SetIcon sets the menu icon.
func (m *RoundMenu) SetIcon(icon interface{}) {
	m.icon = common.ToQIcon(icon)
	m.QMenu.SetIcon(m.icon)
}

// SetItemHeight sets the height of menu items.
func (m *RoundMenu) SetItemHeight(height int) {
	if height == m.itemHeight {
		return
	}
	m.itemHeight = height
	m.view.SetItemHeight(height)
}

// SetMaxVisibleItems sets the maximum visible items.
func (m *RoundMenu) SetMaxVisibleItems(num int) {
	m.view.SetMaxVisibleItems(num)
	m.AdjustSize()
}

// OnClosed registers a callback emitted when the menu hides.
func (m *RoundMenu) OnClosed(f func()) { m.onClosed = f }

// EmitClosed fires the closed signal.
func (m *RoundMenu) EmitClosed() {
	if m.onClosed != nil {
		m.onClosed()
	}
}

// MenuActions returns the actions added through AddAction/InsertAction.
func (m *RoundMenu) MenuActions() []*qt.QAction { return m.actions }

// AddAction adds a QAction to the menu and renders it in the list view.
func (m *RoundMenu) AddAction(action *qt.QAction) {
	// Register the action before building its item so hasItemIcon() and
	// longestShortcutWidth() see it (menu.py appends to self._actions first).
	m.actions = append(m.actions, action)
	item := m.createActionItem(action)
	m.view.AddItemWithItem(item)
	m.itemData[item.UnsafePointer()] = &menuItemData{kind: menuItemAction, action: action}
	action.OnChanged(func() {
		if m.alive.ok() {
			m.onActionChanged(action)
		}
	})
	m.resizeSubMenuItems()
	m.AdjustSize()
}

// AddActions adds multiple actions.
func (m *RoundMenu) AddActions(actions []*qt.QAction) {
	for _, a := range actions {
		m.AddAction(a)
	}
}

// AddActionText creates and adds an action from text.
func (m *RoundMenu) AddActionText(text string) *qt.QAction {
	a := qt.NewQAction2(text)
	m.AddAction(a)
	return a
}

// AddActionIcon creates and adds an action from icon and text.
func (m *RoundMenu) AddActionIcon(icon interface{}, text string) *qt.QAction {
	a := qt.NewQAction3(common.ToQIcon(icon), text)
	m.AddAction(a)
	return a
}

// InsertAction inserts an action before another action.
func (m *RoundMenu) InsertAction(before, action *qt.QAction) {
	row := m.rowOfAction(before)
	if row < 0 {
		return
	}

	// Insert into m.actions before building the item so hasItemIcon() and
	// longestShortcutWidth() see the action being inserted (mirrors menu.py
	// _createActionItem which inserts into self._actions first).
	for i, a := range m.actions {
		if a == before {
			m.actions = append(m.actions[:i], append([]*qt.QAction{action}, m.actions[i:]...)...)
			break
		}
	}

	item := m.createActionItem(action)
	m.view.InsertItem(row, item)
	m.itemData[item.UnsafePointer()] = &menuItemData{kind: menuItemAction, action: action}
	action.OnChanged(func() {
		if m.alive.ok() {
			m.onActionChanged(action)
		}
	})
	m.resizeSubMenuItems()
	m.AdjustSize()
}

// RemoveAction removes an action from the menu.
func (m *RoundMenu) RemoveAction(action *qt.QAction) {
	row := m.rowOfAction(action)
	if row < 0 {
		return
	}
	item := m.view.TakeItem(row)
	if item != nil {
		delete(m.itemData, item.UnsafePointer())
	}
	for i, a := range m.actions {
		if a == action {
			m.actions = append(m.actions[:i], m.actions[i+1:]...)
			break
		}
	}
	m.resizeSubMenuItems()
	m.AdjustSize()
}

// AddMenu adds a sub menu (shown on hover, with a right chevron).
func (m *RoundMenu) AddMenu(menu *RoundMenu) {
	// Register the sub-menu before building its item so hasItemIcon() sees it
	// (menu.py appends to self._subMenus first).
	m.subMenus = append(m.subMenus, menu)
	item := m.createSubMenuItem(menu)
	m.view.AddItemWithItem(item)
	m.itemData[item.UnsafePointer()] = &menuItemData{kind: menuItemSubMenu, submenu: menu}
	menu.isSubMenu = true
	menu.parentMenu = m
	menu.menuItem = item
	m.AdjustSize()
}

// AddSeparator adds a separator.
func (m *RoundMenu) AddSeparator() {
	item := qt.NewQListWidgetItem()
	item.SetFlags(qt.NoItemFlags)
	margins := m.view.ViewportMargins() // GoGC-armed — do NOT Delete
	item.SetSizeHint(qt.NewQSize2(m.view.Width()-margins.Left()-margins.Right(), 9))
	m.view.AddItemWithItem(item)
	m.itemData[item.UnsafePointer()] = &menuItemData{kind: menuItemSeparator}
	m.AdjustSize()
}

// AddWidget adds a custom widget as a (non-selectable) menu item.
func (m *RoundMenu) AddWidget(widget *qt.QWidget) {
	widget.SetParent(m.QWidget)
	item := qt.NewQListWidgetItem2("")
	item.SetFlags(qt.NoItemFlags)
	item.SetSizeHint(widget.Size()) // GoGC-armed — do NOT Delete
	m.view.AddItemWithItem(item)
	m.view.SetItemWidget(item, widget)
	m.itemData[item.UnsafePointer()] = &menuItemData{kind: menuItemWidget, widget: widget}
	m.AdjustSize()
}

// SetDefaultAction selects the given action.
func (m *RoundMenu) SetDefaultAction(action *qt.QAction) {
	row := m.rowOfAction(action)
	if row < 0 {
		return
	}
	m.view.SetCurrentItem(m.view.Item(row))
}

// Clear removes all actions and sub-menus.
func (m *RoundMenu) Clear() {
	m.stopAnim()
	m.actions = nil
	m.subMenus = nil
	m.itemData = map[unsafe.Pointer]*menuItemData{}
	m.view.Clear()
	m.lastHoverItem = nil
	m.lastHoverSubMenuItem = nil
}

// AdjustSize recomputes the menu fixed size from the view plus layout margins.
func (m *RoundMenu) AdjustSize() {
	m.view.AdjustSize()
	mg := m.hBoxLayout.ContentsMargins() // GoGC-armed — do NOT Delete
	m.SetFixedSize2(m.view.Width()+mg.Left()+mg.Right(), m.view.Height()+mg.Top()+mg.Bottom())
}

// Exec shows the menu at pos with the requested pop-up animation (slide and/or
// fade driven by a real ProgressAnimation).
func (m *RoundMenu) Exec(pos *qt.QPoint, aniType MenuAnimationType) {
	m.stopAnim()
	m.AdjustSize()
	// Clamp the height to the available vertical space (mirrors the Python
	// _showMenu's adjustSize(pos, aniType) before exec) so a menu whose content
	// is taller than the screen — or whose anchor button sits near a screen edge
	// — is not clipped and instead scrolls. clampHeight keeps the +3 buffer that
	// AdjustSize already added, so the last row is not clipped.
	h := m.View().clampHeight(pos, aniType)
	m.View().SetFixedHeight(h)
	mg := m.hBoxLayout.ContentsMargins() // GoGC-armed — do NOT Delete
	m.SetFixedHeight(h + mg.Top() + mg.Bottom())
	p := m.adjustedPos(pos, aniType)
	defer p.Delete()

	if aniType == MenuAnimationNone {
		m.SetWindowOpacity(1.0)
		m.Move(p.X(), p.Y())
		m.Show()
		return
	}

	targetX, targetY := p.X(), p.Y()
	startY := targetY
	fadeIn := false
	// The plain (non-fade) variants slide by a fixed 12px. The round-12
	// (height+5)/2 slide was reverted because it overshoots the reference's
	// shallow reveal. The FadeIn variants keep the reference's fixed 8px slide.
	switch aniType {
	case MenuAnimationPullUp:
		startY = targetY + 12
	case MenuAnimationDropDown:
		startY = targetY - 12
	case MenuAnimationFadeInPullUp:
		startY = targetY + 8
		fadeIn = true
	case MenuAnimationFadeInDropDown:
		startY = targetY - 8
		fadeIn = true
	}

	// Only the fade-in variants animate window opacity; the plain slide
	// variants must stay fully opaque for the whole animation (menu.py uses a
	// position-only QPropertyAnimation for DROP_DOWN / PULL_UP).
	if fadeIn {
		m.SetWindowOpacity(0.0)
	} else {
		m.SetWindowOpacity(1.0)
	}
	// Show the menu via the Qt::Popup window flag (exactly like menu.py's
	// RoundMenu.exec which calls self.show()) instead of QMenu::popup(). The
	// native QMenu::popup() machinery renders/reserves a native (empty) item on
	// top of the QListWidget child — the source of the stray blank option — and
	// also repositions the window, which skewed the sub-menu placement.
	m.Move(targetX, startY)
	m.Show()

	curve := common.CreateBezierCurve(0, 0, 0, 1) // OutQuad
	ani := common.NewProgressAnimation(250, curve)
	curve.Delete()
	m.anim = ani

	ani.OnProgress(func(t float64) {
		if fadeIn {
			m.SetWindowOpacity(t)
		}
		m.Move(targetX, startY+int(float64(targetY-startY)*t))
	})
	// Do NOT Delete the animation here: this callback runs inside
	// QAbstractAnimation::finished, and Qt5 continues accessing the object after
	// the signal is emitted. It is released by stopAnim on close/clear.
	ani.OnFinished(func() {
		if fadeIn {
			m.SetWindowOpacity(1.0)
		}
		m.Move(targetX, targetY)
	})
	ani.Start()

	if m.isSubMenu && m.menuItem != nil {
		m.menuItem.SetSelected(true)
	}
}

// ExecAt shows the menu at pos with the drop-down animation.
func (m *RoundMenu) ExecAt(pos *qt.QPoint) {
	m.Exec(pos, MenuAnimationDropDown)
}

// ExecAtCentered shows the menu centered on widget, mirroring Python
// button._showMenu: the menu view's minimum width is set to the widget width so
// the menu matches the button, the menu is horizontally centered on the widget,
// and the animation direction (DROP_DOWN vs PULL_UP) is chosen by the available
// vertical space.
func (m *RoundMenu) ExecAtCentered(widget *qt.QWidget) {
	m.View().SetMinimumWidth(widget.Width())
	m.AdjustSize()

	margins := m.hBoxLayout.ContentsMargins() // GoGC-armed — do NOT Delete
	x := -m.Width()/2 + margins.Left() + widget.Width()/2

	p := qt.NewQPoint2(x, widget.Height())
	pd := widget.MapToGlobal(p) // GoGC-armed — do NOT Delete
	p.Delete()

	p = qt.NewQPoint2(x, 0)
	pu := widget.MapToGlobal(p) // GoGC-armed — do NOT Delete
	p.Delete()

	hd := m.View().HeightForAnimation(pd, MenuAnimationDropDown)
	hu := m.View().HeightForAnimation(pu, MenuAnimationPullUp)
	if hd >= hu {
		m.Exec(pd, MenuAnimationDropDown)
	} else {
		m.Exec(pu, MenuAnimationPullUp)
	}
}

func (m *RoundMenu) adjustedPos(pos *qt.QPoint, aniType MenuAnimationType) *qt.QPoint {
	screen := common.GetCurrentScreenGeometry(false)
	if screen == nil {
		return qt.NewQPoint2(pos.X(), pos.Y())
	}
	// screen is GoGC-armed (QScreen.Geometry) — do NOT Delete
	w := m.Width() + 5
	h := m.Height()
	x := pos.X() - m.hBoxLayout.ContentsMargins().Left()
	if x+w > screen.Right() {
		x = screen.Right() - w
	}
	if x < screen.Left() {
		x = screen.Left()
	}

	var y int
	switch aniType {
	case MenuAnimationPullUp, MenuAnimationFadeInPullUp:
		y = pos.Y() - h + 13
		if y < screen.Top()+4 {
			y = screen.Top() + 4
		}
	default:
		y = pos.Y() - 4
		if y+h > screen.Bottom() {
			y = screen.Bottom() - h + 10
		}
	}
	return qt.NewQPoint2(x, y)
}

// ---------------------------------------------------------------------------
// Item rendering helpers (mirror menu.py)

func (m *RoundMenu) createActionItem(action *qt.QAction) *qt.QListWidgetItem {
	item := qt.NewQListWidgetItem3(m.createItemIcon(action), action.Text())
	m.adjustItemText(item, action)

	if !action.IsEnabled() {
		item.SetFlags(qt.NoItemFlags)
	}
	if action.Text() != action.ToolTip() {
		item.SetToolTip(action.ToolTip())
	}
	return item
}

// onActionChanged re-syncs the item of an action after the action changes
// (icon/text/shortcut/enabled). It mirrors menu.py RoundMenu._onActionChanged so
// a shortcut set after AddAction still reserves space and re-renders the icon.
func (m *RoundMenu) onActionChanged(action *qt.QAction) {
	row := m.rowOfAction(action)
	if row < 0 {
		return
	}
	item := m.view.Item(row)
	if item == nil {
		return
	}

	item.SetIcon(m.createItemIcon(action))
	if action.Text() != action.ToolTip() {
		item.SetToolTip(action.ToolTip())
	}
	m.adjustItemText(item, action)

	if action.IsEnabled() {
		item.SetFlags(qt.ItemIsSelectable | qt.ItemIsEnabled)
	} else {
		item.SetFlags(qt.NoItemFlags)
	}

	m.resizeSubMenuItems()
	m.AdjustSize()
}

func (m *RoundMenu) createSubMenuItem(menu *RoundMenu) *qt.QListWidgetItem {
	item := qt.NewQListWidgetItem3(m.createItemIconForAction(menu.icon), menu.title)
	if m.hasItemIcon() {
		item.SetText(" " + item.Text())
	}
	item.SetSizeHint(qt.NewQSize2(m.subMenuItemWidth(menu.title), m.itemHeight))
	item.SetTextAlignment(int(qt.AlignLeft) | int(qt.AlignVCenter))
	return item
}

// subMenuItemWidth computes the sub-menu item width, accounting for the longest
// shortcut width exactly like adjustItemText does for ordinary action items. The
// reference menu.py leaves sub-menu items at 60/72 + title width, which makes the
// sub-menu row narrower than shortcut-widened action rows (chevron/popup drift
// left) and wider than them when no shortcut exists (drift right). Adding the
// same sw keeps the sub-menu row consistent with the menu's overall width.
func (m *RoundMenu) subMenuItemWidth(title string) int {
	sw := m.shortcutWidth()
	if !m.hasItemIcon() {
		return 60 + m.view.FontMetrics().Width(title) + sw
	}
	return 72 + m.view.FontMetrics().Width(" "+title) + sw
}

// resizeSubMenuItems re-sizes every sub-menu item so its width keeps tracking the
// longest shortcut width. Action shortcuts are often set after the sub-menu was
// added (e.g. examples/menu/menu/main.go inserts shortcut actions after AddMenu),
// so the sub-menu row must be re-measured whenever the action list changes.
func (m *RoundMenu) resizeSubMenuItems() {
	for i := 0; i < m.view.Count(); i++ {
		item := m.view.Item(i)
		info := m.itemData[item.UnsafePointer()]
		if info == nil || info.kind != menuItemSubMenu || info.submenu == nil {
			continue
		}
		item.SetSizeHint(qt.NewQSize2(m.subMenuItemWidth(info.submenu.title), m.itemHeight))
	}
}

func (m *RoundMenu) adjustItemText(item *qt.QListWidgetItem, action *qt.QAction) {
	sw := m.shortcutWidth()

	fm := m.view.FontMetrics() // GoGC-armed — do NOT Delete
	var w int
	if !m.hasItemIcon() {
		item.SetText(action.Text())
		w = 40 + fm.Width(action.Text()) + sw
	} else {
		item.SetText(" " + action.Text())
		space := 4 - fm.Width(" ")
		w = 60 + fm.Width(item.Text()) + sw + space
	}
	// menu.py relies on QStyledItemDelegate's default display alignment
	// (AlignLeft | AlignVCenter) to center the label vertically in the 28px
	// item. Pin it explicitly so text never drifts to the bottom when the item
	// has no icon.
	item.SetTextAlignment(int(qt.AlignLeft) | int(qt.AlignVCenter))
	item.SetSizeHint(qt.NewQSize2(w+m.extraWidth, m.itemHeight))
}

func (m *RoundMenu) longestShortcutWidth() int {
	font := common.GetFont(12, 400)
	defer font.Delete()
	fm := qt.NewQFontMetrics(font)
	defer fm.Delete()

	maxW := 0
	for _, action := range m.actions {
		shortcut := action.Shortcut()
		if shortcut == nil || shortcut.IsEmpty() {
			continue
		}
		if w := fm.Width(shortcut.ToStringWithFormat(qt.QKeySequence__NativeText)); w > maxW {
			maxW = w
		}
	}
	return maxW
}

// shortcutWidth returns the extra horizontal space reserved for the shortcut
// column (longest shortcut + 22px gap), mirroring menu.py _adjustItemText.
func (m *RoundMenu) shortcutWidth() int {
	sw := m.longestShortcutWidth()
	if sw > 0 {
		sw += 22
	}
	return sw
}

// hasItemIcon reports whether any action or sub-menu carries a non-null icon.
func (m *RoundMenu) hasItemIcon() bool {
	for _, a := range m.actions {
		if !a.Icon().IsNull() {
			return true
		}
	}
	for _, sub := range m.subMenus {
		if sub.icon != nil && !sub.icon.IsNull() {
			return true
		}
	}
	return false
}

func (m *RoundMenu) createItemIcon(action *qt.QAction) *qt.QIcon {
	return m.createItemIconForAction(action.Icon())
}

func (m *RoundMenu) createItemIconForAction(icon *qt.QIcon) *qt.QIcon {
	if !m.hasItemIcon() {
		return qt.NewQIcon()
	}
	if icon != nil && !icon.IsNull() {
		return icon
	}
	// Transparent placeholder keeps the text aligned with icon-bearing items.
	pix := qt.NewQPixmap2(m.view.IconSize().Width(), m.view.IconSize().Height())
	defer pix.Delete()
	transparent := qt.NewQColor11(0, 0, 0, 0)
	defer transparent.Delete()
	pix.FillWithFillColor(transparent)
	return qt.NewQIcon2(pix)
}

func (m *RoundMenu) rowOfAction(action *qt.QAction) int {
	for i := 0; i < m.view.Count(); i++ {
		if info := m.itemData[m.view.Item(i).UnsafePointer()]; info != nil && info.kind == menuItemAction && info.action == action {
			return i
		}
	}
	return -1
}

// ---------------------------------------------------------------------------
// Interaction

func (m *RoundMenu) onItemClicked(item *qt.QListWidgetItem) {
	info := m.itemData[item.UnsafePointer()]
	if info == nil || info.kind != menuItemAction {
		return
	}
	if !info.action.IsEnabled() {
		return
	}
	m.hideMenu(false)

	if !m.isSubMenu {
		info.action.Trigger()
		return
	}

	// close the whole parent chain before triggering (mirrors menu.py
	// _onItemClicked / _closeParentMenu).
	m.closeParentMenu()
	info.action.Trigger()
}

func (m *RoundMenu) onItemEntered(item *qt.QListWidgetItem) {
	m.lastHoverItem = item
	info := m.itemData[item.UnsafePointer()]
	if info == nil || info.kind != menuItemSubMenu {
		return
	}
	m.showSubMenu(item)
}

func (m *RoundMenu) showSubMenu(item *qt.QListWidgetItem) {
	m.lastHoverSubMenuItem = item
	m.timer.Stop()
	m.timer.Start(400)
}

func (m *RoundMenu) onShowMenuTimeOut() {
	if m.lastHoverSubMenuItem == nil || m.lastHoverItem != m.lastHoverSubMenuItem {
		return
	}
	info := m.itemData[m.lastHoverSubMenuItem.UnsafePointer()]
	if info == nil || info.submenu == nil {
		return
	}
	sub := info.submenu
	if sub.parentMenu == nil || sub.parentMenu.IsHidden() {
		return
	}

	rect := m.view.VisualItemRect(m.lastHoverSubMenuItem) // GoGC-armed — do NOT Delete
	// VisualItemRect returns viewport coordinates, so map through the viewport
	// (not the QListWidget widget, whose frame/border would offset the point).
	topLeft := m.view.Viewport().MapToGlobal(rect.TopLeft()) // GoGC-armed — do NOT Delete

	// menu.py positions the sub-menu relative to the sub-menu item widget, whose
	// width is the item's sizeHint (the text width). Using the full row width
	// pushed the sub-menu to the parent menu's right edge instead of overlapping
	// the sub-menu button.
	itemW := rect.Width()
	if s := m.lastHoverSubMenuItem.SizeHint(); s != nil { // GoGC-armed — do NOT Delete
		itemW = s.Width()
	}

	// itemRect := qt.NewQre
	// x := topLeft.X() + itemW + 5
	// y := topLeft.Y() - 5
	x := topLeft.X() + itemW - 7
	y := topLeft.Y() - 5

	screen := common.GetCurrentScreenGeometry(false)
	if screen != nil {
		subW := sub.Width()
		subH := sub.Height()
		if x+subW > screen.Right() {
			x = topLeft.X() - subW - 5
			if x < screen.Left() {
				x = screen.Left()
			}
		}
		if y+subH > screen.Bottom() {
			y = screen.Bottom() - subH
		}
		if y < screen.Top() {
			y = screen.Top()
		}
	}
	// fmt.Println("showSubMenu", x, y)
	sub.Exec(qt.NewQPoint2(x, y), MenuAnimationDropDown)
}

func (m *RoundMenu) hideMenu(isHideBySystem bool) {
	m.view.ClearSelection()
	if m.isSubMenu {
		m.Hide()
	} else {
		m.Close()
	}
}

// closeParentMenu closes the whole sub-menu chain starting from m (mirrors
// menu.py RoundMenu._closeParentMenu).
func (m *RoundMenu) closeParentMenu() {
	menu := m
	for menu != nil {
		menu.Close()
		menu = menu.parentMenu
	}
}

// ---------------------------------------------------------------------------
// CheckableMenu

// CheckableMenu is a round menu whose checked actions display a check or radio
// indicator.
type CheckableMenu struct {
	*RoundMenu
}

// NewCheckableMenu builds a checkable menu.
func NewCheckableMenu(title string, parent *qt.QWidget, indicatorType MenuIndicatorType) *CheckableMenu {
	m := &CheckableMenu{RoundMenu: NewRoundMenu(title, parent)}
	m.delegate = NewCheckableMenuItemDelegate(m.RoundMenu, indicatorType)
	m.view.SetItemDelegate(m.delegate.QAbstractItemDelegate)
	m.view.SetObjectName("checkableListWidget")
	// Re-polish so the #checkableListWidget::item padding rule (reserving space
	// for the check indicator) takes effect after the objectName change.
	m.view.SetStyle(qt.QApplication_Style())
	// The extra 36px padding-left needs a matching width compensation so the
	// text and shortcut are not clipped (mirrors CheckableMenu._adjustItemText).
	m.extraWidth = 26
	return m
}

// LabelContextMenu is the context menu shown for text labels.
type LabelContextMenu struct {
	*RoundMenu
	label *qt.QLabel
}

// NewLabelContextMenu builds a label context menu.
func NewLabelContextMenu(label *qt.QLabel) *LabelContextMenu {
	m := &LabelContextMenu{RoundMenu: NewRoundMenu("", label.QWidget), label: label}
	copyAct := m.AddActionIcon(common.FluentIcon(common.Copy), "Copy")
	copyAct.SetShortcut(qt.NewQKeySequence2("Ctrl+C"))
	copyAct.OnTriggered(func() {
		qt.QGuiApplication_Clipboard().SetText(label.SelectedText())
	})
	selectAllAct := m.AddActionText("Select all")
	selectAllAct.SetShortcut(qt.NewQKeySequence2("Ctrl+A"))
	selectAllAct.OnTriggered(func() {
		label.SetSelection(0, len(label.Text()))
	})
	return m
}

func (m *LabelContextMenu) execAt(pos *qt.QPoint) {
	m.RoundMenu.Exec(pos, MenuAnimationDropDown)
}
