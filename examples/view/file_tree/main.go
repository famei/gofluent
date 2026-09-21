// Command file_tree demos the Windows Explorer style widgets.FileTree: the same
// row states as the FileTable (hover, selection outline, highlight), the
// light/dark theming of the rows, the text, the Explorer glyphs and the expander
// chevron. The tree is driven by the local file system through SetModel, so the
// folders are loaded on demand and can be browsed like the Explorer navigation
// pane.
package main

import (
	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
	qt "github.com/mappu/miqt/qt"
)

func newDemo() *qt.QWidget {
	w := qt.NewQWidget2()
	w.SetObjectName("FileTreeDemo")

	vBoxLayout := qt.NewQVBoxLayout(w)
	vBoxLayout.SetContentsMargins(24, 20, 24, 20)
	vBoxLayout.SetSpacing(12)

	titleLabel := widgets.NewTitleLabelText("FileTree", w)
	vBoxLayout.AddWidget(titleLabel.QWidget)

	captionLabel := widgets.NewCaptionLabelText(
		"行悬停/选中/高亮与 FileTable 一致，箭头与图标随主题切换；数据来自本地文件系统", w)
	vBoxLayout.AddWidget(captionLabel.QWidget)

	tree := widgets.NewFileTree(w)
	// QFileSystemModel walks the local drives and loads every folder on demand,
	// which is what makes the tree usable on a real file system.
	model := qt.NewQFileSystemModel2(w.QObject)
	model.SetRootPath(qt.QDir_RootPath())
	tree.SetModel(model.QAbstractItemModel)

	statusLabel := widgets.NewCaptionLabelText("选中一个驱动器、文件夹或文件", w)
	tree.SelectionModel().OnCurrentChanged(func(current *qt.QModelIndex, previous *qt.QModelIndex) {
		if current == nil || !current.IsValid() {
			statusLabel.SetText("选中一个驱动器、文件夹或文件")
			return
		}
		statusLabel.SetText("选中：" + model.FilePath(current))
	})

	toolBar := qt.NewQHBoxLayout2()
	toolBar.SetContentsMargins(0, 0, 0, 0)
	toolBar.SetSpacing(12)

	// Expanding every folder of a drive would walk the whole disk, so the toolbar
	// expands the drives only. miqt's model accessors dereference the parent
	// index, so the root is an invalid index rather than nil (the index lives as
	// long as the window does).
	root := qt.NewQModelIndex()
	expandButton := widgets.NewPushButtonText("展开驱动器", w)
	expandButton.OnClicked(func() {
		for row := 0; row < model.RowCount(root); row++ {
			tree.Expand(model.Index(row, 0, root))
		}
	})
	toolBar.AddWidget(expandButton.QWidget)

	collapseButton := widgets.NewPushButtonText("全部折叠", w)
	collapseButton.OnClicked(func() { tree.CollapseAll() })
	toolBar.AddWidget(collapseButton.QWidget)

	// The row height and the size of the expander chevron are settable.
	rowHeightLabel := widgets.NewBodyLabelText("行高", w)
	toolBar.AddWidget(rowHeightLabel.QWidget)

	rowHeightBox := widgets.NewSpinBox(w)
	rowHeightBox.SetRange(16, 40)
	rowHeightBox.SetSuffix(" px")
	rowHeightBox.SetValue(tree.RowHeight())
	rowHeightBox.OnValueChanged(func(value int) { tree.SetRowHeight(value) })
	toolBar.AddWidget(rowHeightBox.QWidget)

	chevronLabel := widgets.NewBodyLabelText("箭头", w)
	toolBar.AddWidget(chevronLabel.QWidget)

	chevronBox := widgets.NewSpinBox(w)
	// The arrow is right aligned in the 19px indentation of its level, so a very
	// large one reaches into the column of its parent.
	chevronBox.SetRange(8, 16)
	chevronBox.SetSuffix(" px")
	chevronBox.SetValue(tree.ChevronSize())
	chevronBox.OnValueChanged(func(value int) { tree.SetChevronSize(value) })
	toolBar.AddWidget(chevronBox.QWidget)

	toolBar.AddStretch()
	themeButton := widgets.NewPrimaryPushButtonText("切换主题", w)
	themeButton.OnClicked(func() { common.ToggleTheme(true, false) })
	toolBar.AddWidget(themeButton.QWidget)

	vBoxLayout.AddLayout2(toolBar.QLayout, 0)
	vBoxLayout.AddWidget2(tree.QWidget, 1)
	vBoxLayout.AddWidget(statusLabel.QWidget)

	// The tree paints a transparent background, so the window behind it has to
	// follow the theme as well.
	applyBackground := func() {
		if common.IsDarkTheme() {
			w.SetStyleSheet("#FileTreeDemo{background: rgb(32, 32, 32)}")
		} else {
			w.SetStyleSheet("#FileTreeDemo{background: rgb(249, 249, 249)}")
		}
	}
	applyBackground()
	common.QConfigInstance.OnThemeChanged(func(common.Theme) { applyBackground() })

	w.Resize(760, 720)
	return w
}

func main() {
	demo.Run(newDemo)
}
