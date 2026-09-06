// Command navigation3 ports PyQt-Fluent-Widgets' examples/navigation/
// navigation3 demo: a custom top navigation bar (menu button + title) that pops
// out a NavigationPanel when the menu button is clicked.
package main

import (
	"embed"
	"strings"

	qt "github.com/mappu/miqt/qt"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/navigation"
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/asset"
	"github.com/famei/gofluent/examples/internal/demo"
	gfwindow "github.com/famei/gofluent/window"
)

//go:embed resource/dark/*.qss resource/light/*.qss resource/logo.png
var resFS embed.FS

// NavigationBar is the custom top bar (menu button + title) that hosts a
// pop-out NavigationPanel.
type NavigationBar struct {
	*qt.QWidget

	hBoxLayout      *qt.QHBoxLayout
	menuButton      *navigation.NavigationToolButton
	navigationPanel *navigation.NavigationPanel
	titleLabel      *qt.QLabel
}

func newNavigationBar(parent *qt.QWidget) *NavigationBar {
	b := &NavigationBar{QWidget: qt.NewQWidget(parent)}
	b.hBoxLayout = qt.NewQHBoxLayout(b.QWidget)
	b.menuButton = navigation.NewNavigationToolButton(common.Menu, b.QWidget)
	// The Python reference parents the pop-out panel to the window (parent),
	// not to the NavigationBar, so `navigationPanel.move(0, 31)` positions it
	// relative to the window's top-left. Parent it the same way here, otherwise
	// the panel is offset by the bar's own position (the reported wrong
	// position) and it would be re-parented again during expand().
	b.navigationPanel = navigation.NewNavigationPanel(parent, true)
	b.titleLabel = qt.NewQLabel(b.QWidget)

	b.navigationPanel.Move(0, 31)
	b.hBoxLayout.SetContentsMargins(5, 5, 5, 5)
	b.hBoxLayout.AddWidget(b.menuButton.QWidget)
	b.hBoxLayout.AddWidget(b.titleLabel.QWidget)

	b.menuButton.OnClicked(func(bool) { b.showNavigationPanel() })
	b.navigationPanel.SetExpandWidth(260)
	b.navigationPanel.SetMenuButtonVisible(true)
	b.navigationPanel.Hide()
	return b
}

func (b *NavigationBar) setTitle(title string) {
	b.titleLabel.SetText(title)
	b.titleLabel.AdjustSize()
}

func (b *NavigationBar) showNavigationPanel() {
	b.navigationPanel.Show()
	b.navigationPanel.Raise()
	b.navigationPanel.Expand(true)
}

func (b *NavigationBar) addItem(routeKey string, icon interface{}, text string, onClick func(bool), selectable bool, position navigation.NavigationItemPosition) {
	wrapper := func(v bool) {
		onClick(v)
		b.setTitle(text)
	}
	b.navigationPanel.AddItem(routeKey, icon, text, wrapper, selectable, position, "", "")
}

func (b *NavigationBar) addSeparator(position navigation.NavigationItemPosition) {
	b.navigationPanel.AddSeparator(position)
}

func (b *NavigationBar) setCurrentItem(routeKey string) {
	b.navigationPanel.SetCurrentItem(routeKey)
	if tw, ok := b.navigationPanel.Widget(routeKey).(*navigation.NavigationTreeWidget); ok {
		b.setTitle(tw.Text())
	}
}

// resizePanel keeps the panel height in sync with the window (the Go equivalent
// of the Python eventFilter on QEvent.Resize).
func (b *NavigationBar) resizePanel(height int) {
	if height < 0 {
		height = 0
	}
	b.navigationPanel.SetFixedHeight(height)
}

// Window is the frameless demo window.
type Window struct {
	*widgets.FramelessWindow

	vBoxLayout          *qt.QVBoxLayout
	navigationInterface *NavigationBar
	stackWidget         *qt.QStackedWidget
	titleBar            *gfwindow.TitleBar
}

func newWindow() *Window {
	w := &Window{FramelessWindow: widgets.NewFramelessWindow(nil)}

	// NOTE: gofluent has no StandardTitleBar; use window.TitleBar instead.
	// qframelesswindow's StandardTitleBar is 32px tall (TitleBarBase sets
	// setFixedHeight(32)); the panel below is moved to y=31 relative to that
	// height, so an oversized title bar pushes the panel up over the bar.
	w.titleBar = gfwindow.NewTitleBar(w.QWidget)
	w.titleBar.SetFixedHeight(32)
	w.SetTitleBar(w.titleBar.QWidget)

	w.vBoxLayout = qt.NewQVBoxLayout(w.QWidget)
	w.navigationInterface = newNavigationBar(w.QWidget)
	w.stackWidget = qt.NewQStackedWidget(w.QWidget)

	searchInterface := newSubWidget("Search Interface", w.QWidget)
	musicInterface := newSubWidget("Music Interface", w.QWidget)
	videoInterface := newSubWidget("Video Interface", w.QWidget)
	folderInterface := newSubWidget("Folder Interface", w.QWidget)
	settingInterface := newSubWidget("Setting Interface", w.QWidget)

	// The Python demo adds every page twice (here and again inside
	// addSubInterface); that is preserved for fidelity.
	w.stackWidget.AddWidget(searchInterface)
	w.stackWidget.AddWidget(musicInterface)
	w.stackWidget.AddWidget(videoInterface)
	w.stackWidget.AddWidget(folderInterface)
	w.stackWidget.AddWidget(settingInterface)

	w.initLayout()

	w.addSubInterface(searchInterface, common.Search, "Search", navigation.NavigationItemPositionTop)
	w.addSubInterface(musicInterface, common.Music, "Music library", navigation.NavigationItemPositionTop)
	w.addSubInterface(videoInterface, common.Video, "Video library", navigation.NavigationItemPositionTop)

	w.navigationInterface.addSeparator(navigation.NavigationItemPositionTop)

	w.addSubInterface(folderInterface, common.Folder, "Folder library", navigation.NavigationItemPositionScroll)
	w.addSubInterface(settingInterface, common.Setting, "Settings", navigation.NavigationItemPositionBottom)

	w.stackWidget.OnCurrentChanged(w.onCurrentInterfaceChanged)
	w.stackWidget.SetCurrentIndex(1)

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
	w.vBoxLayout.SetSpacing(0)
	w.vBoxLayout.SetContentsMargins(0, w.titleBar.Height(), 0, 0)
	w.vBoxLayout.AddWidget(w.navigationInterface.QWidget)
	w.vBoxLayout.AddWidget2(w.stackWidget.QWidget, 1)

	w.positionTitleBar()
	w.OnResizeEvent(func(super func(e *qt.QResizeEvent), e *qt.QResizeEvent) {
		super(e)
		w.positionTitleBar()
		size := e.Size() // borrowed reference into the event — do NOT Delete
		w.navigationInterface.resizePanel(size.Height() - 31)
	})
}

func (w *Window) positionTitleBar() {
	w.titleBar.Move(0, 0)
	w.titleBar.Resize(w.Width(), w.titleBar.Height())
}

func (w *Window) addSubInterface(iface *qt.QWidget, icon interface{}, text string, position navigation.NavigationItemPosition) {
	w.stackWidget.AddWidget(iface)
	w.navigationInterface.addItem(iface.ObjectName(), icon, text, func(bool) { w.switchTo(iface) }, true, position)
}

func (w *Window) initWindow() {
	w.Resize(500, 600)
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
	w.navigationInterface.setCurrentItem(widget.ObjectName())
}

func main() {
	demo.Run(func() *qt.QWidget { return newWindow().QWidget })
}
