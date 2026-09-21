package view

import (
	gcommon "github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	gallerycommon "github.com/famei/gofluent/examples/gallery/app/common"
	"github.com/famei/gofluent/examples/gallery/app/resource"
	qt "github.com/mappu/miqt/qt"
)

// MenuInterface is the "Menus & toolbars" gallery page (port of
// app/view/menu_interface.py).
type MenuInterface struct {
	*GalleryInterface

	createTimeAction   *gcommon.Action
	shootTimeAction    *gcommon.Action
	modifiedTimeAction *gcommon.Action
	nameAction         *gcommon.Action
	ascendAction       *gcommon.Action
	descendAction      *gcommon.Action

	imageLabel *widgets.ImageLabel
}

// NewMenuInterface builds the menu interface.
func NewMenuInterface(parent *qt.QWidget) *MenuInterface {
	t := gallerycommon.NewTranslator()
	i := &MenuInterface{GalleryInterface: NewGalleryInterface(t.Menus, "github.com/famei/gofluent/components/widgets", parent)}
	i.SetObjectName("menuInterface")

	tr := func(s string) string { return gallerycommon.Tr("MenuInterface", s) }

	i.createTimeAction = gcommon.NewActionFluentIcon(gcommon.Calendar, tr("Create Date"), nil)
	i.shootTimeAction = gcommon.NewActionFluentIcon(gcommon.Camera, tr("Shooting Date"), nil)
	i.modifiedTimeAction = gcommon.NewActionFluentIcon(gcommon.Edit, tr("Modified time"), nil)
	i.nameAction = gcommon.NewActionFluentIcon(gcommon.Font, tr("Name"), nil)
	for _, a := range []*gcommon.Action{i.createTimeAction, i.shootTimeAction, i.modifiedTimeAction, i.nameAction} {
		a.SetCheckable(true)
	}

	actionGroup1 := qt.NewQActionGroup(i.QObject)
	actionGroup1.AddAction(i.createTimeAction.QAction)
	actionGroup1.AddAction(i.shootTimeAction.QAction)
	actionGroup1.AddAction(i.modifiedTimeAction.QAction)
	actionGroup1.AddAction(i.nameAction.QAction)

	i.ascendAction = gcommon.NewActionFluentIcon(gcommon.Up, tr("Ascending"), nil)
	i.descendAction = gcommon.NewActionFluentIcon(gcommon.Down, tr("Descending"), nil)
	i.ascendAction.SetCheckable(true)
	i.descendAction.SetCheckable(true)
	actionGroup2 := qt.NewQActionGroup(i.QObject)
	actionGroup2.AddAction(i.ascendAction.QAction)
	actionGroup2.AddAction(i.descendAction.QAction)

	i.shootTimeAction.SetChecked(true)
	i.ascendAction.SetChecked(true)

	button1 := widgets.NewPushButtonText(tr("Show menu"), nil)
	button1.OnClicked(func() {
		pos := button1.MapToGlobal(qt.NewQPoint2(button1.Width()+5, -100))
		i.createMenu(pos)
	})
	i.AddExampleCard(tr("Rounded corners menu"), button1.QWidget,
		"components/widgets/menu.go", codeMenu, 0)

	button3 := widgets.NewPushButtonText(tr("Show menu"), nil)
	button3.OnClicked(func() {
		pos := button3.MapToGlobal(qt.NewQPoint2(button3.Width()+5, -100))
		i.createCustomWidgetMenu(pos)
	})
	i.AddExampleCard(tr("Rounded corners menu with custom widget"), button3.QWidget,
		"components/widgets/menu.go", codeWidgetMenu, 0)

	button2 := widgets.NewPushButtonText(tr("Show menu"), nil)
	button2.OnClicked(func() {
		pos := button2.MapToGlobal(qt.NewQPoint2(button2.Width()+5, -100))
		i.createCheckableMenu(pos)
	})
	i.AddExampleCard(tr("Checkable menu"), button2.QWidget,
		"components/widgets/menu.go", codeCheckableMenu, 0)

	i.AddExampleCard(tr("Command bar"), i.createCommandBar(),
		"components/widgets/command_bar.go", codeCommandBar, 1)

	widget := qt.NewQWidget(i.QWidget)
	widgetLayout := qt.NewQVBoxLayout(widget)
	widgetLayout.SetContentsMargins(0, 0, 0, 0)
	widgetLayout.SetSpacing(10)

	label := qt.NewQLabel5(tr("Click the image to open a command bar flyout 👇️🥵"), widget)
	i.imageLabel = widgets.NewImageLabelImage(resource.Pixmap("chidanta5.jpg"), widget)
	i.imageLabel.ScaledToWidth(350)
	i.imageLabel.SetBorderRadius(8, 8, 8, 8)
	i.imageLabel.OnClicked(i.createCommandBarFlyout)

	widgetLayout.AddWidget(label.QWidget)
	widgetLayout.AddWidget(i.imageLabel.QWidget)

	i.AddExampleCard(tr("Command bar flyout"), widget,
		"components/widgets/command_bar.go", codeCommandBarFlyout, 1)
	return i
}

func (i *MenuInterface) createMenu(pos *qt.QPoint) {
	tr := func(s string) string { return gallerycommon.Tr("MenuInterface", s) }
	menu := widgets.NewRoundMenu("", i.QWidget)

	menu.AddAction(gcommon.NewActionFluentIcon(gcommon.Copy, tr("Copy"), nil).QAction)
	menu.AddAction(gcommon.NewActionFluentIcon(gcommon.Cut, tr("Cut"), nil).QAction)

	submenu := widgets.NewRoundMenu(tr("Add to"), i.QWidget)
	submenu.SetIcon(gcommon.Add)
	submenu.AddActions([]*qt.QAction{
		gcommon.NewActionFluentIcon(gcommon.Video, tr("Video"), nil).QAction,
		gcommon.NewActionFluentIcon(gcommon.Audio, tr("Music"), nil).QAction,
	})
	menu.AddMenu(submenu)

	menu.AddActions([]*qt.QAction{
		gcommon.NewActionFluentIcon(gcommon.Paste, tr("Paste"), nil).QAction,
		gcommon.NewActionFluentIcon(gcommon.Cancel, tr("Undo"), nil).QAction,
	})

	menu.AddSeparator()
	menu.AddAction(gcommon.NewActionText(tr("Select all"), nil).QAction)

	acts := menu.MenuActions()
	before := acts[len(acts)-1]
	menu.InsertAction(before, gcommon.NewActionFluentIcon(gcommon.Settings, tr("Settings"), nil).QAction)
	acts = menu.MenuActions()
	before = acts[len(acts)-1]
	menu.InsertAction(before, gcommon.NewActionFluentIcon(gcommon.Help, tr("Help"), nil).QAction)
	menu.InsertAction(before, gcommon.NewActionFluentIcon(gcommon.Feedback, tr("Feedback"), nil).QAction)

	menu.Exec(pos, widgets.MenuAnimationDropDown)
}

func (i *MenuInterface) createCustomWidgetMenu(pos *qt.QPoint) {
	tr := func(s string) string { return gallerycommon.Tr("MenuInterface", s) }
	menu := widgets.NewRoundMenu("", i.QWidget)

	card := NewProfileCard(tr("Shoko"), i.QWidget)
	menu.AddWidget(card.QWidget)

	menu.AddSeparator()
	menu.AddActions([]*qt.QAction{
		gcommon.NewActionFluentIcon(gcommon.People, tr("Manage account profile"), nil).QAction,
		gcommon.NewActionFluentIcon(gcommon.ShoppingCart, tr("Payment method"), nil).QAction,
		gcommon.NewActionFluentIcon(gcommon.Code, tr("Redemption code and gift card"), nil).QAction,
	})
	menu.AddSeparator()
	menu.AddAction(gcommon.NewActionFluentIcon(gcommon.Settings, tr("Settings"), nil).QAction)
	menu.Exec(pos, widgets.MenuAnimationDropDown)
}

func (i *MenuInterface) createCheckableMenu(pos *qt.QPoint) *widgets.CheckableMenu {
	menu := widgets.NewCheckableMenu("", i.QWidget, widgets.MenuIndicatorRadio)

	menu.AddActions([]*qt.QAction{
		i.createTimeAction.QAction,
		i.shootTimeAction.QAction,
		i.modifiedTimeAction.QAction,
		i.nameAction.QAction,
	})
	menu.AddSeparator()
	menu.AddActions([]*qt.QAction{i.ascendAction.QAction, i.descendAction.QAction})

	if pos != nil {
		menu.Exec(pos, widgets.MenuAnimationDropDown)
	}
	return menu
}

func (i *MenuInterface) createCommandBar() *qt.QWidget {
	tr := func(s string) string { return gallerycommon.Tr("MenuInterface", s) }
	bar := widgets.NewCommandBar(i.QWidget)
	bar.SetToolButtonStyle(qt.ToolButtonTextBesideIcon)
	bar.AddActions([]*qt.QAction{
		gcommon.NewActionFluentIcon(gcommon.Add, tr("Add"), nil).QAction,
		gcommon.NewActionFluentIcon(gcommon.Rotate, tr("Rotate"), nil).QAction,
	})
	// A pure icon button: only the icon is drawn, the text becomes its tool tip.
	bar.AddIconAction(gcommon.NewActionFluentIcon(gcommon.Copy, tr("Copy"), nil).QAction)
	bar.AddActions([]*qt.QAction{
		gcommon.NewActionFluentIcon(gcommon.ZoomIn, tr("Zoom in"), nil).QAction,
		gcommon.NewActionFluentIcon(gcommon.ZoomOut, tr("Zoom out"), nil).QAction,
	})
	bar.AddSeparator()
	editAction := gcommon.NewActionFluentIcon(gcommon.Edit, tr("Edit"), nil)
	editAction.SetCheckable(true)
	bar.AddActions([]*qt.QAction{
		editAction.QAction,
		gcommon.NewActionFluentIcon(gcommon.Info, tr("Info"), nil).QAction,
		gcommon.NewActionFluentIcon(gcommon.Delete, tr("Delete"), nil).QAction,
		gcommon.NewActionFluentIcon(gcommon.Share, tr("Share"), nil).QAction,
	})

	button := widgets.NewTransparentDropDownPushButtonIcon(gcommon.Sort, tr("Sort"), i.QWidget)
	button.SetMenu(i.createCheckableMenu(nil).RoundMenu)
	button.SetFixedHeight(34)
	gcommon.SetFont(button.QWidget, 12, 400)
	bar.AddWidget(button.QWidget)

	settingsAction := gcommon.NewActionFluentIcon(gcommon.Settings, tr("Settings"), nil)
	settingsAction.SetShortcut(qt.NewQKeySequence2("Ctrl+I"))
	bar.AddHiddenActions([]*qt.QAction{settingsAction.QAction})
	return bar.QWidget
}

func (i *MenuInterface) createCommandBarFlyout() {
	tr := func(s string) string { return gallerycommon.Tr("MenuInterface", s) }
	view := widgets.NewCommandBarView(nil)

	view.AddAction(gcommon.NewActionFluentIcon(gcommon.Share, tr("Share"), nil).QAction)
	saveAction := gcommon.NewActionFluentIcon(gcommon.Save, tr("Save"), nil)
	saveAction.OnTriggered(i.saveImage)
	view.AddAction(saveAction.QAction)
	view.AddAction(gcommon.NewActionFluentIcon(gcommon.Heart, tr("Add to favorate"), nil).QAction)
	view.AddAction(gcommon.NewActionFluentIcon(gcommon.Delete, tr("Delete"), nil).QAction)

	printAction := gcommon.NewActionFluentIcon(gcommon.Print, tr("Print"), nil)
	printAction.SetShortcut(qt.NewQKeySequence2("Ctrl+P"))
	view.AddHiddenAction(printAction.QAction)
	settingsAction := gcommon.NewActionFluentIcon(gcommon.Settings, tr("Settings"), nil)
	settingsAction.SetShortcut(qt.NewQKeySequence2("Ctrl+S"))
	view.AddHiddenAction(settingsAction.QAction)
	view.ResizeToSuitableWidth()

	x := i.imageLabel.Width()
	pos := i.imageLabel.MapToGlobal(qt.NewQPoint2(x, 0))
	widgets.FlyoutMake(view.FlyoutViewBase, pos, i.QWidget, widgets.FlyoutAnimationFadeIn, true)
}

func (i *MenuInterface) saveImage() {
	path := qt.QFileDialog_GetSaveFileName4(i.QWidget, gallerycommon.Tr("MenuInterface", "Save image"), "", "PNG (*.png)")
	if path == "" {
		return
	}
	img := i.imageLabel.Image()
	img.Save(path)
}

// ProfileCard is a small profile widget shown inside a menu (port of
// app/view/menu_interface.py ProfileCard).
type ProfileCard struct {
	*qt.QWidget
}

// NewProfileCard builds a profile card.
func NewProfileCard(name string, parent *qt.QWidget) *ProfileCard {
	card := &ProfileCard{QWidget: qt.NewQWidget(parent)}
	avatar := widgets.NewAvatarWidgetImage(resource.Icon("shoko.png"), card.QWidget)
	nameLabel := widgets.NewBodyLabelText(name, card.QWidget)
	emailLabel := widgets.NewCaptionLabelText("shokokawaii@outlook.com", card.QWidget)
	logoutButton := widgets.NewHyperlinkButtonURL("https://qfluentwidgets.com", "注销", card.QWidget)

	emailLabel.SetStyleSheet("QLabel{color: #cecece}")
	nameLabel.SetStyleSheet("QLabel{color: #ffffff}")
	gcommon.SetFont(logoutButton.QWidget, 13, 400)

	card.SetFixedSize2(307, 82)
	avatar.SetRadius(24)
	avatar.Move(2, 6)
	nameLabel.Move(64, 13)
	emailLabel.Move(64, 32)
	logoutButton.Move(52, 48)
	return card
}
