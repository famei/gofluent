// Command avatar_widget migrates examples/media/avatar_widget/demo.py. It shows
// a single avatar image (resource/shoko.png) rendered at four radii (96, 48, 32,
// 24) laid out in a horizontal row.
package main

import (
	_ "embed"

	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
	qt "github.com/mappu/miqt/qt"
)

//go:embed resource/shoko.png
var shokoPNG []byte

func newDemo() *qt.QWidget {
	w := qt.NewQWidget2()
	w.SetObjectName("Demo")
	w.SetStyleSheet("#Demo{background: white}")
	w.Resize(400, 300)

	hBoxLayout := qt.NewQHBoxLayout(w)

	// Decode the shared avatar image once. Each AvatarWidget keeps the *qt.QImage
	// pointer for its paint event (SetImage stores it without copying), so the
	// image is intentionally not Delete()d — it must outlive the four widgets.
	avatar := qt.QImage_FromDataWithData(shokoPNG)

	for _, s := range []int{96, 48, 32, 24} {
		a := widgets.NewAvatarWidgetImage(avatar, w)
		a.SetRadius(s / 2)
		hBoxLayout.AddWidget(a.QWidget)
	}
	return w
}

func main() {
	demo.Run(newDemo)
}
