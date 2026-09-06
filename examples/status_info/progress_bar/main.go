// Command progress_bar migrates examples/status_info/progress_bar/demo.py.
package main

import (
	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
	qt "github.com/mappu/miqt/qt"
)

type progressBarDemo struct {
	*qt.QWidget
	progressBar *widgets.ProgressBar
	inProgress  *widgets.IndeterminateProgressBar
	button      *widgets.ToggleToolButton
}

func newDemo() *progressBarDemo {
	w := &progressBarDemo{QWidget: qt.NewQWidget2()}
	vBoxLayout := qt.NewQVBoxLayout(w.QWidget)
	w.progressBar = widgets.NewProgressBar(w.QWidget, true)
	w.inProgress = widgets.NewIndeterminateProgressBar(w.QWidget, true)
	w.button = widgets.NewToggleToolButtonIcon(common.PauseBold, w.QWidget)

	w.progressBar.SetValue(50)
	vBoxLayout.AddWidget(w.progressBar.QWidget)
	vBoxLayout.AddWidget(w.inProgress.QWidget)
	vBoxLayout.AddWidget3(w.button.QWidget, 0, qt.AlignHCenter)
	vBoxLayout.SetContentsMargins(30, 30, 30, 30)
	w.Resize(400, 400)

	w.button.OnClicked(w.onButtonClicked)
	return w
}

func (d *progressBarDemo) onButtonClicked() {
	if d.inProgress.IsStarted() {
		d.inProgress.Pause()
		d.progressBar.Pause()
		d.button.SetIcon(common.PlaySolid)
	} else {
		d.inProgress.Resume()
		d.progressBar.Resume()
		d.button.SetIcon(common.PauseBold)
	}
}

func main() {
	demo.Run(func() *qt.QWidget { return newDemo().QWidget })
}
