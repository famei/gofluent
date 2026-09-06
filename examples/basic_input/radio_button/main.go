// Command radio_button migrates examples/basic_input/radio_button/demo.py.
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

	vBoxLayout := qt.NewQVBoxLayout(w)
	button1 := widgets.NewRadioButtonText("Option 1", w)
	button2 := widgets.NewRadioButtonText("Option 2", w)
	button3 := widgets.NewRadioButtonText("Option 3", w)

	vBoxLayout.AddWidget3(button1.QWidget, 0, qt.AlignCenter)
	vBoxLayout.AddWidget3(button2.QWidget, 0, qt.AlignCenter)
	vBoxLayout.AddWidget3(button3.QWidget, 0, qt.AlignCenter)
	w.Resize(300, 150)
	return w
}

func main() {
	demo.Run(newDemo)
}
