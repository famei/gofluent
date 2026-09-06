package dialog_box

import (
	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	qt "github.com/mappu/miqt/qt"
)

// MessageBoxBase is a mask dialog with a title/OK/Cancel button bar and a
// subclass-supplied view layout.
type MessageBoxBase struct {
	*MaskDialogBase
	buttonGroup  *qt.QFrame
	yesButton    *widgets.PrimaryPushButton
	cancelButton *qt.QPushButton
	vBoxLayout   *qt.QVBoxLayout
	viewLayout   *qt.QVBoxLayout
	buttonLayout *qt.QHBoxLayout

	// ValidateFunc is called before accepting; the default returns true.
	ValidateFunc func() bool
}

// NewMessageBoxBase builds a message box base.
func NewMessageBoxBase(parent *qt.QWidget) *MessageBoxBase {
	b := &MessageBoxBase{MaskDialogBase: NewMaskDialogBase(parent)}
	b.buttonGroup = qt.NewQFrame(b.widget.QWidget)
	b.yesButton = widgets.NewPrimaryPushButtonText("OK", b.buttonGroup.QWidget)
	b.cancelButton = qt.NewQPushButton5("Cancel", b.buttonGroup.QWidget)
	b.vBoxLayout = qt.NewQVBoxLayout(b.widget.QWidget)
	b.viewLayout = qt.NewQVBoxLayout2()
	b.buttonLayout = qt.NewQHBoxLayout(b.buttonGroup.QWidget)
	b.initWidget()
	return b
}

func (b *MessageBoxBase) initWidget() {
	b.setQss()
	b.initLayout()

	b.SetShadowEffect(60, 0, 10, qt.NewQColor11(0, 0, 0, 50))
	b.SetMaskColor(qt.NewQColor11(0, 0, 0, 76))

	b.yesButton.SetAttribute(qt.WA_LayoutUsesWidgetRect)
	b.cancelButton.SetAttribute(qt.WA_LayoutUsesWidgetRect)
	b.yesButton.SetAttribute2(qt.WA_MacShowFocusRect, false)

	b.yesButton.SetFocus()
	b.buttonGroup.SetFixedHeight(81)
	// Round the bottom corners of the button bar directly (dialog.qss uses the
	// `MessageBoxBase #buttonGroup` class selector, which cannot match the plain
	// QFrame in the Go port).
	b.buttonGroup.SetStyleSheet("#buttonGroup { border-bottom-left-radius: 8px; border-bottom-right-radius: 8px; }")

	b.yesButton.OnClicked(b.onYesButtonClicked)
	b.cancelButton.OnClicked(b.onCancelButtonClicked)
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

	b.buttonLayout.SetSpacing(12)
	b.buttonLayout.SetContentsMargins(24, 24, 24, 24)
	b.buttonLayout.AddWidget3(b.yesButton.QWidget, 1, qt.AlignVCenter)
	b.buttonLayout.AddWidget3(b.cancelButton.QWidget, 1, qt.AlignVCenter)
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
	b.cancelButton.SetObjectName("cancelButton")
	common.FluentStyleSheet(common.FluentDialog).Apply(b.QWidget, common.ThemeAuto)
}

// HideYesButton hides the yes button.
func (b *MessageBoxBase) HideYesButton() {
	b.yesButton.Hide()
	b.buttonLayout.InsertStretch2(0, 1)
}

// HideCancelButton hides the cancel button.
func (b *MessageBoxBase) HideCancelButton() {
	b.cancelButton.Hide()
	b.buttonLayout.InsertStretch2(0, 1)
}

// ViewLayout exposes the content layout subclasses fill.
func (b *MessageBoxBase) ViewLayout() *qt.QVBoxLayout { return b.viewLayout }
