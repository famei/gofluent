// Command tree_view migrates examples/view/tree_view/demo.py.
package main

import (
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
	qt "github.com/mappu/miqt/qt"
)

func newDemo() *qt.QWidget {
	w := qt.NewQWidget2()
	w.SetObjectName("Demo")
	w.SetStyleSheet("#Demo{background: rgb(255,255,255)}")

	hBoxLayout := qt.NewQHBoxLayout(w)

	view := widgets.NewTreeView(w)
	model := qt.NewQFileSystemModel()
	model.SetRootPath(".")
	view.SetModel(model.QAbstractItemModel)

	view.SetBorderVisible(true)
	view.SetBorderRadius(8)

	hBoxLayout.AddWidget(view.QWidget)
	hBoxLayout.SetContentsMargins(50, 30, 50, 30)
	w.Resize(800, 660)
	return w
}

func main() {
	demo.Run(newDemo)
}
