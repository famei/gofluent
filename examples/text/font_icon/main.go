// Command font_icon migrates examples/text/font_icon/demo.py: it shows custom
// icon fonts (PhotosIcons / MediaPlayerIcons) loaded from embedded .ttf files.
//
// The Python FluentFontIconBase subclasses map to common.FluentFontIconBase;
// the PhotoIcons.json name map is inlined (cloud = \ue753, smile = \ue76e).
package main

import (
	_ "embed"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
	qt "github.com/mappu/miqt/qt"
)

//go:embed font/PhotosIcons.ttf
var photosTTF []byte

//go:embed font/MediaPlayerIcons.ttf
var mediaTTF []byte

func newDemo() *qt.QWidget {
	w := qt.NewQWidget2()
	w.SetObjectName("Demo")
	w.SetStyleSheet("#Demo{background: rgb(242,242,242)}")

	// Load the two custom fonts and resolve their family names (the font ID
	// returned by AddApplicationFontFromData is resolved via
	// ApplicationFontFamilies, equivalent to the Python FluentFontIconBase
	// loadFont logic).
	photoID := qt.QFontDatabase_AddApplicationFontFromData(photosTTF)
	photoFamily := qt.QFontDatabase_ApplicationFontFamilies(photoID)[0]

	mediaID := qt.QFontDatabase_AddApplicationFontFromData(mediaTTF)
	mediaFamily := qt.QFontDatabase_ApplicationFontFamilies(mediaID)[0]

	themeButton := widgets.NewSwitchButton(w, widgets.RIGHT)

	photoDefault := common.NewFluentFontIconBase("\ue77b", photoFamily)
	photoCloud := common.NewFluentFontIconBase("\ue753", photoFamily).Colored(
		qt.NewQColor6("#275EFF"), qt.NewQColor3(0, 139, 139))
	photoSmile := common.NewFluentFontIconBase("\ue76e", photoFamily)
	mediaPlay := common.NewFluentFontIconBase("\uf414", mediaFamily)

	button1 := widgets.NewPushButtonIcon(photoDefault, "Default", nil)
	button2 := widgets.NewPushButtonIcon(photoCloud, "Custom", nil)
	button3 := widgets.NewToggleButtonIcon(photoSmile, "Toggle", nil)
	button4 := widgets.NewHyperlinkButtonIcon(mediaPlay, "http://qfluentwidgets.com", "Hyperlink", nil)
	hBoxLayout := qt.NewQHBoxLayout(w)

	hBoxLayout.AddWidget(button1.QWidget)
	hBoxLayout.AddWidget(button2.QWidget)
	hBoxLayout.AddWidget(button3.QWidget)
	hBoxLayout.AddWidget(button4.QWidget)

	w.Resize(500, 500)
	themeButton.Move(200, 50)
	themeButton.SetOnText("Dark")
	themeButton.SetOffText("Light")

	themeButton.OnCheckedChanged(func(isChecked bool) {
		common.ToggleTheme(false, false)
		if isChecked {
			w.SetStyleSheet("#Demo{background: rgb(32,32,32)}")
		} else {
			w.SetStyleSheet("#Demo{background: rgb(242,242,242)}")
		}
	})

	return w
}

func main() {
	// Fonts must be loaded after the QApplication exists (miqt requires the
	// application before QFontDatabase.AddApplicationFontFromData), so the
	// split entry point is used instead of demo.Run.
	demo.SetupHighDPI()
	_ = demo.NewApp()
	w := newDemo()
	w.Show()
	demo.Exec()
}
