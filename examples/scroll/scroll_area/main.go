// Command scroll_area migrates examples/scroll/scroll_area/demo.py.
package main

import (
	"embed"
	_ "embed"

	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/asset"
	"github.com/famei/gofluent/examples/internal/demo"
	qt "github.com/mappu/miqt/qt"
)

//go:embed resource/demo.qss
var demoQSS string

//go:embed resource/shoko.jpg
var imageFS embed.FS

func newDemo() *qt.QWidget {
	area := widgets.NewSmoothScrollArea(nil)

	label := widgets.NewPixmapLabel(area.QWidget)
	label.SetPixmap(asset.QPixmap(imageFS, "resource/shoko.jpg"))

	// SetScrollAnimation drops the QEasingCurve (see MIGRATION_GUIDE §12).
	area.SetScrollAnimation(qt.Vertical, 400)
	area.SetScrollAnimation(qt.Horizontal, 400)

	area.HorizontalScrollBar().SetValue(1900)
	area.SetWidget(label.QWidget)
	area.Resize(1200, 800)
	area.SetStyleSheet(demoQSS)

	return area.QWidget
}

func main() {
	demo.Run(newDemo)
}
