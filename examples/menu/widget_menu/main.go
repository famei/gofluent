// Command widget_menu migrates examples/menu/widget_menu/demo.py: a context menu
// that embeds a custom ProfileCard widget (avatar, name, email, logout link)
// above the standard menu actions.
package main

import (
	_ "embed"

	qt "github.com/mappu/miqt/qt"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
)

//go:embed resource/shoko.png
var shokoPNG []byte

// shokoImage decodes the avatar image once. AvatarWidget stores the pointer
// without cloning, so it must stay reachable for the whole demo lifetime.
var shokoImage *qt.QImage

func shokoQImage() *qt.QImage {
	if shokoImage == nil {
		img := qt.QImage_FromDataWithData(shokoPNG)
		if img == nil || img.IsNull() {
			panic("widget_menu: decode resource/shoko.png")
		}
		shokoImage = img
	}
	return shokoImage
}

// ProfileCard is the custom widget shown at the top of the context menu.
type ProfileCard struct {
	*qt.QWidget
	avatar       *widgets.AvatarWidget
	nameLabel    *widgets.BodyLabel
	emailLabel   *widgets.CaptionLabel
	logoutButton *widgets.HyperlinkButton
}

func newProfileCard(avatarImage *qt.QImage, name, email string, parent *qt.QWidget) *ProfileCard {
	w := &ProfileCard{QWidget: qt.NewQWidget(parent)}
	w.avatar = widgets.NewAvatarWidget(w.QWidget)
	w.avatar.SetImage(avatarImage)
	w.nameLabel = widgets.NewBodyLabelText(name, w.QWidget)
	w.emailLabel = widgets.NewCaptionLabelText(email, w.QWidget)
	w.logoutButton = widgets.NewHyperlinkButtonURL("https://qfluentwidgets.com", "注销", w.QWidget)

	// email label color (dark: 206,206,206; light: 96,96,96)
	var emailColor *qt.QColor
	if common.IsDarkTheme() {
		emailColor = qt.NewQColor3(206, 206, 206)
	} else {
		emailColor = qt.NewQColor3(96, 96, 96)
	}
	w.emailLabel.SetStyleSheet("QLabel{color: " + emailColor.Name() + "}")
	emailColor.Delete()

	// name label color (dark: white; light: black)
	var nameColor *qt.QColor
	if common.IsDarkTheme() {
		nameColor = qt.NewQColor3(255, 255, 255)
	} else {
		nameColor = qt.NewQColor3(0, 0, 0)
	}
	w.nameLabel.SetStyleSheet("QLabel{color: " + nameColor.Name() + "}")
	nameColor.Delete()

	common.SetFont(w.logoutButton.QWidget, 13, int(qt.QFont__Normal))

	w.SetFixedSize2(307, 82)
	w.avatar.SetRadius(24)
	w.avatar.Move(2, 6)
	w.nameLabel.Move(64, 13)
	w.emailLabel.Move(64, 32)
	w.logoutButton.Move(52, 48)
	return w
}

// Demo is the right-clickable window.
type Demo struct {
	*qt.QWidget
	label *widgets.BodyLabel
}

func newDemo() *Demo {
	w := &Demo{QWidget: qt.NewQWidget2()}
	// The Python demo uses the QSS type selector "Demo{...}"; miqt has no
	// dynamic metaobject for the Go subclass, so the id selector is used with
	// SetObjectName.
	w.SetObjectName("Demo")
	w.SetStyleSheet("#Demo{background: white}")
	hBox := qt.NewQHBoxLayout(w.QWidget)

	w.label = widgets.NewBodyLabelText("Right-click your mouse", w.QWidget)
	w.label.SetAlignment(qt.AlignCenter)
	common.SetFont(w.label.QWidget, 18, int(qt.QFont__Normal))

	hBox.AddWidget(w.label.QWidget)
	w.Resize(400, 400)

	w.OnContextMenuEvent(func(super func(e *qt.QContextMenuEvent), e *qt.QContextMenuEvent) {
		w.showContextMenu(e)
	})
	return w
}

func (w *Demo) showContextMenu(e *qt.QContextMenuEvent) {
	menu := widgets.NewRoundMenu("", w.QWidget)

	// add custom widget
	card := newProfileCard(shokoQImage(), "硝子酱", "shokokawaii@outlook.com", menu.QWidget)
	menu.AddWidget(card.QWidget)

	menu.AddSeparator()
	menu.AddAction(common.NewActionFluentIcon(common.People, "管理账户和设置", nil).QAction)
	menu.AddAction(common.NewActionFluentIcon(common.ShoppingCart, "支付方式", nil).QAction)
	menu.AddAction(common.NewActionFluentIcon(common.Code, "兑换代码和礼品卡", nil).QAction)
	menu.AddSeparator()
	menu.AddAction(common.NewActionFluentIcon(common.Setting, "设置", nil).QAction)

	pos := e.GlobalPos()
	menu.Exec(pos, widgets.MenuAnimationDropDown)

}

func main() {
	demo.Run(func() *qt.QWidget { return newDemo().QWidget })
}
