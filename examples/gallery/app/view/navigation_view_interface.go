package view

import (
	"strconv"

	gcommon "github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/navigation"
	"github.com/famei/gofluent/components/widgets"
	gallerycommon "github.com/famei/gofluent/examples/gallery/app/common"
	"github.com/famei/gofluent/examples/gallery/app/resource"
	qt "github.com/mappu/miqt/qt"
)

// NavigationViewInterface is the "Navigation" gallery page (port of
// app/view/navigation_view_interface.py).
type NavigationViewInterface struct {
	*GalleryInterface
}

// NewNavigationViewInterface builds the navigation view interface.
func NewNavigationViewInterface(parent *qt.QWidget) *NavigationViewInterface {
	t := gallerycommon.NewTranslator()
	i := &NavigationViewInterface{GalleryInterface: NewGalleryInterface(t.Navigation, "github.com/famei/gofluent/components/navigation", parent)}
	i.SetObjectName("navigationViewInterface")

	breadcrumb := navigation.NewBreadcrumbBar(i.QWidget)
	items := []string{
		gallerycommon.Tr("NavigationViewInterface", "Home"),
		gallerycommon.Tr("NavigationViewInterface", "Documents"),
		gallerycommon.Tr("NavigationViewInterface", "Study"),
		gallerycommon.Tr("NavigationViewInterface", "Janpanese Sensei"),
		gallerycommon.Tr("NavigationViewInterface", "Action Film"),
		gallerycommon.Tr("NavigationViewInterface", "G Cup"),
		gallerycommon.Tr("NavigationViewInterface", "Mikami Yua"),
		gallerycommon.Tr("NavigationViewInterface", "Folder1"),
		gallerycommon.Tr("NavigationViewInterface", "Folder2"),
	}
	for _, item := range items {
		breadcrumb.AddItem(item, item)
	}
	i.AddExampleCard(gallerycommon.Tr("NavigationViewInterface", "Breadcrumb bar"), breadcrumb.QWidget,
		"navigation/breadcrumb_bar/main.go", 1)

	i.AddExampleCard(gallerycommon.Tr("NavigationViewInterface", "A basic pivot"), NewPivotInterface(i.QWidget).QWidget,
		"navigation/pivot/main.go", 0)

	i.AddExampleCard(gallerycommon.Tr("NavigationViewInterface", "A segmented control"), NewSegmentedInterface(i.QWidget).QWidget,
		"navigation/segmented_widget/main.go", 0)

	i.AddExampleCard(gallerycommon.Tr("NavigationViewInterface", "Another segmented control"), i.createToggleToolWidget(),
		"navigation/segmented_tool_widget/main.go", 0)

	card := i.AddExampleCard(gallerycommon.Tr("NavigationViewInterface", "A tab bar"), NewTabInterface(i.QWidget).QWidget,
		"navigation/tab_view/main.go", 1)
	card.TopLayout.SetContentsMargins(12, 0, 0, 0)
	return i
}

func (i *NavigationViewInterface) createToggleToolWidget() *qt.QWidget {
	w := navigation.NewSegmentedToggleToolWidget(i.QWidget)
	w.AddItem("k1", gcommon.SquareSparkle, nil)
	w.AddItem("k2", gcommon.Checkbox, nil)
	w.AddItem("k3", gcommon.Light, nil)
	w.SetCurrentItem("k1")
	return w.QWidget
}

// PivotInterface is a pivot + stacked widget demo (port of PivotInterface).
type PivotInterface struct {
	*qt.QWidget
	pivot         *navigation.Pivot
	stackedWidget *qt.QStackedWidget
	vBoxLayout    *qt.QVBoxLayout
}

// NewPivotInterface builds the pivot interface.
func NewPivotInterface(parent *qt.QWidget) *PivotInterface {
	i := &PivotInterface{QWidget: qt.NewQWidget(parent)}
	i.SetFixedSize2(300, 140)

	i.pivot = navigation.NewPivot(i.QWidget)
	i.stackedWidget = qt.NewQStackedWidget(i.QWidget)
	i.vBoxLayout = qt.NewQVBoxLayout(i.QWidget)

	songInterface := qt.NewQLabel5("Song Interface", i.QWidget)
	albumInterface := qt.NewQLabel5("Album Interface", i.QWidget)
	artistInterface := qt.NewQLabel5("Artist Interface", i.QWidget)

	i.addSubInterface(i.pivot, songInterface, "songInterface", gallerycommon.Tr("PivotInterface", "Song"))
	i.addSubInterface(i.pivot, albumInterface, "albumInterface", gallerycommon.Tr("PivotInterface", "Album"))
	i.addSubInterface(i.pivot, artistInterface, "artistInterface", gallerycommon.Tr("PivotInterface", "Artist"))

	i.vBoxLayout.AddWidget3(i.pivot.QWidget, 0, qt.AlignLeft)
	i.vBoxLayout.AddWidget(i.stackedWidget.QWidget)
	i.vBoxLayout.SetContentsMargins(0, 0, 0, 0)
	gallerycommon.StyleNavigationViewInterface.Apply(i.QWidget, gcommon.ThemeAuto)

	i.stackedWidget.OnCurrentChanged(i.onCurrentIndexChanged)
	i.stackedWidget.SetCurrentWidget(songInterface.QWidget)
	i.pivot.SetCurrentItem(songInterface.ObjectName())

	gcommon.RouterInstance.SetDefaultRouteKey(i.stackedWidget, songInterface.ObjectName())
	return i
}

func (i *PivotInterface) addSubInterface(pivot *navigation.Pivot, widget *qt.QLabel, objectName, text string) {
	widget.SetObjectName(objectName)
	widget.SetAlignment(qt.AlignTop | qt.AlignLeft)
	i.stackedWidget.AddWidget(widget.QWidget)
	pivot.AddItem(objectName, text, func(bool) { i.stackedWidget.SetCurrentWidget(widget.QWidget) }, nil)
}

func (i *PivotInterface) onCurrentIndexChanged(index int) {
	widget := i.stackedWidget.Widget(index)
	if widget == nil {
		return
	}
	i.pivot.SetCurrentItem(widget.ObjectName())
	gcommon.RouterInstance.Push(i.stackedWidget, widget.ObjectName())
}

// SegmentedInterface is a segmented + stacked widget demo (port of
// SegmentedInterface).
type SegmentedInterface struct {
	*qt.QWidget
	segmented     *navigation.SegmentedWidget
	stackedWidget *qt.QStackedWidget
	vBoxLayout    *qt.QVBoxLayout
}

// NewSegmentedInterface builds the segmented interface.
func NewSegmentedInterface(parent *qt.QWidget) *SegmentedInterface {
	i := &SegmentedInterface{QWidget: qt.NewQWidget(parent)}
	i.SetFixedSize2(300, 140)

	i.segmented = navigation.NewSegmentedWidget(i.QWidget)
	i.stackedWidget = qt.NewQStackedWidget(i.QWidget)
	i.vBoxLayout = qt.NewQVBoxLayout(i.QWidget)

	songInterface := qt.NewQLabel5("Song Interface", i.QWidget)
	albumInterface := qt.NewQLabel5("Album Interface", i.QWidget)
	artistInterface := qt.NewQLabel5("Artist Interface", i.QWidget)

	i.addSubInterface(songInterface, "songInterface", gallerycommon.Tr("PivotInterface", "Song"))
	i.addSubInterface(albumInterface, "albumInterface", gallerycommon.Tr("PivotInterface", "Album"))
	i.addSubInterface(artistInterface, "artistInterface", gallerycommon.Tr("PivotInterface", "Artist"))

	// The reference SegmentedInterface removes the AlignLeft-inserted pivot and
	// re-inserts it without alignment (insertWidget(0, self.pivot)) so the
	// segmented control stretches to the full card width and its segments share
	// the width equally.
	i.vBoxLayout.AddWidget2(i.segmented.QWidget, 0)
	i.vBoxLayout.AddWidget(i.stackedWidget.QWidget)
	i.vBoxLayout.SetContentsMargins(0, 0, 0, 0)
	gallerycommon.StyleNavigationViewInterface.Apply(i.QWidget, gcommon.ThemeAuto)

	i.stackedWidget.OnCurrentChanged(i.onCurrentIndexChanged)
	i.stackedWidget.SetCurrentWidget(songInterface.QWidget)
	i.segmented.SetCurrentItem(songInterface.ObjectName())

	gcommon.RouterInstance.SetDefaultRouteKey(i.stackedWidget, songInterface.ObjectName())
	return i
}

func (i *SegmentedInterface) addSubInterface(widget *qt.QLabel, objectName, text string) {
	widget.SetObjectName(objectName)
	widget.SetAlignment(qt.AlignTop | qt.AlignLeft)
	i.stackedWidget.AddWidget(widget.QWidget)
	i.segmented.AddItem(objectName, text, func(bool) { i.stackedWidget.SetCurrentWidget(widget.QWidget) }, nil)
}

func (i *SegmentedInterface) onCurrentIndexChanged(index int) {
	widget := i.stackedWidget.Widget(index)
	if widget == nil {
		return
	}
	i.segmented.SetCurrentItem(widget.ObjectName())
	gcommon.RouterInstance.Push(i.stackedWidget, widget.ObjectName())
}

// TabInterface is a tab bar + stacked widget demo (port of TabInterface).
type TabInterface struct {
	*qt.QWidget
	tabCount int

	tabBar        *widgets.TabBar
	stackedWidget *qt.QStackedWidget
	tabView       *qt.QWidget
	controlPanel  *qt.QFrame

	movableCheckBox          *widgets.CheckBox
	scrollableCheckBox       *widgets.CheckBox
	shadowEnabledCheckBox    *widgets.CheckBox
	tabMaxWidthLabel         *widgets.BodyLabel
	tabMaxWidthSpinBox       *widgets.SpinBox
	closeDisplayModeLabel    *widgets.BodyLabel
	closeDisplayModeComboBox *widgets.ComboBox

	hBoxLayout  *qt.QHBoxLayout
	vBoxLayout  *qt.QVBoxLayout
	panelLayout *qt.QVBoxLayout

	labelWidgets map[string]*qt.QLabel
}

// NewTabInterface builds the tab interface.
func NewTabInterface(parent *qt.QWidget) *TabInterface {
	i := &TabInterface{QWidget: qt.NewQWidget(parent), tabCount: 1, labelWidgets: map[string]*qt.QLabel{}}

	i.tabBar = widgets.NewTabBar(i.QWidget)
	i.stackedWidget = qt.NewQStackedWidget(i.QWidget)
	i.tabView = qt.NewQWidget(i.QWidget)
	i.controlPanel = qt.NewQFrame(i.QWidget)

	i.movableCheckBox = widgets.NewCheckBoxText(gallerycommon.Tr("TabInterface", "IsTabMovable"), i.QWidget)
	i.scrollableCheckBox = widgets.NewCheckBoxText(gallerycommon.Tr("TabInterface", "IsTabScrollable"), i.QWidget)
	i.shadowEnabledCheckBox = widgets.NewCheckBoxText(gallerycommon.Tr("TabInterface", "IsTabShadowEnabled"), i.QWidget)
	i.tabMaxWidthLabel = widgets.NewBodyLabelText(gallerycommon.Tr("TabInterface", "TabMaximumWidth"), i.QWidget)
	i.tabMaxWidthSpinBox = widgets.NewSpinBox(i.QWidget)
	i.closeDisplayModeLabel = widgets.NewBodyLabelText(gallerycommon.Tr("TabInterface", "TabCloseButtonDisplayMode"), i.QWidget)
	i.closeDisplayModeComboBox = widgets.NewComboBox(i.QWidget)

	i.hBoxLayout = qt.NewQHBoxLayout(i.QWidget)
	i.vBoxLayout = qt.NewQVBoxLayout(i.tabView)
	i.panelLayout = qt.NewQVBoxLayout(i.controlPanel.QWidget)

	songInterface := qt.NewQLabel5("Song Interface", i.QWidget)
	albumInterface := qt.NewQLabel5("Album Interface", i.QWidget)
	artistInterface := qt.NewQLabel5("Artist Interface", i.QWidget)

	i.initWidget(songInterface, albumInterface, artistInterface)
	return i
}

func (i *TabInterface) initWidget(song, album, artist *qt.QLabel) {
	i.initLayout()

	i.shadowEnabledCheckBox.SetChecked(true)

	i.tabMaxWidthSpinBox.SetRange(60, 400)
	i.tabMaxWidthSpinBox.SetValue(i.tabBar.TabMaximumWidth())

	i.closeDisplayModeComboBox.AddItem(gallerycommon.Tr("TabInterface", "Always"), nil, widgets.TabCloseButtonDisplayAlways)
	i.closeDisplayModeComboBox.AddItem(gallerycommon.Tr("TabInterface", "OnHover"), nil, widgets.TabCloseButtonDisplayOnHover)
	i.closeDisplayModeComboBox.AddItem(gallerycommon.Tr("TabInterface", "Never"), nil, widgets.TabCloseButtonDisplayNever)
	i.closeDisplayModeComboBox.OnCurrentIndexChanged = i.onDisplayModeChanged

	i.addSubInterface(song, "tabSongInterface", gallerycommon.Tr("TabInterface", "Song"), resource.Icon("MusicNote.png"))
	i.addSubInterface(album, "tabAlbumInterface", gallerycommon.Tr("TabInterface", "Album"), resource.Icon("Dvd.png"))
	i.addSubInterface(artist, "tabArtistInterface", gallerycommon.Tr("TabInterface", "Artist"), resource.Icon("Singer.png"))

	i.controlPanel.SetObjectName("controlPanel")
	gallerycommon.StyleNavigationViewInterface.Apply(i.QWidget, gcommon.ThemeAuto)

	i.connectSignalToSlot()

	gcommon.RouterInstance.SetDefaultRouteKey(i.stackedWidget, song.ObjectName())
}

func (i *TabInterface) connectSignalToSlot() {
	i.movableCheckBox.OnStateChanged(func(int) { i.tabBar.SetMovable(i.movableCheckBox.IsChecked()) })
	i.scrollableCheckBox.OnStateChanged(func(int) { i.tabBar.SetScrollable(i.scrollableCheckBox.IsChecked()) })
	i.shadowEnabledCheckBox.OnStateChanged(func(int) { i.tabBar.SetTabShadowEnabled(i.shadowEnabledCheckBox.IsChecked()) })

	i.tabMaxWidthSpinBox.OnValueChanged(func(value int) { i.tabBar.SetTabMaximumWidth(value) })

	i.tabBar.OnTabAddRequested(i.addTab)
	i.tabBar.OnTabCloseRequested(i.removeTab)

	i.stackedWidget.OnCurrentChanged(i.onCurrentIndexChanged)
}

func (i *TabInterface) initLayout() {
	i.tabBar.SetTabMaximumWidth(200)

	i.SetFixedHeight(280)
	i.controlPanel.SetFixedWidth(220)
	i.hBoxLayout.AddWidget2(i.tabView, 1)
	i.hBoxLayout.AddWidget3(i.controlPanel.QWidget, 0, qt.AlignRight)
	i.hBoxLayout.SetContentsMargins(0, 0, 0, 0)

	i.vBoxLayout.AddWidget(i.tabBar.QWidget)
	i.vBoxLayout.AddWidget(i.stackedWidget.QWidget)
	i.vBoxLayout.SetContentsMargins(0, 0, 0, 0)

	i.panelLayout.SetSpacing(8)
	i.panelLayout.SetContentsMargins(14, 16, 14, 14)

	i.panelLayout.AddWidget(i.movableCheckBox.QWidget)
	i.panelLayout.AddWidget(i.scrollableCheckBox.QWidget)
	i.panelLayout.AddWidget(i.shadowEnabledCheckBox.QWidget)

	i.panelLayout.AddSpacing(4)
	i.panelLayout.AddWidget(i.tabMaxWidthLabel.QWidget)
	i.panelLayout.AddWidget(i.tabMaxWidthSpinBox.QWidget)

	i.panelLayout.AddSpacing(4)
	i.panelLayout.AddWidget(i.closeDisplayModeLabel.QWidget)
	i.panelLayout.AddWidget(i.closeDisplayModeComboBox.QWidget)
}

func (i *TabInterface) addSubInterface(widget *qt.QLabel, objectName, text string, icon *qt.QIcon) {
	widget.SetObjectName(objectName)
	widget.SetAlignment(qt.AlignTop | qt.AlignLeft)
	i.stackedWidget.AddWidget(widget.QWidget)
	i.labelWidgets[objectName] = widget
	i.tabBar.AddTab(objectName, text, icon, func() { i.stackedWidget.SetCurrentWidget(widget.QWidget) })
}

func (i *TabInterface) onDisplayModeChanged(index int) {
	mode := i.closeDisplayModeComboBox.ItemData(index)
	if m, ok := mode.(widgets.TabCloseButtonDisplayMode); ok {
		i.tabBar.SetCloseButtonDisplayMode(m)
	}
}

func (i *TabInterface) onCurrentIndexChanged(index int) {
	widget := i.stackedWidget.Widget(index)
	if widget == nil {
		return
	}
	i.tabBar.SetCurrentTab(widget.ObjectName())
	gcommon.RouterInstance.Push(i.stackedWidget, widget.ObjectName())
}

func (i *TabInterface) addTab() {
	text := "硝子酱一级棒卡哇伊×" + strconv.Itoa(i.tabCount)
	label := qt.NewQLabel5("🥰 "+text, i.QWidget)
	i.addSubInterface(label, text, text, resource.Icon("Smiling_with_heart.png"))
	i.tabCount++
}

func (i *TabInterface) removeTab(index int) {
	item := i.tabBar.TabItem(index)
	if item == nil {
		return
	}
	widget := i.labelWidgets[item.RouteKey()]
	if widget == nil {
		return
	}
	i.stackedWidget.RemoveWidget(widget.QWidget)
	i.tabBar.RemoveTab(index)
	delete(i.labelWidgets, item.RouteKey())
	widget.DeleteLater()
}
