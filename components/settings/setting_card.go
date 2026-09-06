package settings

import (
	"reflect"
	"strconv"

	"github.com/famei/gofluent/common"
	dialog_box "github.com/famei/gofluent/components/dialog_box"
	"github.com/famei/gofluent/components/widgets"
	qt "github.com/mappu/miqt/qt"
)

// SettingIconWidget paints a fluent icon with reduced opacity when disabled.
type SettingIconWidget struct {
	*widgets.IconWidget
	icon interface{}
}

// NewSettingIconWidget builds a setting icon widget from an icon source
// (*qt.QIcon, string path or common.FluentIconBase).
func NewSettingIconWidget(icon interface{}, parent *qt.QWidget) *SettingIconWidget {
	w := &SettingIconWidget{IconWidget: widgets.NewIconWidgetIcon(icon, parent), icon: icon}
	w.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		if !w.IsEnabled() {
			painter.SetOpacity(0.36)
		}
		painter.SetRenderHints(qt.QPainter__Antialiasing | qt.QPainter__SmoothPixmapTransform)
		rect := qt.NewQRectF5(w.Rect())
		defer rect.Delete()
		common.DrawIcon(w.icon, painter, rect)
		painter.End()
	})
	return w
}

// SetIcon sets the icon source and repaints.
func (w *SettingIconWidget) SetIcon(icon interface{}) {
	w.icon = icon
	w.IconWidget.SetIcon(icon)
}

// IconSource returns the raw icon source passed to SetIcon.
func (w *SettingIconWidget) IconSource() interface{} { return w.icon }

// SettingCard is the base setting card (title + optional content + icon).
type SettingCard struct {
	*qt.QFrame
	iconLabel    *SettingIconWidget
	titleLabel   *qt.QLabel
	contentLabel *qt.QLabel
	hBoxLayout   *qt.QHBoxLayout
	vBoxLayout   *qt.QVBoxLayout
}

// NewSettingCard builds a setting card. content may be empty to render the
// shorter (50px) card layout.
func NewSettingCard(icon interface{}, title, content string, parent *qt.QWidget) *SettingCard {
	card := &SettingCard{QFrame: qt.NewQFrame(parent)}
	card.iconLabel = NewSettingIconWidget(icon, card.QWidget)
	card.titleLabel = qt.NewQLabel5(title, card.QWidget)
	card.contentLabel = qt.NewQLabel5(content, card.QWidget)
	card.hBoxLayout = qt.NewQHBoxLayout(card.QWidget)
	card.vBoxLayout = qt.NewQVBoxLayout2()

	if content == "" {
		card.contentLabel.Hide()
		card.SetFixedHeight(50)
	} else {
		card.SetFixedHeight(70)
	}
	card.iconLabel.SetFixedSize2(16, 16)

	card.hBoxLayout.SetSpacing(0)
	card.hBoxLayout.SetContentsMargins(16, 0, 0, 0)
	card.vBoxLayout.SetSpacing(0)
	card.vBoxLayout.SetContentsMargins(0, 0, 0, 0)

	card.hBoxLayout.AddWidget3(card.iconLabel.QWidget, 0, qt.AlignLeft)
	card.hBoxLayout.AddSpacing(16)
	card.hBoxLayout.AddLayout(card.vBoxLayout.QLayout)
	card.vBoxLayout.AddWidget3(card.titleLabel.QWidget, 0, qt.AlignLeft)
	card.vBoxLayout.AddWidget3(card.contentLabel.QWidget, 0, qt.AlignLeft)
	card.hBoxLayout.AddSpacing(16)
	card.hBoxLayout.AddStretchWithStretch(1)

	card.contentLabel.SetObjectName("contentLabel")
	common.FluentStyleSheet(common.FluentSettingCard).Apply(card.QFrame.QWidget, common.ThemeAuto)
	card.installPaintEvent()
	return card
}

// SetTitle sets the card title.
func (c *SettingCard) SetTitle(title string) {
	c.titleLabel.SetText(title)
}

// SetContent sets the card content and toggles its visibility.
func (c *SettingCard) SetContent(content string) {
	c.contentLabel.SetText(content)
	c.contentLabel.SetVisible(content != "")
}

// SetValue stores the config value (overridden by subclasses).
func (c *SettingCard) SetValue(value interface{}) {}

// SetIconSize sets the icon fixed size.
func (c *SettingCard) SetIconSize(width, height int) {
	c.iconLabel.SetFixedSize2(width, height)
}

// IconLabel exposes the icon widget.
func (c *SettingCard) IconLabel() *SettingIconWidget { return c.iconLabel }

// TitleLabel exposes the title label.
func (c *SettingCard) TitleLabel() *qt.QLabel { return c.titleLabel }

// ContentLabel exposes the content label.
func (c *SettingCard) ContentLabel() *qt.QLabel { return c.contentLabel }

// HBoxLayout exposes the horizontal layout (subclasses add trailing widgets).
func (c *SettingCard) HBoxLayout() *qt.QHBoxLayout { return c.hBoxLayout }

// VBoxLayout exposes the vertical text layout.
func (c *SettingCard) VBoxLayout() *qt.QVBoxLayout { return c.vBoxLayout }

func (c *SettingCard) installPaintEvent() {
	c.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		super(e)
		painter := qt.NewQPainter2(c.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing)

		var brush, pen *qt.QColor
		if common.IsDarkTheme() {
			brush = qt.NewQColor11(255, 255, 255, 13)
			pen = qt.NewQColor11(0, 0, 0, 50)
		} else {
			brush = qt.NewQColor11(255, 255, 255, 170)
			pen = qt.NewQColor11(0, 0, 0, 19)
		}
		b := qt.NewQBrush3(brush)
		painter.SetBrush(b)
		b.Delete()
		painter.SetPen(pen)
		adjusted := c.Rect().Adjusted(1, 1, -1, -1) // GoGC-armed — do NOT Delete
		painter.DrawRoundedRect3(adjusted, 6, 6)
		painter.End()
	})
}

// SwitchSettingCard is a setting card with an on/off switch.
type SwitchSettingCard struct {
	*SettingCard
	configItem   *common.ConfigItem
	switchButton *widgets.SwitchButton

	OnCheckedChanged func(isChecked bool)
}

// NewSwitchSettingCard builds a switch setting card. configItem may be nil.
func NewSwitchSettingCard(icon interface{}, title, content string, configItem *common.ConfigItem, parent *qt.QWidget) *SwitchSettingCard {
	card := &SwitchSettingCard{SettingCard: NewSettingCard(icon, title, content, parent), configItem: configItem}
	card.switchButton = widgets.NewSwitchButtonText("Off", card.QWidget, widgets.RIGHT)

	if configItem != nil {
		card.SetValue(configItem.Value())
		configItem.OnValueChanged = func(v interface{}) { card.SetValue(v) }
	}

	card.hBoxLayout.AddWidget3(card.switchButton.QWidget, 0, qt.AlignRight)
	card.hBoxLayout.AddSpacing(16)

	card.switchButton.OnCheckedChanged(func(isChecked bool) {
		card.SetValue(isChecked)
		if card.OnCheckedChanged != nil {
			card.OnCheckedChanged(isChecked)
		}
	})
	return card
}

// SetValue stores the checked state.
func (c *SwitchSettingCard) SetValue(isChecked interface{}) {
	checked := asBool(isChecked)
	if c.configItem != nil {
		common.QConfigInstance.Set(c.configItem, checked, true)
	}
	c.switchButton.SetChecked(checked)
	if checked {
		c.switchButton.SetText("On")
	} else {
		c.switchButton.SetText("Off")
	}
}

// SetChecked is an alias of SetValue.
func (c *SwitchSettingCard) SetChecked(isChecked bool) { c.SetValue(isChecked) }

// IsChecked reports the current switch state.
func (c *SwitchSettingCard) IsChecked() bool { return c.switchButton.IsChecked() }

// SwitchButton exposes the underlying switch.
func (c *SwitchSettingCard) SwitchButton() *widgets.SwitchButton { return c.switchButton }

// RangeSettingCard is a setting card with a slider.
type RangeSettingCard struct {
	*SettingCard
	configItem *common.ConfigItem
	slider     *widgets.Slider
	valueLabel *qt.QLabel

	OnValueChanged func(value int)
}

// NewRangeSettingCard builds a range setting card.
func NewRangeSettingCard(configItem *common.ConfigItem, icon interface{}, title, content string, parent *qt.QWidget) *RangeSettingCard {
	card := &RangeSettingCard{SettingCard: NewSettingCard(icon, title, content, parent), configItem: configItem}
	card.QWidget.SetObjectName("rangeSettingCard")
	card.slider = widgets.NewSliderOrientation(qt.Horizontal, card.QWidget)
	card.valueLabel = qt.NewQLabel(card.QWidget)
	card.slider.SetMinimumWidth(268)

	card.slider.SetSingleStep(1)
	min, max := configItem.Range()
	card.slider.SetRange(min, max)
	card.slider.SetValue(asInt(configItem.Value()))
	card.valueLabel.SetText(strconv.Itoa(asInt(configItem.Value())))

	card.hBoxLayout.AddStretchWithStretch(1)
	card.hBoxLayout.AddWidget3(card.valueLabel.QWidget, 0, qt.AlignRight)
	card.hBoxLayout.AddSpacing(6)
	card.hBoxLayout.AddWidget3(card.slider.QWidget, 0, qt.AlignRight)
	card.hBoxLayout.AddSpacing(16)

	card.valueLabel.SetObjectName("valueLabel")
	configItem.OnValueChanged = func(v interface{}) { card.SetValue(v) }
	card.slider.OnValueChanged(func(value int) {
		card.SetValue(value)
		if card.OnValueChanged != nil {
			card.OnValueChanged(value)
		}
	})
	return card
}

// SetValue stores the slider value.
func (c *RangeSettingCard) SetValue(value interface{}) {
	v := asInt(value)
	common.QConfigInstance.Set(c.configItem, v, true)
	c.valueLabel.SetText(strconv.Itoa(v))
	c.valueLabel.AdjustSize()
	c.slider.SetValue(v)
}

// Slider exposes the underlying slider.
func (c *RangeSettingCard) Slider() *widgets.Slider { return c.slider }

// PushSettingCard is a setting card with a push button.
type PushSettingCard struct {
	*SettingCard
	button *qt.QPushButton

	OnClicked func()
}

// NewPushSettingCard builds a push setting card.
func NewPushSettingCard(text string, icon interface{}, title, content string, parent *qt.QWidget) *PushSettingCard {
	card := &PushSettingCard{SettingCard: NewSettingCard(icon, title, content, parent)}
	card.button = qt.NewQPushButton5(text, card.QWidget)
	card.hBoxLayout.AddWidget3(card.button.QWidget, 0, qt.AlignRight)
	card.hBoxLayout.AddSpacing(16)
	card.button.OnClicked(func() {
		if card.OnClicked != nil {
			card.OnClicked()
		}
	})
	return card
}

// Button exposes the underlying push button.
func (c *PushSettingCard) Button() *qt.QPushButton { return c.button }

// PrimaryPushSettingCard is a push setting card with the primary color.
type PrimaryPushSettingCard struct {
	*PushSettingCard
}

// NewPrimaryPushSettingCard builds a primary push setting card.
func NewPrimaryPushSettingCard(text string, icon interface{}, title, content string, parent *qt.QWidget) *PrimaryPushSettingCard {
	card := &PrimaryPushSettingCard{PushSettingCard: NewPushSettingCard(text, icon, title, content, parent)}
	card.button.SetObjectName("primaryButton")
	return card
}

// HyperlinkCard is a setting card with a hyperlink button.
type HyperlinkCard struct {
	*SettingCard
	linkButton *widgets.HyperlinkButton
}

// NewHyperlinkCard builds a hyperlink card.
func NewHyperlinkCard(url, text string, icon interface{}, title, content string, parent *qt.QWidget) *HyperlinkCard {
	card := &HyperlinkCard{SettingCard: NewSettingCard(icon, title, content, parent)}
	card.linkButton = widgets.NewHyperlinkButtonURL(url, text, card.QWidget)
	card.hBoxLayout.AddWidget3(card.linkButton.QWidget, 0, qt.AlignRight)
	card.hBoxLayout.AddSpacing(16)
	return card
}

// LinkButton exposes the underlying hyperlink button.
func (c *HyperlinkCard) LinkButton() *widgets.HyperlinkButton { return c.linkButton }

// ColorPickerButton is a tool button that shows a color and opens a
// ColorDialog when clicked.
type ColorPickerButton struct {
	*qt.QToolButton
	title       string
	enableAlpha bool
	color       *qt.QColor

	OnColorChanged func(color *qt.QColor)
}

// NewColorPickerButton builds a color picker button.
func NewColorPickerButton(color *qt.QColor, title string, parent *qt.QWidget, enableAlpha bool) *ColorPickerButton {
	w := &ColorPickerButton{QToolButton: qt.NewQToolButton(parent), title: title, enableAlpha: enableAlpha}
	w.SetObjectName("colorPickerButton")
	w.SetFixedSize2(96, 32)
	w.SetAttribute(qt.WA_TranslucentBackground)
	w.SetColor(color)
	w.SetCursor(qt.NewQCursor2(qt.PointingHandCursor))
	w.OnClicked(w.showColorDialog)
	w.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		super(e)
		painter := qt.NewQPainter2(w.QPaintDevice)
		defer painter.Delete()
		painter.SetRenderHints(qt.QPainter__Antialiasing)

		var pc *qt.QColor
		if common.IsDarkTheme() {
			pc = qt.NewQColor11(255, 255, 255, 10)
		} else {
			pc = qt.NewQColor3(234, 234, 234)
		}
		painter.SetPen(pc)

		c := qt.NewQColor9(w.color)
		defer c.Delete()
		if !w.enableAlpha {
			c.SetAlpha(255)
		}
		brush := qt.NewQBrush3(c)
		painter.SetBrush(brush)
		brush.Delete()
		adjusted := w.Rect().Adjusted(1, 1, -1, -1) // GoGC-armed — do NOT Delete
		painter.DrawRoundedRect3(adjusted, 5, 5)
		painter.End()
	})
	return w
}

func (w *ColorPickerButton) showColorDialog() {
	d := dialog_box.NewColorDialog(w.color, "Choose "+w.title, w.Window(), w.enableAlpha)
	d.OnColorChanged = func(color *qt.QColor) {
		w.SetColor(color)
		if w.OnColorChanged != nil {
			w.OnColorChanged(color)
		}
	}
	d.Exec()
}

// SetColor stores the color and repaints.
func (w *ColorPickerButton) SetColor(color *qt.QColor) {
	w.color = qt.NewQColor9(color)
	w.Update()
}

// Color returns the current color.
func (w *ColorPickerButton) Color() *qt.QColor { return w.color }

// ColorSettingCard is a setting card with a color picker button.
type ColorSettingCard struct {
	*SettingCard
	configItem  *common.ConfigItem
	colorPicker *ColorPickerButton

	OnColorChanged func(color *qt.QColor)
}

// NewColorSettingCard builds a color setting card.
func NewColorSettingCard(configItem *common.ConfigItem, icon interface{}, title, content string, parent *qt.QWidget, enableAlpha bool) *ColorSettingCard {
	card := &ColorSettingCard{SettingCard: NewSettingCard(icon, title, content, parent), configItem: configItem}
	card.colorPicker = NewColorPickerButton(asColor(configItem.Value()), title, card.QWidget, enableAlpha)
	card.hBoxLayout.AddWidget3(card.colorPicker.QWidget, 0, qt.AlignRight)
	card.hBoxLayout.AddSpacing(16)
	card.colorPicker.OnColorChanged = func(color *qt.QColor) {
		common.QConfigInstance.Set(configItem, color, true)
		if card.OnColorChanged != nil {
			card.OnColorChanged(color)
		}
	}
	configItem.OnValueChanged = func(v interface{}) { card.SetValue(v) }
	return card
}

// SetValue stores the color.
func (c *ColorSettingCard) SetValue(value interface{}) {
	color := asColor(value)
	c.colorPicker.SetColor(color)
	common.QConfigInstance.Set(c.configItem, color, true)
}

// ColorPicker exposes the underlying color picker button.
func (c *ColorSettingCard) ColorPicker() *ColorPickerButton { return c.colorPicker }

// ComboBoxSettingCard is a setting card with a combo box.
type ComboBoxSettingCard struct {
	*SettingCard
	configItem *common.ConfigItem
	comboBox   *widgets.ComboBox
	options    []interface{}
	texts      []string
}

// NewComboBoxSettingCard builds a combo box setting card. texts maps 1:1 to the
// config item options.
func NewComboBoxSettingCard(configItem *common.ConfigItem, icon interface{}, title, content string, texts []string, parent *qt.QWidget) *ComboBoxSettingCard {
	card := &ComboBoxSettingCard{SettingCard: NewSettingCard(icon, title, content, parent), configItem: configItem}
	card.comboBox = widgets.NewComboBox(card.QWidget)
	card.options = configItem.Options()
	card.texts = texts
	card.hBoxLayout.AddWidget3(card.comboBox.QWidget, 0, qt.AlignRight)
	card.hBoxLayout.AddSpacing(16)

	for i, text := range texts {
		var option interface{}
		if i < len(card.options) {
			option = card.options[i]
		}
		card.comboBox.AddItem(text, nil, option)
	}

	card.comboBox.SetCurrentText(card.optionToText(configItem.Value()))
	card.comboBox.OnCurrentIndexChanged = func(index int) {
		common.QConfigInstance.Set(configItem, card.comboBox.ItemData(index), true)
	}
	configItem.OnValueChanged = func(v interface{}) { card.SetValue(v) }
	return card
}

func (c *ComboBoxSettingCard) optionToText(option interface{}) string {
	for i, o := range c.options {
		if reflect.DeepEqual(o, option) {
			if i < len(c.texts) {
				return c.texts[i]
			}
			return ""
		}
	}
	return ""
}

// SetValue selects the item matching the given option value.
func (c *ComboBoxSettingCard) SetValue(value interface{}) {
	text := c.optionToText(value)
	if text == "" {
		return
	}
	c.comboBox.SetCurrentText(text)
	common.QConfigInstance.Set(c.configItem, value, true)
}

// ComboBox exposes the underlying combo box.
func (c *ComboBoxSettingCard) ComboBox() *widgets.ComboBox { return c.comboBox }

// ---------------------------------------------------------------------------
// local value coercions

func asInt(v interface{}) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	case float32:
		return int(n)
	default:
		return 0
	}
}

func asBool(v interface{}) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

func asColor(v interface{}) *qt.QColor {
	if c, ok := v.(*qt.QColor); ok {
		return c
	}
	if s, ok := v.(string); ok {
		return qt.NewQColor6(s)
	}
	return qt.NewQColor6("#000000")
}
