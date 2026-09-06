// Command check_box migrates examples/basic_input/check_box/demo.py.
package main

import (
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
	qt "github.com/mappu/miqt/qt"
)

func newDemo() *qt.QWidget {
	w := qt.NewQWidget2()
	w.SetObjectName("Demo")
	w.SetStyleSheet("#Demo{background: white}")

	hBoxLayout := qt.NewQHBoxLayout(w)
	checkBox := widgets.NewCheckBoxText("This is a check box", w)
	checkBox.SetTristate()

	hBoxLayout.AddWidget3(checkBox.QWidget, 1, qt.AlignCenter)
	w.Resize(400, 400)
	return w
}

func main() {
	demo.Run(newDemo)
}
