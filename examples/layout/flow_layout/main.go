// Command flow_layout migrates examples/layout/flow_layout/demo.py.
package main

import (
	"github.com/famei/gofluent/components/layout"
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
	qt "github.com/mappu/miqt/qt"
)

func newDemo() *qt.QWidget {
	w := qt.NewQWidget2()
	w.SetObjectName("Demo")

	flowLayout := layout.NewFlowLayout(w, true, false)

	// SetAnimation resets the easing curve to Linear (the Go port keeps a
	// single-argument API, mirroring Python setAnimation's default ease).
	flowLayout.SetAnimation(250)

	flowLayout.SetContentsMargins(30, 30, 30, 30)
	flowLayout.SetVerticalSpacing(20)
	flowLayout.SetHorizontalSpacing(10)

	flowLayout.AddWidget(widgets.NewPushButtonText("aiko", nil).QWidget)
	flowLayout.AddWidget(widgets.NewPushButtonText("刘静爱", nil).QWidget)
	flowLayout.AddWidget(widgets.NewPushButtonText("柳井爱子", nil).QWidget)
	flowLayout.AddWidget(widgets.NewPushButtonText("aiko 赛高", nil).QWidget)
	flowLayout.AddWidget(widgets.NewPushButtonText("aiko 太爱啦😘", nil).QWidget)

	flowLayout.InsertWidget(1, widgets.NewPrimaryPushButtonText("西宫硝子", nil).QWidget)

	w.Resize(250, 300)
	w.SetStyleSheet(`#Demo{background: white} QPushButton{padding: 5px 10px; font:15px "Microsoft YaHei"}`)
	return w
}

func main() {
	demo.Run(newDemo)
}
