package widgets

import (
	"io/fs"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/resources"
	qt "github.com/mappu/miqt/qt"
)

// renderFluentIcon draws a fluent icon into painter at rect using the given
// theme (ThemeAuto follows the current theme). Every FluentIconBase honours the
// theme, which is what lets a caller ask for the reversed color (e.g. a checked
// command-bar button drawn on the accent background); other sources are handed
// to common.DrawIcon, which paints a pre-rendered pixmap as it is.
func renderFluentIcon(icon interface{}, painter *qt.QPainter, rect *qt.QRectF, theme common.Theme) {
	switch v := icon.(type) {
	case common.FluentIconBase:
		v.Render(painter, rect, theme)
	default:
		common.DrawIcon(icon, painter, rect)
	}
}

// renderFluentIconWithFill draws a fluent icon in a fixed color instead of the
// theme foreground (used for the accent buttons' glyphs and the grey scroll-bar
// arrows).
func renderFluentIconWithFill(icon common.FluentIconBase, painter *qt.QPainter, rect *qt.QRectF, fill string) {
	if icon == nil {
		return
	}
	if f, ok := icon.(common.SegoeFluentIcon); ok {
		color := qt.NewQColor6(fill)
		defer color.Delete()
		f.RenderGlyph(painter, rect, common.ThemeAuto, color)
		return
	}
	common.DrawIcon(icon, painter, rect)
}

// reversedTheme returns the theme that reverses the icon color (used by
// primary/accent buttons where the foreground is the inverse of the theme).
func reversedTheme() common.Theme {
	if common.IsDarkTheme() {
		return common.ThemeLight
	}
	return common.ThemeDark
}

// readEmbeddedImage reads a raw image asset from the embedded resources FS.
// subpath is relative to the images/ directory, e.g. "check_box/Accept_black.svg".
func readEmbeddedImage(subpath string) []byte {
	b, err := fs.ReadFile(resources.Images, "images/"+subpath)
	if err != nil {
		return nil
	}
	return b
}

// drawSvgBytes renders arbitrary SVG content into a painter (shared by the
// special check-box / spin-box / info-bar icons that live outside the standard
// icons directory).
func drawSvgBytes(svgBytes []byte, painter *qt.QPainter, rect *qt.QRectF) {
	common.DrawSvgIcon(svgBytes, painter, rect)
}
