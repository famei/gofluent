package view

// This file holds the pseudo-code shown in the "Source code" area of every
// gallery card. Each snippet belongs to exactly one card and is written against
// the demo widget that card actually shows (see the *Interface files), using the
// real gofluent API, so a card can never display the code of another control:
// several cards share one control family (basic_input/button/main.go backs 23
// button variants, status_info/info_bar/main.go backs the InfoBar and InfoBadge
// cards, ...), which is why the snippet travels with the card instead of being
// looked up by source path. The neighbouring source path is the Go file that
// implements the control (e.g. components/widgets/button.go).
//
// Convention: `parent` is the card's parent widget, `icon`/`menu`/`images` stand
// for a value built just above the call, and `common`/`qt`/`widgets` are the
// usual gofluent imports.

// ---------------------------------------------------------------------------
// Basic input — buttons

const (
	codeButtonPush = `button := widgets.NewPushButtonText("Standard push button", parent)
button.OnClicked(func() { /* ... */ })`

	codeButtonIcon = `button := widgets.NewToolButtonIcon(icon, parent)
button.SetIconSize(qt.NewQSize2(40, 40))
button.Resize(70, 70)`

	codeButtonPrimaryPush = `button := widgets.NewPrimaryPushButtonText("Accent style button", parent)
button.OnClicked(func() { /* ... */ })`

	codeButtonPrimaryTool = `button := widgets.NewPrimaryToolButtonIcon(common.LeafTwo, parent)`

	codeButtonPillPush = `button := widgets.NewPillPushButtonIcon(common.Tag, "Tag", parent)`

	codeButtonPillTool = `button := widgets.NewPillToolButtonIcon(common.LeafTwo, parent)`

	codeButtonTransparentPush = `button := widgets.NewTransparentPushButtonIcon(common.Library, "Transparent push button", parent)`

	codeButtonTransparentTool = `button := widgets.NewTransparentToolButtonIcon(common.Library, parent)`

	codeButtonDropDownPush = `menu := widgets.NewRoundMenu("", parent)
menu.AddAction(common.NewActionFluentIcon(common.Send, "Send", nil).QAction)

button := widgets.NewDropDownPushButtonIcon(common.Mail, "Email", parent)
button.SetMenu(menu)`

	codeButtonDropDownTool = `button := widgets.NewDropDownToolButtonIcon(common.Mail, parent)
button.SetMenu(menu)`

	codeButtonPrimaryDropDownPush = `button := widgets.NewPrimaryDropDownPushButtonIcon(common.Mail, "Email", parent)
button.SetMenu(menu)`

	codeButtonPrimaryDropDownTool = `button := widgets.NewPrimaryDropDownToolButtonIcon(common.Mail, parent)
button.SetMenu(menu)`

	codeButtonTransparentDropDownPush = `button := widgets.NewTransparentDropDownPushButtonIcon(common.Mail, "Email", parent)
button.SetMenu(menu)`

	codeButtonTransparentDropDownTool = `button := widgets.NewTransparentDropDownToolButtonIcon(common.Mail, parent)
button.SetMenu(menu)`

	codeButtonHyperlink = `button := widgets.NewHyperlinkButtonIcon(common.Link, "https://qfluentwidgets.com", "GitHub", parent)`

	codeButtonSplitPush = `button := widgets.NewSplitPushButtonIcon(common.LeafTwo, "Choose your stand", parent)
button.SetFlyout(widgets.NewRoundMenu("", parent))`

	codeButtonSplitTool = `button := widgets.NewSplitToolButtonIcon(icon, parent)
button.SetFlyout(menu)`

	codeButtonPrimarySplitPush = `button := widgets.NewPrimarySplitPushButtonIcon(common.LeafTwo, "Choose your stand", parent)
button.SetFlyout(widgets.NewRoundMenu("", parent))`

	codeButtonPrimarySplitTool = `button := widgets.NewPrimarySplitToolButtonIcon(common.LeafTwo, parent)
button.SetFlyout(menu)`

	codeButtonTogglePush = `button := widgets.NewToggleButtonIcon(common.LeafTwo, "Start practicing", parent)
button.OnToggled(func(checked bool) { /* ... */ })`

	codeButtonToggleTool = `button := widgets.NewToggleToolButtonIcon(common.LeafTwo, parent)
button.OnToggled(func(checked bool) { /* ... */ })`

	codeButtonTransparentTogglePush = `button := widgets.NewTransparentTogglePushButtonIcon(common.LeafTwo, "Start practicing", parent)
button.OnToggled(func(checked bool) { /* ... */ })`

	codeButtonTransparentToggleTool = `button := widgets.NewTransparentToggleToolButtonIcon(common.LeafTwo, parent)
button.OnToggled(func(checked bool) { /* ... */ })`
)

// Basic input — check box, combo box, radio button, slider, switch button.

const (
	codeCheckBoxTwoState = `checkBox := widgets.NewCheckBoxText("Two-state CheckBox", parent)
checkBox.OnStateChanged(func(state int) { /* ... */ })`

	codeCheckBoxThreeState = `checkBox := widgets.NewCheckBoxText("Three-state CheckBox", parent)
checkBox.SetTristate()
checkBox.OnStateChanged(func(state int) { /* ... */ })`

	codeComboBoxItems = `comboBox := widgets.NewComboBox(parent)
comboBox.AddItems([]string{"shoko 🥰", "西宫硝子 😊", "一级棒卡哇伊的硝子酱 😘"})
comboBox.SetCurrentIndex(0)
comboBox.SetMinimumWidth(210)
comboBox.OnCurrentTextChanged = func(text string) { /* ... */ }`

	codeComboBoxEditable = `comboBox := widgets.NewEditableComboBox(parent)
comboBox.AddItems([]string{"Star Platinum", "Crazy Diamond"})
comboBox.SetPlaceholderText("Choose your stand")
comboBox.OnCurrentTextChanged = func(text string) { /* ... */ }`

	codeRadioButtonGroup = `button1 := widgets.NewRadioButtonText("Star Platinum", parent)
button2 := widgets.NewRadioButtonText("Crazy Diamond", parent)

group := qt.NewQButtonGroup2(parent.QObject)
group.AddButton(button1.QAbstractButton)
group.AddButton(button2.QAbstractButton)
button1.Click()`

	codeSlider = `slider := widgets.NewSliderOrientation(qt.Horizontal, parent)
slider.SetRange(0, 100)
slider.SetValue(30)
slider.SetMinimumWidth(200)
slider.OnValueChanged(func(value int) { /* ... */ })`

	codeSwitchButton = `switchButton := widgets.NewSwitchButtonText("Off", parent, widgets.RIGHT)
switchButton.OnCheckedChanged(func(isChecked bool) {
	switchButton.SetText("On")
})`
)

// ---------------------------------------------------------------------------
// Date & time

const (
	codeCalendarPicker = `picker := date_time.NewCalendarPicker(parent)
picker.OnDateChanged = func(date *qt.QDate) { /* ... */ }`

	codeFastCalendarPicker = `picker := date_time.NewFastCalendarPicker(parent)
picker.OnDateChanged = func(date *qt.QDate) { /* ... */ }`

	codeCalendarPickerFormat = `picker := date_time.NewCalendarPicker(parent)
picker.SetDateFormat("dd/MM/yyyy")
picker.OnDateChanged = func(date *qt.QDate) { /* ... */ }`

	codeDatePicker = `picker := date_time.NewDatePicker(parent, date_time.DatePickerYYYYMMDD, false)
picker.OnDateChanged = func(date *qt.QDate) { /* ... */ }`

	codeZhDatePicker = `picker := date_time.NewZhDatePicker(parent)
picker.OnDateChanged = func(date *qt.QDate) { /* ... */ }`

	codeAMTimePicker = `picker := date_time.NewAMTimePicker(parent, false)
picker.OnTimeChanged = func(t *qt.QTime) { /* ... */ }`

	codeTimePicker = `picker := date_time.NewTimePicker(parent, false)
picker.OnTimeChanged = func(t *qt.QTime) { /* ... */ }`

	codeTimePickerSeconds = `picker := date_time.NewTimePicker(parent, true) // true adds the seconds column
picker.OnTimeChanged = func(t *qt.QTime) { /* ... */ }`
)

// ---------------------------------------------------------------------------
// Dialogs & flyouts

const (
	codeDialog = `dialog := dialog_box.NewDialog(title, content, parent)
dialog.SetContentCopyable(true)
if dialog.Exec() == int(qt.QDialog__Accepted) { /* ... */ }`

	codeMessageBox = `box := dialog_box.NewMessageBox(title, content, parent)
box.SetContentCopyable(true)
if box.Exec() == int(qt.QDialog__Accepted) { /* ... */ }`

	codeCustomMessageBox = `box := dialog_box.NewMessageBoxBase(parent)
box.ViewLayout().AddWidget(widgets.NewLineEdit(box.Widget().QWidget).QWidget)
box.ValidateFunc = func() bool { return valid }
box.Exec()`

	codeColorDialog = `dialog := dialog_box.NewColorDialog(qt.NewQColor6("cyan"), "Choose color", parent, false)
dialog.OnColorChanged = func(c *qt.QColor) { /* ... */ }
dialog.Exec()`

	codeSimpleFlyout = `widgets.FlyoutCreate(
	"Lesson 3", content, widgets.InfoBarIconSuccess, nil, true,
	target, parent, widgets.FlyoutAnimationDropDown, true,
)`

	codeComplexFlyout = `view := widgets.NewFlyoutView(title, content, nil, image, false, nil)
button := widgets.NewPushButtonText("Action", nil)
view.AddWidget(button.QWidget, 0, qt.AlignRight)
widgets.FlyoutMake(view.FlyoutViewBase, target, parent, widgets.FlyoutAnimationSlideRight, true)`

	codeTeachingTip = `widgets.TeachingTipCreate(
	target, "Lesson 4", content, widgets.InfoBarIconSuccess, nil,
	true, -1, widgets.TeachingTipTailBottom, parent, true,
)`

	codeTeachingTipImage = `view := widgets.NewTeachingTipView("Lesson 5", content, nil, image, true, widgets.TeachingTipTailLeftBottom, nil)
button := widgets.NewPushButtonText("Action", nil)
view.AddWidget(button.QWidget, 0, qt.AlignRight)
tip := widgets.TeachingTipMake(view.FlyoutViewBase, target, 3000, widgets.TeachingTipTailLeftBottom, parent, true)
view.SetOnClosed(func() { tip.Close() })`
)

// ---------------------------------------------------------------------------
// Layout

const (
	codeFlowLayout = `flow := layout.NewFlowLayout(widget, false, false)
flow.SetVerticalSpacing(20)
flow.SetHorizontalSpacing(10)
flow.AddWidget(widgets.NewPushButtonText("Star Platinum", widget).QWidget)`

	codeFlowLayoutAnimated = `flow := layout.NewFlowLayout(widget, true, false) // animation enabled
flow.SetVerticalSpacing(20)
flow.SetHorizontalSpacing(10)
flow.AddWidget(widgets.NewPushButtonText("Star Platinum", widget).QWidget)`
)

// ---------------------------------------------------------------------------
// Menus & toolbars

const (
	codeMenu = `menu := widgets.NewRoundMenu("", parent)
menu.AddAction(common.NewActionFluentIcon(common.Copy, "Copy", nil).QAction)

submenu := widgets.NewRoundMenu("Add to", parent)
submenu.SetIcon(common.Add)
submenu.AddAction(common.NewActionFluentIcon(common.Video, "Video", nil).QAction)
menu.AddMenu(submenu)
menu.AddSeparator()

menu.Exec(pos, widgets.MenuAnimationDropDown)`

	codeWidgetMenu = `menu := widgets.NewRoundMenu("", parent)
menu.AddWidget(widgets.NewAvatarWidgetImage(avatar, nil).QWidget)
menu.AddSeparator()
menu.AddAction(common.NewActionFluentIcon(common.Settings, "Settings", nil).QAction)
menu.Exec(pos, widgets.MenuAnimationDropDown)`

	codeCheckableMenu = `action1 := common.NewActionFluentIcon(common.Calendar, "Create Date", nil)
action2 := common.NewActionFluentIcon(common.Camera, "Shooting Date", nil)
for _, a := range []*common.Action{action1, action2} {
	a.SetCheckable(true)
}

group := qt.NewQActionGroup(parent.QObject)
group.AddAction(action1.QAction)
group.AddAction(action2.QAction)

menu := widgets.NewCheckableMenu("", parent, widgets.MenuIndicatorRadio)
menu.AddActions([]*qt.QAction{action1.QAction, action2.QAction})
menu.Exec(pos, widgets.MenuAnimationDropDown)`

	codeCommandBar = `bar := widgets.NewCommandBar(parent)
bar.SetToolButtonStyle(qt.ToolButtonTextBesideIcon)
bar.AddAction(common.NewActionFluentIcon(common.Add, "Add", nil).QAction)
// A pure icon button: only the icon is drawn, the text becomes its tool tip.
bar.AddIconAction(common.NewActionFluentIcon(common.Copy, "Copy", nil).QAction)
bar.AddSeparator()

button := widgets.NewTransparentDropDownPushButtonIcon(common.Sort, "Sort", parent)
bar.AddWidget(button.QWidget)
bar.AddHiddenAction(common.NewActionFluentIcon(common.Settings, "Settings", nil).QAction)`

	codeCommandBarFlyout = `view := widgets.NewCommandBarView(nil)
view.AddAction(common.NewActionFluentIcon(common.Share, "Share", nil).QAction)
view.AddAction(common.NewActionFluentIcon(common.Save, "Save", nil).QAction)
view.AddHiddenAction(common.NewActionFluentIcon(common.Print, "Print", nil).QAction)
view.ResizeToSuitableWidth()

widgets.FlyoutMake(view.FlyoutViewBase, pos, parent, widgets.FlyoutAnimationFadeIn, true)`
)

// ---------------------------------------------------------------------------
// Navigation

const (
	codeBreadcrumbBar = `bar := navigation.NewBreadcrumbBar(parent)
bar.AddItem("Home", "Home")
bar.AddItem("Documents", "Documents")
bar.OnCurrentItemChanged(func(objectName string) { /* ... */ })`

	codePivot = `pivot := navigation.NewPivot(parent)
pivot.AddItem("songInterface", "Song", func(bool) { /* ... */ }, nil)
pivot.AddItem("albumInterface", "Album", func(bool) { /* ... */ }, nil)
pivot.OnCurrentItemChanged(func(key string) { /* ... */ })
pivot.SetCurrentItem("songInterface")`

	codeSegmentedWidget = `segmented := navigation.NewSegmentedWidget(parent)
segmented.AddItem("songInterface", "Song", func(bool) { /* ... */ }, nil)
segmented.AddItem("albumInterface", "Album", func(bool) { /* ... */ }, nil)
segmented.SetCurrentItem("songInterface")`

	codeSegmentedToolWidget = `segmented := navigation.NewSegmentedToggleToolWidget(parent)
segmented.AddItem("k1", common.SquareSparkle, nil)
segmented.AddItem("k2", common.Checkbox, nil)
segmented.SetCurrentItem("k1")
segmented.OnCurrentItemChanged(func(key string) { /* ... */ })`

	codeTabBar = `tabBar := widgets.NewTabBar(parent)
tabBar.SetMovable(true)
tabBar.AddTab("songInterface", "Song", icon, func() { /* ... */ })
tabBar.OnTabAddRequested(func() { /* ... */ })
tabBar.OnTabCloseRequested(func(index int) { /* ... */ })`
)

// ---------------------------------------------------------------------------
// Scrolling

const (
	codeScrollArea = `area := widgets.NewScrollArea(nil)
area.SetWidget(label.QWidget)
area.SetFixedSize2(775, 430)`

	codeSmoothScrollArea = `area := widgets.NewSmoothScrollArea(nil)
area.SetScrollAnimation(qt.Vertical, 400)
area.SetScrollAnimation(qt.Horizontal, 400)
area.SetWidget(label.QWidget)`

	codeSingleDirectionScrollArea = `area := widgets.NewSingleDirectionScrollArea(parent, qt.Horizontal)
area.SetWidget(label.QWidget)
area.SetFixedSize2(660, 498)`

	codePipsPager = `pager := widgets.NewHorizontalPipsPager(parent)
pager.SetPageNumber(15)
pager.SetPreviousButtonDisplayMode(widgets.PipsDisplayAlways)
pager.SetNextButtonDisplayMode(widgets.PipsDisplayAlways)
pager.OnCurrentIndexChanged(func(index int) { /* ... */ })`
)

// ---------------------------------------------------------------------------
// Status & info

const (
	codeStateToolTip = `tip := widgets.NewStateToolTip("Training model", "Please wait patiently", parent)
pos := tip.GetSuitablePos()
tip.Move(pos.X(), pos.Y())
tip.Show()

tip.SetContent("The model training is complete! 😆")
tip.SetState(true)`

	codeToolTipButton = `button := widgets.NewPushButtonText("Button with a simple ToolTip", parent)
widgets.NewToolTipFilter(button.QWidget, 300, widgets.ToolTipPositionTop)
button.SetToolTip("Simple ToolTip")`

	codeToolTipLabel = `label := widgets.NewPixmapLabel(parent)
label.SetPixmap(pixmap)
label.SetFixedSize2(160, 160)
widgets.NewToolTipFilter(label.QWidget, 500, widgets.ToolTipPositionTop)
label.SetToolTip("Label with a ToolTip")
label.SetToolTipDuration(2000)`

	codeInfoBadge = `badge := widgets.NewInfoBadgeText("1", parent, widgets.InfoLevelInformation)
badge.SetCustomBackgroundColor(
	qt.NewQColor6("#005fb8"),
	qt.NewQColor6("#60cdff"),
)`

	codeInfoBar = `bar := widgets.NewInfoBar(
	widgets.InfoBarIconSuccess, "Success", "The Anthem of man is the Anthem of courage.",
	qt.Horizontal, true, -1, widgets.InfoBarPositionNone, parent,
)`

	codeInfoBarLongMessage = `bar := widgets.NewInfoBar(
	widgets.InfoBarIconWarning, "Warning", longContent,
	qt.Vertical, true, -1, widgets.InfoBarPositionNone, parent,
)`

	codeInfoBarCustom = `bar := widgets.NewInfoBar(
	common.Code, "GitHub", content,
	qt.Horizontal, true, -1, widgets.InfoBarPositionNone, parent,
)
bar.AddWidget(widgets.NewPushButtonText("Action", nil).QWidget, 0)
bar.SetCustomBackgroundColor(qt.NewQColor6("white"), qt.NewQColor6("#2a2a2a"))`

	codeInfoBarPosition = `bar := widgets.NewInfoBar(
	widgets.InfoBarIconInformation, "Lesson 3", content,
	qt.Horizontal, true, 2000, widgets.InfoBarPositionTopRight, parent,
)
bar.Show()`

	codeIndeterminateProgressBar = `bar := widgets.NewIndeterminateProgressBar(parent, true)
bar.SetFixedWidth(200)`

	codeProgressBar = `bar := widgets.NewProgressBar(parent, true)
bar.SetFixedWidth(200)

spinBox := widgets.NewSpinBox(parent)
spinBox.SetRange(0, 100)
spinBox.OnValueChanged(func(value int) { bar.SetValue(value) })`

	codeIndeterminateProgressRing = `ring := widgets.NewIndeterminateProgressRing(parent, true)
ring.SetFixedSize2(70, 70)`

	codeProgressRing = `ring := widgets.NewProgressRing(parent, true)
ring.SetFixedSize2(80, 80)
ring.SetTextVisible(true)

spinBox := widgets.NewSpinBox(parent)
spinBox.SetRange(0, 100)
spinBox.OnValueChanged(func(value int) { ring.SetValue(value) })`
)

// ---------------------------------------------------------------------------
// Text

const (
	codeLineEdit = `lineEdit := widgets.NewLineEdit(parent)
lineEdit.SetText("ko no dio da！")
lineEdit.SetClearButtonEnabled(true)`

	codeSearchLineEdit = `edit := widgets.NewSearchLineEdit(parent)
edit.SetPlaceholderText("Type a stand name")
edit.SetClearButtonEnabled(true)

completer := qt.NewQCompleter3(stands)
completer.SetCaseSensitivity(qt.CaseInsensitive)
edit.SetCompleter(completer)`

	codePasswordLineEdit = `edit := widgets.NewPasswordLineEdit(parent)
edit.SetPlaceholderText("Enter your password")
edit.SetFixedWidth(230)`

	codeSpinBox = `spinBox := widgets.NewSpinBox(parent)
spinBox.SetRange(0, 100)
spinBox.OnValueChanged(func(value int) { /* ... */ })`

	codeDoubleSpinBox = `spinBox := widgets.NewDoubleSpinBox(parent)
spinBox.SetRange(0, 100)
spinBox.OnValueChanged(func(value float64) { /* ... */ })`

	codeDateEdit = `dateEdit := widgets.NewDateEdit(parent)
dateEdit.OnDateChanged(func(date *qt.QDate) { /* ... */ })`

	codeTimeEdit = `timeEdit := widgets.NewTimeEdit(parent)
timeEdit.OnTimeChanged(func(t *qt.QTime) { /* ... */ })`

	codeDateTimeEdit = `dateTimeEdit := widgets.NewDateTimeEdit(parent)
dateTimeEdit.OnDateTimeChanged(func(dt *qt.QDateTime) { /* ... */ })`

	codeTextEdit = `textEdit := widgets.NewTextEdit(parent)
textEdit.SetFixedHeight(150)
textEdit.SetMarkdown("## Steel Ball Run \n * Johnny Joestar 🦄 \n * Gyro Zeppeli 🐴")`
)

// ---------------------------------------------------------------------------
// Views

const (
	codeListWidget = `listWidget := widgets.NewListWidget(parent)
listWidget.AddItem("白金之星")
listWidget.AddItem("绿色法皇")
listWidget.AddItem("天堂制造")`

	codeTableWidget = `table := widgets.NewTableWidget(parent)
table.SetBorderVisible(true)
table.SetBorderRadius(8)
table.VerticalHeader().Hide()

table.SetColumnCount(5)
table.SetRowCount(60)
table.SetHorizontalHeaderLabels([]string{"Title", "Artist", "Album", "Year", "Duration"})
table.SetItem(0, 0, qt.NewQTableWidgetItem2("かばん"))`

	codeTreeWidget = `tree := widgets.NewTreeWidget(parent)
tree.SetHeaderHidden(true)

item := qt.NewQTreeWidgetItem2([]string{"JoJo 1 - Phantom Blood"})
item.AddChildren([]*qt.QTreeWidgetItem{
	qt.NewQTreeWidgetItem2([]string{"Jonathan Joestar"}),
	qt.NewQTreeWidgetItem2([]string{"Dio Brando"}),
})
tree.AddTopLevelItem(item)
tree.ExpandAll()`

	codeTreeWidgetMultiSelect = `tree := widgets.NewTreeWidget(parent)
tree.SetHeaderHidden(true)
tree.SetSelectionMode(qt.QAbstractItemView__ExtendedSelection)

item := qt.NewQTreeWidgetItem2([]string{"JoJo 3 - Stardust Crusaders"})
child := qt.NewQTreeWidgetItem2([]string{"Jotaro Kujo"})
child.SetCheckState(0, qt.Unchecked)
item.AddChild(child)
tree.AddTopLevelItem(item)
tree.ExpandAll()`

	codeFileTable = `table := widgets.NewFileTable(parent)
table.SetColumnCount(4)
table.SetHorizontalHeaderLabels([]string{"名称", "修改日期", "类型", "大小"})
table.SetRowCount(300)
table.SetMaxColumn(3)

// Marquee selection (switchable) and the Explorer drag of the selected rows.
table.SetRubberBandEnabled(true)
table.SetItemDragHandler(func(rows []int, x, y int) { /* ... */ })`

	codeFlipView = `flipView := widgets.NewHorizontalFlipView(parent)
flipView.AddImages(images)

pager := widgets.NewHorizontalPipsPager(parent)
pager.SetPageNumber(flipView.Count())
pager.OnCurrentIndexChanged(func(index int) { flipView.SetCurrentIndex(index) })`
)

// ---------------------------------------------------------------------------
// Material (acrylic)

const codeAcrylicCard = `card := acrylic.NewAcrylicOpenGLWidget(parent)
card.SetGeometry(x, y, w, h)
card.SetTintColor("#3c8cff59")
card.SetCornerRadius(20)
card.SetBorderColor("#ffffff59")
card.SetBorderWidth(2)
card.SetCaption("Azure")
card.SetDraggable(true)`
