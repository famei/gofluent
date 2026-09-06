// Command adaptive_flow_layout migrates examples/layout/adaptive_flow_layout/demo.py.
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

	adaptive := layout.NewAdaptiveFlowLayout(w, false, false)

	adaptive.SetWidgetMinimumWidth(150)
	adaptive.SetContentsMargins(30, 30, 30, 30)
	adaptive.SetVerticalSpacing(20)
	adaptive.SetHorizontalSpacing(10)

	adaptive.AddWidget(widgets.NewPushButtonText("aiko", nil).QWidget)
	adaptive.AddWidget(widgets.NewPushButtonText("刘静爱", nil).QWidget)
	adaptive.AddWidget(widgets.NewPushButtonText("柳井爱子", nil).QWidget)
	adaptive.AddWidget(widgets.NewPushButtonText("aiko 赛高", nil).QWidget)
	adaptive.AddWidget(widgets.NewPushButtonText("aiko 太爱啦😘", nil).QWidget)

	adaptive.InsertWidget(1, widgets.NewPrimaryPushButtonText("西宫硝子", nil).QWidget)

	w.Resize(400, 300)
	w.SetStyleSheet(`#Demo{background: white} QPushButton{padding: 5px 10px; font:15px "Microsoft YaHei"}`)
	return w
}

func main() {
	demo.Run(newDemo)
}
