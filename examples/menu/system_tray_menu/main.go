// Command system_tray_menu migrates examples/menu/system_tray_menu/demo.py.
//
// The demo builds a QSystemTrayIcon whose context menu is a fluent round menu
// (gofluent RoundMenu; the Python SystemTrayMenu is a RoundMenu whose only
// change is a sizeHint tweak that the Go port's fixed-size layout does not
// need). Clicking the tray icon pops the menu, and its "篮球" action shows the
// "坤家军！集合！" message box.
package main

import (
	qt "github.com/mappu/miqt/qt"

	"github.com/famei/gofluent/components/dialog_box"
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/asset"
	"github.com/famei/gofluent/examples/internal/demo"
	"github.com/famei/gofluent/resources"
)

// systemTrayIcon keeps the tray icon reachable for the whole process lifetime so
// its Go wrapper (and therefore the C++ object) is never garbage collected. The
// Python demo stores the same object on the Demo window.
var systemTrayIcon *qt.QSystemTrayIcon

// ikun shows the "坤家军！集合！" message box that the Python demo triggers from
// its system-tray "篮球" action. Python also renames the yes/cancel buttons
// ("献出心脏"/"你干嘛~"); gofluent's MessageDialog keeps those buttons
// unexported, so they keep their default labels.
func ikun(parent *qt.QWidget) {
	content := `巅峰产生虚伪的拥护，黄昏见证真正的使徒 🏀

                         ⠀⠰⢷⢿⠄
                   ⠀⠀⠀⠀⠀⣼⣷⣄
                   ⠀⠀⣤⣿⣇⣿⣿⣧⣿⡄
                   ⢴⠾⠋⠀⠀⠻⣿⣷⣿⣿⡀
                   ⠀⢀⣿⣿⡿⢿⠈⣿
                   ⠀⠀⠀⢠⣿⡿⠁⠀⡊⠀⠙
                   ⠀⠀⠀⢿⣿⠀⠀⠹⣿
                   ⠀⠀⠀⠀⠹⣷⡀⠀⣿⡄
                   ⠀⠀⠀⠀⣀⣼⣿⠀⢈⣧
        `
	w := dialog_box.NewMessageDialog("坤家军！集合！", content, parent)
	w.Exec()
}

func newDemo() *qt.QWidget {
	w := qt.NewQWidget2()
	hBox := qt.NewQHBoxLayout(w)
	label := qt.NewQLabel3("Right-click system tray icon")
	label.SetAlignment(qt.AlignCenter)
	hBox.AddWidget(label.QWidget)

	w.Resize(500, 500)
	// The Python demo uses the QSS type selector "Demo{...}"; miqt has no
	// dynamic metaobject for the Go subclass, so the id selector is used with
	// SetObjectName.
	w.SetObjectName("Demo")
	w.SetStyleSheet("#Demo{background: white} QLabel{font-size: 20px}")

	// Window icon; the tray icon mirrors it, exactly like the Python demo.
	logo := asset.QIcon(resources.Images, "images/logo.png")
	w.SetWindowIcon(logo)
	logo.Delete()

	// Build the system tray icon and its fluent round context menu.
	tray := qt.NewQSystemTrayIcon3(w.QObject)
	tray.SetIcon(w.WindowIcon()) // GoGC-armed copy — do NOT Delete
	tray.SetToolTip("硝子酱一级棒卡哇伊🥰")

	menu := widgets.NewRoundMenu("", w)
	menu.AddActionText("🎤   唱")
	menu.AddActionText("🕺   跳")
	menu.AddActionText("🤘🏼   RAP")
	menu.AddActionText("🎶   Music")
	basketball := menu.AddActionText("🏀   篮球")
	basketball.OnTriggered(func() { ikun(w) })
	tray.SetContextMenu(menu.QMenu)

	tray.Show()
	systemTrayIcon = tray

	return w
}

func main() {
	demo.Run(newDemo)
}
