package view

import (
	"github.com/famei/gofluent/components/layout"
	"github.com/famei/gofluent/components/widgets"
	gallerycommon "github.com/famei/gofluent/examples/gallery/app/common"
	qt "github.com/mappu/miqt/qt"
)

// LayoutInterface is the "Layout" gallery page (port of
// app/view/layout_interface.py).
type LayoutInterface struct {
	*GalleryInterface
}

// NewLayoutInterface builds the layout interface.
func NewLayoutInterface(parent *qt.QWidget) *LayoutInterface {
	t := gallerycommon.NewTranslator()
	i := &LayoutInterface{GalleryInterface: NewGalleryInterface(t.Layout, "github.com/famei/gofluent/components/layout", parent)}
	i.SetObjectName("layoutInterface")

	tr := func(s string) string { return gallerycommon.Tr("LayoutInterface", s) }

	i.AddExampleCard(
		tr("Flow layout without animation"),
		i.createWidget(false),
		"layout/flow_layout/main.go",
		1,
	)

	i.AddExampleCard(
		tr("Flow layout with animation"),
		i.createWidget(true),
		"layout/flow_layout/main.go",
		1,
	)
	return i
}

func (i *LayoutInterface) createWidget(animation bool) *qt.QWidget {
	texts := []string{
		gallerycommon.Tr("LayoutInterface", "Star Platinum"),
		gallerycommon.Tr("LayoutInterface", "Hierophant Green"),
		gallerycommon.Tr("LayoutInterface", "Silver Chariot"),
		gallerycommon.Tr("LayoutInterface", "Crazy diamond"),
		gallerycommon.Tr("LayoutInterface", "Heaven's Door"),
		gallerycommon.Tr("LayoutInterface", "Killer Queen"),
		gallerycommon.Tr("LayoutInterface", "Gold Experience"),
		gallerycommon.Tr("LayoutInterface", "Sticky Fingers"),
		gallerycommon.Tr("LayoutInterface", "Sex Pistols"),
		gallerycommon.Tr("LayoutInterface", "Dirty Deeds Done Dirt Cheap"),
	}

	widget := qt.NewQWidget2()
	fl := layout.NewFlowLayout(widget, animation, false)

	fl.SetContentsMargins(0, 0, 0, 0)
	fl.SetVerticalSpacing(20)
	fl.SetHorizontalSpacing(10)

	for _, text := range texts {
		fl.AddWidget(widgets.NewPushButtonText(text, widget).QWidget)
	}
	return widget
}
