package window

import (
	"github.com/famei/gofluent/common"

	qt "github.com/mappu/miqt/qt"
)

// Title icons from a FluentIconBase source.
//
// The title bars can show any icon source (a Segoe font glyph, an SVG icon, the
// gallery's own icon type) instead of a pre-rendered QIcon, and the icon follows the
// theme: the source is re-rendered whenever the theme changes, so a black glyph does
// not stay black in the dark theme.
//
// A title bar may also be driven by the window icon (SetWindowIcon), which the
// existing SetIcon(*qt.QIcon) handles; both can be used, the last call wins.

// SetTitleIcon sets the icon label from a fluent icon source and keeps it in step with
// the theme.
func (t *FluentTitleBar) SetTitleIcon(icon common.FluentIconBase) {
	t.iconSource = icon
	ensureTitleIconHook(t.QWidget, t.iconLabel, &t.iconSource, &t.iconHooked, &t.iconGone)
	renderTitleIcon(t.iconLabel, icon)
}

// TitleIcon returns the icon source set with SetTitleIcon (nil when none).
func (t *FluentTitleBar) TitleIcon() common.FluentIconBase { return t.iconSource }

// SetTitleIcon sets the icon label from a fluent icon source and keeps it in step with
// the theme.
func (t *SplitTitleBar) SetTitleIcon(icon common.FluentIconBase) {
	t.iconSource = icon
	ensureTitleIconHook(t.QWidget, t.iconLabel, &t.iconSource, &t.iconHooked, &t.iconGone)
	renderTitleIcon(t.iconLabel, icon)
}

// TitleIcon returns the icon source set with SetTitleIcon (nil when none).
func (t *SplitTitleBar) TitleIcon() common.FluentIconBase { return t.iconSource }

// ensureTitleIconHook registers the theme-changed re-render once per title bar. It
// reads the source through the pointer so later SetTitleIcon calls are picked up, and
// stops touching the label once the bar is destroyed.
func ensureTitleIconHook(owner *qt.QWidget, label *qt.QLabel, source *common.FluentIconBase, hooked, gone *bool) {
	if *hooked || owner == nil || label == nil {
		return
	}
	*hooked = true
	owner.OnDestroyed(func() { *gone = true })
	common.QConfigInstance.OnThemeChanged(func(common.Theme) {
		if *gone {
			return
		}
		renderTitleIcon(label, *source)
	})
}

// renderTitleIcon paints the source into the 18x18 icon label.
func renderTitleIcon(label *qt.QLabel, icon common.FluentIconBase) {
	if label == nil || icon == nil {
		return
	}
	qicon := icon.Icon(common.ThemeAuto)
	if qicon == nil {
		return
	}
	defer qicon.Delete()
	// SetPixmap copies the pixmap, and the pixmap from the QIcon is borrowed.
	label.SetPixmap(qicon.Pixmap2(18, 18))
}
