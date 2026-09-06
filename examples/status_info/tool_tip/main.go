// Command tool_tip migrates examples/status_info/tool_tip/demo.py.
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

	hBox := qt.NewQHBoxLayout(w)
	button1 := widgets.NewPushButtonText("キラキラ", w)
	button2 := widgets.NewPushButtonText("食べた愛", w)
	button3 := widgets.NewPushButtonText("シアワセ", w)

	button1.SetToolTip("aiko - キラキラ ✨")
	button2.SetToolTip("aiko - 食べた愛 🥰")
	button3.SetToolTip("aiko - シアワセ 😊")
	button1.SetToolTipDuration(1000)

	// NewToolTipFilter installs itself as the widget's event filter.
	_ = widgets.NewToolTipFilter(button1.QWidget, 0, widgets.ToolTipPositionTop)
	_ = widgets.NewToolTipFilter(button2.QWidget, 0, widgets.ToolTipPositionBottom)
	_ = widgets.NewToolTipFilter(button3.QWidget, 300, widgets.ToolTipPositionRight)

	button1.OnClicked(func() {
		qt.QDesktopServices_OpenUrl(qt.NewQUrl3("https://www.youtube.com/watch?v=S0bXDRY1DGM&list=RDMM&index=1"))
	})
	button2.OnClicked(func() {
		qt.QDesktopServices_OpenUrl(qt.NewQUrl3("https://www.youtube.com/watch?v=CZLs8GuCq2U&list=RDMM&index=4"))
	})
	button3.OnClicked(func() {
		qt.QDesktopServices_OpenUrl(qt.NewQUrl3("https://www.youtube.com/watch?v=fp-yJUB7sS8&list=RDMM&index=3"))
	})

	hBox.SetContentsMargins(24, 24, 24, 24)
	hBox.SetSpacing(16)
	hBox.AddWidget(button1.QWidget)
	hBox.AddWidget(button2.QWidget)
	hBox.AddWidget(button3.QWidget)

	w.Resize(480, 240)
	return w
}

func main() {
	demo.Run(newDemo)
}
