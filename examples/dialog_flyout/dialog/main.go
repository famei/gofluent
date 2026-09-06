// Command dialog migrates examples/dialog_flyout/dialog/demo.py: a frameless
// Dialog with OK/Cancel buttons.
package main

import (
	"fmt"

	qt "github.com/mappu/miqt/qt"

	"github.com/famei/gofluent/components/dialog_box"
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
)

func newDemo() *qt.QWidget {
	w := qt.NewQWidget2()
	w.SetObjectName("Demo")
	w.Resize(950, 500)
	w.SetStyleSheet("#Demo{background: white}")

	btn := widgets.NewPrimaryPushButtonText("Click Me", w)
	btn.Move(425, 225)
	btn.OnClicked(func() { showDialog(w) })

	return w
}

func showDialog(parent *qt.QWidget) {
	title := "Are you sure you want to delete the folder?"
	content := `If you delete the "Music" folder from the list, the folder will no longer appear in the list, but will not be deleted.`
	dialog := dialog_box.NewDialog(title, content, parent)
	if dialog.Exec() == int(qt.QDialog__Accepted) {
		fmt.Println("Yes button is pressed")
	} else {
		fmt.Println("Cancel button is pressed")
	}
}

func main() {
	demo.Run(newDemo)
}
