// Command navigation_user_card ports PyQt-Fluent-Widgets' examples/navigation/
// navigation_user_card demo: a FluentWindow whose navigation panel shows an
// avatar user card above the item list.
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
	"github.com/famei/gofluent/resources"
	gfwindow "github.com/famei/gofluent/window"
)

//go:embed resource/shoko.png
var cardFS embed.FS

// Widget is the placeholder sub-interface used by the window demos.
type Widget struct {
	*qt.QFrame
	label      *widgets.SubtitleLabel
	hBoxLayout *qt.QHBoxLayout
}

func newWidget(text string, parent *qt.QWidget) *Widget {
	w := &Widget{QFrame: qt.NewQFrame(parent)}
	w.label = widgets.NewSubtitleLabelText(text, w.QWidget)
	w.hBoxLayout = qt.NewQHBoxLayout(w.QWidget)

	common.SetFont(w.label.QWidget, 24, int(qt.QFont__Normal))
	w.label.SetAlignment(qt.AlignCenter)
	w.hBoxLayout.AddWidget3(w.label.QWidget, 1, qt.AlignCenter)
	w.SetObjectName(strings.ReplaceAll(text, " ", "-"))
	return w
}

// Window is the FluentWindow demo with a navigation user card.
type Window struct {
	*gfwindow.FluentWindow

	homeInterface    *Widget
	musicInterface   *Widget
	videoInterface   *Widget
	settingInterface *Widget

	userCard *navigation.NavigationUserCard
}

func newWindow() *Window {
	w := &Window{FluentWindow: gfwindow.NewFluentWindow(nil)}

	w.homeInterface = newWidget("Home Interface", w.QWidget)
	w.musicInterface = newWidget("Music Interface", w.QWidget)
	w.videoInterface = newWidget("Video Interface", w.QWidget)
	w.settingInterface = newWidget("Setting Interface", w.QWidget)

	w.initNavigation()
	w.initWindow()
	return w
}

func (w *Window) initNavigation() {
	// add user card with custom parameters.
	w.userCard = w.NavigationInterface().AddUserCard(
		"userCard",
		shokoAvatar(),
		"zhiyiYo",
		"shokokawaii@outlook.com",
		func(bool) { w.showMessageBox() },
		navigation.NavigationItemPositionTop,
		false, // place below the expand/collapse button
	)

	w.AddSubInterface(w.homeInterface.QWidget, common.Home, "Home", navigation.NavigationItemPositionTop, nil, false)
	w.AddSubInterface(w.musicInterface.QWidget, common.Audio, "Music library", navigation.NavigationItemPositionTop, nil, false)

	w.NavigationInterface().AddSeparator(navigation.NavigationItemPositionScroll)

	w.AddSubInterface(w.videoInterface.QWidget, common.Video, "Video library", navigation.NavigationItemPositionScroll, nil, false)
	w.AddSubInterface(w.settingInterface.QWidget, common.Settings, "Settings", navigation.NavigationItemPositionBottom, nil, false)

	w.NavigationInterface().SetUpdateIndicatorPosOnCollapseFinished(true)
}

func (w *Window) initWindow() {
	w.Resize(900, 700)
	w.SetWindowIcon(asset.QIcon(resources.Images, "images/logo.png"))
	w.SetWindowTitle("Navigation User Card")

	desktop := qt.QApplication_Desktop().AvailableGeometry2()
	wd, ht := desktop.Width(), desktop.Height()
	w.Move(wd/2-w.Width()/2, ht/2-w.Height()/2)
}

func (w *Window) showMessageBox() {
	box := dialog_box.NewMessageBox(
		"User Card",
		"This is a navigation user card that displays avatar, title and subtitle.\n\n"+
			"Placement:\n"+
			"• aboveMenuButton=True: Place above expand/collapse button\n"+
			"• aboveMenuButton=False: Place below menu button (default)",
		w.QWidget,
	)
	box.Exec()
}

func shokoAvatar() *qt.QPixmap {
	return asset.QPixmap(cardFS, "resource/shoko.png")
}

func main() {
	demo.Run(func() *qt.QWidget { return newWindow().QWidget })
}
