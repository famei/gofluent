package view

import (
	"strings"

	"github.com/famei/gofluent/examples"
)

// snippetByPath maps an example source path (the value passed to AddExampleCard)
// to a short 3-8 line call pseudo-code snippet. Every line uses the real
// gofluent Go API — the constructors and methods are taken verbatim from each
// example's main.go — so the snippet stays a faithful, copyable summary of how
// the control is used instead of the previous full-source dump.
var snippetByPath = map[string]string{
	"basic_input/button/main.go": `button := widgets.NewPushButtonText("Standard push button", nil)
button.OnClicked(func() { /* ... */ })
toggle := widgets.NewToggleToolButtonIcon(common.Setting, nil)
toggle.OnToggled(func(bool) { /* ... */ })`,

	"basic_input/check_box/main.go": `checkBox := widgets.NewCheckBoxText("This is a check box", parent)
checkBox.SetTristate()
layout.AddWidget3(checkBox.QWidget, 1, qt.AlignCenter)`,

	"basic_input/combo_box/main.go": `comboBox := widgets.NewComboBox(parent)
comboBox.AddItems([]string{"shoko 🥰", "西宫硝子"})
comboBox.SetCurrentIndex(0)
comboBox.OnCurrentTextChanged = func(text string) { /* ... */ }`,

	"basic_input/radio_button/main.go": `button1 := widgets.NewRadioButtonText("Option 1", parent)
button2 := widgets.NewRadioButtonText("Option 2", parent)
group := qt.NewQButtonGroup2(parent.QObject)
group.AddButton(button1.QAbstractButton)
group.AddButton(button2.QAbstractButton)`,

	"basic_input/slider/main.go": `slider := widgets.NewSliderOrientation(qt.Horizontal, parent)
slider.SetRange(0, 100)
slider.SetValue(30)`,

	"basic_input/switch_button/main.go": `switchButton := widgets.NewSwitchButton(parent, widgets.RIGHT)
switchButton.SetText("Off")
switchButton.OnCheckedChanged(func(isChecked bool) { /* ... */ })`,

	"date_time/calendar_picker/main.go": `picker := date_time.NewCalendarPicker(parent)
picker.OnDateChanged = func(date *qt.QDate) { /* ... */ }
layout.AddWidget3(picker.QWidget, 0, qt.AlignCenter)`,

	"date_time/time_picker/main.go": `datePicker := date_time.NewDatePicker(parent, date_time.DatePickerYYYYMMDD, false)
timePicker := date_time.NewTimePicker(parent, false)
timePicker.OnTimeChanged = func(t *qt.QTime) { /* ... */ }`,

	"dialog_flyout/dialog/main.go": `dialog := dialog_box.NewDialog(title, content, parent)
if dialog.Exec() == int(qt.QDialog__Accepted) {
    // user confirmed
}`,

	"dialog_flyout/message_dialog/main.go": `box := dialog_box.NewMessageBox(title, content, parent)
box.SetClosableOnMaskClicked(true)
box.SetDraggable(true)
box.Exec()`,

	"dialog_flyout/custom_message_box/main.go": `box := dialog_box.NewMessageBoxBase(parent)
box.ValidateFunc = func() bool { return valid }
if box.Exec() == int(qt.QDialog__Accepted) { /* ... */ }`,

	"dialog_flyout/color_dialog/main.go": `color := qt.NewQColor6("#5012aaa2")
button := settings.NewColorPickerButton(color, "Background Color", parent, true)
button.Move(352, 312)`,

	"dialog_flyout/flyout/main.go": `view := widgets.NewFlyoutView("Lesson 4", content, nil, image, true, nil)
view.SetOnClosed(func() { flyout.Close() })
flyout := widgets.FlyoutMake(view.FlyoutViewBase, target, parent, widgets.FlyoutAnimationPullUp, false)`,

	"dialog_flyout/teaching_tip/main.go": `view := widgets.NewTeachingTipView("Lesson 5", content, nil, image, true, widgets.TeachingTipTailBottom, nil)
tip := widgets.TeachingTipMake(view.FlyoutViewBase, target, -1, widgets.TeachingTipTailBottom, parent, false)
view.SetOnClosed(func() { tip.Close() })`,

	"layout/flow_layout/main.go": `flow := layout.NewFlowLayout(parent, true, false)
flow.SetAnimation(250)
flow.AddWidget(widgets.NewPushButtonText("aiko", nil).QWidget)
flow.AddWidget(widgets.NewPushButtonText("刘静爱", nil).QWidget)`,

	"menu/menu/main.go": `menu := widgets.NewRoundMenu("", parent)
menu.AddAction(common.NewActionFluentIcon(common.Copy, "Copy", nil).QAction)
submenu := widgets.NewRoundMenu("Add to", parent)
menu.AddMenu(submenu)
menu.Exec(pos, widgets.MenuAnimationDropDown)`,

	"menu/widget_menu/main.go": `menu := widgets.NewRoundMenu("", parent)
menu.AddWidget(profileCard.QWidget)
menu.AddSeparator()
menu.AddAction(common.NewActionFluentIcon(common.Setting, "设置", nil).QAction)
menu.Exec(pos, widgets.MenuAnimationDropDown)`,

	"menu/command_bar/main.go": `bar := widgets.NewCommandBar(parent)
bar.AddAction(common.NewActionFluentIcon(common.Copy, "Copy", nil).QAction)
bar.AddSeparator()
bar.AddHiddenAction(common.NewActionFluentIcon(common.Setting, "Settings", nil).QAction)`,

	"navigation/breadcrumb_bar/main.go": `bar := navigation.NewBreadcrumbBar(parent)
bar.AddItem(objectName, text)
bar.OnCurrentItemChanged(func(objectName string) { /* ... */ })`,

	"navigation/pivot/main.go": `pivot := navigation.NewPivot(parent)
pivot.AddItem(routeKey, text, nil, nil)
pivot.OnCurrentItemChanged(func(k string) { /* ... */ })`,

	"navigation/segmented_widget/main.go": `seg := navigation.NewSegmentedWidget(parent)
seg.AddItem(routeKey, text, nil, nil)
seg.OnCurrentItemChanged(func(k string) { /* ... */ })`,

	"navigation/segmented_tool_widget/main.go": `seg := navigation.NewSegmentedToggleToolWidget(parent)
seg.AddItem(routeKey, common.Music, nil)
seg.OnCurrentItemChanged(func(k string) { /* ... */ })`,

	"navigation/tab_view/main.go": `tabBar := widgets.NewTabBar(parent)
tabBar.SetMovable(true)
tabBar.AddTab(routeKey, text, icon, nil)
tabBar.OnCurrentChanged(func(index int) { /* ... */ })`,

	"scroll/scroll_area/main.go": `area := widgets.NewSmoothScrollArea(nil)
area.SetScrollAnimation(qt.Vertical, 400)
area.SetWidget(label.QWidget)`,

	"scroll/pips_pager/main.go": `pager := widgets.NewHorizontalPipsPager(parent)
pager.SetPageNumber(15)
pager.SetVisibleNumber(8)`,

	"status_info/state_tool_tip/main.go": `tip := widgets.NewStateToolTip("正在训练模型", "客官请耐心等待哦~~", parent)
tip.Move(510, 30)
tip.Show()
tip.SetState(true)`,

	"status_info/tool_tip/main.go": `button.SetToolTip("aiko - キラキラ ✨")
button.SetToolTipDuration(1000)
widgets.NewToolTipFilter(button.QWidget, 300, widgets.ToolTipPositionTop)`,

	"status_info/info_bar/main.go": `bar := widgets.NewInfoBar(widgets.InfoBarIconSuccess, "Lesson 4", content, qt.Horizontal, true, 2000, widgets.InfoBarPositionTop, parent)
bar.AddWidget(widgets.NewPushButtonText("Action", nil).QWidget, 0)
bar.Show()`,

	"status_info/progress_bar/main.go": `bar := widgets.NewProgressBar(parent, true)
bar.SetValue(50)
bar.Pause()
bar.Resume()`,

	"status_info/progress_ring/main.go": `ring := widgets.NewProgressRing(parent, true)
ring.SetValue(50)
ring.SetTextVisible(true)
ring.Pause()
ring.Resume()`,

	"text/line_edit/main.go": `edit := widgets.NewSearchLineEdit(parent)
edit.SetClearButtonEnabled(true)
edit.SetPlaceholderText("Search stand")
completer := qt.NewQCompleter6(stands, edit.QObject)
edit.SetCompleter(completer)`,

	"text/spin_box/main.go": `spinBox := widgets.NewSpinBox(parent)
spinBox.SetAccelerated(true)
doubleSpinBox := widgets.NewDoubleSpinBox(parent)
dateEdit := widgets.NewDateEdit(parent)
timeEdit := widgets.NewTimeEdit(parent)
dateTimeEdit := widgets.NewDateTimeEdit(parent)`,

	"view/list_view/main.go": `listWidget := widgets.NewListWidget(parent)
listWidget.AddItem("白金之星")
listWidget.AddItem("绿色法皇")`,

	"view/table_view/main.go": `table := widgets.NewTableWidget(parent)
table.SetRowCount(60)
table.SetColumnCount(5)
table.SetItem(0, 0, qt.NewQTableWidgetItem2("Title"))
table.SetHorizontalHeaderLabels([]string{"Title", "Artist"})`,

	"view/tree_view/main.go": `view := widgets.NewTreeView(parent)
model := qt.NewQFileSystemModel()
model.SetRootPath(".")
view.SetModel(model.QAbstractItemModel)
view.SetBorderVisible(true)`,

	"view/flip_view/main.go": `flipView := widgets.NewHorizontalFlipView(parent)
pager := widgets.NewHorizontalPipsPager(parent)
flipView.AddImages(images)
pager.SetPageNumber(flipView.Count())`,
	"acrylic/acrylic_opengl.go": `acrylic := acrylic.NewAcrylicOpenGLWidget(parent)
acrylic.SetTintColor("#3c8cff59")`,
}

// codeSnippet returns the 3-8 line pseudo-code snippet for the given source
// path. Unknown paths fall back to deriving a short snippet from the embedded
// example source, so the gallery never shows the old full main.go dump.
func codeSnippet(sourcePath string) string {
	if s, ok := snippetByPath[sourcePath]; ok {
		return s
	}
	if src, ok := examples.Source(sourcePath); ok {
		return derivedSnippet(src)
	}
	return "// source not found: " + sourcePath
}

// derivedSnippet extracts a handful of representative API-call lines from an
// embedded example source. It is only a safety net for paths without a curated
// entry above.
func derivedSnippet(src string) string {
	var picked []string
	for _, ln := range strings.Split(src, "\n") {
		t := strings.TrimSpace(ln)
		if t == "" || strings.HasPrefix(t, "//") || strings.HasPrefix(t, "/*") || strings.HasPrefix(t, "*") {
			continue
		}
		if isAPILine(t) {
			picked = append(picked, t)
			if len(picked) >= 6 {
				break
			}
		}
	}
	if len(picked) < 3 {
		picked = nil
		for _, ln := range strings.Split(src, "\n") {
			t := strings.TrimSpace(ln)
			if t == "" || strings.HasPrefix(t, "//") {
				continue
			}
			picked = append(picked, t)
			if len(picked) >= 4 {
				break
			}
		}
	}
	return strings.Join(picked, "\n")
}

// isAPILine reports whether a source line constructs or configures a widget.
func isAPILine(s string) bool {
	for _, p := range []string{
		"widgets.New", "common.New", "qt.New", "navigation.New",
		"date_time.New", "dialog_box.New", "layout.New", "settings.New",
	} {
		if strings.Contains(s, p) {
			return true
		}
	}
	for _, m := range []string{
		".OnClicked", ".OnToggled", ".OnCheckedChanged", ".OnCurrentTextChanged",
		".OnCurrentItemChanged", ".OnCurrentChanged", ".Set", ".Add", ".Exec",
		".Show", ".Move", ".Insert", ".Resize", ".Toggle", ".Pause", ".Resume",
	} {
		if strings.Contains(s, m) {
			return true
		}
	}
	return false
}
