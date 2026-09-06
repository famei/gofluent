package widgets

import (
	"io/fs"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/resources"
	qt "github.com/mappu/miqt/qt"
)

// renderFluentIcon draws a FluentIcon into painter at rect using the given
// theme (ThemeAuto follows the current theme). Non-FluentIcon sources are
// delegated to common.DrawIcon.
func renderFluentIcon(icon interface{}, painter *qt.QPainter, rect *qt.QRectF, theme common.Theme) {
	if fi, ok := icon.(common.FluentIcon); ok {
		fi.Render(painter, rect, theme)
		return
	}
	common.DrawIcon(icon, painter, rect)
}

// renderFluentIconWithFill renders a FluentIcon recolored to the given hex fill
// color (the Go equivalent of FluentIcon.render(..., fill="#rrggbb")).
func renderFluentIconWithFill(icon common.FluentIcon, painter *qt.QPainter, rect *qt.QRectF, fill string) {
	raw := resources.IconSVG(string(icon), common.GetIconColor(common.ThemeAuto, false))
	if len(raw) == 0 {
		return
	}
	common.DrawSvgIcon([]byte(common.RecolorSvg(string(raw), fill)), painter, rect)
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
