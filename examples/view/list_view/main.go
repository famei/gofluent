// Command list_view migrates examples/view/list_view/demo.py.
package main

import (
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
	qt "github.com/mappu/miqt/qt"
)

func newDemo() *qt.QWidget {
	w := qt.NewQWidget2()
	w.SetObjectName("Demo")
	w.SetStyleSheet("#Demo{background: rgb(249, 249, 249)}")

	hBoxLayout := qt.NewQHBoxLayout(w)
	listWidget := widgets.NewListWidget(w)

	stands := []string{
		"白金之星", "绿色法皇", "天堂制造", "绯红之王",
		"银色战车", "疯狂钻石", "壮烈成仁", "败者食尘",
		"黑蚊子多", "杀手皇后", "金属制品", "石之自由",
		"砸瓦鲁多", "钢链手指", "臭氧宝宝", "华丽挚爱",
		"隐者之紫", "黄金体验", "虚无之王", "纸月之王",
		"骇人恶兽", "男子领域", "20世纪男孩", "牙 Act 4",
		"铁球破坏者", "性感手枪", "D4C • 爱之列车", "天生完美",
		"软又湿", "佩斯利公园", "奇迹于你", "行走的心",
		"护霜旅行者", "十一月雨", "调情圣手", "片刻静候",
	}
	for _, stand := range stands {
		listWidget.AddItem(stand)
	}

	hBoxLayout.SetContentsMargins(0, 0, 0, 0)
	hBoxLayout.AddWidget(listWidget.QWidget)
	w.Resize(300, 400)
	return w
}

func main() {
	demo.Run(newDemo)
}
