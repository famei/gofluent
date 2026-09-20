package navigation

import (
	"math"
	"unsafe"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	qt "github.com/mappu/miqt/qt"
)

// navigationSeparatorPtrs tracks live NavigationSeparator instances by their
// C++ pointer so NavigationItemLayout can re-anchor them to x=0.
var navigationSeparatorPtrs = map[unsafe.Pointer]bool{}

// setNavigationCompacted dispatches SetCompacted to the concrete widget type
// (Go has no virtual dispatch through embedded structs).
func setNavigationCompacted(v interface{}, isCompacted bool) {
	switch w := v.(type) {
	case *NavigationSeparator:
		w.SetCompacted(isCompacted)
	case *NavigationItemHeader:
		w.SetCompacted(isCompacted)
	case *NavigationTreeWidget:
		w.SetCompacted(isCompacted)
	case *NavigationUserCard:
		w.SetCompacted(isCompacted)
	case *NavigationPushButton:
		w.SetCompacted(isCompacted)
	case *NavigationToolButton:
		w.SetCompacted(isCompacted)
	case *NavigationAvatarWidget:
		w.SetCompacted(isCompacted)
	case *NavigationWidget:
		w.SetCompacted(isCompacted)
	}
}

// navWidgetOf returns the embedded *NavigationWidget for any navigation widget
// type (Go has no implicit up-cast through struct embedding).
func navWidgetOf(v interface{}) *NavigationWidget {
	switch w := v.(type) {
	case *NavigationWidget:
		return w
	case *NavigationPushButton:
		return w.NavigationWidget
	case *NavigationBarPushButton:
		return w.NavigationPushButton.NavigationWidget
	case *NavigationToolButton:
		return w.NavigationWidget
	case *NavigationSeparator:
		return w.NavigationWidget
	case *NavigationItemHeader:
		return w.NavigationWidget
	case *NavigationTreeWidget:
		return w.NavigationWidget
	case *NavigationAvatarWidget:
		return w.NavigationWidget
	case *NavigationUserCard:
		return w.NavigationWidget
	default:
		return nil
	}
}

// navigationWidgetExpandWidth mirrors the mutable Python class attribute
// NavigationWidget.EXPAND_WIDTH (312 by default). NavigationPanel.setExpandWidth
// updates it, exactly like the Python port.
var navigationWidgetExpandWidth = 312

// cloneColor returns a copy of c (QColor is a value type in miqt, so painting
// code must clone before mutating a shared color).
func cloneColor(c *qt.QColor) *qt.QColor {
	if c == nil {
		return nil
	}
	return qt.NewQColor9(c)
}

// coerceColor converts a color-like value (string, *qt.QColor or
// common.FluentThemeColor) into a caller-owned *qt.QColor.
func coerceColor(v interface{}) *qt.QColor {
	switch c := v.(type) {
	case *qt.QColor:
		return cloneColor(c)
	case string:
		return qt.NewQColor6(c)
	case common.FluentThemeColor:
		return c.Color()
	default:
		return qt.NewQColor()
	}
}

// fallbackThemeColor returns a caller-owned indicator color: the light/dark
// color when valid, otherwise the primary theme color.
func fallbackThemeColor(light, dark *qt.QColor) *qt.QColor {
	if common.IsDarkTheme() {
		if dark != nil && dark.IsValid() {
			return cloneColor(dark)
		}
	} else if light != nil && light.IsValid() {
		return cloneColor(light)
	}
	return common.ThemeColorPrimary.Color()
}

// iconIsNull reports whether an icon source produces an empty QIcon without
// forcing an expensive SVG render (common.ToQIcon would render a pixmap).
func iconIsNull(icon interface{}) bool {
	switch v := icon.(type) {
	case *qt.QIcon:
		return v.IsNull()
	case string:
		return v == ""
	case nil:
		return true
	default:
		return false
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ---------------------------------------------------------------------------
// NavigationWidget

// NavigationWidget is the base class of every navigation item. It replaces the
// PyQt5 NavigationWidget (a QWidget subclass) and stores the hover/selection
// state plus the light/dark text and indicator colors.
type NavigationWidget struct {
	*qt.QWidget

	IsCompacted     bool
	IsSelected      bool
	IsPressed       bool
	IsEnter         bool
	IsAboutSelected bool
	IsSelectable    bool
	TreeParent      *NavigationTreeWidget
	NodeDepth       int

	routeKey       string
	parentRouteKey string

	lightTextColor      *qt.QColor
	darkTextColor       *qt.QColor
	lightIndicatorColor *qt.QColor
	darkIndicatorColor  *qt.QColor

	marginsFunc func() (int, int)
	// indicatorRectFunc is the per-widget IndicatorRect override. Go has no
	// virtual dispatch: the indicator animators only ever hold a
	// *NavigationWidget, so an outer override (NavigationBarPushButton) has to be
	// installed here as well, otherwise the animated indicator would use the
	// base 3x16 geometry while the button paints its own 4x24 bar.
	indicatorRectFunc func() *qt.QRectF

	clickedSig         boolSignal
	selectedChangedSig boolSignal
}

// margins returns the left/right drawing margins (overridden per widget via
// marginsFunc; the base widget has none).
func (w *NavigationWidget) margins() (int, int) {
	if w.marginsFunc != nil {
		return w.marginsFunc()
	}
	return 0, 0
}

func newNavigationWidget(isSelectable bool, parent *qt.QWidget) *NavigationWidget {
	w := &NavigationWidget{QWidget: qt.NewQWidget(parent)}
	w.IsCompacted = true
	w.IsSelectable = isSelectable
	w.lightTextColor = qt.NewQColor3(0, 0, 0)
	w.darkTextColor = qt.NewQColor3(255, 255, 255)
	w.lightIndicatorColor = qt.NewQColor()
	w.darkIndicatorColor = qt.NewQColor()
	w.SetFixedSize2(40, 36)
	return w
}

// OnClicked registers a clicked listener (replaces clicked.connect).
func (w *NavigationWidget) OnClicked(f func(bool)) { w.clickedSig.connect(f) }

// OnSelectedChanged registers a selectedChanged listener.
func (w *NavigationWidget) OnSelectedChanged(f func(bool)) { w.selectedChangedSig.connect(f) }

func (w *NavigationWidget) emitClicked(v bool)         { w.clickedSig.emit(v) }
func (w *NavigationWidget) emitSelectedChanged(v bool) { w.selectedChangedSig.emit(v) }

// Click emits the clicked signal with true (user-triggered).
func (w *NavigationWidget) Click() { w.emitClicked(true) }

func (w *NavigationWidget) handleEnter() {
	w.IsEnter = true
	w.Update()
}

func (w *NavigationWidget) handleLeave() {
	w.IsEnter = false
	w.IsPressed = false
	w.Update()
}

func (w *NavigationWidget) handlePress() {
	w.IsPressed = true
	w.Update()
}

func (w *NavigationWidget) handleRelease() {
	w.IsPressed = false
	w.Update()
	w.emitClicked(true)
}

// registerInteractionEvents wires the hover/press/release behavior shared by
// NavigationPushButton, NavigationAvatarWidget and NavigationUserCard.
func (w *NavigationWidget) registerInteractionEvents() {
	w.OnEnterEvent(func(super func(event *qt.QEvent), event *qt.QEvent) { w.handleEnter() })
	w.OnLeaveEvent(func(super func(event *qt.QEvent), event *qt.QEvent) { w.handleLeave() })
	w.OnMousePressEvent(func(super func(event *qt.QMouseEvent), event *qt.QMouseEvent) {
		super(event)
		w.handlePress()
	})
	w.OnMouseReleaseEvent(func(super func(event *qt.QMouseEvent), event *qt.QMouseEvent) {
		super(event)
		w.handleRelease()
	})
}

// SetCompacted toggles compacted mode (40x36 vs expanded width x36).
func (w *NavigationWidget) SetCompacted(isCompacted bool) {
	if isCompacted == w.IsCompacted {
		return
	}
	w.IsCompacted = isCompacted
	if isCompacted {
		w.SetFixedSize2(40, 36)
	} else {
		w.SetFixedSize2(navigationWidgetExpandWidth, 36)
	}
	w.Update()
}

// SetSelected sets the selected state; non-selectable widgets ignore it.
func (w *NavigationWidget) SetSelected(isSelected bool) {
	if !w.IsSelectable {
		return
	}
	w.IsSelected = isSelected
	w.IsAboutSelected = false
	w.Update()
	w.emitSelectedChanged(isSelected)
}

// TextColor returns the theme-appropriate text color (borrowed).
func (w *NavigationWidget) TextColor() *qt.QColor {
	if common.IsDarkTheme() {
		return w.darkTextColor
	}
	return w.lightTextColor
}

// SetLightTextColor sets the light-theme text color.
func (w *NavigationWidget) SetLightTextColor(color interface{}) {
	if w.lightTextColor != nil {
		w.lightTextColor.Delete()
	}
	w.lightTextColor = coerceColor(color)
	w.Update()
}

// SetDarkTextColor sets the dark-theme text color.
func (w *NavigationWidget) SetDarkTextColor(color interface{}) {
	if w.darkTextColor != nil {
		w.darkTextColor.Delete()
	}
	w.darkTextColor = coerceColor(color)
	w.Update()
}

// SetTextColor sets both light and dark text colors.
func (w *NavigationWidget) SetTextColor(light, dark interface{}) {
	w.SetLightTextColor(light)
	w.SetDarkTextColor(dark)
}

// SetAboutSelected sets the transient "about to be selected" state.
func (w *NavigationWidget) SetAboutSelected(selected bool) {
	w.IsAboutSelected = selected
	w.Update()
}

// IndicatorRect returns the indicator geometry (caller owns the QRectF). It is
// indented by the widget's left margin so nested tree items place the vertical
// bar after their depth offset (mirrors navigation_widget.py indicatorRect()).
// Widgets that override the geometry install it through indicatorRectFunc.
func (w *NavigationWidget) IndicatorRect() *qt.QRectF {
	if w.indicatorRectFunc != nil {
		return w.indicatorRectFunc()
	}
	left, _ := w.margins()
	return qt.NewQRectF4(float64(left), 10, 3, 16)
}

// SetIndicatorColor sets the light/dark indicator colors.
func (w *NavigationWidget) SetIndicatorColor(light, dark interface{}) {
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

// ---------------------------------------------------------------------------
// NavigationPushButton

// NavigationPushButton is a selectable navigation item with an icon and text.
type NavigationPushButton struct {
	*NavigationWidget
	icon             interface{}
	text             string
	canDrawIndicator func() bool
}

// NewNavigationPushButton builds a navigation push button.
func NewNavigationPushButton(icon interface{}, text string, isSelectable bool, parent *qt.QWidget) *NavigationPushButton {
	w := &NavigationPushButton{
		NavigationWidget: newNavigationWidget(isSelectable, parent),
		icon:             icon,
		text:             text,
	}
	w.canDrawIndicator = func() bool { return w.IsSelected }
	common.SetFont(w.QWidget, 14, int(qt.QFont__Normal))
	w.registerInteractionEvents()
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) { w.paint() })
	return w
}

// Text returns the button text.
func (w *NavigationPushButton) Text() string { return w.text }

// SetText sets the button text.
func (w *NavigationPushButton) SetText(text string) {
	w.text = text
	w.Update()
}

// Icon returns the icon as a *qt.QIcon.
func (w *NavigationPushButton) Icon() *qt.QIcon { return common.ToQIcon(w.icon) }

// SetIcon sets the icon source.
func (w *NavigationPushButton) SetIcon(icon interface{}) {
	w.icon = icon
	w.Update()
}

// paint draws the icon, text, hover/selected background and the indicator.
func (w *NavigationPushButton) paint() {
	painter := qt.NewQPainter2(w.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__Antialiasing | qt.QPainter__TextAntialiasing | qt.QPainter__SmoothPixmapTransform)
	painter.SetPenWithStyle(qt.NoPen)

	if w.IsPressed {
		painter.SetOpacity(0.7)
	}
	if !w.IsEnabled() {
		painter.SetOpacity(0.4)
	}

	c := 0
	if common.IsDarkTheme() {
		c = 255
	}
	left, right := w.margins()

	origin := qt.NewQPoint()
	global := w.MapToGlobal(origin)
	origin.Delete()
	size := w.Size()
	globalRect := qt.NewQRect4(global.X(), global.Y(), size.Width(), size.Height())

	drawBG := false
	bgAlpha := 10
	if w.canDrawIndicator() {
		if w.IsEnter {
			bgAlpha = 6
		} else {
			bgAlpha = 10
		}
		drawBG = true
	} else if w.IsEnabled() {
		hovered := w.IsEnter
		if hovered {
			cursor := qt.QCursor_Pos()
			hovered = globalRect.ContainsWithQPoint(cursor)
		}
		if hovered || w.IsAboutSelected {
			drawBG = true
			if w.IsAboutSelected {
				bgAlpha = 6
			} else {
				bgAlpha = 10
			}
		}
	}
	globalRect.Delete()

	if drawBG {
		bg := qt.NewQColor11(c, c, c, bgAlpha)
		brush := qt.NewQBrush3(bg)
		bg.Delete()
		defer brush.Delete()
		painter.SetBrush(brush)
		rect := w.Rect()
		painter.DrawRoundedRect3(rect, 5, 5)

	}

	if w.canDrawIndicator() {
		indicator := fallbackThemeColor(w.lightIndicatorColor, w.darkIndicatorColor)
		brush := qt.NewQBrush3(indicator)
		indicator.Delete()
		defer brush.Delete()
		painter.SetBrush(brush)
		indicatorRect := w.IndicatorRect()
		painter.DrawRoundedRect(indicatorRect, 1.5, 1.5)
		indicatorRect.Delete()
	}

	iconRect := qt.NewQRectF4(11.5+float64(left), 10, 16, 16)
	common.DrawIcon(w.icon, painter, iconRect)
	iconRect.Delete()

	if w.IsCompacted {
		painter.End()
		return
	}

	painter.SetFont(w.Font())
	painter.SetPen(w.TextColor())
	textLeft := left + 16
	if !iconIsNull(w.icon) {
		textLeft = 44 + left
	}
	textRect := qt.NewQRectF4(float64(textLeft), 0, float64(w.Width()-13-textLeft-right), float64(w.Height()))
	painter.DrawText5(textRect, int(qt.AlignVCenter), w.text)
	textRect.Delete()
	painter.End()
}

// ---------------------------------------------------------------------------
// NavigationToolButton

// NavigationToolButton is a square tool button used for the menu/return buttons.
type NavigationToolButton struct{ *NavigationPushButton }

// NewNavigationToolButton builds a navigation tool button.
func NewNavigationToolButton(icon interface{}, parent *qt.QWidget) *NavigationToolButton {
	w := &NavigationToolButton{NavigationPushButton: NewNavigationPushButton(icon, "", false, parent)}
	w.SetFixedSize2(40, 36)
	return w
}

// SetCompacted keeps a tool button always square.
func (w *NavigationToolButton) SetCompacted(isCompacted bool) {
	w.SetFixedSize2(40, 36)
}

// ---------------------------------------------------------------------------
// NavigationSeparator

// NavigationSeparator is a thin horizontal divider between navigation items.
type NavigationSeparator struct{ *NavigationWidget }

// NewNavigationSeparator builds a navigation separator.
func NewNavigationSeparator(parent *qt.QWidget) *NavigationSeparator {
	w := &NavigationSeparator{NavigationWidget: newNavigationWidget(false, parent)}
	navigationSeparatorPtrs[w.QWidget.UnsafePointer()] = true
	w.setCompactedSeparator(true)
	w.installPaint()
	return w
}

func (w *NavigationSeparator) setCompactedSeparator(isCompacted bool) {
	if isCompacted {
		w.SetFixedSize2(48, 3)
	} else {
		w.SetFixedSize2(navigationWidgetExpandWidth+10, 3)
	}
	w.Update()
}

// SetCompacted resizes the separator.
func (w *NavigationSeparator) SetCompacted(isCompacted bool) {
	w.IsCompacted = isCompacted
	w.setCompactedSeparator(isCompacted)
}

func (w *NavigationSeparator) installPaint() {
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		c := 0
		if common.IsDarkTheme() {
			c = 255
		}
		color := qt.NewQColor11(c, c, c, 15)
		pen := qt.NewQPen3(color)
		color.Delete()
		defer pen.Delete()
		pen.SetCosmetic(true)
		painter.SetPenWithPen(pen)
		painter.DrawLine2(0, 1, w.Width(), 1)
		painter.End()
	})
}

// ---------------------------------------------------------------------------
// NavigationItemHeader

// NavigationItemHeader is a non-clickable header used to group navigation items.
type NavigationItemHeader struct {
	*NavigationWidget
	text         string
	targetHeight int
	heightAni    *common.ProgressAnimation
}

// NewNavigationItemHeader builds a header.
func NewNavigationItemHeader(text string, parent *qt.QWidget) *NavigationItemHeader {
	w := &NavigationItemHeader{
		NavigationWidget: newNavigationWidget(false, parent),
		text:             text,
		targetHeight:     30,
	}
	common.SetFont(w.QWidget, 12, int(qt.QFont__Normal))
	w.lightTextColor = qt.NewQColor3(96, 96, 96)
	w.darkTextColor = qt.NewQColor3(160, 160, 160)
	w.SetCursor(qt.NewQCursor2(qt.ArrowCursor))
	w.SetFixedHeight(0)
	w.OnMousePressEvent(func(super func(event *qt.QMouseEvent), event *qt.QMouseEvent) {
		event.Ignore()
	})
	w.OnMouseReleaseEvent(func(super func(event *qt.QMouseEvent), event *qt.QMouseEvent) {
		event.Ignore()
	})
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) { w.paint() })
	return w
}

// Text returns the header text.
func (w *NavigationItemHeader) Text() string { return w.text }

// SetText sets the header text.
func (w *NavigationItemHeader) SetText(text string) {
	w.text = text
	w.Update()
}

// SetCompacted animates the header height (150ms OutQuad): it collapses to 0
// and hides when compacted, or reveals and grows to targetHeight when expanded.
func (w *NavigationItemHeader) SetCompacted(isCompacted bool) {
	if isCompacted == w.IsCompacted {
		return
	}
	w.IsCompacted = isCompacted
	w.stopHeightAni()

	if isCompacted {
		w.SetFixedWidth(40)
		w.animateHeight(0, true)
	} else {
		w.SetFixedWidth(navigationWidgetExpandWidth)
		w.SetVisible(true)
		w.animateHeight(w.targetHeight, false)
	}
	w.Update()
}

func (w *NavigationItemHeader) animateHeight(target int, hideOnFinish bool) {
	from := w.Height()
	curve := qt.NewQEasingCurve3(qt.QEasingCurve__OutQuad)
	w.heightAni = common.NewProgressAnimation(150, curve)
	curve.Delete()
	w.heightAni.OnProgress(func(t float64) {
		w.SetFixedHeight(from + int(float64(target-from)*t+0.5))
		w.Update()
	})
	w.heightAni.OnFinished(func() {
		w.SetFixedHeight(target)
		if hideOnFinish {
			w.SetVisible(false)
		}
		w.Update()
	})
	w.heightAni.Start()
}

func (w *NavigationItemHeader) stopHeightAni() {
	if w.heightAni != nil {
		w.heightAni.Stop()
		w.heightAni.Delete()
		w.heightAni = nil
	}
}

func (w *NavigationItemHeader) paint() {
	if w.Height() == 0 || !w.IsVisible() {
		return
	}
	painter := qt.NewQPainter2(w.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__Antialiasing | qt.QPainter__TextAntialiasing)

	opacity := 1.0
	if w.targetHeight > 0 {
		opacity = math.Min(1.0, float64(w.Height())/float64(w.targetHeight))
	}
	painter.SetOpacity(opacity)

	if !w.IsCompacted {
		painter.SetFont(w.Font())
		painter.SetPen(w.TextColor())
		rect := qt.NewQRectF4(16, 0, float64(w.Width()-16), float64(w.Height()))
		painter.DrawText5(rect, int(qt.AlignLeft|qt.AlignVCenter), w.text)
		rect.Delete()
	}
	painter.End()
}

// ---------------------------------------------------------------------------
// navigationTreeItem (NavigationTreeItem)

// navigationTreeItem is the clickable header row of a NavigationTreeWidget. It
// draws the icon/text plus a drop-down arrow and emits itemClicked.
type navigationTreeItem struct {
	*NavigationPushButton
	treeWidget     *NavigationTreeWidget
	arrowAngle     float64
	rotateAni      *common.ProgressAnimation
	itemClickedSig bool2Signal
}

func newNavigationTreeItem(icon interface{}, text string, isSelectable bool, tree *NavigationTreeWidget) *navigationTreeItem {
	w := &navigationTreeItem{
		NavigationPushButton: NewNavigationPushButton(icon, text, isSelectable, tree.QWidget),
		treeWidget:           tree,
	}
	w.NavigationWidget.marginsFunc = func() (int, int) {
		return w.treeWidget.NodeDepth * 28, 20 * boolToInt(len(w.treeWidget.treeChildren) > 0)
	}
	w.canDrawIndicator = w.canDrawIndicatorTree
	w.registerInteractionEvents()
	// Replace the push-button paint handler with the tree-item paint handler.
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
		w.paint()
		w.drawDropDownArrow()
	})
	w.OnMouseReleaseEvent(func(super func(event *qt.QMouseEvent), event *qt.QMouseEvent) {
		super(event)
		w.handleRelease()
		pos := event.Pos()
		arrowRect := qt.NewQRectF4(float64(w.Width()-30), 8, 20, 20)
		clickArrow := arrowRect.Contains2(float64(pos.X()), float64(pos.Y()))

		arrowRect.Delete()
		w.itemClickedSig.emit(true, clickArrow && !w.treeWidget.IsLeaf())
		w.Update()
	})
	return w
}

// OnItemClicked registers the itemClicked listener.
func (w *navigationTreeItem) OnItemClicked(f func(bool, bool)) { w.itemClickedSig.connect(f) }

func (w *navigationTreeItem) canDrawIndicatorTree() bool {
	p := w.treeWidget
	if p.IsLeaf() || p.IsSelected {
		return p.IsSelected
	}
	for _, child := range p.treeChildren {
		if child.itemWidget.canDrawIndicator() && !child.IsVisible() {
			return true
		}
	}
	return false
}

// SetExpanded rotates the drop-down arrow over 150ms (OutCubic).
func (w *navigationTreeItem) SetExpanded(isExpanded bool) {
	w.stopRotateAni()

	target := 0.0
	if isExpanded {
		target = 180.0
	}
	from := w.arrowAngle
	curve := qt.NewQEasingCurve3(qt.QEasingCurve__OutCubic)
	w.rotateAni = common.NewProgressAnimation(150, curve)
	curve.Delete()
	w.rotateAni.OnProgress(func(t float64) {
		w.arrowAngle = from + (target-from)*t
		w.Update()
	})
	w.rotateAni.OnFinished(func() {
		w.arrowAngle = target
		w.Update()
	})
	w.rotateAni.Start()
}

func (w *navigationTreeItem) stopRotateAni() {
	if w.rotateAni != nil {
		w.rotateAni.Stop()
		w.rotateAni.Delete()
		w.rotateAni = nil
	}
}

func (w *navigationTreeItem) drawDropDownArrow() {
	if w.IsCompacted || w.treeWidget.IsLeaf() {
		return
	}
	painter := qt.NewQPainter2(w.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__Antialiasing)
	painter.SetPenWithStyle(qt.NoPen)
	if w.IsPressed {
		painter.SetOpacity(0.7)
	}
	if !w.IsEnabled() {
		painter.SetOpacity(0.4)
	}
	painter.Translate2(float64(w.Width()-20), 18)
	painter.Rotate(w.arrowAngle)
	rect := qt.NewQRectF4(-5, -5, 9.6, 9.6)
	common.ChevronDown.Render(painter, rect, common.ThemeAuto)
	rect.Delete()
	painter.End()
}

// ---------------------------------------------------------------------------
// NavigationTreeWidgetBase

// NavigationTreeWidgetBase is the abstract tree-node interface. NavigationTreeWidget
// is the only concrete implementation in this port.
type NavigationTreeWidgetBase interface {
	IsRoot() bool
	IsLeaf() bool
	SetExpanded(isExpanded bool, ani bool)
	SaveExpandState()
	RestoreExpandState(ani bool)
	SetRememberExpandState(remember bool)
	ChildItems() []*NavigationTreeWidget
	AddChild(child *NavigationTreeWidget)
	InsertChild(index int, child *NavigationTreeWidget)
	RemoveChild(child *NavigationTreeWidget)
}

// ---------------------------------------------------------------------------
// NavigationTreeWidget

// NavigationTreeWidget is a collapsible tree node shown in a navigation panel.
type NavigationTreeWidget struct {
	*NavigationWidget
	treeChildren        []*NavigationTreeWidget
	IsExpanded          bool
	icon                interface{}
	rememberExpandState bool
	wasExpanded         bool
	itemWidget          *navigationTreeItem
	vBoxLayout          *qt.QVBoxLayout
	expandAni           *common.ProgressAnimation
	expandedSig         voidSignal
}

// NewNavigationTreeWidget builds a tree widget.
func NewNavigationTreeWidget(icon interface{}, text string, isSelectable bool, parent *qt.QWidget) *NavigationTreeWidget {
	w := &NavigationTreeWidget{
		NavigationWidget: newNavigationWidget(isSelectable, parent),
		icon:             icon,
	}
	w.itemWidget = newNavigationTreeItem(icon, text, isSelectable, w)
	w.vBoxLayout = qt.NewQVBoxLayout(w.QWidget)
	w.vBoxLayout.SetSpacing(4)
	w.vBoxLayout.SetContentsMargins(0, 0, 0, 0)
	w.vBoxLayout.AddWidget3(w.itemWidget.QWidget, 0, qt.AlignTop)
	w.itemWidget.OnItemClicked(func(triggerByUser, clickArrow bool) { w.onClicked(triggerByUser, clickArrow) })
	w.SetAttribute(qt.WA_TranslucentBackground)
	return w
}

// OnExpanded registers the expanded listener.
func (w *NavigationTreeWidget) OnExpanded(f func()) { w.expandedSig.connect(f) }

func (w *NavigationTreeWidget) emitExpanded() { w.expandedSig.emit() }

// AddChild appends a child node.
func (w *NavigationTreeWidget) AddChild(child *NavigationTreeWidget) {
	w.InsertChild(-1, child)
}

// Text returns the item text.
func (w *NavigationTreeWidget) Text() string { return w.itemWidget.Text() }

// SetText sets the item text.
func (w *NavigationTreeWidget) SetText(text string) { w.itemWidget.SetText(text) }

// Icon returns the item icon.
func (w *NavigationTreeWidget) Icon() *qt.QIcon { return w.itemWidget.Icon() }

// SetIcon sets the item icon.
func (w *NavigationTreeWidget) SetIcon(icon interface{}) { w.itemWidget.SetIcon(icon) }

// TextColor returns the item text color.
func (w *NavigationTreeWidget) TextColor() *qt.QColor { return w.itemWidget.TextColor() }

// SetLightTextColor sets the item light text color.
func (w *NavigationTreeWidget) SetLightTextColor(color interface{}) {
	w.itemWidget.SetLightTextColor(color)
}

// SetDarkTextColor sets the item dark text color.
func (w *NavigationTreeWidget) SetDarkTextColor(color interface{}) {
	w.itemWidget.SetDarkTextColor(color)
}

// SetTextColor sets the item light/dark text colors.
func (w *NavigationTreeWidget) SetTextColor(light, dark interface{}) {
	w.NavigationWidget.SetTextColor(light, dark)
	w.itemWidget.SetTextColor(light, dark)
}

// SetIndicatorColor sets the item light/dark indicator colors.
func (w *NavigationTreeWidget) SetIndicatorColor(light, dark interface{}) {
	w.NavigationWidget.SetIndicatorColor(light, dark)
	w.itemWidget.SetIndicatorColor(light, dark)
}

// SetFont sets the font on the tree widget and its item row.
func (w *NavigationTreeWidget) SetFont(font *qt.QFont) {
	w.QWidget.SetFont(font)
	w.itemWidget.SetFont(font)
}

// Clone deep-copies this tree widget (used by the flyout menu).
func (w *NavigationTreeWidget) Clone() *NavigationTreeWidget {
	root := NewNavigationTreeWidget(w.icon, w.Text(), w.IsSelectable, w.ParentWidget())
	root.SetSelected(w.IsSelected)
	size := w.Size()
	root.SetFixedSize2(size.Width(), size.Height())

	root.SetTextColor(w.lightTextColor, w.darkTextColor)
	root.SetIndicatorColor(w.itemWidget.lightIndicatorColor, w.itemWidget.darkIndicatorColor)
	root.NodeDepth = w.NodeDepth
	root.OnClicked(func(v bool) { w.emitClicked(v) })
	w.OnSelectedChanged(func(v bool) { root.SetSelected(v) })
	for _, child := range w.treeChildren {
		root.AddChild(child.Clone())
	}
	return root
}

// SuitableWidth returns the preferred width for the fully expanded node.
func (w *NavigationTreeWidget) SuitableWidth() int {
	left, right := w.itemWidget.margins()
	textLeft := left + 29
	if !iconIsNull(w.itemWidget.icon) {
		textLeft = 57 + left
	}
	fm := w.itemWidget.FontMetrics()
	bound := fm.BoundingRectWithText(w.itemWidget.text)
	tw := bound.Width()
	return textLeft + tw + right
}

// InsertChild inserts a child node at index.
func (w *NavigationTreeWidget) InsertChild(index int, child *NavigationTreeWidget) {
	for _, c := range w.treeChildren {
		if c == child {
			return
		}
	}
	child.TreeParent = w
	child.NodeDepth = w.NodeDepth + 1
	child.SetVisible(w.IsExpanded)

	if index < 0 || index > len(w.treeChildren) {
		index = len(w.treeChildren)
	}
	w.treeChildren = append(w.treeChildren, nil)
	copy(w.treeChildren[index+1:], w.treeChildren[index:])
	w.treeChildren[index] = child
	// The item widget occupies layout index 0, so children start at index 1.
	w.vBoxLayout.InsertWidget3(index+1, child.QWidget, 0, qt.AlignTop)

	if w.IsExpanded {
		w.SetFixedSize2(w.sizeHintWidth(), w.sizeHintHeight())
		for p := w.TreeParent; p != nil; p = p.TreeParent {
			p.SetFixedSize2(p.sizeHintWidth(), p.sizeHintHeight())
		}
	}
	w.Update()
}

// RemoveChild removes a child node.
func (w *NavigationTreeWidget) RemoveChild(child *NavigationTreeWidget) {
	for i, c := range w.treeChildren {
		if c == child {
			w.treeChildren = append(w.treeChildren[:i], w.treeChildren[i+1:]...)
			break
		}
	}
	w.vBoxLayout.RemoveWidget(child.QWidget)
	w.SetFixedSize2(w.sizeHintWidth(), w.sizeHintHeight())
}

// ChildItems returns the child nodes.
func (w *NavigationTreeWidget) ChildItems() []*NavigationTreeWidget { return w.treeChildren }

// resizeAndUpdateAncestors sets this node's fixed size and re-sizes every
// ancestor to its own size hint. A nested node's height change must propagate
// up the tree, otherwise a collapsed/expanded child leaves its parent at a
// stale fixed size and the sibling below overlaps the grown node (the reported
// third-level overlap).
func (w *NavigationTreeWidget) resizeAndUpdateAncestors(width, height int) {
	w.SetFixedSize2(width, height)
	for p := w.TreeParent; p != nil; p = p.TreeParent {
		p.SetFixedSize2(p.sizeHintWidth(), p.sizeHintHeight())
	}
}

// SetExpanded expands/collapses the node. With ani it animates the fixed size
// from the current geometry to the size hint (120ms OutQuad) and emits expanded
// on completion; otherwise it applies the target size directly.
func (w *NavigationTreeWidget) SetExpanded(isExpanded bool, ani bool) {
	if isExpanded == w.IsExpanded {
		return
	}
	w.IsExpanded = isExpanded
	w.itemWidget.SetExpanded(isExpanded)
	for _, child := range w.treeChildren {
		child.SetVisible(isExpanded)
		child.SetFixedSize2(child.sizeHintWidth(), child.sizeHintHeight())
	}
	w.stopExpandAni()

	tw := w.sizeHintWidth()
	th := w.sizeHintHeight()
	if ani {
		cw, ch := w.Width(), w.Height()
		curve := qt.NewQEasingCurve3(qt.QEasingCurve__OutQuad)
		w.expandAni = common.NewProgressAnimation(120, curve)
		curve.Delete()
		w.expandAni.OnProgress(func(t float64) {
			w.resizeAndUpdateAncestors(cw+int(float64(tw-cw)*t+0.5), ch+int(float64(th-ch)*t+0.5))
		})
		w.expandAni.OnFinished(func() {
			w.resizeAndUpdateAncestors(tw, th)
			w.emitExpanded()
		})
		w.expandAni.Start()
	} else {
		w.resizeAndUpdateAncestors(tw, th)
	}
}

func (w *NavigationTreeWidget) stopExpandAni() {
	if w.expandAni != nil {
		w.expandAni.Stop()
		w.expandAni.Delete()
		w.expandAni = nil
	}
}

// IsRoot reports whether this is a root node.
func (w *NavigationTreeWidget) IsRoot() bool { return w.TreeParent == nil }

// IsLeaf reports whether this node has no children.
func (w *NavigationTreeWidget) IsLeaf() bool { return len(w.treeChildren) == 0 }

// SetSelected selects the node and its item row.
func (w *NavigationTreeWidget) SetSelected(isSelected bool) {
	w.NavigationWidget.SetSelected(isSelected)
	w.itemWidget.SetSelected(isSelected)
}

// SetCompacted compacts the node and its item row.
func (w *NavigationTreeWidget) SetCompacted(isCompacted bool) {
	w.NavigationWidget.SetCompacted(isCompacted)
	w.itemWidget.SetCompacted(isCompacted)
}

// SetAboutSelected sets the transient selection state on the item row.
func (w *NavigationTreeWidget) SetAboutSelected(selected bool) {
	w.IsAboutSelected = selected
	w.itemWidget.SetAboutSelected(selected)
}

func (w *NavigationTreeWidget) onClicked(triggerByUser, clickArrow bool) {
	if !w.IsCompacted {
		if w.IsSelectable && !w.IsSelected && !clickArrow {
			w.SetExpanded(true, true)
		} else {
			w.SetExpanded(!w.IsExpanded, true)
		}
	}
	if !clickArrow || w.IsCompacted {
		w.emitClicked(triggerByUser)
	}
}

// SetRememberExpandState toggles expand-state persistence.
func (w *NavigationTreeWidget) SetRememberExpandState(remember bool) {
	w.rememberExpandState = remember
}

// SaveExpandState saves the current expanded state.
func (w *NavigationTreeWidget) SaveExpandState() {
	if w.rememberExpandState {
		w.wasExpanded = w.IsExpanded
	} else {
		w.wasExpanded = false
	}
}

// RestoreExpandState restores the saved expanded state.
func (w *NavigationTreeWidget) RestoreExpandState(ani bool) {
	if w.wasExpanded {
		w.SetExpanded(true, ani)
	}
}

func (w *NavigationTreeWidget) sizeHintWidth() int {
	s := w.SizeHint()

	return s.Width()
}

func (w *NavigationTreeWidget) sizeHintHeight() int {
	s := w.SizeHint()

	return s.Height()
}

// ---------------------------------------------------------------------------
// NavigationAvatarWidget

// NavigationAvatarWidget shows an avatar plus an optional name.
type NavigationAvatarWidget struct {
	*NavigationWidget
	name   string
	avatar *widgets.AvatarWidget
}

func newNavigationAvatarWidget(name string, avatar interface{}, parent *qt.QWidget) *NavigationAvatarWidget {
	w := &NavigationAvatarWidget{
		NavigationWidget: newNavigationWidget(false, parent),
		name:             name,
	}
	w.avatar = widgets.NewAvatarWidget(w.QWidget)
	w.avatar.SetRadius(12)
	w.avatar.SetText(name)
	w.avatar.Move(8, 6)
	common.SetFont(w.QWidget, 14, int(qt.QFont__Normal))
	if avatar != nil {
		w.SetAvatar(avatar)
	}
	w.registerInteractionEvents()
	return w
}

// NewNavigationAvatarWidget builds an avatar widget.
func NewNavigationAvatarWidget(name string, avatar interface{}, parent *qt.QWidget) *NavigationAvatarWidget {
	w := newNavigationAvatarWidget(name, avatar, parent)
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) { w.paint() })
	return w
}

// SetName sets the display name.
func (w *NavigationAvatarWidget) SetName(name string) {
	w.name = name
	w.avatar.SetText(name)
	w.Update()
}

// SetAvatar sets the avatar image.
func (w *NavigationAvatarWidget) SetAvatar(avatar interface{}) {
	w.avatar.SetImage(avatar)
	w.avatar.SetRadius(12)
	w.Update()
}

func (w *NavigationAvatarWidget) paint() {
	painter := qt.NewQPainter2(w.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__SmoothPixmapTransform | qt.QPainter__Antialiasing)
	if w.IsPressed {
		painter.SetOpacity(0.7)
	}
	w.drawBackground(painter)
	w.drawText(painter)
	painter.End()
}

func (w *NavigationAvatarWidget) drawText(painter *qt.QPainter) {
	if w.IsCompacted {
		return
	}
	painter.SetPen(w.TextColor())
	painter.SetFont(w.Font())
	rect := qt.NewQRect4(44, 0, 255, 36)
	painter.DrawText6(rect, int(qt.AlignVCenter), w.name)
	rect.Delete()
}

func (w *NavigationAvatarWidget) drawBackground(painter *qt.QPainter) {
	if !w.IsEnter {
		return
	}
	c := 0
	if common.IsDarkTheme() {
		c = 255
	}
	color := qt.NewQColor11(c, c, c, 10)
	brush := qt.NewQBrush3(color)
	color.Delete()
	defer brush.Delete()
	painter.SetBrush(brush)
	painter.SetPenWithStyle(qt.NoPen)
	rect := w.Rect()
	painter.DrawRoundedRect3(rect, 5, 5)

}

// ---------------------------------------------------------------------------
// NavigationUserCard

// NavigationUserCard is an avatar card with title and subtitle text. The avatar
// radius and text opacity are animated together when compact mode toggles.
type NavigationUserCard struct {
	*NavigationAvatarWidget
	title             string
	subtitle          string
	titleSize         int
	subtitleSize      int
	subtitleColor     *qt.QColor
	textOpacity       float64
	animationDuration int
	anim              *common.ProgressAnimation
}

// NewNavigationUserCard builds a user card.
func NewNavigationUserCard(parent *qt.QWidget) *NavigationUserCard {
	w := &NavigationUserCard{
		NavigationAvatarWidget: newNavigationAvatarWidget("", nil, parent),
		titleSize:              14,
		subtitleSize:           12,
		animationDuration:      250,
	}
	w.SetFixedSize2(40, 36)
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) { w.paint() })
	return w
}

// SetAvatarIcon sets the avatar from a fluent icon.
func (w *NavigationUserCard) SetAvatarIcon(icon interface{}) {
	ic := common.ToQIcon(icon)
	pm := ic.Pixmap2(64, 64)
	w.avatar.SetImage(pm)
	w.Update()
}

// SetAvatarBackgroundColor sets the avatar background color.
func (w *NavigationUserCard) SetAvatarBackgroundColor(light, dark *qt.QColor) {
	w.avatar.SetBackgroundColor(light, dark)
	w.Update()
}

// Title returns the card title.
func (w *NavigationUserCard) Title() string { return w.title }

// SetTitle sets the card title.
func (w *NavigationUserCard) SetTitle(title string) {
	w.title = title
	w.SetName(title)
	w.Update()
}

// Subtitle returns the card subtitle.
func (w *NavigationUserCard) Subtitle() string { return w.subtitle }

// SetSubtitle sets the card subtitle.
func (w *NavigationUserCard) SetSubtitle(subtitle string) {
	w.subtitle = subtitle
	w.Update()
}

// SetTitleFontSize sets the title font size.
func (w *NavigationUserCard) SetTitleFontSize(size int) {
	w.titleSize = size
	w.Update()
}

// SetSubtitleFontSize sets the subtitle font size.
func (w *NavigationUserCard) SetSubtitleFontSize(size int) {
	w.subtitleSize = size
	w.Update()
}

// SetAnimationDuration sets the (now unused) animation duration.
func (w *NavigationUserCard) SetAnimationDuration(duration int) { w.animationDuration = duration }

// SetCompacted toggles compact/expanded mode, animating the avatar radius and
// text opacity together (OutCubic over animationDuration).
func (w *NavigationUserCard) SetCompacted(isCompacted bool) {
	if isCompacted == w.IsCompacted {
		return
	}
	w.IsCompacted = isCompacted
	w.stopAnim()

	fromRadius := float64(w.avatar.Radius())
	fromOpacity := w.textOpacity
	var toRadius, toOpacity float64
	if isCompacted {
		w.SetFixedSize2(40, 36)
		toRadius, toOpacity = 12, 0.0
	} else {
		w.SetFixedSize2(navigationWidgetExpandWidth, 80)
		toRadius, toOpacity = 32, 1.0
	}

	curve := qt.NewQEasingCurve3(qt.QEasingCurve__OutCubic)
	w.anim = common.NewProgressAnimation(w.animationDuration, curve)
	curve.Delete()
	w.anim.OnProgress(func(t float64) {
		w.avatar.SetRadius(int(math.Round(fromRadius + (toRadius-fromRadius)*t)))
		w.textOpacity = fromOpacity + (toOpacity-fromOpacity)*t
		w.updateAvatarPosition()
		w.Update()
	})
	w.anim.OnFinished(func() {
		w.avatar.SetRadius(int(toRadius))
		w.textOpacity = toOpacity
		w.updateAvatarPosition()
		w.Update()
	})
	w.anim.Start()
}

func (w *NavigationUserCard) stopAnim() {
	if w.anim != nil {
		w.anim.Stop()
		w.anim.Delete()
		w.anim = nil
	}
}

func (w *NavigationUserCard) paint() {
	painter := qt.NewQPainter2(w.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__SmoothPixmapTransform | qt.QPainter__Antialiasing | qt.QPainter__TextAntialiasing)
	if w.IsPressed {
		painter.SetOpacity(0.7)
	}
	w.drawBackground(painter)
	if !w.IsCompacted && w.textOpacity > 0 {
		w.drawText(painter)
	}
	painter.End()
}

func (w *NavigationUserCard) drawText(painter *qt.QPainter) {
	textX := 16 + w.avatar.Radius()*2 + 12
	textWidth := w.Width() - textX - 16

	titleFont := common.GetFont(w.titleSize, int(qt.QFont__Bold))
	painter.SetFont(titleFont)
	titleFont.Delete()
	c := cloneColor(w.TextColor())
	c.SetAlpha(int(255 * w.textOpacity))
	painter.SetPen(c)
	c.Delete()

	titleY := w.Height()/2 - 2
	titleRect := qt.NewQRectF4(float64(textX), 0, float64(textWidth), float64(titleY))
	painter.DrawText5(titleRect, int(qt.AlignLeft|qt.AlignBottom), w.title)
	titleRect.Delete()

	if w.subtitle != "" {
		subFont := common.GetFont(w.subtitleSize, int(qt.QFont__Normal))
		painter.SetFont(subFont)
		subFont.Delete()
		sc := w.subtitleColor
		if sc == nil {
			sc = w.TextColor()
		}
		c2 := cloneColor(sc)
		c2.SetAlpha(int(150 * w.textOpacity))
		painter.SetPen(c2)
		c2.Delete()
		subtitleY := w.Height()/2 + 2
		subRect := qt.NewQRectF4(float64(textX), float64(subtitleY), float64(textWidth), float64(w.Height()-subtitleY))
		painter.DrawText5(subRect, int(qt.AlignLeft|qt.AlignTop), w.subtitle)
		subRect.Delete()
	}
}

func (w *NavigationUserCard) updateAvatarPosition() {
	if w.IsCompacted {
		w.avatar.Move(8, 6)
	} else {
		w.avatar.Move(16, (w.Height()-w.avatar.Height())/2)
	}
}

// ---------------------------------------------------------------------------
// NavigationIndicator

// NavigationIndicator is the sliding selection indicator. It reproduces the
// WinUI squash-and-stretch ScaleSlideAnimation by driving a QRectF transition
// (see slideRectAnimator) and repainting the indicator geometry each frame.
type NavigationIndicator struct {
	*qt.QWidget
	lightColor     *qt.QColor
	darkColor      *qt.QColor
	curRect        *qt.QRectF
	anim           *slideRectAnimator
	aniFinishedSig voidSignal
}

// NewNavigationIndicator builds an indicator.
func NewNavigationIndicator(parent *qt.QWidget) *NavigationIndicator {
	w := &NavigationIndicator{QWidget: qt.NewQWidget(parent)}
	w.lightColor = qt.NewQColor()
	w.darkColor = qt.NewQColor()
	w.curRect = qt.NewQRectF4(0, 0, 3, 16)
	w.Resize(3, 16)
	w.SetAttribute(qt.WA_TransparentForMouseEvents)
	w.SetAttribute(qt.WA_TranslucentBackground)
	w.Hide()
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) { w.paint() })
	return w
}

// OnAniFinished registers the aniFinished listener.
func (w *NavigationIndicator) OnAniFinished(f func()) { w.aniFinishedSig.connect(f) }

// StartAnimation slides the indicator from startRect to endRect and emits
// aniFinished on completion.
func (w *NavigationIndicator) StartAnimation(startRect, endRect *qt.QRectF, useCrossFade bool) {
	w.StopAnimation()

	w.curRect.SetRect(startRect.X(), startRect.Y(), startRect.Width(), startRect.Height())
	w.SetGeometryWithGeometry(w.curRect.ToRect()) // GoGC-armed — do NOT Delete
	w.Show()
	// Raise above the navigation items / scroll area so the slide is never
	// covered by a button or the scroll viewport (z-order fix).
	w.Raise()

	w.anim = newSlideRectAnimator(w.curRect, false, func() {
		w.SetGeometryWithGeometry(w.curRect.ToRect()) // GoGC-armed — do NOT Delete
		w.Update()
	}, func() {
		w.anim = nil
		w.aniFinishedSig.emit()
	})
	w.anim.start(endRect, useCrossFade)
}

// StopAnimation cancels any in-flight slide and hides the indicator.
func (w *NavigationIndicator) StopAnimation() {
	if w.anim != nil {
		w.anim.stop()
		w.anim = nil
	}
	w.Hide()
}

// SetIndicatorColor sets the light/dark indicator colors.
func (w *NavigationIndicator) SetIndicatorColor(light, dark interface{}) {
	if w.lightColor != nil {
		w.lightColor.Delete()
	}
	if w.darkColor != nil {
		w.darkColor.Delete()
	}
	w.lightColor = coerceColor(light)
	w.darkColor = coerceColor(dark)
	w.Update()
}

func (w *NavigationIndicator) paint() {
	painter := qt.NewQPainter2(w.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__Antialiasing)
	painter.SetPenWithStyle(qt.NoPen)
	color := fallbackThemeColor(w.lightColor, w.darkColor)
	brush := qt.NewQBrush3(color)
	color.Delete()
	defer brush.Delete()
	painter.SetBrush(brush)
	rect := w.Rect()
	painter.DrawRoundedRect3(rect, 1.5, 1.5)

	painter.End()
}

// ---------------------------------------------------------------------------
// NavigationFlyoutMenu

// NavigationFlyoutMenu shows a cloned tree's children inside a flyout when a
// collapsed root item is clicked.
type NavigationFlyoutMenu struct {
	*widgets.ScrollArea
	view         *qt.QWidget
	treeWidget   *NavigationTreeWidget
	treeChildren []*NavigationTreeWidget
	vBoxLayout   *qt.QVBoxLayout
	expandedSig  voidSignal
}

// NewNavigationFlyoutMenu builds a flyout menu for tree.
func NewNavigationFlyoutMenu(tree *NavigationTreeWidget, parent *qt.QWidget) *NavigationFlyoutMenu {
	w := &NavigationFlyoutMenu{
		ScrollArea: widgets.NewScrollArea(parent),
		treeWidget: tree,
	}
	w.view = qt.NewQWidget(w.QWidget)
	w.vBoxLayout = qt.NewQVBoxLayout(w.view)
	w.SetWidget(w.view)
	w.SetWidgetResizable(true)
	w.SetHorizontalScrollBarPolicy(qt.ScrollBarAlwaysOff)
	w.EnableTransparentBackground()
	w.vBoxLayout.SetSpacing(5)
	w.vBoxLayout.SetContentsMargins(5, 8, 5, 8)

	for _, child := range tree.treeChildren {
		node := child.Clone()
		node.OnExpanded(func() { w.adjustViewSize(true) })
		w.treeChildren = append(w.treeChildren, node)
		w.vBoxLayout.AddWidget(node.QWidget)
	}
	w.initNodeChildren(w.treeChildren)
	w.adjustViewSize(false)
	return w
}

// OnExpanded registers the expanded listener.
func (w *NavigationFlyoutMenu) OnExpanded(f func()) { w.expandedSig.connect(f) }

func (w *NavigationFlyoutMenu) initNodeChildren(children []*NavigationTreeWidget) {
	for _, c := range children {
		c.NodeDepth--
		c.SetCompacted(false)
		if c.IsLeaf() {
			c.OnClicked(func(v bool) { w.Window().Close() })
		}
		w.initNodeChildren(c.treeChildren)
	}
}

func (w *NavigationFlyoutMenu) adjustViewSize(emit bool) {
	width := w.suitableWidth()
	for _, node := range w.visibleTreeNodes() {
		node.SetFixedWidth(width - 10)
		node.itemWidget.SetFixedWidth(width - 10)
	}
	sh := w.view.SizeHint()
	w.view.SetFixedSize2(width, sh.Height())

	parentHeight := 0
	if w.Window() != nil && w.Window().ParentWidget() != nil {
		parentHeight = w.Window().ParentWidget().Height()
	}
	h := parentHeight - 48
	if h < 0 || w.view.Height() < h {
		h = w.view.Height()
	}
	w.SetFixedSize2(width, h)
	if emit {
		w.expandedSig.emit()
	}
}

func (w *NavigationFlyoutMenu) suitableWidth() int {
	width := 0
	for _, node := range w.visibleTreeNodes() {
		if !node.IsHidden() {
			if sw := node.SuitableWidth() + 10; sw > width {
				width = sw
			}
		}
	}
	var window *qt.QWidget
	if w.Window() != nil {
		window = w.Window().ParentWidget()
	}
	maxW := 0
	if window != nil {
		maxW = window.Width()/2 - 25
	}
	if width > maxW {
		width = maxW
	}
	return width + 10
}

func (w *NavigationFlyoutMenu) visibleTreeNodes() []*NavigationTreeWidget {
	var nodes []*NavigationTreeWidget
	queue := append([]*NavigationTreeWidget(nil), w.treeChildren...)
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		nodes = append(nodes, node)
		for _, child := range node.treeChildren {
			if !child.IsHidden() {
				queue = append(queue, child)
			}
		}
	}
	return nodes
}
