package widgets

import (
	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// CommandButton is a fluent command-bar tool button. It embeds a QToolButton
// and self-draws its icon and text (the native QToolButton text/icon painting is
// left unused so the two do not paint on top of each other).
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
	b.SetToolTip(action.ToolTip())
	b.SetEnabled(action.IsEnabled())
	b.SetCheckable(action.IsCheckable())
	b.SetChecked(action.IsChecked())
}

func (b *CommandButton) isIconOnly() bool {
	if b._text == "" {
		return true
	}
	style := b.ToolButtonStyle()
	return style == qt.ToolButtonIconOnly || style == qt.ToolButtonFollowStyle
}

// SizeHint returns the computed preferred size.
func (b *CommandButton) SizeHint() *qt.QSize { return b.calcSizeHint() }

func (b *CommandButton) calcSizeHint() *qt.QSize {
	if b.isIconOnly() {
		if b._isTight {
			return qt.NewQSize2(36, 34)
		}
		return qt.NewQSize2(48, 34)
	}

	fm := b.FontMetrics() // GoGC-armed — do NOT Delete
	tw := fm.Width(b._text)

	switch b.ToolButtonStyle() {
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

		style := b.ToolButtonStyle()
		isz := b.IconSize()
		iw := isz.Width()
		ih := isz.Height()

		switch style {
		case qt.ToolButtonIconOnly, qt.ToolButtonFollowStyle:
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

	for _, it := range visibles {
		qw := commandBarWidget(it)
		qw.Show()
		qw.Move(x, (h-qw.Height())/2)
		x += qw.Width() + w.spacing
	}

	if len(w._hiddenActions) > 0 || len(visibles) < len(w._widgets) {
		w.moreButton.Show()
		w.moreButton.Move(x, (h-w.moreButton.Height())/2)
	}

	for _, it := range w._widgets[len(visibles):] {
		commandBarWidget(it).Hide()
		w._hiddenWidgets = append(w._hiddenWidgets, it)
	}
}

func (w *CommandBar) visibleWidgets() []interface{} {
	if w.suitableWidth() <= w.Width() {
		return w._widgets
	}

	width := w.moreButton.Width()
	for i, it := range w._widgets {
		width += commandBarWidget(it).Width()
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
	total := 0
	n := len(w._widgets)
	for _, it := range w._widgets {
		total += commandBarWidget(it).Width()
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
	qw := commandBarWidget(widget)
	qw.SetParent(w.QWidget)
	qw.Show()

	if index < 0 || index > len(w._widgets) {
		w._widgets = append(w._widgets, widget)
	} else {
		w._widgets = append(w._widgets, nil)
		copy(w._widgets[index+1:], w._widgets[index:])
		w._widgets[index] = widget
	}

	maxH := 0
	for _, it := range w._widgets {
		if h := commandBarWidget(it).Height(); h > maxH {
			maxH = h
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

func commandBarWidget(w interface{}) *qt.QWidget {
	switch v := w.(type) {
	case *CommandButton:
		return v.QWidget
	case *CommandSeparator:
		return v.QWidget
	case *qt.QWidget:
		return v
	default:
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
