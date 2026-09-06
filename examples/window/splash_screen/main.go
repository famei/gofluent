// Command splash_screen ports PyQt-Fluent-Widgets' examples/window/
// splash_screen demo: a frameless window that shows a centered splash screen
// for three seconds while its sub-interface is "created".
package main

import (
	qt "github.com/mappu/miqt/qt"

	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/asset"
	"github.com/famei/gofluent/examples/internal/demo"
	"github.com/famei/gofluent/resources"
	gfwindow "github.com/famei/gofluent/window"
)

// Window is the frameless demo window.
type Window struct {
	*widgets.FramelessWindow
	splashScreen *gfwindow.SplashScreen
}

func newWindow(icon *qt.QIcon) *Window {
	w := &Window{FramelessWindow: widgets.NewFramelessWindow(nil)}
	w.Resize(700, 600)
	w.SetWindowTitle("PyQt-Fluent-Widgets")
	w.SetWindowIcon(icon)

	// The reference FramelessWindow auto-creates a StandardTitleBar (icon +
	// title + min/max/close); the Go FramelessWindow has no default title bar,
	// so add a FluentTitleBar so the window keeps its system buttons after the
	// splash screen finishes.
	titleBar := gfwindow.NewFluentTitleBar(w.QWidget)
	titleBar.SetTitle("PyQt-Fluent-Widgets")
	titleBar.SetIcon(icon)
	w.SetTitleBar(titleBar.QWidget)

	// create splash screen and show the window
	w.splashScreen = gfwindow.NewSplashScreen(icon, w.QWidget, true)
	w.splashScreen.SetIconSize(qt.NewQSize2(102, 102))
	return w
}

func main() {
	demo.SetupHighDPI()
	app := demo.NewApp()

	logo := asset.QIcon(resources.Images, "images/logo.png")
	w := newWindow(logo)
	w.Show()

	// The Python port blocks on a QEventLoop for three seconds to simulate
	// creating sub-interfaces, then closes the splash screen. Approximate that
	// with a single-shot timer so the splash stays visible after the event loop
	// starts.
	timer := qt.NewQTimer2(app.QObject)
	timer.SetSingleShot(true)
	timer.OnTimeout(func() { w.splashScreen.Finish() })
	timer.Start(3000)

	demo.Exec()
}
