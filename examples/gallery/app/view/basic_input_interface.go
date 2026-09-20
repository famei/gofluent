package view

import (
	gcommon "github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	gallerycommon "github.com/famei/gofluent/examples/gallery/app/common"
	"github.com/famei/gofluent/examples/gallery/app/resource"
	qt "github.com/mappu/miqt/qt"
)

// BasicInputInterface is the "Basic input" gallery page (port of
// app/view/basic_input_interface.py).
type BasicInputInterface struct {
	*GalleryInterface
	switchButton *widgets.SwitchButton
}

// NewBasicInputInterface builds the basic input interface.
func NewBasicInputInterface(parent *qt.QWidget) *BasicInputInterface {
	translator := gallerycommon.NewTranslator()
	i := &BasicInputInterface{GalleryInterface: NewGalleryInterface(translator.BasicInput, "github.com/famei/gofluent/components/widgets", parent)}
	i.SetObjectName("basicInputInterface")

	tr := func(s string) string { return gallerycommon.Tr("BasicInputInterface", s) }

	const buttonSrc = "basic_input/button/main.go"

	i.AddExampleCard(tr("A simple button with text content"), widgets.NewPushButtonText(tr("Standard push button"), nil).QWidget, "components/widgets/button.go", codeButtonPush, 0)

	button := widgets.NewToolButtonIcon(resource.Icon("kunkun.png"), nil)
	button.SetIconSize(qt.NewQSize2(40, 40))
	button.Resize(70, 70)
	i.AddExampleCard(tr("A button with graphical content"), button.QWidget, "components/widgets/button.go", codeButtonIcon, 0)

	i.AddExampleCard(tr("Accent style applied to push button"), widgets.NewPrimaryPushButtonText(tr("Accent style button"), nil).QWidget, "components/widgets/button.go", codeButtonPrimaryPush, 0)
	i.AddExampleCard(tr("Accent style applied to tool button"), widgets.NewPrimaryToolButtonIcon(gcommon.LeafTwo, nil).QWidget, "components/widgets/button.go", codeButtonPrimaryTool, 0)
	i.AddExampleCard(tr("Pill push button"), widgets.NewPillPushButtonIcon(gcommon.Tag, tr("Tag"), i.QWidget).QWidget, "components/widgets/button.go", codeButtonPillPush, 0)
	i.AddExampleCard(tr("Pill tool button"), widgets.NewPillToolButtonIcon(gcommon.LeafTwo, nil).QWidget, "components/widgets/button.go", codeButtonPillTool, 0)
	i.AddExampleCard(tr("A transparent push button"), widgets.NewTransparentPushButtonIcon(gcommon.Library, tr("Transparent push button"), i.QWidget).QWidget, "components/widgets/button.go", codeButtonTransparentPush, 0)
	i.AddExampleCard(tr("A transparent tool button"), widgets.NewTransparentToolButtonIcon(gcommon.Library, i.QWidget).QWidget, "components/widgets/button.go", codeButtonTransparentTool, 0)

	i.AddExampleCard(tr("A 2-state CheckBox"), widgets.NewCheckBoxText(tr("Two-state CheckBox"), nil).QWidget,
		"components/widgets/check_box.go", codeCheckBoxTwoState, 0)

	checkBox := widgets.NewCheckBoxText(tr("Three-state CheckBox"), nil)
	checkBox.SetTristate()
	i.AddExampleCard(tr("A 3-state CheckBox"), checkBox.QWidget,
		"components/widgets/check_box.go", codeCheckBoxThreeState, 0)

	comboBox := widgets.NewComboBox(nil)
	comboBox.AddItems([]string{"shoko 🥰", "西宫硝子 😊", "一级棒卡哇伊的硝子酱 😘"})
	comboBox.SetCurrentIndex(0)
	comboBox.SetMinimumWidth(210)
	i.AddExampleCard(tr("A ComboBox with items"), comboBox.QWidget,
		"components/widgets/combo_box.go", codeComboBoxItems, 0)

	editableComboBox := widgets.NewEditableComboBox(nil)
	editableComboBox.AddItems([]string{tr("Star Platinum"), tr("Crazy Diamond"), tr("Gold Experience"), tr("Sticky Fingers")})
	editableComboBox.SetPlaceholderText(tr("Choose your stand"))
	editableComboBox.SetMinimumWidth(210)
	i.AddExampleCard(tr("An editable ComboBox"), editableComboBox.QWidget,
		"components/widgets/combo_box.go", codeComboBoxEditable, 0)

	menu := widgets.NewRoundMenu("", i.QWidget)
	menu.AddAction(gcommon.NewActionFluentIcon(gcommon.Send, tr("Send"), nil).QAction)
	menu.AddAction(gcommon.NewActionFluentIcon(gcommon.Save, tr("Save"), nil).QAction)

	ddPush := widgets.NewDropDownPushButtonIcon(gcommon.Mail, tr("Email"), i.QWidget)
	ddPush.SetMenu(menu)
	i.AddExampleCard(tr("A push button with drop down menu"), ddPush.QWidget, "components/widgets/button.go", codeButtonDropDownPush, 0)

	ddTool := widgets.NewDropDownToolButtonIcon(gcommon.Mail, i.QWidget)
	ddTool.SetMenu(menu)
	i.AddExampleCard(tr("A tool button with drop down menu"), ddTool.QWidget, "components/widgets/button.go", codeButtonDropDownTool, 0)

	pddPush := widgets.NewPrimaryDropDownPushButtonIcon(gcommon.Mail, tr("Email"), i.QWidget)
	pddPush.SetMenu(menu)
	i.AddExampleCard(tr("A primary color push button with drop down menu"), pddPush.QWidget, "components/widgets/button.go", codeButtonPrimaryDropDownPush, 0)

	pddTool := widgets.NewPrimaryDropDownToolButtonIcon(gcommon.Mail, i.QWidget)
	pddTool.SetMenu(menu)
	i.AddExampleCard(tr("A primary color tool button with drop down menu"), pddTool.QWidget, "components/widgets/button.go", codeButtonPrimaryDropDownTool, 0)

	tddPush := widgets.NewTransparentDropDownPushButtonIcon(gcommon.Mail, tr("Email"), i.QWidget)
	tddPush.SetMenu(menu)
	i.AddExampleCard(tr("A transparent push button with drop down menu"), tddPush.QWidget, "components/widgets/button.go", codeButtonTransparentDropDownPush, 0)

	tddTool := widgets.NewTransparentDropDownToolButtonIcon(gcommon.Mail, i.QWidget)
	tddTool.SetMenu(menu)
	i.AddExampleCard(tr("A transparent tool button with drop down menu"), tddTool.QWidget, "components/widgets/button.go", codeButtonTransparentDropDownTool, 0)

	i.AddExampleCard(tr("A hyperlink button that navigates to a URI"),
		widgets.NewHyperlinkButtonIcon(gcommon.Link, "https://qfluentwidgets.com", "GitHub", i.QWidget).QWidget,
		"components/widgets/button.go", codeButtonHyperlink, 0)

	radioWidget := qt.NewQWidget2()
	radioLayout := qt.NewQVBoxLayout(radioWidget)
	radioLayout.SetContentsMargins(2, 0, 0, 0)
	radioLayout.SetSpacing(15)
	radioButton1 := widgets.NewRadioButtonText(tr("Star Platinum"), radioWidget)
	radioButton2 := widgets.NewRadioButtonText(tr("Crazy Diamond"), radioWidget)
	radioButton3 := widgets.NewRadioButtonText(tr("Soft and Wet"), radioWidget)
	buttonGroup := qt.NewQButtonGroup2(radioWidget.QObject)
	buttonGroup.AddButton(radioButton1.QAbstractButton)
	buttonGroup.AddButton(radioButton2.QAbstractButton)
	buttonGroup.AddButton(radioButton3.QAbstractButton)
	radioLayout.AddWidget(radioButton1.QWidget)
	radioLayout.AddWidget(radioButton2.QWidget)
	radioLayout.AddWidget(radioButton3.QWidget)
	radioButton1.Click()
	i.AddExampleCard(tr("A group of RadioButton controls in a button group"), radioWidget,
		"components/widgets/button.go", codeRadioButtonGroup, 0)

	slider := widgets.NewSliderOrientation(qt.Horizontal, nil)
	slider.SetRange(0, 100)
	slider.SetValue(30)
	slider.SetMinimumWidth(200)
	i.AddExampleCard(tr("A simple horizontal slider"), slider.QWidget,
		"components/widgets/slider.go", codeSlider, 0)

	splitPush := widgets.NewSplitPushButtonIcon(gcommon.LeafTwo, tr("Choose your stand"), i.QWidget)
	splitPush.SetFlyout(i.createStandMenu(splitPush))
	i.AddExampleCard(tr("A split push button with drop down menu"), splitPush.QWidget, "components/widgets/button.go", codeButtonSplitPush, 0)

	ikunMenu := widgets.NewRoundMenu("", i.QWidget)
	ikunMenu.AddActions([]*qt.QAction{
		gcommon.NewActionText(tr("Sing"), nil).QAction,
		gcommon.NewActionText(tr("Jump"), nil).QAction,
		gcommon.NewActionText(tr("Rap"), nil).QAction,
		gcommon.NewActionText(tr("Music"), nil).QAction,
	})
	splitTool := widgets.NewSplitToolButtonIcon(resource.Icon("kunkun.png"), i.QWidget)
	splitTool.SetIconSize(qt.NewQSize2(30, 30))
	splitTool.SetFlyout(ikunMenu)
	i.AddExampleCard(tr("A split tool button with drop down menu"), splitTool.QWidget, "components/widgets/button.go", codeButtonSplitTool, 0)

	pspPush := widgets.NewPrimarySplitPushButtonIcon(gcommon.LeafTwo, tr("Choose your stand"), i.QWidget)
	pspPush.SetFlyout(i.createStandMenu(pspPush.SplitPushButton))
	i.AddExampleCard(tr("A primary color split push button with drop down menu"), pspPush.QWidget, "components/widgets/button.go", codeButtonPrimarySplitPush, 0)

	pspTool := widgets.NewPrimarySplitToolButtonIcon(gcommon.LeafTwo, i.QWidget)
	pspTool.SetFlyout(ikunMenu)
	i.AddExampleCard(tr("A primary color split tool button with drop down menu"), pspTool.QWidget, "components/widgets/button.go", codeButtonPrimarySplitTool, 0)

	i.switchButton = widgets.NewSwitchButtonText(tr("Off"), nil, widgets.RIGHT)
	i.switchButton.OnCheckedChanged(i.onSwitchCheckedChanged)
	i.AddExampleCard(tr("A simple switch button"), i.switchButton.QWidget,
		"components/widgets/switch_button.go", codeSwitchButton, 0)

	i.AddExampleCard(tr("A simple toggle push button"), widgets.NewToggleButtonIcon(gcommon.LeafTwo, tr("Start practicing"), i.QWidget).QWidget, "components/widgets/button.go", codeButtonTogglePush, 0)
	i.AddExampleCard(tr("A simple toggle tool button"), widgets.NewToggleToolButtonIcon(gcommon.LeafTwo, i.QWidget).QWidget, "components/widgets/button.go", codeButtonToggleTool, 0)
	i.AddExampleCard(tr("A transparent toggle push button"), widgets.NewTransparentTogglePushButtonIcon(gcommon.LeafTwo, tr("Start practicing"), i.QWidget).QWidget, "components/widgets/button.go", codeButtonTransparentTogglePush, 0)
	i.AddExampleCard(tr("A transparent toggle tool button"), widgets.NewTransparentToggleToolButtonIcon(gcommon.LeafTwo, i.QWidget).QWidget, "components/widgets/button.go", codeButtonTransparentToggleTool, 0)
	return i
}

func (i *BasicInputInterface) onSwitchCheckedChanged(isChecked bool) {
	if isChecked {
		i.switchButton.SetText(gallerycommon.Tr("BasicInputInterface", "On"))
	} else {
		i.switchButton.SetText(gallerycommon.Tr("BasicInputInterface", "Off"))
	}
}

func (i *BasicInputInterface) createStandMenu(button *widgets.SplitPushButton) *widgets.RoundMenu {
	menu := widgets.NewRoundMenu("", i.QWidget)
	names := []string{
		gallerycommon.Tr("BasicInputInterface", "Star Platinum"),
		gallerycommon.Tr("BasicInputInterface", "Crazy Diamond"),
		gallerycommon.Tr("BasicInputInterface", "Gold Experience"),
		gallerycommon.Tr("BasicInputInterface", "Sticky Fingers"),
	}
	for _, name := range names {
		name := name
		action := gcommon.NewActionText(name, nil)
		action.OnTriggered(func() { button.SetText(name) })
		menu.AddAction(action.QAction)
	}
	return menu
}
