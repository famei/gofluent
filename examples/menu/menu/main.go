// Command menu migrates examples/menu/menu/demo.py: a right-click context menu
// built from a RoundMenu with checkable items, an "Add to" sub menu, a separator
// and actions inserted before the last item.
package main

import (
	qt "github.com/mappu/miqt/qt"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
)

// Demo is the right-clickable window.
type Demo struct {
	*qt.QWidget
	label *qt.QLabel
}

func newDemo() *Demo {
	w := &Demo{QWidget: qt.NewQWidget2()}
	hBox := qt.NewQHBoxLayout(w.QWidget)
	w.label = qt.NewQLabel3("Right-click your mouse")
	w.label.SetAlignment(qt.AlignCenter)
	hBox.AddWidget(w.label.QWidget)
	w.Resize(400, 400)

	// The Python demo uses the QSS type selector "Demo{...}"; miqt has no
	// dynamic metaobject for the Go subclass, so the equivalent id selector is
	// used together with SetObjectName (see examples/media/avatar_widget).
	w.SetObjectName("Demo")
	w.SetStyleSheet("#Demo{background: white} QLabel{font-size: 20px}")

	w.OnContextMenuEvent(func(super func(e *qt.QContextMenuEvent), e *qt.QContextMenuEvent) {
		w.showContextMenu(e)
	})
	return w
}

func (w *Demo) showContextMenu(e *qt.QContextMenuEvent) {
	menu := widgets.NewRoundMenu("", w.QWidget)

	// add actions
	menu.AddAction(common.NewActionFluentIcon(common.Copy, "Copy", nil).QAction)
	menu.AddAction(common.NewActionFluentIcon(common.Cut, "Cut", nil).QAction)
	menu.MenuActions()[0].SetCheckable(true)
	menu.MenuActions()[0].SetChecked(true)

	// add sub menu
	submenu := widgets.NewRoundMenu("Add to", w.QWidget)
	submenu.SetIcon(common.Add)
	submenu.AddAction(common.NewActionFluentIcon(common.Video, "Video", nil).QAction)
	submenu.AddAction(common.NewActionFluentIcon(common.Music, "Music", nil).QAction)
	menu.AddMenu(submenu)

	// add actions
	menu.AddAction(common.NewActionFluentIcon(common.Paste, "Paste", nil).QAction)
	menu.AddAction(common.NewActionFluentIcon(common.Cancel, "Undo", nil).QAction)

	// add separator

	menu.AddSeparator()
	selectAll := common.NewActionText("Select all", nil)
	selectAll.SetShortcut(qt.NewQKeySequence2("Ctrl+A"))
	menu.AddAction(selectAll.QAction)

	// insert actions before the last action ("Select all")
	before := menu.MenuActions()[len(menu.MenuActions())-1]
	settingsAct := common.NewActionFluentIcon(common.Setting, "Settings", nil)
	settingsAct.SetShortcut(qt.NewQKeySequence2("Ctrl+S"))
	menu.InsertAction(before, settingsAct.QAction)

	helpAct := common.NewActionFluentIcon(common.Help, "Help", nil)
	helpAct.SetShortcut(qt.NewQKeySequence2("Ctrl+H"))
	feedbackAct := common.NewActionFluentIcon(common.Feedback, "Feedback", nil)
	feedbackAct.SetShortcut(qt.NewQKeySequence2("Ctrl+F"))
	menu.InsertAction(before, helpAct.QAction)
	menu.InsertAction(before, feedbackAct.QAction)

	// the second-to-last action ("Feedback") is checkable and checked
	actions := menu.MenuActions()
	actions[len(actions)-2].SetCheckable(true)
	actions[len(actions)-2].SetChecked(true)

	// show menu
	pos := e.GlobalPos()
	menu.Exec(pos, widgets.MenuAnimationDropDown)

}

func main() {
	demo.Run(func() *qt.QWidget { return newDemo().QWidget })
}
