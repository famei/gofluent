// Command ms_fluent_window ports PyQt-Fluent-Widgets' examples/window/
// ms_fluent_window demo: a Microsoft Store style window with a bottom
// navigation bar.
package main

import (
	"strings"

	qt "github.com/mappu/miqt/qt"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/dialog_box"
	"github.com/famei/gofluent/components/navigation"
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/asset"
	"github.com/famei/gofluent/examples/internal/demo"
	"github.com/famei/gofluent/resources"
	gfwindow "github.com/famei/gofluent/window"
)

// Widget is the placeholder sub-interface used by the window demos.
type Widget struct {
	*qt.QWidget
	label      *widgets.SubtitleLabel
	hBoxLayout *qt.QHBoxLayout
}

func newWidget(text string, parent *qt.QWidget) *Widget {
	w := &Widget{QWidget: qt.NewQWidget(parent)}
	w.label = widgets.NewSubtitleLabelText(text, w.QWidget)
	w.hBoxLayout = qt.NewQHBoxLayout(w.QWidget)

	common.SetFont(w.label.QWidget, 24, int(qt.QFont__Normal))
	w.label.SetAlignment(qt.AlignCenter)
	w.hBoxLayout.AddWidget3(w.label.QWidget, 1, qt.AlignCenter)
	w.SetObjectName(strings.ReplaceAll(text, " ", "-"))
	return w
}

// Window is the MSFluentWindow demo window.
type Window struct {
	*gfwindow.MSFluentWindow

	homeInterface    *Widget
	appInterface     *Widget
	videoInterface   *Widget
	libraryInterface *Widget
}

func newWindow() *Window {
	w := &Window{MSFluentWindow: gfwindow.NewMSFluentWindow(nil)}

	w.homeInterface = newWidget("Home Interface", w.QWidget)
	w.appInterface = newWidget("Application Interface", w.QWidget)
	w.videoInterface = newWidget("Video Interface", w.QWidget)
	w.libraryInterface = newWidget("library Interface", w.QWidget)

	w.initNavigation()
	w.initWindow()
	return w
}

func (w *Window) initNavigation() {
	w.AddSubInterface(w.homeInterface.QWidget, common.Home, "主页", common.HomeFill, navigation.NavigationItemPositionTop, false)
	w.AddSubInterface(w.appInterface.QWidget, common.Application, "应用", nil, navigation.NavigationItemPositionTop, false)
	w.AddSubInterface(w.videoInterface.QWidget, common.Video, "视频", nil, navigation.NavigationItemPositionTop, false)

	w.AddSubInterface(w.libraryInterface.QWidget, common.Library, "库", common.LibraryFill, navigation.NavigationItemPositionBottom, false)
	w.NavigationBar().AddItem(
		"Help",
		common.Help,
		"帮助",
		func(bool) { w.showMessageBox() },
		false,
		nil,
		navigation.NavigationItemPositionBottom,
	)

	w.NavigationBar().SetCurrentItem(w.homeInterface.ObjectName())
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
	demo.Run(func() *qt.QWidget { return newWindow().QWidget })
}
