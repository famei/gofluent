// Command button migrates examples/basic_input/button/demo.py: it showcases
// every fluent button flavour (tool buttons, push buttons, drop-down/split
// buttons, toggle/pill/hyperlink variants) in two grid-based views.
package main

import (
	"fmt"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
	qt "github.com/mappu/miqt/qt"
)

// newButtonView builds the shared white-background base view used by the two
// demo panes (the Python ButtonView class).
func newButtonView() *qt.QWidget {
	w := qt.NewQWidget2()
	w.SetObjectName("ButtonView")
	w.SetStyleSheet("#ButtonView{background: rgb(255,255,255)}")
	return w
}

// toolButtonDemo is the ToolButtonDemo pane.
type toolButtonDemo struct {
	*qt.QWidget
	grid *qt.QGridLayout
}

func newToolButtonDemo() *toolButtonDemo {
	w := &toolButtonDemo{QWidget: newButtonView()}

	menu := widgets.NewRoundMenu("", w.QWidget)
	menu.AddAction(common.NewActionFluentIcon(common.SendFill, "Send", nil).QAction)
	menu.AddAction(common.NewActionFluentIcon(common.Save, "Save", nil).QAction)

	toolButton := widgets.NewToolButtonIcon(common.Settings, w.QWidget)

	dropDownToolButton := widgets.NewDropDownToolButtonIcon(common.Mail, w.QWidget)
	dropDownToolButton.SetMenu(menu)

	splitToolButton := widgets.NewSplitToolButtonIcon(common.Code, w.QWidget)
	splitToolButton.SetFlyout(menu)

	primaryToolButton := widgets.NewPrimaryToolButtonIcon(common.Settings, w.QWidget)

	primaryDropDownToolButton := widgets.NewPrimaryDropDownToolButtonIcon(common.Mail, w.QWidget)
	primaryDropDownToolButton.SetMenu(menu)

	primarySplitToolButton := widgets.NewPrimarySplitToolButtonIcon(common.Code, w.QWidget)
	primarySplitToolButton.SetFlyout(menu)

	toggleToolButton := widgets.NewToggleToolButtonIcon(common.Settings, w.QWidget)
	toggleToolButton.OnToggled(func(bool) { fmt.Println("Toggled") })
	toggleToolButton.Toggle()

	transparentToggleToolButton := widgets.NewTransparentToggleToolButtonIcon(common.Code, w.QWidget)

	transparentToolButton := widgets.NewTransparentToolButtonIcon(common.Mail, w.QWidget)

	transparentDropDownToolButton := widgets.NewTransparentDropDownToolButtonIcon(common.Mail, w.QWidget)
	transparentDropDownToolButton.SetMenu(menu)

	pillToolButton1 := widgets.NewPillToolButtonIcon(common.Calendar, w.QWidget)
	pillToolButton2 := widgets.NewPillToolButtonIcon(common.Calendar, w.QWidget)
	pillToolButton3 := widgets.NewPillToolButtonIcon(common.Calendar, w.QWidget)
	pillToolButton2.SetDisabled(true)
	pillToolButton3.SetChecked(true)
	pillToolButton3.SetDisabled(true)

	w.grid = qt.NewQGridLayout(w.QWidget)
	w.grid.AddWidget2(toolButton.QWidget, 0, 0)
	w.grid.AddWidget2(dropDownToolButton.QWidget, 0, 1)
	w.grid.AddWidget2(splitToolButton.QWidget, 0, 2)
	w.grid.AddWidget2(primaryToolButton.QWidget, 1, 0)
	w.grid.AddWidget2(primaryDropDownToolButton.QWidget, 1, 1)
	w.grid.AddWidget2(primarySplitToolButton.QWidget, 1, 2)
	w.grid.AddWidget2(toggleToolButton.QWidget, 2, 0)
	w.grid.AddWidget2(transparentToggleToolButton.QWidget, 2, 1)
	w.grid.AddWidget2(transparentToolButton.QWidget, 3, 0)
	w.grid.AddWidget2(transparentDropDownToolButton.QWidget, 3, 1)
	w.grid.AddWidget2(pillToolButton1.QWidget, 4, 0)
	w.grid.AddWidget2(pillToolButton2.QWidget, 4, 1)
	w.grid.AddWidget2(pillToolButton3.QWidget, 4, 2)

	w.Resize(300, 300)
	return w
}

// pushButtonDemo is the PushButtonDemo pane.
type pushButtonDemo struct {
	*qt.QWidget
	grid *qt.QGridLayout
}

func newPushButtonDemo() *pushButtonDemo {
	w := &pushButtonDemo{QWidget: newButtonView()}

	menu := widgets.NewRoundMenu("", w.QWidget)
	menu.AddAction(common.NewActionFluentIcon(common.LeafTwo, "Basketball", nil).QAction)
	menu.AddAction(common.NewActionFluentIcon(common.MusicAlbum, "Sing", nil).QAction)
	menu.AddAction(common.NewActionFluentIcon(common.Audio, "Music", nil).QAction)

	pushButton1 := widgets.NewPushButtonText("Standard push button", nil)
	pushButton2 := widgets.NewPushButtonIcon(common.Folder, "Standard push button with icon", w.QWidget)

	primaryButton1 := widgets.NewPrimaryPushButtonText("Accent style button", w.QWidget)
	primaryButton2 := widgets.NewPrimaryPushButtonIcon(common.Sync, "Accent style button with icon", w.QWidget)

	transparentPushButton1 := widgets.NewTransparentPushButtonText("Transparent push button", w.QWidget)
	transparentPushButton2 := widgets.NewTransparentPushButtonIcon(common.Library, "Transparent push button", w.QWidget)

	toggleButton1 := widgets.NewToggleButtonText("Toggle push button", w.QWidget)
	toggleButton2 := widgets.NewToggleButtonIcon(common.Send, "Toggle push button", w.QWidget)

	transparentTogglePushButton1 := widgets.NewTransparentTogglePushButtonText("Transparent toggle button", w.QWidget)
	transparentTogglePushButton2 := widgets.NewTransparentTogglePushButtonIcon(common.Library, "Transparent toggle button", w.QWidget)

	dropDownPushButton1 := widgets.NewDropDownPushButtonText("Email", w.QWidget)
	dropDownPushButton2 := widgets.NewDropDownPushButtonIcon(common.Mail, "Email", w.QWidget)
	dropDownPushButton1.SetMenu(menu)
	dropDownPushButton2.SetMenu(menu)

	primaryDropDownPushButton1 := widgets.NewPrimaryDropDownPushButtonText("Email", w.QWidget)
	primaryDropDownPushButton2 := widgets.NewPrimaryDropDownPushButtonIcon(common.Mail, "Email", w.QWidget)
	primaryDropDownPushButton1.SetMenu(menu)
	primaryDropDownPushButton2.SetMenu(menu)

	transparentDropDownPushButton1 := widgets.NewTransparentDropDownPushButtonText("Email", w.QWidget)
	transparentDropDownPushButton2 := widgets.NewTransparentDropDownPushButtonIcon(common.Mail, "Email", w.QWidget)
	transparentDropDownPushButton1.SetMenu(menu)
	transparentDropDownPushButton2.SetMenu(menu)

	splitPushButton1 := widgets.NewSplitPushButtonText("Split push button", w.QWidget)
	splitPushButton2 := widgets.NewSplitPushButtonIcon(common.Code, "Split push button", w.QWidget)
	splitPushButton1.SetFlyout(menu)
	splitPushButton2.SetFlyout(menu)

	primarySplitPushButton1 := widgets.NewPrimarySplitPushButtonText("Split push button", w.QWidget)
	primarySplitPushButton2 := widgets.NewPrimarySplitPushButtonIcon(common.Code, "Split push button", w.QWidget)
	primarySplitPushButton1.SetFlyout(menu)
	primarySplitPushButton2.SetFlyout(menu)

	hyperlinkButton1 := widgets.NewHyperlinkButtonURL("https://qfluentwidgets.com", "Hyper link button", w.QWidget)
	hyperlinkButton2 := widgets.NewHyperlinkButtonIcon(common.Link, "https://qfluentwidgets.com", "Hyper link button", w.QWidget)

	pillPushButton1 := widgets.NewPillPushButtonText("Pill Push Button", w.QWidget)
	pillPushButton2 := widgets.NewPillPushButtonIcon(common.Calendar, "Pill Push Button", w.QWidget)

	w.grid = qt.NewQGridLayout(w.QWidget)
	w.grid.AddWidget2(pushButton1.QWidget, 0, 0)
	w.grid.AddWidget2(pushButton2.QWidget, 0, 1)
	w.grid.AddWidget2(primaryButton1.QWidget, 1, 0)
	w.grid.AddWidget2(primaryButton2.QWidget, 1, 1)
	w.grid.AddWidget2(transparentPushButton1.QWidget, 2, 0)
	w.grid.AddWidget2(transparentPushButton2.QWidget, 2, 1)

	w.grid.AddWidget2(toggleButton1.QWidget, 3, 0)
	w.grid.AddWidget2(toggleButton2.QWidget, 3, 1)
	w.grid.AddWidget2(transparentTogglePushButton1.QWidget, 4, 0)
	w.grid.AddWidget2(transparentTogglePushButton2.QWidget, 4, 1)

	w.grid.AddWidget2(splitPushButton1.QWidget, 5, 0)
	w.grid.AddWidget2(splitPushButton2.QWidget, 5, 1)
	w.grid.AddWidget2(primarySplitPushButton1.QWidget, 6, 0)
	w.grid.AddWidget2(primarySplitPushButton2.QWidget, 6, 1)

	w.grid.AddWidget4(dropDownPushButton1.QWidget, 7, 0, qt.AlignLeft)
	w.grid.AddWidget4(dropDownPushButton2.QWidget, 7, 1, qt.AlignLeft)
	w.grid.AddWidget4(primaryDropDownPushButton1.QWidget, 8, 0, qt.AlignLeft)
	w.grid.AddWidget4(primaryDropDownPushButton2.QWidget, 8, 1, qt.AlignLeft)
	w.grid.AddWidget4(transparentDropDownPushButton1.QWidget, 9, 0, qt.AlignLeft)
	w.grid.AddWidget4(transparentDropDownPushButton2.QWidget, 9, 1, qt.AlignLeft)

	w.grid.AddWidget4(pillPushButton1.QWidget, 10, 0, qt.AlignLeft)
	w.grid.AddWidget4(pillPushButton2.QWidget, 10, 1, qt.AlignLeft)

	w.grid.AddWidget4(hyperlinkButton1.QWidget, 11, 0, qt.AlignLeft)
	w.grid.AddWidget4(hyperlinkButton2.QWidget, 11, 1, qt.AlignLeft)

	w.Resize(600, 700)
	return w
}

func main() {
	demo.Run(
		func() *qt.QWidget { return newToolButtonDemo().QWidget },
		func() *qt.QWidget { return newPushButtonDemo().QWidget },
	)
}
