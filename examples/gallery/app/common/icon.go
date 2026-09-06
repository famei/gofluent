package common

import (
	gcommon "github.com/famei/gofluent/common"
	"github.com/famei/gofluent/examples/gallery/app/resource"
	qt "github.com/mappu/miqt/qt"
	"github.com/mappu/miqt/qt/svg"
)

// Icon enumerates the five gallery icons (port of app/common/icon.py). It
// implements gcommon.FluentIconBase so it can be passed to IconWidget, actions
// and navigation items.
type Icon string

const (
	IconGrid            Icon = "Grid"
	IconMenu            Icon = "Menu"
	IconText            Icon = "Text"
	IconPrice           Icon = "Price"
	IconEmojiTabSymbols Icon = "EmojiTabSymbols"
)

// Name returns the icon name.
func (i Icon) Name() string { return string(i) }

// Path returns the logical Qt resource path of the icon.
func (i Icon) Path(theme gcommon.Theme) string {
	return ":/gallery/images/icons/" + string(i) + "_" + gcommon.GetIconColor(theme, false) + ".svg"
}

func (i Icon) svgBytes(theme gcommon.Theme) []byte {
	return resource.IconSVG(string(i), gcommon.GetIconColor(theme, false))
}

// svgToIcon renders embedded SVG bytes into a QIcon backed by a 128px pixmap.
func svgToIcon(svgBytes []byte) *qt.QIcon {
	const size = 128
	pm := qt.NewQPixmap2(size, size)
	transparent := qt.NewQColor11(0, 0, 0, 0)
	pm.FillWithFillColor(transparent)
	transparent.Delete()

	painter := qt.NewQPainter2(pm.QPaintDevice)
	renderer := svg.NewQSvgRenderer3(svgBytes)
	renderer.Render2(painter, qt.NewQRectF4(0, 0, size, size))
	renderer.Delete()
	painter.End()
	painter.Delete()

	return qt.NewQIcon2(pm)
}

// Icon returns a QIcon rendered from the embedded SVG.
func (i Icon) Icon(theme gcommon.Theme) *qt.QIcon { return svgToIcon(i.svgBytes(theme)) }

// Render draws the icon into a painter.
func (i Icon) Render(painter *qt.QPainter, rect *qt.QRectF, theme gcommon.Theme) {
	gcommon.DrawSvgIcon(i.svgBytes(theme), painter, rect)
}

// Colored returns nil: the gallery icons are never recolored.
func (i Icon) Colored(light, dark *qt.QColor) *gcommon.ColoredFluentIcon { return nil }

// QIcon returns a theme-following QIcon (reverse swaps black/white).
func (i Icon) QIcon(reverse bool) *qt.QIcon {
	return svgToIcon(resource.IconSVG(string(i), gcommon.GetIconColor(gcommon.ThemeAuto, reverse)))
}
