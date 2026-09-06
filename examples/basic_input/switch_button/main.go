// Command switch_button migrates examples/basic_input/switch_button/demo.py.
package main

import (
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
	qt "github.com/mappu/miqt/qt"
)

type window struct {
	*qt.QWidget
	switchButton *widgets.SwitchButton
}

func newWindow() *window {
	w := &window{QWidget: qt.NewQWidget2()}
	w.Resize(160, 80)
	w.switchButton = widgets.NewSwitchButton(w.QWidget, widgets.RIGHT)
	w.switchButton.Move(48, 24)
	w.switchButton.OnCheckedChanged(func(isChecked bool) {
		text := "Off"
		if isChecked {
			text = "On"
		}
		w.switchButton.SetText(text)
	})
	return w
}

func main() {
	demo.Run(func() *qt.QWidget { return newWindow().QWidget })
}
