// Command clock ports PyQt-Fluent-Widgets' examples/window/clock multi-file
// demo: a SplitFluentWindow hosting the FocusInterface and StopWatchInterface
// sub-interfaces, plus a bottom avatar and settings entry.
package main

import (
	qt "github.com/mappu/miqt/qt"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/dialog_box"
	"github.com/famei/gofluent/components/navigation"
	"github.com/famei/gofluent/examples/internal/asset"
	"github.com/famei/gofluent/examples/internal/demo"
	"github.com/famei/gofluent/examples/window/clock/resource"
	"github.com/famei/gofluent/examples/window/clock/view"
	"github.com/famei/gofluent/resources"
	gfwindow "github.com/famei/gofluent/window"
)

// Window is the clock demo window.
type Window struct {
	*gfwindow.SplitFluentWindow

	focusInterface     *view.FocusInterface
	stopWatchInterface *view.StopWatchInterface
}

func newWindow() *Window {
	w := &Window{SplitFluentWindow: gfwindow.NewSplitFluentWindow(nil)}

	w.focusInterface = view.NewFocusInterface(w.QWidget)
	w.stopWatchInterface = view.NewStopWatchInterface(w.QWidget)

	w.initNavigation()
	w.initWindow()
	return w
}

func (w *Window) initNavigation() {
	w.AddSubInterface(w.focusInterface.QWidget, common.Ringer, "专注时段", navigation.NavigationItemPositionTop, nil, false)
	w.AddSubInterface(w.stopWatchInterface.QWidget, common.DateTime, "秒表", navigation.NavigationItemPositionTop, nil, false)

	avatar := navigation.NewNavigationAvatarWidget("zhiyiYo", resource.Shoko(), nil)
	w.NavigationInterface().AddWidget(
		"avatar",
		avatar,
		func(bool) { w.showMessageBox() },
		navigation.NavigationItemPositionBottom,
		"",
		"",
	)
	w.NavigationInterface().AddItem(
		"settingInterface",
		common.Settings,
		"设置",
		nil,
		true,
		navigation.NavigationItemPositionBottom,
		"",
		"",
	)

	w.NavigationInterface().SetExpandWidth(280)
}

func (w *Window) initWindow() {
	w.Resize(900, 700)
	w.SetWindowIcon(asset.QIcon(resources.Images, "images/logo.png"))
	w.SetWindowTitle("PyQt-Fluent-Widgets")

	desktop := qt.QApplication_Desktop().AvailableGeometry2()
	wd, ht := desktop.Width(), desktop.Height()
	w.Move(wd/2-w.Width()/2, ht/2-w.Height()/2)
}

func (w *Window) showMessageBox() {
	box := dialog_box.NewMessageBox(
		"支持作者🥰",
		"个人开发不易，如果这个项目帮助到了您，可以考虑请作者喝一瓶快乐水🥤。您的支持就是作者开发和维护项目的动力🚀",
		w.QWidget,
	)
	box.OnYes = func() {
		qt.QDesktopServices_OpenUrl(qt.NewQUrl3("https://afdian.net/a/zhiyiYo"))
	}
	box.Exec()
}

func main() {
	demo.SetupHighDPI()
	app := demo.NewApp()

	// install the fluent translator (the Python port installs FluentTranslator).
	translator := common.NewFluentTranslator(nil, app.QObject)
	qt.QCoreApplication_InstallTranslator(translator.QTranslator)

	w := newWindow()
	w.Show()
	demo.Exec()
}
