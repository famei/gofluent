// Command command_bar migrates examples/menu/command_bar/demo.py. It builds two
// windows: Demo1 holds a CommandBar with a drop-down menu button and hidden
// overflow actions, and Demo2 shows an image that opens a CommandBarView flyout
// on click.
//
// The Python Demo2 derives from qframelesswindow.FramelessWindow with a
// StandardTitleBar; gofluent/window exposes no StandardTitleBar wrapper, so
// Demo2 is downgraded to a plain QWidget (see newDemo2).
package main

import (
	_ "embed"
	"fmt"

	qt "github.com/mappu/miqt/qt"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
)

//go:embed resource/pink_memory.jpg
var pinkMemoryJPG []byte

// Demo1 is a plain QWidget holding a CommandBar.
type Demo1 struct {
	*qt.QWidget
	hBoxLayout     *qt.QHBoxLayout
	commandBar     *widgets.CommandBar
	dropDownButton *widgets.TransparentDropDownPushButton
}

func newDemo1() *Demo1 {
	w := &Demo1{QWidget: qt.NewQWidget2()}
	w.hBoxLayout = qt.NewQHBoxLayout(w.QWidget)
	w.commandBar = widgets.NewCommandBar(w.QWidget)
	w.dropDownButton = w.createDropDownButton()

	w.hBoxLayout.AddWidget2(w.commandBar.QWidget, 0)

	// change button style
	w.commandBar.SetToolButtonStyle(qt.ToolButtonTextBesideIcon)

	w.addButton(common.Add, "Add")
	w.commandBar.AddSeparator()

	editAct := common.NewActionFluentIcon(common.Edit, "Edit", nil)
	editAct.SetCheckable(true)
	editAct.OnTriggeredWithChecked(w.onEdit)
	w.commandBar.AddAction(editAct.QAction)

	w.addButton(common.Copy, "Copy")
	w.addButton(common.Share, "Share")

	// add custom widget
	w.commandBar.AddWidget(w.dropDownButton.QWidget)

	// add hidden actions
	sortAct := common.NewActionFluentIcon(common.Scroll, "Sort", nil)
	sortAct.OnTriggered(func() { fmt.Println("排序") })
	w.commandBar.AddHiddenAction(sortAct.QAction)

	settingsAct := common.NewActionFluentIcon(common.Setting, "Settings", nil)
	settingsAct.SetShortcut(qt.NewQKeySequence2("Ctrl+S"))
	w.commandBar.AddHiddenAction(settingsAct.QAction)

	w.Resize(240, 40)
	w.SetWindowTitle("Drag window")
	return w
}

// addButton mirrors the Python Demo1.addButton helper: create an action, print
// its text on trigger, and append it to the command bar.
func (w *Demo1) addButton(icon common.FluentIcon, text string) {
	action := common.NewActionFluentIcon(icon, text, nil)
	action.OnTriggered(func() { fmt.Println(text) })
	w.commandBar.AddAction(action.QAction)
}

func (w *Demo1) onEdit(isChecked bool) {
	if isChecked {
		fmt.Println("Enter edit mode")
	} else {
		fmt.Println("Exit edit mode")
	}
}

func (w *Demo1) createDropDownButton() *widgets.TransparentDropDownPushButton {
	button := widgets.NewTransparentDropDownPushButtonIcon(common.Menu, "Menu", w.QWidget)
	button.SetFixedHeight(34)
	common.SetFont(button.QWidget, 12, int(qt.QFont__Normal))

	menu := widgets.NewRoundMenu("", w.QWidget)
	menu.AddAction(common.NewActionFluentIcon(common.Copy, "Copy", nil).QAction)
	menu.AddAction(common.NewActionFluentIcon(common.Cut, "Cut", nil).QAction)
	menu.AddAction(common.NewActionFluentIcon(common.Paste, "Paste", nil).QAction)
	menu.AddAction(common.NewActionFluentIcon(common.Cancel, "Cancel", nil).QAction)
	menu.AddAction(common.NewActionText("Select all", nil).QAction)
	button.SetMenu(menu)
	return button
}

// Demo2 is the downgraded second window (see the package doc comment).
type Demo2 struct {
	*qt.QWidget
	hBoxLayout *qt.QHBoxLayout
	imageLabel *widgets.ImageLabel
}

func newDemo2() *Demo2 {
	w := &Demo2{QWidget: qt.NewQWidget2()}
	w.hBoxLayout = qt.NewQHBoxLayout(w.QWidget)
	w.imageLabel = widgets.NewImageLabelImage(qt.QImage_FromDataWithData(pinkMemoryJPG), w.QWidget)

	w.imageLabel.ScaledToWidth(380)
	w.imageLabel.OnClicked(w.showCommandBar)
	w.hBoxLayout.AddWidget(w.imageLabel.QWidget)

	w.hBoxLayout.SetContentsMargins(0, 80, 0, 0)
	// The Python demo uses the QSS type selector "Demo2{...}"; miqt has no
	// dynamic metaobject for the Go subclass, so the id selector is used with
	// SetObjectName.
	w.SetObjectName("Demo2")
	w.SetStyleSheet("#Demo2{background: white}")
	// The Python demo also sets a window icon from ":/qfluentwidgets/images/logo.png";
	// that Qt resource is not embedded in this demo, so the icon is omitted.
	w.SetWindowTitle("Click Image 👇️🥵")

	return w
}

// showCommandBar opens a CommandBarView flyout anchored to the image label.
func (w *Demo2) showCommandBar() {
	view := newCommandBarFlyoutView(nil)

	view.bar.AddAction(common.NewActionFluentIcon(common.Share, "Share", nil).QAction)
	view.bar.AddAction(common.NewActionFluentIcon(common.Save, "Save", nil).QAction)
	view.bar.AddAction(common.NewActionFluentIcon(common.Delete, "Delete", nil).QAction)

	appAct := common.NewActionFluentIcon(common.Application, "App", nil)
	appAct.SetShortcut(qt.NewQKeySequence2("Ctrl+A"))
	view.bar.AddHiddenAction(appAct.QAction)

	settingsAct := common.NewActionFluentIcon(common.Setting, "Settings", nil)
	settingsAct.SetShortcut(qt.NewQKeySequence2("Ctrl+S"))
	view.bar.AddHiddenAction(settingsAct.QAction)

	view.bar.ResizeToSuitableWidth()

	widgets.FlyoutMake(view.FlyoutViewBase, w.imageLabel.QWidget, w.QWidget, widgets.FlyoutAnimationFadeIn, true)
}

// commandBarFlyoutView composes a CommandBarView inside a FlyoutViewBase so it
// can be passed to widgets.FlyoutMake. The Python CommandBarView derives from
// FlyoutViewBase; the Go port embeds a plain QWidget instead, so the base's own
// rounded background is disabled (CommandBarView paints its own).
type commandBarFlyoutView struct {
	*widgets.FlyoutViewBase
	bar *widgets.CommandBarView
}

func newCommandBarFlyoutView(parent *qt.QWidget) *commandBarFlyoutView {
	w := &commandBarFlyoutView{FlyoutViewBase: widgets.NewFlyoutViewBase(parent)}
	w.SetDrawBackground(false)
	w.bar = widgets.NewCommandBarView(w.QWidget)
	layout := qt.NewQHBoxLayout(w.QWidget)
	layout.SetContentsMargins(0, 0, 0, 0)
	layout.SetSpacing(0)
	layout.AddWidget(w.bar.QWidget)
	return w
}

func main() {
	demo.Run(
		func() *qt.QWidget { return newDemo1().QWidget },
		func() *qt.QWidget { return newDemo2().QWidget },
	)
}
