package view

import (
	"github.com/famei/gofluent/acrylic"
	gallerycommon "github.com/famei/gofluent/examples/gallery/app/common"
	"github.com/famei/gofluent/examples/gallery/app/resource"
	"github.com/mappu/miqt/qt"
)

type AcrylicInterface struct {
	*GalleryInterface
}

func NewAcrylicInterface(parent *qt.QWidget) *AcrylicInterface {
	t := gallerycommon.NewTranslator()
	i := &AcrylicInterface{GalleryInterface: NewGalleryInterface(t.Material, `github.com/famei/gofluent/acrylic`, parent)}
	i.SetObjectName("AcrylicInterface")
	//tr := func(s string) string { return gallerycommon.Tr("AcrylicInterface", s) }
	box := NewAcrylicLabel(i.QWidget)
	i.AddExampleCard("Acrylic", box, "acrylic/acrylic_opengl.go", codeAcrylicCard, 0)

	return i
}

func NewAcrylicLabel(parent *qt.QWidget) *qt.QWidget {
	w1 := qt.NewQLabel(parent)
	w1.SetMaximumSize2(787, 579)
	w1.SetMinimumSize2(787, 579)
	jpg := resource.Pixmap("chidanta.jpg")
	w1.OnPaintEvent(func(super func(param1 *qt.QPaintEvent), param1 *qt.QPaintEvent) {
		painter := qt.NewQPainter2(w1.QPaintDevice)
		painter.SetRenderHints(qt.QPainter__Antialiasing | qt.QPainter__SmoothPixmapTransform)
		painter.DrawPixmap10(w1.Rect(), jpg)
		painter.End()
		painter.Delete()
	})
	var CardList []*acrylic.AcrylicOpenGLWidget
	for _, c := range cardSpecs("GL - ") {
		card := acrylic.NewAcrylicOpenGLWidget(w1.QWidget)
		configureCard(card.QWidget, c)
		card.SetTintColor(c.tint)
		card.SetCornerRadius(c.radius)
		card.SetBorderColor(c.border)
		card.SetBorderWidth(c.borderW)
		card.SetCaption(c.title)
		card.SetDraggable(true)
		CardList = append(CardList, card)

		// No OnUpdated -> RefreshWindow wiring here on purpose. Every card
		// re-captures its backdrop whenever Qt paints it, and raising a card to
		// drag it invalidates the cards underneath for us. Refreshing the whole
		// window from every callback instead re-captures all four cards on every
		// mouse move, which is what made dragging stutter (measured: ~40 ms per
		// drag step instead of ~14 ms).
	}
	acrylic.RefreshAll()
	return w1.QWidget
}

func configureCard(w *qt.QWidget, c cardSpec) {
	w.SetGeometry(c.x, c.y, c.w, c.h)
}

// cardSpec describes one acrylic card. tint and border are hex strings, e.g.
// "#3c8cff59" (blue at 35% alpha), so a single call configures colour+opacity.
type cardSpec struct {
	title   string
	x, y    int
	w, h    int
	tint    string // e.g. "#RRGGBB" or "#RRGGBBAA"
	radius  int    // corner radius in px (0 = square)
	border  string // border colour, e.g. "#ffffff66"
	borderW int    // border width in px (0 = none)
}

func cardSpecs(prefix string) []cardSpec {
	return []cardSpec{
		{prefix + "Azure", 40, 80, 260, 180, "#3c8cff59", 20, "#ffffff59", 2},
		{prefix + "Rose", 320, 80, 260, 180, "#ff5a5a80", 20, "#ffffff59", 2},
		{prefix + "Mint", 600, 80, 260, 180, "#46c88240", 0, "#ffffff59", 2},
		{prefix + "Amber", 180, 300, 260, 180, "#ffb43c99", 40, "#ffd28c99", 3},
	}
}
