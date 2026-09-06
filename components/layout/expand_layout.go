package layout

import (
	qt "github.com/mappu/miqt/qt"
)

// ExpandLayout is a vertical-only layout that stacks child widgets at their
// natural heights, re-flowing them whenever the parent geometry changes.
//
// Like the Python port it also grows/shrinks its parent widget when a child
// resizes vertically (see the event filter), so an expanding setting card
// pushes the following groups down instead of drawing over them.
type ExpandLayout struct {
	*qt.QLayout
	items    []*qt.QLayoutItem
	widgets  []*qt.QWidget
	inResize bool
}

// NewExpandLayout builds an expand layout. parent may be nil for a layout that
// is attached later.
func NewExpandLayout(parent *qt.QWidget) *ExpandLayout {
	l := &ExpandLayout{QLayout: qt.NewQLayout(parent)}

	l.OnAddItem(func(item *qt.QLayoutItem) {
		l.items = append(l.items, item)
	})
	l.OnCount(func() int { return len(l.items) })
	l.OnItemAt(func(index int) *qt.QLayoutItem {
		if 0 <= index && index < len(l.items) {
			return l.items[index]
		}
		return nil
	})
	l.OnTakeAt(func(index int) *qt.QLayoutItem {
		if 0 <= index && index < len(l.items) {
			if index < len(l.widgets) {
				l.widgets = append(l.widgets[:index], l.widgets[index+1:]...)
			}
			item := l.items[index]
			l.items = append(l.items[:index], l.items[index+1:]...)
			return item
		}
		return nil
	})
	l.OnExpandingDirections(func(super func() qt.Orientation) qt.Orientation { return qt.Vertical })
	l.OnHasHeightForWidth(func(super func() bool) bool { return true })
	l.OnHeightForWidth(func(super func(width int) int, width int) int {
		rect := qt.NewQRect4(0, 0, width, 0)
		defer rect.Delete()
		return l.doLayout(rect, false)
	})
	l.OnSetGeometry(func(super func(geometry *qt.QRect), geometry *qt.QRect) {
		super(geometry)
		l.doLayout(geometry, true)
	})
	l.OnSizeHint(func() *qt.QSize { return l.minimumSize() })
	l.OnMinimumSize(func(super func() *qt.QSize) *qt.QSize { return l.minimumSize() })
	l.OnEventFilter(func(super func(watched *qt.QObject, event *qt.QEvent) bool, watched *qt.QObject, event *qt.QEvent) bool {
		if l.containsWidgetQObject(watched) && event.Type() == qt.QEvent__Resize {
			// Mirror Python ExpandLayout.eventFilter: when a child's height
			// changes (and its width does not), grow/shrink the parent widget
			// by the same delta so an expanding card pushes the following
			// groups down instead of drawing over them. The parent resize then
			// re-runs the layout through QLayout.setGeometry.
			//
			// The inResize flag breaks the re-entrant resize feedback loop that
			// otherwise ping-pongs between this parent resize and a QScrollArea
			// (setWidgetResizable) auto-resizing the same widget, recursing until
			// the C++ stack overflows (the "100% CPU then silent crash" symptom).
			re := qt.UnsafeNewQResizeEvent(event.UnsafePointer())
			old := re.OldSize() // reference into the event — do NOT Delete
			size := re.Size()   // reference into the event — do NOT Delete
			dh := size.Height() - old.Height()
			dw := size.Width() - old.Width()
			if dh != 0 && dw == 0 && !l.inResize {
				l.inResize = true
				if p := l.ParentWidget(); p != nil {
					p.Resize(p.Width(), p.Height()+dh)
				}
				l.inResize = false
			}

			// Keep child positions in sync (also covers width-only resizes).
			g := l.Geometry() // GoGC-armed (QLayout.Geometry) — do NOT Delete
			if g != nil && !l.inResize {
				l.doLayout(g, true)
			}
		}
		return super(watched, event)
	})

	return l
}

// AddWidget appends a widget to the layout and installs the layout event
// filter on it (so child resize events keep the layout in sync).
func (l *ExpandLayout) AddWidget(widget *qt.QWidget) {
	if widget == nil {
		return
	}
	for _, w := range l.widgets {
		if w.UnsafePointer() == widget.UnsafePointer() {
			return
		}
	}
	l.widgets = append(l.widgets, widget)
	widget.InstallEventFilter(l.QObject)
}

// AddItem appends a raw layout item (used when the widget is added through the
// base QLayout.addItem virtual path).
func (l *ExpandLayout) AddItem(item *qt.QLayoutItem) {
	if item != nil {
		l.items = append(l.items, item)
	}
}

// Count returns the number of layout items.
func (l *ExpandLayout) Count() int { return len(l.items) }

// ItemAt returns the item at index, or nil when out of range.
func (l *ExpandLayout) ItemAt(index int) *qt.QLayoutItem {
	if 0 <= index && index < len(l.items) {
		return l.items[index]
	}
	return nil
}

// WidgetAt returns the widget at index among the widgets added via AddWidget.
// AddWidget fills `widgets` (not `items`), so ItemAt is nil for these widgets.
func (l *ExpandLayout) WidgetAt(index int) *qt.QWidget {
	if 0 <= index && index < len(l.widgets) {
		return l.widgets[index]
	}
	return nil
}

// TakeAt removes and returns the item at index, or nil when out of range.
func (l *ExpandLayout) TakeAt(index int) *qt.QLayoutItem {
	if 0 <= index && index < len(l.items) {
		if index < len(l.widgets) {
			l.widgets = append(l.widgets[:index], l.widgets[index+1:]...)
		}
		item := l.items[index]
		l.items = append(l.items[:index], l.items[index+1:]...)
		return item
	}
	return nil
}

// MinimumSize returns the widest child width and the sum of every child height
// plus inter-item spacing and margins. A vertical stack needs the summed height
// (not the maximum) so that a QVBoxLayout parent (used by the settings example)
// sizes a SettingCardGroup to fit all of its cards; the maximum-height form left
// the group's sizeHint too short.
//
// It reads the child's current Height() (not MinimumSize().Height()) so the
// reported sizeHint always matches what doLayout actually lays out. The two used
// to disagree for Resize'd children (Height() != MinimumSize()), so a
// QScrollArea with setWidgetResizable(true) — which sizes its widget from
// sizeHint() — fought the ExpandLayout filter's ParentWidget().Resize(+dh),
// ping-ponging until the stack overflowed.
func (l *ExpandLayout) minimumSize() *qt.QSize {
	size := qt.NewQSize()
	totalH := 0
	for i, w := range l.widgets {
		size.SetWidth(maxInt(size.Width(), w.MinimumWidth()))
		if i > 0 {
			totalH += l.Spacing()
		}
		totalH += w.Height()
	}
	m := l.ContentsMargins() // GoGC-armed — do NOT Delete
	size.SetWidth(size.Width() + m.Left() + m.Right())
	size.SetHeight(totalH + m.Top() + m.Bottom())
	return size
}

// doLayout positions every visible child below the previous one and returns
// the total consumed height relative to rect.
func (l *ExpandLayout) doLayout(rect *qt.QRect, move bool) int {
	margin := l.ContentsMargins()

	x := rect.X() + margin.Left()
	y := rect.Y() + margin.Top()
	width := rect.Width() - margin.Left() - margin.Right()

	for i, w := range l.widgets {
		if w.IsHidden() {
			continue
		}
		if i > 0 {
			y += l.Spacing()
		}
		if move {
			pos := qt.NewQPoint2(x, y)
			size := qt.NewQSize2(width, w.Height())
			rect := qt.NewQRect4(pos.X(), pos.Y(), size.Width(), size.Height())
			w.SetGeometryWithGeometry(rect)
			pos.Delete()
			size.Delete()
			rect.Delete()
		}
		y += w.Height()
	}
	return y - rect.Y()
}

func (l *ExpandLayout) containsWidgetQObject(obj *qt.QObject) bool {
	for _, w := range l.widgets {
		if w.QObject.UnsafePointer() == obj.UnsafePointer() {
			return true
		}
	}
	return false
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
