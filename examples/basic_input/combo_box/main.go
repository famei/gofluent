// Command combo_box migrates examples/basic_input/combo_box/demo.py.
package main

import (
	"fmt"

	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
	qt "github.com/mappu/miqt/qt"
)

func newDemo() *qt.QWidget {
	w := qt.NewQWidget2()
	w.SetObjectName("Demo")
	w.SetStyleSheet("#Demo{background: white}")

	comboBox := widgets.NewComboBox(w)
	hBoxLayout := qt.NewQHBoxLayout(w)

	comboBox.SetPlaceholderText("选择一个脑婆")

	items := []string{"shoko 🥰", "西宫硝子", "宝多六花", "小鸟游六花"}
	comboBox.AddItems(items)
	comboBox.SetCurrentIndex(-1)

	comboBox.OnCurrentTextChanged = func(text string) { fmt.Println(text) }

	w.Resize(500, 500)
	hBoxLayout.AddWidget3(comboBox.QWidget, 0, qt.AlignCenter)
	return w
}

func main() {
	demo.Run(newDemo)
}
