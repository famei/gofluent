package common

import (
	"strconv"
	"strings"
	"sync"

	"github.com/famei/gofluent/resources"
	qt "github.com/mappu/miqt/qt"
)

// DefaultFontFamilies is the font family list used when none is configured.
var DefaultFontFamilies = []string{"Segoe UI", "Microsoft YaHei", "PingFang SC"}

// CJKFontFamily is the family provided by the embedded CJK font.
const CJKFontFamily = "Microsoft YaHei"

var (
	cjkFontMu      sync.Mutex
	cjkFontChecked bool
)

// ensureCJKFont registers the embedded Microsoft YaHei when the platform does not
// have that family, so Chinese text renders on a machine with no Chinese font at
// all (a bare Linux container draws empty boxes otherwise). Registering it also
// makes the family in DefaultFontFamilies resolve there, so no other code has to
// know about the fallback.
//
// QFontDatabase needs a live QGuiApplication: an early call (before the
// application exists) leaves the check pending and the next call retries.
func ensureCJKFont() {
	// The font is embedded on Linux only (see resources/font_linux.go); everywhere else
	// there is nothing to register and the platform's own family is used.
	if len(resources.MicrosoftYaHei) == 0 {
		return
	}

	cjkFontMu.Lock()
	defer cjkFontMu.Unlock()
	if cjkFontChecked {
		return
	}
	if qt.QCoreApplication_Instance() == nil {
		return
	}

	db := qt.NewQFontDatabase()
	hasFamily := db.HasFamily(CJKFontFamily)
	db.Delete()
	if hasFamily {
		cjkFontChecked = true
		return
	}

	// A negative id means Qt rejected the data; retry on the next call rather
	// than giving up silently.
	if qt.QFontDatabase_AddApplicationFontFromData(resources.MicrosoftYaHei) >= 0 {
		cjkFontChecked = true
	}
}

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
	ensureCJKFont()
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
