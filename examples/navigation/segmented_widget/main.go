// Command segmented_widget ports PyQt-Fluent-Widgets' examples/navigation/
// segmented_widget demo: a SegmentedWidget whose text items switch a
// QStackedWidget.
package main

import (
	qt "github.com/mappu/miqt/qt"

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

// Demo is the segmented widget demo window.
type Demo struct {
	*qt.QWidget

	pivot         *navigation.SegmentedWidget
	stackedWidget *qt.QStackedWidget
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

	w.pivot = navigation.NewSegmentedWidget(w.QWidget)
	w.stackedWidget = qt.NewQStackedWidget(w.QWidget)
	w.vBoxLayout = qt.NewQVBoxLayout(w.QWidget)

	w.songInterface = qt.NewQLabel5("Song Interface", w.QWidget)
	w.albumInterface = qt.NewQLabel5("Album Interface", w.QWidget)
	w.artistInterface = qt.NewQLabel5("Artist Interface", w.QWidget)

	// add items to pivot.
	w.addSubInterface(w.songInterface, "songInterface", "Song")
	w.addSubInterface(w.albumInterface, "albumInterface", "Album")
	w.addSubInterface(w.artistInterface, "artistInterface", "Artist")

	w.vBoxLayout.AddWidget(w.pivot.QWidget)
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

func (w *Demo) addSubInterface(label *qt.QLabel, objectName, text string) {
	label.SetObjectName(objectName)
	label.SetAlignment(qt.AlignCenter)
	w.stackedWidget.AddWidget(label.QWidget)
	w.pivot.AddItem(objectName, text, nil, nil)
	w.labelMap[objectName] = label
}

func main() {
	demo.Run(func() *qt.QWidget { return newDemo().QWidget })
}
