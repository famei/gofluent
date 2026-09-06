package settings

import (
	"github.com/famei/gofluent/common"
	dialog_box "github.com/famei/gofluent/components/dialog_box"
	"github.com/famei/gofluent/components/widgets"
	qt "github.com/mappu/miqt/qt"
)

// CustomColorSettingCard is an expand group card that switches between the
// default color and a custom color chosen with a ColorDialog.
type CustomColorSettingCard struct {
	*ExpandGroupSettingCard
	enableAlpha        bool
	configItem         *common.ConfigItem
	defaultColor       *qt.QColor
	customColor        *qt.QColor
	choiceLabel        *qt.QLabel
	radioWidget        *qt.QWidget
	radioLayout        *qt.QVBoxLayout
	defaultRadioButton *widgets.RadioButton
	customRadioButton  *widgets.RadioButton
	buttonGroup        *qt.QButtonGroup
	customColorWidget  *qt.QWidget
	customColorLayout  *qt.QHBoxLayout
	customLabel        *qt.QLabel
	chooseColorButton  *qt.QPushButton

	OnColorChanged func(color *qt.QColor)
}

// NewCustomColorSettingCard builds a custom color setting card.
func NewCustomColorSettingCard(configItem *common.ConfigItem, icon interface{}, title, content string, parent *qt.QWidget, enableAlpha bool) *CustomColorSettingCard {
	c := &CustomColorSettingCard{ExpandGroupSettingCard: NewExpandGroupSettingCard(icon, title, content, parent)}
	c.enableAlpha = enableAlpha
	c.configItem = configItem
	c.defaultColor = cloneColor(asColor(configItem.DefaultValue))
	c.customColor = cloneColor(asColor(configItem.Value()))

	c.choiceLabel = qt.NewQLabel(c.QWidget)

	c.radioWidget = qt.NewQWidget(c.view.QWidget)
	c.radioLayout = qt.NewQVBoxLayout(c.radioWidget)
	c.defaultRadioButton = widgets.NewRadioButtonText("Default color", c.radioWidget)
	c.customRadioButton = widgets.NewRadioButtonText("Custom color", c.radioWidget)
	c.buttonGroup = qt.NewQButtonGroup2(c.QObject)

	c.customColorWidget = qt.NewQWidget(c.view.QWidget)
	c.customColorLayout = qt.NewQHBoxLayout(c.customColorWidget)
	c.customLabel = qt.NewQLabel5("Custom color", c.customColorWidget)
	c.chooseColorButton = qt.NewQPushButton5("Choose color", c.customColorWidget)

	c.initWidget()
	return c
}

func (c *CustomColorSettingCard) initWidget() {
	c.initLayout()

	if !colorEquals(c.defaultColor, c.customColor) {
		c.customRadioButton.SetChecked(true)
		c.chooseColorButton.SetEnabled(true)
	} else {
		c.defaultRadioButton.SetChecked(true)
		c.chooseColorButton.SetEnabled(false)
	}

	c.choiceLabel.SetText(c.buttonGroup.CheckedButton().Text())
	c.choiceLabel.AdjustSize()

	c.choiceLabel.SetObjectName("titleLabel")
	c.customLabel.SetObjectName("titleLabel")
	c.chooseColorButton.SetObjectName("chooseColorButton")

	c.buttonGroup.OnButtonClicked(func(button *qt.QAbstractButton) {
		c.onRadioButtonClicked(button)
	})
	c.chooseColorButton.OnClicked(c.showColorDialog)
}

func (c *CustomColorSettingCard) initLayout() {
	c.AddWidget(c.choiceLabel.QWidget)

	c.radioLayout.SetSpacing(19)
	c.radioLayout.SetContentsMargins(48, 18, 0, 18)
	c.buttonGroup.AddButton(c.customRadioButton.QAbstractButton)
	c.buttonGroup.AddButton(c.defaultRadioButton.QAbstractButton)
	c.radioLayout.AddWidget(c.customRadioButton.QWidget)
	c.radioLayout.AddWidget(c.defaultRadioButton.QWidget)
	c.radioLayout.SetSizeConstraint(qt.QLayout__SetMinimumSize)

	c.customColorLayout.SetContentsMargins(48, 18, 44, 18)
	c.customColorLayout.AddWidget3(c.customLabel.QWidget, 0, qt.AlignLeft)
	c.customColorLayout.AddWidget3(c.chooseColorButton.QWidget, 0, qt.AlignRight)
	c.customColorLayout.SetSizeConstraint(qt.QLayout__SetMinimumSize)

	c.viewLayout.SetSpacing(0)
	c.viewLayout.SetContentsMargins(0, 0, 0, 0)
	c.AddGroupWidget(c.radioWidget)
	c.AddGroupWidget(c.customColorWidget)
}

func (c *CustomColorSettingCard) onRadioButtonClicked(button *qt.QAbstractButton) {
	if button.Text() == c.choiceLabel.Text() {
		return
	}

	c.choiceLabel.SetText(button.Text())
	c.choiceLabel.AdjustSize()

	if button.UnsafePointer() == c.defaultRadioButton.QAbstractButton.UnsafePointer() {
		c.chooseColorButton.SetDisabled(true)
		common.QConfigInstance.Set(c.configItem, c.defaultColor, true)
		if !colorEquals(c.defaultColor, c.customColor) && c.OnColorChanged != nil {
			c.OnColorChanged(c.defaultColor)
		}
	} else {
		c.chooseColorButton.SetDisabled(false)
		common.QConfigInstance.Set(c.configItem, c.customColor, true)
		if !colorEquals(c.defaultColor, c.customColor) && c.OnColorChanged != nil {
			c.OnColorChanged(c.customColor)
		}
	}
}

func (c *CustomColorSettingCard) showColorDialog() {
	d := dialog_box.NewColorDialog(asColor(c.configItem.Value()), "Choose color", c.Window(), c.enableAlpha)
	d.OnColorChanged = func(color *qt.QColor) {
		common.QConfigInstance.Set(c.configItem, color, true)
		c.customColor = cloneColor(color)
		if c.OnColorChanged != nil {
			c.OnColorChanged(color)
		}
	}
	d.Exec()
}

// ConfigItem returns the backing config item.
func (c *CustomColorSettingCard) ConfigItem() *common.ConfigItem { return c.configItem }

// ---------------------------------------------------------------------------
// helpers

func cloneColor(c *qt.QColor) *qt.QColor {
	if c == nil {
		return qt.NewQColor6("#000000")
	}
	return qt.NewQColor9(c)
}

func colorEquals(a, b *qt.QColor) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.NameWithFormat(qt.QColor__HexArgb) == b.NameWithFormat(qt.QColor__HexArgb)
}
