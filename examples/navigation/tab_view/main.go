// Command tab_view ports PyQt-Fluent-Widgets' examples/navigation/tab_view demo:
// an MSFluentWindow whose home interface hosts a TabBar above a QStackedWidget
// of icon tabs.
package main

import (
	"embed"
	"fmt"
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

//go:embed resource/Heart.png resource/Smiling_with_heart.png
var tabIcons embed.FS

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

// TabInterface is a single tab page with an icon and a subtitle.
type TabInterface struct {
	*qt.QWidget

	iconWidget *widgets.IconWidget
	label      *widgets.SubtitleLabel
	vBoxLayout *qt.QVBoxLayout
}

func newTabInterface(text string, icon interface{}, objectName string, parent *qt.QWidget) *TabInterface {
	w := &TabInterface{QWidget: qt.NewQWidget(parent)}
	w.iconWidget = widgets.NewIconWidgetIcon(icon, w.QWidget)
	w.label = widgets.NewSubtitleLabelText(text, w.QWidget)
	w.iconWidget.SetFixedSize2(120, 120)

	w.vBoxLayout = qt.NewQVBoxLayout(w.QWidget)
	w.vBoxLayout.SetSpacing(30)
	w.vBoxLayout.AddWidget3(w.iconWidget.QWidget, 0, qt.AlignCenter)
	w.vBoxLayout.AddWidget3(w.label.QWidget, 0, qt.AlignCenter)
	common.SetFont(w.label.QWidget, 24, int(qt.QFont__Normal))

	w.SetObjectName(objectName)
	return w
}

// Window is the tab view demo window (MSFluentWindow).
//
// NOTE: the Python port embeds the TabBar inside a CustomTitleBar(MSFluentTitleBar)
// subclass. Because MSFluentTitleBar's internal hBoxLayout is private to the
// window package, this port places the TabBar at the top of the home interface
// container instead of inside the title bar.
type Window struct {
	*gfwindow.MSFluentWindow

	tabBar        *widgets.TabBar
	homeStack     *qt.QStackedWidget
	homeContainer *qt.QWidget

	appInterface     *Widget
	videoInterface   *Widget
	libraryInterface *Widget

	tabCount int
	// tabPages replaces the Python findChild(TabInterface, objectName) lookup.
	tabPages map[string]*TabInterface
}

func newWindow() *Window {
	w := &Window{MSFluentWindow: gfwindow.NewMSFluentWindow(nil)}
	w.tabCount = 1
	w.tabPages = map[string]*TabInterface{}

	// The home sub-interface hosts only the stacked tab pages; the tab bar and
	// tool buttons live in the title bar (mirrors the Python CustomTitleBar).
	w.homeContainer = qt.NewQWidget(nil)
	w.homeContainer.SetObjectName("homeInterface")
	homeLayout := qt.NewQVBoxLayout(w.homeContainer)
	homeLayout.SetContentsMargins(0, 0, 0, 0)
	homeLayout.SetSpacing(0)

	w.homeStack = qt.NewQStackedWidget(w.homeContainer)
	homeLayout.AddWidget2(w.homeStack.QWidget, 1)

	// Build the title bar tool buttons (search / back / forward) and the tab
	// bar, then insert them into the MSFluentTitleBar hBoxLayout between the
	// title and the system buttons, exactly like the Python demo's
	// CustomTitleBar.insertLayout(4)/insertWidget(5).
	titleBar := w.MSFTitleBar()
	searchButton := widgets.NewTransparentToolButtonIcon(common.SearchMirror, titleBar.QWidget)
	forwardButton := widgets.NewTransparentToolButtonIcon(common.RightArrow, titleBar.QWidget)
	backButton := widgets.NewTransparentToolButtonIcon(common.LeftArrow, titleBar.QWidget)
	forwardButton.SetDisabled(true)

	toolButtonLayout := qt.NewQHBoxLayout2()
	toolButtonLayout.SetContentsMargins(20, 0, 20, 0)
	toolButtonLayout.SetSpacing(15)
	toolButtonLayout.AddWidget(searchButton.QWidget)
	toolButtonLayout.AddWidget(backButton.QWidget)
	toolButtonLayout.AddWidget(forwardButton.QWidget)

	w.tabBar = widgets.NewTabBar(titleBar.QWidget)
	w.tabBar.SetMovable(true)
	w.tabBar.SetTabMaximumWidth(220)
	w.tabBar.SetTabShadowEnabled(false)
	w.tabBar.SetTabSelectedBackgroundColor(qt.NewQColor11(255, 255, 255, 125), qt.NewQColor11(255, 255, 255, 50))

	// hBoxLayout order: [0] spacing(20), [1] icon, [2] spacing(2), [3] title,
	// [4] stretch, [5] button vBoxLayout. Insert the tool buttons before the
	// stretch and the tab bar after them, then zero the old stretch so the tab
	// bar absorbs the free space (the system buttons stay pinned right).
	hbox := titleBar.HBoxLayout()
	hbox.InsertLayout2(4, toolButtonLayout.QLayout, 0)
	hbox.InsertWidget3(5, w.tabBar.QWidget, 1, qt.AlignVCenter)
	hbox.SetStretch(6, 0)

	// The tool buttons are interactive; exclude them from the title-bar drag
	// area so they receive clicks. (The tab items are excluded separately in
	// TabBar.InsertTab, while the tab bar's empty area stays draggable.)
	for _, btn := range []*qt.QWidget{searchButton.QWidget, backButton.QWidget, forwardButton.QWidget} {
		btn.SetProperty("isTitleBarButton", qt.NewQVariant11(true))
	}

	w.appInterface = newWidget("Application Interface", w.QWidget)
	w.videoInterface = newWidget("Video Interface", w.QWidget)
	w.libraryInterface = newWidget("library Interface", w.QWidget)

	w.initNavigation()
	w.initWindow()
	return w
}

func (w *Window) initNavigation() {
	w.AddSubInterface(w.homeContainer, common.Home, "主页", common.HomeFill, navigation.NavigationItemPositionTop, false)
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

	w.NavigationBar().SetCurrentItem(w.homeContainer.ObjectName())

	// add tab.
	w.addTab("Heart", "As long as you love me", heartIcon())

	w.tabBar.OnCurrentChanged(w.onTabChanged)
	w.tabBar.OnTabAddRequested(w.onTabAddRequested)
	// Mirrors the reference CustomTitleBar: closing a tab removes it.
	w.tabBar.OnTabCloseRequested(func(index int) { w.tabBar.RemoveTab(index) })
}

func (w *Window) initWindow() {
	w.Resize(1100, 750)
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

func (w *Window) onTabChanged(_ int) {
	objectName := w.tabBar.CurrentTab().RouteKey()
	if page, ok := w.tabPages[objectName]; ok {
		w.homeStack.SetCurrentWidget(page.QWidget)
	}
	w.SwitchTo(w.homeContainer)
}

func (w *Window) onTabAddRequested() {
	text := fmt.Sprintf("硝子酱一级棒卡哇伊×%d", w.tabCount)
	w.addTab(text, text, smilingIcon())
	w.tabCount++
}

func (w *Window) addTab(routeKey, text string, icon interface{}) {
	w.tabBar.AddTab(routeKey, text, icon, nil)
	page := newTabInterface(text, icon, routeKey, w.QWidget)
	w.tabPages[routeKey] = page
	w.homeStack.AddWidget(page.QWidget)
}

func heartIcon() *qt.QIcon {
	return asset.QIcon(tabIcons, "resource/Heart.png")
}

func smilingIcon() *qt.QIcon {
	return asset.QIcon(tabIcons, "resource/Smiling_with_heart.png")
}

func main() {
	demo.Run(func() *qt.QWidget { return newWindow().QWidget })
}
