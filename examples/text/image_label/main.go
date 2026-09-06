// Command image_label migrates examples/text/image_label/demo.py.
package main

import (
	"embed"

	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/asset"
	"github.com/famei/gofluent/examples/internal/demo"
	qt "github.com/mappu/miqt/qt"
)

//go:embed resource/Gyro.jpg resource/boqi.gif
var imageFS embed.FS

func newDemo() *qt.QWidget {
	w := qt.NewQWidget2()

	imageLabel := widgets.NewImageLabelImage(asset.QImage(imageFS, "resource/Gyro.jpg"), nil)
	gifLabel := widgets.NewImageLabelImage(asset.QImage(imageFS, "resource/boqi.gif"), nil)
	vBoxLayout := qt.NewQVBoxLayout(w)

	imageLabel.ScaledToHeight(300)
	gifLabel.ScaledToHeight(300)

	imageLabel.SetBorderRadius(0, 30, 30, 0)
	gifLabel.SetBorderRadius(10, 10, 10, 10)

	vBoxLayout.AddWidget(imageLabel.QWidget)
	vBoxLayout.AddWidget(gifLabel.QWidget)

	return w
}

func main() {
	demo.Run(newDemo)
}
