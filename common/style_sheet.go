package common

import (
	"math"
	"regexp"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/famei/gofluent/resources"
	qt "github.com/mappu/miqt/qt"
)

// ThemeColor enumerates the theme colour placeholders referenced by QSS files
// (e.g. "--ThemeColorPrimary").
type ThemeColor int

const (
	ThemeColorPrimary ThemeColor = iota
	ThemeColorDark1
	ThemeColorDark2
	ThemeColorDark3
	ThemeColorLight1
	ThemeColorLight2
	ThemeColorLight3
)

var themeColorKeys = [...]string{
	"ThemeColorPrimary",
	"ThemeColorDark1",
	"ThemeColorDark2",
	"ThemeColorDark3",
	"ThemeColorLight1",
	"ThemeColorLight2",
	"ThemeColorLight3",
}

// Key returns the placeholder name (without the "--" prefix).
func (t ThemeColor) Key() string { return themeColorKeys[t] }

// Name returns the "#RRGGBB" name of the resolved theme color.
func (t ThemeColor) Name() string {
	c := t.Color()
	defer c.Delete()
	return c.Name()
}

// Color computes the theme colour for this slot by transforming the configured
// theme colour in HSV space (faithful port of ThemeColor.color()).
func (t ThemeColor) Color() *qt.QColor {
	color := QConfigInstance.ThemeColor()
	var h, s, v float64
	color.GetHsvF(&h, &s, &v)

	if IsDarkTheme() {
		s *= 0.84
		v = 1
		switch t {
		case ThemeColorDark1:
			v *= 0.9
		case ThemeColorDark2:
			s *= 0.977
			v *= 0.82
		case ThemeColorDark3:
			s *= 0.95
			v *= 0.7
		case ThemeColorLight1:
			s *= 0.92
		case ThemeColorLight2:
			s *= 0.78
		case ThemeColorLight3:
			s *= 0.65
		}
	} else {
		switch t {
		case ThemeColorDark1:
			v *= 0.75
		case ThemeColorDark2:
			s *= 1.05
			v *= 0.5
		case ThemeColorDark3:
			s *= 1.1
			v *= 0.4
		case ThemeColorLight1:
			v *= 1.05
		case ThemeColorLight2:
			s *= 0.75
			v *= 1.05
		case ThemeColorLight3:
			s *= 0.65
			v *= 1.05
		}
	}

	c := qt.NewQColor()
	c.SetHsvF(h, math.Min(s, 1), math.Min(v, 1))
	return c
}

// themeColorMap returns the placeholder -> hex mapping used by renderQss.
func themeColorMap() map[string]string {
	m := make(map[string]string, len(themeColorKeys))
	for _, t := range allThemeColors() {
		c := t.Color() // owned (NewQColor + SetHsvF) — safe to Delete
		m[t.Key()] = c.Name()
		c.Delete()
	}
	return m
}

func allThemeColors() []ThemeColor {
	return []ThemeColor{
		ThemeColorPrimary, ThemeColorDark1, ThemeColorDark2, ThemeColorDark3,
		ThemeColorLight1, ThemeColorLight2, ThemeColorLight3,
	}
}

// ApplyThemeColor substitutes every "--ThemeColor*" placeholder in qss with its
// resolved "#RRGGBB" value.
func ApplyThemeColor(qss string) string {
	for k, v := range themeColorMap() {
		qss = strings.ReplaceAll(qss, "--"+k, v)
	}
	return qss
}

// RenderQss substitutes both theme colour placeholders and "--FontFamilies",
// and rewrites PyQt-Fluent-Widgets class-name selectors to the native Qt class
// names used by the Go port (see translateSelectors).
func RenderQss(qss string) string {
	qss = translateSelectors(qss)
	qss = ApplyThemeColor(qss)

	families := QConfigInstance.FontFamilies()
	quoted := make([]string, 0, len(families))
	for _, f := range families {
		quoted = append(quoted, "'"+f+"'")
	}
	return strings.ReplaceAll(qss, "--FontFamilies", strings.Join(quoted, ","))
}

// selectorTranslation maps a PyQt-Fluent-Widgets custom class name to the
// selector(s) that identify the corresponding Go widget.
//
// In Python every fluent widget is a QWidget subclass with its own meta-object
// class name (PushButton, PrimaryPushButton, ToggleButton, ...), so the bundled
// QSS uses those names as type selectors and relies on Qt's subclass matching.
// The Go port implements every variant as a native Qt widget (QPushButton /
// QToolButton / QCheckBox / ...), so those type selectors never match and the
// fluent look silently disappears. This table rewrites each custom class name
// to the native Qt class name, using the objectName set in the widget
// constructor to disambiguate variants (#primaryPushButton etc.).
//
// Keys are matched as whole selector tokens (word boundaries) in a single
// pass, so a replacement value containing another key's text can never be
// re-matched and corrupted.
var selectorTranslation = map[string]string{
	// Push-button variants.
	"PushButton":                    "QPushButton",
	"PrimaryPushButton":             "QPushButton#primaryPushButton, QPushButton#primaryDropDownPushButton, QPushButton#primarySplitPushButton",
	"ToggleButton":                  "QPushButton#toggleButton, QPushButton#transparentTogglePushButton, QPushButton#pillPushButton",
	"TransparentPushButton":         "QPushButton#transparentPushButton",
	"TransparentTogglePushButton":   "QPushButton#transparentTogglePushButton",
	"DropDownPushButton":            "QPushButton#dropDownPushButton, QPushButton#transparentDropDownPushButton",
	"PrimaryDropDownPushButton":     "QPushButton#primaryDropDownPushButton",
	"TransparentDropDownPushButton": "QPushButton#transparentDropDownPushButton",
	"PillPushButton":                "QPushButton#pillPushButton",
	"HyperlinkButton":               "QPushButton#hyperlinkButton",
	// Tool-button variants.
	"ToolButton":                    "QToolButton",
	"PrimaryToolButton":             "QToolButton#primaryToolButton, QToolButton#primaryDropDownToolButton, QToolButton#primarySplitToolButton, QToolButton#primarySplitDropButton",
	"ToggleToolButton":              "QToolButton#toggleToolButton, QToolButton#transparentToggleToolButton, QToolButton#pillToolButton",
	"TransparentToolButton":         "QToolButton#transparentToolButton",
	"TransparentToggleToolButton":   "QToolButton#transparentToggleToolButton",
	"DropDownToolButton":            "QToolButton#dropDownToolButton, QToolButton#transparentDropDownToolButton",
	"PrimaryDropDownToolButton":     "QToolButton#primaryDropDownToolButton",
	"TransparentDropDownToolButton": "QToolButton#transparentDropDownToolButton",
	"PillToolButton":                "QToolButton#pillToolButton",
	"SplitDropButton":               "QToolButton#splitDropButton",
	"PrimarySplitDropButton":        "QToolButton#primarySplitDropButton",
	// Other basic inputs. The Go port implements RadioButton as a plain
	// QRadioButton, so the class selector is rewritten to the object-name
	// selector set in NewRadioButton (mirroring PyQt's "RadioButton" meta-class
	// matching).
	"RadioButton":          "QRadioButton#radioButton",
	"CheckBox":             "QCheckBox",
	"ComboBox":             "QPushButton#comboBox",
	"ModelComboBox":        "QPushButton#modelComboBox",
	"LineEdit":             "QLineEdit",
	"TextEdit":             "QTextEdit",
	"PlainTextEdit":        "QPlainTextEdit",
	"TextBrowser":          "QTextBrowser",
	"SpinBox":              "QSpinBox",
	"DoubleSpinBox":        "QDoubleSpinBox",
	"DateEdit":             "QDateEdit",
	"DateTimeEdit":         "QDateTimeEdit",
	"TimeEdit":             "QTimeEdit",
	"CompactSpinBox":       "QSpinBox#compactSpinBox",
	"CompactDoubleSpinBox": "QDoubleSpinBox#compactDoubleSpinBox",
	"CompactDateEdit":      "QDateEdit#compactDateEdit",
	"CompactDateTimeEdit":  "QDateTimeEdit#compactDateTimeEdit",
	"CompactTimeEdit":      "QTimeEdit#compactTimeEdit",
	"SpinButton":           "QToolButton#spinButton",
	// Date/time pickers (the Go port implements CalendarPicker and PickerBase as
	// plain QPushButtons, so the class selectors are rewritten to the object-name
	// selectors set in NewCalendarPicker / NewPickerBase).
	"CalendarPicker":  "QPushButton#calendarPicker",
	"PickerBase":      "QPushButton#pickerBase",
	"FluentLabelBase": "QLabel",
	"HyperlinkLabel":  "QPushButton#hyperlinkLabel",
	"SwitchButton":    "QWidget#switchButton",
	// Menus.
	"RoundMenu":            "QMenu",
	"MenuActionListWidget": "QListWidget",
	// Views (the Go port uses the native QAbstractItemView types directly, so
	// the fluent class names map to the plain Qt class selectors).
	"ListView":   "QListView",
	"ListWidget": "QListWidget",
	// PipsPager is implemented as a plain QListWidget, so its QSS type selector
	// must be rewritten too (otherwise "PipsPager { border: none;
	// background-color: transparent; }" never matches and the pager keeps its
	// default frame/background).
	"PipsPager": "QListWidget",
	// Overlay close buttons (the Go port implements them as TransparentToolButton).
	"InfoBarCloseButton": "QToolButton#transparentToolButton",
	// Gallery app widgets (examples/gallery): the Go port implements them as
	// plain QWidget/QFrame with a unique objectName, so the gallery QSS class
	// selectors are rewritten to object-name selectors.
	"SettingCardGroup":  "QWidget#settingCardGroup",
	"RangeSettingCard":  "QWidget#rangeSettingCard",
	"ColorPickerButton": "QToolButton#colorPickerButton",
	"LinkCard":          "QFrame#linkCard",
	"BannerWidget":      "QWidget#bannerWidget",
	"ExampleCard":       "QWidget#exampleCard",
	"GalleryInterface":  "QScrollArea",
	"ToolBar":           "QWidget#toolBar",
	"IconCard":          "QFrame#iconCard",
	"IconCardView":      "QWidget#iconCardView",
	"IconInfoPanel":     "QFrame#iconInfoPanel",
	"ExpandSettingCard": "QScrollArea",
	"StackedWidget":     "QFrame#stackedWidget",
	// Window title bars. The Go port implements FluentTitleBar/MSFluentTitleBar/
	// SplitTitleBar as plain QWidgets, so the fluent_window.qss type selectors are
	// rewritten to object-name selectors. MSFluentTitleBar is a FluentTitleBar in
	// Python, so it keeps the shared #fluentTitleBar object name and is told apart
	// by the dynamic isMsTitleBar property (attribute selector).
	"FluentTitleBar":   "QWidget#fluentTitleBar",
	"MSFluentTitleBar": "QWidget#fluentTitleBar[isMsTitleBar=true]",
	"SplitTitleBar":    "QWidget#splitTitleBar",
}

// selectorSuffixPattern matches the pseudo-classes (":hover", ":checked",
// ":checked:hover") and attribute selectors ("[hasIcon=true]") that may follow
// a translated class-name token. Sub-controls ("::indicator") intentionally do
// not match, so "RadioButton::indicator" keeps its sub-control intact.
const selectorSuffixPattern = `(?::[a-zA-Z_][a-zA-Z0-9_-]*|\[[^\]]*\])*`

// selectorTranslateRe matches any selectorTranslation key as a whole
// word-bounded token, plus any trailing pseudo-class/attribute suffix, with the
// longest keys tried first so that e.g. "PrimaryDropDownPushButton" wins over
// "PushButton" at the same position.
var selectorTranslateRe = func() *regexp.Regexp {
	keys := make([]string, 0, len(selectorTranslation))
	for k := range selectorTranslation {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return len(keys[i]) > len(keys[j]) })

	quoted := make([]string, len(keys))
	for i, k := range keys {
		quoted[i] = regexp.QuoteMeta(k)
	}
	return regexp.MustCompile(`\b(` + strings.Join(quoted, "|") + `)\b(` + selectorSuffixPattern + `)`)
}()

// translateSelectors rewrites the PyQt-Fluent-Widgets class-name selectors in a
// QSS string to the native Qt selectors used by the Go port. It is a single
// regex pass (not sequential ReplaceAll), so replacement text is never scanned
// again and cannot be corrupted by later keys.
//
// A translation value may be a comma-separated selector list (to reproduce the
// reference's subclass matching, e.g. ToggleButton matching both #toggleButton
// and #transparentTogglePushButton). The trailing pseudo-class / attribute
// suffix is re-attached to every selector in the list so that
// "ToggleButton:checked" becomes
// "QPushButton#toggleButton:checked, QPushButton#transparentTogglePushButton:checked"
// instead of "QPushButton#toggleButton, QPushButton#transparentTogglePushButton:checked".
func translateSelectors(qss string) string {
	matches := selectorTranslateRe.FindAllStringSubmatchIndex(qss, -1)
	if len(matches) == 0 {
		return qss
	}

	var b strings.Builder
	last := 0
	for _, m := range matches {
		b.WriteString(qss[last:m[0]])
		key := qss[m[2]:m[3]]
		suffix := qss[m[4]:m[5]]
		to, ok := selectorTranslation[key]
		if !ok {
			b.WriteString(qss[m[0]:m[1]])
		} else if !strings.Contains(to, ",") {
			b.WriteString(to)
			b.WriteString(suffix)
		} else {
			for i, p := range strings.Split(to, ",") {
				if i > 0 {
					b.WriteString(", ")
				}
				b.WriteString(strings.TrimSpace(p))
				b.WriteString(suffix)
			}
		}
		last = m[1]
	}
	b.WriteString(qss[last:])
	return b.String()
}

// StyleSheetBase is implemented by every QSS source.
type StyleSheetBase interface {
	Path(theme Theme) string
	Content(theme Theme) string
	Apply(widget *qt.QWidget, theme Theme)
}

// resolvedTheme maps ThemeAuto to the current theme.
func resolvedTheme(theme Theme) Theme {
	if theme == ThemeAuto {
		return QConfigInstance.Theme()
	}
	return theme
}

// FluentStyleSheet enumerates the 34 built-in stylesheets (matching the QSS
// filenames under resources/qss/{light,dark}).
type FluentStyleSheet int

const (
	FluentMenu FluentStyleSheet = iota
	FluentLabel
	FluentPivot
	FluentButton
	FluentDialog
	FluentSlider
	FluentInfoBar
	FluentSpinBox
	FluentTabView
	FluentToolTip
	FluentCheckBox
	FluentComboBox
	FluentFlipView
	FluentLineEdit
	FluentListView
	FluentTreeView
	FluentInfoBadge
	FluentPipsPager
	FluentTableView
	FluentCardWidget
	FluentTimePicker
	FluentColorDialog
	FluentMediaPlayer
	FluentSettingCard
	FluentTeachingTip
	FluentWindow
	FluentSwitchButton
	FluentMessageDialog
	FluentStateToolTip
	FluentCalendarPicker
	FluentFolderListDialog
	FluentSettingCardGroup
	FluentExpandSettingCard
	FluentNavigationInterface
)

var fluentStyleSheetNames = [...]string{
	"menu", "label", "pivot", "button", "dialog", "slider", "info_bar",
	"spin_box", "tab_view", "tool_tip", "check_box", "combo_box", "flip_view",
	"line_edit", "list_view", "tree_view", "info_badge", "pips_pager",
	"table_view", "card_widget", "time_picker", "color_dialog", "media_player",
	"setting_card", "teaching_tip", "fluent_window", "switch_button",
	"message_dialog", "state_tool_tip", "calendar_picker", "folder_list_dialog",
	"setting_card_group", "expand_setting_card", "navigation_interface",
}

// Name returns the stylesheet name (QSS filename without extension).
func (s FluentStyleSheet) Name() string { return fluentStyleSheetNames[s] }

// Path returns the logical Qt resource path of the stylesheet.
func (s FluentStyleSheet) Path(theme Theme) string {
	t := resolvedTheme(theme)
	return ":/qfluentwidgets/qss/" + t.Lower() + "/" + s.Name() + ".qss"
}

// Content returns the raw QSS content for the given theme.
func (s FluentStyleSheet) Content(theme Theme) string {
	t := resolvedTheme(theme)
	return resources.QSS(t.Lower(), s.Name())
}

// Apply applies the stylesheet to the widget.
func (s FluentStyleSheet) Apply(widget *qt.QWidget, theme Theme) {
	SetStyleSheet(widget, s, theme)
}

// StyleSheetFile is a stylesheet loaded from an explicit file path.
type StyleSheetFile struct {
	filePath string
}

// NewStyleSheetFile builds a StyleSheetFile from a path.
func NewStyleSheetFile(path string) *StyleSheetFile { return &StyleSheetFile{filePath: path} }

// Path returns the file path.
func (f *StyleSheetFile) Path(theme Theme) string { return f.filePath }

// Content reads the file content from disk.
func (f *StyleSheetFile) Content(theme Theme) string {
	return readFileContent(f.filePath)
}

// Apply applies the stylesheet to the widget.
func (f *StyleSheetFile) Apply(widget *qt.QWidget, theme Theme) {
	SetStyleSheet(widget, f, theme)
}

// CustomStyleSheet stores light/dark QSS on a widget via dynamic properties.
type CustomStyleSheet struct {
	widget *qt.QWidget
}

// Property keys used for the light/dark custom QSS.
const (
	DarkQSSKey  = "darkCustomQss"
	LightQSSKey = "lightCustomQss"
)

// NewCustomStyleSheet wraps a widget.
func NewCustomStyleSheet(widget *qt.QWidget) *CustomStyleSheet {
	return &CustomStyleSheet{widget: widget}
}

// Path returns an empty string (custom stylesheets have no path).
func (c *CustomStyleSheet) Path(theme Theme) string { return "" }

// Content returns the light or dark custom QSS.
func (c *CustomStyleSheet) Content(theme Theme) string {
	if resolvedTheme(theme) == ThemeLight {
		return c.LightStyleSheet()
	}
	return c.DarkStyleSheet()
}

// Apply applies the current custom QSS to the widget.
func (c *CustomStyleSheet) Apply(widget *qt.QWidget, theme Theme) {
	widget.SetStyleSheet(RenderQss(c.Content(theme)))
}

// SetCustomStyleSheet stores both light and dark QSS.
func (c *CustomStyleSheet) SetCustomStyleSheet(lightQss, darkQss string) *CustomStyleSheet {
	c.SetLightStyleSheet(lightQss)
	c.SetDarkStyleSheet(darkQss)
	return c
}

// SetLightStyleSheet stores the light QSS on the widget.
func (c *CustomStyleSheet) SetLightStyleSheet(qss string) *CustomStyleSheet {
	if c.widget != nil {
		c.widget.SetProperty(LightQSSKey, qt.NewQVariant14(qss))
	}
	return c
}

// SetDarkStyleSheet stores the dark QSS on the widget.
func (c *CustomStyleSheet) SetDarkStyleSheet(qss string) *CustomStyleSheet {
	if c.widget != nil {
		c.widget.SetProperty(DarkQSSKey, qt.NewQVariant14(qss))
	}
	return c
}

// LightStyleSheet returns the stored light QSS.
func (c *CustomStyleSheet) LightStyleSheet() string {
	if c.widget == nil {
		return ""
	}
	v := c.widget.Property(LightQSSKey)
	if v == nil || v.IsNull() {
		return ""
	}
	return v.ToString()
}

// DarkStyleSheet returns the stored dark QSS.
func (c *CustomStyleSheet) DarkStyleSheet() string {
	if c.widget == nil {
		return ""
	}
	v := c.widget.Property(DarkQSSKey)
	if v == nil || v.IsNull() {
		return ""
	}
	return v.ToString()
}

// StyleSheetCompose concatenates several stylesheet sources.
type StyleSheetCompose struct {
	sources []StyleSheetBase
}

// NewStyleSheetCompose builds a compose from sources.
func NewStyleSheetCompose(sources ...StyleSheetBase) *StyleSheetCompose {
	return &StyleSheetCompose{sources: sources}
}

// Path is empty for a compose.
func (s *StyleSheetCompose) Path(theme Theme) string { return "" }

// Content joins the content of every source with newlines.
func (s *StyleSheetCompose) Content(theme Theme) string {
	parts := make([]string, 0, len(s.sources))
	for _, src := range s.sources {
		if src != nil {
			parts = append(parts, src.Content(theme))
		}
	}
	return strings.Join(parts, "\n")
}

// Apply applies the composed stylesheet to the widget.
func (s *StyleSheetCompose) Apply(widget *qt.QWidget, theme Theme) {
	widget.SetStyleSheet(RenderQss(s.Content(theme)))
}

// Add appends a source if it is not already present. Two CustomStyleSheet
// sources wrapping the same widget are treated as duplicates (mirrors Python
// CustomStyleSheet.__eq__, which compares by widget rather than identity).
func (s *StyleSheetCompose) Add(source StyleSheetBase) {
	if source == nil {
		return
	}
	for _, existing := range s.sources {
		if existing == source {
			return
		}
		if ec, ok := existing.(*CustomStyleSheet); ok {
			if sc, ok2 := source.(*CustomStyleSheet); ok2 && ec.widget == sc.widget {
				return
			}
		}
	}
	s.sources = append(s.sources, source)
}

// ---------------------------------------------------------------------------
// Style sheet manager

type styleSheetManagerT struct {
	mu      sync.Mutex
	sources map[*qt.QWidget]StyleSheetBase
}

var styleSheetManager = &styleSheetManagerT{sources: map[*qt.QWidget]StyleSheetBase{}}

func (m *styleSheetManagerT) register(source StyleSheetBase, widget *qt.QWidget, reset bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.sources[widget]; !ok {
		widget.OnDestroyed(func() { m.deregister(widget) })
		// Re-apply lazily-skipped stylesheets the next time the widget paints
		// (mirrors Python's DirtyStyleSheetWatcher, see UpdateStyleSheet).
		widget.InstallEventFilter(dirtyStyleSheetWatcher)
		m.sources[widget] = NewStyleSheetCompose(source, NewCustomStyleSheet(widget))
		return
	}
	if !reset {
		m.sources[widget].(*StyleSheetCompose).Add(source)
		return
	}
	m.sources[widget] = NewStyleSheetCompose(source, NewCustomStyleSheet(widget))
}

func (m *styleSheetManagerT) deregister(widget *qt.QWidget) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sources, widget)
}

// widgetForQObject maps a watched QObject back to its registered widget. It is
// only called from the dirty watcher when a widget is actually marked dirty, so
// the linear scan is negligible.
func (m *styleSheetManagerT) widgetForQObject(obj *qt.QObject) *qt.QWidget {
	m.mu.Lock()
	defer m.mu.Unlock()
	for w := range m.sources {
		if w.QObject.UnsafePointer() == obj.UnsafePointer() {
			return w
		}
	}
	return nil
}

// dirtyStyleSheetWatcher re-applies the stylesheet of a widget that was skipped
// by a lazy UpdateStyleSheet (marked "dirty-qss") on its next paint, mirroring
// PyQt-Fluent-Widgets' DirtyStyleSheetWatcher.
var dirtyStyleSheetWatcher = qt.NewQObject()

func init() {
	dirtyStyleSheetWatcher.OnEventFilter(func(super func(watched *qt.QObject, event *qt.QEvent) bool, watched *qt.QObject, event *qt.QEvent) bool {
		if event.Type() != qt.QEvent__Paint {
			return super(watched, event)
		}
		v := watched.Property("dirty-qss")
		if v == nil || v.IsNull() || !v.ToBool() {
			return super(watched, event)
		}
		falseV := qt.NewQVariant11(false)
		watched.SetProperty("dirty-qss", falseV)
		falseV.Delete()
		if w := styleSheetManager.widgetForQObject(watched); w != nil {
			if src := styleSheetManager.source(w); src != nil {
				w.SetStyleSheet(GetStyleSheet(src, QConfigInstance.Theme()))
			}
		}
		return super(watched, event)
	})
}

func (m *styleSheetManagerT) source(widget *qt.QWidget) StyleSheetBase {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.sources[widget]; ok {
		return s
	}
	return NewStyleSheetCompose()
}

// ---------------------------------------------------------------------------
// Public API

// GetStyleSheet renders a stylesheet source to its final QSS string.
func GetStyleSheet(source interface{}, theme Theme) string {
	s := toStyleSheet(source)
	return renderQssCached(s.Content(theme))
}

// renderCache caches the rendered form of a raw QSS body, keyed by the raw
// content. RenderQss (selector translation + theme-color/font substitution) is
// not free: without a cache a gallery theme switch re-runs the ~80-alternation
// selector regex and 7 theme-color substitutions once per widget for every
// registered widget, which takes seconds. Invalidated when the theme color or
// font families change (RenderQss substitutes both at render time).
var renderCache sync.Map

// renderQssCached returns the cached render of qss, rendering it on a miss.
func renderQssCached(qss string) string {
	if v, ok := renderCache.Load(qss); ok {
		return v.(string)
	}
	rendered := RenderQss(qss)
	renderCache.Store(qss, rendered)
	return rendered
}

// DebugWidgetCounts returns (total, visible) counts of registered stylesheet
// widgets. Temporary diagnostic for the theme-switch performance work.
func DebugWidgetCounts() (int, int) {
	total, visible := 0, 0
	styleSheetManager.mu.Lock()
	for w := range styleSheetManager.sources {
		total++
		if w.IsVisible() {
			visible++
		}
	}
	styleSheetManager.mu.Unlock()
	return total, visible
}

// invalidateRenderCache drops every cached render (theme color / font change).
func invalidateRenderCache() {
	renderCache.Range(func(k, _ interface{}) bool {
		renderCache.Delete(k)
		return true
	})
}

// SetStyleSheet registers the widget (when register is true) and applies the
// rendered stylesheet.
func SetStyleSheet(widget *qt.QWidget, source interface{}, theme Theme) {
	SetStyleSheetWithRegister(widget, source, theme, true)
}

// SetStyleSheetWithRegister applies a stylesheet and optionally registers the
// widget for automatic theme updates.
func SetStyleSheetWithRegister(widget *qt.QWidget, source interface{}, theme Theme, register bool) {
	s := toStyleSheet(source)
	if register {
		styleSheetManager.register(s, widget, true)
	}
	widget.SetStyleSheet(GetStyleSheet(s, theme))
}

// AddStyleSheet adds a stylesheet to a widget (combining with its current QSS).
func AddStyleSheet(widget *qt.QWidget, source interface{}, theme Theme) {
	AddStyleSheetWithRegister(widget, source, theme, true)
}

// AddStyleSheetWithRegister adds a stylesheet, optionally registering it.
func AddStyleSheetWithRegister(widget *qt.QWidget, source interface{}, theme Theme, register bool) {
	s := toStyleSheet(source)
	var qss string
	if register {
		styleSheetManager.register(s, widget, false)
		qss = GetStyleSheet(styleSheetManager.source(widget), theme)
	} else {
		qss = widget.StyleSheet() + "\n" + GetStyleSheet(s, theme)
	}
	if strings.TrimRight(qss, " \t\r\n") != strings.TrimRight(widget.StyleSheet(), " \t\r\n") {
		widget.SetStyleSheet(qss)
	}
}

// SetCustomStyleSheet stores the light/dark custom QSS for a widget and re-applies
// the widget's registered stylesheet so the custom QSS is composed on top of the
// fluent stylesheet (the Go equivalent of Python's setCustomStyleSheet plus the
// CustomStyleSheetWatcher that triggers addStyleSheet). The previous implementation
// replaced the whole stylesheet with only the custom QSS, which erased the fluent
// look (header/selection/item padding) after e.g. TableView.SetBorderRadius.
func SetCustomStyleSheet(widget *qt.QWidget, lightQss, darkQss string) {
	cs := NewCustomStyleSheet(widget)
	cs.SetCustomStyleSheet(lightQss, darkQss)
	AddStyleSheet(widget, cs, ThemeAuto)
}

// updateStyleSheetDepth guards against re-entrant theme changes: applying a
// stylesheet can (via Qt re-polish/resize) re-enter SetTheme/UpdateStyleSheet,
// which would otherwise recurse until the stack overflows (the "100% CPU then
// crash" symptom). The depth counter breaks that cycle; nested calls become
// no-ops and the outermost call re-applies everything once.
var updateStyleSheetDepth int32

// UpdateStyleSheet re-applies the registered stylesheets (theme switch).
func UpdateStyleSheet(lazy bool) {
	if atomic.AddInt32(&updateStyleSheetDepth, 1) != 1 {
		atomic.AddInt32(&updateStyleSheetDepth, -1)
		return
	}
	defer atomic.AddInt32(&updateStyleSheetDepth, -1)

	styleSheetManager.mu.Lock()
	widgets := make([]*qt.QWidget, 0, len(styleSheetManager.sources))
	for w := range styleSheetManager.sources {
		widgets = append(widgets, w)
	}
	styleSheetManager.mu.Unlock()

	theme := QConfigInstance.Theme()
	for _, w := range widgets {
		styleSheetManager.mu.Lock()
		src := styleSheetManager.sources[w]
		styleSheetManager.mu.Unlock()
		if src == nil {
			continue
		}
		// Lazily mark hidden widgets as dirty; the dirty watcher re-applies
		// their stylesheet on their next paint. This is the Go equivalent of
		// Python's `lazy and widget.visibleRegion().isNull()` branch and avoids
		// re-polishing every widget of every hidden page on a theme switch.
		if lazy && !w.IsVisible() {
			trueV := qt.NewQVariant11(true)
			w.SetProperty("dirty-qss", trueV)
			trueV.Delete()
			continue
		}
		// Skip the re-polish when the rendered QSS is unchanged (Qt's
		// setStyleSheet always repolishes the widget subtree, which is the
		// dominant theme-switch cost for nested widgets). Matches the guard in
		// AddStyleSheet.
		qss := GetStyleSheet(src, theme)
		if strings.TrimRight(qss, " \t\r\n") != strings.TrimRight(w.StyleSheet(), " \t\r\n") {
			w.SetStyleSheet(qss)
		}
	}
}

// setThemeDepth breaks re-entrant theme changes (see updateStyleSheetDepth).
var setThemeDepth int32

// SetTheme sets the application theme and refreshes every registered widget.
func SetTheme(theme Theme, save bool, lazy bool) {
	if atomic.AddInt32(&setThemeDepth, 1) != 1 {
		atomic.AddInt32(&setThemeDepth, -1)
		return
	}
	defer atomic.AddInt32(&setThemeDepth, -1)

	QConfigInstance.Set(QConfigInstance.ThemeMode(), theme, save)
	UpdateStyleSheet(lazy)
	QConfigInstance.EmitThemeChangedFinished()
}

// ToggleTheme switches between light and dark.
func ToggleTheme(save bool, lazy bool) {
	if IsDarkTheme() {
		SetTheme(ThemeLight, save, lazy)
	} else {
		SetTheme(ThemeDark, save, lazy)
	}
}

// themeColor returns the primary theme color.
func themeColor() *qt.QColor { return ThemeColorPrimary.Color() }

// ThemeColorValue returns the primary theme color (exported alias).
func ThemeColorValue() *qt.QColor { return themeColor() }

// SetThemeColor sets the theme color and refreshes stylesheets.
func SetThemeColor(color *qt.QColor, save bool, lazy bool) {
	QConfigInstance.Set(QConfigInstance.ThemeColorItem(), color, save)
	invalidateRenderCache()
	UpdateStyleSheet(lazy)
}

// ---------------------------------------------------------------------------
// helpers

func toStyleSheet(source interface{}) StyleSheetBase {
	switch s := source.(type) {
	case StyleSheetBase:
		return s
	case string:
		return NewStyleSheetFile(s)
	default:
		return NewStyleSheetCompose()
	}
}

func readFileContent(path string) string {
	f := qt.NewQFile2(path)
	defer f.DeleteLater()
	if !f.Open(qt.QIODevice__ReadOnly) {
		return ""
	}
	b := f.ReadAll()
	f.Close()
	return string(b)
}
