// Command custom_message_box migrates examples/dialog_flyout/custom_message_box/demo.py:
// a MessageBoxBase subclass that validates a URL line edit before accepting.
package main

import (
	"fmt"
	"strings"

	qt "github.com/mappu/miqt/qt"

	"github.com/famei/gofluent/components/dialog_box"
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
)

// customMessageBox is the Go equivalent of the Python CustomMessageBox
// (MessageBoxBase) subclass. gofluent's MessageBoxBase exposes a ValidateFunc
// hook in place of the overridable validate() method.
type customMessageBox struct {
	*dialog_box.MessageBoxBase
	urlLineEdit  *widgets.LineEdit
	warningLabel *widgets.CaptionLabel
}

func newCustomMessageBox(parent *qt.QWidget) *customMessageBox {
	w := &customMessageBox{MessageBoxBase: dialog_box.NewMessageBoxBase(parent)}
	center := w.Widget().QWidget

	titleLabel := widgets.NewSubtitleLabelText("打开 URL", center)
	w.urlLineEdit = widgets.NewLineEdit(center)
	w.urlLineEdit.SetPlaceholderText("输入文件、流或者播放列表的 URL")
	w.urlLineEdit.SetClearButtonEnabled(true)

	w.warningLabel = widgets.NewCaptionLabelText("The url is invalid", center)
	light := qt.NewQColor6("#cf1010")
	dark := qt.NewQColor3(255, 28, 32)
	w.warningLabel.SetTextColor(light, dark)
	light.Delete()
	dark.Delete()

	viewLayout := w.ViewLayout()
	viewLayout.AddWidget(titleLabel.QWidget)
	viewLayout.AddWidget(w.urlLineEdit.QWidget)
	viewLayout.AddWidget(w.warningLabel.QWidget)
	w.warningLabel.Hide()

	// NOTE: the gofluent MessageBoxBase does not expose its yes/cancel buttons
	// for re-labelling, so the default "OK"/"Cancel" texts are kept here (the
	// Python port renames them to "打开"/"取消").
	w.Widget().SetMinimumWidth(350)

	w.ValidateFunc = func() bool {
		isValid := strings.HasPrefix(strings.ToLower(w.urlLineEdit.Text()), "http://")
		w.warningLabel.SetHidden(isValid)
		w.urlLineEdit.SetError(!isValid)
		return isValid
	}
	return w
}

func newDemo() *qt.QWidget {
	w := qt.NewQWidget2()
	w.Resize(600, 600)

	hBoxLayout := qt.NewQHBoxLayout(w)
	button := widgets.NewPushButtonText("打开 URL", w)
	hBoxLayout.AddWidget3(button.QWidget, 0, qt.AlignCenter)
	button.OnClicked(func() { showDialog(w) })

	return w
}

func showDialog(parent *qt.QWidget) {
	box := newCustomMessageBox(parent)
	if box.Exec() == int(qt.QDialog__Accepted) {
		fmt.Println(box.urlLineEdit.Text())
	}
}

func main() {
	demo.Run(newDemo)
}
