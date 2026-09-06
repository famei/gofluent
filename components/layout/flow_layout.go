package layout

import (
	qt "github.com/mappu/miqt/qt"
)

// FlowLayout is a flowing layout: widgets are laid out left-to-right and wrap
// to the next row when they no longer fit the available width.
//
// When needAni is true the layout animates every reflow: each widget owns a
// QPropertyAnimation on its "geometry" property, all collected in a single
// QParallelAnimationGroup. A 80ms single-shot debounce timer coalesces the
// geometry changes caused by rapid window resizes before the group is
// restarted, mirroring PyQt-Fluent-Widgets flow_layout.py.
//
// A QObject event filter is installed on the first widget's parent so that a
// Show event re-runs the layout. This matters for isTight layouts: during
// construction the widgets are not visible yet and are skipped, so without the
// Show hook the first (and sometimes only) layout pass would leave every
// widget stacked at (0,0).
type FlowLayout struct {
	*qt.QLayout
	items             []*qt.QLayoutItem
	anis              []*qt.QPropertyAnimation
	aniGroup          *qt.QParallelAnimationGroup
	deBounceTimer     *qt.QTimer
	verticalSpacing   int
	horizontalSpacing int
	needAni           bool
	isTight           bool
	duration          int
	layoutFunc        func(*qt.QRect, bool) int
	wParent           *qt.QWidget
	isFilterInstalled bool
}

// NewFlowLayout builds a flow layout. parent may be nil.
func NewFlowLayout(parent *qt.QWidget, needAni, isTight bool) *FlowLayout {
	l := &FlowLayout{
		QLayout:           qt.NewQLayout(parent),
		verticalSpacing:   10,
		horizontalSpacing: 10,
		duration:          300,
		needAni:           needAni,
		isTight:           isTight,
	}
	l.layoutFunc = l.doLayout

	if needAni {
		l.aniGroup = qt.NewQParallelAnimationGroup2(l.QObject)
		l.deBounceTimer = qt.NewQTimer2(l.QObject)
		l.deBounceTimer.SetSingleShot(true)
		l.deBounceTimer.SetInterval(80)
		l.deBounceTimer.OnTimeout(func() {
			g := l.Geometry() // GoGC-armed — do NOT Delete
			if g != nil {
				l.layoutFunc(g, true)
			}
		})
	}

	l.OnAddItem(func(item *qt.QLayoutItem) {
		if item == nil {
			return
		}
		l.items = append(l.items, item)

		w := item.Widget()
		if w == nil {
			return
		}
		l.installEventFilter(w)
		if l.needAni {
			l.addWidgetAnimation(w)
		}
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
			item := l.items[index]
			l.items = append(l.items[:index], l.items[index+1:]...)
			if l.needAni && index < len(l.anis) {
				ani := l.anis[index]
				l.anis = append(l.anis[:index], l.anis[index+1:]...)
				l.aniGroup.RemoveAnimation(ani.QAbstractAnimation)
				ani.Stop()
				ani.DeleteLater()
			}
			return item
		}
		return nil
	})
	l.OnExpandingDirections(func(super func() qt.Orientation) qt.Orientation { return qt.Orientation(0) })
	l.OnHasHeightForWidth(func(super func() bool) bool { return true })
	l.OnHeightForWidth(func(super func(width int) int, width int) int {
		rect := qt.NewQRect4(0, 0, width, 0)
		defer rect.Delete()
		return l.layoutFunc(rect, false)
	})
	l.OnSetGeometry(func(super func(geometry *qt.QRect), geometry *qt.QRect) {
		super(geometry)
		if l.needAni {
			l.deBounceTimer.Start(80)
		} else {
			l.layoutFunc(geometry, true)
		}
	})
	l.OnSizeHint(func() *qt.QSize { return l.minimumSize() })
	l.OnMinimumSize(func(super func() *qt.QSize) *qt.QSize { return l.minimumSize() })
	l.OnEventFilter(func(super func(watched *qt.QObject, event *qt.QEvent) bool, watched *qt.QObject, event *qt.QEvent) bool {
		if l.wParent != nil && event.Type() == qt.QEvent__Show &&
			watched.UnsafePointer() == l.wParent.QObject.UnsafePointer() {
			// Mirror Python FlowLayout.eventFilter: when the managed parent
			// widget is shown, re-flow immediately so isTight layouts pick up
			// the now-visible widgets instead of leaving them at (0,0).
			g := l.Geometry() // GoGC-armed — do NOT Delete
			if g != nil {
				l.layoutFunc(g, true)
			}
		}
		return super(watched, event)
	})

	return l
}

// AddWidget appends a widget to the layout.
func (l *FlowLayout) AddWidget(w *qt.QWidget) {
	l.QLayout.AddWidget(w)
}

// InsertWidget inserts a widget at index. When the index cannot be honoured
// (the underlying Qt item is created by AddWidget), the widget is appended.
func (l *FlowLayout) InsertWidget(index int, w *qt.QWidget) {
	_ = index
	l.AddWidget(w)
}

// InsertItem inserts a raw layout item at index.
func (l *FlowLayout) InsertItem(index int, item *qt.QLayoutItem) {
	if item == nil {
		return
	}
	if index < 0 {
		index = 0
	}
	if index > len(l.items) {
		index = len(l.items)
	}
	l.items = append(l.items[:index], append([]*qt.QLayoutItem{item}, l.items[index:]...)...)
}

// AddItem appends a raw layout item.
func (l *FlowLayout) AddItem(item *qt.QLayoutItem) {
	if item != nil {
		l.items = append(l.items, item)
	}
}

// RemoveWidget removes the widget (and its layout item) without deleting it.
func (l *FlowLayout) RemoveWidget(widget *qt.QWidget) {
	if widget == nil {
		return
	}
	for i, item := range l.items {
		if item.Widget() != nil && item.Widget().UnsafePointer() == widget.UnsafePointer() {
			l.TakeAt(i)
			return
		}
	}
}

// RemoveAllWidgets removes every widget from the layout without deleting them.
func (l *FlowLayout) RemoveAllWidgets() {
	for len(l.items) > 0 {
		l.TakeAt(0)
	}
}

// TakeAllWidgets removes every widget from the layout and schedules them for
// deletion.
func (l *FlowLayout) TakeAllWidgets() {
	for len(l.items) > 0 {
		item := l.TakeAt(0)
		if item != nil {
			if w := item.Widget(); w != nil {
				w.DeleteLater()
			}
			item.Delete()
		}
	}
}

// Count returns the number of layout items.
func (l *FlowLayout) Count() int { return len(l.items) }

// ItemAt returns the item at index, or nil when out of range.
func (l *FlowLayout) ItemAt(index int) *qt.QLayoutItem {
	if 0 <= index && index < len(l.items) {
		return l.items[index]
	}
	return nil
}

// TakeAt removes and returns the item at index, or nil when out of range.
func (l *FlowLayout) TakeAt(index int) *qt.QLayoutItem {
	if 0 <= index && index < len(l.items) {
		item := l.items[index]
		l.items = append(l.items[:index], l.items[index+1:]...)
		if l.needAni && index < len(l.anis) {
			ani := l.anis[index]
			l.anis = append(l.anis[:index], l.anis[index+1:]...)
			l.aniGroup.RemoveAnimation(ani.QAbstractAnimation)
			ani.Stop()
			ani.DeleteLater()
		}
		return item
	}
	return nil
}

// SetAnimation sets the moving animation duration and resets the easing curve
// to Linear (mirrors Python setAnimation(duration, ease=QEasingCurve.Linear)).
func (l *FlowLayout) SetAnimation(duration int) {
	l.duration = duration
	ease := qt.NewQEasingCurve3(qt.QEasingCurve__Linear)
	for _, ani := range l.anis {
		ani.SetDuration(duration)
		ani.SetEasingCurve(ease)
	}
	ease.Delete()
}

// SetVerticalSpacing sets the vertical spacing between rows.
func (l *FlowLayout) SetVerticalSpacing(spacing int) { l.verticalSpacing = spacing }

// VerticalSpacing returns the vertical spacing between rows.
func (l *FlowLayout) VerticalSpacing() int { return l.verticalSpacing }

// SetHorizontalSpacing sets the horizontal spacing between widgets.
func (l *FlowLayout) SetHorizontalSpacing(spacing int) { l.horizontalSpacing = spacing }

// HorizontalSpacing returns the horizontal spacing between widgets.
func (l *FlowLayout) HorizontalSpacing() int { return l.horizontalSpacing }

// minimumSize returns the union of every item minimum size plus margins.
func (l *FlowLayout) minimumSize() *qt.QSize {
	size := qt.NewQSize()
	for _, item := range l.items {
		m := item.MinimumSize() // GoGC-armed — do NOT Delete
		size.SetWidth(maxInt(size.Width(), m.Width()))
		size.SetHeight(maxInt(size.Height(), m.Height()))
	}
	m := l.ContentsMargins() // GoGC-armed — do NOT Delete
	size.SetWidth(size.Width() + m.Left() + m.Right())
	size.SetHeight(size.Height() + m.Top() + m.Bottom())
	return size
}

// installEventFilter installs this layout as an event filter on the first
// managed widget's parent (or on the widget itself when it has no parent) so
// the Show event can re-run the layout.
func (l *FlowLayout) installEventFilter(w *qt.QWidget) {
	if l.isFilterInstalled {
		return
	}
	if w == nil {
		return
	}
	if p := w.ParentWidget(); p != nil {
		l.wParent = p
		p.InstallEventFilter(l.QObject)
	} else {
		l.wParent = w
		w.InstallEventFilter(l.QObject)
	}
	l.isFilterInstalled = true
}

// addWidgetAnimation creates and registers the geometry animation for w.
func (l *FlowLayout) addWidgetAnimation(w *qt.QWidget) {
	ani := qt.NewQPropertyAnimation4(w.QObject, []byte("geometry"), l.QObject)

	end := qt.NewQRect4(0, 0, w.Width(), w.Height())
	endValue := qt.NewQVariant31(end)
	ani.SetEndValue(endValue)
	endValue.Delete()
	end.Delete()

	ani.SetDuration(l.duration)
	ease := qt.NewQEasingCurve3(qt.QEasingCurve__Linear)
	ani.SetEasingCurve(ease)
	ease.Delete()

	l.aniGroup.AddAnimation(ani.QAbstractAnimation)
	l.anis = append(l.anis, ani)
}

// sameRect reports whether the animation's current end value equals target.
func (l *FlowLayout) sameRect(target *qt.QRect, ani *qt.QPropertyAnimation) bool {
	end := ani.EndValue() // GoGC-armed — do NOT Delete
	if end == nil {
		return false
	}
	r := end.ToRect() // GoGC-armed — do NOT Delete
	if r == nil {
		return false
	}
	return target.X() == r.X() && target.Y() == r.Y() &&
		target.Width() == r.Width() && target.Height() == r.Height()
}

// applyTarget positions item at target, either immediately or by updating the
// item's animation end value. It returns true when the animation group needs
// to be restarted.
func (l *FlowLayout) applyTarget(index int, item *qt.QLayoutItem, target *qt.QRect) bool {
	if !l.needAni {
		item.SetGeometry(target)
		return false
	}
	if index >= len(l.anis) {
		// Raw item added without an animation; fall back to immediate geometry.
		item.SetGeometry(target)
		return false
	}
	ani := l.anis[index]
	if l.sameRect(target, ani) {
		return false
	}
	ani.Stop()
	v := qt.NewQVariant31(target)
	ani.SetEndValue(v)
	v.Delete()
	return true
}

// finishAnimation restarts the animation group when any target changed.
func (l *FlowLayout) finishAnimation(restart bool) {
	if l.needAni && restart {
		l.aniGroup.Stop()
		l.aniGroup.Start()
	}
}

// doLayout positions every visible widget in flowing rows and returns the
// total consumed height relative to rect.
func (l *FlowLayout) doLayout(rect *qt.QRect, move bool) int {
	margin := l.ContentsMargins()

	x := rect.X() + margin.Left()
	y := rect.Y() + margin.Top()
	rowHeight := 0
	spaceX := l.horizontalSpacing
	spaceY := l.verticalSpacing
	aniRestart := false

	for i, item := range l.items {
		w := item.Widget()
		if w != nil && w.IsHidden() && l.isTight {
			continue
		}

		sh := item.SizeHint()
		nextX := x + sh.Width() + spaceX

		if nextX-spaceX > rect.Right()-margin.Right() && rowHeight > 0 {
			x = rect.X() + margin.Left()
			y = y + rowHeight + spaceY
			nextX = x + sh.Width() + spaceX
			rowHeight = 0
		}

		if move {
			pos := qt.NewQPoint2(x, y)
			target := qt.NewQRect4(pos.X(), pos.Y(), sh.Width(), sh.Height())
			pos.Delete()
			if l.applyTarget(i, item, target) {
				aniRestart = true
			}
			target.Delete()
		}

		x = nextX
		rowHeight = maxInt(rowHeight, sh.Height())
	}

	l.finishAnimation(aniRestart)

	return y + rowHeight + margin.Bottom() - rect.Y()
}

// AdaptiveFlowLayout is a flow layout whose items are stretched to fill each
// row, given a configurable minimum/maximum card width.
type AdaptiveFlowLayout struct {
	*FlowLayout
	widgetMinimumWidth int
	widgetMaximumWidth int
}

// NewAdaptiveFlowLayout builds an adaptive flow layout.
func NewAdaptiveFlowLayout(parent *qt.QWidget, needAni, isTight bool) *AdaptiveFlowLayout {
	l := &AdaptiveFlowLayout{
		FlowLayout:         NewFlowLayout(parent, needAni, isTight),
		widgetMinimumWidth: 200,
		widgetMaximumWidth: 0,
	}
	l.layoutFunc = l.doLayout
	return l
}

// SetWidgetMinimumWidth sets the minimum card width used to derive the number
// of cards per row.
func (l *AdaptiveFlowLayout) SetWidgetMinimumWidth(width int) { l.widgetMinimumWidth = width }

// WidgetMinimumWidth returns the minimum card width.
func (l *AdaptiveFlowLayout) WidgetMinimumWidth() int { return l.widgetMinimumWidth }

// SetWidgetMaximumWidth sets the maximum card width (0 disables the limit).
func (l *AdaptiveFlowLayout) SetWidgetMaximumWidth(width int) { l.widgetMaximumWidth = width }

// WidgetMaximumWidth returns the maximum card width.
func (l *AdaptiveFlowLayout) WidgetMaximumWidth() int { return l.widgetMaximumWidth }

// doLayout positions every visible widget in evenly sized flowing rows.
func (l *AdaptiveFlowLayout) doLayout(rect *qt.QRect, move bool) int {
	margin := l.ContentsMargins()

	spaceX := l.horizontalSpacing
	spaceY := l.verticalSpacing

	availableWidth := rect.Width() - margin.Left() - margin.Right()
	cardsPerRow := 1
	if l.widgetMinimumWidth+spaceX > 0 {
		cardsPerRow = (availableWidth + spaceX) / (l.widgetMinimumWidth + spaceX)
		if cardsPerRow < 1 {
			cardsPerRow = 1
		}
	}

	cardWidth := availableWidth
	if cardsPerRow > 1 {
		cardWidth = (availableWidth - (cardsPerRow-1)*spaceX) / cardsPerRow
	}
	if l.widgetMaximumWidth > 0 && cardWidth > l.widgetMaximumWidth {
		cardWidth = l.widgetMaximumWidth
	}

	x := rect.X() + margin.Left()
	y := rect.Y() + margin.Top()
	rowHeight := 0
	colIndex := 0
	aniRestart := false

	for i, item := range l.items {
		w := item.Widget()
		if w != nil && w.IsHidden() && l.isTight {
			continue
		}

		sh := item.SizeHint()

		nextX := x + cardWidth + spaceX
		if colIndex >= cardsPerRow && cardsPerRow > 0 {
			x = rect.X() + margin.Left()
			y = y + rowHeight + spaceY
			nextX = x + cardWidth + spaceX
			rowHeight = 0
			colIndex = 0
		}

		if move {
			pos := qt.NewQPoint2(x, y)
			target := qt.NewQRect4(pos.X(), pos.Y(), cardWidth, sh.Height())
			pos.Delete()
			if l.applyTarget(i, item, target) {
				aniRestart = true
			}
			target.Delete()
		}

		x = nextX
		colIndex++
		rowHeight = maxInt(rowHeight, sh.Height())
	}

	l.finishAnimation(aniRestart)

	return y + rowHeight + margin.Bottom() - rect.Y()
}
