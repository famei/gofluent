package dialog_box

import (
	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	qt "github.com/mappu/miqt/qt"
)

// messageBoxUi is the shared title/content/buttons UI used by both Dialog and
// MessageBox (the Go equivalent of the Python Ui_MessageBox mixin).
type messageBoxUi struct {
	content      string
	titleLabel   *qt.QLabel
	contentLabel *widgets.BodyLabel
	buttonGroup  *qt.QFrame
	yesButton    *widgets.PrimaryPushButton
	cancelButton *qt.QPushButton
	vBoxLayout   *qt.QVBoxLayout
	textLayout   *qt.QVBoxLayout
	buttonLayout *qt.QHBoxLayout
}

func (ui *messageBoxUi) setUpUi(host *qt.QWidget, parent *qt.QWidget, title, content string, onYes, onCancel func()) {
	ui.content = content
	ui.titleLabel = qt.NewQLabel5(title, parent)
	ui.contentLabel = widgets.NewBodyLabelText(content, parent)
	ui.buttonGroup = qt.NewQFrame(parent)
	ui.yesButton = widgets.NewPrimaryPushButtonText("OK", ui.buttonGroup.QWidget)
	ui.cancelButton = qt.NewQPushButton5("Cancel", ui.buttonGroup.QWidget)
	ui.vBoxLayout = qt.NewQVBoxLayout(parent)
	ui.textLayout = qt.NewQVBoxLayout2()
	ui.buttonLayout = qt.NewQHBoxLayout(ui.buttonGroup.QWidget)

	ui.initWidget(host, onYes, onCancel)
}

func (ui *messageBoxUi) initWidget(host *qt.QWidget, onYes, onCancel func()) {
	ui.setQss(host)
	ui.initLayout()

	ui.yesButton.SetAttribute(qt.WA_LayoutUsesWidgetRect)
	ui.cancelButton.SetAttribute(qt.WA_LayoutUsesWidgetRect)
	ui.yesButton.SetFocus()
	ui.buttonGroup.SetFixedHeight(81)
	ui.contentLabel.SetContextMenuPolicy(qt.CustomContextMenu)

	ui.yesButton.OnClicked(onYes)
	ui.cancelButton.OnClicked(onCancel)
}

// adjustText wraps the content text to the host window width.
func (ui *messageBoxUi) adjustText(host *qt.QWidget) {
	var chars int
	if host.IsWindow() {
		if host.ParentWidget() != nil {
			w := ui.titleLabel.Width()
			if host.ParentWidget().Width() > w {
				w = host.ParentWidget().Width()
			}
			chars = clampChars(w, 140)
		} else {
			chars = 100
		}
	} else {
		w := ui.titleLabel.Width()
		if host.Window().Width() > w {
			w = host.Window().Width()
		}
		chars = clampChars(w, 100)
	}

	wrapped, _ := common.Wrap(ui.content, chars, false)
	ui.contentLabel.SetText(wrapped)
}

func (ui *messageBoxUi) initLayout() {
	ui.vBoxLayout.SetSpacing(0)
	ui.vBoxLayout.SetContentsMargins(0, 0, 0, 0)
	ui.vBoxLayout.AddLayout2(ui.textLayout.QLayout, 1)
	ui.vBoxLayout.AddWidget3(ui.buttonGroup.QWidget, 0, qt.AlignBottom)
	ui.vBoxLayout.SetSizeConstraint(qt.QLayout__SetMinimumSize)

	ui.textLayout.SetSpacing(12)
	ui.textLayout.SetContentsMargins(24, 24, 24, 24)
	ui.textLayout.AddWidget3(ui.titleLabel.QWidget, 0, qt.AlignTop)
	ui.textLayout.AddWidget3(ui.contentLabel.QWidget, 0, qt.AlignTop)

	ui.buttonLayout.SetSpacing(12)
	ui.buttonLayout.SetContentsMargins(24, 24, 24, 24)
	ui.buttonLayout.AddWidget3(ui.yesButton.QWidget, 1, qt.AlignVCenter)
	ui.buttonLayout.AddWidget3(ui.cancelButton.QWidget, 1, qt.AlignVCenter)
}

func (ui *messageBoxUi) setQss(host *qt.QWidget) {
	ui.titleLabel.SetObjectName("titleLabel")
	ui.contentLabel.SetObjectName("contentLabel")
	ui.buttonGroup.SetObjectName("buttonGroup")
	ui.cancelButton.SetObjectName("cancelButton")

	common.FluentStyleSheet(common.FluentDialog).Apply(host, common.ThemeAuto)
	common.FluentStyleSheet(common.FluentDialog).Apply(ui.contentLabel.QWidget, common.ThemeAuto)

	ui.yesButton.AdjustSize()
	ui.cancelButton.AdjustSize()
}

// SetContentCopyable toggles whether the content can be selected by mouse.
func (ui *messageBoxUi) SetContentCopyable(isCopyable bool) {
	if isCopyable {
		ui.contentLabel.SetTextInteractionFlags(qt.TextSelectableByMouse)
	} else {
		ui.contentLabel.SetTextInteractionFlags(qt.NoTextInteraction)
	}
}

// HideYesButton hides the yes button and inserts a leading stretch.
func (ui *messageBoxUi) HideYesButton() {
	ui.yesButton.Hide()
	ui.buttonLayout.InsertStretch2(0, 1)
}

// HideCancelButton hides the cancel button and inserts a leading stretch.
func (ui *messageBoxUi) HideCancelButton() {
	ui.cancelButton.Hide()
	ui.buttonLayout.InsertStretch2(0, 1)
}

// Dialog is a frameless message dialog box.
type Dialog struct {
	*qt.QDialog
	messageBoxUi
	windowTitleLabel *qt.QLabel

	OnYes    func()
	OnCancel func()
}

// NewDialog builds a frameless dialog box.
func NewDialog(title, content string, parent *qt.QWidget) *Dialog {
	d := &Dialog{QDialog: qt.NewQDialog(parent)}
	// Keep the Qt::Dialog type alongside the frameless hint. Setting only
	// FramelessWindowHint replaces the whole flag word and clears the
	// Dialog/Window type bit, which makes QDialog::exec() show an embedded
	// child widget instead of a top-level modal window (the "dialog never
	// appears" bug).
	d.SetWindowFlags(qt.FramelessWindowHint | qt.Dialog)
	d.messageBoxUi.setUpUi(d.QWidget, d.QWidget, title, content, d.acceptYes, d.rejectCancel)
	d.windowTitleLabel = qt.NewQLabel5(title, d.QWidget)

	d.Resize(240, 192)
	// The external qframelesswindow title bar is not available in this port;
	// the frameless flag above replaces FramelessDialog.
	d.vBoxLayout.InsertWidget3(0, d.windowTitleLabel.QWidget, 0, qt.AlignTop)
	d.windowTitleLabel.SetObjectName("windowTitleLabel")
	common.FluentStyleSheet(common.FluentDialog).Apply(d.QWidget, common.ThemeAuto)
	d.SetFixedSize2(d.Width(), d.Height())
	d.messageBoxUi.adjustText(d.QWidget)
	installFramelessDrag(d.QDialog, d.windowTitleLabel.QWidget)
	installFramelessShadow(d.QDialog)
	return d
}

func (d *Dialog) acceptYes() {
	d.Accept()
	if d.OnYes != nil {
		d.OnYes()
	}
}

func (d *Dialog) rejectCancel() {
	d.Reject()
	if d.OnCancel != nil {
		d.OnCancel()
	}
}

// SetTitleBarVisible toggles the manual title label.
func (d *Dialog) SetTitleBarVisible(isVisible bool) {
	d.windowTitleLabel.SetVisible(isVisible)
}

// MessageBox is a mask-backed message box.
type MessageBox struct {
	*MaskDialogBase
	messageBoxUi

	OnYes    func()
	OnCancel func()
}

// NewMessageBox builds a mask message box.
func NewMessageBox(title, content string, parent *qt.QWidget) *MessageBox {
	m := &MessageBox{MaskDialogBase: NewMaskDialogBase(parent)}
	m.messageBoxUi.setUpUi(m.QWidget, m.widget.QWidget, title, content, m.acceptYes, m.rejectCancel)

	m.SetShadowEffect(60, 0, 10, qt.NewQColor11(0, 0, 0, 50))
	m.SetMaskColor(qt.NewQColor11(0, 0, 0, 76))
	m.hBoxLayout.RemoveWidget(m.widget.QWidget)
	m.hBoxLayout.AddWidget3(m.widget.QWidget, 1, qt.AlignCenter)

	m.buttonGroup.SetMinimumWidth(280)
	// dialog.qss rounds the button bar through the `MessageBox #buttonGroup`
	// class selector, which cannot match the plain QFrame in the Go port; round
	// the bottom corners directly so the card's lower edge is not square (bug
	// #9). The inline radius only conflicts with the inherited background rule
	// for these two properties, so the theme-aware buttonGroup background and
	// top border still come from dialog.qss.
	m.buttonGroup.SetStyleSheet("#buttonGroup { border-bottom-left-radius: 8px; border-bottom-right-radius: 8px; }")
	m.messageBoxUi.adjustText(m.widget.QWidget)

	// Compute the dialog size from the label size hints. The widget has not been
	// shown yet, so Width()/Height()/Y() would all read zero here — the previous
	// code therefore produced a degenerate 48x105 box (the "MessageBox layout
	// broken" bug). SizeHint() returns the natural label size right after the
	// wrapped text is set.
	tw := m.titleLabel.SizeHint()   // GoGC-armed — do NOT Delete
	cw := m.contentLabel.SizeHint() // GoGC-armed — do NOT Delete
	w := maxInt(cw.Width(), tw.Width()) + 48
	w = maxInt(w, 280)
	// Mirrors contentLabel.y() + contentLabel.height() + 105, where
	// contentLabel.y() == titleLabel.height() + 36 (24 top margin + 12 spacing).
	h := tw.Height() + cw.Height() + 141
	m.widget.SetFixedSize2(w, h)

	m.eventFilterHook = func(watched *qt.QObject, event *qt.QEvent) {
		if watched.UnsafePointer() == m.Window().QObject.UnsafePointer() && event.Type() == qt.QEvent__Resize {
			m.messageBoxUi.adjustText(m.widget.QWidget)
		}
	}
	return m
}

func (m *MessageBox) acceptYes() {
	m.Accept()
	if m.OnYes != nil {
		m.OnYes()
	}
}

func (m *MessageBox) rejectCancel() {
	m.Reject()
	if m.OnCancel != nil {
		m.OnCancel()
	}
}

func clampChars(w, maxVal int) int {
	f := float64(w) / 9
	if f > float64(maxVal) {
		f = float64(maxVal)
	}
	if f < 30 {
		f = 30
	}
	return int(f)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
