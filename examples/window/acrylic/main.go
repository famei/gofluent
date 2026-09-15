// Command acrylic-demo shows draggable "Acrylic" (frosted-glass) cards in three
// separate windows, one per back end:
//
//   - AcrylicWidget:       CPU blur, a plain QWidget (child widgets welcome).
//   - AcrylicGLWidget:     GPU blur on a native QGLWidget — cheapest, but being a
//     native window it cannot host ordinary child widgets.
//   - AcrylicOpenGLWidget: GPU blur on a non-native QOpenGLWidget, same API, and
//     ordinary child widgets work normally.
//
// Each card uses a different tint colour and opacity, and every card can be
// dragged around with the mouse. The third window puts real child widgets inside
// its cards.
package main

import (
	"os"

	"github.com/mappu/miqt/qt"

	"github.com/famei/gofluent/acrylic"
)

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

func main() {
	qt.NewQApplication(os.Args)
	qt.QGuiApplication_SetApplicationDisplayName("Acrylic Cards Demo")

	cpuWin := makeWindow("AcrylicWidget - CPU blur (draggable cards)")
	gpuWin := makeWindow("AcrylicGLWidget - GPU blur, native QGLWidget (no child widgets)")
	oglWin := makeWindow("AcrylicOpenGLWidget - GPU blur + child widgets inside the cards")

	// Ordinary child widgets placed *under* the cards (created first, so they
	// sit lower in the z order). Acrylic cards blur them as well — this is the
	// "draw on top of other widgets" case.
	for _, win := range []*qt.QWidget{cpuWin, gpuWin, oglWin} {
		addStripes(win, 70, 110, 240, 150)
	}

	// Same call sequence for every back end: colours come from hex strings.
	for _, c := range cardSpecs("CPU - ") {
		card := acrylic.NewAcrylicWidget(cpuWin)
		configureCard(card.QWidget, c)
		card.SetTintColor(c.tint)
		card.SetCornerRadius(c.radius)
		card.SetBorderColor(c.border)
		card.SetBorderWidth(c.borderW)
		card.SetCaption(c.title)
		card.SetDraggable(true)
	}

	for _, c := range cardSpecs("GPU - ") {
		card := acrylic.NewAcrylicGLWidget(gpuWin)
		configureCard(card.QWidget, c)
		card.SetTintColor(c.tint)
		card.SetCornerRadius(c.radius)
		card.SetBorderColor(c.border)
		card.SetBorderWidth(c.borderW)
		card.SetCaption(c.title)
		card.SetDraggable(true)
	}

	// The non-native GPU back end is a normal widget, so its cards can carry
	// their own content just like the CPU ones.
	var oglCards []*acrylic.AcrylicOpenGLWidget
	for _, c := range cardSpecs("GL - ") {
		card := acrylic.NewAcrylicOpenGLWidget(oglWin)
		configureCard(card.QWidget, c)
		card.SetTintColor(c.tint)
		card.SetCornerRadius(c.radius)
		card.SetBorderColor(c.border)
		card.SetBorderWidth(c.borderW)
		card.SetCaption(c.title)
		card.SetDraggable(true)
		oglCards = append(oglCards, card)
	}

	s := qt.NewQVBoxLayout(oglCards[0].QWidget)
	s2 := qt.NewQPushButton5("WidgetTest", oglCards[0].QWidget)
	s.AddWidget(s2.QWidget)

	cpuWin.Show()
	gpuWin.Show()
	oglWin.Show()

	qt.QApplication_Exec()
}

// configureCard applies the geometry every back end shares.
func configureCard(w *qt.QWidget, c cardSpec) {
	w.SetGeometry(c.x, c.y, c.w, c.h)
}

// addStripes creates an ordinary (non-acrylic) child widget painting bright
// vertical bands. Acrylic cards drawn above it must blur those bands, proving
// that the backdrop also covers regular widgets — not just the window
// background.
func addStripes(parent *qt.QWidget, x, y, w, h int) {
	stripes := qt.NewQWidget(parent)
	stripes.SetGeometry(x, y, w, h)
	stripes.OnPaintEvent(func(super func(ev *qt.QPaintEvent), ev *qt.QPaintEvent) {
		p := qt.NewQPainter2(stripes.QPaintDevice)
		defer p.Delete()

		bg := qt.NewQColor11(20, 20, 30, 255)
		defer bg.Delete()
		p.FillRect5(0, 0, w, h, bg)

		const band = 36
		for i := 0; i < w; i += band * 2 {
			col := qt.NewQColor11(255, 220, 80, 255)
			p.FillRect5(i, 0, band, h, col)
			col.Delete()
		}
		p.End()
	})
}

// makeWindow creates a top-level window whose background is a colourful
// checkerboard, giving the acrylic cards something to blur.
func makeWindow(title string) *qt.QWidget {
	w := qt.NewQWidget2()
	w.SetWindowTitle(title)
	w.Resize(920, 520)

	w.OnPaintEvent(func(super func(ev *qt.QPaintEvent), ev *qt.QPaintEvent) {
		paintBackground(w)
	})

	s := qt.NewQVBoxLayout(w)
	s2 := qt.NewQPushButton5("123123", w)
	s.AddWidget(s2.QWidget)

	return w
}

// paintBackground draws a colourful checkerboard with a few bright shapes.
func paintBackground(w *qt.QWidget) {
	painter := qt.NewQPainter2(w.QPaintDevice)
	defer painter.Delete()

	cols := []int{0x2b2b45, 0x3a4a6b, 0x2f6f6f, 0x6b4a3a, 0x4a2b5a}
	const cell = 64
	width, height := w.Width(), w.Height()
	for y := 0; y*cell < height; y++ {
		for x := 0; x*cell < width; x++ {
			c := cols[(x+y)%len(cols)]
			color := qt.NewQColor3((c>>16)&0xff, (c>>8)&0xff, c&0xff)
			defer color.Delete()
			painter.FillRect5(x*cell, y*cell, cell, cell, color)
		}
	}

	accent := qt.NewQColor11(230, 120, 60, 255)
	defer accent.Delete()
	brush := qt.NewQBrush3(accent)
	defer brush.Delete()
	painter.SetBrush(brush)
	painter.SetPenWithStyle(qt.NoPen)

	r1 := qt.NewQRectF4(140, 140, 90, 90)
	defer r1.Delete()
	painter.DrawEllipse(r1)

	r2 := qt.NewQRectF4(560, 160, 110, 110)
	defer r2.Delete()
	painter.DrawEllipse(r2)

	r3 := qt.NewQRectF4(700, 120, 70, 70)
	defer r3.Delete()
	painter.DrawRect(r3)

	painter.End()
}
