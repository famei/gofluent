package date_time

import (
	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// monthNamesShort are the untranslated month source strings. They are passed to
// QCoreApplication_Translate with the MonthScrollView context, so the embedded
// .qm maps them to 一月..十二月 under zh_CN (mirrors the Python self.tr('Jan')).
var monthNamesShort = []string{
	"Jan", "Feb", "Mar", "Apr", "May", "Jun",
	"Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
}

// weekDayNames are the untranslated week-day source strings (DayScrollView
// context; mapped to 一..日 under zh_CN).
var weekDayNames = []string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}

// translatedMonthName returns the localized month name for 1..12.
func translatedMonthName(m int) string {
	return qt.QCoreApplication_Translate("MonthScrollView", monthNamesShort[m-1])
}

// translatedWeekDay returns the localized week-day label for 0..6.
func translatedWeekDay(i int) string {
	return qt.QCoreApplication_Translate("DayScrollView", weekDayNames[i])
}

// gridKind identifies the three stacked calendar grids.
type gridKind int

const (
	gridDay gridKind = iota
	gridMonth
	gridYear
)

// cellGranularity controls how the "today" accent is matched against a cell's
// date (mirrors the Python delegate which receives a granularity-adjusted
// current date: QDate(y,1,1) for years, QDate(y,m,1) for months, exact day).
type cellGranularity int

const (
	cellDay cellGranularity = iota
	cellMonth
	cellYear
)

// Calendar page geometry (mirrors the fast reference: 7x6 day cells of 44px and
// 4x4 month/year cells of 76px). Each grid is pre-filled with three stacked
// pages (previous / current / next) so the adjacent page is already on screen
// while the smooth scroll slides — the reference ScrollViewBase scrolls a fully
// pre-populated list through its smooth scrollbar instead of repopulating at
// the animation midpoint.
const (
	dayGridCols  = 7
	dayGridRows  = 6
	dayGridCell  = 44
	dayPageCells = dayGridCols * dayGridRows // 42

	monthGridCols  = 4
	monthGridRows  = 4
	monthGridCell  = 76
	monthPageCells = monthGridCols * monthGridRows // 16

	yearGridCols  = 4
	yearGridRows  = 4
	yearGridCell  = 76
	yearPageCells = yearGridCols * yearGridRows // 16
)

// ---------------------------------------------------------------------------
// ScrollButton

// ScrollButton is the calendar header navigation button (reset / up / down). It
// paints a transparent rounded background with hover/pressed feedback and draws
// its icon with the Fluent gray fill (#5e5e5e light / #9c9c9c dark, 10x10,
// 9x9 when pressed), mirroring the Python ScrollButton.
type ScrollButton struct {
	*qt.QToolButton
	icon      common.FluentIconBase
	isPressed bool
}

// NewScrollButton builds a calendar scroll button.
func NewScrollButton(icon common.FluentIconBase, parent *qt.QWidget) *ScrollButton {
	b := &ScrollButton{QToolButton: qt.NewQToolButton(parent), icon: icon}
	b.SetObjectName("scrollButton")
	b.SetStyleSheet(scrollButtonStyleSheet())
	b.installEvents()
	return b
}

func scrollButtonStyleSheet() string {
	if common.IsDarkTheme() {
		return "#scrollButton { background-color: transparent; border: none; border-radius: 4px; margin: 0; }" +
			" #scrollButton:hover { background-color: rgba(255, 255, 255, 9); }" +
			" #scrollButton:pressed { background-color: rgba(255, 255, 255, 6); }"
	}
	// TransparentToolButton (the Python ScrollButton base) uses a 5px radius in
	// the light theme and 4px in the dark theme (button.qss).
	return "#scrollButton { background-color: transparent; border: none; border-radius: 5px; margin: 0; }" +
		" #scrollButton:hover { background-color: rgba(0, 0, 0, 9); }" +
		" #scrollButton:pressed { background-color: rgba(0, 0, 0, 6); }"
}

func (b *ScrollButton) installEvents() {
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
		b.paintIcon()
	})
}

func (b *ScrollButton) paintIcon() {
	painter := qt.NewQPainter2(b.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__Antialiasing)

	size := 10
	if b.isPressed {
		size = 9
	}
	rect := qt.NewQRectF4(
		float64((b.Width()-size))/2, float64((b.Height()-size))/2,
		float64(size), float64(size))
	defer rect.Delete()

	fill := "#9c9c9c"
	if !common.IsDarkTheme() {
		fill = "#5e5e5e"
	}
	color := qt.NewQColor6(fill)
	defer color.Delete()
	b.icon.Colored(color, color).Render(painter, rect, common.ThemeAuto)
	painter.End()
}

// ---------------------------------------------------------------------------
// ScrollItemDelegate

// ScrollItemDelegate paints the calendar cells with the Fluent three-state
// circle: theme-coloured fill for today, theme-coloured ring for the selected
// day, and subtle hover / pressed highlights otherwise (mirrors the Python
// ScrollItemDelegate).
type ScrollItemDelegate struct {
	*qt.QStyledItemDelegate
	today       *qt.QDate
	selected    *qt.QDate
	min         *qt.QDate
	max         *qt.QDate
	hoveredRow  int
	pressedRow  int
	margin      int
	granularity cellGranularity
}

// NewScrollItemDelegate builds a calendar cell delegate.
func NewScrollItemDelegate(view *qt.QListWidget, margin int, granularity cellGranularity) *ScrollItemDelegate {
	d := &ScrollItemDelegate{
		QStyledItemDelegate: qt.NewQStyledItemDelegate2(view.QObject),
		today:               qt.NewQDate(),
		selected:            qt.NewQDate(),
		hoveredRow:          -1,
		pressedRow:          -1,
		margin:              margin,
		granularity:         granularity,
	}
	d.installPaint()
	return d
}

// SetToday sets the "today" date (granularity-matched by the delegate).
func (d *ScrollItemDelegate) SetToday(date *qt.QDate) { d.today = cloneDate(date) }

// SetSelected sets the selected date (nil clears the selection ring).
func (d *ScrollItemDelegate) SetSelected(date *qt.QDate) { d.selected = cloneDate(date) }

// SetRange sets the inclusive page range; cells outside it are drawn dimmed
// (mirrors the reference delegate, which fades adjacent-month cells to 60%).
func (d *ScrollItemDelegate) SetRange(min, max *qt.QDate) {
	d.min = cloneDate(min)
	d.max = cloneDate(max)
}

// SetHoveredRow sets the hovered row (-1 clears).
func (d *ScrollItemDelegate) SetHoveredRow(row int) { d.hoveredRow = row }

// SetPressedRow sets the pressed row (-1 clears).
func (d *ScrollItemDelegate) SetPressedRow(row int) { d.pressedRow = row }

func (d *ScrollItemDelegate) installPaint() {
	d.OnPaint(func(super func(painter *qt.QPainter, option *qt.QStyleOptionViewItem, index *qt.QModelIndex), painter *qt.QPainter, option *qt.QStyleOptionViewItem, index *qt.QModelIndex) {
		d.paint(painter, option, index)
	})
}

func (d *ScrollItemDelegate) isToday(date *qt.QDate) bool {
	if d.today == nil || date == nil || !date.IsValid() || !d.today.IsValid() {
		return false
	}
	switch d.granularity {
	case cellYear:
		return date.Year() == d.today.Year()
	case cellMonth:
		return date.Year() == d.today.Year() && date.Month() == d.today.Month()
	default:
		return sameDate(date, d.today)
	}
}

func (d *ScrollItemDelegate) isSelected(date *qt.QDate) bool {
	return d.selected != nil && d.selected.IsValid() && date != nil && date.IsValid() && sameDate(date, d.selected)
}

// isOutOfRange reports whether date lies outside the configured page range.
func (d *ScrollItemDelegate) isOutOfRange(date *qt.QDate) bool {
	if date == nil || !date.IsValid() {
		return false
	}
	if d.min != nil && d.min.IsValid() && date.DaysTo(d.min) > 0 {
		return true // date < min
	}
	if d.max != nil && d.max.IsValid() && d.max.DaysTo(date) > 0 {
		return true // date > max
	}
	return false
}

func (d *ScrollItemDelegate) paint(painter *qt.QPainter, option *qt.QStyleOptionViewItem, index *qt.QModelIndex) {
	if index == nil || !index.IsValid() {
		return
	}

	dateVar := index.DataWithRole(int(qt.UserRole)) // GoGC-armed — do NOT Delete
	if dateVar == nil || dateVar.IsNull() {
		return
	}
	date := dateVar.ToDate() // GoGC-armed value copy — do NOT Delete
	if date == nil || !date.IsValid() {
		return
	}

	painter.Save()
	painter.SetRenderHints(qt.QPainter__Antialiasing | qt.QPainter__TextAntialiasing)
	// The Python ScrollItemDelegate renders the day number with getFont() (14px);
	// without this the number is drawn with the inherited QListWidget font, which
	// has a different size/ascent and lands off-center in the highlight circle.
	font := common.GetFont(14, int(qt.QFont__Normal))
	painter.SetFont(font)
	font.Delete()

	rect := option.Rect()                                          // GoGC-armed — do NOT Delete
	adj := rect.Adjusted(d.margin, d.margin, -d.margin, -d.margin) // GoGC-armed — do NOT Delete

	row := index.Row()
	isToday := d.isToday(date)
	isSelected := d.isSelected(date)
	isHover := row == d.hoveredRow
	isPressed := row == d.pressedRow

	// outer ring (only for the selected day)
	if isSelected {
		c := common.ThemeColorValue()
		painter.SetPen(c)
		c.Delete()
	} else {
		painter.SetPenWithStyle(qt.NoPen)
	}

	// fill
	switch {
	case isToday:
		var c *qt.QColor
		switch {
		case isPressed:
			c = common.ThemeColorLight2.Color()
		case isHover:
			c = common.ThemeColorLight1.Color()
		default:
			c = common.ThemeColorValue()
		}
		brush := qt.NewQBrush3(c)
		painter.SetBrush(brush)
		brush.Delete()
		c.Delete()
	case isPressed || isHover:
		g := 255
		if !common.IsDarkTheme() {
			g = 0
		}
		alpha := 7
		if isHover {
			alpha = 9
		}
		c := qt.NewQColor11(g, g, g, alpha)
		brush := qt.NewQBrush3(c)
		painter.SetBrush(brush)
		brush.Delete()
		c.Delete()
	default:
		painter.SetBrushWithStyle(qt.NoBrush)
	}

	painter.DrawEllipseWithQRect(adj)

	// text
	if isToday {
		cv := 0
		if !common.IsDarkTheme() {
			cv = 255
		}
		c := qt.NewQColor3(cv, cv, cv)
		painter.SetPen(c)
		c.Delete()
	} else {
		if common.IsDarkTheme() {
			c := qt.NewQColor3(255, 255, 255)
			painter.SetPen(c)
			c.Delete()
		} else {
			c := qt.NewQColor3(0, 0, 0)
			painter.SetPen(c)
			c.Delete()
		}
		if isPressed || (d.isOutOfRange(date) && !isHover) {
			painter.SetOpacity(0.6)
		}
	}

	text := index.DataWithRole(int(qt.DisplayRole)).ToString() // GoGC-armed — do NOT Delete
	painter.DrawText6(rect, int(qt.AlignCenter), text)
	painter.Restore()
}

// ---------------------------------------------------------------------------
// CalendarView

// CalendarView is the pop-up calendar used by CalendarPicker. It shows a
// day/month/year stacked grid with title navigation, up/down paging and an
// optional reset button.
type CalendarView struct {
	*qt.QWidget
	date           *qt.QDate
	today          *qt.QDate
	isResetEnabled bool
	currentYear    int
	currentMonth   int
	activeGrid     gridKind

	stackedWidget *qt.QStackedWidget
	dayContainer  *qt.QWidget
	dayView       *qt.QListWidget
	monthView     *qt.QListWidget
	yearView      *qt.QListWidget
	dayDelegate   *ScrollItemDelegate
	monthDelegate *ScrollItemDelegate
	yearDelegate  *ScrollItemDelegate
	titleButton   *qt.QPushButton
	resetButton   *ScrollButton
	upButton      *ScrollButton
	downButton    *ScrollButton
	view          *qt.QFrame
	hBoxLayout    *qt.QHBoxLayout
	vBoxLayout    *qt.QVBoxLayout
	outerLayout   *qt.QHBoxLayout
	shadowEffect  *qt.QGraphicsDropShadowEffect

	scrollAni       *common.ProgressAnimation
	scrollAnimating bool

	OnDateChanged func(*qt.QDate)
	OnResetted    func()
}

// NewCalendarView builds a calendar view.
func NewCalendarView(parent *qt.QWidget) *CalendarView {
	v := &CalendarView{
		QWidget:      qt.NewQWidget(parent),
		date:         qt.NewQDate(),
		activeGrid:   gridDay,
		currentYear:  1970,
		currentMonth: 1,
	}
	// `view` is the rounded, solid-background card; every child lives inside it so
	// the translucent popup still shows an opaque panel (the transparent
	// background bug). The outer layout adds the shadow margins.
	v.view = qt.NewQFrame(v.QWidget)
	v.stackedWidget = qt.NewQStackedWidget(v.view.QWidget)
	v.dayContainer = qt.NewQWidget(v.view.QWidget)
	v.dayView = qt.NewQListWidget(v.dayContainer)
	v.monthView = qt.NewQListWidget(v.view.QWidget)
	v.yearView = qt.NewQListWidget(v.view.QWidget)
	v.titleButton = qt.NewQPushButton(v.view.QWidget)
	v.resetButton = NewScrollButton(common.Cancel, v.view.QWidget)
	v.upButton = NewScrollButton(common.CaretSolidUp, v.view.QWidget)
	v.downButton = NewScrollButton(common.CaretSolidDown, v.view.QWidget)

	v.dayDelegate = NewScrollItemDelegate(v.dayView, 3, cellDay)
	v.monthDelegate = NewScrollItemDelegate(v.monthView, 8, cellMonth)
	v.yearDelegate = NewScrollItemDelegate(v.yearView, 8, cellYear)

	v.hBoxLayout = qt.NewQHBoxLayout2()
	v.vBoxLayout = qt.NewQVBoxLayout(v.view.QWidget)
	v.outerLayout = qt.NewQHBoxLayout(v.QWidget)

	v.initWidget()
	return v
}

func (v *CalendarView) initWidget() {
	v.SetWindowFlags(qt.Popup | qt.FramelessWindowHint | qt.NoDropShadowWindowHint)
	v.SetAttribute(qt.WA_TranslucentBackground)
	v.SetAttribute(qt.WA_DeleteOnClose)
	v.OnDestroyed(func() {
		if v.scrollAni != nil {
			v.scrollAni.Delete()
			v.scrollAni = nil
		}
	})
	// The scrollbar range is only known once the layout has given each grid its
	// viewport size, which happens when the popup is first shown. Re-centre the
	// three stacked pages then; refresh() alone (run pre-show) may see a zero
	// range and clamp the value to the previous page.
	v.OnShowEvent(func(super func(ev *qt.QShowEvent), ev *qt.QShowEvent) {
		super(ev)
		v.resetScrollPositions()
	})

	// The inner card is the 314x355 surface (matching the Python CalendarViewBase
	// frame); the popup itself is sized by outerLayout (card + shadow margins).
	v.view.SetFixedSize2(314, 355)
	v.view.SetObjectName("view")
	v.view.SetAttribute(qt.WA_StyledBackground)

	v.configureGrid(v.dayView, dayGridCell, v.dayDelegate)
	v.configureGrid(v.monthView, monthGridCell, v.monthDelegate)
	v.configureGrid(v.yearView, yearGridCell, v.yearDelegate)

	v.buildDayHeader()
	v.stackedWidget.AddWidget(v.dayContainer)
	v.stackedWidget.AddWidget(v.monthView.QWidget)
	v.stackedWidget.AddWidget(v.yearView.QWidget)

	v.resetButton.SetVisible(false)
	v.titleButton.SetFixedHeight(34)
	v.upButton.SetFixedSize2(32, 34)
	v.downButton.SetFixedSize2(32, 34)
	v.resetButton.SetFixedSize2(32, 34)
	v.titleButton.SetObjectName("titleButton")

	v.hBoxLayout.SetContentsMargins(9, 8, 9, 8)
	v.hBoxLayout.SetSpacing(7)
	v.hBoxLayout.AddWidget3(v.titleButton.QWidget, 1, qt.AlignVCenter)
	v.hBoxLayout.AddWidget3(v.resetButton.QWidget, 0, qt.AlignVCenter)
	v.hBoxLayout.AddWidget3(v.upButton.QWidget, 0, qt.AlignVCenter)
	v.hBoxLayout.AddWidget3(v.downButton.QWidget, 0, qt.AlignVCenter)

	v.vBoxLayout.SetContentsMargins(0, 0, 0, 0)
	v.vBoxLayout.SetSpacing(0)
	v.vBoxLayout.AddLayout2(v.hBoxLayout.QLayout, 0)
	v.vBoxLayout.AddWidget(v.stackedWidget.QWidget)

	v.outerLayout.SetContentsMargins(12, 8, 12, 20)
	v.outerLayout.AddWidget(v.view.QWidget)

	v.setShadowEffect(30, 0, 8)
	common.FluentStyleSheet(common.FluentCalendarPicker).Apply(v.QWidget, common.ThemeAuto)
	v.applyViewBackground()

	cur := qt.QDate_CurrentDate() // GoGC-armed — do NOT Delete
	v.currentYear = cur.Year()
	v.currentMonth = cur.Month()
	v.today = cloneDate(cur)
	v.date = cloneDate(cur)

	v.titleButton.OnClicked(v.onTitleClicked)
	v.resetButton.OnClicked(func() {
		if v.OnResetted != nil {
			v.OnResetted()
		}
		v.Close()
	})
	v.upButton.OnClicked(v.scrollUp)
	v.downButton.OnClicked(v.scrollDown)

	v.dayView.OnItemClicked(func(item *qt.QListWidgetItem) { v.onDayClicked(item) })
	v.monthView.OnItemClicked(func(item *qt.QListWidgetItem) { v.onMonthClicked(item) })
	v.yearView.OnItemClicked(func(item *qt.QListWidgetItem) { v.onYearClicked(item) })

	v.refresh()
}

func (v *CalendarView) configureGrid(grid *qt.QListWidget, cellSize int, delegate *ScrollItemDelegate) {
	grid.SetViewMode(qt.QListView__IconMode)
	grid.SetMovement(qt.QListView__Static)
	grid.SetResizeMode(qt.QListView__Adjust)
	grid.SetUniformItemSizes(true)
	size := qt.NewQSize2(cellSize, cellSize)
	grid.SetGridSize(size)
	size.Delete()
	grid.SetSpacing(0)
	grid.SetViewportMargins(0, 0, 0, 0)
	grid.SetVerticalScrollMode(qt.QAbstractItemView__ScrollPerPixel)
	grid.SetVerticalScrollBarPolicy(qt.ScrollBarAlwaysOff)
	grid.SetHorizontalScrollBarPolicy(qt.ScrollBarAlwaysOff)
	grid.SetMouseTracking(true)
	grid.SetObjectName("scrollView")
	grid.SetStyleSheet(scrollViewStyleSheet())
	grid.SetItemDelegate(delegate.QAbstractItemDelegate)
	installGridTracking(grid, delegate)
	// Mirror ScrollViewBase.wheelEvent: the wheel pages the active grid
	// (day -> next/previous month, month -> next/previous year, year -> +/- 10
	// years) instead of scrolling the underlying QListWidget.
	grid.OnWheelEvent(func(super func(ev *qt.QWheelEvent), ev *qt.QWheelEvent) {
		v.handleWheel(ev)
	})
}

// installGridTracking wires the hover / pressed rows into the delegate so the
// cells repaint their transient states (mirrors ScrollViewBase's mouse handlers).
func installGridTracking(grid *qt.QListWidget, delegate *ScrollItemDelegate) {
	update := func() {
		if vp := grid.Viewport(); vp != nil {
			vp.Update()
		}
	}
	grid.OnItemEntered(func(item *qt.QListWidgetItem) {
		delegate.SetHoveredRow(grid.Row(item))
		update()
	})
	grid.OnItemPressed(func(item *qt.QListWidgetItem) {
		delegate.SetPressedRow(grid.Row(item))
		update()
	})
	grid.OnMouseReleaseEvent(func(super func(ev *qt.QMouseEvent), ev *qt.QMouseEvent) {
		super(ev)
		delegate.SetPressedRow(-1)
		update()
	})
	grid.OnLeaveEvent(func(super func(ev *qt.QEvent), ev *qt.QEvent) {
		super(ev)
		delegate.SetHoveredRow(-1)
		update()
	})
}

// buildDayHeader lays out the day grid inside a container together with the
// week-day label row (Mo..Su). The Python DayScrollView paints those labels in
// a top viewport margin; here they are a real widget row so the native
// QListWidget needs no custom viewport painting. The #weekDayGroup /
// #weekDayLabel selectors of calendar_picker.qss style them.
func (v *CalendarView) buildDayHeader() {
	layout := qt.NewQVBoxLayout(v.dayContainer)
	layout.SetContentsMargins(0, 0, 0, 0)
	layout.SetSpacing(0)

	header := qt.NewQWidget(v.dayContainer)
	header.SetObjectName("weekDayGroup")
	header.SetAttribute(qt.WA_StyledBackground)
	header.SetFixedHeight(38)
	// In Python the week-day labels live inside the DayScrollView viewport, so a
	// wheel over them pages the day grid too; route the header's wheel the same
	// way (the labels themselves ignore wheel events, so they bubble up here).
	header.OnWheelEvent(func(super func(ev *qt.QWheelEvent), ev *qt.QWheelEvent) {
		v.handleWheel(ev)
	})
	headerLayout := qt.NewQHBoxLayout(header)
	headerLayout.SetSpacing(0)
	headerLayout.SetContentsMargins(3, 12, 3, 12)
	for i := range weekDayNames {
		label := qt.NewQLabel5(translatedWeekDay(i), header)
		label.SetObjectName("weekDayLabel")
		label.SetAlignment(qt.AlignCenter)
		headerLayout.AddWidget3(label.QWidget, 1, qt.AlignHCenter)
	}

	layout.AddWidget(header)
	layout.AddWidget(v.dayView.QWidget)
}

// scrollViewStyleSheet returns the themed ScrollViewBase QSS used by the three
// calendar grids. The FluentCalendarPicker QSS `ScrollViewBase` class selector
// cannot match the native QListWidget, so the equivalent top separator,
// transparent background and item text colour are applied under the #scrollView
// objectName selector (mirrors calendar_picker.qss light/dark).
func scrollViewStyleSheet() string {
	if common.IsDarkTheme() {
		return "QListWidget#scrollView { border: none; padding: 0px 1px 0px 1px; border-bottom-left-radius: 8px; border-bottom-right-radius: 8px; border-top: 1px solid rgb(52, 52, 52); background-color: transparent; outline: none; }"
	}
	return "QListWidget#scrollView { border: none; padding: 0px 1px 0px 1px; border-bottom-left-radius: 8px; border-bottom-right-radius: 8px; border-top: 1px solid rgb(240, 240, 240); background-color: transparent; outline: none; }"
}

func (v *CalendarView) setShadowEffect(blurRadius float64, dx, dy float64) {
	color := qt.NewQColor11(0, 0, 0, 30)
	defer color.Delete()
	v.shadowEffect = qt.NewQGraphicsDropShadowEffect2(v.view.QObject)
	v.shadowEffect.SetBlurRadius(blurRadius)
	v.shadowEffect.SetOffset2(dx, dy)
	v.shadowEffect.SetColor(color)
	v.view.SetGraphicsEffect(nil)
	v.view.SetGraphicsEffect(v.shadowEffect.QGraphicsEffect)
}

// applyViewBackground paints the rounded solid card background on the inner
// frame. The FluentCalendarPicker QSS only matches the Python CalendarViewBase
// class name, which Go cannot register, so the equivalent background is applied
// here explicitly (mirrors calendar_picker.qss light/dark).
func (v *CalendarView) applyViewBackground() {
	if common.IsDarkTheme() {
		v.view.SetStyleSheet("#view { background-color: rgb(37, 37, 37); border: 1px solid rgba(255, 255, 255, 0.1); border-radius: 8px; }")
	} else {
		v.view.SetStyleSheet("#view { background-color: rgb(255, 255, 255); border: 1px solid rgba(0, 0, 0, 0.1); border-radius: 8px; }")
	}
}

// IsResetEnabled reports whether the reset button is enabled.
func (v *CalendarView) IsResetEnabled() bool { return v.isResetEnabled }

// SetResetEnabled toggles the reset button.
func (v *CalendarView) SetResetEnabled(isEnabled bool) {
	v.isResetEnabled = isEnabled
	v.resetButton.SetVisible(isEnabled)
}

// SetDate sets the selected date and refreshes the view.
func (v *CalendarView) SetDate(date *qt.QDate) {
	if date != nil && date.IsValid() && !date.IsNull() {
		v.date = cloneDate(date)
		v.currentYear = date.Year()
		v.currentMonth = date.Month()
	}
	v.activeGrid = gridDay
	v.refresh()
}

// Exec shows the calendar at pos. When ani is true the pop-up slides down 8px
// while fading in (mirrors the Python opacity+geometry parallel animation).
func (v *CalendarView) Exec(pos *qt.QPoint, ani bool) {
	if v.IsVisible() {
		return
	}
	v.refresh()

	rect := common.GetCurrentScreenGeometry(false)
	if rect == nil {
		rect = qt.NewQRect4(0, 0, 1920, 1080)
	}

	sh := v.SizeHint() // GoGC-armed — do NOT Delete
	w, h := sh.Width()+5, sh.Height()

	x := maxInt(rect.Left(), minInt(pos.X(), rect.Right()-w))
	y := maxInt(rect.Top(), minInt(pos.Y()-4, rect.Bottom()-h+5))

	if !ani {
		v.Move(x, y)
		v.Show()
		return
	}

	curve := qt.NewQEasingCurve3(qt.QEasingCurve__OutQuad)
	v.SetWindowOpacity(0)
	v.Move(x, y)
	v.Show()

	common.FadeWindowIn(v.QWidget, 150, curve)
	start := qt.NewQRect4(x, y-8, w, h)
	end := qt.NewQRect4(x, y, w, h)
	common.SlideGeometry(v.QWidget, start, end, 150, curve)
	curve.Delete()
	start.Delete()
	end.Delete()
}

func (v *CalendarView) onTitleClicked() {
	switch v.activeGrid {
	case gridDay:
		v.activeGrid = gridMonth
	case gridMonth:
		v.activeGrid = gridYear
	case gridYear:
		v.activeGrid = gridDay
	}
	v.refresh()
}

// handleWheel pages the active grid in response to a wheel event, mirroring
// ScrollViewBase.wheelEvent (wheel down scrolls forward, wheel up backward) and
// accepting the event so it does not bubble to the popup / parent widgets.
func (v *CalendarView) handleWheel(ev *qt.QWheelEvent) {
	y := ev.AngleDelta().Y()
	if y < 0 {
		v.scrollDown()
	} else if y > 0 {
		v.scrollUp()
	}
	ev.SetAccepted(true)
}

func (v *CalendarView) scrollUp()   { v.animateScroll(-1) }
func (v *CalendarView) scrollDown() { v.animateScroll(1) }

// advanceState moves the current page by direction (+1 forward / -1 backward),
// mirroring ScrollViewBase.scrollToPage(currentPage ± 1): the day grid pages by
// month, the month grid by year and the year grid by decade.
func (v *CalendarView) advanceState(direction int) {
	switch v.activeGrid {
	case gridDay:
		v.currentMonth += direction
		for v.currentMonth < 1 {
			v.currentMonth += 12
			v.currentYear--
		}
		for v.currentMonth > 12 {
			v.currentMonth -= 12
			v.currentYear++
		}
	case gridMonth:
		v.currentYear += direction
	case gridYear:
		v.currentYear += 10 * direction
	}
}

// activeGridWidget returns the QListWidget backing the active grid.
func (v *CalendarView) activeGridWidget() *qt.QListWidget {
	switch v.activeGrid {
	case gridDay:
		return v.dayView
	case gridMonth:
		return v.monthView
	case gridYear:
		return v.yearView
	}
	return nil
}

// monthAt returns the year/month reached by shifting the current day page by
// delta months (negative values wrap into the previous year).
func (v *CalendarView) monthAt(delta int) (year, month int) {
	month = v.currentMonth + delta
	year = v.currentYear
	for month < 1 {
		month += 12
		year--
	}
	for month > 12 {
		month -= 12
		year++
	}
	return year, month
}

// pageHeight returns the pixel height of one grid page. It is the scrollbar
// address step between the three vertically stacked pages.
func (v *CalendarView) pageHeight(kind gridKind) int {
	switch kind {
	case gridDay:
		return dayGridRows * dayGridCell
	case gridMonth:
		return monthGridRows * monthGridCell
	default:
		return yearGridRows * yearGridCell
	}
}

// resetScrollToCurrent parks a grid's scrollbar on its middle (current) page.
// The three pre-generated pages are stacked vertically, so the current page sits
// exactly one pageHeight below the top of the content.
func (v *CalendarView) resetScrollToCurrent(grid *qt.QListWidget, kind gridKind) {
	if bar := grid.VerticalScrollBar(); bar != nil {
		bar.SetValue(v.pageHeight(kind))
	}
}

// resetScrollPositions re-centres all three grids onto their middle page.
func (v *CalendarView) resetScrollPositions() {
	v.resetScrollToCurrent(v.dayView, gridDay)
	v.resetScrollToCurrent(v.monthView, gridMonth)
	v.resetScrollToCurrent(v.yearView, gridYear)
}

// animateScroll pages the active grid with a smooth slide by scrolling the
// pre-generated three-page list (previous / current / next stacked vertically).
// Because the adjacent page is already populated, both the outgoing and the
// incoming page stay visible for the whole slide — the reference
// ScrollViewBase.scrollToPage scrolls a fully populated list through its smooth
// scrollbar (300ms OutQuad) instead of repopulating at the midpoint. The
// viewport clips the scrolling content, so cells can never paint over the fixed
// week-day header row above the day grid.
func (v *CalendarView) animateScroll(direction int) {
	if v.scrollAnimating {
		return
	}

	grid := v.activeGridWidget()
	if grid == nil {
		return
	}

	// Commit the page change up front (ScrollViewBase.scrollToPage updates
	// currentPage and emits pageChanged before the smooth scroll starts), so the
	// title flips to the destination immediately while the old page slides out.
	v.advanceState(direction)
	v.updateTitle()
	v.updateDelegates()

	bar := grid.VerticalScrollBar()
	if bar == nil {
		v.refresh()
		return
	}

	pageH := v.pageHeight(v.activeGrid)
	start := pageH                 // current page
	end := pageH + direction*pageH // next page (direction +1) / previous (-1)

	if v.scrollAni != nil {
		v.scrollAni.Stop()
		v.scrollAni.Delete()
		v.scrollAni = nil
	}

	v.scrollAnimating = true
	curve := qt.NewQEasingCurve3(qt.QEasingCurve__OutQuad)
	v.scrollAni = common.AnimateFloat(float64(start), float64(end), 300, curve, func(t float64) {
		bar.SetValue(int(t))
	})
	v.scrollAni.OnFinished(func() {
		v.scrollAnimating = false
		// Rebuild the three pages around the newly current page and snap the
		// scrollbar back onto the middle page.
		v.refresh()
	})
	curve.Delete() // SetEasingCurve copies; the temporary is freed here.
}

func (v *CalendarView) onDayClicked(item *qt.QListWidgetItem) {
	date := itemData(item)
	if date == nil || !date.IsValid() {
		return
	}
	v.date = cloneDate(date)
	v.Close()
	if v.OnDateChanged != nil {
		v.OnDateChanged(cloneDate(date))
	}
}

func (v *CalendarView) onMonthClicked(item *qt.QListWidgetItem) {
	date := itemData(item)
	if date == nil || !date.IsValid() {
		return
	}
	v.currentYear = date.Year()
	v.currentMonth = date.Month()
	v.activeGrid = gridDay
	v.refresh()
}

func (v *CalendarView) onYearClicked(item *qt.QListWidgetItem) {
	date := itemData(item)
	if date == nil || !date.IsValid() {
		return
	}
	v.currentYear = date.Year()
	v.activeGrid = gridMonth
	v.refresh()
}

func (v *CalendarView) refresh() {
	v.populateDayGrid()
	v.populateMonthGrid()
	v.populateYearGrid()
	v.updateDelegates()

	// Each grid holds three stacked pages; park its scrollbar on the middle
	// (current) page so the destination month/year/decade is centred.
	v.resetScrollPositions()

	switch v.activeGrid {
	case gridDay:
		v.stackedWidget.SetCurrentWidget(v.dayContainer)
	case gridMonth:
		v.stackedWidget.SetCurrentWidget(v.monthView.QWidget)
	case gridYear:
		v.stackedWidget.SetCurrentWidget(v.yearView.QWidget)
	}
	v.updateTitle()
}

// updateTitle sets the header title for the active grid (the localized month
// name for the day view, the year for the month view and the decade range for
// the year view).
func (v *CalendarView) updateTitle() {
	switch v.activeGrid {
	case gridDay:
		v.titleButton.SetText(translatedMonthName(v.currentMonth) + " " + itoa(v.currentYear))
	case gridMonth:
		v.titleButton.SetText(itoa(v.currentYear))
	case gridYear:
		start := v.currentYear - v.currentYear%10
		v.titleButton.SetText(itoa(start) + " - " + itoa(start+9))
	}
}

// updateDelegates pushes the current "today" / "selected" dates into each grid's
// delegate. Only the day grid shows the selected ring; the month/year grids show
// the granularity-matched today accent (mirrors the Python view behaviour).
func (v *CalendarView) updateDelegates() {
	v.dayDelegate.SetToday(v.today)
	v.dayDelegate.SetSelected(v.date)
	v.monthDelegate.SetToday(v.today)
	v.monthDelegate.SetSelected(nil)
	v.yearDelegate.SetToday(v.today)
	v.yearDelegate.SetSelected(nil)

	// Page ranges drive the adjacent-cell dimming (mirrors the reference
	// delegate, which fades cells outside the current page to 60% opacity).
	monthFirst := qt.NewQDate2(v.currentYear, v.currentMonth, 1)
	monthLast := qt.NewQDate2(v.currentYear, v.currentMonth, monthFirst.DaysInMonth())
	v.dayDelegate.SetRange(monthFirst, monthLast)
	monthFirst.Delete()
	monthLast.Delete()

	yearFirst := qt.NewQDate2(v.currentYear, 1, 1)
	yearLast := qt.NewQDate2(v.currentYear, 12, 31)
	v.monthDelegate.SetRange(yearFirst, yearLast)
	yearFirst.Delete()
	yearLast.Delete()

	decadeStart := v.currentYear - v.currentYear%10
	decadeFirst := qt.NewQDate2(decadeStart, 1, 1)
	decadeLast := qt.NewQDate2(decadeStart+9, 12, 31)
	v.yearDelegate.SetRange(decadeFirst, decadeLast)
	decadeFirst.Delete()
	decadeLast.Delete()
}

func (v *CalendarView) populateDayGrid() {
	v.dayView.Clear()

	// Three stacked pages: the previous month, the current month and the next
	// month. Each page is 6 weeks (42 cells) starting from the Monday on/before
	// the 1st, so the adjacent month is already on screen while the smooth
	// scroll slides (mirrors the reference pre-populated ScrollViewBase list).
	for page := -1; page <= 1; page++ {
		year, month := v.monthAt(page)
		first := qt.NewQDate2(year, month, 1)
		offset := first.DayOfWeek() - 1       // 0 = Monday (QDate dayOfWeek: 1=Mon..7=Sun)
		left := first.AddDays(int64(-offset)) // GoGC-armed — do NOT Delete
		first.Delete()

		for i := 0; i < dayPageCells; i++ {
			date := left.AddDays(int64(i)) // GoGC-armed — do NOT Delete
			item := qt.NewQListWidgetItem2(itoa(date.Day()))
			size := qt.NewQSize2(dayGridCell, dayGridCell)
			item.SetSizeHint(size)
			size.Delete()
			item.SetTextAlignment(int(qt.AlignCenter))
			item.SetFlags(qt.ItemIsEnabled | qt.ItemIsSelectable)
			setItemDate(item, date.Year(), date.Month(), date.Day())
			v.dayView.AddItemWithItem(item)
		}
	}

	// Keep the selected day as the current row when it lies on the middle page.
	if v.date != nil && v.date.IsValid() && v.date.Year() == v.currentYear && v.date.Month() == v.currentMonth {
		first := qt.NewQDate2(v.currentYear, v.currentMonth, 1)
		offset := first.DayOfWeek() - 1
		first.Delete()
		v.dayView.SetCurrentRow(dayPageCells + offset + v.date.Day() - 1)
	}
}

func (v *CalendarView) populateMonthGrid() {
	v.monthView.Clear()
	// Three stacked pages: the previous year, the current year and the next
	// year. Each page is a 4x4 window (the 12 months of that year plus the first
	// 4 months of the following year), mirroring FastMonthScrollView._updateItems.
	for page := -1; page <= 1; page++ {
		year := v.currentYear + page
		for i := 0; i < monthPageCells; i++ {
			m := i%12 + 1
			y := year
			if i > 11 {
				y++
			}
			item := qt.NewQListWidgetItem2(translatedMonthName(m))
			size := qt.NewQSize2(monthGridCell, monthGridCell)
			item.SetSizeHint(size)
			size.Delete()
			item.SetTextAlignment(int(qt.AlignCenter))
			setItemDate(item, y, m, 1)
			v.monthView.AddItemWithItem(item)
		}
	}
}

func (v *CalendarView) populateYearGrid() {
	v.yearView.Clear()
	// Three stacked pages: the previous decade, the current decade and the next
	// decade. Each page is a 4x4 window (the decade containing the current year
	// plus the six following years), mirroring FastYearScrollView._updateItems.
	base := v.currentYear - v.currentYear%10
	for page := -1; page <= 1; page++ {
		start := base + page*10
		for y := start; y < start+yearPageCells; y++ {
			item := qt.NewQListWidgetItem2(itoa(y))
			size := qt.NewQSize2(yearGridCell, yearGridCell)
			item.SetSizeHint(size)
			size.Delete()
			item.SetTextAlignment(int(qt.AlignCenter))
			setItemDate(item, y, 1, 1)
			v.yearView.AddItemWithItem(item)
		}
	}
}

// setItemDate stores a date in the item's UserRole, releasing the temporary
// QDate and QVariant afterwards (both SetData and QVariant copy the value).
func setItemDate(item *qt.QListWidgetItem, year, month, day int) {
	date := qt.NewQDate2(year, month, day)
	variant := qt.NewQVariant20(date)
	item.SetData(int(qt.UserRole), variant)
	variant.Delete()
	date.Delete()
}

func itemData(item *qt.QListWidgetItem) *qt.QDate {
	if item == nil {
		return nil
	}
	v := item.Data(int(qt.UserRole))
	if v == nil || v.IsNull() {
		return nil
	}
	return v.ToDate()
}

// sameDate reports whether two dates share the same year/month/day.
func sameDate(a, b *qt.QDate) bool {
	if a == nil || b == nil || !a.IsValid() || !b.IsValid() {
		return false
	}
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [12]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
