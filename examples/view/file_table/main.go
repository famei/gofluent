// Command file_table demonstrates the Windows Explorer style widgets.FileTable:
// light/dark theme switching, the marquee (rubber band) drag selection with
// auto scrolling, the search highlight and the folder/document icons.
package main

import (
	"fmt"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
	"github.com/famei/gofluent/examples/internal/filetabledata"
	qt "github.com/mappu/miqt/qt"
)

// fileCount is the number of fake files listed after the demo folders.
const fileCount = 800

func newDemo() *qt.QWidget {
	w := qt.NewQWidget2()
	w.SetObjectName("FileTableDemo")

	vBoxLayout := qt.NewQVBoxLayout(w)
	vBoxLayout.SetContentsMargins(24, 20, 24, 20)
	vBoxLayout.SetSpacing(12)

	titleLabel := widgets.NewTitleLabelText("FileTable", w)
	vBoxLayout.AddWidget(titleLabel.QWidget)

	// The table is created before the toolbar so the controls can talk to it.
	table := widgets.NewFileTable(w)
	filetabledata.Fill(table, fileCount)

	// Shows the callback registered below.
	statusLabel := widgets.NewCaptionLabelText(
		"拖动已选中的行会调用 SetItemDragHandler，拖拽动作由调用方实现", w)

	toolBar := qt.NewQHBoxLayout2()
	toolBar.SetContentsMargins(0, 0, 0, 0)
	toolBar.SetSpacing(16)

	// 1. Marquee (rubber band) selection: drag with the left button to draw a
	//    selection rectangle, hold Ctrl/Shift to keep the current selection.
	marqueeSwitch := widgets.NewSwitchButtonText("拖拽框选", w, widgets.RIGHT)
	marqueeSwitch.SetChecked(table.IsRubberBandEnabled())
	marqueeSwitch.OnCheckedChanged(func(checked bool) {
		table.SetRubberBandEnabled(checked)
	})
	toolBar.AddWidget(marqueeSwitch.QWidget)

	// 2. Automatic scrolling while the marquee is held against the top/bottom
	//    edge: this is what makes the marquee usable with thousands of rows.
	scrollSwitch := widgets.NewSwitchButtonText("边缘自动滚动", w, widgets.RIGHT)
	scrollSwitch.SetChecked(table.IsRubberBandAutoScrollEnabled())
	scrollSwitch.OnCheckedChanged(func(checked bool) {
		table.SetRubberBandAutoScrollEnabled(checked)
	})
	toolBar.AddWidget(scrollSwitch.QWidget)

	// 3. Search highlight: matching rows are painted with the highlight color
	//    (they are not selected), which the delegate draws below the selection.
	searchLineEdit := widgets.NewSearchLineEdit(w)
	searchLineEdit.SetPlaceholderText("高亮名称包含…")
	searchLineEdit.SetFixedWidth(220)
	searchLineEdit.SetOnTextChangedExtra(func(text string) {
		table.SetHighlight(filetabledata.HighlightMatching(table, text))
	})
	searchLineEdit.ClearSignal = func() {
		table.SetHighlight(nil)
	}
	toolBar.AddWidget(searchLineEdit.QWidget)

	toolBar.AddStretch()
	selectAllButton := widgets.NewPushButtonText("全选", w)
	selectAllButton.OnClicked(func() { table.SelectAll() })
	toolBar.AddWidget(selectAllButton.QWidget)

	themeButton := widgets.NewPrimaryPushButtonText("切换主题", w)
	themeButton.OnClicked(func() { common.ToggleTheme(true, false) })
	toolBar.AddWidget(themeButton.QWidget)

	vBoxLayout.AddLayout2(toolBar.QLayout, 0)
	vBoxLayout.AddWidget2(table.QWidget, 1)
	vBoxLayout.AddWidget(statusLabel.QWidget)

	// Explorer semantics: dragging a selected row does not draw the marquee but
	// reports the drag to the application, which implements the drag and drop
	// itself (QDrag, a context menu, a file move, ...).
	table.SetItemDragHandler(func(rows []int, x, y int) {
		statusLabel.SetText(fmt.Sprintf("拖拽 %d 个已选中项目，光标屏幕坐标 (%d, %d)", len(rows), x, y))
	})

	// The table paints a transparent background, so the window behind it has to
	// follow the theme as well.
	applyBackground := func() {
		if common.IsDarkTheme() {
			w.SetStyleSheet("#FileTableDemo{background: rgb(32, 32, 32)}")
		} else {
			w.SetStyleSheet("#FileTableDemo{background: rgb(249, 249, 249)}")
		}
	}
	applyBackground()
	common.QConfigInstance.OnThemeChanged(func(common.Theme) { applyBackground() })

	w.Resize(1000, 680)
	return w
}

func main() {
	demo.Run(newDemo)
}
