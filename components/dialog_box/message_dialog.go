package dialog_box

import (
	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// MessageDialog is a Win10-style message dialog box with a mask.
type MessageDialog struct {
	*MaskDialogBase
	content      string
	titleLabel   *qt.QLabel
	contentLabel *qt.QLabel
	yesButton    *qt.QPushButton
	cancelButton *qt.QPushButton

	OnYes    func()
	OnCancel func()
}

// NewMessageDialog builds a mask message dialog.
func NewMessageDialog(title, content string, parent *qt.QWidget) *MessageDialog {
	d := &MessageDialog{MaskDialogBase: NewMaskDialogBase(parent)}
	d.content = content
	d.titleLabel = qt.NewQLabel5(title, d.widget.QWidget)
	d.contentLabel = qt.NewQLabel5(content, d.widget.QWidget)
	d.yesButton = qt.NewQPushButton5("OK", d.widget.QWidget)
	d.cancelButton = qt.NewQPushButton5("Cancel", d.widget.QWidget)
	d.initWidget()
	return d
}

func (d *MessageDialog) initWidget() {
	d.windowMask.Resize(d.Width(), d.Height())
	d.widget.SetMaximumWidth(540)
	d.titleLabel.Move(24, 24)
	d.contentLabel.Move(24, 56)

	wrapped, _ := common.Wrap(d.content, 71, false)
	d.contentLabel.SetText(wrapped)

	d.setQss()
	d.initLayout()

	d.yesButton.OnClicked(d.onYesButtonClicked)
	d.cancelButton.OnClicked(d.onCancelButtonClicked)
}

func (d *MessageDialog) initLayout() {
	d.contentLabel.AdjustSize()
	w := 48 + d.contentLabel.Width()
	h := d.contentLabel.Y() + d.contentLabel.Height() + 92
	d.widget.SetFixedSize2(w, h)

	bw := (d.widget.Width() - 54) / 2
	d.yesButton.Resize(bw, 32)
	d.cancelButton.Resize(bw, 32)
	d.yesButton.Move(24, d.widget.Height()-56)
	d.cancelButton.Move(d.widget.Width()-24-d.cancelButton.Width(), d.widget.Height()-56)
}

func (d *MessageDialog) onCancelButtonClicked() {
	if d.OnCancel != nil {
		d.OnCancel()
	}
	d.SetResult(int(qt.QDialog__Rejected))
	d.Close()
}

func (d *MessageDialog) onYesButtonClicked() {
	d.SetEnabled(false)
	if d.OnYes != nil {
		d.OnYes()
	}
	d.SetResult(int(qt.QDialog__Accepted))
	d.Close()
}

func (d *MessageDialog) setQss() {
	d.windowMask.SetObjectName("windowMask")
	d.titleLabel.SetObjectName("titleLabel")
	d.contentLabel.SetObjectName("contentLabel")
	common.FluentStyleSheet(common.FluentMessageDialog).Apply(d.QWidget, common.ThemeAuto)
}
