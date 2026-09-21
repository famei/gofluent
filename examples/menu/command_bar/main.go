// Command command_bar migrates examples/menu/command_bar/demo.py. It builds two
// windows: Demo1 holds a CommandBar with a drop-down menu button and hidden
// overflow actions, and Demo2 shows an image that opens a CommandBarView flyout
// on click.
//
// Beyond the Python demo, Demo1 also shows the additions of the Go port: a pure
// icon button (AddIconAction, no text, with a fluent tool tip instead of the
// native one) and AddStretch, which pushes the actions added after it to the right
// edge of the bar.
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
	// toolTips holds the fluent tool tip filters of the icon-only buttons. The
	// filters are owned by their widget, so keeping them is only needed to call
	// HideToolTip / SetToolTipDelay later on.
	toolTips []*widgets.ToolTipFilter
}

func newDemo1() *Demo1 {
	w := &Demo1{QWidget: qt.NewQWidget2()}
	w.hBoxLayout = qt.NewQHBoxLayout(w.QWidget)
	w.commandBar = widgets.NewCommandBar(w.QWidget)
	w.dropDownButton = w.createDropDownButton()

	// The bar fills the window: AddStretch below needs room to push the right hand
	// group to the edge.
	w.hBoxLayout.AddWidget2(w.commandBar.QWidget, 1)
	// The bar is 34px tall, so the window has to leave room for it (the Python
	// demo keeps the default margins and clips the bar instead).
	w.hBoxLayout.SetContentsMargins(8, 3, 8, 3)

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

	// Go port addition: a pure icon button, which is what an action without text
	// looks like as well.
	w.addIconButton(common.Heart, "Favorite")

	// Go port addition: a stretch pushes everything added after it to the right
	// edge of the bar, so a bar can keep a left and a right hand group apart.
	w.commandBar.AddStretch()
	w.commandBar.AddSeparator()
	w.addButton(common.Help, "Help")
	w.addIconButton(common.Feedback, "Feedback")

	// add hidden actions
	sortAct := common.NewActionFluentIcon(common.Sort, "Sort", nil)
	sortAct.OnTriggered(func() { fmt.Println("排序") })
	w.commandBar.AddHiddenAction(sortAct.QAction)

	settingsAct := common.NewActionFluentIcon(common.Settings, "Settings", nil)
	settingsAct.SetShortcut(qt.NewQKeySequence2("Ctrl+S"))
	w.commandBar.AddHiddenAction(settingsAct.QAction)

	// The bar needs a window wider than its content for the stretch to show, and
	// the size is pinned: the window otherwise grows to the size hint of the bar's
	// manually placed children.
	w.SetFixedWidth(780)
	w.SetFixedHeight(40)
	w.SetWindowTitle("Drag window")
	return w
}

// addButton mirrors the Python Demo1.addButton helper: create an action, print
// its text on trigger, and append it to the command bar.
func (w *Demo1) addButton(icon common.FluentIconBase, text string) {
	action := common.NewActionFluentIcon(icon, text, nil)
	action.OnTriggered(func() { fmt.Println(text) })
	w.commandBar.AddAction(action.QAction)
}

// addIconButton is the Go port addition to addButton: the action is added as a
// pure icon button (nothing but the icon is drawn) and shows the fluent tool tip
// instead of the native one.
func (w *Demo1) addIconButton(icon common.FluentIconBase, text string) {
	action := common.NewActionFluentIcon(icon, text, nil)
	action.OnTriggered(func() { fmt.Println(text) })
	button := w.commandBar.AddIconAction(action.QAction)
	w.toolTips = append(w.toolTips,
		widgets.NewToolTipFilter(button.QWidget, 300, widgets.ToolTipPositionTop))
}

func (w *Demo1) onEdit(isChecked bool) {
	if isChecked {
		fmt.Println("Enter edit mode")
	} else {
		fmt.Println("Exit edit mode")
	}
}

func (w *Demo1) createDropDownButton() *widgets.TransparentDropDownPushButton {
	button := widgets.NewTransparentDropDownPushButtonIcon(common.GlobalNavButton, "Menu", w.QWidget)
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

	// The Python demo uses FluentIcon.APPLICATION; the icon set of the Go port is
	// the Segoe Fluent Icons font, whose generic app glyph stands in for it.
	appAct := common.NewActionFluentIcon(common.GenericApp, "App", nil)
	appAct.SetShortcut(qt.NewQKeySequence2("Ctrl+A"))
	view.bar.AddHiddenAction(appAct.QAction)

	settingsAct := common.NewActionFluentIcon(common.Settings, "Settings", nil)
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
