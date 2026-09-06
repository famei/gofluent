// Command line_edit migrates examples/text/line_edit/demo.py.
package main

import (
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
	qt "github.com/mappu/miqt/qt"
)

func newDemo() *qt.QWidget {
	w := qt.NewQWidget2()

	hBoxLayout := qt.NewQHBoxLayout(w)
	lineEdit := widgets.NewSearchLineEdit(w)
	button := widgets.NewPushButtonText("Search", w)

	stands := []string{
		"Star Platinum", "Hierophant Green",
		"Made in Haven", "King Crimson",
		"Silver Chariot", "Crazy diamond",
		"Metallica", "Another One Bites The Dust",
		"Heaven's Door", "Killer Queen",
		"The Grateful Dead", "Stone Free",
		"The World", "Sticky Fingers",
		"Ozone Baby", "Love Love Deluxe",
		"Hermit Purple", "Gold Experience",
		"King Nothing", "Paper Moon King",
		"Scary Monster", "Mandom",
		"20th Century Boy", "Tusk Act 4",
		"Ball Breaker", "Sex Pistols",
		"D4C • Love Train", "Born This Way",
		"SOFT & WET", "Paisley Park",
		"Wonder of U", "Walking Heart",
		"Cream Starter", "November Rain",
		"Smooth Operators", "The Matte Kudasai",
	}
	completer := qt.NewQCompleter6(stands, lineEdit.QObject)
	completer.SetCaseSensitivity(qt.CaseInsensitive)
	completer.SetMaxVisibleItems(10)
	lineEdit.SetCompleter(completer)

	w.Resize(400, 400)
	hBoxLayout.AddWidget3(lineEdit.QWidget, 0, qt.AlignCenter)
	hBoxLayout.AddWidget3(button.QWidget, 0, qt.AlignCenter)

	lineEdit.SetFixedSize2(200, 33)
	lineEdit.SetClearButtonEnabled(true)
	lineEdit.SetPlaceholderText("Search stand")

	return w
}

func main() {
	demo.Run(newDemo)
}
