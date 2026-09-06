// Command gallery is the gofluent port of PyQt-Fluent-Widgets' gallery demo
// (examples/gallery/demo.py). It wires up High-DPI scaling, the translators,
// the theme-change style refresh and the MainWindow, then runs the Qt event
// loop.
package main

import (
	"os"
	"strconv"

	gcommon "github.com/famei/gofluent/common"
	"github.com/famei/gofluent/examples/gallery/app/common"
	"github.com/famei/gofluent/examples/gallery/app/view"
	"github.com/famei/gofluent/examples/internal/demo"

	qt "github.com/mappu/miqt/qt"
)

func main() {
	cfg := common.ConfigInstance

	// Enable DPI scaling, mirroring demo.py.
	if cfg.Get(cfg.DpiScale) == "Auto" {
		qt.QGuiApplication_SetHighDpiScaleFactorRoundingPolicy(qt.PassThrough)
		qt.QCoreApplication_SetAttribute(qt.AA_EnableHighDpiScaling)
	} else {
		os.Setenv("QT_ENABLE_HIGHDPI_SCALING", "0")
		if v, ok := cfg.Get(cfg.DpiScale).(float64); ok {
			os.Setenv("QT_SCALE_FACTOR", strconv.FormatFloat(v, 'f', -1, 64))
		}
	}
	qt.QCoreApplication_SetAttribute(qt.AA_UseHighDpiPixmaps)

	demo.NewApp()
	qt.QCoreApplication_SetAttribute(qt.AA_DontCreateNativeWidgetSiblings)

	// Internationalization. The translators are installed once here and then
	// reloaded by common.SwitchLanguage when the user changes the language in
	// the settings interface.
	common.InstallTranslators(cfg.Get(cfg.Language).(common.Language).Locale())

	// The theme-color card updates the config singleton directly (not through
	// SetThemeColor), so re-apply stylesheets when the color changes. The theme
	// MODE change is handled by setting_interface.go's OnThemeChanged -> SetTheme
	// (which also emits themeChangedFinished); registering UpdateStyleSheet here
	// too would re-apply every registered stylesheet a second time.
	gcommon.QConfigInstance.OnThemeColorChanged(func(c *qt.QColor) { gcommon.UpdateStyleSheet(false) })

	// Language switching rebuilds the main window so every translated string is
	// re-resolved against the freshly installed translators (the Go equivalent
	// of the languageChanged → retranslate round-trip).
	common.OnLanguageChanged(func() { rebuildMainWindow() })
	rebuildMainWindow()
	demo.Exec()
}

// currentWindow is the active MainWindow, rebuilt on every language change.
var currentWindow *view.MainWindow

// rebuildMainWindow closes the previous window and builds a fresh one whose
// Tr() lookups resolve against the newly installed translators.
func rebuildMainWindow() {
	// Disable quit-on-last-window-closed while swapping windows so closing the
	// old window (which also stops its system-theme listener via OnCloseEvent)
	// never terminates the application.
	qt.QGuiApplication_SetQuitOnLastWindowClosed(false)
	if currentWindow != nil {
		currentWindow.Close()
		currentWindow = nil
	}
	currentWindow = view.NewMainWindow()
	qt.QGuiApplication_SetQuitOnLastWindowClosed(true)
}
