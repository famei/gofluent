// Command navigation1 ports PyQt-Fluent-Widgets' examples/navigation/
// navigation1 demo: a frameless window with a standard title bar and a
// NavigationInterface that mixes top items, a scrollable tree menu and a bottom
// avatar widget.
package main

import (
	"embed"
	"strings"

	qt "github.com/mappu/miqt/qt"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/dialog_box"
	"github.com/famei/gofluent/components/navigation"
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/asset"
	"github.com/famei/gofluent/examples/internal/demo"
	gfwindow "github.com/famei/gofluent/window"
)

//go:embed resource/dark/*.qss resource/light/*.qss resource/shoko.png resource/logo.png
var resFS embed.FS

// Window is the frameless demo window.
type Window struct {
	*widgets.FramelessWindow

	hBoxLayout          *qt.QHBoxLayout
	navigationInterface *navigation.NavigationInterface
	stackWidget         *qt.QStackedWidget
	titleBar            *gfwindow.TitleBar
}

func newWindow() *Window {
	w := &Window{FramelessWindow: widgets.NewFramelessWindow(nil)}

	// NOTE: gofluent has no StandardTitleBar; the closest equivalent is
	// window.TitleBar (minimize/maximize/close buttons). Its height is fixed
	// explicitly because the Go FramelessWindow does not auto-size the title bar
	// the way qframelesswindow does.
	w.titleBar = gfwindow.NewTitleBar(w.QWidget)
	w.titleBar.SetFixedHeight(46)
	w.SetTitleBar(w.titleBar.QWidget)

	w.hBoxLayout = qt.NewQHBoxLayout(w.QWidget)
	w.navigationInterface = navigation.NewNavigationInterface(w.QWidget, true, false, true)
	w.stackWidget = qt.NewQStackedWidget(w.QWidget)

	searchInterface := newSubWidget("Search Interface", w.QWidget)
	musicInterface := newSubWidget("Music Interface", w.QWidget)
	videoInterface := newSubWidget("Video Interface", w.QWidget)
	folderInterface := newSubWidget("Folder Interface", w.QWidget)
	settingInterface := newSubWidget("Setting Interface", w.QWidget)
	albumInterface := newSubWidget("Album Interface", w.QWidget)
	albumInterface1 := newSubWidget("Album Interface 1", w.QWidget)
	albumInterface2 := newSubWidget("Album Interface 2", w.QWidget)
	albumInterface11 := newSubWidget("Album Interface 1-1", w.QWidget)

	w.initLayout()

	w.addSubInterface(searchInterface, common.Search, "Search", navigation.NavigationItemPositionTop, nil)
	w.addSubInterface(musicInterface, common.Audio, "Music library", navigation.NavigationItemPositionTop, nil)
	w.addSubInterface(videoInterface, common.Video, "Video library", navigation.NavigationItemPositionTop, nil)

	w.navigationInterface.AddSeparator(navigation.NavigationItemPositionTop)

	w.addSubInterface(albumInterface, common.MusicAlbum, "Albums", navigation.NavigationItemPositionScroll, nil)
	w.addSubInterface(albumInterface1, common.MusicAlbum, "Album 1", navigation.NavigationItemPositionTop, albumInterface)
	w.addSubInterface(albumInterface11, common.MusicAlbum, "Album 1.1", navigation.NavigationItemPositionTop, albumInterface1)
	w.addSubInterface(albumInterface2, common.MusicAlbum, "Album 2", navigation.NavigationItemPositionTop, albumInterface)

	// Enable expand-state memory for the tree nodes.
	if tw, ok := w.navigationInterface.Widget("Album-Interface").(*navigation.NavigationTreeWidget); ok {
		tw.SetRememberExpandState(true)
	}
	if tw, ok := w.navigationInterface.Widget("Album-Interface-1").(*navigation.NavigationTreeWidget); ok {
		tw.SetRememberExpandState(true)
	}

	w.addSubInterface(folderInterface, common.Folder, "Folder library", navigation.NavigationItemPositionScroll, nil)

	// Custom avatar widget at the bottom (shoko.png embedded via go:embed).
	avatar := navigation.NewNavigationAvatarWidget("zhiyiYo", asset.QImage(resFS, "resource/shoko.png"), w.QWidget)
	w.navigationInterface.AddWidget("avatar", avatar, func(bool) { w.showMessageBox() }, navigation.NavigationItemPositionBottom, "", "")

	w.addSubInterface(settingInterface, common.Settings, "Settings", navigation.NavigationItemPositionBottom, nil)

	w.stackWidget.OnCurrentChanged(w.onCurrentInterfaceChanged)
	w.stackWidget.SetCurrentIndex(1)

	w.initWindow()
	return w
}

// newSubWidget builds the QFrame page that shows a centered label. The object
// name mirrors the Python text.replace(' ', '-').
func newSubWidget(text string, parent *qt.QWidget) *qt.QFrame {
	f := qt.NewQFrame(parent)
	f.SetObjectName(strings.ReplaceAll(text, " ", "-"))
	label := qt.NewQLabel5(text, f.QWidget)
	label.SetAlignment(qt.AlignCenter)
	layout := qt.NewQHBoxLayout(f.QWidget)
	layout.AddWidget3(label.QWidget, 1, qt.AlignCenter)
	return f
}

func (w *Window) initLayout() {
	w.hBoxLayout.SetSpacing(0)
	w.hBoxLayout.SetContentsMargins(0, w.titleBar.Height(), 0, 0)
	w.hBoxLayout.AddWidget(w.navigationInterface.QWidget)
	w.hBoxLayout.AddWidget2(w.stackWidget.QWidget, 1)

	w.positionTitleBar()
	w.OnResizeEvent(func(super func(e *qt.QResizeEvent), e *qt.QResizeEvent) {
		super(e)
		w.positionTitleBar()
	})
}

func (w *Window) positionTitleBar() {
	w.titleBar.Move(0, 0)
	w.titleBar.Resize(w.Width(), w.titleBar.Height())
}

func (w *Window) addSubInterface(iface *qt.QFrame, icon interface{}, text string, position navigation.NavigationItemPosition, parent *qt.QFrame) {
	w.stackWidget.AddWidget(iface.QWidget)
	parentRouteKey := ""
	if parent != nil {
		parentRouteKey = parent.ObjectName()
	}
	w.navigationInterface.AddItem(iface.ObjectName(), icon, text, func(bool) { w.switchTo(iface.QWidget) }, true, position, text, parentRouteKey)
}

func (w *Window) initWindow() {
	w.Resize(900, 700)
	icon := asset.QIcon(resFS, "resource/logo.png")
	w.SetWindowIcon(icon)
	icon.Delete()
	w.SetWindowTitle("PyQt-Fluent-Widgets")
	w.titleBar.SetAttribute(qt.WA_StyledBackground)

	desktop := qt.QApplication_Desktop().AvailableGeometry2()
	wd, ht := desktop.Width(), desktop.Height()
	w.Move(wd/2-w.Width()/2, ht/2-w.Height()/2)

	w.setQss()
}

func (w *Window) setQss() {
	dir := "resource/light/"
	if common.IsDarkTheme() {
		dir = "resource/dark/"
	}
	w.SetStyleSheet(asset.String(resFS, dir+"demo.qss"))
}

func (w *Window) switchTo(widget *qt.QWidget) {
	w.stackWidget.SetCurrentWidget(widget)
}

func (w *Window) onCurrentInterfaceChanged(index int) {
	widget := w.stackWidget.Widget(index)
	if widget == nil {
		return
	}
	w.navigationInterface.SetCurrentItem(widget.ObjectName())
}

func (w *Window) showMessageBox() {
	box := dialog_box.NewMessageBox(
		"支持作者🥰",
		"个人开发不易，如果这个项目帮助到了您，可以考虑请作者喝一瓶快乐水🥤。您的支持就是作者开发和维护项目的动力🚀",
		w.QWidget,
	)
	// NOTE: gofluent's MessageBox keeps yesButton/cancelButton unexported, so the
	// Python w.yesButton.setText('来啦老弟') / w.cancelButton.setText('下次一定')
	// customization is skipped. The URL opens on the yes (OK) button instead.
	box.OnYes = func() {
		url := qt.NewQUrl3("https://afdian.net/a/zhiyiYo")
		qt.QDesktopServices_OpenUrl(url)
		url.Delete()
	}
	box.Exec()
}

func main() {
	demo.Run(func() *qt.QWidget { return newWindow().QWidget })
}
