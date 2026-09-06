package widgets

import (
	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// HorizontalSeparator is a thin horizontal line used to separate content.
type HorizontalSeparator struct {
	*qt.QWidget
}

// NewHorizontalSeparator builds a horizontal separator.
func NewHorizontalSeparator(parent *qt.QWidget) *HorizontalSeparator {
	w := &HorizontalSeparator{QWidget: qt.NewQWidget(parent)}
	w.SetFixedHeight(3)
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
		super(event)
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing)
		var c *qt.QColor
		if common.IsDarkTheme() {
			c = qt.NewQColor11(255, 255, 255, 51)
		} else {
			c = qt.NewQColor11(0, 0, 0, 22)
		}
		defer c.Delete()
		painter.SetPen(c)
		painter.DrawLine2(0, 1, w.Width(), 1)
		painter.End()
	})
	return w
}

// VerticalSeparator is a thin vertical line used to separate content.
type VerticalSeparator struct {
	*qt.QWidget
}

// NewVerticalSeparator builds a vertical separator.
func NewVerticalSeparator(parent *qt.QWidget) *VerticalSeparator {
	w := &VerticalSeparator{QWidget: qt.NewQWidget(parent)}
	w.SetFixedWidth(3)
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
		super(event)
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing)
		var c *qt.QColor
		if common.IsDarkTheme() {
			c = qt.NewQColor11(255, 255, 255, 51)
		} else {
			c = qt.NewQColor11(0, 0, 0, 22)
		}
		defer c.Delete()
		painter.SetPen(c)
		painter.DrawLine2(1, 0, 1, w.Height())
		painter.End()
	})
	return w
}
