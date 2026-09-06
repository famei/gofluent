// Command slider migrates examples/basic_input/slider/demo.py.
//
// Demo1 demonstrates customising the slider sub-page colour through
// HollowHandleStyle. HollowHandleStyle (a QProxyStyle) is not ported to Go, so
// the equivalent QSS is applied to a native QSlider (documented degradation,
// see MIGRATION_GUIDE §10). Demo2 uses the fluent Slider widgets directly.
package main

import (
	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/demo"
	qt "github.com/mappu/miqt/qt"
)

func newDemo1() *qt.QWidget {
	w := qt.NewQWidget2()
	w.SetObjectName("Demo1")
	w.SetStyleSheet("#Demo1{background: rgb(184, 106, 106)}")
	w.Resize(300, 150)

	// HollowHandleStyle({"sub-page.color": QColor(70, 23, 180)}) approximated by
	// the equivalent Qt style sheet on a native QSlider. A bare sub-page rule has
	// no effect (the sub-page derives its height from the groove, which is not
	// defined), so the full groove/sub-page/add-page/handle geometry is spelled
	// out to match HollowHandleStyle: 3px groove, purple filled part, translucent
	// white unfilled part, and a hollow white ring handle.
	slider := qt.NewQSlider4(qt.Horizontal, w)
	slider.SetStyleSheet(`
QSlider::groove:horizontal {
    height: 3px;
    background: rgba(255, 255, 255, 64);
}
QSlider::sub-page:horizontal {
    height: 3px;
    background: rgb(70, 23, 180);
}
QSlider::add-page:horizontal {
    height: 3px;
    background: rgba(255, 255, 255, 64);
}
QSlider::handle:horizontal {
    width: 14px;
    height: 14px;
    margin: -6px 0;
    border-radius: 7px;
    background: white;
}`)

	slider.Resize(200, 28)
	slider.Move(50, 61)
	return w
}

func newDemo2() *qt.QWidget {
	w := qt.NewQWidget2()
	w.SetObjectName("Demo2")
	w.SetStyleSheet("#Demo2{background: white}")
	w.Resize(300, 300)

	slider1 := widgets.NewSliderOrientation(qt.Horizontal, w)
	slider1.SetFixedWidth(200)
	slider1.Move(50, 30)

	slider2 := widgets.NewSliderOrientation(qt.Vertical, w)
	slider2.SetFixedHeight(150)
	slider2.Move(140, 80)
	return w
}

func main() {
	demo.Run(newDemo1, newDemo2)
}
