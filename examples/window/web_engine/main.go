// Command web_engine is a placeholder for PyQt-Fluent-Widgets' examples/window/
// web_engine demo.
//
// The upstream demo embeds qframelesswindow.webengine.FramelessWebEngineView,
// which wraps QWebEngineView. miqt does ship Qt WebEngine bindings (see
// miqt-master/qt/webengine/gen_qwebengineview.go: NewQWebEngineView and
// (*QWebEngineView).Load), but the MXE static Qt5 used for the windows/amd64
// cross-compile does not include QtWebEngine (Chromium is not statically
// linkable through MXE). Importing the webengine package would therefore break
// `go vet --tags=windowsqtstatic ./...`.
//
// This port keeps the SplitFluentWindow shell and degrades the web view to a
// placeholder label. Re-introduce the `github.com/mappu/miqt/qt/webengine`
// import and replace the label with webengine.NewQWebEngineView(parent) once a
// WebEngine-capable static Qt is available.
package main

import (
	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/navigation"
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/asset"
	"github.com/famei/gofluent/examples/internal/demo"
	"github.com/famei/gofluent/resources"
	gfwindow "github.com/famei/gofluent/window"
	qt "github.com/mappu/miqt/qt"
)

// Widget is the placeholder sub-interface. The Python port hosts a
// FramelessWebEngineView here.
type Widget struct {
	*qt.QWidget
	vBoxLayout *qt.QVBoxLayout
}

func newWidget(parent *qt.QWidget) *Widget {
	w := &Widget{QWidget: qt.NewQWidget(parent)}
	w.SetObjectName("homeInterface")

	w.vBoxLayout = qt.NewQVBoxLayout(w.QWidget)
	// The margins are wider than the frameless window's 5px resize border: the web
	// view is a native child window, so it owns the mouse over the whole area it
	// covers, and with the upstream 4px margins it covered the border and left the
	// window practically un-resizable.
	w.vBoxLayout.SetContentsMargins(8, 48, 8, 8)

	// Placeholder for webengine.NewQWebEngineView(w.QWidget) + Load(QUrl).
	wv := widgets.NewQWebEngineView(w.QWidget)
	wv.SetUrl(`https://www.bilibili.com`)
	w.vBoxLayout.AddWidget(wv.QWidget)
	return w
}

// Window is the SplitFluentWindow demo window.
type Window struct {
	*gfwindow.SplitFluentWindow
	homeInterface *Widget
}

func newWindow() *Window {
	w := &Window{SplitFluentWindow: gfwindow.NewSplitFluentWindow(nil)}
	w.homeInterface = newWidget(w.QWidget)

	w.initNavigation()
	w.initWindow()
	return w
}

func (w *Window) initNavigation() {
	w.AddSubInterface(w.homeInterface.QWidget, common.Home, "Home", navigation.NavigationItemPositionTop, nil, false)
}

func (w *Window) initWindow() {
	w.Resize(900, 700)
	w.SetWindowIcon(asset.QIcon(resources.Images, "images/logo.png"))
	w.SetWindowTitle("PyQt-Fluent-Widgets")

	desktop := qt.QApplication_Desktop().AvailableGeometry2()
	wd, ht := desktop.Width(), desktop.Height()
	w.Move(wd/2-w.Width()/2, ht/2-w.Height()/2)

	w.SetMicaEffectEnabled(true)
	// The embedded browser subclasses the window procedure and answers WM_NCHITTEST
	// itself, so both the drag and the border resize are started from Qt mouse events
	// instead of the native hit test.
	w.SetMouseMoveDragEnabled(true)
	w.SetMouseMoveResizeEnabled(true)
}

func main() {
	demo.Run(func() *qt.QWidget { return newWindow().QWidget })
}
