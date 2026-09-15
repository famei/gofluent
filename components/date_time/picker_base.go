package date_time

import (
	"fmt"
	"strconv"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	qt "github.com/mappu/miqt/qt"
)

// ---------------------------------------------------------------------------
// Formatters

// PickerColumnFormatter converts a raw column value to its display string and
// back.
type PickerColumnFormatter interface {
	Encode(value interface{}) string
	Decode(value string) interface{}
}

// baseFormatter is the default formatter (identity string conversion).
type baseFormatter struct{}

func (baseFormatter) Encode(value interface{}) string { return fmt.Sprint(value) }
func (baseFormatter) Decode(value string) interface{} { return value }

// DigitFormatter decodes strings back to int.
type DigitFormatter struct{}

func (DigitFormatter) Encode(value interface{}) string { return fmt.Sprint(value) }
func (DigitFormatter) Decode(value string) interface{} {
	n, _ := strconv.Atoi(value)
	return n
}

// asInt coerces a decoded value to int.
func asInt(v interface{}) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case string:
		i, _ := strconv.Atoi(n)
		return i
	default:
		return 0
	}
}

func toInterfaceSlice(items []int) []interface{} {
	out := make([]interface{}, 0, len(items))
	for _, v := range items {
		out = append(out, v)
	}
	return out
}

func toStringSlice(items []string) []interface{} {
	out := make([]interface{}, 0, len(items))
	for _, v := range items {
		out = append(out, v)
	}
	return out
}

// ---------------------------------------------------------------------------
// SeparatorWidget

// SeparatorWidget is a thin 1-pixel separator used inside the picker panel.
type SeparatorWidget struct {
	*qt.QWidget
}

// NewSeparatorWidget builds a separator for the given orientation.
func NewSeparatorWidget(orient qt.Orientation, parent *qt.QWidget) *SeparatorWidget {
	w := &SeparatorWidget{QWidget: qt.NewQWidget(parent)}
	if orient == qt.Horizontal {
		w.SetFixedHeight(1)
	} else {
		w.SetFixedWidth(1)
	}
	w.SetAttribute(qt.WA_StyledBackground)
	w.SetObjectName("separatorWidget")
	common.FluentStyleSheet(common.FluentTimePicker).Apply(w.QWidget, common.ThemeAuto)
	w.applyBaseStyle()
	return w
}

// applyBaseStyle restores the separator's 1px themed background. The
// FluentTimePicker QSS `SeparatorWidget` class selector cannot match the native
// QWidget, so the equivalent colour is applied under the #separatorWidget
// objectName selector (mirrors time_picker.qss light/dark).
func (w *SeparatorWidget) applyBaseStyle() {
	if common.IsDarkTheme() {
		w.SetStyleSheet("#separatorWidget { background-color: rgb(61, 61, 61); }")
	} else {
		w.SetStyleSheet("#separatorWidget { background-color: rgb(234, 234, 234); }")
	}
}

// ---------------------------------------------------------------------------
// ItemMaskWidget

// ItemMaskWidget draws the solid selected-row band above the picker columns and
// paints the per-item text overlay on top of it so the selected value stays
// readable (mirrors ItemMaskWidget in the Python port).
type ItemMaskWidget struct {
	*qt.QWidget
	// listWidgets is a pointer to the PickerPanel's live column slice. A plain
	// []*CycleListWidget value would copy the (empty) slice header at
	// construction time and never observe the columns added later by
	// PickerPanel.AddColumn, leaving the text overlay permanently blank and the
	// selected value hidden behind the theme-coloured band.
	listWidgets          *[]*widgets.CycleListWidget
	lightBackgroundColor *qt.QColor
	darkBackgroundColor  *qt.QColor
}

// NewItemMaskWidget builds an item mask widget.
func NewItemMaskWidget(listWidgets *[]*widgets.CycleListWidget, parent *qt.QWidget) *ItemMaskWidget {
	w := &ItemMaskWidget{
		QWidget:              qt.NewQWidget(parent),
		listWidgets:          listWidgets,
		lightBackgroundColor: qt.NewQColor(),
		darkBackgroundColor:  qt.NewQColor(),
	}
	w.SetFixedHeight(37)
	// The mask is positioned over the list but painted with its own font; set the
	// same 14px font the reference applies via `ItemMaskWidget { font: 14px }` so
	// the selected value's overlay text matches the list items.
	common.SetFont(w.QWidget, 14, int(qt.QFont__Normal))
	// The mask must not swallow mouse input: wheel/click events have to reach
	// the CycleListWidget underneath.
	w.SetAttribute(qt.WA_TransparentForMouseEvents)
	common.FluentStyleSheet(common.FluentTimePicker).Apply(w.QWidget, common.ThemeAuto)
	w.installPaintEvent()
	return w
}

// SetCustomBackgroundColor sets the light/dark band colors.
func (w *ItemMaskWidget) SetCustomBackgroundColor(light, dark *qt.QColor) {
	w.lightBackgroundColor = cloneColor(light)
	w.darkBackgroundColor = cloneColor(dark)
	w.Update()
}

func (w *ItemMaskWidget) installPaintEvent() {
	w.OnPaintEvent(func(super func(event *qt.QPaintEvent), event *qt.QPaintEvent) {
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing | qt.QPainter__TextAntialiasing)

		painter.SetPenWithStyle(qt.NoPen)
		color := common.AutoFallbackThemeColor(w.lightBackgroundColor, w.darkBackgroundColor)
		brush := qt.NewQBrush3(color)
		defer brush.Delete()
		painter.SetBrush(brush)

		rect := w.Rect().Adjusted(4, 0, -3, 0) // GoGC-armed — do NOT Delete
		painter.DrawRoundedRect3(rect, 5, 5)

		// Draw the per-item text on top of the opaque band. The Python
		// ItemMaskWidget paints this overlay so the selected value stays readable
		// over the solid theme-coloured band; without it the value is hidden
		// behind the band (the round-6 "green / occluded selection" bug).
		w.paintText(painter)

		painter.End()
		super(event)
	})
}

// paintText draws the item text overlay above the selected-row band. For every
// column it paints the item crossing the band's top edge at its visual position
// and the item crossing the bottom edge, so the text follows the scroll
// animation (mirrors ItemMaskWidget.paintEvent's text pass; the text colour is
// white in light theme / black in dark theme, exactly as the Python port).
func (w *ItemMaskWidget) paintText(painter *qt.QPainter) {
	if w.listWidgets == nil || len(*w.listWidgets) == 0 {
		return
	}

	tc := 255
	if common.IsDarkTheme() {
		tc = 0
	}
	color := qt.NewQColor3(tc, tc, tc)
	painter.SetPen(color)
	color.Delete()

	// QWidget.Font() is a borrowed reference in miqt (no GoGC); do NOT Delete it.
	font := w.Font()
	painter.SetFont(font)

	// Port of ItemMaskWidget.paintEvent's text pass: use ItemAt + VisualItemRect
	// directly (no MapTo), exactly like the Python reference. MapTo across the
	// panel/view parent boundary segfaults on a dangling QWidget, so avoid it.
	h := w.Height()
	acc := 0
	for _, list := range *w.listWidgets {
		if list == nil {
			continue
		}
		painter.Save()

		// item center x = itemSize.width()/2 + 4 == list.Width()/2 (the fixed
		// width is itemSize.width()+8).
		x := list.Width()/2 + w.X()
		p := qt.NewQPoint2(x, w.Y()+6)
		item1 := list.ItemAt(p)
		p.Delete()
		if item1 == nil {
			painter.Restore()
			continue
		}

		iw := item1.SizeHint().Width()       // GoGC-armed — do NOT Delete
		vy := list.VisualItemRect(item1).Y() // GoGC-armed — do NOT Delete
		painter.Translate2(float64(acc), float64(vy-w.Y()+7))
		w.drawItemText(painter, item1, 0)

		p = qt.NewQPoint2(w.X()+x, w.Y()+h-6)
		item2 := list.ItemAt(p)
		p.Delete()
		if item2 != nil {
			w.drawItemText(painter, item2, h)
		}

		painter.Restore()
		acc += iw + 8
	}
}

// drawItemText draws item's text at the given y offset using the painter's
// current translation (mirrors ItemMaskWidget._drawText).
func (w *ItemMaskWidget) drawItemText(painter *qt.QPainter, item *qt.QListWidgetItem, y int) {
	flags := item.TextAlignment()
	size := item.SizeHint() // GoGC-armed — do NOT Delete
	iw, ih := size.Width(), size.Height()

	var dx, dw int
	if flags&int(qt.AlignLeft) != 0 {
		dx, dw = 15, iw
	} else if flags&int(qt.AlignRight) != 0 {
		dx, dw = 4, iw-15
	} else {
		dx, dw = 4, iw
	}
	painter.DrawText7(dx, y, dw, ih, flags, item.Text())
}

// ---------------------------------------------------------------------------
// PickerColumnButton

// PickerColumnButton is one column of a picker (rendered as a transparent
// button that only shows the current formatted value).
type PickerColumnButton struct {
	*qt.QPushButton
	name      string
	value     interface{}
	items     []interface{}
	align     int
	formatter PickerColumnFormatter
}

// NewPickerColumnButton builds a picker column button.
func NewPickerColumnButton(name string, items []interface{}, width int, align int, formatter PickerColumnFormatter, parent *qt.QWidget) *PickerColumnButton {
	if formatter == nil {
		formatter = baseFormatter{}
	}
	b := &PickerColumnButton{
		QPushButton: qt.NewQPushButton5(name, parent),
		name:        name,
		align:       align,
		formatter:   formatter,
	}
	b.SetItems(items)
	b.SetAlignment(align)
	b.SetFixedSize2(width, 30)
	b.SetObjectName("pickerButton")
	b.SetProperty("hasBorder", qt.NewQVariant11(false))
	b.SetAttribute(qt.WA_TransparentForMouseEvents)
	return b
}

// Align returns the text alignment.
func (b *PickerColumnButton) Align() int { return b.align }

// SetAlignment sets the text alignment and updates the "align" property.
func (b *PickerColumnButton) SetAlignment(align int) {
	b.align = align
	a := "center"
	if align == int(qt.AlignLeft) {
		a = "left"
	} else if align == int(qt.AlignRight) {
		a = "right"
	}
	b.SetProperty("align", qt.NewQVariant14(a))
	b.SetStyle(qt.QApplication_Style())
}

// Value returns the formatted value ("" when no value is set).
func (b *PickerColumnButton) Value() string {
	if b.value == nil {
		return ""
	}
	return b.formatter.Encode(b.value)
}

// SetValue sets the raw value and refreshes the text.
func (b *PickerColumnButton) SetValue(v interface{}) {
	b.value = v
	if v == nil {
		b.SetText(b.name)
		b.SetProperty("hasValue", qt.NewQVariant11(false))
	} else {
		b.SetText(b.Value())
		b.SetProperty("hasValue", qt.NewQVariant11(true))
	}
	b.SetStyle(qt.QApplication_Style())
}

// Items returns the formatted item strings.
func (b *PickerColumnButton) Items() []string {
	out := make([]string, 0, len(b.items))
	for _, it := range b.items {
		out = append(out, b.formatter.Encode(it))
	}
	return out
}

// SetItems replaces the raw items.
func (b *PickerColumnButton) SetItems(items []interface{}) {
	b.items = append([]interface{}(nil), items...)
}

// Formatter returns the column formatter.
func (b *PickerColumnButton) Formatter() PickerColumnFormatter { return b.formatter }

// SetFormatter replaces the column formatter.
func (b *PickerColumnButton) SetFormatter(f PickerColumnFormatter) {
	if f == nil {
		f = baseFormatter{}
	}
	b.formatter = f
}

// Name returns the column name.
func (b *PickerColumnButton) Name() string { return b.name }

// SetName sets the column name (and refreshes the placeholder text).
func (b *PickerColumnButton) SetName(name string) {
	if b.Text() == b.name {
		b.SetText(name)
	}
	b.name = name
}

// InitialValue returns the current value (kept for API parity).
func (b *PickerColumnButton) InitialValue() string { return b.Value() }

// SetInitialValue sets the value (kept for API parity).
func (b *PickerColumnButton) SetInitialValue(v interface{}) { b.SetValue(v) }

// ---------------------------------------------------------------------------
// PickerBase

// PickerBase is the base class of the date/time pickers. It renders a set of
// column buttons and opens a PickerPanel pop-up when clicked.
type PickerBase struct {
	*qt.QPushButton
	columns                      []*PickerColumnButton
	lightSelectedBackgroundColor *qt.QColor
	darkSelectedBackgroundColor  *qt.QColor
	isResetEnabled               bool
	hBoxLayout                   *qt.QHBoxLayout
	isScrollButtonRepeatEnabled  bool

	// Overridable hooks (Go has no virtual dispatch; subclasses replace these).
	panelInitValueFunc func() []string
	confirmValuesFunc  func([]string)
	columnChangedFunc  func(panel *PickerPanel, index int, value string)
}

// NewPickerBase builds a picker base.
func NewPickerBase(parent *qt.QWidget) *PickerBase {
	b := &PickerBase{
		QPushButton:                  qt.NewQPushButton(parent),
		lightSelectedBackgroundColor: qt.NewQColor(),
		darkSelectedBackgroundColor:  qt.NewQColor(),
		isScrollButtonRepeatEnabled:  true,
	}
	b.hBoxLayout = qt.NewQHBoxLayout(b.QWidget)
	b.hBoxLayout.SetSpacing(0)
	b.hBoxLayout.SetContentsMargins(0, 0, 0, 0)
	// The Python port sets SetFixedSize; without it the picker button collapses
	// to a zero/native size and the column buttons are not laid out horizontally.
	b.hBoxLayout.SetSizeConstraint(qt.QLayout__SetFixedSize)

	b.panelInitValueFunc = b.panelInitialValue
	b.confirmValuesFunc = b.confirmValues
	b.columnChangedFunc = func(panel *PickerPanel, index int, value string) {}

	// The PickerBase class selector in time_picker.qss is rewritten to
	// QPushButton#pickerBase (selectorTranslation), so the objectName must be set
	// before the stylesheet is applied. It is applied via the registered
	// FluentTimePicker source, so it follows dark/light theme switches.
	b.SetObjectName("pickerBase")
	common.FluentStyleSheet(common.FluentTimePicker).Apply(b.QWidget, common.ThemeAuto)
	b.OnClicked(b.showPanel)
	return b
}

// SetSelectedBackgroundColor sets the panel selected-row colors.
func (b *PickerBase) SetSelectedBackgroundColor(light, dark *qt.QColor) {
	b.lightSelectedBackgroundColor = cloneColor(light)
	b.darkSelectedBackgroundColor = cloneColor(dark)
}

// AddColumn appends a column to the picker.
func (b *PickerBase) AddColumn(name string, items []interface{}, width int, align int, formatter PickerColumnFormatter) {
	button := NewPickerColumnButton(name, items, width, align, formatter, b.QWidget)
	b.columns = append(b.columns, button)
	b.hBoxLayout.AddWidget3(button.QWidget, 0, qt.AlignLeft)

	for _, btn := range b.columns[:maxInt(0, len(b.columns)-1)] {
		btn.SetProperty("hasBorder", qt.NewQVariant11(true))
		btn.SetStyle(qt.QApplication_Style())
	}
}

// SetColumnAlignment sets the alignment of a column.
func (b *PickerBase) SetColumnAlignment(index, align int) {
	if 0 <= index && index < len(b.columns) {
		b.columns[index].SetAlignment(align)
	}
}

// SetColumnWidth sets the width of a column.
func (b *PickerBase) SetColumnWidth(index, width int) {
	if 0 <= index && index < len(b.columns) {
		b.columns[index].SetFixedWidth(width)
	}
}

// SetColumnTight makes a column as narrow as its widest item.
func (b *PickerBase) SetColumnTight(index int) {
	if index < 0 || index >= len(b.columns) {
		return
	}
	fm := b.FontMetrics()
	w := 0
	for _, it := range b.columns[index].Items() {
		w = maxInt(w, fm.Width(it))
	}
	b.SetColumnWidth(index, w+30)
}

// SetColumnVisible sets the visibility of a column.
func (b *PickerBase) SetColumnVisible(index int, isVisible bool) {
	if 0 <= index && index < len(b.columns) {
		b.columns[index].SetVisible(isVisible)
	}
}

// Value returns the formatted values of the visible columns.
func (b *PickerBase) Value() []string {
	out := make([]string, 0, len(b.columns))
	for _, c := range b.columns {
		if c.IsVisible() {
			out = append(out, c.Value())
		}
	}
	return out
}

// SetColumnValue sets the raw value of a column.
func (b *PickerBase) SetColumnValue(index int, value interface{}) {
	if 0 <= index && index < len(b.columns) {
		b.columns[index].SetValue(value)
	}
}

// SetColumnFormatter replaces the formatter of a column.
func (b *PickerBase) SetColumnFormatter(index int, formatter PickerColumnFormatter) {
	if 0 <= index && index < len(b.columns) {
		b.columns[index].SetFormatter(formatter)
	}
}

// SetColumnItems replaces the items of a column.
func (b *PickerBase) SetColumnItems(index int, items []interface{}) {
	if 0 <= index && index < len(b.columns) {
		b.columns[index].SetItems(items)
	}
}

// EncodeValue formats a raw value using the column formatter.
func (b *PickerBase) EncodeValue(index int, value interface{}) string {
	if 0 <= index && index < len(b.columns) {
		return b.columns[index].Formatter().Encode(value)
	}
	return ""
}

// DecodeValue parses a formatted value using the column formatter.
func (b *PickerBase) DecodeValue(index int, value string) interface{} {
	if 0 <= index && index < len(b.columns) {
		return b.columns[index].Formatter().Decode(value)
	}
	return value
}

// SetColumn replaces the name, width and alignment of a column.
func (b *PickerBase) SetColumn(index int, name string, items []interface{}, width, align int) {
	if index < 0 || index >= len(b.columns) {
		return
	}
	button := b.columns[index]
	button.SetText(name)
	button.SetFixedWidth(width)
	button.SetAlignment(align)
	button.SetItems(items)
}

// ClearColumns removes and deletes every column.
func (b *PickerBase) ClearColumns() {
	for len(b.columns) > 0 {
		btn := b.columns[len(b.columns)-1]
		b.columns = b.columns[:len(b.columns)-1]
		b.hBoxLayout.RemoveWidget(btn.QWidget)
		btn.SetParent(nil)
		btn.DeleteLater()
	}
}

// SetScrollButtonRepeatEnabled toggles the panel scroll button auto-repeat.
func (b *PickerBase) SetScrollButtonRepeatEnabled(isEnabled bool) {
	b.isScrollButtonRepeatEnabled = isEnabled
}

// IsResetEnabled reports whether the reset button is enabled.
func (b *PickerBase) IsResetEnabled() bool { return b.isResetEnabled }

// SetResetEnabled toggles the reset button.
func (b *PickerBase) SetResetEnabled(isEnabled bool) { b.isResetEnabled = isEnabled }

// Reset clears every column value.
func (b *PickerBase) Reset() {
	for i := range b.columns {
		b.SetColumnValue(i, nil)
	}
}

// PanelInitialValue returns the values used to initialise the panel.
func (b *PickerBase) PanelInitialValue() []string { return b.Value() }

// OnColumnValueChanged is the hook called when a panel column changes.
func (b *PickerBase) OnColumnValueChanged(panel *PickerPanel, index int, value string) {}

func (b *PickerBase) panelInitialValue() []string { return b.Value() }

func (b *PickerBase) showPanel() {
	panel := NewPickerPanel(b.QWidget)
	for _, column := range b.columns {
		if column.IsVisible() {
			panel.AddColumn(column.Items(), column.Width(), column.Align())
		}
	}

	panel.SetValue(b.panelInitValueFunc())
	panel.SetResetEnabled(b.isResetEnabled)
	panel.SetScrollButtonRepeatEnabled(b.isScrollButtonRepeatEnabled)
	panel.SetSelectedBackgroundColor(b.lightSelectedBackgroundColor, b.darkSelectedBackgroundColor)

	panel.OnConfirmed = b.confirmValuesFunc
	panel.OnResetted = func() { b.Reset() }
	panel.OnColumnValueChanged = func(index int, value string) { b.columnChangedFunc(panel, index, value) }

	sh := panel.vBoxLayout.SizeHint()
	w := sh.Width() - b.Width()
	p := qt.NewQPoint2(-w/2, -37*4)
	pos := b.MapToGlobal(p)
	p.Delete()
	panel.Exec(pos, true)
}

func (b *PickerBase) confirmValues(values []string) {
	for i, v := range values {
		b.SetColumnValue(i, v)
	}
}

// ---------------------------------------------------------------------------
// PickerPanel

// PickerPanel is the pop-up list panel used by a picker.
type PickerPanel struct {
	*qt.QWidget
	itemHeight                int
	listWidgets               []*widgets.CycleListWidget
	view                      *qt.QFrame
	itemMaskWidget            *ItemMaskWidget
	hSeparatorWidget          *SeparatorWidget
	yesButton                 *widgets.TransparentToolButton
	resetButton               *widgets.TransparentToolButton
	cancelButton              *widgets.TransparentToolButton
	hBoxLayout                *qt.QHBoxLayout
	listLayout                *qt.QHBoxLayout
	buttonLayout              *qt.QHBoxLayout
	vBoxLayout                *qt.QVBoxLayout
	scrollButtonRepeatEnabled bool
	shadowEffect              *qt.QGraphicsDropShadowEffect
	isExpanded                bool
	ani                       *common.ProgressAnimation

	OnConfirmed          func([]string)
	OnResetted           func()
	OnColumnValueChanged func(int, string)
}

// NewPickerPanel builds a picker panel.
func NewPickerPanel(parent *qt.QWidget) *PickerPanel {
	p := &PickerPanel{
		QWidget:                   qt.NewQWidget(parent),
		itemHeight:                37,
		scrollButtonRepeatEnabled: true,
	}
	p.view = qt.NewQFrame(p.QWidget)
	p.itemMaskWidget = NewItemMaskWidget(&p.listWidgets, p.QWidget)
	p.hSeparatorWidget = NewSeparatorWidget(qt.Horizontal, p.view.QWidget)
	p.yesButton = widgets.NewTransparentToolButtonIcon(common.Accept, p.view.QWidget)
	p.resetButton = widgets.NewTransparentToolButtonIcon(common.Cancel, p.view.QWidget)
	p.cancelButton = widgets.NewTransparentToolButtonIcon(common.Cancel, p.view.QWidget)

	p.hBoxLayout = qt.NewQHBoxLayout(p.QWidget)
	p.listLayout = qt.NewQHBoxLayout2()
	p.buttonLayout = qt.NewQHBoxLayout2()
	p.vBoxLayout = qt.NewQVBoxLayout(p.view.QWidget)

	p.initWidget()
	return p
}

func (p *PickerPanel) initWidget() {
	p.SetWindowFlags(qt.Popup | qt.FramelessWindowHint | qt.NoDropShadowWindowHint)
	p.SetAttribute(qt.WA_TranslucentBackground)
	p.SetAttribute(qt.WA_DeleteOnClose)

	// Configure the card frame before attaching the drop shadow: WA_StyledBackground
	// + objectName must be in place first so the QGraphicsDropShadowEffect renders
	// the #view QSS background instead of the default black (mirrors
	// CalendarView.initWidget ordering).
	p.view.SetObjectName("view")
	p.view.SetAttribute(qt.WA_StyledBackground)

	p.yesButton.SetIconSize(qt.NewQSize2(16, 16))
	p.resetButton.SetIconSize(qt.NewQSize2(16, 16))
	p.cancelButton.SetIconSize(qt.NewQSize2(13, 13))
	p.yesButton.SetFixedHeight(33)
	p.cancelButton.SetFixedHeight(33)
	p.resetButton.SetFixedHeight(33)

	p.hBoxLayout.SetContentsMargins(12, 8, 12, 20)
	p.hBoxLayout.SetSizeConstraint(qt.QLayout__SetMinimumSize)
	p.hBoxLayout.AddWidget3(p.view.QWidget, 1, qt.AlignCenter)

	p.vBoxLayout.SetSpacing(0)
	p.vBoxLayout.SetContentsMargins(0, 0, 0, 0)
	p.vBoxLayout.AddLayout2(p.listLayout.QLayout, 1)
	p.vBoxLayout.AddWidget(p.hSeparatorWidget.QWidget)
	p.vBoxLayout.AddLayout2(p.buttonLayout.QLayout, 1)
	p.vBoxLayout.SetSizeConstraint(qt.QLayout__SetMinimumSize)

	p.buttonLayout.SetSpacing(6)
	p.buttonLayout.SetContentsMargins(3, 3, 3, 3)
	p.buttonLayout.AddWidget(p.yesButton.QWidget)
	p.buttonLayout.AddWidget(p.resetButton.QWidget)
	p.buttonLayout.AddWidget(p.cancelButton.QWidget)
	p.yesButton.SetSizePolicy2(qt.QSizePolicy__Expanding, qt.QSizePolicy__Expanding)
	p.resetButton.SetSizePolicy2(qt.QSizePolicy__Expanding, qt.QSizePolicy__Expanding)
	p.cancelButton.SetSizePolicy2(qt.QSizePolicy__Expanding, qt.QSizePolicy__Expanding)

	p.yesButton.OnClicked(func() { p.confirm() })
	p.cancelButton.OnClicked(func() { p.closePanel() })
	p.resetButton.OnClicked(func() {
		if p.OnResetted != nil {
			p.OnResetted()
		}
		p.closePanel()
	})

	p.SetResetEnabled(false)
	p.setShadowEffect(30, 0, 8)
	common.FluentStyleSheet(common.FluentTimePicker).Apply(p.QWidget, common.ThemeAuto)
	p.applyViewBackground()

	// The Python port repositions the translucent selected-row band (item mask)
	// on every resize; the mask is a sibling of `view`, so it must be laid out
	// manually instead of through a layout.
	p.OnResizeEvent(func(super func(e *qt.QResizeEvent), e *qt.QResizeEvent) {
		super(e)
		p.layoutMask()
	})

	// Lifecycle guard: the show/fade animation (p.ani) is an unparented
	// QVariantAnimation whose callback calls back into this widget. A Popup can
	// also be closed by an outside click without going through closePanel, in
	// which case WA_DeleteOnClose deletes the widget while the 150ms fade-in may
	// still be running. Stop the animation from the destroyed signal so its
	// callback never dereferences the freed widget.
	p.OnDestroyed(func() { p.stopAnimation() })
}

// applyViewBackground paints the rounded solid panel background on the inner
// view frame. The FluentTimePicker QSS selector `PickerPanel > #view` requires
// the Python class name which Go cannot register, so the equivalent background
// is applied explicitly (mirrors time_picker.qss light/dark).
func (p *PickerPanel) applyViewBackground() {
	if common.IsDarkTheme() {
		p.view.SetStyleSheet("#view { background-color: rgb(44, 44, 44); border: 1px solid rgba(255, 255, 255, 0.1); border-radius: 7px; }")
	} else {
		p.view.SetStyleSheet("#view { background-color: rgb(249, 249, 249); border: 1px solid rgba(0, 0, 0, 0.14); border-radius: 7px; }")
	}
}

// layoutMask positions and sizes the selected-row band over the list columns.
func (p *PickerPanel) layoutMask() {
	if p.view == nil || p.itemMaskWidget == nil {
		return
	}
	p.itemMaskWidget.Resize(p.view.Width()-3, p.itemHeight)
	m := p.hBoxLayout.ContentsMargins() // GoGC-armed — do NOT Delete
	p.itemMaskWidget.Move(m.Left()+2, m.Top()+148)
}

func (p *PickerPanel) setShadowEffect(blurRadius float64, dx, dy float64) {
	color := qt.NewQColor11(0, 0, 0, 30)
	defer color.Delete()
	p.shadowEffect = qt.NewQGraphicsDropShadowEffect2(p.view.QObject)
	p.shadowEffect.SetBlurRadius(blurRadius)
	p.shadowEffect.SetOffset2(dx, dy)
	p.shadowEffect.SetColor(color)
	p.view.SetGraphicsEffect(nil)
	p.view.SetGraphicsEffect(p.shadowEffect.QGraphicsEffect)
}

// SetResetEnabled toggles the reset button.
func (p *PickerPanel) SetResetEnabled(isEnabled bool) { p.resetButton.SetVisible(isEnabled) }

// IsResetEnabled reports whether the reset button is visible.
func (p *PickerPanel) IsResetEnabled() bool { return p.resetButton.IsVisible() }

// SetScrollButtonRepeatEnabled toggles the scroll buttons auto-repeat.
func (p *PickerPanel) SetScrollButtonRepeatEnabled(isEnabled bool) {
	p.scrollButtonRepeatEnabled = isEnabled
	for _, w := range p.listWidgets {
		w.SetScrollButtonRepeatEnabled(isEnabled)
	}
}

// SetSelectedBackgroundColor sets the mask colors.
func (p *PickerPanel) SetSelectedBackgroundColor(light, dark *qt.QColor) {
	p.itemMaskWidget.SetCustomBackgroundColor(light, dark)
}

// AddColumn appends one list column to the panel.
func (p *PickerPanel) AddColumn(items []string, width, align int) {
	if len(p.listWidgets) > 0 {
		sep := NewSeparatorWidget(qt.Vertical, p.QWidget)
		p.listLayout.AddWidget(sep.QWidget)
	}

	size := qt.NewQSize2(width, p.itemHeight)
	w := widgets.NewCycleListWidget(items, size, align, p.QWidget)
	size.Delete()
	w.SetScrollButtonRepeatEnabled(p.scrollButtonRepeatEnabled)
	w.SetObjectName("cycleListWidget")
	w.SetStyleSheet(cycleListStyleSheet())
	w.SetOnScrollChanged(func() { p.itemMaskWidget.Update() })

	n := len(p.listWidgets)
	w.OnCurrentItemChanged(func(item *qt.QListWidgetItem) {
		if p.OnColumnValueChanged != nil {
			p.OnColumnValueChanged(n, item.Text())
		}
	})

	p.listWidgets = append(p.listWidgets, w)
	p.listLayout.AddWidget(w.QWidget)
}

// Value returns the current formatted value of every column.
func (p *PickerPanel) Value() []string {
	out := make([]string, 0, len(p.listWidgets))
	for _, w := range p.listWidgets {
		if item := w.CurrentItem(); item != nil {
			out = append(out, item.Text())
		}
	}
	return out
}

// SetValue sets the selected item of every column.
func (p *PickerPanel) SetValue(values []string) {
	if len(values) != len(p.listWidgets) {
		return
	}
	for i, v := range values {
		p.listWidgets[i].SetSelectedItem(v)
	}
}

// ColumnValue returns the current value of a column.
func (p *PickerPanel) ColumnValue(index int) string {
	if 0 <= index && index < len(p.listWidgets) {
		if item := p.listWidgets[index].CurrentItem(); item != nil {
			return item.Text()
		}
	}
	return ""
}

// SetColumnValue sets the selected item of a column.
func (p *PickerPanel) SetColumnValue(index int, value string) {
	if 0 <= index && index < len(p.listWidgets) {
		p.listWidgets[index].SetSelectedItem(value)
	}
}

// Column returns the list widget of a column.
func (p *PickerPanel) Column(index int) *widgets.CycleListWidget {
	if 0 <= index && index < len(p.listWidgets) {
		return p.listWidgets[index]
	}
	return nil
}

// Exec shows the panel at pos. When ani is true the panel fades in while its
// QRegion mask unfolds from the vertical center (150ms OutQuad), a faithful port
// of PickerPanel.exec's QPropertyAnimation + _onAniValueChanged mask reveal.
func (p *PickerPanel) Exec(pos *qt.QPoint, ani bool) {
	if p.IsVisible() {
		return
	}

	if ani {
		// Start fully transparent so the fade-in has no visible flash; the mask
		// is applied on the first animation frame after Show finalises geometry.
		p.SetWindowOpacity(0)
	}

	// Show before running the animation so the layout/geometry is final when the
	// mask animation reads the view size.
	p.Show()

	rect := common.GetCurrentScreenGeometry(false)
	if rect == nil {
		rect = qt.NewQRect4(0, 0, 1920, 1080)
	}

	m := p.hBoxLayout.ContentsMargins() // GoGC-armed — do NOT Delete
	w, h := p.Width()+5, p.Height()
	x := minInt(pos.X()-m.Left(), rect.Right()-w)
	y := maxInt(rect.Top(), minInt(pos.Y()-4, rect.Bottom()-h+5))
	p.Move(x, y)

	if ani {
		p.playShowAnimation()
	} else {
		p.SetWindowOpacity(1)
	}
}

// applyMask recomputes the reveal mask for the current window opacity t. When
// isExpanded is false the panel unfolds from its vertical center (h/2 factor);
// when true it collapses toward the center with the Python fade-out h/3 factor.
// Both match picker_base.py _onAniValueChanged, whose parameter is the opacity
// value (0 -> 1 on show, 1 -> 0 on fade-out).
func (p *PickerPanel) applyMask(t float64) {
	m := p.hBoxLayout.ContentsMargins() // GoGC-armed — do NOT Delete
	w := p.view.Width() + m.Left() + m.Right() + 120
	h := p.view.Height() + m.Top() + m.Bottom() + 12
	if h <= 0 {
		return
	}

	var y int
	if p.isExpanded {
		y = int(float64(h)/3*(1-t) + 0.5)
	} else {
		y = int(float64(h)/2*(1-t) + 0.5)
	}

	region := qt.NewQRegion2(0, y, w, h-y*2)
	p.SetMaskWithMask(region)
	region.Delete()
}

func (p *PickerPanel) playShowAnimation() {
	p.stopAnimation()
	p.isExpanded = false

	curve := qt.NewQEasingCurve3(qt.QEasingCurve__OutQuad)
	p.ani = common.NewProgressAnimation(150, curve)
	curve.Delete() // SetEasingCurve copies; the temporary is freed here.
	p.ani.OnProgress(func(t float64) {
		p.SetWindowOpacity(t)
		p.applyMask(t)
	})
	p.ani.OnFinished(func() {
		p.ClearMask()
		p.SetWindowOpacity(1)
		// Do not Delete the animation from inside its own finished handler
		// (use-after-free); stopAnimation() reclaims it when the panel is next
		// closed or reopened.
	})
	p.ani.Start()
}

func (p *PickerPanel) fadeOut() {
	p.stopAnimation()
	p.isExpanded = true

	curve := qt.NewQEasingCurve3(qt.QEasingCurve__OutQuad)
	p.ani = common.NewProgressAnimation(150, curve)
	curve.Delete() // SetEasingCurve copies; the temporary is freed here.
	p.ani.OnProgress(func(t float64) {
		p.SetWindowOpacity(1 - t)
		// applyMask takes the *opacity* value (1 -> 0 during fade-out), exactly
		// like the Python _onAniValueChanged; passing the 0 -> 1 progress here
		// reversed the collapse and left a ghosting/afterimage behind the
		// closing panel.
		p.applyMask(1 - t)
	})
	p.ani.OnFinished(func() {
		p.ClearMask()
		p.ani = nil
		p.Close()
	})
	p.ani.Start()
}

func (p *PickerPanel) stopAnimation() {
	if p.ani != nil {
		p.ani.Stop()
		p.ani.Delete()
		p.ani = nil
	}
}

func (p *PickerPanel) confirm() {
	values := p.Value()
	p.closePanel()
	if p.OnConfirmed != nil {
		p.OnConfirmed(values)
	}
}

func (p *PickerPanel) closePanel() {
	p.fadeOut()
}

// cloneColor returns an independent copy of c, preserving an invalid (null)
// color and the alpha channel. QColor.Name() maps an invalid color to
// "#000000" and drops alpha, so copying through Name() would silently turn the
// default null selected-background into an opaque black band and would flatten
// semi-transparent custom colours; FallbackThemeColor relies on invalidity to
// substitute the theme color (mirrors Python autoFallbackThemeColor).
func cloneColor(c *qt.QColor) *qt.QColor {
	if c == nil || !c.IsValid() {
		return qt.NewQColor()
	}
	return qt.NewQColor9(c)
}

// cycleListStyleSheet returns the themed CycleListWidget QSS used by the picker
// panel. The FluentTimePicker QSS `CycleListWidget` class selector cannot match
// the native QListWidget, so the equivalent transparent background, item
// padding and hover/selected highlight are applied under the #cycleListWidget
// objectName selector (mirrors time_picker.qss light/dark).
func cycleListStyleSheet() string {
	if common.IsDarkTheme() {
		return "QListWidget#cycleListWidget { background-color: transparent; border: none; outline: none; }" +
			" QListWidget#cycleListWidget::item { color: white; background-color: transparent; border: none; border-radius: 5px; margin: 0 4px; padding-left: 11px; padding-right: 11px; }" +
			" QListWidget#cycleListWidget::item:hover { background-color: rgba(255, 255, 255, 9); }" +
			" QListWidget#cycleListWidget::item:selected { background-color: rgba(255, 255, 255, 9); }"
	}
	return "QListWidget#cycleListWidget { background-color: transparent; border: none; outline: none; }" +
		" QListWidget#cycleListWidget::item { color: black; background-color: transparent; border: none; border-radius: 5px; margin: 0 4px; padding-left: 11px; padding-right: 11px; }" +
		" QListWidget#cycleListWidget::item:hover { background-color: rgba(0, 0, 0, 9); }" +
		" QListWidget#cycleListWidget::item:selected { background-color: rgba(0, 0, 0, 9); }"
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
