// Package resources embeds the QSS stylesheets, SVG/PNG images and i18n
// translation files copied from PyQt-Fluent-Widgets' qfluentwidgets/_rc
// directory. Every other gofluent package reads assets through this package so
// that the resources are compiled into the final static binary.
package resources

import (
	"embed"
	"io/fs"
	"strings"
)

// LightQSS embeds the 34 light-theme stylesheets (qss/light/*.qss).
//
//go:embed qss/light/*.qss
var LightQSS embed.FS

// DarkQSS embeds the 34 dark-theme stylesheets (qss/dark/*.qss).
//
//go:embed qss/dark/*.qss
var DarkQSS embed.FS

// Images embeds every SVG/PNG asset under images/ (icons and control sprites).
//
//go:embed all:images
var Images embed.FS

// I18n embeds the .qm/.ts translation files under i18n/.
//
//go:embed all:i18n
var I18n embed.FS

// QSS returns the raw content of a stylesheet for the given theme and name.
// theme must be "light" or "dark" and name must not contain the ".qss" suffix,
// e.g. QSS("light", "button").
func QSS(theme, name string) string {
	f := "qss/" + theme + "/" + name + ".qss"
	var (
		b   []byte
		err error
	)
	if theme == "dark" {
		b, err = fs.ReadFile(DarkQSS, f)
	} else {
		b, err = fs.ReadFile(LightQSS, f)
	}
	if err != nil {
		return ""
	}
	return string(b)
}

// IconSVG returns the SVG content for a FluentIcon name and color suffix.
// name is the icon enum value (e.g. "Up") and color is "black" or "white".
func IconSVG(name, color string) []byte {
	b, err := fs.ReadFile(Images, "images/icons/"+name+"_"+color+".svg")
	if err != nil {
		return nil
	}
	return b
}

// IconExists reports whether the given icon SVG is present in the embedded FS.
func IconExists(name, color string) bool {
	_, err := fs.Stat(Images, "images/icons/"+name+"_"+color+".svg")
	return err == nil
}

// Translation returns the content of the ".qm" translation file for a locale
// name (e.g. "zh_CN"). An empty slice is returned when the file is missing.
func Translation(localeName string) []byte {
	base := "qfluentwidgets." + localeName
	b, err := fs.ReadFile(I18n, "i18n/"+base+".qm")
	if err != nil {
		return nil
	}
	return b
}

// LocaleName normalises a Qt QLocale name (which uses an underscore such as
// "zh_CN") for use with Translation.
func LocaleName(name string) string {
	return strings.ReplaceAll(name, "-", "_")
}
