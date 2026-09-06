// Command folder_list_dialog migrates examples/dialog_flyout/folder_list_dialog/demo.py:
// a FolderListDialog that edits a list of watched folders.
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
	w.Resize(800, 720)
	w.SetStyleSheet("#Demo{background: white}")

	btn := widgets.NewPrimaryPushButtonText("Click Me", w)
	btn.Move(352, 300)
	btn.OnClicked(func() { showDialog(w) })

	return w
}

func showDialog(parent *qt.QWidget) {
	folderPaths := []string{"D:/KuGou", "C:/Users/shoko/Documents/Music"}
	title := "Build your collection from your local music files"
	content := "Right now, we're watching these folders:"
	dialog := dialog_box.NewFolderListDialog(folderPaths, title, content, parent)
	dialog.OnFolderChanged = func(folders []string) { fmt.Println(folders) }
	dialog.Exec()
}

func main() {
	demo.Run(newDemo)
}
