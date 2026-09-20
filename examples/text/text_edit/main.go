// Command text_edit migrates examples/text/text_edit/demo.py: a markdown
// QTextEdit with the fluent style.
package main

import (
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
	qt "github.com/mappu/miqt/qt"
)

func newDemo() *qt.QWidget {
	w := qt.NewQWidget2()

	hBoxLayout := qt.NewQHBoxLayout(w)
	textEdit := widgets.NewTextEdit(w)

	textEdit.SetFixedHeight(150)
	textEdit.SetMarkdown("## Steel Ball Run \n * Johnny Joestar 🦄 \n * Gyro Zeppeli 🐴")

	hBoxLayout.AddWidget3(textEdit.QWidget, 0, qt.AlignCenter)
	hBoxLayout.SetContentsMargins(50, 30, 50, 30)
	w.Resize(400, 400)
	return w
}

func main() {
	demo.Run(newDemo)
}
