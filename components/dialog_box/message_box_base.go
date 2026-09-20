package dialog_box

import (
	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	qt "github.com/mappu/miqt/qt"
)

// MessageBoxBase is a mask dialog with a title/OK/Cancel button bar and a
// subclass-supplied view layout.
//
// With a parent it masks that parent window (the fluent mask dialog). Without a
// parent there is no window to mask, so it becomes a standalone draggable window
// exactly like dialog_box.NewDialog (see initStandaloneWindow).
type MessageBoxBase struct {
	*MaskDialogBase
	buttonGroup      *qt.QFrame
	YesButton        *widgets.PrimaryPushButton
	CancelButton     *qt.QPushButton
	vBoxLayout       *qt.QVBoxLayout
	viewLayout       *qt.QVBoxLayout
	ButtonLayout     *qt.QHBoxLayout
	windowTitleLabel *qt.QLabel

	// ValidateFunc is called before accepting; the default returns true.
	ValidateFunc func() bool
}

// NewMessageBoxBase builds a message box base.
func NewMessageBoxBase(parent *qt.QWidget) *MessageBoxBase {
	b := &MessageBoxBase{MaskDialogBase: NewMaskDialogBase(parent)}
	b.buttonGroup = qt.NewQFrame(b.widget.QWidget)
	b.YesButton = widgets.NewPrimaryPushButtonText("OK", b.buttonGroup.QWidget)
	b.CancelButton = qt.NewQPushButton5("Cancel", b.buttonGroup.QWidget)
	b.vBoxLayout = qt.NewQVBoxLayout(b.widget.QWidget)
	b.viewLayout = qt.NewQVBoxLayout2()
	b.ButtonLayout = qt.NewQHBoxLayout(b.buttonGroup.QWidget)
	b.initWidget()
	if parent == nil {
		b.initStandaloneWindow()
	}
	return b
}

// initStandaloneWindow turns a parent-less message box into a draggable frameless
// window, mirroring dialog_box.NewDialog.
//
// Without a parent the mask has no window to dim: it would only paint a dark
// sheet over the desktop around the card, the window would have no native shadow
// or rounded corners and nothing could drag it. So the mask is made transparent,
// a title strip is added at the top of the card to drag the window through the
// native hit test, and the window gets the same native shadow/rounded corners as
// Dialog.
func (b *MessageBoxBase) initStandaloneWindow() {
	// No parent window to dim.
	b.SetMaskColor(qt.NewQColor11(0, 0, 0, 0))

	// The mask dialog floats its card over the mask and gives it a
	// graphics-effect drop shadow. A standalone window is shadowed natively
	// (installFramelessShadow below) — like NewDialog, which carries no effect —
	// and keeping the effect makes the card's paint region larger than the layered
	// window: closing it then fails Qt's layered-window update with
	// "UpdateLayeredWindowIndirect failed ... dirty=(...) (参数错误)".
	b.widget.SetGraphicsEffect(nil)

	// The drag strip must not contain interactive children, otherwise their
	// clicks would start a window move instead of reaching them. Its text is the
	// window title (empty unless the caller sets one).
	b.windowTitleLabel = qt.NewQLabel5(b.WindowTitle(), b.widget.QWidget)
	b.windowTitleLabel.SetObjectName("windowTitleLabel")
	b.vBoxLayout.InsertWidget3(0, b.windowTitleLabel.QWidget, 0, qt.AlignTop)

	// Size the window around the card. Subclasses add their content to ViewLayout
	// after this constructor returns, and the layout constraint keeps the window
	// in sync with it.
	//
	// The outer margins are zeroed so the window hugs the card exactly: the native
	// frame/shadow installed below wraps the *whole window*, so Qt's default
	// margins would leave the (invisible) mask area outside the card and draw the
	// window border in mid-air around it. Dialog zeroes the same margins for the
	// same reason.
	b.hBoxLayout.SetContentsMargins(0, 0, 0, 0)
	b.hBoxLayout.SetSizeConstraint(qt.QLayout__SetFixedSize)

	// The strip is added after setQss applied the stylesheet, so the
	// QLabel#windowTitleLabel rule has to be re-applied.
	common.FluentStyleSheet(common.FluentDialog).Apply(b.QWidget, common.ThemeAuto)

	installFramelessDrag(b.QDialog, b.windowTitleLabel.QWidget)
	installFramelessShadow(b.QDialog)
}

// SetTitle sets the standalone window's title: the window title and the text of
// the draggable title strip. It is a no-op for a mask dialog (which has no title
// strip), but the window title is still set.
func (b *MessageBoxBase) SetTitle(title string) {
	b.SetWindowTitle(title)
	if b.windowTitleLabel != nil {
		b.windowTitleLabel.SetText(title)
	}
}

// TitleBarVisible reports whether the standalone window shows its title strip.
func (b *MessageBoxBase) TitleBarVisible() bool {
	return b.windowTitleLabel != nil && b.windowTitleLabel.IsVisible()
}

// SetTitleBarVisible toggles the standalone window's title strip (a no-op for a
// mask dialog, which has no title strip).
func (b *MessageBoxBase) SetTitleBarVisible(isVisible bool) {
	if b.windowTitleLabel == nil {
		return
	}
	b.windowTitleLabel.SetVisible(isVisible)
}

func (b *MessageBoxBase) initWidget() {
	b.setQss()
	b.initLayout()

	b.SetShadowEffect(60, 0, 10, qt.NewQColor11(0, 0, 0, 50))
	b.SetMaskColor(qt.NewQColor11(0, 0, 0, 76))

	b.YesButton.SetAttribute(qt.WA_LayoutUsesWidgetRect)
	b.CancelButton.SetAttribute(qt.WA_LayoutUsesWidgetRect)
	b.YesButton.SetAttribute2(qt.WA_MacShowFocusRect, false)

	b.YesButton.SetFocus()
	b.buttonGroup.SetFixedHeight(81)
	// Round the bottom corners of the button bar directly (dialog.qss uses the
	// `MessageBoxBase #buttonGroup` class selector, which cannot match the plain
	// QFrame in the Go port).
	b.buttonGroup.SetStyleSheet("#buttonGroup { border-bottom-left-radius: 8px; border-bottom-right-radius: 8px; }")

	b.YesButton.OnClicked(b.onYesButtonClicked)
	b.CancelButton.OnClicked(b.onCancelButtonClicked)
}

func (b *MessageBoxBase) initLayout() {
	b.hBoxLayout.RemoveWidget(b.widget.QWidget)
	b.hBoxLayout.AddWidget3(b.widget.QWidget, 1, qt.AlignCenter)

	b.vBoxLayout.SetSpacing(0)
	b.vBoxLayout.SetContentsMargins(0, 0, 0, 0)
	b.vBoxLayout.AddLayout2(b.viewLayout.QLayout, 1)
	b.vBoxLayout.AddWidget3(b.buttonGroup.QWidget, 0, qt.AlignBottom)

	b.viewLayout.SetSpacing(12)
	b.viewLayout.SetContentsMargins(24, 24, 24, 24)

	b.ButtonLayout.SetSpacing(12)
	b.ButtonLayout.SetContentsMargins(24, 24, 24, 24)
	b.ButtonLayout.AddWidget3(b.YesButton.QWidget, 1, qt.AlignVCenter)
	b.ButtonLayout.AddWidget3(b.CancelButton.QWidget, 1, qt.AlignVCenter)
}

// Validate reports whether the form data is legal before closing.
func (b *MessageBoxBase) Validate() bool {
	if b.ValidateFunc != nil {
		return b.ValidateFunc()
	}
	return true
}

func (b *MessageBoxBase) onCancelButtonClicked() {
	b.Reject()
}

func (b *MessageBoxBase) onYesButtonClicked() {
	if b.Validate() {
		b.Accept()
	}
}

func (b *MessageBoxBase) setQss() {
	b.buttonGroup.SetObjectName("buttonGroup")
	b.CancelButton.SetObjectName("cancelButton")
	common.FluentStyleSheet(common.FluentDialog).Apply(b.QWidget, common.ThemeAuto)
}

// HideYesButton hides the yes button.
func (b *MessageBoxBase) HideYesButton() {
	b.YesButton.Hide()
	b.ButtonLayout.InsertStretch2(0, 1)
}

// HideCancelButton hides the cancel button.
func (b *MessageBoxBase) HideCancelButton() {
	b.CancelButton.Hide()
	b.ButtonLayout.InsertStretch2(0, 1)
}

// ViewLayout exposes the content layout subclasses fill.
func (b *MessageBoxBase) ViewLayout() *qt.QVBoxLayout { return b.viewLayout }
