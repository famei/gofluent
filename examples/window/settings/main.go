// Command settings ports PyQt-Fluent-Widgets' examples/window/settings
// multi-file demo: a frameless window hosting the SettingInterface, with
// theme-aware QSS and i18n translation loading.
package main

import (
	qt "github.com/mappu/miqt/qt"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/asset"
	"github.com/famei/gofluent/examples/internal/demo"
	"github.com/famei/gofluent/resources"
	gfwindow "github.com/famei/gofluent/window"
)

// Window is the settings demo window.
type Window struct {
	*widgets.FramelessWindow

	settingInterface *SettingInterface
	hBoxLayout       *qt.QHBoxLayout
}

func newWindow() *Window {
	w := &Window{FramelessWindow: widgets.NewFramelessWindow(nil)}
	w.SetTitleBar(gfwindow.NewTitleBar(w.QWidget).QWidget)

	w.hBoxLayout = qt.NewQHBoxLayout(w.QWidget)
	w.settingInterface = NewSettingInterface(w.QWidget)
	w.hBoxLayout.SetContentsMargins(0, 0, 0, 0)
	w.hBoxLayout.AddWidget(w.settingInterface.QWidget)

	w.SetWindowIcon(asset.QIcon(resources.Images, "images/logo.png"))
	w.SetWindowTitle("PyQt-Fluent-Widgets")

	w.Resize(1080, 784)
	desktop := qt.QApplication_Desktop().AvailableGeometry2()
	wd, ht := desktop.Width(), desktop.Height()
	w.Move(wd/2-w.Width()/2, ht/2-w.Height()/2)

	w.TitleBar().Raise()

	w.setQss()
	common.QConfigInstance.OnThemeChanged(func(common.Theme) { w.setQss() })
	return w
}

// setQss applies the theme-specific demo stylesheet (the Python demo.py
// implementation).
func (w *Window) setQss() {
	dir := "light"
	if common.IsDarkTheme() {
		dir = "dark"
	}
	w.SetStyleSheet(asset.String(settingsQSS, "resource/qss/"+dir+"/demo.qss"))
}

// localeForLanguage converts a Language config value to a QLocale (nil for
// auto/system).
func localeForLanguage(lang Language) *qt.QLocale {
	if name := lang.LocaleName(); name != "" {
		return qt.NewQLocale2(name)
	}
	return nil
}

// loadSettingsTranslator loads the settings .qm file for the configured
// language (only zh_CN and zh_HK are shipped).
func loadSettingsTranslator(app *qt.QApplication) *qt.QTranslator {
	lang := cfg.Language.Value().(Language)
	name := lang.LocaleName()

	t := qt.NewQTranslator2(app.QObject)
	if name == "zh_CN" || name == "zh_HK" {
		data := asset.Bytes(settingsI18n, "resource/i18n/settings."+name+".qm")
		t.Load3(&data[0], len(data))
	}
	return t
}

func main() {
	demo.SetupHighDPI()
	app := demo.NewApp()
	qt.QCoreApplication_SetAttribute(qt.AA_DontCreateNativeWidgetSiblings)

	// internationalization (mirrors the Python demo.py flow).
	fluentTranslator := common.NewFluentTranslator(localeForLanguage(cfg.Language.Value().(Language)), app.QObject)
	qt.QCoreApplication_InstallTranslator(fluentTranslator.QTranslator)
	qt.QCoreApplication_InstallTranslator(loadSettingsTranslator(app))

	// Load the persisted theme (and other built-in config) so a theme chosen
	// earlier survives a restart, mirroring config.py's
	// qconfig.load('config/config.json', cfg). The extra demo items stay
	// in-memory (documented in config.go).
	common.QConfigInstance.Load("config/config.json")

	w := newWindow()
	w.Show()
	demo.Exec()
}
