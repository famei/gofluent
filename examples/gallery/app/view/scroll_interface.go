package view

import (
	"github.com/famei/gofluent/components/widgets"
	gallerycommon "github.com/famei/gofluent/examples/gallery/app/common"
	"github.com/famei/gofluent/examples/gallery/app/resource"
	qt "github.com/mappu/miqt/qt"
)

// ScrollInterface is the "Scrolling" gallery page (port of
// app/view/scroll_interface.py).
type ScrollInterface struct {
	*GalleryInterface
}

// NewScrollInterface builds the scroll interface.
func NewScrollInterface(parent *qt.QWidget) *ScrollInterface {
	t := gallerycommon.NewTranslator()
	i := &ScrollInterface{GalleryInterface: NewGalleryInterface(t.Scroll, "qfluentwidgets.components.widgets", parent)}
	i.SetObjectName("scrollInterface")

	tr := func(s string) string { return gallerycommon.Tr("ScrollInterface", s) }

	const src = "scroll/scroll_area/main.go"

	w := widgets.NewScrollArea(nil)
	label := widgets.NewImageLabelImage(resource.Pixmap("chidanta2.jpg"), i.QWidget)
	label.ScaledToWidth(775)
	label.SetBorderRadius(8, 8, 8, 8)
	w.HorizontalScrollBar().SetValue(0)
	w.SetWidget(label.QWidget)
	w.SetFixedSize2(775, 430)

	card := i.AddExampleCard(tr("Smooth scroll area"), w.QWidget, src, 0)
	widgets.NewToolTipFilter(card.Card.QWidget, 500, widgets.ToolTipPositionTop)
	card.Card.SetToolTip(tr("Chitanda Eru is too hot 🥵"))
	card.Card.SetToolTipDuration(2000)

	smooth := widgets.NewSmoothScrollArea(nil)
	label2 := widgets.NewImageLabelImage(resource.Pixmap("chidanta3.jpg"), i.QWidget)
	label2.SetBorderRadius(8, 8, 8, 8)
	smooth.SetWidget(label2.QWidget)
	smooth.SetFixedSize2(660, 540)

	card = i.AddExampleCard(tr("Smooth scroll area implemented by animation"), smooth.QWidget, src, 0)
	widgets.NewToolTipFilter(card.Card.QWidget, 500, widgets.ToolTipPositionTop)
	card.Card.SetToolTip(tr("Chitanda Eru is so hot 🥵🥵"))
	card.Card.SetToolTipDuration(2000)

	single := widgets.NewSingleDirectionScrollArea(i.QWidget, qt.Horizontal)
	label3 := widgets.NewImageLabelImage(resource.Pixmap("chidanta4.jpg"), i.QWidget)
	label3.SetBorderRadius(8, 8, 8, 8)
	single.SetWidget(label3.QWidget)
	single.SetFixedSize2(660, 498)

	card = i.AddExampleCard(tr("Single direction scroll scroll area"), single.QWidget, src, 0)
	widgets.NewToolTipFilter(card.Card.QWidget, 500, widgets.ToolTipPositionTop)
	card.Card.SetToolTip(tr("Chitanda Eru is so hot 🥵🥵🥵"))
	card.Card.SetToolTipDuration(2000)

	pager := widgets.NewHorizontalPipsPager(i.QWidget)
	pager.SetPageNumber(15)
	pager.SetPreviousButtonDisplayMode(widgets.PipsDisplayAlways)
	pager.SetNextButtonDisplayMode(widgets.PipsDisplayAlways)
	card = i.AddExampleCard(tr("Pips pager"), pager.QWidget,
		"scroll/pips_pager/main.go", 0)
	card.TopLayout.SetContentsMargins(12, 20, 12, 20)
	return i
}
