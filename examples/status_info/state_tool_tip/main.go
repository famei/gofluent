// Command state_tool_tip migrates examples/status_info/state_tool_tip/demo.py.
package main

import (
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
	qt "github.com/mappu/miqt/qt"
)

type stateToolTipDemo struct {
	*qt.QWidget
	stateTooltip *widgets.StateToolTip
}

func newDemo() *stateToolTipDemo {
	w := &stateToolTipDemo{QWidget: qt.NewQWidget2()}
	w.Resize(800, 300)
	w.SetObjectName("Demo")
	w.SetStyleSheet("#Demo{background: white}")

	btn := widgets.NewPushButtonText("Click Me", w.QWidget)
	btn.Move(360, 225)
	btn.OnClicked(w.onButtonClicked)

	return w
}

func (d *stateToolTipDemo) onButtonClicked() {
	if d.stateTooltip != nil {
		d.stateTooltip.SetContent("模型训练完成啦 😆")
		d.stateTooltip.SetState(true)
		d.stateTooltip = nil
	} else {
		d.stateTooltip = widgets.NewStateToolTip("正在训练模型", "客官请耐心等待哦~~", d.QWidget)
		d.stateTooltip.Move(510, 30)
		d.stateTooltip.Show()
	}
}

func main() {
	demo.Run(func() *qt.QWidget { return newDemo().QWidget })
}
