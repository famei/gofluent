// Command tab_widget ports PyQt-Fluent-Widgets' examples/navigation/tab_widget
// demo: a TabWidget (tab bar + stacked widget) with closable, addable tabs.
package main

import (
	"embed"
	"fmt"

	qt "github.com/mappu/miqt/qt"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/asset"
	"github.com/famei/gofluent/examples/internal/demo"
	"github.com/famei/gofluent/resources"
)

//go:embed resource/Heart.png resource/Smiling_with_heart.png
var tabIcons embed.FS

// TabInterface is a single tab page with an icon and a subtitle.
type TabInterface struct {
	*qt.QWidget

	iconWidget *widgets.IconWidget
	label      *widgets.SubtitleLabel
	vBoxLayout *qt.QVBoxLayout
}

func newTabInterface(text string, icon interface{}, parent *qt.QWidget) *TabInterface {
	w := &TabInterface{QWidget: qt.NewQWidget(parent)}
	w.iconWidget = widgets.NewIconWidgetIcon(icon, w.QWidget)
	w.label = widgets.NewSubtitleLabelText(text, w.QWidget)
	w.iconWidget.SetFixedSize2(120, 120)

	w.vBoxLayout = qt.NewQVBoxLayout(w.QWidget)
	w.vBoxLayout.SetSpacing(30)
	w.vBoxLayout.AddWidget3(w.iconWidget.QWidget, 0, qt.AlignCenter)
	w.vBoxLayout.AddWidget3(w.label.QWidget, 0, qt.AlignCenter)
	common.SetFont(w.label.QWidget, 24, int(qt.QFont__Normal))
	return w
}

// Window is the tab widget demo window.
type Window struct {
	*qt.QWidget

	tabCount   int
	tabWidget  *widgets.TabWidget
	hBoxLayout *qt.QVBoxLayout
}

func newWindow() *Window {
	w := &Window{QWidget: qt.NewQWidget2()}
	w.tabCount = 1
	w.tabWidget = widgets.NewTabWidget(w.QWidget)
	w.hBoxLayout = qt.NewQVBoxLayout(w.QWidget)

	w.tabWidget.SetMovable(true)

	w.initNavigation()
	w.initWindow()
	return w
}

func (w *Window) initNavigation() {
	w.hBoxLayout.AddWidget(w.tabWidget.QWidget)

	// add tab.
	w.tabWidget.AddTab(
		newTabInterface("Heart", heartIcon(), nil).QWidget,
		"As long as you love me",
		heartIcon(),
		"",
	)

	w.tabWidget.OnCurrentChanged(func(index int) { fmt.Println("current index:", index) })
	w.tabWidget.OnTabCloseRequested(func(index int) { w.tabWidget.RemoveTab(index) })
	w.tabWidget.OnTabAddRequested(w.addNewPage)
}

func (w *Window) initWindow() {
	w.Resize(1100, 750)
	w.SetWindowIcon(asset.QIcon(resources.Images, "images/logo.png"))
	w.SetWindowTitle("PyQt-Fluent-Widgets")

	desktop := qt.QApplication_Desktop().AvailableGeometry2()
	wd, ht := desktop.Width(), desktop.Height()
	w.Move(wd/2-w.Width()/2, ht/2-w.Height()/2)
}

func (w *Window) addNewPage() {
	text := fmt.Sprintf("硝子酱一级棒卡哇伊×%d", w.tabCount)
	w.tabWidget.AddTab(newTabInterface(text, smilingIcon(), nil).QWidget, text, smilingIcon(), "")
	w.tabCount++
}

func heartIcon() *qt.QIcon {
	return asset.QIcon(tabIcons, "resource/Heart.png")
}

func smilingIcon() *qt.QIcon {
	return asset.QIcon(tabIcons, "resource/Smiling_with_heart.png")
}

func main() {
	demo.Run(func() *qt.QWidget { return newWindow().QWidget })
}
