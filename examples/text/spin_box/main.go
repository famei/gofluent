// Command spin_box migrates examples/text/spin_box/demo.py.
package main

import (
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
	qt "github.com/mappu/miqt/qt"
)

func newDemo() *qt.QWidget {
	w := qt.NewQWidget2()
	w.SetObjectName("Demo")
	w.SetStyleSheet("#Demo{background: rgb(255, 255, 255)}")

	gridLayout := qt.NewQGridLayout(w)

	spinBox := widgets.NewSpinBox(w)
	compactSpinBox := widgets.NewCompactSpinBox(w)
	spinBox.SetAccelerated(true)
	compactSpinBox.SetAccelerated(true)

	timeEdit := widgets.NewTimeEdit(w)
	compactTimeEdit := widgets.NewCompactTimeEdit(w)

	dateEdit := widgets.NewDateEdit(w)
	compactDateEdit := widgets.NewCompactDateEdit(w)

	dateTimeEdit := widgets.NewDateTimeEdit(w)
	compactDateTimeEdit := widgets.NewCompactDateTimeEdit(w)

	doubleSpinBox := widgets.NewDoubleSpinBox(w)
	compactDoubleSpinBox := widgets.NewCompactDoubleSpinBox(w)

	w.Resize(500, 500)
	gridLayout.SetHorizontalSpacing(30)

	gridLayout.SetContentsMargins(100, 50, 100, 50)
	gridLayout.AddWidget2(spinBox.QWidget, 0, 0)
	gridLayout.AddWidget4(compactSpinBox.QWidget, 0, 1, qt.AlignLeft)

	gridLayout.AddWidget2(doubleSpinBox.QWidget, 1, 0)
	gridLayout.AddWidget4(compactDoubleSpinBox.QWidget, 1, 1, qt.AlignLeft)

	gridLayout.AddWidget2(timeEdit.QWidget, 2, 0)
	gridLayout.AddWidget4(compactTimeEdit.QWidget, 2, 1, qt.AlignLeft)

	gridLayout.AddWidget2(dateEdit.QWidget, 3, 0)
	gridLayout.AddWidget4(compactDateEdit.QWidget, 3, 1, qt.AlignLeft)

	gridLayout.AddWidget2(dateTimeEdit.QWidget, 4, 0)
	gridLayout.AddWidget4(compactDateTimeEdit.QWidget, 4, 1, qt.AlignLeft)

	return w
}

func main() {
	demo.Run(newDemo)
}
