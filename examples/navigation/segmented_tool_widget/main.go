// Command segmented_tool_widget ports PyQt-Fluent-Widgets' examples/navigation/
// segmented_tool_widget demo: a SegmentedToggleToolWidget whose icon-only items
// switch a QStackedWidget.
package main

import (
	qt "github.com/mappu/miqt/qt"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/navigation"
	"github.com/famei/gofluent/examples/internal/demo"
)

// segmentedStyle is the demo stylesheet kept verbatim from the Python source.
const segmentedStyle = `
Demo{background: white}
QLabel{
    font: 20px 'Segoe UI';
    background: rgb(242,242,242);
    border-radius: 8px;
}
`

// Demo is the segmented tool widget demo window.
type Demo struct {
	*qt.QWidget

	pivot         *navigation.SegmentedToggleToolWidget
	stackedWidget *qt.QStackedWidget
	hBoxLayout    *qt.QHBoxLayout
	vBoxLayout    *qt.QVBoxLayout

	songInterface   *qt.QLabel
	albumInterface  *qt.QLabel
	artistInterface *qt.QLabel

	labelMap map[string]*qt.QLabel
}

func newDemo() *Demo {
	w := &Demo{QWidget: qt.NewQWidget2()}
	w.labelMap = map[string]*qt.QLabel{}
	// setTheme(Theme.DARK)
	w.SetStyleSheet(segmentedStyle)
	w.Resize(400, 400)

	// NOTE: the Python demo toggles between SegmentedToolWidget and
	// SegmentedToggleToolWidget; the toggle variant is kept here.
	w.pivot = navigation.NewSegmentedToggleToolWidget(w.QWidget)
	w.stackedWidget = qt.NewQStackedWidget(w.QWidget)

	w.hBoxLayout = qt.NewQHBoxLayout2()
	w.vBoxLayout = qt.NewQVBoxLayout(w.QWidget)

	w.songInterface = qt.NewQLabel5("Song Interface", w.QWidget)
	w.albumInterface = qt.NewQLabel5("Album Interface", w.QWidget)
	w.artistInterface = qt.NewQLabel5("Artist Interface", w.QWidget)

	// add items to pivot.
	w.addSubInterface(w.songInterface, "songInterface", common.Music)
	w.addSubInterface(w.albumInterface, "albumInterface", common.Album)
	w.addSubInterface(w.artistInterface, "artistInterface", common.People)

	w.hBoxLayout.AddWidget3(w.pivot.QWidget, 0, qt.AlignCenter)
	w.vBoxLayout.AddLayout(w.hBoxLayout.QLayout)
	w.vBoxLayout.AddWidget(w.stackedWidget.QWidget)
	w.vBoxLayout.SetContentsMargins(30, 10, 30, 30)

	w.stackedWidget.SetCurrentWidget(w.songInterface.QWidget)
	w.pivot.SetCurrentItem(w.songInterface.ObjectName())
	w.pivot.OnCurrentItemChanged(func(k string) {
		if label, ok := w.labelMap[k]; ok {
			w.stackedWidget.SetCurrentWidget(label.QWidget)
		}
	})
	return w
}

func (w *Demo) addSubInterface(label *qt.QLabel, objectName string, icon interface{}) {
	label.SetObjectName(objectName)
	label.SetAlignment(qt.AlignCenter)
	w.stackedWidget.AddWidget(label.QWidget)
	w.pivot.AddItem(objectName, icon, nil)
	w.labelMap[objectName] = label
}

func main() {
	demo.Run(func() *qt.QWidget { return newDemo().QWidget })
}
