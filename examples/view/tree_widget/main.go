// Command tree_widget migrates examples/view/tree_widget/demo.py.
package main

import (
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
	qt "github.com/mappu/miqt/qt"
)

func jojoChildren() []*qt.QTreeWidgetItem {
	names := []string{"空条承太郎", "空条蕉太狼", "阿强", "卖鱼强", "那个无敌的男人"}
	children := make([]*qt.QTreeWidgetItem, 0, len(names)*10)
	for i := 0; i < 10; i++ {
		for _, name := range names {
			children = append(children, qt.NewQTreeWidgetItem2([]string{name}))
		}
	}
	return children
}

func newDemo() *qt.QWidget {
	w := qt.NewQWidget2()

	tree := widgets.NewTreeWidget(w)
	hBoxLayout := qt.NewQHBoxLayout(w)
	hBoxLayout.SetContentsMargins(0, 8, 0, 0)
	hBoxLayout.AddWidget(tree.QWidget)

	item1 := qt.NewQTreeWidgetItem2([]string{"JoJo 1 - Phantom Blood"})
	item1.AddChildren([]*qt.QTreeWidgetItem{
		qt.NewQTreeWidgetItem2([]string{"Jonathan Joestar"}),
		qt.NewQTreeWidgetItem2([]string{"Dio Brando"}),
		qt.NewQTreeWidgetItem2([]string{"Will A. Zeppeli"}),
	})
	tree.AddTopLevelItem(item1)

	item2 := qt.NewQTreeWidgetItem2([]string{"JoJo 3 - Stardust Crusaders"})
	item21 := qt.NewQTreeWidgetItem2([]string{"Jotaro Kujo"})
	item21.AddChildren(jojoChildren())
	item22 := qt.NewQTreeWidgetItem2([]string{"Jotaro Kujo"})
	item22.AddChildren(jojoChildren())
	item23 := qt.NewQTreeWidgetItem2([]string{"Jotaro Kujo"})
	item23.AddChildren(jojoChildren())
	item2.AddChild(item21)
	item2.AddChild(item22)
	item2.AddChild(item23)

	tree.AddTopLevelItem(item2)
	tree.ExpandAll()
	tree.SetHeaderHidden(true)

	w.Resize(400, 500)
	return w
}

func main() {
	demo.Run(newDemo)
}
