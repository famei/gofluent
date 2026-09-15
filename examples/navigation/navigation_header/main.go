// Command navigation_header ports PyQt-Fluent-Widgets' examples/navigation/
// navigation_header demo: a NavigationInterface with item headers, separators
// and a QStackedWidget of plain QFrame pages.
package main

import (
	"strings"

	qt "github.com/mappu/miqt/qt"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/navigation"
	"github.com/famei/gofluent/examples/internal/demo"
)

// DemoInterface is a plain QFrame page identified by its route key.
type DemoInterface struct {
	*qt.QFrame
}

func newDemoInterface(text string, parent *qt.QWidget) *DemoInterface {
	w := &DemoInterface{QFrame: qt.NewQFrame(parent)}
	w.SetObjectName(strings.ReplaceAll(text, " ", "-"))
	return w
}

// Window is the navigation header demo window.
type Window struct {
	*qt.QWidget

	hBox                *qt.QHBoxLayout
	navigationInterface *navigation.NavigationInterface
	stackedWidget       *qt.QStackedWidget
	interfaces          map[string]*DemoInterface
}

func newWindow() *Window {
	w := &Window{QWidget: qt.NewQWidget2()}
	w.SetWindowTitle("Navigation Header Demo")
	w.Resize(900, 600)

	w.hBox = qt.NewQHBoxLayout(w.QWidget)
	w.hBox.SetContentsMargins(0, 0, 0, 0)
	w.hBox.SetSpacing(0)

	w.navigationInterface = navigation.NewNavigationInterface(w.QWidget, true, false, true)
	w.stackedWidget = qt.NewQStackedWidget(w.QWidget)
	w.interfaces = map[string]*DemoInterface{}

	w.initNavigation()

	w.hBox.AddWidget(w.navigationInterface.QWidget)
	w.hBox.AddWidget(w.stackedWidget.QWidget)
	w.hBox.SetStretchFactor(w.stackedWidget.QWidget, 1)
	return w
}

func (w *Window) initNavigation() {
	// home
	w.addInterface("home", common.Home, "Home", navigation.NavigationItemPositionTop)
	w.navigationInterface.AddSeparator(navigation.NavigationItemPositionScroll)

	// basic group
	w.navigationInterface.AddItemHeader("Basic Input", navigation.NavigationItemPositionScroll)
	w.addInterface("button", common.Checkbox, "Button", navigation.NavigationItemPositionScroll)
	w.addInterface("input", common.Edit, "Input", navigation.NavigationItemPositionScroll)

	// data group
	w.navigationInterface.AddItemHeader("Data", navigation.NavigationItemPositionScroll)
	w.addInterface("table", common.Document, "Table", navigation.NavigationItemPositionScroll)
	w.addInterface("list", common.GlobalNavButton, "List", navigation.NavigationItemPositionScroll)

	// settings
	w.addInterface("settings", common.Settings, "Settings", navigation.NavigationItemPositionBottom)

	// default
	w.stackedWidget.SetCurrentIndex(0)
	w.navigationInterface.SetCurrentItem("home")

	w.navigationInterface.SetUpdateIndicatorPosOnCollapseFinished(true)
}

func (w *Window) addInterface(routeKey string, icon interface{}, text string, position navigation.NavigationItemPosition) {
	page := newDemoInterface(text, w.QWidget)
	w.interfaces[routeKey] = page
	w.stackedWidget.AddWidget(page.QWidget)
	w.navigationInterface.AddItem(routeKey, icon, text, func(bool) {
		w.stackedWidget.SetCurrentWidget(page.QWidget)
	}, true, position, text, "")
}

func main() {
	demo.Run(func() *qt.QWidget { return newWindow().QWidget })
}
