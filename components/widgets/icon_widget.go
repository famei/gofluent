package widgets

import (
	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// IconWidget is a widget that paints an icon centered in its area.
//
// Constructors
//   - NewIconWidget(parent *qt.QWidget)
//   - NewIconWidgetIcon(icon interface{}, parent *qt.QWidget)
type IconWidget struct {
	*qt.QWidget
	_icon interface{}
}

// NewIconWidget builds an empty icon widget.
func NewIconWidget(parent *qt.QWidget) *IconWidget {
	w := &IconWidget{QWidget: qt.NewQWidget(parent)}
	w.installPaintEvent()
	return w
}

// NewIconWidgetIcon builds an icon widget from an icon source
// (*qt.QIcon, string or common.FluentIconBase).
func NewIconWidgetIcon(icon interface{}, parent *qt.QWidget) *IconWidget {
	w := &IconWidget{QWidget: qt.NewQWidget(parent)}
	w.SetIcon(icon)
	w.installPaintEvent()
	return w
}

// GetIcon returns the current icon as a *qt.QIcon.
func (w *IconWidget) GetIcon() *qt.QIcon {
	return common.ToQIcon(w._icon)
}

// SetIcon sets the icon source.
func (w *IconWidget) SetIcon(icon interface{}) {
	w._icon = icon
	w.Update()
}

// Icon returns the icon property value.
func (w *IconWidget) Icon() *qt.QIcon { return w.GetIcon() }

func (w *IconWidget) installPaintEvent() {
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
		super(event)
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing | qt.QPainter__SmoothPixmapTransform)
		rect := qt.NewQRectF5(w.Rect())
		defer rect.Delete()
		common.DrawIcon(w._icon, painter, rect)
		painter.End()
	})
}
