package widgets

import (
	"strings"

	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// Line edit action positions (mirrors QLineEdit.ActionPosition).
const (
	LineEditLeadingPosition = iota
	LineEditTrailingPosition
)

// LineEditButton is a tool button that renders a fluent icon centered inside
// a line edit. It keeps the source icon and paints it manually (theme aware).
type LineEditButton struct {
	*qt.QToolButton
	icon      interface{}
	action    *qt.QAction
	isPressed bool
	onPress   func()
	onRelease func()
}

// NewLineEditButton builds a line edit button from an icon source
// (*qt.QIcon, string or common.FluentIconBase).
func NewLineEditButton(icon interface{}, parent *qt.QWidget) *LineEditButton {
	w := &LineEditButton{QToolButton: qt.NewQToolButton(parent)}
	w.icon = icon
	w.SetFixedSize2(31, 23)
	size := qt.NewQSize2(10, 10)
	w.SetIconSize(size)
	size.Delete()
	w.SetCursor(qt.NewQCursor2(qt.PointingHandCursor))
	w.SetObjectName("lineEditButton")
	common.FluentStyleSheet(common.FluentLineEdit).Apply(w.QWidget, common.ThemeAuto)
	w.OnMousePressEvent(func(super func(event *qt.QMouseEvent), event *qt.QMouseEvent) {
		w.isPressed = true
		if w.onPress != nil {
			w.onPress()
		}
		super(event)
	})
	w.OnMouseReleaseEvent(func(super func(event *qt.QMouseEvent), event *qt.QMouseEvent) {
		w.isPressed = false
		if w.onRelease != nil {
			w.onRelease()
		}
		super(event)
	})
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
		super(event)
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing | qt.QPainter__SmoothPixmapTransform)

		isize := w.IconSize()
		iw, ih := isize.Width(), isize.Height()
		wd, ht := w.Width(), w.Height()
		rect := qt.NewQRectF4(float64(wd-iw)/2, float64(ht-ih)/2, float64(iw), float64(ih))
		defer rect.Delete()

		if w.isPressed {
			painter.SetOpacity(0.7)
		}

		renderFluentIcon(w.icon, painter, rect, common.ThemeAuto)
		painter.End()
	})
	return w
}

// SetIcon sets the icon source.
func (w *LineEditButton) SetIcon(icon interface{}) {
	w.icon = icon
	w.Update()
}

// SetAction binds a QAction to the button, mirroring its state.
func (w *LineEditButton) SetAction(action *qt.QAction) {
	w.action = action
	w.syncAction()
	// The action is usually created by the caller and outlives the button (the
	// whole widget tree can be rebuilt, e.g. on a language change), so its toggled
	// signal must not call into a destroyed button.
	alive := trackWidget(w.OnDestroyed)
	w.OnClicked(func() {
		if alive.ok() {
			action.Trigger()
		}
	})
	action.OnToggled(func(checked bool) {
		if alive.ok() {
			w.SetChecked(checked)
		}
	})
}

// Action returns the bound action.
func (w *LineEditButton) Action() *qt.QAction { return w.action }

func (w *LineEditButton) syncAction() {
	a := w.action
	w.SetIcon(a.Icon())
	w.SetToolTip(a.ToolTip())
	w.SetEnabled(a.IsEnabled())
	w.SetCheckable(a.IsCheckable())
	w.SetChecked(a.IsChecked())
}

// LineEdit is a fluent line edit with an internal clear button and a themed
// focus underline. The QCompleter interaction is simplified away: the
// completer field is kept for API compatibility but no completer menu is shown.
type LineEdit struct {
	*qt.QLineEdit
	clearButtonEnabled      bool
	completer               *qt.QCompleter
	isError                 bool
	lightFocusedBorderColor *qt.QColor
	darkFocusedBorderColor  *qt.QColor
	leftButtons             []*LineEditButton
	rightButtons            []*LineEditButton
	hBoxLayout              *qt.QHBoxLayout
	clearButton             *LineEditButton

	onClearButtonClicked func()
	onTextChangedExtra   func(string)
}

// NewLineEdit builds a line edit.
func NewLineEdit(parent *qt.QWidget) *LineEdit {
	w := &LineEdit{QLineEdit: qt.NewQLineEdit(parent)}
	w.lightFocusedBorderColor = qt.NewQColor()
	w.darkFocusedBorderColor = qt.NewQColor()
	w.SetProperty("transparent", qt.NewQVariant11(true))
	common.FluentStyleSheet(common.FluentLineEdit).Apply(w.QWidget, common.ThemeAuto)
	w.SetFixedHeight(33)
	w.SetAttribute2(qt.WA_MacShowFocusRect, false)
	common.SetFont(w.QWidget, 14, 400)

	w.hBoxLayout = qt.NewQHBoxLayout(w.QWidget)
	w.clearButton = NewLineEditButton(common.Cancel, w.QWidget)
	w.clearButton.SetFixedSize2(29, 25)
	w.clearButton.Hide()

	w.hBoxLayout.SetSpacing(3)
	w.hBoxLayout.SetContentsMargins(4, 4, 4, 4)
	// Port of line_edit.py's `hBoxLayout.setAlignment(Qt.AlignRight |
	// Qt.AlignVCenter)`: packs the right-side buttons (clear/search/view) flush
	// against the line edit's right edge instead of spreading them across the
	// width. The alignment lives on QLayoutItem, so it is reached through the
	// promoted embedded field (the same-name QLayout.SetAlignment overload takes
	// a child widget).
	w.hBoxLayout.QLayoutItem.SetAlignment(qt.AlignRight | qt.AlignVCenter)
	w.hBoxLayout.AddWidget3(w.clearButton.QWidget, 0, qt.AlignRight)

	w.clearButton.OnClicked(func() {
		w.Clear()
		if w.onClearButtonClicked != nil {
			w.onClearButtonClicked()
		}
	})
	w.OnTextChanged(func(text string) {
		if w.clearButtonEnabled {
			w.clearButton.SetVisible(text != "" && w.HasFocus())
		}
		if w.onTextChangedExtra != nil {
			w.onTextChangedExtra(text)
		}
	})

	w.OnFocusOutEvent(func(super func(event *qt.QFocusEvent), event *qt.QFocusEvent) {
		super(event)
		w.clearButton.Hide()
	})
	w.OnFocusInEvent(func(super func(event *qt.QFocusEvent), event *qt.QFocusEvent) {
		super(event)
		if w.clearButtonEnabled {
			w.clearButton.SetVisible(w.Text() != "")
		}
	})

	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
		super(event)
		if !w.HasFocus() {
			return
		}
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing)
		painter.SetPenWithStyle(qt.NoPen)

		m := w.ContentsMargins()

		wd, ht := w.Width(), w.Height()

		path := qt.NewQPainterPath()
		defer path.Delete()
		rect := qt.NewQRectF4(float64(m.Left()), float64(ht-10), float64(wd-m.Left()-m.Right()), 10)
		defer rect.Delete()
		path.AddRoundedRect(rect, 5, 5)

		cutPath := qt.NewQPainterPath()
		defer cutPath.Delete()
		cut := qt.NewQRectF4(float64(m.Left()), float64(ht-10), float64(wd-m.Left()-m.Right()), 8)
		defer cut.Delete()
		cutPath.AddRect(cut)

		sub := path.Subtracted(cutPath) // GoGC-armed — do NOT Delete

		color := w.focusedBorderColor()
		defer color.Delete()
		brush := qt.NewQBrush3(color)
		defer brush.Delete()
		painter.FillPath(sub, brush)
		painter.End()
	})
	return w
}

// IsError reports whether the line edit is in error status.
func (w *LineEdit) IsError() bool { return w.isError }

// SetError sets the error status.
func (w *LineEdit) SetError(isError bool) {
	if isError == w.isError {
		return
	}
	w.isError = isError
	w.Update()
}

// SetCustomFocusedBorderColor sets the focused border color for light/dark mode.
func (w *LineEdit) SetCustomFocusedBorderColor(light, dark *qt.QColor) {
	w.lightFocusedBorderColor = cloneColor(light)
	w.darkFocusedBorderColor = cloneColor(dark)
	w.Update()
}

func (w *LineEdit) focusedBorderColor() *qt.QColor {
	if w.isError {
		return common.CriticalForeground.Color(common.ThemeAuto)
	}
	var c *qt.QColor
	if common.IsDarkTheme() {
		c = w.darkFocusedBorderColor
	} else {
		c = w.lightFocusedBorderColor
	}
	if c != nil && c.IsValid() {
		return cloneColor(c)
	}
	return common.ThemeColorPrimary.Color()
}

// SetClearButtonEnabled toggles the clear button.
func (w *LineEdit) SetClearButtonEnabled(enable bool) {
	w.clearButtonEnabled = enable
	w.adjustTextMargins()
}

// IsClearButtonEnabled reports whether the clear button is enabled.
func (w *LineEdit) IsClearButtonEnabled() bool { return w.clearButtonEnabled }

// SetOnTextChangedExtra registers an additional text-changed callback invoked
// by the built-in OnTextChanged handler (after the clear-button visibility
// update). It lets callers reposition suffixes etc. without overriding the
// internal handler.
func (w *LineEdit) SetOnTextChangedExtra(f func(string)) { w.onTextChangedExtra = f }

// SetCompleter stores a completer (kept for API compatibility; no QCompleter
// interaction is performed in this port).
func (w *LineEdit) SetCompleter(completer *qt.QCompleter) { w.completer = completer }

// Completer returns the stored completer.
func (w *LineEdit) Completer() *qt.QCompleter { return w.completer }

// AddAction adds a QAction as an internal leading/trailing icon button.
func (w *LineEdit) AddAction(action *qt.QAction, position int) {
	w.QWidget.AddAction(action)

	button := NewLineEditButton(action.Icon(), w.QWidget)
	button.SetAction(action)
	button.SetFixedWidth(29)

	if position == LineEditLeadingPosition {
		w.hBoxLayout.InsertWidget3(len(w.leftButtons), button.QWidget, 0, qt.AlignLeading)
		if len(w.leftButtons) == 0 {
			// Port of line_edit.py's `insertStretch(1, 1)` after the first
			// leading action: the stretch sits between the leading and trailing
			// buttons so the clear button stays flush right.
			w.hBoxLayout.InsertStretch2(1, 1)
		}
		w.leftButtons = append(w.leftButtons, button)
	} else {
		w.hBoxLayout.AddWidget3(button.QWidget, 0, qt.AlignRight)
		w.rightButtons = append(w.rightButtons, button)
	}
	w.adjustTextMargins()
}

// AddActions adds multiple QActions at the given position.
func (w *LineEdit) AddActions(actions []*qt.QAction, position int) {
	for _, action := range actions {
		w.AddAction(action, position)
	}
}

func (w *LineEdit) adjustTextMargins() {
	left := len(w.leftButtons) * 30
	right := len(w.rightButtons) * 30
	if w.clearButtonEnabled {
		right += 28
	}
	m := w.TextMargins() // GoGC-armed — do NOT Delete
	w.SetTextMargins(left, m.Top(), right, m.Bottom())
}

// SearchLineEdit is a line edit with a search button.
type SearchLineEdit struct {
	*LineEdit
	searchButton *LineEditButton

	SearchSignal func(string)
	ClearSignal  func()
}

// NewSearchLineEdit builds a search line edit.
func NewSearchLineEdit(parent *qt.QWidget) *SearchLineEdit {
	w := &SearchLineEdit{LineEdit: NewLineEdit(parent)}
	w.searchButton = NewLineEditButton(common.Search, w.QWidget)
	w.hBoxLayout.AddWidget3(w.searchButton.QWidget, 0, qt.AlignRight)
	w.SetClearButtonEnabled(true)
	w.SetTextMargins(0, 0, 59, 0)

	w.searchButton.OnClicked(w.search)
	w.onClearButtonClicked = func() {
		if w.ClearSignal != nil {
			w.ClearSignal()
		}
	}
	return w
}

// SetClearButtonEnabled overrides the base to account for the search button.
func (w *SearchLineEdit) SetClearButtonEnabled(enable bool) {
	w.clearButtonEnabled = enable
	right := 30
	if enable {
		right += 28
	}
	w.SetTextMargins(0, 0, right, 0)
}

func (w *SearchLineEdit) search() {
	text := strings.TrimSpace(w.Text())
	if text != "" {
		if w.SearchSignal != nil {
			w.SearchSignal(text)
		}
	} else if w.ClearSignal != nil {
		w.ClearSignal()
	}
}

// PasswordLineEdit is a password line edit with a reveal (view) button.
type PasswordLineEdit struct {
	*LineEdit
	viewButton *LineEditButton
}

// NewPasswordLineEdit builds a password line edit.
func NewPasswordLineEdit(parent *qt.QWidget) *PasswordLineEdit {
	w := &PasswordLineEdit{LineEdit: NewLineEdit(parent)}
	w.viewButton = NewLineEditButton(common.View, w.QWidget)

	w.SetEchoMode(qt.QLineEdit__Password)
	w.SetContextMenuPolicy(qt.NoContextMenu)
	w.hBoxLayout.AddWidget3(w.viewButton.QWidget, 0, qt.AlignRight)
	w.SetClearButtonEnabled(false)

	size := qt.NewQSize2(13, 13)
	w.viewButton.SetIconSize(size)
	size.Delete()
	w.viewButton.SetFixedSize2(29, 25)

	w.viewButton.onPress = func() { w.SetPasswordVisible(true) }
	w.viewButton.onRelease = func() { w.SetPasswordVisible(false) }
	return w
}

// SetPasswordVisible sets the visibility of the password.
func (w *PasswordLineEdit) SetPasswordVisible(isVisible bool) {
	if isVisible {
		w.SetEchoMode(qt.QLineEdit__Normal)
	} else {
		w.SetEchoMode(qt.QLineEdit__Password)
	}
}

// IsPasswordVisible reports whether the password is visible.
func (w *PasswordLineEdit) IsPasswordVisible() bool {
	return w.EchoMode() == qt.QLineEdit__Normal
}

// SetClearButtonEnabled overrides the base to account for the view button width.
func (w *PasswordLineEdit) SetClearButtonEnabled(enable bool) {
	w.clearButtonEnabled = enable
	right := 0
	if enable {
		right = 28
	}
	if !w.viewButton.IsHidden() {
		right += 30
	}
	w.SetTextMargins(0, 0, right, 0)
}

// SetViewPasswordButtonVisible sets the visibility of the view password button.
func (w *PasswordLineEdit) SetViewPasswordButtonVisible(isVisible bool) {
	w.viewButton.SetVisible(isVisible)
}

// TextEdit is a fluent multi-line text edit.
type TextEdit struct {
	*qt.QTextEdit
}

// NewTextEdit builds a text edit.
func NewTextEdit(parent *qt.QWidget) *TextEdit {
	w := &TextEdit{QTextEdit: qt.NewQTextEdit(parent)}
	common.FluentStyleSheet(common.FluentLineEdit).Apply(w.QWidget, common.ThemeAuto)
	common.SetFont(w.QWidget, 14, 400)
	return w
}

// PlainTextEdit is a fluent plain text edit.
type PlainTextEdit struct {
	*qt.QPlainTextEdit
}

// NewPlainTextEdit builds a plain text edit.
func NewPlainTextEdit(parent *qt.QWidget) *PlainTextEdit {
	w := &PlainTextEdit{QPlainTextEdit: qt.NewQPlainTextEdit(parent)}
	common.FluentStyleSheet(common.FluentLineEdit).Apply(w.QWidget, common.ThemeAuto)
	common.SetFont(w.QWidget, 14, 400)
	return w
}

// TextBrowser is a fluent read-only rich text browser.
type TextBrowser struct {
	*qt.QTextBrowser
}

// NewTextBrowser builds a text browser.
func NewTextBrowser(parent *qt.QWidget) *TextBrowser {
	w := &TextBrowser{QTextBrowser: qt.NewQTextBrowser(parent)}
	common.FluentStyleSheet(common.FluentLineEdit).Apply(w.QWidget, common.ThemeAuto)
	common.SetFont(w.QWidget, 14, 400)
	return w
}
