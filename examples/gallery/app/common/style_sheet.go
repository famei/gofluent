package common

import (
	gcommon "github.com/famei/gofluent/common"
	"github.com/famei/gofluent/examples/gallery/app/resource"
	qt "github.com/mappu/miqt/qt"
)

// StyleSheet enumerates the gallery-specific stylesheets (port of
// app/common/style_sheet.py). It implements gcommon.StyleSheetBase so it can be
// applied through common.SetStyleSheet with theme-color rendering.
type StyleSheet string

const (
	StyleLinkCard                StyleSheet = "link_card"
	StyleSampleCard              StyleSheet = "sample_card"
	StyleHomeInterface           StyleSheet = "home_interface"
	StyleIconInterface           StyleSheet = "icon_interface"
	StyleViewInterface           StyleSheet = "view_interface"
	StyleSettingInterface        StyleSheet = "setting_interface"
	StyleGalleryInterface        StyleSheet = "gallery_interface"
	StyleNavigationViewInterface StyleSheet = "navigation_view_interface"
)

// Name returns the QSS file name without extension.
func (s StyleSheet) Name() string { return string(s) }

// Path returns the logical Qt resource path of the stylesheet.
func (s StyleSheet) Path(theme gcommon.Theme) string {
	return ":/gallery/qss/" + s.resolve(theme) + "/" + s.Name() + ".qss"
}

// Content returns the raw QSS content for the theme.
func (s StyleSheet) Content(theme gcommon.Theme) string {
	return resource.QSS(s.resolve(theme), s.Name())
}

// Apply applies the stylesheet to the widget.
func (s StyleSheet) Apply(widget *qt.QWidget, theme gcommon.Theme) {
	gcommon.SetStyleSheet(widget, s, theme)
}

func (s StyleSheet) resolve(theme gcommon.Theme) string {
	if theme == gcommon.ThemeAuto {
		return gcommon.CurrentTheme().Lower()
	}
	return theme.Lower()
}
