package view

import (
	gcommon "github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	gallerycommon "github.com/famei/gofluent/examples/gallery/app/common"
	"github.com/famei/gofluent/examples/gallery/app/resource"
	qt "github.com/mappu/miqt/qt"
)

// ViewInterface is the "View" gallery page (port of app/view/view_interface.py).
type ViewInterface struct {
	*GalleryInterface
}

// NewViewInterface builds the view interface.
func NewViewInterface(parent *qt.QWidget) *ViewInterface {
	t := gallerycommon.NewTranslator()
	i := &ViewInterface{GalleryInterface: NewGalleryInterface(t.View, "qfluentwidgets.components.widgets", parent)}
	i.SetObjectName("viewInterface")

	tr := func(s string) string { return gallerycommon.Tr("ViewInterface", s) }

	i.AddExampleCard(tr("A simple ListView"), NewListFrame(i.QWidget).QWidget,
		"view/list_view/main.go", 0)
	i.AddExampleCard(tr("A simple TableView"), NewTableFrame(i.QWidget).QWidget,
		"view/table_view/main.go", 0)
	i.AddExampleCard(tr("A simple TreeView"), NewTreeFrame(i.QWidget, false).QWidget,
		"view/tree_view/main.go", 0)
	i.AddExampleCard(tr("A TreeView with Multi-selection enabled"), NewTreeFrame(i.QWidget, true).QWidget,
		"view/tree_view/main.go", 0)

	w := widgets.NewHorizontalFlipView(i.QWidget)
	w.AddImages([]interface{}{
		resource.Pixmap("Shoko1.jpg"),
		resource.Pixmap("Shoko2.jpg"),
		resource.Pixmap("Shoko3.jpg"),
		resource.Pixmap("Shoko4.jpg"),
	})
	i.AddExampleCard(tr("Flip view"), w.QWidget,
		"view/flip_view/main.go", 0)
	return i
}

// Frame is the shared base of the list/tree/table example frames.
type Frame struct {
	*qt.QFrame
	HBoxLayout *qt.QHBoxLayout
}

func newFrame(parent *qt.QWidget) *Frame {
	f := &Frame{QFrame: qt.NewQFrame(parent)}
	f.HBoxLayout = qt.NewQHBoxLayout(f.QWidget)
	f.HBoxLayout.SetContentsMargins(0, 8, 0, 0)
	f.SetObjectName("frame")
	gallerycommon.StyleViewInterface.Apply(f.QWidget, gcommon.ThemeAuto)
	return f
}

// AddWidget appends a widget to the frame layout.
func (f *Frame) AddWidget(widget *qt.QWidget) {
	f.HBoxLayout.AddWidget(widget)
}

// ListFrame shows a fluent list widget.
type ListFrame struct {
	*Frame
	listWidget *widgets.ListWidget
}

// NewListFrame builds a list frame.
func NewListFrame(parent *qt.QWidget) *ListFrame {
	f := &ListFrame{Frame: newFrame(parent)}
	f.listWidget = widgets.NewListWidget(f.QWidget)
	f.AddWidget(f.listWidget.QWidget)

	tr := func(s string) string { return gallerycommon.Tr("ListFrame", s) }
	stands := []string{
		tr("Star Platinum"), tr("Hierophant Green"),
		tr("Made in Haven"), tr("King Crimson"),
		tr("Silver Chariot"), tr("Crazy diamond"),
		tr("Metallica"), tr("Another One Bites The Dust"),
		tr("Heaven's Door"), tr("Killer Queen"),
		tr("The Grateful Dead"), tr("Stone Free"),
		tr("The World"), tr("Sticky Fingers"),
		tr("Ozone Baby"), tr("Love Love Deluxe"),
		tr("Hermit Purple"), tr("Gold Experience"),
		tr("King Nothing"), tr("Paper Moon King"),
		tr("Scary Monster"), tr("Mandom"),
		tr("20th Century Boy"), tr("Tusk Act 4"),
		tr("Ball Breaker"), tr("Sex Pistols"),
		tr("D4C • Love Train"), tr("Born This Way"),
		tr("SOFT & WET"), tr("Paisley Park"),
		tr("Wonder of U"), tr("Walking Heart"),
		tr("Cream Starter"), tr("November Rain"),
		tr("Smooth Operators"), tr("The Matte Kudasai"),
	}
	for _, stand := range stands {
		f.listWidget.AddItem(stand)
	}
	f.SetFixedSize2(300, 380)
	return f
}

// TreeFrame shows a fluent tree widget.
type TreeFrame struct {
	*Frame
	tree *widgets.TreeWidget
}

// NewTreeFrame builds a tree frame; enableCheck toggles check boxes.
func NewTreeFrame(parent *qt.QWidget, enableCheck bool) *TreeFrame {
	f := &TreeFrame{Frame: newFrame(parent)}
	f.tree = widgets.NewTreeWidget(f.QWidget)
	f.AddWidget(f.tree.QWidget)

	tr := func(s string) string { return gallerycommon.Tr("TreeFrame", s) }

	item1 := qt.NewQTreeWidgetItem2([]string{tr("JoJo 1 - Phantom Blood")})
	item1.AddChildren([]*qt.QTreeWidgetItem{
		qt.NewQTreeWidgetItem2([]string{tr("Jonathan Joestar")}),
		qt.NewQTreeWidgetItem2([]string{tr("Dio Brando")}),
		qt.NewQTreeWidgetItem2([]string{tr("Will A. Zeppeli")}),
	})
	f.tree.AddTopLevelItem(item1)

	item2 := qt.NewQTreeWidgetItem2([]string{tr("JoJo 3 - Stardust Crusaders")})
	item21 := qt.NewQTreeWidgetItem2([]string{tr("Jotaro Kujo")})
	item21.AddChildren([]*qt.QTreeWidgetItem{
		qt.NewQTreeWidgetItem2([]string{"空条承太郎"}),
		qt.NewQTreeWidgetItem2([]string{"空条蕉太狼"}),
		qt.NewQTreeWidgetItem2([]string{"阿强"}),
		qt.NewQTreeWidgetItem2([]string{"卖鱼强"}),
		qt.NewQTreeWidgetItem2([]string{"那个无敌的男人"}),
	})
	item2.AddChild(item21)
	f.tree.AddTopLevelItem(item2)
	f.tree.ExpandAll()
	f.tree.SetHeaderHidden(true)

	f.SetFixedSize2(300, 380)

	if enableCheck {
		setTreeItemsCheckable(f.tree)
	}
	return f
}

func setTreeItemsCheckable(tree *widgets.TreeWidget) {
	count := tree.TopLevelItemCount()
	for i := 0; i < count; i++ {
		item := tree.TopLevelItem(i)
		if item != nil {
			setItemCheckable(tree, item)
		}
	}
}

func setItemCheckable(tree *widgets.TreeWidget, item *qt.QTreeWidgetItem) {
	item.SetCheckState(0, qt.Unchecked)
	for i := 0; i < item.ChildCount(); i++ {
		child := item.Child(i)
		if child != nil {
			setItemCheckable(tree, child)
		}
	}
}

// TableFrame shows a fluent table widget.
type TableFrame struct {
	*widgets.TableWidget
}

// NewTableFrame builds a table frame.
func NewTableFrame(parent *qt.QWidget) *TableFrame {
	f := &TableFrame{TableWidget: widgets.NewTableWidget(parent)}

	f.VerticalHeader().Hide()
	f.SetBorderRadius(8)
	f.SetBorderVisible(true)

	f.SetColumnCount(5)
	f.SetRowCount(60)
	f.SetHorizontalHeaderLabels([]string{
		gallerycommon.Tr("TableFrame", "Title"),
		gallerycommon.Tr("TableFrame", "Artist"),
		gallerycommon.Tr("TableFrame", "Album"),
		gallerycommon.Tr("TableFrame", "Year"),
		gallerycommon.Tr("TableFrame", "Duration"),
	})

	songInfos := [][]string{
		{"かばん", "aiko", "かばん", "2004", "5:04"},
		{"爱你", "王心凌", "爱你", "2004", "3:39"},
		{"星のない世界", "aiko", "星のない世界/横顔", "2007", "5:30"},
		{"横顔", "aiko", "星のない世界/横顔", "2007", "5:06"},
		{"秘密", "aiko", "秘密", "2008", "6:27"},
		{"シアワセ", "aiko", "秘密", "2008", "5:25"},
		{"二人", "aiko", "二人", "2008", "5:00"},
		{"スパークル", "RADWIMPS", "君の名は。", "2016", "8:54"},
		{"なんでもないや", "RADWIMPS", "君の名は。", "2016", "3:16"},
		{"前前前世", "RADWIMPS", "人間開花", "2016", "4:35"},
		{"恋をしたのは", "aiko", "恋をしたのは", "2016", "6:02"},
		{"夏バテ", "aiko", "恋をしたのは", "2016", "4:41"},
		{"もっと", "aiko", "もっと", "2016", "4:50"},
		{"問題集", "aiko", "もっと", "2016", "4:18"},
		{"半袖", "aiko", "もっと", "2016", "5:50"},
		{"ひねくれ", "鎖那", "Hush a by little girl", "2017", "3:54"},
		{"シュテルン", "鎖那", "Hush a by little girl", "2017", "3:16"},
		{"愛は勝手", "aiko", "湿った夏の始まり", "2018", "5:31"},
		{"ドライブモード", "aiko", "湿った夏の始まり", "2018", "3:37"},
		{"うん。", "aiko", "湿った夏の始まり", "2018", "5:48"},
		{"キラキラ", "aikoの詩。", "2019", "5:08", "aiko"},
		{"恋のスーパーボール", "aiko", "aikoの詩。", "2019", "4:31"},
		{"磁石", "aiko", "どうしたって伝えられないから", "2021", "4:24"},
		{"食べた愛", "aiko", "食べた愛/あたしたち", "2021", "5:17"},
		{"列車", "aiko", "食べた愛/あたしたち", "2021", "4:18"},
		{"花の塔", "さユり", "花の塔", "2022", "4:35"},
		{"夏恋のライフ", "aiko", "夏恋のライフ", "2022", "5:03"},
		{"あかときリロード", "aiko", "あかときリロード", "2023", "4:04"},
		{"荒れた唇は恋を失くす", "aiko", "今の二人をお互いが見てる", "2023", "4:07"},
		{"ワンツースリー", "aiko", "今の二人をお互いが見てる", "2023", "4:47"},
	}
	songInfos = append(songInfos, songInfos...)
	for i, songInfo := range songInfos {
		for j := 0; j < 5; j++ {
			f.SetItem(i, j, qt.NewQTableWidgetItem2(songInfo[j]))
		}
	}

	f.SetFixedSize2(625, 440)
	f.ResizeColumnsToContents()
	return f
}
