package common

import (
	"regexp"
	"sync"
	"unsafe"

	"github.com/famei/gofluent/resources"
	qt "github.com/mappu/miqt/qt"
	"github.com/mappu/miqt/qt/svg"
)

// svgFillRe matches a `fill="#rrggbb"` attribute in the Fluent SVG icons.
var svgFillRe = regexp.MustCompile(`fill="#[0-9a-fA-F]+"`)

// GetIconColor returns the icon color suffix ("black" for light theme, "white"
// for dark theme). reverse swaps the two.
func GetIconColor(theme Theme, reverse bool) string {
	lc, dc := "black", "white"
	if reverse {
		lc, dc = "white", "black"
	}
	if theme == ThemeAuto {
		if IsDarkTheme() {
			return dc
		}
		return lc
	}
	if theme == ThemeDark {
		return dc
	}
	return lc
}

// DrawSvgIcon renders SVG byte content into a painter at the given rect.
func DrawSvgIcon(svgBytes []byte, painter *qt.QPainter, rect *qt.QRectF) {
	if len(svgBytes) == 0 {
		return
	}
	renderer := svg.NewQSvgRenderer3(svgBytes)
	defer renderer.Delete()
	renderer.Render2(painter, rect)
}

// svgBytesToIcon renders SVG content into a QIcon backed by a pixmap.
func svgBytesToIcon(svgBytes []byte) *qt.QIcon {
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

// RecolorSvg rewrites the `fill` attribute of every <path> element in the SVG
// (the Go equivalent of writeSvg(path, fill=...)).
func RecolorSvg(svgText, fill string) string {
	return svgFillRe.ReplaceAllString(svgText, `fill="`+fill+`"`)
}

// FluentIconBase is implemented by every icon source that can be rendered.
type FluentIconBase interface {
	// Path returns the logical Qt resource path of the icon.
	Path(theme Theme) string
	// Icon returns a QIcon for the given theme.
	Icon(theme Theme) *qt.QIcon
	// Colored returns an icon recolored for light/dark mode.
	Colored(light, dark *qt.QColor) *ColoredFluentIcon
	// Render draws the icon into a painter.
	Render(painter *qt.QPainter, rect *qt.QRectF, theme Theme)
	// QIcon returns a theme-following QIcon.
	QIcon(reverse bool) *qt.QIcon
}

// FluentIcon is the Fluent 2 icon set. The Go constant value is the icon name
// (the same string used by the SVG filenames).
type FluentIcon string

/*
const (

	Up                   FluentIcon = "Up"
	Add                  FluentIcon = "Add"
	Bus                  FluentIcon = "Bus"
	Car                  FluentIcon = "Car"
	Cut                  FluentIcon = "Cut"
	Iot                  FluentIcon = "IOT"
	Pin                  FluentIcon = "Pin"
	Tag                  FluentIcon = "Tag"
	Vpn                  FluentIcon = "VPN"
	Cafe                 FluentIcon = "Cafe"
	Chat                 FluentIcon = "Chat"
	Copy                 FluentIcon = "Copy"
	Code                 FluentIcon = "Code"
	Down                 FluentIcon = "Down"
	Edit                 FluentIcon = "Edit"
	Flag                 FluentIcon = "Flag"
	Font                 FluentIcon = "Font"
	Game                 FluentIcon = "Game"
	Help                 FluentIcon = "Help"
	Hide                 FluentIcon = "Hide"
	Home                 FluentIcon = "Home"
	Info                 FluentIcon = "Info"
	Leaf                 FluentIcon = "Leaf"
	Link                 FluentIcon = "Link"
	Mail                 FluentIcon = "Mail"
	Menu                 FluentIcon = "Menu"
	Mute                 FluentIcon = "Mute"
	More                 FluentIcon = "More"
	Move                 FluentIcon = "Move"
	Play                 FluentIcon = "Play"
	Save                 FluentIcon = "Save"
	Send                 FluentIcon = "Send"
	Sync                 FluentIcon = "Sync"
	Unit                 FluentIcon = "Unit"
	View                 FluentIcon = "View"
	Wifi                 FluentIcon = "Wifi"
	Zoom                 FluentIcon = "Zoom"
	Album                FluentIcon = "Album"
	Brush                FluentIcon = "Brush"
	Broom                FluentIcon = "Broom"
	Close                FluentIcon = "Close"
	Cloud                FluentIcon = "Cloud"
	Embed                FluentIcon = "Embed"
	Globe                FluentIcon = "Globe"
	Heart                FluentIcon = "Heart"
	Label                FluentIcon = "Label"
	Media                FluentIcon = "Media"
	Movie                FluentIcon = "Movie"
	Music                FluentIcon = "Music"
	Robot                FluentIcon = "Robot"
	Pause                FluentIcon = "Pause"
	Paste                FluentIcon = "Paste"
	Photo                FluentIcon = "Photo"
	Phone                FluentIcon = "Phone"
	Print                FluentIcon = "Print"
	Share                FluentIcon = "Share"
	Tiles                FluentIcon = "Tiles"
	Unpin                FluentIcon = "Unpin"
	Video                FluentIcon = "Video"
	Train                FluentIcon = "Train"
	AddTo                FluentIcon = "AddTo"
	Accept               FluentIcon = "Accept"
	Camera               FluentIcon = "Camera"
	Cancel               FluentIcon = "Cancel"
	Delete               FluentIcon = "Delete"
	Folder               FluentIcon = "Folder"
	Filter               FluentIcon = "Filter"
	Market               FluentIcon = "Market"
	Scroll               FluentIcon = "Scroll"
	Layout               FluentIcon = "Layout"
	GitHub               FluentIcon = "GitHub"
	Update               FluentIcon = "Update"
	Remove               FluentIcon = "Remove"
	Return               FluentIcon = "Return"
	People               FluentIcon = "People"
	QRCode               FluentIcon = "QRCode"
	Ringer               FluentIcon = "Ringer"
	Rotate               FluentIcon = "Rotate"
	Search               FluentIcon = "Search"
	Volume               FluentIcon = "Volume"
	Frigid               FluentIcon = "Frigid"
	SaveAs               FluentIcon = "SaveAs"
	ZoomIn               FluentIcon = "ZoomIn"
	Connect              FluentIcon = "Connect"
	History              FluentIcon = "History"
	Setting              FluentIcon = "Setting"
	Palette              FluentIcon = "Palette"
	Message              FluentIcon = "Message"
	FitPage              FluentIcon = "FitPage"
	ZoomOut              FluentIcon = "ZoomOut"
	Airplane             FluentIcon = "Airplane"
	Asterisk             FluentIcon = "Asterisk"
	Calories             FluentIcon = "Calories"
	Calendar             FluentIcon = "Calendar"
	Feedback             FluentIcon = "Feedback"
	Library              FluentIcon = "BookShelf"
	Minimize             FluentIcon = "Minimize"
	Checkbox             FluentIcon = "CheckBox"
	Document             FluentIcon = "Document"
	Language             FluentIcon = "Language"
	Download             FluentIcon = "Download"
	Question             FluentIcon = "Question"
	Speakers             FluentIcon = "Speakers"
	DateTime             FluentIcon = "DateTime"
	FontSize             FluentIcon = "FontSize"
	HomeFill             FluentIcon = "HomeFill"
	PageLeft             FluentIcon = "PageLeft"
	SaveCopy             FluentIcon = "SaveCopy"
	SendFill             FluentIcon = "SendFill"
	SkipBack             FluentIcon = "SkipBack"
	SpeedOff             FluentIcon = "SpeedOff"
	Alignment            FluentIcon = "Alignment"
	Bluetooth            FluentIcon = "Bluetooth"
	Completed            FluentIcon = "Completed"
	Constract            FluentIcon = "Constract"
	Headphone            FluentIcon = "Headphone"
	Megaphone            FluentIcon = "Megaphone"
	Projector            FluentIcon = "Projector"
	Education            FluentIcon = "Education"
	LeftArrow            FluentIcon = "LeftArrow"
	EraseTool            FluentIcon = "EraseTool"
	PageRight            FluentIcon = "PageRight"
	PlaySolid            FluentIcon = "PlaySolid"
	BookShelf            FluentIcon = "BookShelf"
	Highlight            FluentIcon = "Highlight"
	FolderAdd            FluentIcon = "FolderAdd"
	PauseBold            FluentIcon = "PauseBold"
	PencilInk            FluentIcon = "PencilInk"
	PieSingle            FluentIcon = "PieSingle"
	QuickNote            FluentIcon = "QuickNote"
	SpeedHigh            FluentIcon = "SpeedHigh"
	StopWatch            FluentIcon = "StopWatch"
	ZipFolder            FluentIcon = "ZipFolder"
	Basketball           FluentIcon = "Basketball"
	Brightness           FluentIcon = "Brightness"
	Dictionary           FluentIcon = "Dictionary"
	Microphone           FluentIcon = "Microphone"
	ArrowDown            FluentIcon = "ChevronDown"
	FullScreen           FluentIcon = "FullScreen"
	MixVolumes           FluentIcon = "MixVolumes"
	RemoveFrom           FluentIcon = "RemoveFrom"
	RightArrow           FluentIcon = "RightArrow"
	QuietHours           FluentIcon = "QuietHours"
	Fingerprint          FluentIcon = "Fingerprint"
	Application          FluentIcon = "Application"
	Certificate          FluentIcon = "Certificate"
	Transparent          FluentIcon = "Transparent"
	ImageExport          FluentIcon = "ImageExport"
	SpeedMedium          FluentIcon = "SpeedMedium"
	LibraryFill          FluentIcon = "LibraryFill"
	MusicFolder          FluentIcon = "MusicFolder"
	PowerButton          FluentIcon = "PowerButton"
	SkipForward          FluentIcon = "SkipForward"
	CareUpSolid          FluentIcon = "CareUpSolid"
	AcceptMedium         FluentIcon = "AcceptMedium"
	CancelMedium         FluentIcon = "CancelMedium"
	ChevronRight         FluentIcon = "ChevronRight"
	ClippingTool         FluentIcon = "ClippingTool"
	SearchMirror         FluentIcon = "SearchMirror"
	ShoppingCart         FluentIcon = "ShoppingCart"
	FontIncrease         FluentIcon = "FontIncrease"
	BackToWindow         FluentIcon = "BackToWindow"
	CommandPrompt        FluentIcon = "CommandPrompt"
	CloudDownload        FluentIcon = "CloudDownload"
	DictionaryAdd        FluentIcon = "DictionaryAdd"
	CareDownSolid        FluentIcon = "CareDownSolid"
	CareLeftSolid        FluentIcon = "CareLeftSolid"
	ClearSelection       FluentIcon = "ClearSelection"
	DeveloperTools       FluentIcon = "DeveloperTools"
	BackgroundFill       FluentIcon = "BackgroundColor"
	CareRightSolid       FluentIcon = "CareRightSolid"
	ChevronDownMed       FluentIcon = "ChevronDownMed"
	ChevronRightMed      FluentIcon = "ChevronRightMed"
	EmojiTabSymbols      FluentIcon = "EmojiTabSymbols"
	ExpressiveInputEntry FluentIcon = "ExpressiveInputEntry"

)
*/

const (
	GitHub FluentIcon = "GitHub"
)

// Path returns the logical Qt resource path for the icon.
func (f FluentIcon) Path(theme Theme) string {
	return ":/qfluentwidgets/images/icons/" + string(f) + "_" + GetIconColor(theme, false) + ".svg"
}

// svgBytes returns the embedded SVG content for the icon and theme.
func (f FluentIcon) svgBytes(theme Theme) []byte {
	return resources.IconSVG(string(f), GetIconColor(theme, false))
}

// Icon returns a QIcon rendered from the embedded SVG.
func (f FluentIcon) Icon(theme Theme) *qt.QIcon {
	return svgBytesToIcon(f.svgBytes(theme))
}

// IconWithColor returns a QIcon whose SVG fill attribute is set to color.
func (f FluentIcon) IconWithColor(theme Theme, color *qt.QColor) *qt.QIcon {
	recolorSvg := RecolorSvg(string(f.svgBytes(theme)), color.Name())
	return svgBytesToIcon([]byte(recolorSvg))
}

// Colored returns an icon recolored for light/dark mode.
func (f FluentIcon) Colored(light, dark *qt.QColor) *ColoredFluentIcon {
	return &ColoredFluentIcon{fluentIcon: f, lightColor: light, darkColor: dark}
}

// Render draws the icon into a painter.
func (f FluentIcon) Render(painter *qt.QPainter, rect *qt.QRectF, theme Theme) {
	DrawSvgIcon(f.svgBytes(theme), painter, rect)
}

// QIcon returns a QIcon for the current theme (reverse swaps black/white).
func (f FluentIcon) QIcon(reverse bool) *qt.QIcon {
	return svgBytesToIcon(resources.IconSVG(string(f), GetIconColor(ThemeAuto, reverse)))
}

// ColoredFluentIcon wraps a FluentIconBase and applies light/dark colors.
type ColoredFluentIcon struct {
	fluentIcon FluentIconBase
	lightColor *qt.QColor
	darkColor  *qt.QColor
}

// Path delegates to the wrapped icon.
func (c *ColoredFluentIcon) Path(theme Theme) string { return c.fluentIcon.Path(theme) }

// Icon returns a recolored QIcon.
func (c *ColoredFluentIcon) Icon(theme Theme) *qt.QIcon {
	//svgBytes := c.svgBytes(theme)
	//return svgBytesToIcon(svgBytes)
	if f, ok := c.fluentIcon.(SegoeFluentIcon); ok {
		return f.renderIcon(theme, c.color(theme))
	}
	return c.fluentIcon.Icon(theme)
}

func (c *ColoredFluentIcon) color(theme Theme) *qt.QColor {
	if resolvedTheme(theme) == ThemeDark {
		return c.darkColor
	}
	return c.lightColor
}

func (c *ColoredFluentIcon) svgBytes(theme Theme) []byte {
	if fi, ok := c.fluentIcon.(FluentIcon); ok {
		raw := fi.svgBytes(theme)
		return []byte(RecolorSvg(string(raw), c.color(theme).Name()))
	}
	return nil
}

// Colored returns the receiver (already colored).
func (c *ColoredFluentIcon) Colored(light, dark *qt.QColor) *ColoredFluentIcon {
	c.lightColor, c.darkColor = light, dark
	return c
}

// Render draws the recolored icon into a painter.
func (c *ColoredFluentIcon) Render(painter *qt.QPainter, rect *qt.QRectF, theme Theme) {
	//if b := c.svgBytes(theme); len(b) > 0 {
	//	DrawSvgIcon(b, painter, rect)
	//}
	if f, ok := c.fluentIcon.(SegoeFluentIcon); ok {
		f.RenderGlyph(painter, rect, theme, c.color(theme))
		return
	}
	c.fluentIcon.Render(painter, rect, theme)
}

// QIcon returns a theme-following recolored QIcon.
func (c *ColoredFluentIcon) QIcon(reverse bool) *qt.QIcon {
	return c.Icon(ThemeAuto)
}

// FluentFontIconBase renders glyphs from an application font (e.g. Segoe
// Fluent Icons). loadFont must be called after QApplication exists.
type FluentFontIconBase struct {
	char       string
	fontFamily string
	lightColor *qt.QColor
	darkColor  *qt.QColor
	isBold     bool
}

// NewFluentFontIconBase builds a font icon for the given character.
func NewFluentFontIconBase(char, fontFamily string) *FluentFontIconBase {
	return &FluentFontIconBase{
		char:       char,
		fontFamily: fontFamily,
		lightColor: qt.NewQColor3(0, 0, 0),
		darkColor:  qt.NewQColor3(255, 255, 255),
	}
}

// Bold marks the glyph bold and returns the receiver.
func (f *FluentFontIconBase) Bold() *FluentFontIconBase {
	f.isBold = true
	return f
}

// Path returns an empty string (font icons have no file path).
func (f *FluentFontIconBase) Path(theme Theme) string { return "" }

func (f *FluentFontIconBase) iconColor(theme Theme) *qt.QColor {
	if resolvedTheme(theme) == ThemeDark {
		return f.darkColor
	}
	return f.lightColor
}

// renderGlyph draws the glyph into a painter (shared by Icon and Render).
func (f *FluentFontIconBase) renderGlyph(painter *qt.QPainter, rect *qt.QRectF, theme Theme) {
	font := qt.NewQFont2(f.fontFamily)
	defer font.Delete()
	font.SetBold(f.isBold)
	font.SetPixelSize(int(rect.Height()))

	// Save/restore the painter state so the glyph font and its pen/brush never
	// leak into the caller (which may draw its own text right after the icon).
	painter.Save()
	defer painter.Restore()

	painter.SetFont(font)
	painter.SetPenWithStyle(qt.NoPen)
	brush := qt.NewQBrush3(f.iconColor(theme))
	defer brush.Delete()
	painter.SetBrush(brush)
	painter.SetRenderHints(qt.QPainter__Antialiasing | qt.QPainter__TextAntialiasing)

	path := qt.NewQPainterPath()
	defer path.Delete()
	path.AddText2(rect.X(), rect.Y()+rect.Height(), font, f.char)
	painter.DrawPath(path)
}

// Icon returns a QIcon rendered from the glyph.
func (f *FluentFontIconBase) Icon(theme Theme) *qt.QIcon {
	const size = 128
	pm := qt.NewQPixmap2(size, size)
	transparent := qt.NewQColor11(0, 0, 0, 0)
	pm.FillWithFillColor(transparent)
	transparent.Delete()

	painter := qt.NewQPainter2(pm.QPaintDevice)
	f.renderGlyph(painter, qt.NewQRectF4(0, 0, size, size), theme)
	painter.End()
	painter.Delete()

	return qt.NewQIcon2(pm)
}

// Colored sets the light/dark glyph colors.
func (f *FluentFontIconBase) Colored(light, dark *qt.QColor) *ColoredFluentIcon {
	f.lightColor = light
	f.darkColor = dark
	return &ColoredFluentIcon{fluentIcon: f, lightColor: light, darkColor: dark}
}

// Render draws the glyph into a painter.
func (f *FluentFontIconBase) Render(painter *qt.QPainter, rect *qt.QRectF, theme Theme) {
	f.renderGlyph(painter, rect, theme)
}

// QIcon returns a theme-following glyph QIcon.
func (f *FluentFontIconBase) QIcon(reverse bool) *qt.QIcon { return f.Icon(ThemeAuto) }

// Icon wraps a FluentIcon together with its current QIcon.
type Icon struct {
	FluentIcon FluentIcon
	QIcon      *qt.QIcon
}

// NewIcon builds an Icon from a FluentIcon.
func NewIcon(fluentIcon FluentIcon) *Icon {
	return &Icon{FluentIcon: fluentIcon, QIcon: fluentIcon.Icon(ThemeAuto)}
}

// ToQIcon converts an icon source (string path, FluentIconBase or *qt.QIcon)
// to a *qt.QIcon.
func ToQIcon(icon interface{}) *qt.QIcon {
	switch v := icon.(type) {
	case string:
		return qt.NewQIcon4(v)
	case FluentIconBase:
		return v.Icon(ThemeAuto)
	case *qt.QIcon:
		return v
	default:
		return qt.NewQIcon()
	}
}

// DrawIcon draws an icon source into a painter.
func DrawIcon(icon interface{}, painter *qt.QPainter, rect *qt.QRectF) {
	switch v := icon.(type) {
	case FluentIconBase:
		v.Render(painter, rect, ThemeAuto)
	case *Icon:
		v.FluentIcon.Render(painter, rect, ThemeAuto)
	case *qt.QIcon:
		v.Paint(painter, rect.ToRect())
	default:
		ToQIcon(icon).Paint(painter, rect.ToRect())
	}
}

// Action is a QAction that remembers the FluentIcon it was created with.
type Action struct {
	*qt.QAction
	fluentIcon FluentIconBase
}

// NewAction builds an action with an optional parent.
func NewAction(parent *qt.QObject) *Action {
	return &Action{QAction: qt.NewQAction4(parent)}
}

// NewActionText builds an action with text.
func NewActionText(text string, parent *qt.QObject) *Action {
	return &Action{QAction: qt.NewQAction5(text, parent)}
}

// NewActionIcon builds an action from a QIcon and text.
func NewActionIcon(icon *qt.QIcon, text string, parent *qt.QObject) *Action {
	return &Action{QAction: qt.NewQAction6(icon, text, parent)}
}

// NewActionFluentIcon builds an action from a FluentIconBase and text.
//
// The action's icon is always the icon of the *current theme*: an action's icon
// is shown by many widgets (menu rows, command-bar buttons, tool buttons) and
// only some of them draw it on an accent background, so the inversion of a
// checked (accent) state belongs to those widgets, not to the shared action. A
// checked RoundMenu row, for example, keeps the ordinary menu surface and would
// otherwise show a white icon on a light menu.
func NewActionFluentIcon(icon FluentIconBase, text string, parent *qt.QObject) *Action {
	a := &Action{QAction: qt.NewQAction6(icon.Icon(ThemeAuto), text, parent), fluentIcon: icon}
	registerActionIcon(a.QAction, icon)

	// Re-render the icon when the theme changes. The action is often handed to a
	// RoundMenu / CommandBar as its embedded *qt.QAction, so the fluentIcon is
	// no longer reachable there; without this the menu/command-bar icons keep the
	// color they were rendered with at construction (black icons in dark mode).
	destroyed := false
	a.QAction.OnDestroyed(func() {
		destroyed = true
		unregisterActionIcon(a.QAction)
	})
	QConfigInstance.OnThemeChanged(func(Theme) {
		if destroyed {
			return
		}
		a.QAction.SetIcon(icon.Icon(ThemeAuto))
	})

	return a
}

// actionIconSources maps the C++ QAction of a fluent action to the icon source
// it was built from.
//
// A *qt.QAction only carries a rendered *qt.QIcon, but a widget that draws a
// checkable action itself sometimes needs the source instead: the command bar's
// toggle button re-renders the icon in the reversed color while it is checked
// (its checked state paints on the accent background), which a pre-rendered
// pixmap cannot express. The key is the C++ pointer, because miqt hands out a
// fresh Go wrapper for the same QAction from every accessor.
var (
	actionIconSourcesMu sync.RWMutex
	actionIconSources   = map[unsafe.Pointer]FluentIconBase{}
)

// FluentIconOf returns the fluent icon source that action was built from with
// NewActionFluentIcon, or nil for a plain QAction (or an unknown action).
func FluentIconOf(action *qt.QAction) FluentIconBase {
	if action == nil {
		return nil
	}
	actionIconSourcesMu.RLock()
	defer actionIconSourcesMu.RUnlock()
	return actionIconSources[action.UnsafePointer()]
}

func registerActionIcon(action *qt.QAction, icon FluentIconBase) {
	if action == nil || icon == nil {
		return
	}
	actionIconSourcesMu.Lock()
	defer actionIconSourcesMu.Unlock()
	actionIconSources[action.UnsafePointer()] = icon
}

func unregisterActionIcon(action *qt.QAction) {
	if action == nil {
		return
	}
	actionIconSourcesMu.Lock()
	defer actionIconSourcesMu.Unlock()
	delete(actionIconSources, action.UnsafePointer())
}

// FluentIcon returns the wrapped fluent icon (nil when the action was built
// from a plain QIcon).
func (a *Action) FluentIcon() FluentIconBase { return a.fluentIcon }

// Icon returns the current icon, rebuilding it from the fluent icon when set.
func (a *Action) Icon() *qt.QIcon {
	if a.fluentIcon != nil {
		return a.fluentIcon.Icon(ThemeAuto)
	}
	return a.QAction.Icon()
}

// SetIcon accepts a FluentIconBase or a *qt.QIcon.
func (a *Action) SetIcon(icon interface{}) {
	switch v := icon.(type) {
	case FluentIconBase:
		a.fluentIcon = v
		a.QAction.SetIcon(v.Icon(ThemeAuto))
	case *qt.QIcon:
		a.fluentIcon = nil
		a.QAction.SetIcon(v)
	}
}
