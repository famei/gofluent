package common

import (
	"strconv"
	"strings"

	qt "github.com/mappu/miqt/qt"
)

// DefaultFontFamilies is the font family list used when none is configured.
var DefaultFontFamilies = []string{"Segoe UI", "Microsoft YaHei", "PingFang SC"}

// SetFontFamilies sets the font families used by all widgets; when save is true
// the change is persisted to the config file.
func SetFontFamilies(families []string, save bool) {
	QConfigInstance.SetFontFamilies(families, save)
}

// FontFamilies returns a copy of the configured font family list.
func FontFamilies() []string {
	return QConfigInstance.FontFamilies()
}

// SetFont applies a generated font to the widget.
func SetFont(widget *qt.QWidget, fontSize int, weight int) {
	font := GetFont(fontSize, weight)
	widget.SetFont(font)
}

// GetFont creates a font with the configured families, pixel size and weight.
// The caller owns the returned *qt.QFont (defer Delete when temporary).
func GetFont(fontSize int, weight int) *qt.QFont {
	font := qt.NewQFont()
	font.SetFamilies(QConfigInstance.FontFamilies())
	font.SetPixelSize(fontSize)
	font.SetWeight(weight)
	return font
}

// FontStyleSheet returns the CSS `font:` declaration for a font, mirroring the
// Python fontStyleSheet helper.
func FontStyleSheet(font *qt.QFont) string {
	families := font.Families()
	quoted := make([]string, 0, len(families))
	for _, f := range families {
		quoted = append(quoted, "'"+f+"'")
	}
	return "font: " + strconv.Itoa(font.PixelSize()) + "px " + strings.Join(quoted, ",")
}
