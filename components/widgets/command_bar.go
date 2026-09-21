package widgets

import (
	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// CommandButton is a fluent command-bar tool button. It embeds a QToolButton
// and self-draws its icon and text (the native QToolButton text/icon painting is
// left unused so the two do not paint on top of each other).
//
// A button is drawn as an icon-only button when it carries no text or when
// SetIconOnly(true) forces it, whatever the tool button style of the bar is; the
// text of its action stays available as the tool tip and in the more-actions
// menu.
//
// Constructors
//   - NewCommandButton(parent *qt.QWidget)
type CommandButton struct {
	*qt.QToolButton
	_icon     interface{}
	_text     string
	_action   *qt.QAction
	_isTight  bool
	isPressed bool

	// _iconOnly overrides the derived icon-only state once SetIconOnly was called.
	_iconOnly    bool
	_iconOnlySet bool

	// _toolTip overrides the tool tip taken from the action once SetToolTip was
	// called (an empty string suppresses the tool tip entirely).
	_toolTip    string
	_hasToolTip bool
}

// NewCommandButton builds a command button.
func NewCommandButton(parent *qt.QWidget) *CommandButton {
	b := &CommandButton{QToolButton: qt.NewQToolButton(parent)}
	b._icon = qt.NewQIcon()
	b.SetCheckable(false)
	b.SetToolButtonStyle(qt.ToolButtonIconOnly)
	// CommandButton is the Go equivalent of Python's TransparentToggleToolButton:
	// the fluent button QSS (transparent background + hover/press states) is
	// applied via the #transparentToggleToolButton objectName, and the icon/text
	// are self-drawn on top.
	b.SetObjectName("transparentToggleToolButton")
	common.FluentStyleSheet(common.FluentButton).Apply(b.QWidget, common.ThemeAuto)
	common.SetFont(b.QWidget, 12, 400)
	b.installEvents()
	b.applyPreferredSize()
	return b
}

// Text returns the command button text (kept separately from the native
// QToolButton text to avoid double painting).
func (b *CommandButton) Text() string { return b._text }

// SetText sets the command button text and recomputes its preferred size.
func (b *CommandButton) SetText(text string) {
	b._text = text
	b.applyPreferredSize()
	b.Update()
}

// SetIcon sets the icon source (*qt.QIcon, string or FluentIconBase).
func (b *CommandButton) SetIcon(icon interface{}) {
	b._icon = icon
	b.Update()
}

// Icon returns the current icon as a *qt.QIcon.
func (b *CommandButton) Icon() *qt.QIcon { return common.ToQIcon(b._icon) }

// SetTight toggles the tight layout.
func (b *CommandButton) SetTight(isTight bool) {
	b._isTight = isTight
	b.applyPreferredSize()
	b.Update()
}

// IsTight reports whether the button uses the tight layout.
func (b *CommandButton) IsTight() bool { return b._isTight }

// SetToolButtonStyle sets the tool button style and recomputes the size.
func (b *CommandButton) SetToolButtonStyle(style qt.ToolButtonStyle) {
	b.QToolButton.SetToolButtonStyle(style)
	b.applyPreferredSize()
}

// SetAction binds a QAction to the button.
func (b *CommandButton) SetAction(action *qt.QAction) {
	b._action = action
	b.onActionChanged()

	// The action outlives the button whenever the button's tree is deleted (a
	// command-bar flyout deletes its whole bar when it closes, while the actions
	// stay alive in the page that created them), so every callback is dropped once
	// the button is gone. Without the guard a theme switch — which re-renders each
	// action icon and therefore fires changed() — calls into freed widget memory.
	alive := trackWidget(b.OnDestroyed)
	b.OnClicked(func() {
		if alive.ok() {
			action.Trigger()
		}
	})
	action.OnToggled(func(checked bool) {
		if alive.ok() {
			b.SetChecked(checked)
		}
	})
	action.OnChanged(func() {
		if alive.ok() {
			b.onActionChanged()
		}
	})
}

// SetToolTip overrides the tool tip of the button. An empty string suppresses the
// tool tip entirely — an icon-only button otherwise falls back to the text of its
// action, because QAction.toolTip() answers with the action text while no explicit
// tool tip is set:
//
//	bar.AddIconAction(action).SetToolTip("")   // icon only, no hover text
//
// A non-empty tool tip is kept when the action changes.
func (b *CommandButton) SetToolTip(tooltip string) {
	b._toolTip = tooltip
	b._hasToolTip = true
	b.QToolButton.SetToolTip(tooltip)
}

// ToolTip returns the tool tip currently shown by the button.
func (b *CommandButton) ToolTip() string { return b.QToolButton.ToolTip() }

// applyToolTip takes the tool tip of the bound action unless the caller set one.
func (b *CommandButton) applyToolTip(action *qt.QAction) {
	if b._hasToolTip {
		b.QToolButton.SetToolTip(b._toolTip)
		return
	}
	b.QToolButton.SetToolTip(action.ToolTip())
}

// Action returns the bound action (nil when none).
func (b *CommandButton) Action() *qt.QAction { return b._action }

func (b *CommandButton) onActionChanged() {
	action := b._action
	if action == nil {
		return
	}
	// Keep the fluent icon source when the action has one: the button paints its
	// own icon and reverses it while checked (the checked state sits on the
	// accent background), which the action's pre-rendered QIcon cannot express.
	if src := common.FluentIconOf(action); src != nil {
		b.SetIcon(src)
	} else {
		b.SetIcon(action.Icon())
	}
	b.SetText(action.Text())
	b.applyToolTip(action)
	b.SetEnabled(action.IsEnabled())
	b.SetCheckable(action.IsCheckable())
	b.SetChecked(action.IsChecked())
}

// SetIconOnly forces the button into the icon-only layout (nothing but the icon
// is drawn) or restores the derived layout. A button whose action carries no text
// is icon-only anyway; this is what makes a labeled action render as a pure icon
// button — its text is still used as the tool tip and by the more-actions menu.
func (b *CommandButton) SetIconOnly(iconOnly bool) {
	b._iconOnly = iconOnly
	b._iconOnlySet = true
	b.applyPreferredSize()
	b.Update()
}

// IsIconOnly reports whether the button draws its icon without text.
func (b *CommandButton) IsIconOnly() bool { return b.isIconOnly() }

func (b *CommandButton) isIconOnly() bool {
	if b._iconOnlySet {
		return b._iconOnly
	}
	if b._text == "" {
		return true
	}
	style := b.ToolButtonStyle()
	if style != qt.ToolButtonIconOnly && style != qt.ToolButtonFollowStyle {
		return false
	}
	// The icon-only style still needs an icon to draw: an action without one (see
	// common.NewActionText) would render as an empty button, so its text is drawn
	// instead.
	return b.hasIcon()
}

// hasIcon reports whether the button has an icon to draw.
func (b *CommandButton) hasIcon() bool {
	switch icon := b._icon.(type) {
	case nil:
		return false
	case *qt.QIcon:
		return icon != nil && !icon.IsNull()
	case string:
		return icon != ""
	default:
		return b._icon != nil
	}
}

// effectiveStyle returns the layout the button paints and measures with: an
// icon-only button ignores the tool button style of the bar, an icon-only style
// falls back to the icon-beside-text layout once SetIconOnly(false) asks for the
// text back, and a button without an icon is text only — the icon slot of the
// beside/under layouts would stay empty and push the text off center.
func (b *CommandButton) effectiveStyle() qt.ToolButtonStyle {
	if b.isIconOnly() {
		return qt.ToolButtonIconOnly
	}
	if !b.hasIcon() {
		return qt.ToolButtonTextOnly
	}
	style := b.ToolButtonStyle()
	if style == qt.ToolButtonIconOnly || style == qt.ToolButtonFollowStyle {
		return qt.ToolButtonTextBesideIcon
	}
	return style
}

// SizeHint returns the computed preferred size.
func (b *CommandButton) SizeHint() *qt.QSize { return b.calcSizeHint() }

func (b *CommandButton) calcSizeHint() *qt.QSize {
	switch b.effectiveStyle() {
	case qt.ToolButtonIconOnly:
		if b._isTight {
			return qt.NewQSize2(36, 34)
		}
		return qt.NewQSize2(48, 34)
	}

	fm := b.FontMetrics() // GoGC-armed — do NOT Delete
	tw := fm.Width(b._text)

	switch b.effectiveStyle() {
	case qt.ToolButtonTextBesideIcon:
		return qt.NewQSize2(tw+47, 34)
	case qt.ToolButtonTextOnly:
		return qt.NewQSize2(tw+32, 34)
	default:
		return qt.NewQSize2(tw+32, 50)
	}
}

func (b *CommandButton) applyPreferredSize() {
	s := b.calcSizeHint()
	defer s.Delete()
	b.SetFixedSize(s)
}

func (b *CommandButton) commandTextColor() *qt.QColor {
	dark := common.IsDarkTheme()
	if !b.IsChecked() {
		if dark {
			return qt.NewQColor3(255, 255, 255)
		}
		return qt.NewQColor3(0, 0, 0)
	}
	if dark {
		return qt.NewQColor3(0, 0, 0)
	}
	return qt.NewQColor3(255, 255, 255)
}

func (b *CommandButton) iconTheme() common.Theme {
	if b.IsChecked() {
		return reversedTheme()
	}
	return common.ThemeAuto
}

func (b *CommandButton) installEvents() {
	b.OnPressed(func() {
		b.isPressed = true
		b.Update()
	})
	b.OnReleased(func() {
		b.isPressed = false
		b.Update()
	})
	b.OnPaintEvent(func(super func(ev *qt.QPaintEvent), ev *qt.QPaintEvent) {
		super(ev)
		painter := qt.NewQPainter2(b.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing | qt.QPainter__SmoothPixmapTransform)

		pen := b.commandTextColor()
		defer pen.Delete()
		painter.SetPen(pen)

		if !b.IsEnabled() {
			painter.SetOpacity(0.43)
		} else if b.isPressed {
			painter.SetOpacity(0.63)
		}

		style := b.effectiveStyle()
		isz := b.IconSize()
		iw := isz.Width()
		ih := isz.Height()

		// An icon-only button centers its icon whatever the tool button style of
		// the bar is: the text layouts below would leave the icon in the corner
		// reserved for a label that is never drawn.
		switch style {
		case qt.ToolButtonIconOnly:
			rect := qt.NewQRectF4(float64((b.Width()-iw))/2, float64((b.Height()-ih))/2, float64(iw), float64(ih))
			defer rect.Delete()
			renderFluentIcon(b._icon, painter, rect, b.iconTheme())
		case qt.ToolButtonTextOnly:
			r := b.Rect()
			painter.DrawText6(r, int(qt.AlignCenter), b._text)

		case qt.ToolButtonTextBesideIcon:
			rect := qt.NewQRectF4(11, float64((b.Height()-ih))/2, float64(iw), float64(ih))
			defer rect.Delete()
			renderFluentIcon(b._icon, painter, rect, b.iconTheme())

			textRect := qt.NewQRectF4(26, 0, float64(b.Width()-26), float64(b.Height()))
			defer textRect.Delete()
			tr := textRect.ToRect() // GoGC-armed — do NOT Delete
			painter.DrawText6(tr, int(qt.AlignCenter), b._text)
		case qt.ToolButtonTextUnderIcon:
			rect := qt.NewQRectF4(float64((b.Width()-iw))/2, 9, float64(iw), float64(ih))
			defer rect.Delete()
			renderFluentIcon(b._icon, painter, rect, b.iconTheme())

			textRect := qt.NewQRectF4(0, float64(ih+13), float64(b.Width()), float64(b.Height()-ih-13))
			defer textRect.Delete()
			tr := textRect.ToRect() // GoGC-armed — do NOT Delete
			painter.DrawText6(tr, int(qt.AlignHCenter|qt.AlignTop), b._text)
		}
		painter.End()
	})
}

// MoreActionsButton is the "more actions" button shown at the end of a
// command bar when there is not enough space for every action.
type MoreActionsButton struct {
	*CommandButton
}

// NewMoreActionsButton builds a more-actions button.
func NewMoreActionsButton(parent *qt.QWidget) *MoreActionsButton {
	b := &MoreActionsButton{CommandButton: NewCommandButton(parent)}
	b.SetIcon(common.More)
	b.SetFixedSize2(40, 34)
	return b
}

// SizeHint returns the fixed more-button size.
func (b *MoreActionsButton) SizeHint() *qt.QSize { return qt.NewQSize2(40, 34) }

// ClearState clears the hover/pressed visual state of the button.
func (b *MoreActionsButton) ClearState() {
	b.SetAttribute2(qt.WA_UnderMouse, false)
	b.isPressed = false

	// Mirror Python MoreActionsButton.clearState: dispatch a HoverLeave event so
	// Qt fully drops the under-mouse/hover style state. Resetting WA_UnderMouse
	// alone can leave the button painted with its :hover/:pressed QSS background
	// until the next real mouse move.
	pos := qt.NewQPointF3(-1, -1)
	defer pos.Delete()
	oldPos := qt.NewQPointF3(0, 0)
	defer oldPos.Delete()
	ev := qt.NewQHoverEvent(qt.QEvent__HoverLeave, pos, oldPos)
	defer ev.Delete()
	qt.QCoreApplication_SendEvent(b.QObject, ev.QEvent)

	b.Update()
}

// CommandSeparator is a thin vertical line used to separate command buttons.
type CommandSeparator struct {
	*qt.QWidget
}

// NewCommandSeparator builds a command separator.
func NewCommandSeparator(parent *qt.QWidget) *CommandSeparator {
	w := &CommandSeparator{QWidget: qt.NewQWidget(parent)}
	w.SetFixedSize2(9, 34)
	w.OnPaintEvent(func(super func(ev *qt.QPaintEvent), ev *qt.QPaintEvent) {
		super(ev)
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		var c *qt.QColor
		if common.IsDarkTheme() {
			c = qt.NewQColor11(255, 255, 255, 21)
		} else {
			c = qt.NewQColor11(0, 0, 0, 15)
		}
		defer c.Delete()
		painter.SetPen(c)
		painter.DrawLine2(5, 2, 5, w.Height()-2)
		painter.End()
	})
	return w
}

// CommandMenu is the round menu shown by a command bar's more-actions button.
type CommandMenu struct {
	*RoundMenu
}

// NewCommandMenu builds a command menu.
func NewCommandMenu(parent *qt.QWidget) *CommandMenu {
	m := &CommandMenu{RoundMenu: NewRoundMenu("", parent)}
	m.SetItemHeight(32)
	m.view.SetIconSize(qt.NewQSize2(16, 16))
	return m
}

// CommandViewMenu is the round menu shown by a CommandBarView's more-actions
// button. It carries the #commandListWidget objectName and dropDown/long dynamic
// properties so the menu QSS attaches it to the command bar shape (mirrors
// command_bar.py CommandViewMenu).
type CommandViewMenu struct {
	*CommandMenu
}

// NewCommandViewMenu builds a command view menu.
func NewCommandViewMenu(parent *qt.QWidget) *CommandViewMenu {
	m := &CommandViewMenu{CommandMenu: NewCommandMenu(parent)}
	m.view.SetObjectName("commandListWidget")
	// Re-polish so the #commandListWidget border rule takes effect after the
	// objectName change.
	m.view.SetStyle(qt.QApplication_Style())
	return m
}

// SetDropDown sets the drop-down direction and long shape properties that drive
// the #commandListWidget[dropDown=…][long=…] QSS border-radius rules.
func (m *CommandViewMenu) SetDropDown(down, long bool) {
	m.view.SetProperty("dropDown", qt.NewQVariant11(down))
	m.view.SetProperty("long", qt.NewQVariant11(long))
	m.view.SetStyle(qt.QApplication_Style())
	m.view.QWidget.Update()
}

// CommandBar is a horizontal bar of command buttons with manual layout. When
// there is not enough room, overflow actions are folded into a more-actions
// button that opens a CommandMenu.
type CommandBar struct {
	*qt.QFrame
	_widgets       []interface{}
	_hiddenWidgets []interface{}
	_hiddenActions []*qt.QAction

	_menuAnimation   MenuAnimationType
	_toolButtonStyle qt.ToolButtonStyle
	_iconSize        *qt.QSize
	_isButtonTight   bool
	spacing          int

	moreButton *MoreActionsButton
	// showMenuFunc dispatches the more-button click. CommandViewBar replaces it
	// to show a CommandViewMenu attached to the surrounding CommandBarView.
	showMenuFunc func()
}

// NewCommandBar builds a command bar.
func NewCommandBar(parent *qt.QWidget) *CommandBar {
	w := &CommandBar{QFrame: qt.NewQFrame(parent)}
	w._menuAnimation = MenuAnimationDropDown
	w._toolButtonStyle = qt.ToolButtonIconOnly
	w._iconSize = qt.NewQSize2(16, 16)
	w._isButtonTight = false
	w.spacing = 4

	w.moreButton = NewMoreActionsButton(w.QWidget)
	w.showMenuFunc = w.showMoreActionsMenu
	w.moreButton.OnClicked(func() { w.showMenuFunc() })
	w.moreButton.Hide()

	common.SetFont(w.QWidget, 12, 400)
	w.SetAttribute(qt.WA_TranslucentBackground)
	w.OnResizeEvent(func(super func(ev *qt.QResizeEvent), ev *qt.QResizeEvent) {
		super(ev)
		w.UpdateGeometry()
	})
	return w
}

// SetSpacing sets the spacing between widgets.
func (w *CommandBar) SetSpacing(spacing int) {
	if spacing == w.spacing {
		return
	}
	w.spacing = spacing
	w.UpdateGeometry()
}

// Spacing returns the spacing between widgets.
func (w *CommandBar) Spacing() int { return w.spacing }

// AddAction adds an action and returns the created command button.
func (w *CommandBar) AddAction(action *qt.QAction) *CommandButton {
	if containsAction(w.Actions(), action) {
		return nil
	}
	button := w.createButton(action)
	w.insertWidgetToLayout(-1, button)
	w.QWidget.AddAction(action)
	return button
}

// AddIconAction adds an action as a pure icon button: only the icon is drawn, no
// text. The action text is kept as the button's tool tip and is still shown by the
// more-actions menu, so an icon-only button stays reachable and labelled:
//
//	bar.AddIconAction(common.NewActionFluentIcon(common.Settings, "设置", nil).QAction)
func (w *CommandBar) AddIconAction(action *qt.QAction) *CommandButton {
	button := w.AddAction(action)
	if button != nil {
		button.SetIconOnly(true)
	}
	return button
}

// AddActions adds multiple actions.
func (w *CommandBar) AddActions(actions []*qt.QAction) {
	for _, a := range actions {
		w.AddAction(a)
	}
}

// AddHiddenAction adds an action that is only shown in the more-actions menu.
func (w *CommandBar) AddHiddenAction(action *qt.QAction) {
	if containsAction(w.Actions(), action) {
		return
	}
	w._hiddenActions = append(w._hiddenActions, action)
	w.UpdateGeometry()
	w.QWidget.AddAction(action)
}

// AddHiddenActions adds multiple hidden actions.
func (w *CommandBar) AddHiddenActions(actions []*qt.QAction) {
	for _, a := range actions {
		w.AddHiddenAction(a)
	}
}

// InsertAction inserts an action before another action.
func (w *CommandBar) InsertAction(before, action *qt.QAction) *CommandButton {
	if !containsAction(w.Actions(), before) {
		return nil
	}
	index := -1
	for i, a := range w.Actions() {
		if a == before {
			index = i
			break
		}
	}
	button := w.createButton(action)
	w.insertWidgetToLayout(index, button)
	w.QWidget.InsertAction(before, action)
	return button
}

// AddStretch appends a flexible spacer. It takes the space that is left after the
// visible items are placed, so every action added afterwards is pushed to the right
// edge of the bar (the QBoxLayout.AddStretch behaviour):
//
//	bar.AddAction(openAction)    // left aligned
//	bar.AddStretch()
//	bar.AddAction(themeAction)   // right aligned
//
// Several stretches share the leftover space equally.
func (w *CommandBar) AddStretch() {
	w.insertWidgetToLayout(-1, commandBarStretch{})
}

// InsertStretch inserts a flexible spacer at index.
func (w *CommandBar) InsertStretch(index int) {
	w.insertWidgetToLayout(index, commandBarStretch{})
}

// AddSeparator adds a separator to the end of the bar.
func (w *CommandBar) AddSeparator() {
	w.InsertSeparator(-1)
}

// InsertSeparator inserts a separator at index.
func (w *CommandBar) InsertSeparator(index int) {
	w.insertWidgetToLayout(index, NewCommandSeparator(w.QWidget))
}

// AddWidget adds an arbitrary widget to the bar.
func (w *CommandBar) AddWidget(widget *qt.QWidget) {
	w.insertWidgetToLayout(-1, widget)
}

// RemoveAction removes the command button bound to the action.
func (w *CommandBar) RemoveAction(action *qt.QAction) {
	if !containsAction(w.Actions(), action) {
		return
	}
	for i, it := range w._widgets {
		if b, ok := it.(*CommandButton); ok && b.Action() == action {
			w._widgets = append(w._widgets[:i], w._widgets[i+1:]...)
			b.Hide()
			b.DeleteLater()
			break
		}
	}
	w.UpdateGeometry()
}

// RemoveWidget removes a widget from the bar.
func (w *CommandBar) RemoveWidget(widget *qt.QWidget) {
	if widget == nil {
		return
	}
	for i, it := range w._widgets {
		if commandBarWidget(it) == widget {
			w._widgets = append(w._widgets[:i], w._widgets[i+1:]...)
			w.UpdateGeometry()
			return
		}
	}
}

// RemoveHiddenAction removes a hidden action.
func (w *CommandBar) RemoveHiddenAction(action *qt.QAction) {
	for i, a := range w._hiddenActions {
		if a == action {
			w._hiddenActions = append(w._hiddenActions[:i], w._hiddenActions[i+1:]...)
			return
		}
	}
}

// SetToolButtonStyle sets the tool button style of every command button.
func (w *CommandBar) SetToolButtonStyle(style qt.ToolButtonStyle) {
	if w.ToolButtonStyle() == style {
		return
	}
	w._toolButtonStyle = style
	for _, b := range w.CommandButtons() {
		b.SetToolButtonStyle(style)
	}
}

// ToolButtonStyle returns the tool button style.
func (w *CommandBar) ToolButtonStyle() qt.ToolButtonStyle { return w._toolButtonStyle }

// SetButtonTight toggles the tight layout of every command button.
func (w *CommandBar) SetButtonTight(isTight bool) {
	if w.IsButtonTight() == isTight {
		return
	}
	w._isButtonTight = isTight
	for _, b := range w.CommandButtons() {
		b.SetTight(isTight)
	}
	w.UpdateGeometry()
}

// IsButtonTight reports whether buttons use the tight layout.
func (w *CommandBar) IsButtonTight() bool { return w._isButtonTight }

// SetIconSize sets the icon size of every command button.
func (w *CommandBar) SetIconSize(size *qt.QSize) {
	if size.Width() == w._iconSize.Width() && size.Height() == w._iconSize.Height() {
		return
	}
	w._iconSize = qt.NewQSize2(size.Width(), size.Height())
	for _, b := range w.CommandButtons() {
		b.SetIconSize(w._iconSize)
	}
}

// IconSize returns the icon size.
func (w *CommandBar) IconSize() *qt.QSize { return w._iconSize }

// SetFont applies a font to the bar and every command button.
func (w *CommandBar) SetFont(font *qt.QFont) {
	w.QWidget.SetFont(font)
	for _, b := range w.CommandButtons() {
		b.SetFont(font)
		b.applyPreferredSize()
	}
	w.UpdateGeometry()
}

// CommandButtons returns the command buttons in layout order.
func (w *CommandBar) CommandButtons() []*CommandButton {
	var out []*CommandButton
	for _, it := range w._widgets {
		if b, ok := it.(*CommandButton); ok {
			out = append(out, b)
		}
	}
	return out
}

// SetMenuDropDown sets the animation direction of the more-actions menu.
func (w *CommandBar) SetMenuDropDown(down bool) {
	if down {
		w._menuAnimation = MenuAnimationDropDown
	} else {
		w._menuAnimation = MenuAnimationPullUp
	}
}

// IsMenuDropDown reports whether the more-actions menu drops down.
func (w *CommandBar) IsMenuDropDown() bool { return w._menuAnimation == MenuAnimationDropDown }

// SuitableWidth returns the width needed to show all widgets.
func (w *CommandBar) SuitableWidth() int { return w.suitableWidth() }

// ResizeToSuitableWidth fixes the bar width to show all widgets.
func (w *CommandBar) ResizeToSuitableWidth() {
	w.SetFixedWidth(w.suitableWidth())
}

// UpdateGeometry re-runs the manual layout, hiding overflow widgets behind the
// more-actions button.
func (w *CommandBar) UpdateGeometry() {
	w._hiddenWidgets = nil
	w.moreButton.Hide()

	visibles := w.visibleWidgets()
	m := w.ContentsMargins()
	x := m.Left()

	h := w.Height()

	// The flexible spacers take an equal share of the space that is left once
	// every visible item and the more-actions button are placed, which is what
	// pushes the items behind a stretch to the right edge of the bar.
	fixed, items, stretches := 0, 0, 0
	for _, it := range visibles {
		if qw := commandBarWidget(it); qw != nil {
			fixed += qw.Width()
			items++
		} else {
			stretches++
		}
	}
	spacing := 0
	if items+stretches > 1 {
		spacing = w.spacing * (items + stretches - 1)
	}
	showMore := len(w._hiddenActions) > 0 || len(visibles) < len(w._widgets)
	reserved := 0
	if showMore {
		reserved = w.moreButton.Width() + w.spacing
	}
	share := 0
	if stretches > 0 {
		share = (w.Width() - m.Left() - m.Right() - fixed - spacing - reserved) / stretches
		if share < 0 {
			share = 0
		}
	}

	for _, it := range visibles {
		qw := commandBarWidget(it)
		if qw == nil {
			x += share + w.spacing
			continue
		}
		qw.Show()
		qw.Move(x, (h-qw.Height())/2)
		x += qw.Width() + w.spacing
	}

	if showMore {
		w.moreButton.Show()
		w.moreButton.Move(x, (h-w.moreButton.Height())/2)
	}

	for _, it := range w._widgets[len(visibles):] {
		qw := commandBarWidget(it)
		if qw == nil {
			continue
		}
		qw.Hide()
		w._hiddenWidgets = append(w._hiddenWidgets, it)
	}
}

func (w *CommandBar) visibleWidgets() []interface{} {
	if w.suitableWidth() <= w.Width() {
		return w._widgets
	}

	width := w.moreButton.Width()
	for i, it := range w._widgets {
		if qw := commandBarWidget(it); qw != nil {
			width += qw.Width()
		}
		if i > 0 {
			width += w.spacing
		}
		if width > w.Width() {
			return w._widgets[:i]
		}
	}
	return w._widgets
}

func (w *CommandBar) suitableWidth() int {
	total, n := 0, 0
	for _, it := range w._widgets {
		if qw := commandBarWidget(it); qw != nil {
			total += qw.Width()
			n++
		}
	}
	if len(w._hiddenActions) > 0 {
		total += w.moreButton.Width()
		n++
	}
	if n > 0 {
		total += w.spacing * (n - 1)
	}
	return total
}

func (w *CommandBar) createButton(action *qt.QAction) *CommandButton {
	button := NewCommandButton(w.QWidget)
	button.SetAction(action)
	button.SetToolButtonStyle(w.ToolButtonStyle())
	button.SetTight(w.IsButtonTight())
	button.SetIconSize(w.IconSize())

	font := w.Font() // borrowed reference — do NOT Delete
	button.SetFont(font)

	button.applyPreferredSize()
	return button
}

func (w *CommandBar) insertWidgetToLayout(index int, widget interface{}) {
	// A stretch has no widget of its own; it only takes part in the layout pass.
	if qw := commandBarWidget(widget); qw != nil {
		qw.SetParent(w.QWidget)
		qw.Show()
	}

	if index < 0 || index > len(w._widgets) {
		w._widgets = append(w._widgets, widget)
	} else {
		w._widgets = append(w._widgets, nil)
		copy(w._widgets[index+1:], w._widgets[index:])
		w._widgets[index] = widget
	}

	maxH := 0
	for _, it := range w._widgets {
		if qw := commandBarWidget(it); qw != nil {
			if h := qw.Height(); h > maxH {
				maxH = h
			}
		}
	}
	w.SetFixedHeight(maxH)
	w.UpdateGeometry()
}

func (w *CommandBar) showMoreActionsMenu() {
	w.moreButton.ClearState()

	actions := append([]*qt.QAction{}, w._hiddenActions...)
	for i := len(w._hiddenWidgets) - 1; i >= 0; i-- {
		if b, ok := w._hiddenWidgets[i].(*CommandButton); ok {
			actions = append([]*qt.QAction{b.Action()}, actions...)
		}
	}

	menu := NewCommandMenu(w.QWidget)
	menu.AddActions(actions)

	// Right-align the menu so its right edge sits near the more button (mirrors
	// command_bar.py _showMoreActionsMenu). The menu width is known now because
	// AddActions ran AdjustSize on each insertion.
	x := -menu.Width() + menu.hBoxLayout.ContentsMargins().Right() + w.moreButton.Width() + 18
	var y int
	if w._menuAnimation == MenuAnimationDropDown {
		y = w.moreButton.Height()
	} else {
		y = -5
	}
	p := qt.NewQPoint2(x, y)
	pos := w.moreButton.MapToGlobal(p) // GoGC-armed — do NOT Delete
	p.Delete()
	menu.Exec(pos, w._menuAnimation)
}

func containsAction(actions []*qt.QAction, action *qt.QAction) bool {
	for _, a := range actions {
		if a == action {
			return true
		}
	}
	return false
}

// commandBarStretch is the flexible spacer of a command bar (see AddStretch). It
// owns no widget: the layout gives it the space that is left over once every
// visible item has been placed.
type commandBarStretch struct{}

func commandBarWidget(w interface{}) *qt.QWidget {
	switch v := w.(type) {
	case *CommandButton:
		return v.QWidget
	case *CommandSeparator:
		return v.QWidget
	case *qt.QWidget:
		return v
	default:
		// A commandBarStretch is a placeholder without a widget.
		return nil
	}
}

// CommandViewBar is the command bar used inside a CommandBarView flyout. Its
// more-actions menu is a CommandViewMenu that attaches to the flyout shape and
// uses fade-in popup animations (mirrors command_bar.py CommandViewBar).
type CommandViewBar struct {
	*CommandBar
	view *CommandBarView
}

// NewCommandViewBar builds a command view bar.
func NewCommandViewBar(parent *qt.QWidget) *CommandViewBar {
	w := &CommandViewBar{CommandBar: NewCommandBar(parent)}
	w.showMenuFunc = w.showMoreActionsMenu
	w.SetMenuDropDown(true)
	return w
}

// SetMenuDropDown sets the fade-in drop-down/pull-up menu animation.
func (w *CommandViewBar) SetMenuDropDown(down bool) {
	if down {
		w._menuAnimation = MenuAnimationFadeInDropDown
	} else {
		w._menuAnimation = MenuAnimationFadeInPullUp
	}
}

// IsMenuDropDown reports whether the menu animates as a fade-in drop-down.
func (w *CommandViewBar) IsMenuDropDown() bool {
	return w._menuAnimation == MenuAnimationFadeInDropDown
}

// showMoreActionsMenu shows a CommandViewMenu that is visually attached to the
// parent CommandBarView (mirrors command_bar.py CommandViewBar._showMoreActionsMenu).
func (w *CommandViewBar) showMoreActionsMenu() {
	w.moreButton.ClearState()

	actions := append([]*qt.QAction{}, w._hiddenActions...)
	for i := len(w._hiddenWidgets) - 1; i >= 0; i-- {
		if b, ok := w._hiddenWidgets[i].(*CommandButton); ok {
			actions = append([]*qt.QAction{b.Action()}, actions...)
		}
	}

	menu := NewCommandViewMenu(w.QWidget)
	menu.AddActions(actions)

	view := w.view
	if view != nil {
		view.SetMenuVisible(true)
		menu.OnClosed(func() { view.SetMenuVisible(false) })
	}

	long := false
	if view != nil {
		long = menu.view.Width() > view.Width()+5
	}
	menu.SetDropDown(w.IsMenuDropDown(), long)

	if view != nil && menu.view.Width() < view.Width() {
		menu.view.SetFixedWidth(view.Width())
		menu.AdjustSize()
	}

	x := -menu.Width() + menu.hBoxLayout.ContentsMargins().Right() + w.moreButton.Width() + 18
	var y int
	if w.IsMenuDropDown() {
		y = w.moreButton.Height()
	} else {
		y = -13
		// The pull-up menu attaches to the top edge of the view: drop the shadow
		// and reverse the layout margins so the menu hugs the command bar.
		menu.view.SetGraphicsEffect(nil)
		menu.hBoxLayout.SetContentsMargins(12, 20, 12, 8)
	}

	p := qt.NewQPoint2(x, y)
	pos := w.moreButton.MapToGlobal(p) // GoGC-armed — do NOT Delete
	p.Delete()
	menu.Exec(pos, w._menuAnimation)
}

// CommandBarView is a command bar container that self-draws a rounded
// background. In the Python port it derives from FlyoutViewBase (so it can be
// shown inside a Flyout popup); the Go port embeds FlyoutViewBase directly and
// keeps the extra "menu visible" shape painted here.
type CommandBarView struct {
	*FlyoutViewBase
	bar            *CommandViewBar
	hBoxLayout     *qt.QHBoxLayout
	_isMenuVisible bool
}

// NewCommandBarView builds a command bar view.
func NewCommandBarView(parent *qt.QWidget) *CommandBarView {
	w := &CommandBarView{FlyoutViewBase: NewFlyoutViewBase(parent)}
	w.bar = NewCommandViewBar(w.QWidget)
	w.bar.view = w
	w.hBoxLayout = qt.NewQHBoxLayout(w.QWidget)
	w.hBoxLayout.SetContentsMargins(6, 6, 6, 6)
	w.hBoxLayout.AddWidget(w.bar.QWidget)
	w.hBoxLayout.SetSizeConstraint(qt.QLayout__SetMinAndMaxSize)

	w.SetButtonTight(true)
	s := qt.NewQSize2(14, 14)
	w.SetIconSize(s)
	s.Delete()

	w._isMenuVisible = false
	w.installPaintEvent()
	return w
}

// SetMenuVisible toggles the menu-visible shape adjustment.
func (w *CommandBarView) SetMenuVisible(isVisible bool) {
	w._isMenuVisible = isVisible
	w.Update()
}

// AddWidget adds a widget to the inner bar.
func (w *CommandBarView) AddWidget(widget *qt.QWidget) { w.bar.AddWidget(widget) }

// SetSpacing sets the inner bar spacing.
func (w *CommandBarView) SetSpacing(spacing int) { w.bar.SetSpacing(spacing) }

// Spacing returns the inner bar spacing.
func (w *CommandBarView) Spacing() int { return w.bar.Spacing() }

// AddAction adds an action to the inner bar.
func (w *CommandBarView) AddAction(action *qt.QAction) *CommandButton { return w.bar.AddAction(action) }

// AddActions adds multiple actions to the inner bar.
func (w *CommandBarView) AddActions(actions []*qt.QAction) { w.bar.AddActions(actions) }

// AddHiddenAction adds a hidden action to the inner bar.
func (w *CommandBarView) AddHiddenAction(action *qt.QAction) { w.bar.AddHiddenAction(action) }

// AddHiddenActions adds multiple hidden actions to the inner bar.
func (w *CommandBarView) AddHiddenActions(actions []*qt.QAction) { w.bar.AddHiddenActions(actions) }

// InsertAction inserts an action into the inner bar.
func (w *CommandBarView) InsertAction(before, action *qt.QAction) *CommandButton {
	return w.bar.InsertAction(before, action)
}

// AddSeparator adds a separator to the inner bar.
func (w *CommandBarView) AddSeparator() { w.bar.AddSeparator() }

// InsertSeparator inserts a separator into the inner bar.
func (w *CommandBarView) InsertSeparator(index int) { w.bar.InsertSeparator(index) }

// RemoveAction removes an action from the inner bar.
func (w *CommandBarView) RemoveAction(action *qt.QAction) { w.bar.RemoveAction(action) }

// RemoveWidget removes a widget from the inner bar.
func (w *CommandBarView) RemoveWidget(widget *qt.QWidget) { w.bar.RemoveWidget(widget) }

// RemoveHiddenAction removes a hidden action from the inner bar.
func (w *CommandBarView) RemoveHiddenAction(action *qt.QAction) { w.bar.RemoveHiddenAction(action) }

// SetToolButtonStyle sets the inner bar tool button style.
func (w *CommandBarView) SetToolButtonStyle(style qt.ToolButtonStyle) {
	w.bar.SetToolButtonStyle(style)
}

// ToolButtonStyle returns the inner bar tool button style.
func (w *CommandBarView) ToolButtonStyle() qt.ToolButtonStyle { return w.bar.ToolButtonStyle() }

// SetButtonTight toggles the inner bar tight layout.
func (w *CommandBarView) SetButtonTight(isTight bool) { w.bar.SetButtonTight(isTight) }

// IsButtonTight reports whether the inner bar uses the tight layout.
func (w *CommandBarView) IsButtonTight() bool { return w.bar.IsButtonTight() }

// SetIconSize sets the inner bar icon size.
func (w *CommandBarView) SetIconSize(size *qt.QSize) { w.bar.SetIconSize(size) }

// IconSize returns the inner bar icon size.
func (w *CommandBarView) IconSize() *qt.QSize { return w.bar.IconSize() }

// SetFont applies a font to the inner bar.
func (w *CommandBarView) SetFont(font *qt.QFont) { w.bar.SetFont(font) }

// SetMenuDropDown sets the inner bar menu animation direction.
func (w *CommandBarView) SetMenuDropDown(down bool) { w.bar.SetMenuDropDown(down) }

// SuitableWidth returns the width needed to show every action.
func (w *CommandBarView) SuitableWidth() int {
	m := w.ContentsMargins()

	return m.Left() + m.Right() + w.bar.SuitableWidth()
}

// ResizeToSuitableWidth fixes the view width to show every action.
func (w *CommandBarView) ResizeToSuitableWidth() {
	w.bar.ResizeToSuitableWidth()
	w.SetFixedWidth(w.SuitableWidth())
}

// Actions returns the actions of the inner bar.
func (w *CommandBarView) Actions() []*qt.QAction { return w.bar.Actions() }

func (w *CommandBarView) installPaintEvent() {
	w.OnPaintEvent(func(super func(ev *qt.QPaintEvent), ev *qt.QPaintEvent) {
		super(ev)
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing)

		path := qt.NewQPainterPath()
		defer path.Delete()
		path.SetFillRule(qt.WindingFill)

		adjusted := w.Rect().Adjusted(1, 1, -1, -1) // GoGC-armed — do NOT Delete
		rectF := qt.NewQRectF5(adjusted)
		defer rectF.Delete()
		path.AddRoundedRect(rectF, 8, 8)

		if w._isMenuVisible {
			y := w.Height() - 10
			if !w.bar.IsMenuDropDown() {
				y = 1
			}
			path.AddRect2(1, float64(y), float64(w.Width()-2), 9)
		}

		var bg, border *qt.QColor
		if common.IsDarkTheme() {
			bg = qt.NewQColor3(40, 40, 40)
			border = qt.NewQColor3(56, 56, 56)
		} else {
			bg = qt.NewQColor3(248, 248, 248)
			border = qt.NewQColor3(233, 233, 233)
		}
		defer bg.Delete()
		defer border.Delete()

		brush := qt.NewQBrush3(bg)
		defer brush.Delete()
		painter.SetBrush(brush)
		painter.SetPen(border)

		simplified := path.Simplified() // GoGC-armed — do NOT Delete
		painter.DrawPath(simplified)
		painter.End()
	})
}
