// Command color_dialog migrates examples/dialog_flyout/color_dialog/demo.py:
// a ColorPickerButton that opens a ColorDialog with alpha channel editing.
package main

import (
	qt "github.com/mappu/miqt/qt"

	"github.com/famei/gofluent/components/settings"
	"github.com/famei/gofluent/examples/internal/demo"
)

func newDemo() *qt.QWidget {
	w := qt.NewQWidget2()
	w.SetObjectName("Demo")
	w.Resize(800, 720)
	w.SetStyleSheet("#Demo{background: white}")

	color := qt.NewQColor6("#5012aaa2")
	defer color.Delete()
	button := settings.NewColorPickerButton(color, "Background Color", w, true)
	button.Move(352, 312)

	return w
}

func main() {
	demo.Run(newDemo)
}
