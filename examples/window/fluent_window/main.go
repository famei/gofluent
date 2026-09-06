// Command fluent_window ports PyQt-Fluent-Widgets' examples/window/fluent_window
// demo: a FluentWindow with a left navigation interface, a nested "Albums"
// route tree, a bottom avatar entry and a badge on the video item.
package main

import (
	"embed"
	_ "embed"
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

//go:embed resource/shoko.png
var windowFS embed.FS

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

// Window is the FluentWindow demo window.
type Window struct {
	*gfwindow.FluentWindow

	homeInterface     *Widget
	musicInterface    *Widget
	videoInterface    *Widget
	folderInterface   *Widget
	settingInterface  *Widget
	albumInterface    *Widget
	albumInterface1   *Widget
	albumInterface2   *Widget
	albumInterface1_1 *Widget
}

func newWindow() *Window {
	w := &Window{FluentWindow: gfwindow.NewFluentWindow(nil)}

	w.homeInterface = newWidget("Search Interface", w.QWidget)
	w.musicInterface = newWidget("Music Interface", w.QWidget)
	w.videoInterface = newWidget("Video Interface", w.QWidget)
	w.folderInterface = newWidget("Folder Interface", w.QWidget)
	w.settingInterface = newWidget("Setting Interface", w.QWidget)
	w.albumInterface = newWidget("Album Interface", w.QWidget)
	w.albumInterface1 = newWidget("Album Interface 1", w.QWidget)
	w.albumInterface2 = newWidget("Album Interface 2", w.QWidget)
	w.albumInterface1_1 = newWidget("Album Interface 1-1", w.QWidget)

	w.initNavigation()
	w.initWindow()
	return w
}

func (w *Window) initNavigation() {
	w.AddSubInterface(w.homeInterface.QWidget, common.Home, "Home", navigation.NavigationItemPositionTop, nil, false)
	w.AddSubInterface(w.musicInterface.QWidget, common.Music, "Music library", navigation.NavigationItemPositionTop, nil, false)
	w.AddSubInterface(w.videoInterface.QWidget, common.Video, "Video library", navigation.NavigationItemPositionTop, nil, false)

	w.NavigationInterface().AddSeparator(navigation.NavigationItemPositionScroll)

	w.AddSubInterface(w.albumInterface.QWidget, common.Album, "Albums", navigation.NavigationItemPositionScroll, nil, false)
	w.AddSubInterface(w.albumInterface1.QWidget, common.Album, "Album 1", navigation.NavigationItemPositionScroll, w.albumInterface.QWidget, false)
	w.AddSubInterface(w.albumInterface1_1.QWidget, common.Album, "Album 1.1", navigation.NavigationItemPositionScroll, w.albumInterface1.QWidget, false)
	w.AddSubInterface(w.albumInterface2.QWidget, common.Album, "Album 2", navigation.NavigationItemPositionScroll, w.albumInterface.QWidget, false)
	w.AddSubInterface(w.folderInterface.QWidget, common.Folder, "Folder library", navigation.NavigationItemPositionScroll, nil, false)

	// add custom avatar widget to the bottom
	avatar := navigation.NewNavigationAvatarWidget("zhiyiYo", shokoPixmap(), nil)
	w.NavigationInterface().AddWidget(
		"avatar",
		avatar,
		func(bool) { w.showMessageBox() },
		navigation.NavigationItemPositionBottom,
		"",
		"",
	)

	w.AddSubInterface(w.settingInterface.QWidget, common.Setting, "Settings", navigation.NavigationItemPositionBottom, nil, false)

	// add badge to the video navigation item
	if item := w.NavigationInterface().Widget(w.videoInterface.ObjectName()); item != nil {
		if tw, ok := item.(*navigation.NavigationTreeWidget); ok {
			widgets.InfoBadgeMake(9, tw.ParentWidget(), widgets.InfoLevelAttention, tw.QWidget, widgets.InfoBadgePositionNavigationItem)
		}
	}
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
	// NOTE: the gofluent MessageBox does not expose the yes/cancel buttons for
	// re-labelling, so the default "OK"/"Cancel" texts are kept here (the
	// Python port renames them to "来啦老弟"/"下次一定").
	box.OnYes = func() {
		qt.QDesktopServices_OpenUrl(qt.NewQUrl3("https://afdian.net/a/zhiyiYo"))
	}
	box.Exec()
}

func shokoPixmap() *qt.QPixmap {
	return asset.QPixmap(windowFS, "resource/shoko.png")
}

func main() {
	demo.Run(func() *qt.QWidget { return newWindow().QWidget })
}
