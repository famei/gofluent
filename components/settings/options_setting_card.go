package settings

import (
	"reflect"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	qt "github.com/mappu/miqt/qt"
)

// OptionsSettingCard is an expandable card with a group of radio buttons.
type OptionsSettingCard struct {
	*ExpandSettingCard
	configItem  *common.ConfigItem
	texts       []string
	options     []interface{}
	configName  string
	choiceLabel *qt.QLabel
	buttonGroup *qt.QButtonGroup
	buttons     []*widgets.RadioButton

	OnOptionChanged func()
}

// NewOptionsSettingCard builds an options setting card. texts maps 1:1 to the
// config item options.
func NewOptionsSettingCard(configItem *common.ConfigItem, icon interface{}, title, content string, texts []string, parent *qt.QWidget) *OptionsSettingCard {
	c := &OptionsSettingCard{ExpandSettingCard: NewExpandSettingCard(icon, title, content, parent)}
	c.configItem = configItem
	c.texts = texts
	c.options = configItem.Options()
	c.configName = configItem.Name
	c.choiceLabel = qt.NewQLabel(c.QWidget)
	c.buttonGroup = qt.NewQButtonGroup2(c.QObject)

	c.choiceLabel.SetObjectName("titleLabel")
	c.AddWidget(c.choiceLabel.QWidget)

	c.viewLayout.SetSpacing(19)
	c.viewLayout.SetContentsMargins(48, 18, 0, 18)
	for _, text := range texts {
		button := widgets.NewRadioButtonText(text, c.view.QWidget)
		c.buttonGroup.AddButton(button.QAbstractButton)
		c.viewLayout.AddWidget(button.QWidget)
		c.buttons = append(c.buttons, button)
	}

	c.adjustViewSize()
	c.SetValue(configItem.Value())
	configItem.OnValueChanged = func(v interface{}) { c.SetValue(v) }
	c.buttonGroup.OnButtonClicked(func(button *qt.QAbstractButton) {
		c.onButtonClicked(button)
	})
	return c
}

func (c *OptionsSettingCard) onButtonClicked(button *qt.QAbstractButton) {
	index := -1
	for i, b := range c.buttons {
		if b.QAbstractButton.UnsafePointer() == button.UnsafePointer() {
			index = i
			break
		}
	}
	if index < 0 || button.Text() == c.choiceLabel.Text() {
		return
	}

	var value interface{}
	if index < len(c.options) {
		value = c.options[index]
	}
	common.QConfigInstance.Set(c.configItem, value, true)

	c.choiceLabel.SetText(button.Text())
	c.choiceLabel.AdjustSize()
	if c.OnOptionChanged != nil {
		c.OnOptionChanged()
	}
}

// SetValue selects the radio button whose option matches value.
func (c *OptionsSettingCard) SetValue(value interface{}) {
	common.QConfigInstance.Set(c.configItem, value, true)

	for i, button := range c.buttons {
		var option interface{}
		if i < len(c.options) {
			option = c.options[i]
		}
		isChecked := reflect.DeepEqual(option, value)
		button.SetChecked(isChecked)
		if isChecked {
			c.choiceLabel.SetText(button.Text())
			c.choiceLabel.AdjustSize()
		}
	}
}

// Buttons returns the radio buttons.
func (c *OptionsSettingCard) Buttons() []*widgets.RadioButton { return c.buttons }

// ButtonGroup returns the underlying button group.
func (c *OptionsSettingCard) ButtonGroup() *qt.QButtonGroup { return c.buttonGroup }
