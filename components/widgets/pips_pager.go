package widgets

import (
	"math"

	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// PipsScrollButtonDisplayMode enumerates when the previous/next scroll buttons
// of a pips pager are shown.
type PipsScrollButtonDisplayMode int

const (
	PipsDisplayAlways PipsScrollButtonDisplayMode = iota
	PipsDisplayOnHover
	PipsDisplayNever
)

// PipsScrollButton is the previous/next scroll button of a pips pager.
type PipsScrollButton struct {
	*qt.QToolButton
	_icon     common.FluentIcon
	isHover   bool
	isPressed bool
}

// NewPipsScrollButton builds a pips scroll button.
func NewPipsScrollButton(icon common.FluentIcon, parent *qt.QWidget) *PipsScrollButton {
	b := &PipsScrollButton{QToolButton: qt.NewQToolButton(parent), _icon: icon}
	b.SetFixedSize2(12, 12)
	b.installEvents()
	return b
}

func (b *PipsScrollButton) installEvents() {
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
	b.OnEnterEvent(func(super func(ev *qt.QEvent), ev *qt.QEvent) {
		super(ev)
		b.isHover = true
		b.Update()
	})
	b.OnLeaveEvent(func(super func(ev *qt.QEvent), ev *qt.QEvent) {
		super(ev)
		b.isHover = false
		b.Update()
	})
	// The base QToolButton paintEvent is deliberately skipped so the scroll
	// button stays transparent (the Python ScrollButton.paintEvent also draws
	// only the icon and never paints a background or border).
	b.OnPaintEvent(func(_ func(ev *qt.QPaintEvent), _ *qt.QPaintEvent) {
		painter := qt.NewQPainter2(b.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing)
		painter.SetPenWithStyle(qt.NoPen)

		var color *qt.QColor
		var opacity float64
		if common.IsDarkTheme() {
			color = qt.NewQColor3(255, 255, 255)
			if b.isHover || b.isPressed {
				opacity = 0.773
			} else {
				opacity = 0.541
			}
		} else {
			color = qt.NewQColor3(0, 0, 0)
			if b.isHover || b.isPressed {
				opacity = 0.616
			} else {
				opacity = 0.45
			}
		}
		defer color.Delete()
		painter.SetOpacity(opacity)

		var rect *qt.QRectF
		if b.isPressed {
			rect = qt.NewQRectF4(3, 3, 6, 6)
		} else {
			rect = qt.NewQRectF4(2, 2, 8, 8)
		}
		defer rect.Delete()
		renderFluentIconWithFill(b._icon, painter, rect, color.Name())
		painter.End()
	})
}

// PipsDelegate is the item delegate of a pips pager. It paints each pip as an
// antialiased vector ellipse (the Go port of PipsDelegate.paint), so the dots
// stay crisp at any device-pixel ratio instead of relying on pre-rendered
// pixmaps. The hovered/pressed rows only enlarge the corresponding dot.
type PipsDelegate struct {
	*qt.QStyledItemDelegate
	view       *qt.QAbstractItemView
	hoveredRow int
	pressedRow int
}

// NewPipsDelegate builds a pips delegate for the given view.
func NewPipsDelegate(view *qt.QAbstractItemView) *PipsDelegate {
	d := &PipsDelegate{
		QStyledItemDelegate: qt.NewQStyledItemDelegate2(view.QObject),
		view:                view,
		hoveredRow:          -1,
		pressedRow:          -1,
	}
	d.installPaint()
	return d
}

// installPaint overrides the delegate paint virtual to draw the pip dot,
// mirroring the Python PipsDelegate.paint: an antialiased filled circle whose
// radius is 3 when selected (or hovered but not pressed) and 2 otherwise.
func (d *PipsDelegate) installPaint() {
	d.OnPaint(func(_ func(painter *qt.QPainter, option *qt.QStyleOptionViewItem, index *qt.QModelIndex), painter *qt.QPainter, option *qt.QStyleOptionViewItem, index *qt.QModelIndex) {
		// The base paint is intentionally skipped, exactly like the Python
		// PipsDelegate.paint override (no background / selection chrome).
		painter.Save()
		painter.SetRenderHints(qt.QPainter__Antialiasing)
		painter.SetPenWithStyle(qt.NoPen)

		isHover := index.Row() == d.hoveredRow
		isPressed := index.Row() == d.pressedRow

		var color *qt.QColor
		if common.IsDarkTheme() {
			if isHover || isPressed {
				color = qt.NewQColor11(255, 255, 255, 197)
			} else {
				color = qt.NewQColor11(255, 255, 255, 138)
			}
		} else {
			if isHover || isPressed {
				color = qt.NewQColor11(0, 0, 0, 157)
			} else {
				color = qt.NewQColor11(0, 0, 0, 114)
			}
		}
		brush := qt.NewQBrush3(color)
		painter.SetBrush(brush)
		brush.Delete()
		color.Delete()

		r := 2
		if (option.State()&qt.QStyle__State_Selected) != 0 || (isHover && !isPressed) {
			r = 3
		}

		rect := option.Rect() // GoGC-armed — do NOT Delete
		x := rect.X() + 6 - r
		y := rect.Y() + 6 - r
		painter.DrawEllipse2(x, y, 2*r, 2*r)
		painter.Restore()
	})
}

// SetPressedRow sets the pressed row and repaints the viewport.
func (d *PipsDelegate) SetPressedRow(row int) {
	d.pressedRow = row
	if vp := d.view.Viewport(); vp != nil {
		vp.Update()
	}
}

// SetHoveredRow sets the hovered row and repaints the viewport.
func (d *PipsDelegate) SetHoveredRow(row int) {
	d.hoveredRow = row
	if vp := d.view.Viewport(); vp != nil {
		vp.Update()
	}
}

// PipsPager is a row/column of pips used to indicate the current page. The
// Python SmoothScrollBar is replaced by animating the QListWidget scrollbar to
// the reference position s*(index - visibleNumber/2) with an OutCubic easing
// (see animateScroll), so the current pip stays centered while scrolling.
//
// Constructors
//   - NewPipsPager(parent *qt.QWidget)
//   - NewPipsPagerOrientation(orientation qt.Orientation, parent *qt.QWidget)
type PipsPager struct {
	*qt.QListWidget
	orientation    qt.Orientation
	delegate       *PipsDelegate
	preButton      *PipsScrollButton
	nextButton     *PipsScrollButton
	_visibleNumber int
	isHover        bool
	_currentIndex  int

	scrollAni *common.ProgressAnimation

	previousButtonDisplayMode PipsScrollButtonDisplayMode
	nextButtonDisplayMode     PipsScrollButtonDisplayMode

	onCurrentIndexChanged func(index int)
}

// NewPipsPager builds a horizontal pips pager.
func NewPipsPager(parent *qt.QWidget) *PipsPager {
	return newPipsPager(qt.Horizontal, parent)
}

// NewPipsPagerOrientation builds a pips pager with the given orientation.
func NewPipsPagerOrientation(orientation qt.Orientation, parent *qt.QWidget) *PipsPager {
	return newPipsPager(orientation, parent)
}

func newPipsPager(orientation qt.Orientation, parent *qt.QWidget) *PipsPager {
	p := &PipsPager{QListWidget: qt.NewQListWidget(parent)}
	p.orientation = orientation
	p._visibleNumber = 5
	p.isHover = false
	p._currentIndex = 0

	p.delegate = NewPipsDelegate(p.QAbstractItemView)
	p.SetItemDelegate(p.delegate.QAbstractItemDelegate)

	p.SetMouseTracking(true)
	p.SetUniformItemSizes(true)
	p.SetGridSize(qt.NewQSize2(12, 12))
	p.SetMovement(qt.QListView__Static)
	p.SetVerticalScrollMode(qt.QAbstractItemView__ScrollPerPixel)
	p.SetHorizontalScrollMode(qt.QAbstractItemView__ScrollPerPixel)
	p.SetVerticalScrollBarPolicy(qt.ScrollBarAlwaysOff)
	p.SetHorizontalScrollBarPolicy(qt.ScrollBarAlwaysOff)
	p.SetIconSize(qt.NewQSize2(12, 12))
	common.FluentStyleSheet(common.FluentPipsPager).Apply(p.QWidget, common.ThemeAuto)

	if p.isHorizontal() {
		p.SetFlow(qt.QListView__LeftToRight)
		p.SetViewportMargins(15, 0, 15, 0)
		p.preButton = NewPipsScrollButton(common.CareLeftSolid, p.QWidget)
		p.nextButton = NewPipsScrollButton(common.CareRightSolid, p.QWidget)
		p.SetFixedHeight(12)
	} else {
		p.SetViewportMargins(0, 15, 0, 15)
		p.preButton = NewPipsScrollButton(common.CareUpSolid, p.QWidget)
		p.nextButton = NewPipsScrollButton(common.CareDownSolid, p.QWidget)
		p.SetFixedWidth(12)
	}

	p.SetPreviousButtonDisplayMode(PipsDisplayNever)
	p.SetNextButtonDisplayMode(PipsDisplayNever)
	p.preButton.SetToolTip("Previous Page")
	p.nextButton.SetToolTip("Next Page")

	p.preButton.OnClicked(func() { p.ScrollPrevious() })
	p.nextButton.OnClicked(func() { p.ScrollNext() })
	p.OnItemPressed(func(item *qt.QListWidgetItem) { p.setPressedItem(item) })
	p.OnItemEntered(func(item *qt.QListWidgetItem) { p.delegate.SetHoveredRow(p.Row(item)) })

	p.OnMouseReleaseEvent(func(super func(ev *qt.QMouseEvent), ev *qt.QMouseEvent) {
		super(ev)
		p.delegate.SetPressedRow(-1)
	})
	p.OnEnterEvent(func(super func(ev *qt.QEvent), ev *qt.QEvent) {
		super(ev)
		p.isHover = true
		p.updateScrollButtonVisibility()
	})
	p.OnLeaveEvent(func(super func(ev *qt.QEvent), ev *qt.QEvent) {
		super(ev)
		p.isHover = false
		p.delegate.SetHoveredRow(-1)
		p.updateScrollButtonVisibility()
	})
	p.OnResizeEvent(func(super func(ev *qt.QResizeEvent), ev *qt.QResizeEvent) {
		super(ev)
		w, h := p.Width(), p.Height()
		bw, bh := p.preButton.Width(), p.preButton.Height()
		if p.isHorizontal() {
			p.preButton.Move(0, h/2-bh/2)
			p.nextButton.Move(w-bw, h/2-bh/2)
		} else {
			p.preButton.Move(w/2-bw/2, 0)
			p.nextButton.Move(w/2-bw/2, h-bh)
		}
	})
	// Swallow wheel events (the Python implementation does nothing here).
	p.OnWheelEvent(func(super func(ev *qt.QWheelEvent), ev *qt.QWheelEvent) {})
	// Release the unparented scroll animation when the pager is destroyed.
	p.OnDestroyed(func() {
		if p.scrollAni != nil {
			p.scrollAni.Stop()
			p.scrollAni.Delete()
			p.scrollAni = nil
		}
	})
	return p
}

// IsHorizontal reports whether the pager is horizontal.
func (p *PipsPager) IsHorizontal() bool { return p.orientation == qt.Horizontal }

// CurrentIndex returns the current page index.
func (p *PipsPager) CurrentIndex() int { return p._currentIndex }

// OnCurrentIndexChanged registers a callback fired after the index changes.
func (p *PipsPager) OnCurrentIndexChanged(cb func(index int)) { p.onCurrentIndexChanged = cb }

// SetPageNumber sets the number of pages.
func (p *PipsPager) SetPageNumber(n int) {
	p.Clear()
	labels := make([]string, n)
	p.AddItems(labels)
	for i := 0; i < n; i++ {
		item := p.Item(i)
		item.SetData(int(qt.UserRole), qt.NewQVariant7(i+1))
		item.SetSizeHint(qt.NewQSize2(12, 12))
	}
	p._currentIndex = 0
	p.SetCurrentIndex(0)
	p.AdjustSize()
}

// GetPageNumber returns the number of pages.
func (p *PipsPager) GetPageNumber() int { return p.Count() }

// GetVisibleNumber returns the number of visible pips.
func (p *PipsPager) GetVisibleNumber() int { return p._visibleNumber }

// SetVisibleNumber sets the number of visible pips.
func (p *PipsPager) SetVisibleNumber(n int) {
	p._visibleNumber = n
	p.AdjustSize()
}

// ScrollNext scrolls to the next pip.
func (p *PipsPager) ScrollNext() { p.SetCurrentIndex(p._currentIndex + 1) }

// ScrollPrevious scrolls to the previous pip.
func (p *PipsPager) ScrollPrevious() { p.SetCurrentIndex(p._currentIndex - 1) }

// SetCurrentIndex sets and clamps the current page index.
func (p *PipsPager) SetCurrentIndex(index int) {
	if index < 0 || index >= p.Count() {
		return
	}
	p._currentIndex = index

	item := p.Item(index)
	p.ScrollToItem(item)
	p.SetCurrentItem(item)

	p.updateScrollButtonVisibility()
	if p.onCurrentIndexChanged != nil {
		p.onCurrentIndexChanged(index)
	}
}

// ScrollToItem scrolls the given item to the center position, matching the
// Python PipsPager.scrollToItem: the scrollbar is animated to
// s * (index - visibleNumber/2) so the current pip stays centered while the
// list keeps scrolling through every page. The smooth OutCubic animation is the
// Go port of SmoothScrollBar.scrollTo (the reference animates via a hidden
// SmoothScrollBar; a plain SetValue would snap instantly and read as "stuck").
func (p *PipsPager) ScrollToItem(item *qt.QListWidgetItem) {
	index := p.Row(item)
	size := item.SizeHint() // GoGC-armed — do NOT Delete
	s := size.Width()
	if !p.isHorizontal() {
		s = size.Height()
	}

	pos := s * (index - p._visibleNumber/2)

	var bar *qt.QScrollBar
	if p.isHorizontal() {
		bar = p.HorizontalScrollBar() // borrowed — do NOT Delete
	} else {
		bar = p.VerticalScrollBar() // borrowed — do NOT Delete
	}
	p.animateScroll(bar, pos)

	// clear selection
	p.ClearSelection()
	item.SetSelected(false)
}

// animateScroll smoothly scrolls bar to the target value with an OutCubic
// easing. It is the port of SmoothScrollBar.scrollTo: the target is clamped to
// the bar range and the 500ms duration is shortened for short distances.
func (p *PipsPager) animateScroll(bar *qt.QScrollBar, to int) {
	if to < bar.Minimum() {
		to = bar.Minimum()
	}
	if to > bar.Maximum() {
		to = bar.Maximum()
	}

	if p.scrollAni != nil {
		p.scrollAni.Stop()
		p.scrollAni.Delete()
		p.scrollAni = nil
	}

	from := bar.Value()
	if from == to {
		return
	}
	dv := to - from
	if dv < 0 {
		dv = -dv
	}
	duration := 500
	if dv < 50 {
		duration = 500 * dv / 70
	}

	curve := qt.NewQEasingCurve3(qt.QEasingCurve__OutCubic)
	p.scrollAni = common.AnimateFloat(float64(from), float64(to), duration, curve, func(v float64) {
		bar.SetValue(int(math.Round(v)))
	})
	curve.Delete() // SetEasingCurve copies; the temporary is freed here.
}

// SetPreviousButtonDisplayMode sets the display mode of the previous button.
func (p *PipsPager) SetPreviousButtonDisplayMode(mode PipsScrollButtonDisplayMode) {
	p.previousButtonDisplayMode = mode
	p.preButton.SetVisible(p.IsPreviousButtonVisible())
}

// SetNextButtonDisplayMode sets the display mode of the next button.
func (p *PipsPager) SetNextButtonDisplayMode(mode PipsScrollButtonDisplayMode) {
	p.nextButtonDisplayMode = mode
	p.nextButton.SetVisible(p.IsNextButtonVisible())
}

// IsPreviousButtonVisible reports whether the previous button is visible.
func (p *PipsPager) IsPreviousButtonVisible() bool {
	if p._currentIndex <= 0 || p.previousButtonDisplayMode == PipsDisplayNever {
		return false
	}
	if p.previousButtonDisplayMode == PipsDisplayOnHover {
		return p.isHover
	}
	return true
}

// IsNextButtonVisible reports whether the next button is visible.
func (p *PipsPager) IsNextButtonVisible() bool {
	if p._currentIndex >= p.Count()-1 || p.nextButtonDisplayMode == PipsDisplayNever {
		return false
	}
	if p.nextButtonDisplayMode == PipsDisplayOnHover {
		return p.isHover
	}
	return true
}

// AdjustSize fixes the pager size to show visibleNumber pips.
func (p *PipsPager) AdjustSize() {
	m := p.ViewportMargins()
	g := p.GridSize() // GoGC-armed — do NOT Delete
	if p.isHorizontal() {
		p.SetFixedWidth(p._visibleNumber*g.Width() + m.Left() + m.Right())
	} else {
		p.SetFixedHeight(p._visibleNumber*g.Height() + m.Top() + m.Bottom())
	}
}

func (p *PipsPager) isHorizontal() bool { return p.orientation == qt.Horizontal }

func (p *PipsPager) setPressedItem(item *qt.QListWidgetItem) {
	p.delegate.SetPressedRow(p.Row(item))
	p.SetCurrentIndex(p.Row(item))
}

func (p *PipsPager) updateScrollButtonVisibility() {
	p.preButton.SetVisible(p.IsPreviousButtonVisible())
	p.nextButton.SetVisible(p.IsNextButtonVisible())
}

// HorizontalPipsPager is a horizontally oriented pips pager.
type HorizontalPipsPager struct {
	*PipsPager
}

// NewHorizontalPipsPager builds a horizontal pips pager.
func NewHorizontalPipsPager(parent *qt.QWidget) *HorizontalPipsPager {
	return &HorizontalPipsPager{PipsPager: NewPipsPagerOrientation(qt.Horizontal, parent)}
}

// VerticalPipsPager is a vertically oriented pips pager.
type VerticalPipsPager struct {
	*PipsPager
}

// NewVerticalPipsPager builds a vertical pips pager.
func NewVerticalPipsPager(parent *qt.QWidget) *VerticalPipsPager {
	return &VerticalPipsPager{PipsPager: NewPipsPagerOrientation(qt.Vertical, parent)}
}
