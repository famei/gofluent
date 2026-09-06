// Command progress_ring migrates examples/status_info/progress_ring/demo.py.
package main

import (
	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
	qt "github.com/mappu/miqt/qt"
)

type progressRingDemo struct {
	*qt.QWidget
	progressRing *widgets.ProgressRing
	button       *widgets.ToggleToolButton
}

func newDemo() *progressRingDemo {
	w := &progressRingDemo{QWidget: qt.NewQWidget2()}

	vBoxLayout := qt.NewQVBoxLayout(w.QWidget)
	hBoxLayout := qt.NewQHBoxLayout2()

	w.button = widgets.NewToggleToolButtonIcon(common.PauseBold, w.QWidget)
	spinner := widgets.NewIndeterminateProgressRing(w.QWidget, true)
	w.progressRing = widgets.NewProgressRing(w.QWidget, true)
	spinBox := widgets.NewSpinBox(w.QWidget)

	w.progressRing.SetValue(50)
	w.progressRing.SetTextVisible(true)
	w.progressRing.SetFixedSize2(80, 80)

	spinBox.SetRange(0, 100)
	spinBox.SetValue(50)
	spinBox.OnValueChanged(func(value int) { w.progressRing.SetValue(value) })

	hBoxLayout.AddWidget3(w.progressRing.QWidget, 0, qt.AlignHCenter)
	hBoxLayout.AddWidget3(spinBox.QWidget, 0, qt.AlignHCenter)

	vBoxLayout.SetContentsMargins(30, 30, 30, 30)
	vBoxLayout.AddLayout(hBoxLayout.QLayout)
	vBoxLayout.AddWidget3(spinner.QWidget, 0, qt.AlignHCenter)
	vBoxLayout.AddWidget3(w.button.QWidget, 0, qt.AlignHCenter)
	w.Resize(400, 400)

	w.button.OnClicked(w.onButtonClicked)
	return w
}

func (d *progressRingDemo) onButtonClicked() {
	if !d.progressRing.IsPaused() {
		d.progressRing.Pause()
		d.button.SetIcon(common.PlaySolid)
	} else {
		d.progressRing.Resume()
		d.button.SetIcon(common.PauseBold)
	}
}

func main() {
	demo.Run(func() *qt.QWidget { return newDemo().QWidget })
}
