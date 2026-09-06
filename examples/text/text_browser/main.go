// Command text_browser migrates examples/text/text_browser/demo.py.
package main

import (
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
	qt "github.com/mappu/miqt/qt"
)

func newDemo() *qt.QWidget {
	w := qt.NewQWidget2()

	hBoxLayout := qt.NewQHBoxLayout(w)
	textBrowser := widgets.NewTextBrowser(w)

	w.Resize(400, 400)
	hBoxLayout.AddWidget3(textBrowser.QWidget, 0, qt.AlignCenter)

	textBrowser.SetPlaceholderText("Search stand")
	textBrowser.SetMarkdown("## Steel Ball Run \n * Johnny Joestar 🦄 \n * Gyro Zeppeli 🐴")

	return w
}

func main() {
	demo.Run(newDemo)
}
