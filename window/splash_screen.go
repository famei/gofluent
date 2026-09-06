package window

import (
	"runtime"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	qt "github.com/mappu/miqt/qt"
)

// SplashScreen shows a centered icon with a drop shadow over its parent.
type SplashScreen struct {
	*qt.QWidget
	icon         interface{}
	iconSize     *qt.QSize
	titleBar     *qt.QWidget
	iconWidget   *widgets.IconWidget
	shadowEffect *qt.QGraphicsDropShadowEffect
	parentWidget *qt.QWidget
}

// NewSplashScreen builds a splash screen.
func NewSplashScreen(icon interface{}, parent *qt.QWidget, enableShadow bool) *SplashScreen {
	w := &SplashScreen{QWidget: qt.NewQWidget(parent), icon: icon, iconSize: qt.NewQSize2(96, 96), parentWidget: parent}
	w.titleBar = NewTitleBar(w.QWidget).QWidget
	w.iconWidget = widgets.NewIconWidgetIcon(icon, w.QWidget)
	w.shadowEffect = qt.NewQGraphicsDropShadowEffect2(w.QObject)

	w.iconWidget.SetFixedSize(w.iconSize)
	shadowColor := qt.NewQColor11(0, 0, 0, 50)
	w.shadowEffect.SetColor(shadowColor)
	shadowColor.Delete()
	w.shadowEffect.SetBlurRadius(15)
	w.shadowEffect.SetOffset2(0, 4)

	common.FluentStyleSheet(common.FluentWindow).Apply(w.titleBar, common.ThemeAuto)

	if enableShadow {
		w.iconWidget.SetGraphicsEffect(w.shadowEffect.QGraphicsEffect)
	}

	if parent != nil {
		parent.InstallEventFilter(w.QObject)
	}

	if runtime.GOOS == "darwin" {
		w.titleBar.Hide()
	}

	w.installEvents()
	return w
}

func (w *SplashScreen) installEvents() {
	w.OnEventFilter(func(super func(watched *qt.QObject, event *qt.QEvent) bool, watched *qt.QObject, event *qt.QEvent) bool {
		if w.parentWidget != nil && watched.UnsafePointer() == w.parentWidget.QObject.UnsafePointer() {
			switch event.Type() {
			case qt.QEvent__Resize:
				// The parent is resizing; track its current size (the Go port
				// approximates QResizeEvent.size() with parentWidget.Size()).
				sz := w.parentWidget.Size()
				w.Resize(sz.Width(), sz.Height())

			case qt.QEvent__ChildAdded:
				w.Raise()
			}
		}
		return super(watched, event)
	})
	w.OnResizeEvent(func(super func(e *qt.QResizeEvent), e *qt.QResizeEvent) {
		sz := w.IconSize()
		iw, ih := sz.Width(), sz.Height()
		w.iconWidget.Move(w.Width()/2-iw/2, w.Height()/2-ih/2)
		w.titleBar.Resize(w.Width(), w.titleBar.Height())
	})
	w.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		painter.SetPenWithStyle(qt.NoPen)
		c := 255
		if common.IsDarkTheme() {
			c = 32
		}
		bg := qt.NewQColor3(c, c, c)
		brush := qt.NewQBrush3(bg)
		painter.SetBrush(brush)
		brush.Delete()
		bg.Delete()
		rect := w.Rect()
		painter.DrawRect2(rect.X(), rect.Y(), rect.Width(), rect.Height())

		painter.End()
	})
}

// SetIcon sets the splash icon.
func (w *SplashScreen) SetIcon(icon interface{}) {
	w.icon = icon
	w.Update()
}

// Icon returns the splash icon as a QIcon.
func (w *SplashScreen) Icon() *qt.QIcon { return common.ToQIcon(w.icon) }

// SetIconSize sets the icon size.
func (w *SplashScreen) SetIconSize(size *qt.QSize) {
	w.iconSize = size
	w.iconWidget.SetFixedSize(size)
	w.Update()
}

// IconSize returns the icon size.
func (w *SplashScreen) IconSize() *qt.QSize { return w.iconSize }

// SetTitleBar replaces the title bar.
func (w *SplashScreen) SetTitleBar(titleBar *qt.QWidget) {
	if w.titleBar != nil {
		w.titleBar.DeleteLater()
	}
	w.titleBar = titleBar
	titleBar.SetParent(w.QWidget)
	titleBar.Raise()
	w.titleBar.Resize(w.Width(), w.titleBar.Height())
}

// Finish closes the splash screen.
func (w *SplashScreen) Finish() { w.Close() }
