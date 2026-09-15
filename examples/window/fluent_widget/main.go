// Command fluent_widget ports PyQt-Fluent-Widgets' examples/window/fluent_widget
// demo: a plain FluentWidget whose single button toggles the light/dark theme.
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

// Window is the FluentWidget demo window.
type Window struct {
	*gfwindow.FluentWidget
	button     *widgets.PushButton
	vBoxLayout *qt.QVBoxLayout
}

func newWindow() *Window {
	w := &Window{FluentWidget: gfwindow.NewFluentWidget(nil)}
	w.button = widgets.NewPushButtonIcon(common.Light, "Toggle theme", w.QWidget)
	w.vBoxLayout = qt.NewQVBoxLayout(w.QWidget)

	// toggle theme when the button is clicked
	w.button.OnClicked(func() { common.ToggleTheme(false, false) })

	// leave some space for the title bar
	w.vBoxLayout.SetContentsMargins(0, w.TitleBar().Height(), 0, 0)
	w.vBoxLayout.AddWidget3(w.button.QWidget, 0, qt.AlignCenter)

	w.initWindow()
	return w
}

func (w *Window) initWindow() {
	w.Resize(900, 700)
	w.SetWindowIcon(asset.QIcon(resources.Images, "images/logo.png"))
	w.SetWindowTitle("PyQt-Fluent-Widgets")

	desktop := qt.QApplication_Desktop().AvailableGeometry2()
	wd, ht := desktop.Width(), desktop.Height()
	w.Move(wd/2-w.Width()/2, ht/2-w.Height()/2)
}

func main() {
	demo.Run(func() *qt.QWidget { return newWindow().QWidget })
}
