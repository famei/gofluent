package widgets

import (
	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// ComboItem is a single combo box item.
type ComboItem struct {
	Text      string
	Icon      interface{}
	UserData  interface{}
	IsEnabled bool
}

// NewComboItem builds a combo box item.
func NewComboItem(text string, icon interface{}, userData interface{}) ComboItem {
	return ComboItem{Text: text, Icon: icon, UserData: userData, IsEnabled: true}
}

// insertComboItem inserts item at index into items (index is clamped).
func insertComboItem(items []ComboItem, index int, item ComboItem) []ComboItem {
	if index < 0 {
		index = 0
	}
	if index > len(items) {
		index = len(items)
	}
	return append(items[:index], append([]ComboItem{item}, items[index:]...)...)
}

// ComboBoxMenu is the drop-down menu of a combo box.
type ComboBoxMenu struct {
	*RoundMenu
}

// NewComboBoxMenu builds a combo box menu.
func NewComboBoxMenu(parent *qt.QWidget) *ComboBoxMenu {
	m := &ComboBoxMenu{RoundMenu: NewRoundMenu("", parent)}
	m.View().SetViewportMargins(0, 2, 0, 6)
	m.View().SetVerticalScrollBarPolicy(qt.ScrollBarAsNeeded)
	m.View().SetObjectName("comboListWidget")
	m.SetItemHeight(33)
	return m
}

// ComboBox is a fluent combo box backed by a QPushButton. The pop-up menu is a
// RoundMenu populated from the item list.
type ComboBox struct {
	*qt.QPushButton
	isHover         bool
	isPressed       bool
	isPlaceholder   bool
	items           []ComboItem
	currentIndex    int
	maxVisibleItems int
	dropMenu        *ComboBoxMenu
	placeholderText string
	arrowAni        *translateYAnimation

	OnCurrentIndexChanged func(int)
	OnCurrentTextChanged  func(string)
	OnActivated           func(int)
	OnTextActivated       func(string)
}

// NewComboBox builds a combo box.
func NewComboBox(parent *qt.QWidget) *ComboBox {
	w := &ComboBox{
		QPushButton:     qt.NewQPushButton(parent),
		currentIndex:    -1,
		maxVisibleItems: -1,
	}
	w.SetObjectName("comboBox")
	common.SetFont(w.QWidget, 14, 400)
	common.FluentStyleSheet(common.FluentComboBox).Apply(w.QWidget, common.ThemeAuto)
	w.arrowAni = newTranslateYAnimation(w.QWidget, 2)
	w.installEvents()
	return w
}

func (w *ComboBox) installEvents() {
	w.OnMousePressEvent(func(super func(event *qt.QMouseEvent), event *qt.QMouseEvent) {
		w.isPressed = true
		super(event)
		w.arrowAni.press()
	})
	w.OnMouseReleaseEvent(func(super func(event *qt.QMouseEvent), event *qt.QMouseEvent) {
		w.isPressed = false
		super(event)
		w.arrowAni.release()
		w.toggleComboMenu()
	})
	w.OnEnterEvent(func(super func(event *qt.QEvent), event *qt.QEvent) {
		super(event)
		w.isHover = true
		w.Update()
	})
	w.OnLeaveEvent(func(super func(event *qt.QEvent), event *qt.QEvent) {
		super(event)
		w.isHover = false
		w.Update()
	})
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
		super(event)
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing)
		if w.isHover {
			painter.SetOpacity(0.8)
		} else if w.isPressed {
			painter.SetOpacity(0.7)
		}

		rect := qt.NewQRectF4(float64(w.Width()-22), float64(w.Height())/2-5+w.arrowAni.y, 10, 10)
		defer rect.Delete()
		if common.IsDarkTheme() {
			renderFluentIcon(common.ChevronDown, painter, rect, common.ThemeAuto)
		} else {
			renderFluentIconWithFill(common.ChevronDown, painter, rect, "#646464")
		}
		painter.End()
	})
}

// SetText sets the button text and adjusts the size.
func (w *ComboBox) SetText(text string) {
	w.QPushButton.SetText(text)
	w.AdjustSize()
}

// AddItem appends an item.
func (w *ComboBox) AddItem(text string, icon interface{}, userData interface{}) {
	w.items = append(w.items, NewComboItem(text, icon, userData))
	if len(w.items) == 1 {
		w.SetCurrentIndex(0)
	}
}

// AddItems appends multiple text items.
func (w *ComboBox) AddItems(texts []string) {
	for _, t := range texts {
		w.AddItem(t, nil, nil)
	}
}

// RemoveItem removes the item at index and updates the current index.
func (w *ComboBox) RemoveItem(index int) {
	if index < 0 || index >= len(w.items) {
		return
	}
	w.items = append(w.items[:index], w.items[index+1:]...)

	if index < w.currentIndex {
		w.SetCurrentIndex(w.currentIndex - 1)
	} else if index == w.currentIndex {
		if index > 0 {
			w.SetCurrentIndex(w.currentIndex - 1)
		} else {
			w.SetText(w.ItemText(0))
			if w.OnCurrentTextChanged != nil {
				w.OnCurrentTextChanged(w.CurrentText())
			}
			if w.OnCurrentIndexChanged != nil {
				w.OnCurrentIndexChanged(0)
			}
		}
	}
	if w.Count() == 0 {
		w.Clear()
	}
}

// CurrentIndex returns the current index.
func (w *ComboBox) CurrentIndex() int { return w.currentIndex }

// SetCurrentIndex sets the current index.
func (w *ComboBox) SetCurrentIndex(index int) {
	if index < 0 {
		w.currentIndex = -1
		w.SetPlaceholderText(w.placeholderText)
		return
	}
	if index >= len(w.items) || index == w.currentIndex {
		return
	}
	w.updateTextState(false)
	w.setCurrentIndexBase(index)
}

func (w *ComboBox) setCurrentIndexBase(index int) {
	oldText := w.CurrentText()
	w.currentIndex = index
	w.SetText(w.items[index].Text)
	if oldText != w.CurrentText() {
		if w.OnCurrentTextChanged != nil {
			w.OnCurrentTextChanged(w.CurrentText())
		}
	}
	if w.OnCurrentIndexChanged != nil {
		w.OnCurrentIndexChanged(index)
	}
}

// CurrentText returns the current text.
func (w *ComboBox) CurrentText() string {
	if w.currentIndex < 0 || w.currentIndex >= len(w.items) {
		return ""
	}
	return w.items[w.currentIndex].Text
}

// CurrentData returns the user data of the current item.
func (w *ComboBox) CurrentData() interface{} {
	if w.currentIndex < 0 || w.currentIndex >= len(w.items) {
		return nil
	}
	return w.items[w.currentIndex].UserData
}

// SetCurrentText selects the item with the given text.
func (w *ComboBox) SetCurrentText(text string) {
	if text == w.CurrentText() {
		return
	}
	index := w.FindText(text)
	if index >= 0 {
		w.SetCurrentIndex(index)
	}
}

// SetItemText sets the text of an item.
func (w *ComboBox) SetItemText(index int, text string) {
	if index < 0 || index >= len(w.items) {
		return
	}
	w.items[index].Text = text
	if w.currentIndex == index {
		w.SetText(text)
	}
}

// ItemData returns the user data of an item.
func (w *ComboBox) ItemData(index int) interface{} {
	if index < 0 || index >= len(w.items) {
		return nil
	}
	return w.items[index].UserData
}

// ItemText returns the text of an item.
func (w *ComboBox) ItemText(index int) string {
	if index < 0 || index >= len(w.items) {
		return ""
	}
	return w.items[index].Text
}

// ItemIcon returns the icon of an item (an icon source, may be nil).
func (w *ComboBox) ItemIcon(index int) interface{} {
	if index < 0 || index >= len(w.items) {
		return nil
	}
	return w.items[index].Icon
}

// SetItemData sets the user data of an item.
func (w *ComboBox) SetItemData(index int, value interface{}) {
	if index < 0 || index >= len(w.items) {
		return
	}
	w.items[index].UserData = value
}

// SetItemIcon sets the icon of an item.
func (w *ComboBox) SetItemIcon(index int, icon interface{}) {
	if index < 0 || index >= len(w.items) {
		return
	}
	w.items[index].Icon = icon
}

// SetItemEnabled sets the enabled status of an item.
func (w *ComboBox) SetItemEnabled(index int, isEnabled bool) {
	if index < 0 || index >= len(w.items) {
		return
	}
	w.items[index].IsEnabled = isEnabled
}

// FindData returns the index of the item with the given user data.
func (w *ComboBox) FindData(data interface{}) int {
	for i, item := range w.items {
		if item.UserData == data {
			return i
		}
	}
	return -1
}

// FindText returns the index of the item with the given text.
func (w *ComboBox) FindText(text string) int {
	for i, item := range w.items {
		if item.Text == text {
			return i
		}
	}
	return -1
}

// Clear removes all items.
func (w *ComboBox) Clear() {
	if w.currentIndex >= 0 {
		w.SetText("")
	}
	w.items = nil
	w.currentIndex = -1
}

// Count returns the number of items.
func (w *ComboBox) Count() int { return len(w.items) }

// InsertItem inserts an item at the given index.
func (w *ComboBox) InsertItem(index int, text string, icon interface{}, userData interface{}) {
	w.items = insertComboItem(w.items, index, NewComboItem(text, icon, userData))
	if index <= w.currentIndex {
		w.SetCurrentIndex(w.currentIndex + 1)
	}
}

// InsertItems inserts multiple text items starting at index.
func (w *ComboBox) InsertItems(index int, texts []string) {
	pos := index
	for _, t := range texts {
		w.items = insertComboItem(w.items, pos, NewComboItem(t, nil, nil))
		pos++
	}
	if index <= w.currentIndex {
		w.SetCurrentIndex(w.currentIndex + pos - index)
	}
}

// SetMaxVisibleItems sets the maximum visible item count of the drop menu.
func (w *ComboBox) SetMaxVisibleItems(num int) { w.maxVisibleItems = num }

// MaxVisibleItems returns the maximum visible item count.
func (w *ComboBox) MaxVisibleItems() int { return w.maxVisibleItems }

// SetPlaceholderText sets the placeholder text shown when nothing is selected.
func (w *ComboBox) SetPlaceholderText(text string) {
	w.placeholderText = text
	if w.currentIndex < 0 {
		w.updateTextState(true)
		w.QPushButton.SetText(text)
	}
}

func (w *ComboBox) updateTextState(isPlaceholder bool) {
	if w.isPlaceholder == isPlaceholder {
		return
	}
	w.isPlaceholder = isPlaceholder
	w.SetProperty("isPlaceholderText", qt.NewQVariant11(isPlaceholder))
	w.SetStyle(qt.QApplication_Style())
}

func (w *ComboBox) toggleComboMenu() {
	if w.dropMenu != nil {
		w.dropMenu.Hide()
		w.dropMenu = nil
	} else {
		w.showComboMenu()
	}
}

func (w *ComboBox) showComboMenu() {
	if len(w.items) == 0 {
		return
	}

	menu := NewComboBoxMenu(w.QWidget)
	for i, item := range w.items {
		idx := i
		action := menu.AddActionIcon(item.Icon, item.Text)
		action.SetEnabled(item.IsEnabled)
		action.OnTriggered(func() { w.onItemClicked(idx) })
	}

	menu.SetMaxVisibleItems(w.maxVisibleItems)

	if w.currentIndex >= 0 && w.currentIndex < len(menu.MenuActions()) {
		menu.SetDefaultAction(menu.MenuActions()[w.currentIndex])
	}

	menu.OnClosed(func() { w.dropMenu = nil })
	w.dropMenu = menu

	// Center the menu on the combo box and choose DROP_DOWN vs PULL_UP from the
	// available vertical space, mirroring Python ComboBoxBase._showComboMenu (the
	// previous top-left positioning left the menu misaligned with the button).
	menu.ExecAtCentered(w.QWidget)
}

func (w *ComboBox) onItemClicked(index int) {
	if index != w.currentIndex {
		w.SetCurrentIndex(index)
	}
	if w.OnActivated != nil {
		w.OnActivated(index)
	}
	if w.OnTextActivated != nil {
		w.OnTextActivated(w.CurrentText())
	}
}

// EditableComboBox is a fluent editable combo box backed by a LineEdit with a
// drop-down button.
type EditableComboBox struct {
	*LineEdit
	items           []ComboItem
	currentIndex    int
	maxVisibleItems int
	dropMenu        *ComboBoxMenu
	placeholderText string
	dropButton      *LineEditButton

	OnCurrentIndexChanged func(int)
	OnCurrentTextChanged  func(string)
	OnActivated           func(int)
	OnTextActivated       func(string)
}

// NewEditableComboBox builds an editable combo box.
func NewEditableComboBox(parent *qt.QWidget) *EditableComboBox {
	w := &EditableComboBox{
		LineEdit:        NewLineEdit(parent),
		currentIndex:    -1,
		maxVisibleItems: -1,
	}
	w.dropButton = NewLineEditButton(common.ChevronDown, w.QWidget)
	w.SetTextMargins(0, 0, 29, 0)
	w.dropButton.SetFixedSize2(30, 25)
	w.hBoxLayout.AddWidget3(w.dropButton.QWidget, 0, qt.AlignRight)

	w.dropButton.OnClicked(w.toggleComboMenu)
	w.onTextChangedExtra = w.onComboTextChanged
	w.OnReturnPressed(w.onReturnPressed)
	w.onClearButtonClicked = func() { w.currentIndex = -1 }
	return w
}

// AddItem appends an item.
func (w *EditableComboBox) AddItem(text string, icon interface{}, userData interface{}) {
	w.items = append(w.items, NewComboItem(text, icon, userData))
	if len(w.items) == 1 {
		w.SetCurrentIndex(0)
	}
}

// AddItems appends multiple text items.
func (w *EditableComboBox) AddItems(texts []string) {
	for _, t := range texts {
		w.AddItem(t, nil, nil)
	}
}

// RemoveItem removes the item at index and updates the current index.
func (w *EditableComboBox) RemoveItem(index int) {
	if index < 0 || index >= len(w.items) {
		return
	}
	w.items = append(w.items[:index], w.items[index+1:]...)

	if index < w.currentIndex {
		w.SetCurrentIndex(w.currentIndex - 1)
	} else if index == w.currentIndex {
		if index > 0 {
			w.SetCurrentIndex(w.currentIndex - 1)
		} else {
			w.SetText(w.ItemText(0))
			if w.OnCurrentTextChanged != nil {
				w.OnCurrentTextChanged(w.CurrentText())
			}
			if w.OnCurrentIndexChanged != nil {
				w.OnCurrentIndexChanged(0)
			}
		}
	}
	if w.Count() == 0 {
		w.Clear()
	}
}

// CurrentIndex returns the current index.
func (w *EditableComboBox) CurrentIndex() int { return w.currentIndex }

// SetCurrentIndex sets the current index.
func (w *EditableComboBox) SetCurrentIndex(index int) {
	if index >= w.Count() || index == w.currentIndex {
		return
	}
	if index < 0 {
		w.currentIndex = -1
		w.SetText("")
		w.SetPlaceholderText(w.placeholderText)
	} else {
		w.currentIndex = index
		w.SetText(w.items[index].Text)
	}
}

// CurrentText returns the current text.
func (w *EditableComboBox) CurrentText() string { return w.Text() }

// CurrentData returns the user data of the current item.
func (w *EditableComboBox) CurrentData() interface{} { return w.ItemData(w.currentIndex) }

// SetCurrentText selects the item with the given text.
func (w *EditableComboBox) SetCurrentText(text string) {
	if text == w.CurrentText() {
		return
	}
	index := w.FindText(text)
	if index >= 0 {
		w.SetCurrentIndex(index)
	}
}

// SetItemText sets the text of an item.
func (w *EditableComboBox) SetItemText(index int, text string) {
	if index < 0 || index >= len(w.items) {
		return
	}
	w.items[index].Text = text
	if w.currentIndex == index {
		w.SetText(text)
	}
}

// ItemData returns the user data of an item.
func (w *EditableComboBox) ItemData(index int) interface{} {
	if index < 0 || index >= len(w.items) {
		return nil
	}
	return w.items[index].UserData
}

// ItemText returns the text of an item.
func (w *EditableComboBox) ItemText(index int) string {
	if index < 0 || index >= len(w.items) {
		return ""
	}
	return w.items[index].Text
}

// ItemIcon returns the icon of an item (an icon source, may be nil).
func (w *EditableComboBox) ItemIcon(index int) interface{} {
	if index < 0 || index >= len(w.items) {
		return nil
	}
	return w.items[index].Icon
}

// SetItemData sets the user data of an item.
func (w *EditableComboBox) SetItemData(index int, value interface{}) {
	if index < 0 || index >= len(w.items) {
		return
	}
	w.items[index].UserData = value
}

// SetItemIcon sets the icon of an item.
func (w *EditableComboBox) SetItemIcon(index int, icon interface{}) {
	if index < 0 || index >= len(w.items) {
		return
	}
	w.items[index].Icon = icon
}

// SetItemEnabled sets the enabled status of an item.
func (w *EditableComboBox) SetItemEnabled(index int, isEnabled bool) {
	if index < 0 || index >= len(w.items) {
		return
	}
	w.items[index].IsEnabled = isEnabled
}

// FindData returns the index of the item with the given user data.
func (w *EditableComboBox) FindData(data interface{}) int {
	for i, item := range w.items {
		if item.UserData == data {
			return i
		}
	}
	return -1
}

// FindText returns the index of the item with the given text.
func (w *EditableComboBox) FindText(text string) int {
	for i, item := range w.items {
		if item.Text == text {
			return i
		}
	}
	return -1
}

// Clear removes all items.
func (w *EditableComboBox) Clear() {
	if w.currentIndex >= 0 {
		w.SetText("")
	}
	w.items = nil
	w.currentIndex = -1
}

// Count returns the number of items.
func (w *EditableComboBox) Count() int { return len(w.items) }

// InsertItem inserts an item at the given index.
func (w *EditableComboBox) InsertItem(index int, text string, icon interface{}, userData interface{}) {
	w.items = insertComboItem(w.items, index, NewComboItem(text, icon, userData))
	if index <= w.currentIndex {
		w.SetCurrentIndex(w.currentIndex + 1)
	}
}

// InsertItems inserts multiple text items starting at index.
func (w *EditableComboBox) InsertItems(index int, texts []string) {
	pos := index
	for _, t := range texts {
		w.items = insertComboItem(w.items, pos, NewComboItem(t, nil, nil))
		pos++
	}
	if index <= w.currentIndex {
		w.SetCurrentIndex(w.currentIndex + pos - index)
	}
}

// SetMaxVisibleItems sets the maximum visible item count of the drop menu.
func (w *EditableComboBox) SetMaxVisibleItems(num int) { w.maxVisibleItems = num }

// MaxVisibleItems returns the maximum visible item count.
func (w *EditableComboBox) MaxVisibleItems() int { return w.maxVisibleItems }

// SetPlaceholderText sets the placeholder text.
func (w *EditableComboBox) SetPlaceholderText(text string) {
	w.placeholderText = text
	w.QLineEdit.SetPlaceholderText(text)
}

func (w *EditableComboBox) onReturnPressed() {
	if w.Text() == "" {
		return
	}
	index := w.FindText(w.Text())
	if index >= 0 && index != w.currentIndex {
		w.currentIndex = index
		if w.OnCurrentIndexChanged != nil {
			w.OnCurrentIndexChanged(index)
		}
	} else if index == -1 {
		w.AddItem(w.Text(), nil, nil)
		w.SetCurrentIndex(w.Count() - 1)
	}
}

func (w *EditableComboBox) onComboTextChanged(text string) {
	w.currentIndex = -1
	if w.OnCurrentTextChanged != nil {
		w.OnCurrentTextChanged(text)
	}
	for i, item := range w.items {
		if item.Text == text {
			w.currentIndex = i
			if w.OnCurrentIndexChanged != nil {
				w.OnCurrentIndexChanged(i)
			}
			return
		}
	}
}

func (w *EditableComboBox) toggleComboMenu() {
	if w.dropMenu != nil {
		w.dropMenu.Hide()
		w.dropMenu = nil
	} else {
		w.showComboMenu()
	}
}

func (w *EditableComboBox) showComboMenu() {
	if len(w.items) == 0 {
		return
	}

	menu := NewComboBoxMenu(w.QWidget)
	for i, item := range w.items {
		idx := i
		action := menu.AddActionIcon(item.Icon, item.Text)
		action.SetEnabled(item.IsEnabled)
		action.OnTriggered(func() { w.onItemClicked(idx) })
	}

	menu.SetMaxVisibleItems(w.maxVisibleItems)

	if w.currentIndex >= 0 && w.currentIndex < len(menu.MenuActions()) {
		menu.SetDefaultAction(menu.MenuActions()[w.currentIndex])
	}

	menu.OnClosed(func() { w.dropMenu = nil })
	w.dropMenu = menu

	// Center the menu on the combo box and choose DROP_DOWN vs PULL_UP from the
	// available vertical space, mirroring Python ComboBoxBase._showComboMenu.
	menu.ExecAtCentered(w.QWidget)
}

func (w *EditableComboBox) onItemClicked(index int) {
	if index != w.currentIndex {
		w.SetCurrentIndex(index)
	}
	if w.OnActivated != nil {
		w.OnActivated(index)
	}
	if w.OnTextActivated != nil {
		w.OnTextActivated(w.CurrentText())
	}
}
