// Command label migrates examples/text/label/demo.py.
package main

import (
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
	qt "github.com/mappu/miqt/qt"
)

func newDemo() *qt.QWidget {
	w := qt.NewQWidget2()
	vBoxLayout := qt.NewQVBoxLayout(w)
	vBoxLayout.SetContentsMargins(30, 30, 30, 30)
	vBoxLayout.SetSpacing(20)

	hyperlinkLabel := widgets.NewHyperlinkLabelURL("https://github.com/", "GitHub", nil)

	vBoxLayout.AddWidget(hyperlinkLabel.QWidget)
	vBoxLayout.AddWidget(widgets.NewCaptionLabelText("Caption", nil).QWidget)
	vBoxLayout.AddWidget(widgets.NewBodyLabelText("Body", nil).QWidget)
	vBoxLayout.AddWidget(widgets.NewStrongBodyLabelText("Body Strong", nil).QWidget)
	vBoxLayout.AddWidget(widgets.NewSubtitleLabelText("Subtitle", nil).QWidget)
	vBoxLayout.AddWidget(widgets.NewTitleLabelText("Title", nil).QWidget)
	vBoxLayout.AddWidget(widgets.NewLargeTitleLabelText("Title Large", nil).QWidget)
	vBoxLayout.AddWidget(widgets.NewDisplayLabelText("Display", nil).QWidget)

	return w
}

func main() {
	demo.Run(newDemo)
}
