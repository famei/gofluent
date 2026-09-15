// Package resource embeds the gallery assets — images, the light/dark QSS
// themes and the i18n ".qm" files — and exposes them by name.
//
// It replaces the Python app/resource/resource.qrc and its pyrcc-compiled
// counterpart app/common/resource.py: instead of a Qt resource bundle that is
// registered at runtime, the files are read from the embedded filesystem
// (MIGRATION_GUIDE §11).
package resource

import (
	"embed"
	"encoding/json"
	"io/fs"

	"github.com/famei/gofluent/examples/internal/asset"
	qt "github.com/mappu/miqt/qt"
)

//go:embed images qss i18n
var files embed.FS

// Image returns a *qt.QImage decoded from an embedded gallery image located at
// images/<name>. The caller owns the returned object.
func Image(name string) *qt.QImage {
	return asset.QImage(files, "images/"+name)
}

// Pixmap returns a *qt.QPixmap decoded from an embedded gallery image located
// at images/<name>. The caller owns the returned object.
func Pixmap(name string) *qt.QPixmap {
	return asset.QPixmap(files, "images/"+name)
}

// Icon returns a *qt.QIcon decoded from an embedded gallery image located at
// images/<name>. The caller owns the returned object.
func Icon(name string) *qt.QIcon {
	return asset.QIcon(files, "images/"+name)
}

// QSS returns the raw QSS content of the theme/name stylesheet
// (qss/<theme>/<name>.qss).
func QSS(theme, name string) string {
	return asset.String(files, "qss/"+theme+"/"+name+".qss")
}

// IconSVG returns the SVG bytes of a gallery icon (images/icons/<name>_<color>.svg).
func IconSVG(name, color string) []byte {
	return asset.Bytes(files, "images/icons/"+name+"_"+color+".svg")
}

// Translation returns the ".qm" translation bytes for a locale name (e.g.
// "zh_CN"), or nil when no translation exists for that locale.
func Translation(localeName string) []byte {
	b, err := fs.ReadFile(files, "i18n/gallery."+localeName+".qm")
	if err != nil {
		return nil
	}
	return b
}

// IconsDataByte use https://github.com/microsoft/WinUI-Gallery/blob/main/WinUIGallery/Samples/Iconography/IconsData.json
//
//go:embed IconsData.json
var IconsDataByte []byte

type IconsData []struct {
	Code              string   `json:"Code"`
	Name              string   `json:"Name"`
	Tags              []string `json:"Tags"`
	IsSegoeFluentOnly bool     `json:"IsSegoeFluentOnly,omitempty"`
}

func GetIconsData() IconsData {
	var data IconsData
	err := json.Unmarshal(IconsDataByte, &data)
	if err != nil {
		panic(err)
	}
	return data
}
