package view

import (
	gcommon "github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/navigation"
	gallerycommon "github.com/famei/gofluent/examples/gallery/app/common"
	"github.com/famei/gofluent/examples/gallery/app/resource"
	"github.com/famei/gofluent/window"
	qt "github.com/mappu/miqt/qt"
)

// MainWindow is the gallery main window (port of app/view/main_window.py).
//
// The gofluent window package's FluentWindow does not expose its navigation
// interface (needed for the separator and the bottom "Price" item), so the
// window is assembled from the exported building blocks, mirroring
// window.NewFluentWindow's layout.
type MainWindow struct {
	*window.FluentWidget

	hBoxLayout    *qt.QHBoxLayout
	stackedWidget *window.StackedWidget
	nav           *navigation.NavigationInterface
	widgetLayout  *qt.QHBoxLayout

	splashScreen  *window.SplashScreen
	themeListener *gcommon.SystemThemeListener

	interfaces []*GalleryInterface

	homeInterface           *HomeInterface
	iconInterface           *IconInterface
	basicInputInterface     *BasicInputInterface
	dateTimeInterface       *DateTimeInterface
	dialogInterface         *DialogInterface
	layoutInterface         *LayoutInterface
	menuInterface           *MenuInterface
	navigationViewInterface *NavigationViewInterface
	scrollInterface         *ScrollInterface
	statusInfoInterface     *StatusInfoInterface
	settingInterface        *SettingInterface
	textInterface           *TextInterface
	viewInterface           *ViewInterface
}

// NewMainWindow builds the gallery main window.
func NewMainWindow() *MainWindow {
	w := &MainWindow{FluentWidget: window.NewFluentWidget(nil)}
	w.SetTitleBar(window.NewFluentTitleBar(w.QWidget).QWidget)

	w.hBoxLayout = qt.NewQHBoxLayout(w.QWidget)
	w.stackedWidget = window.NewStackedWidget(w.QWidget)
	w.hBoxLayout.SetSpacing(0)
	w.hBoxLayout.SetContentsMargins(0, 0, 0, 0)
	gcommon.FluentStyleSheet(gcommon.FluentWindow).Apply(w.stackedWidget.QWidget, gcommon.ThemeAuto)

	w.nav = navigation.NewNavigationInterface(w.QWidget, true, true, true)
	w.widgetLayout = qt.NewQHBoxLayout2()

	w.hBoxLayout.AddWidget(w.nav.QWidget)
	w.hBoxLayout.AddLayout(w.widgetLayout.QLayout)
	w.hBoxLayout.SetStretchFactor2(w.widgetLayout.QLayout, 1)

	w.widgetLayout.AddWidget(w.stackedWidget.QWidget)
	// Reserve the top 48px for the title bar on the content only, matching
	// window.NewFluentWindow: the navigation interface stays at y=0 (its top
	// strip is covered by the raised title bar) and the stacked widget starts
	// below the title bar.
	w.widgetLayout.SetContentsMargins(0, 48, 0, 0)

	w.nav.OnDisplayModeChanged(func(navigation.NavigationDisplayMode) { w.TitleBar().Raise() })
	w.TitleBar().Raise()
	w.positionTitleBar()
	w.OnResizeEvent(func(super func(e *qt.QResizeEvent), e *qt.QResizeEvent) {
		super(e)
		w.positionTitleBar()
	})

	w.initWindow()

	w.themeListener = gcommon.NewSystemThemeListener()

	w.homeInterface = NewHomeInterface(w.QWidget)
	w.iconInterface = NewIconInterface(w.QWidget)
	w.basicInputInterface = NewBasicInputInterface(w.QWidget)
	w.dateTimeInterface = NewDateTimeInterface(w.QWidget)
	w.dialogInterface = NewDialogInterface(w.QWidget)
	w.layoutInterface = NewLayoutInterface(w.QWidget)
	w.menuInterface = NewMenuInterface(w.QWidget)
	w.navigationViewInterface = NewNavigationViewInterface(w.QWidget)
	w.scrollInterface = NewScrollInterface(w.QWidget)
	w.statusInfoInterface = NewStatusInfoInterface(w.QWidget)
	w.settingInterface = NewSettingInterface(w.QWidget)
	w.textInterface = NewTextInterface(w.QWidget)
	w.viewInterface = NewViewInterface(w.QWidget)

	w.interfaces = []*GalleryInterface{
		w.iconInterface.GalleryInterface,
		w.basicInputInterface.GalleryInterface,
		w.dateTimeInterface.GalleryInterface,
		w.dialogInterface.GalleryInterface,
		w.layoutInterface.GalleryInterface,
		w.menuInterface.GalleryInterface,
		w.navigationViewInterface.GalleryInterface,
		w.scrollInterface.GalleryInterface,
		w.statusInfoInterface.GalleryInterface,
		w.textInterface.GalleryInterface,
		w.viewInterface.GalleryInterface,
	}

	w.connectSignalToSlot()
	w.initNavigation()
	w.splashScreen.Finish()

	w.themeListener.Start()
	return w
}

func (w *MainWindow) positionTitleBar() {
	tb := w.TitleBar()
	if tb == nil {
		return
	}
	tb.Move(46, 0)
	tb.Resize(w.Width()-46, tb.Height())
}

func (w *MainWindow) initWindow() {
	w.Resize(960, 780)
	w.SetMinimumWidth(760)
	w.SetWindowIcon(resource.Icon("logo.png"))
	w.SetWindowTitle("PyQt-Fluent-Widgets")

	w.SetMicaEffectEnabled(gallerycommon.ConfigInstance.Get(gallerycommon.ConfigInstance.MicaEnabled).(bool))

	w.splashScreen = window.NewSplashScreen(w.WindowIcon(), w.QWidget, false)
	w.splashScreen.SetIconSize(qt.NewQSize2(106, 106))
	w.splashScreen.Raise()

	desktop := qt.QApplication_Desktop().AvailableGeometry2()
	w.Move(desktop.Width()/2-w.Width()/2, desktop.Height()/2-w.Height()/2)
	w.Show()
	qt.QCoreApplication_ProcessEvents()

	w.OnCloseEvent(func(super func(e *qt.QCloseEvent), e *qt.QCloseEvent) {
		w.themeListener.Stop()
		super(e)
	})
}

func (w *MainWindow) connectSignalToSlot() {
	gallerycommon.SignalBusInstance.OnMicaEnableChanged(func(enabled bool) { w.SetMicaEffectEnabled(enabled) })
	gallerycommon.SignalBusInstance.OnSwitchToSample(func(routeKey string, index int) { w.switchToSample(routeKey, index) })
	gallerycommon.SignalBusInstance.OnSupport(w.onSupport)
}

func (w *MainWindow) addSubInterface(widget *qt.QWidget, icon interface{}, text string, position navigation.NavigationItemPosition) {
	widget.SetProperty("isStackedTransparent", qt.NewQVariant11(false))
	w.stackedWidget.AddWidget(widget)

	routeKey := widget.ObjectName()
	w.nav.AddItem(routeKey, icon, text, func(bool) { w.stackedWidget.SetCurrentWidget(widget, false) }, true, position, text, "")

	if w.stackedWidget.Count() == 1 {
		w.stackedWidget.OnCurrentChanged(w.onCurrentInterfaceChanged)
		w.nav.SetCurrentItem(routeKey)
		gcommon.RouterInstance.SetDefaultRouteKey(w.stackedWidget.QStackedWidget(), routeKey)
	}
}

func (w *MainWindow) onCurrentInterfaceChanged(index int) {
	widget := w.stackedWidget.Widget(index)
	if widget == nil {
		return
	}
	w.nav.SetCurrentItem(widget.ObjectName())
	gcommon.RouterInstance.Push(w.stackedWidget.QStackedWidget(), widget.ObjectName())
}

func (w *MainWindow) initNavigation() {
	t := gallerycommon.NewTranslator()
	w.addSubInterface(w.homeInterface.QWidget, gcommon.Home, gallerycommon.Tr("MainWindow", "Home"), navigation.NavigationItemPositionTop)
	w.addSubInterface(w.iconInterface.QWidget, gallerycommon.IconEmojiTabSymbols, t.Icons, navigation.NavigationItemPositionTop)
	w.nav.AddSeparator(navigation.NavigationItemPositionScroll)

	pos := navigation.NavigationItemPositionScroll
	w.addSubInterface(w.basicInputInterface.QWidget, gcommon.Checkbox, t.BasicInput, pos)
	w.addSubInterface(w.dateTimeInterface.QWidget, gcommon.DateTime, t.DateTime, pos)
	w.addSubInterface(w.dialogInterface.QWidget, gcommon.Message, t.Dialogs, pos)
	w.addSubInterface(w.layoutInterface.QWidget, gcommon.Layout, t.Layout, pos)
	w.addSubInterface(w.menuInterface.QWidget, gallerycommon.IconMenu, t.Menus, pos)
	w.addSubInterface(w.navigationViewInterface.QWidget, gcommon.Menu, t.Navigation, pos)
	w.addSubInterface(w.scrollInterface.QWidget, gcommon.Scroll, t.Scroll, pos)
	w.addSubInterface(w.statusInfoInterface.QWidget, gcommon.Chat, t.StatusInfo, pos)
	w.addSubInterface(w.textInterface.QWidget, gallerycommon.IconText, t.Text, pos)
	w.addSubInterface(w.viewInterface.QWidget, gallerycommon.IconGrid, t.View, pos)

	w.nav.AddItem("price", gallerycommon.IconPrice, t.Price, func(bool) { w.onSupport() }, false, navigation.NavigationItemPositionBottom, t.Price, "")
	w.addSubInterface(w.settingInterface.QWidget, gcommon.Setting, gallerycommon.Tr("MainWindow", "Settings"), navigation.NavigationItemPositionBottom)
}

func (w *MainWindow) onSupport() {
	lang := gallerycommon.ConfigInstance.Get(gallerycommon.ConfigInstance.Language).(gallerycommon.Language)
	if lang.Name() == "zh_CN" {
		qt.QDesktopServices_OpenUrl(qt.NewQUrl3(gallerycommon.ZH_SUPPORT_URL))
	} else {
		qt.QDesktopServices_OpenUrl(qt.NewQUrl3(gallerycommon.EN_SUPPORT_URL))
	}
}

func (w *MainWindow) switchToSample(routeKey string, index int) {
	for _, iface := range w.interfaces {
		if iface.ObjectName() == routeKey {
			w.stackedWidget.SetCurrentWidget(iface.QWidget, false)
			iface.ScrollToCard(index)
			return
		}
	}
}
