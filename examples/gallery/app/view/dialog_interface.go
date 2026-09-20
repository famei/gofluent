package view

import (
	"fmt"

	"github.com/famei/gofluent/components/dialog_box"
	"github.com/famei/gofluent/components/widgets"
	gallerycommon "github.com/famei/gofluent/examples/gallery/app/common"
	"github.com/famei/gofluent/examples/gallery/app/resource"
	qt "github.com/mappu/miqt/qt"
)

// DialogInterface is the "Dialogs & flyouts" gallery page (port of
// app/view/dialog_interface.py).
type DialogInterface struct {
	*GalleryInterface

	simpleFlyoutButton  *widgets.PushButton
	complexFlyoutButton *widgets.PushButton
	teachingButton      *widgets.PushButton
	teachingRightButton *widgets.PushButton
}

// NewDialogInterface builds the dialog interface.
func NewDialogInterface(parent *qt.QWidget) *DialogInterface {
	t := gallerycommon.NewTranslator()
	i := &DialogInterface{GalleryInterface: NewGalleryInterface(t.Dialogs, "github.com/famei/gofluent/components/dialog_box", parent)}
	i.SetObjectName("dialogInterface")

	tr := func(s string) string { return gallerycommon.Tr("DialogInterface", s) }

	const dialogSrc = "dialog_flyout/dialog/main.go"

	button := widgets.NewPushButtonText(tr("Show dialog"), nil)
	button.OnClicked(i.showDialog)
	i.AddExampleCard(tr("A frameless message box"), button.QWidget, "components/dialog_box/dialog.go", codeDialog, 0)

	button = widgets.NewPushButtonText(tr("Show dialog"), nil)
	button.OnClicked(i.showMessageDialog)
	i.AddExampleCard(tr("A message box with mask"), button.QWidget,
		"components/dialog_box/dialog.go", codeMessageBox, 0)

	button = widgets.NewPushButtonText(tr("Show dialog"), nil)
	button.OnClicked(i.showCustomDialog)
	i.AddExampleCard(tr("A custom message box"), button.QWidget,
		"components/dialog_box/message_box_base.go", codeCustomMessageBox, 0)

	button = widgets.NewPushButtonText(tr("Show dialog"), nil)
	button.OnClicked(i.showColorDialog)
	i.AddExampleCard(tr("A color dialog"), button.QWidget,
		"components/dialog_box/color_dialog.go", codeColorDialog, 0)

	i.simpleFlyoutButton = widgets.NewPushButtonText(tr("Show flyout"), nil)
	i.simpleFlyoutButton.OnClicked(i.showSimpleFlyout)
	i.AddExampleCard(tr("A simple flyout"), i.simpleFlyoutButton.QWidget,
		"components/widgets/flyout.go", codeSimpleFlyout, 0)

	i.complexFlyoutButton = widgets.NewPushButtonText(tr("Show flyout"), nil)
	i.complexFlyoutButton.OnClicked(i.showComplexFlyout)
	i.AddExampleCard(tr("A flyout with image and button"), i.complexFlyoutButton.QWidget,
		"components/widgets/flyout.go", codeComplexFlyout, 0)

	i.teachingButton = widgets.NewPushButtonText(tr("Show teaching tip"), nil)
	i.teachingButton.OnClicked(i.showBottomTeachingTip)
	i.AddExampleCard(tr("A teaching tip"), i.teachingButton.QWidget,
		"components/widgets/teaching_tip.go", codeTeachingTip, 0)

	i.teachingRightButton = widgets.NewPushButtonText(tr("Show teaching tip"), nil)
	i.teachingRightButton.OnClicked(i.showLeftBottomTeachingTip)
	i.AddExampleCard(tr("A teaching tip with image and button"), i.teachingRightButton.QWidget,
		"components/widgets/teaching_tip.go", codeTeachingTipImage, 0)
	return i
}

func (i *DialogInterface) showDialog() {
	title := gallerycommon.Tr("DialogInterface", "This is a frameless message dialog")
	content := gallerycommon.Tr("DialogInterface", "If the content of the message box is veeeeeeeeeeeeeeeeeeeeeeeeeery long, it will automatically wrap like this.")
	w := dialog_box.NewDialog(title, content, i.Window())
	w.SetContentCopyable(true)
	if w.Exec() != 0 {
		fmt.Println("Yes button is pressed")
	} else {
		fmt.Println("Cancel button is pressed")
	}
}

func (i *DialogInterface) showMessageDialog() {
	title := gallerycommon.Tr("DialogInterface", "This is a message dialog with mask")
	content := gallerycommon.Tr("DialogInterface", "If the content of the message box is veeeeeeeeeeeeeeeeeeeeeeeeeery long, it will automatically wrap like this.")
	w := dialog_box.NewMessageBox(title, content, i.Window())
	w.SetContentCopyable(true)
	if w.Exec() != 0 {
		fmt.Println("Yes button is pressed")
	} else {
		fmt.Println("Cancel button is pressed")
	}
}

func (i *DialogInterface) showCustomDialog() {
	w := NewCustomMessageBox(i.Window())
	if w.Exec() != 0 {
		fmt.Println(w.urlLineEdit.Text())
	}
}

func (i *DialogInterface) showColorDialog() {
	w := dialog_box.NewColorDialog(qt.NewQColor6("cyan"), gallerycommon.Tr("DialogInterface", "Choose color"), i.Window(), false)
	w.OnColorChanged = func(c *qt.QColor) { fmt.Println(c.Name()) }
	w.Exec()
}

func (i *DialogInterface) showBottomTeachingTip() {
	widgets.TeachingTipCreate(
		i.teachingButton.QWidget,
		"Lesson 4",
		gallerycommon.Tr("DialogInterface", "With respect, let's advance towards a new stage of the spin."),
		widgets.InfoBarIconSuccess,
		nil,
		true,
		-1,
		widgets.TeachingTipTailBottom,
		i.QWidget,
		true,
	)
}

func (i *DialogInterface) showLeftBottomTeachingTip() {
	view := widgets.NewTeachingTipView(
		"Lesson 5",
		gallerycommon.Tr("DialogInterface", "The shortest shortcut is to take a detour."),
		nil,
		resource.Pixmap("Gyro.jpg"),
		true,
		widgets.TeachingTipTailLeftBottom,
		nil,
	)

	button := widgets.NewPushButtonText("Action", nil)
	button.SetFixedWidth(120)
	view.AddWidget(button.QWidget, 0, qt.AlignRight)

	tip := widgets.TeachingTipMake(view.FlyoutViewBase, i.teachingRightButton.QWidget, 3000, widgets.TeachingTipTailLeftBottom, i.QWidget, true)
	view.SetOnClosed(func() { tip.Close() })
}

func (i *DialogInterface) showSimpleFlyout() {
	widgets.FlyoutCreate(
		"Lesson 3",
		gallerycommon.Tr("DialogInterface", "Believe in the spin, just keep believing!"),
		widgets.InfoBarIconSuccess,
		nil,
		true,
		i.simpleFlyoutButton.QWidget,
		i.Window(),
		widgets.FlyoutAnimationDropDown,
		true,
	)
}

func (i *DialogInterface) showComplexFlyout() {
	view := widgets.NewFlyoutView(
		gallerycommon.Tr("DialogInterface", "Julius·Zeppeli"),
		"触网而起的网球会落到哪一侧，谁也无法知晓。\n如果那种时刻到来，我希望「女神」是存在的。\n这样的话，不管网球落到哪一边，我都会坦然接受的吧。",
		nil,
		resource.Pixmap("SBR.jpg"),
		false,
		nil,
	)

	button := widgets.NewPushButtonText("Action", nil)
	button.SetFixedWidth(120)
	view.AddWidget(button.QWidget, 0, qt.AlignRight)

	widgets.FlyoutMake(view.FlyoutViewBase, i.complexFlyoutButton.QWidget, i.Window(), widgets.FlyoutAnimationSlideRight, true)
}

// CustomMessageBox is a message box with a URL line edit (port of
// app/view/dialog_interface.py CustomMessageBox).
type CustomMessageBox struct {
	*dialog_box.MessageBoxBase
	urlLineEdit *widgets.LineEdit
}

// NewCustomMessageBox builds a custom message box.
func NewCustomMessageBox(parent *qt.QWidget) *CustomMessageBox {
	w := &CustomMessageBox{MessageBoxBase: dialog_box.NewMessageBoxBase(parent)}

	titleLabel := widgets.NewSubtitleLabelText(gallerycommon.Tr("CustomMessageBox", "Open URL"), w.Widget().QWidget)
	w.urlLineEdit = widgets.NewLineEdit(w.Widget().QWidget)

	w.urlLineEdit.SetPlaceholderText(gallerycommon.Tr("CustomMessageBox", "Enter the URL of a file, stream, or playlist"))
	w.urlLineEdit.SetClearButtonEnabled(true)

	w.ViewLayout().AddWidget(titleLabel.QWidget)
	w.ViewLayout().AddWidget(w.urlLineEdit.QWidget)

	w.Widget().SetMinimumWidth(360)

	// The underlying yes/cancel button labels are not exposed by the library,
	// so the "Open"/"Cancel" text customization is dropped; URL validity is
	// enforced through the Validate hook instead.
	w.ValidateFunc = func() bool {
		return qt.NewQUrl3(w.urlLineEdit.Text()).IsValid()
	}
	return w
}
