package widgets

import (
	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// CycleScrollButton is the up/down scroll button of a CycleListWidget. It
// self-draws the chevron icon.
type CycleScrollButton struct {
	*qt.QToolButton
	_icon     common.FluentIcon
	isPressed bool
}

// NewCycleScrollButton builds a cycle scroll button.
func NewCycleScrollButton(icon common.FluentIcon, parent *qt.QWidget) *CycleScrollButton {
	b := &CycleScrollButton{QToolButton: qt.NewQToolButton(parent), _icon: icon}
	b.SetObjectName("scrollButton")
	b.SetStyleSheet(scrollButtonStyleSheet())
	b.installEvents()
	return b
}

// scrollButtonStyleSheet restores the PickerPanel list scroll button background.
// The FluentTimePicker QSS `ScrollButton` class selector cannot match the native
// QToolButton, so the equivalent solid background and 7px radius are applied
// under the #scrollButton objectName selector (mirrors time_picker.qss
// light/dark).
func scrollButtonStyleSheet() string {
	if common.IsDarkTheme() {
		return "#scrollButton { background-color: rgb(44, 44, 44); border: none; border-radius: 7px; }"
	}
	return "#scrollButton { background-color: rgb(249, 249, 249); border: none; border-radius: 7px; }"
}

func (b *CycleScrollButton) installEvents() {
	b.OnMousePressEvent(func(super func(ev *qt.QMouseEvent), ev *qt.QMouseEvent) {
		b.isPressed = true
		b.Update()
		super(ev)
	})
	b.OnMouseReleaseEvent(func(super func(ev *qt.QMouseEvent), ev *qt.QMouseEvent) {
		b.isPressed = false
		b.Update()
		super(ev)
	})
	b.OnPaintEvent(func(super func(ev *qt.QPaintEvent), ev *qt.QPaintEvent) {
		super(ev)
		painter := qt.NewQPainter2(b.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing)

		w, h := 10, 10
		if b.isPressed {
			w, h = 8, 8
		}
		rect := qt.NewQRectF4(float64((b.Width()-w))/2, float64((b.Height()-h))/2, float64(w), float64(h))
		defer rect.Delete()

		if !common.IsDarkTheme() {
			renderFluentIconWithFill(b._icon, painter, rect, "#5e5e5e")
		} else {
			renderFluentIcon(b._icon, painter, rect, common.ThemeAuto)
		}
		painter.End()
	})
}

// CycleListWidget is a list widget with up/down scroll buttons and a centered
// selected row. Columns with more items than the visible window repeat their
// items to enable circular scrolling; shorter columns are padded with disabled
// empty items so every value can still scroll to the center (mirrors the Python
// CycleListWidget).
type CycleListWidget struct {
	*qt.QListWidget
	itemSize       *qt.QSize
	align          int
	upButton       *CycleScrollButton
	downButton     *CycleScrollButton
	scrollDuration int
	originItems    []string
	visibleNumber  int

	isCycle bool // true when the column repeats its items to allow circular scrolling

	_currentIndex              int
	_scrollButtonRepeatEnabled bool
	onCurrentItemChanged       func(item *qt.QListWidgetItem)
	onScrollChanged            func()
	scrollAni                  *common.ProgressAnimation
}

// NewCycleListWidget builds a cycle list widget. align is the item text
// alignment (use int(qt.AlignCenter) for centered text).
func NewCycleListWidget(items []string, itemSize *qt.QSize, align int, parent *qt.QWidget) *CycleListWidget {
	w := &CycleListWidget{QListWidget: qt.NewQListWidget(parent)}
	// Match the reference `CycleListWidget { font: 14px }` so the list items and
	// the selected-value overlay (ItemMaskWidget) use the same font.
	common.SetFont(w.QWidget, 14, int(qt.QFont__Normal))
	w.itemSize = qt.NewQSize2(itemSize.Width(), itemSize.Height())
	w.align = align
	w.upButton = NewCycleScrollButton(common.CareUpSolid, w.QWidget)
	w.downButton = NewCycleScrollButton(common.CareDownSolid, w.QWidget)
	w.scrollDuration = 250
	w.originItems = append([]string{}, items...)
	w.visibleNumber = 9
	w._currentIndex = 0

	w.setItems(items)

	w.SetVerticalScrollMode(qt.QAbstractItemView__ScrollPerPixel)
	w.SetViewportMargins(0, 0, 0, 0)
	w.SetFixedSize2(w.itemSize.Width()+8, w.itemSize.Height()*w.visibleNumber)
	w.SetHorizontalScrollBarPolicy(qt.ScrollBarAlwaysOff)
	w.SetVerticalScrollBarPolicy(qt.ScrollBarAlwaysOff)

	w.OnItemClicked(func(item *qt.QListWidgetItem) { w.onItemClicked(item) })

	w.upButton.OnClicked(func() { w.ScrollUp() })
	w.downButton.OnClicked(func() { w.ScrollDown() })
	w.upButton.SetAutoRepeatDelay(500)
	w.upButton.SetAutoRepeatInterval(50)
	w.downButton.SetAutoRepeatDelay(500)
	w.downButton.SetAutoRepeatInterval(50)

	w.SetScrollButtonRepeatEnabled(true)
	w.setButtonsVisible(false)

	w.OnEnterEvent(func(super func(ev *qt.QEvent), ev *qt.QEvent) {
		super(ev)
		w.setButtonsVisible(true)
	})
	w.OnLeaveEvent(func(super func(ev *qt.QEvent), ev *qt.QEvent) {
		super(ev)
		w.setButtonsVisible(false)
	})
	w.OnResizeEvent(func(super func(ev *qt.QResizeEvent), ev *qt.QResizeEvent) {
		super(ev)
		w.upButton.Resize(w.Width(), 34)
		w.downButton.Resize(w.Width(), 34)
		w.downButton.Move(0, w.Height()-34)
	})
	// Replace the native wheel scrolling (mirrors Python CycleListWidget.wheelEvent,
	// which does not call the base class): the current index drives the scroll.
	w.OnWheelEvent(func(super func(ev *qt.QWheelEvent), ev *qt.QWheelEvent) {
		y := ev.AngleDelta().Y()
		if y < 0 {
			w.ScrollDown()
		} else if y > 0 {
			w.ScrollUp()
		}
		// Accept so the event does not bubble to the popup panel / parent.
		ev.SetAccepted(true)
	})

	// Lifecycle guard: the scroll animation is an unparented QVariantAnimation
	// whose per-frame callback dereferences this widget's vertical scrollbar and
	// the PickerPanel's itemMaskWidget. When the owning PickerPanel is closed it
	// has WA_DeleteOnClose, so Qt deletes the list widget while a 250ms scroll
	// may still be in flight. Stop the animation from the destroyed signal
	// (emitted before the internal scrollbar is torn down) so the callback never
	// runs against freed native objects.
	w.OnDestroyed(func() { w.stopScrollAnimation() })

	return w
}

// SetItems replaces the items of the list.
func (w *CycleListWidget) SetItems(items []string) {
	w.originItems = append([]string{}, items...)
	w.setItems(items)
}

// OnCurrentItemChanged registers a callback fired after the current item
// changes.
func (w *CycleListWidget) OnCurrentItemChanged(f func(item *qt.QListWidgetItem)) {
	w.onCurrentItemChanged = f
}

// SetOnScrollChanged registers a callback fired on every animated scroll frame
// (used by PickerPanel to repaint the translucent selected-row mask so its text
// overlay follows the scrolling columns).
func (w *CycleListWidget) SetOnScrollChanged(f func()) {
	w.onScrollChanged = f
}

// CurrentIndex returns the clamped current index.
func (w *CycleListWidget) CurrentIndex() int { return w._currentIndex }

// CurrentItem returns the current item.
func (w *CycleListWidget) CurrentItem() *qt.QListWidgetItem { return w.Item(w._currentIndex) }

// normalizeIndex clamps (non-cycle list) or wraps (cycle list) the raw index
// into the valid logical range without scrolling, mirroring the wrap math in
// CycleListWidget.setCurrentIndex.
func (w *CycleListWidget) normalizeIndex(index int) int {
	n := w.Count()
	if n == 0 {
		return 0
	}

	if !w.isCycle {
		pad := w.visibleNumber / 2
		lo := pad
		hi := pad + len(w.originItems) - 1
		if index < lo {
			return lo
		}
		if index > hi {
			return hi
		}
		return index
	}

	N := n / 2
	m := (w.visibleNumber + 1) / 2
	if index >= n-m {
		return N + index - n
	}
	if index <= m-1 {
		return N + index
	}
	return index
}

// SetCurrentIndex sets the current index (clamped for a non-cycle list, wrapped
// for a cycle list) and scrolls the item into the centered position. A circular
// wrap first recenters on the adjacent row instantly so the following animated
// step stays a single row instead of a long jump across the doubled items
// (mirrors CycleListWidget.setCurrentIndex's wrap pre-jump in the Python port).
func (w *CycleListWidget) SetCurrentIndex(index int) {
	n := w.Count()
	if n == 0 {
		return
	}

	if w.isCycle {
		N := n / 2
		m := (w.visibleNumber + 1) / 2
		if index >= n-m {
			w._currentIndex = N + index - n
			if pre := w.Item(w._currentIndex - 1); pre != nil {
				w.ScrollToItem2(pre, qt.QAbstractItemView__PositionAtCenter)
			}
			w.scrollToCurrent()
			return
		}
		if index <= m-1 {
			w._currentIndex = N + index
			if pre := w.Item(w._currentIndex + 1); pre != nil {
				w.ScrollToItem2(pre, qt.QAbstractItemView__PositionAtCenter)
			}
			w.scrollToCurrent()
			return
		}
	}

	w._currentIndex = w.normalizeIndex(index)
	w.scrollToCurrent()
}

// SetSelectedItem selects the item whose text matches text.
func (w *CycleListWidget) SetSelectedItem(text string) {
	if text == "" {
		return
	}
	items := w.FindItems(text, qt.MatchExactly)
	if len(items) == 0 {
		return
	}

	// For a repeated (cycle) list the second copy is the preferred one: it sits in
	// the middle of the doubled items, so every value can be scrolled to the
	// center band (mirrors CycleListWidget.setSelectedItem).
	idx := 0
	if len(items) >= 2 {
		idx = 1
	}

	// setSelectedItem scrolls instantly (native scrollToItem) rather than
	// animating, so resetting a column (e.g. the day column when the month
	// changes) never re-animates and never makes adjacent columns appear to
	// jitter (the round-6 animated scroll regression).
	w._currentIndex = w.normalizeIndex(w.Row(items[idx]))
	w.scrollToCurrentInstant()
}

// ScrollUp scrolls up one item.
func (w *CycleListWidget) ScrollUp() { w.SetCurrentIndex(w._currentIndex - 1) }

// ScrollDown scrolls down one item.
func (w *CycleListWidget) ScrollDown() { w.SetCurrentIndex(w._currentIndex + 1) }

// SetScrollButtonRepeatEnabled toggles auto-repeat of the scroll buttons.
func (w *CycleListWidget) SetScrollButtonRepeatEnabled(isEnabled bool) {
	if w._scrollButtonRepeatEnabled == isEnabled {
		return
	}
	w._scrollButtonRepeatEnabled = isEnabled
	w.upButton.SetAutoRepeat(isEnabled)
	w.downButton.SetAutoRepeat(isEnabled)
}

func (w *CycleListWidget) setItems(items []string) {
	w.stopScrollAnimation()
	w.Clear()

	n := len(items)
	w.isCycle = n > w.visibleNumber

	if w.isCycle {
		// Repeat the items twice to enable circular scrolling around the middle.
		w.addColumnItems(items, false)
		w.addColumnItems(items, false)
		w._currentIndex = n
		w.ScrollToItem2(w.Item(w._currentIndex-w.visibleNumber/2), qt.QAbstractItemView__PositionAtTop)
		return
	}

	// Pad with disabled empty items on both sides so the first/last real item can
	// also be scrolled to the centered band (mirrors CycleListWidget._createItems).
	pad := w.visibleNumber / 2
	w.addColumnItems(make([]string, pad), true)
	w.addColumnItems(items, false)
	w.addColumnItems(make([]string, pad), true)
	w._currentIndex = pad
}

func (w *CycleListWidget) addColumnItems(items []string, disabled bool) {
	for _, it := range items {
		w.addItem(it, disabled)
	}
}

func (w *CycleListWidget) addItem(text string, disabled bool) {
	item := qt.NewQListWidgetItem2(text)
	item.SetSizeHint(w.itemSize)
	item.SetTextAlignment(w.align | int(qt.AlignVCenter))
	if disabled {
		item.SetFlags(qt.NoItemFlags)
	}
	w.AddItemWithItem(item)
}

func (w *CycleListWidget) onItemClicked(item *qt.QListWidgetItem) {
	w.SetCurrentIndex(w.Row(item))
}

// scrollToCurrent scrolls the current item to the centered band position. Before
// the panel is shown the posted item layout has not run yet, so the scrollbar
// range is stale: ScrollToItem2 forces the layout and centers correctly. Once
// visible the scroll is animated for a smooth 250ms OutQuad glide (the same
// manual y = itemHeight*(index-visibleNumber/2) formula as the Python
// CycleListWidget.scrollToItem).
func (w *CycleListWidget) scrollToCurrent() {
	item := w.Item(w._currentIndex)
	if item == nil {
		return
	}

	if w.IsVisible() {
		target := w.itemSize.Height() * (w._currentIndex - w.visibleNumber/2)
		w.animateScroll(target)
	} else {
		w.ScrollToItem2(item, qt.QAbstractItemView__PositionAtCenter)
	}

	w.ClearSelection()
	if w.onCurrentItemChanged != nil {
		w.onCurrentItemChanged(item)
	}
}

// scrollToCurrentInstant centers the current item without animation and clears
// the selection, mirroring CycleListWidget.setSelectedItem's native
// super().scrollToItem(PositionAtCenter) call. It is used for programmatic
// selection so column resets never trigger an animated re-scroll.
func (w *CycleListWidget) scrollToCurrentInstant() {
	item := w.Item(w._currentIndex)
	if item == nil {
		return
	}
	w.ScrollToItem2(item, qt.QAbstractItemView__PositionAtCenter)
	w.ClearSelection()
	if w.onCurrentItemChanged != nil {
		w.onCurrentItemChanged(item)
	}
}

// animateScroll animates the vertical scroll bar to target, interrupting any
// in-flight scroll so rapid wheel/button input never accumulates drift.
func (w *CycleListWidget) animateScroll(target int) {
	bar := w.VerticalScrollBar()
	if bar == nil {
		return
	}
	if target < bar.Minimum() {
		target = bar.Minimum()
	}
	if target > bar.Maximum() {
		target = bar.Maximum()
	}

	if w.scrollAni != nil {
		w.scrollAni.Stop()
		w.scrollAni.Delete()
		w.scrollAni = nil
	}

	from := bar.Value()
	if from == target {
		return
	}

	curve := qt.NewQEasingCurve3(qt.QEasingCurve__OutQuad)
	w.scrollAni = common.AnimateFloat(float64(from), float64(target), 250, curve, func(v float64) {
		bar.SetValue(int(v + 0.5))
		if w.onScrollChanged != nil {
			w.onScrollChanged()
		}
	})
	curve.Delete() // SetEasingCurve copies; the temporary is freed here.
}

func (w *CycleListWidget) stopScrollAnimation() {
	if w.scrollAni != nil {
		w.scrollAni.Stop()
		w.scrollAni.Delete()
		w.scrollAni = nil
	}
}

func (w *CycleListWidget) setButtonsVisible(visible bool) {
	w.upButton.SetVisible(visible)
	w.downButton.SetVisible(visible)
}
