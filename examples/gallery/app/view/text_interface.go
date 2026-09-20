package view

import (
	"github.com/famei/gofluent/components/widgets"
	gallerycommon "github.com/famei/gofluent/examples/gallery/app/common"
	qt "github.com/mappu/miqt/qt"
)

// TextInterface is the "Text" gallery page (port of app/view/text_interface.py).
type TextInterface struct {
	*GalleryInterface
}

// NewTextInterface builds the text interface.
func NewTextInterface(parent *qt.QWidget) *TextInterface {
	t := gallerycommon.NewTranslator()
	i := &TextInterface{GalleryInterface: NewGalleryInterface(t.Text, "github.com/famei/gofluent/components/widgets", parent)}
	i.SetObjectName("textInterface")

	tr := func(s string) string { return gallerycommon.Tr("TextInterface", s) }

	const src = "text/line_edit/main.go"

	lineEdit := widgets.NewLineEdit(i.QWidget)
	lineEdit.SetText(tr("ko no dio da！"))
	lineEdit.SetClearButtonEnabled(true)
	i.AddExampleCard(tr("A LineEdit with a clear button"), lineEdit.QWidget, "components/widgets/line_edit.go", codeLineEdit, 0)

	searchEdit := widgets.NewSearchLineEdit(i.QWidget)
	searchEdit.SetPlaceholderText(tr("Type a stand name"))
	searchEdit.SetClearButtonEnabled(true)
	searchEdit.SetFixedWidth(230)
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
	completer := qt.NewQCompleter3(stands)
	completer.SetCaseSensitivity(qt.CaseInsensitive)
	completer.SetMaxVisibleItems(10)
	searchEdit.SetCompleter(completer)
	i.AddExampleCard(tr("A autosuggest line edit"), searchEdit.QWidget, "components/widgets/line_edit.go", codeSearchLineEdit, 0)

	passwordLineEdit := widgets.NewPasswordLineEdit(i.QWidget)
	passwordLineEdit.SetFixedWidth(230)
	passwordLineEdit.SetPlaceholderText(tr("Enter your password"))
	i.AddExampleCard(tr("A password line edit"), passwordLineEdit.QWidget, "components/widgets/line_edit.go", codePasswordLineEdit, 0)

	const ssrc = "text/spin_box/main.go"
	i.AddExampleCard(tr("A SpinBox with a spin button"), widgets.NewSpinBox(i.QWidget).QWidget, "components/widgets/spin_box.go", codeSpinBox, 0)
	i.AddExampleCard(tr("A DoubleSpinBox with a spin button"), widgets.NewDoubleSpinBox(i.QWidget).QWidget, "components/widgets/spin_box.go", codeDoubleSpinBox, 0)
	i.AddExampleCard(tr("A DateEdit with a spin button"), widgets.NewDateEdit(i.QWidget).QWidget, "components/widgets/spin_box.go", codeDateEdit, 0)
	i.AddExampleCard(tr("A TimeEdit with a spin button"), widgets.NewTimeEdit(i.QWidget).QWidget, "components/widgets/spin_box.go", codeTimeEdit, 0)
	i.AddExampleCard(tr("A DateTimeEdit with a spin button"), widgets.NewDateTimeEdit(i.QWidget).QWidget, "components/widgets/spin_box.go", codeDateTimeEdit, 0)

	textEdit := widgets.NewTextEdit(i.QWidget)
	textEdit.SetMarkdown("## Steel Ball Run \n * Johnny Joestar 🦄 \n * Gyro Zeppeli 🐴 ")
	textEdit.SetFixedHeight(150)
	i.AddExampleCard(tr("A simple TextEdit"), textEdit.QWidget, "components/widgets/line_edit.go", codeTextEdit, 1)
	return i
}
