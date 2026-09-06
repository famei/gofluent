// Command breadcrumb_bar ports PyQt-Fluent-Widgets' examples/navigation/
// breadcrumb_bar demo: a BreadcrumbBar drives a QStackedWidget whose pages are
// added on demand from a line edit.
package main

import (
	"fmt"

	qt "github.com/mappu/miqt/qt"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/navigation"
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
)

// Demo is the breadcrumb bar demo window.
type Demo struct {
	*qt.QWidget

	breadcrumbBar *navigation.BreadcrumbBar
	stackedWidget *qt.QStackedWidget
	lineEdit      *widgets.LineEdit
	addButton     *widgets.PrimaryToolButton

	vBoxLayout     *qt.QVBoxLayout
	lineEditLayout *qt.QHBoxLayout

	// ifaceMap replaces the Python findChild(SubtitleLabel, objectName) lookup.
	ifaceMap map[string]*widgets.SubtitleLabel
	counter  int
}

func newDemo() *Demo {
	w := &Demo{QWidget: qt.NewQWidget2()}
	w.ifaceMap = map[string]*widgets.SubtitleLabel{}

	// NOTE: setTheme(Theme.DARK) is active in this demo (unlike most others).
	common.SetTheme(common.ThemeDark, false, false)
	w.SetStyleSheet("Demo{background:rgb(32,32,32)}")

	w.breadcrumbBar = navigation.NewBreadcrumbBar(w.QWidget)
	w.stackedWidget = qt.NewQStackedWidget(w.QWidget)
	w.lineEdit = widgets.NewLineEdit(w.QWidget)
	w.addButton = widgets.NewPrimaryToolButtonIcon(common.Send, w.QWidget)

	w.vBoxLayout = qt.NewQVBoxLayout(w.QWidget)
	w.lineEditLayout = qt.NewQHBoxLayout2()

	w.addButton.OnClicked(func() { w.addInterface(w.lineEdit.Text()) })
	w.lineEdit.OnReturnPressed(func() { w.addInterface(w.lineEdit.Text()) })
	w.lineEdit.SetPlaceholderText("Enter the name of interface")

	// NOTE: adjust the size of breadcrumb item.
	common.SetFont(w.breadcrumbBar.QWidget, 26, int(qt.QFont__DemiBold))
	w.breadcrumbBar.SetSpacing(20)
	w.breadcrumbBar.OnCurrentItemChanged(func(objectName string) { w.switchInterface(objectName) })

	w.addInterface("Home")
	w.addInterface("Documents")

	w.vBoxLayout.SetContentsMargins(20, 20, 20, 20)
	w.vBoxLayout.AddWidget(w.breadcrumbBar.QWidget)
	w.vBoxLayout.AddWidget(w.stackedWidget.QWidget)
	w.vBoxLayout.AddLayout(w.lineEditLayout.QLayout)

	w.lineEditLayout.AddWidget2(w.lineEdit.QWidget, 1)
	w.lineEditLayout.AddWidget(w.addButton.QWidget)
	w.Resize(500, 500)
	return w
}

func (w *Demo) addInterface(text string) {
	if text == "" {
		return
	}

	label := widgets.NewSubtitleLabelText(text, w.QWidget)
	w.counter++
	objectName := fmt.Sprintf("iface%d", w.counter)
	label.SetObjectName(objectName)
	label.SetAlignment(qt.AlignCenter)

	w.lineEdit.Clear()
	w.stackedWidget.AddWidget(label.QWidget)
	w.stackedWidget.SetCurrentWidget(label.QWidget)

	// !IMPORTANT: add breadcrumb item.
	w.ifaceMap[objectName] = label
	w.breadcrumbBar.AddItem(objectName, text)
}

// switchInterface switches to the page for the given breadcrumb item.
func (w *Demo) switchInterface(objectName string) {
	if label, ok := w.ifaceMap[objectName]; ok {
		w.stackedWidget.SetCurrentWidget(label.QWidget)
	}
}

func main() {
	demo.Run(func() *qt.QWidget { return newDemo().QWidget })
}
