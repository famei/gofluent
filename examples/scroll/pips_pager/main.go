// Command pips_pager migrates examples/scroll/pips_pager/demo.py.
package main

import (
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
	qt "github.com/mappu/miqt/qt"
)

func newDemo() *qt.QWidget {
	w := qt.NewQWidget2()

	vPager := widgets.NewVerticalPipsPager(w)
	hPager := widgets.NewHorizontalPipsPager(w)

	hPager.SetPageNumber(15)
	vPager.SetPageNumber(15)

	hPager.SetVisibleNumber(8)
	hPager.SetNextButtonDisplayMode(widgets.PipsDisplayAlways)
	hPager.SetPreviousButtonDisplayMode(widgets.PipsDisplayAlways)

	vPager.SetNextButtonDisplayMode(widgets.PipsDisplayAlways)
	vPager.SetPreviousButtonDisplayMode(widgets.PipsDisplayOnHover)

	layout := qt.NewQHBoxLayout(w)
	layout.AddWidget(hPager.QWidget)
	layout.AddWidget(vPager.QWidget)

	w.Resize(500, 500)
	return w
}

func main() {
	demo.Run(newDemo)
}
