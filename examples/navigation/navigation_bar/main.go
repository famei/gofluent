// Command navigation_bar ports PyQt-Fluent-Widgets' examples/navigation/
// navigation_bar demo: a frameless window with a horizontal NavigationBar and a
// PopUpAniStackedWidget (the pop-out animation is dropped, see MIGRATION_GUIDE
// §12).
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

//go:embed resource/dark/*.qss resource/light/*.qss
var resFS embed.FS

// StackedWidget wraps PopUpAniStackedWidget in a frame.
type StackedWidget struct {
	*qt.QFrame

	hBoxLayout *qt.QHBoxLayout
	view       *widgets.PopUpAniStackedWidget
}

func newStackedWidget(parent *qt.QWidget) *StackedWidget {
	s := &StackedWidget{QFrame: qt.NewQFrame(parent)}
	s.hBoxLayout = qt.NewQHBoxLayout(s.QWidget)
	s.view = widgets.NewPopUpAniStackedWidget(s.QWidget)
	s.hBoxLayout.SetContentsMargins(0, 0, 0, 0)
	s.hBoxLayout.AddWidget(s.view.QWidget)
	return s
}

func (s *StackedWidget) addWidget(widget *qt.QWidget) {
	// deltaX/deltaY are unused because the pop-out animation is dropped.
	s.view.AddWidget(widget, 0, 0)
}

func (s *StackedWidget) widget(index int) *qt.QWidget {
	return s.view.Widget(index)
}

func (s *StackedWidget) setCurrentWidget(widget *qt.QWidget) {
	s.view.SetCurrentWidget(widget)
}

func (s *StackedWidget) setCurrentIndex(index int) {
	s.view.SetCurrentIndex(index)
}

func (s *StackedWidget) onCurrentChanged(f func(int)) {
	s.view.OnCurrentChanged(f)
}

// Window is the frameless demo window.
type Window struct {
	*widgets.FramelessWindow

	hBoxLayout     *qt.QHBoxLayout
	navigationBar  *navigation.NavigationBar
	stackWidget    *StackedWidget
	titleBar       *gfwindow.MSFluentTitleBar
	searchLineEdit *widgets.SearchLineEdit
}

func newWindow() *Window {
	w := &Window{FramelessWindow: widgets.NewFramelessWindow(nil)}

	// NOTE: the Python CustomTitleBar (icon + title + search box + buttons) is
	// approximated by gfwindow.MSFluentTitleBar plus a floating SearchLineEdit;
	// the search box is positioned manually because the title-bar layout fields
	// are unexported in the Go port.
	w.titleBar = gfwindow.NewMSFluentTitleBar(w.QWidget)
	w.SetTitleBar(w.titleBar.QWidget)

	w.searchLineEdit = widgets.NewSearchLineEdit(w.titleBar.QWidget)
	w.searchLineEdit.SetPlaceholderText("搜索应用、游戏、电影、设备等")
	w.searchLineEdit.SetFixedWidth(400)
	w.searchLineEdit.SetClearButtonEnabled(true)
	w.searchLineEdit.Show()

	w.hBoxLayout = qt.NewQHBoxLayout(w.QWidget)
	w.navigationBar = navigation.NewNavigationBar(w.QWidget)
	w.stackWidget = newStackedWidget(w.QWidget)

	homeInterface := newSubWidget("Home Interface", w.QWidget)
	appInterface := newSubWidget("Application Interface", w.QWidget)
	videoInterface := newSubWidget("Video Interface", w.QWidget)
	libraryInterface := newSubWidget("library Interface", w.QWidget)

	w.initLayout()

	w.addSubInterface(homeInterface, common.Home, "主页", navigation.NavigationItemPositionTop, common.HomeFill)
	w.addSubInterface(appInterface, common.Application, "应用", navigation.NavigationItemPositionTop, nil)
	w.addSubInterface(videoInterface, common.Video, "视频", navigation.NavigationItemPositionTop, nil)
	w.addSubInterface(libraryInterface, common.BookShelf, "库", navigation.NavigationItemPositionBottom, common.LibraryFill)

	w.navigationBar.AddItem("Help", common.Help, "帮助", func(bool) { w.showMessageBox() }, false, nil, navigation.NavigationItemPositionBottom)

	w.stackWidget.onCurrentChanged(w.onCurrentInterfaceChanged)
	w.navigationBar.SetCurrentItem(homeInterface.ObjectName())

	w.initWindow()
	return w
}

// newSubWidget builds the QWidget page that shows a centered label.
func newSubWidget(text string, parent *qt.QWidget) *qt.QWidget {
	w := qt.NewQWidget(parent)
	w.SetObjectName(strings.ReplaceAll(text, " ", "-"))
	label := qt.NewQLabel5(text, w)
	label.SetAlignment(qt.AlignCenter)
	layout := qt.NewQHBoxLayout(w)
	layout.AddWidget3(label.QWidget, 1, qt.AlignCenter)
	return w
}

func (w *Window) initLayout() {
	w.hBoxLayout.SetSpacing(0)
	w.hBoxLayout.SetContentsMargins(0, 48, 0, 0)
	w.hBoxLayout.AddWidget(w.navigationBar.QWidget)
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
	w.searchLineEdit.Move((w.titleBar.Width()-w.searchLineEdit.Width())/2, 8)
}

func (w *Window) addSubInterface(iface *qt.QWidget, icon interface{}, text string, position navigation.NavigationItemPosition, selectedIcon interface{}) {
	w.stackWidget.addWidget(iface)
	w.navigationBar.AddItem(iface.ObjectName(), icon, text, func(bool) { w.switchTo(iface) }, true, selectedIcon, position)
}

func (w *Window) initWindow() {
	w.Resize(900, 700)
	// NOTE: the Python uses the Qt resource ':/qfluentwidgets/images/logo.png'
	// (not embedded here), so the window icon is skipped.
	w.SetWindowTitle("PyQt-Fluent-Widgets")
	w.titleBar.SetAttribute(qt.WA_StyledBackground)
	w.positionTitleBar()

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
	w.stackWidget.setCurrentWidget(widget)
}

func (w *Window) onCurrentInterfaceChanged(index int) {
	widget := w.stackWidget.widget(index)
	if widget == nil {
		return
	}
	w.navigationBar.SetCurrentItem(widget.ObjectName())
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
