package common

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	qt "github.com/mappu/miqt/qt"
)

// Alert is the startup tip printed by the original qfluentwidgets package.
const Alert = "\n\033[1;33m📢 Tips:\033[0m QFluentWidgets Pro is now released. Click \033[1;96mhttps://qfluentwidgets.com/pages/pro\033[0m to learn more about it.\n"

// Theme enumerates the application theme mode.
type Theme int

const (
	// ThemeLight forces the light theme.
	ThemeLight Theme = iota
	// ThemeDark forces the dark theme.
	ThemeDark
	// ThemeAuto follows the operating system theme.
	ThemeAuto
)

// String returns the canonical theme name used by the Python enum values.
func (t Theme) String() string {
	switch t {
	case ThemeDark:
		return "Dark"
	case ThemeAuto:
		return "Auto"
	default:
		return "Light"
	}
}

// Lower returns the lowercase theme name used in QSS resource paths.
func (t Theme) Lower() string {
	return strings.ToLower(t.String())
}

// themeFromString converts a Python-style theme name to a Theme.
func themeFromString(s string) Theme {
	switch strings.ToLower(s) {
	case "dark":
		return ThemeDark
	case "auto":
		return ThemeAuto
	default:
		return ThemeLight
	}
}

// ---------------------------------------------------------------------------
// Validators

// ConfigValidator validates and corrects config values. All values are passed
// as interface{} because different ConfigItems hold different Go types.
type ConfigValidator interface {
	Validate(value interface{}) bool
	Correct(value interface{}) interface{}
}

// DefaultValidator accepts every value unchanged.
type DefaultValidator struct{}

// Validate reports whether value is legal (always true).
func (DefaultValidator) Validate(interface{}) bool { return true }

// Correct returns value unchanged.
func (DefaultValidator) Correct(value interface{}) interface{} { return value }

// RangeValidator clamps a numeric value into [Min, Max].
type RangeValidator struct {
	Min int
	Max int
}

// Validate reports whether value is an int inside [Min, Max].
func (v *RangeValidator) Validate(value interface{}) bool {
	n, ok := asInt(value)
	return ok && v.Min <= n && n <= v.Max
}

// Correct clamps value into [Min, Max].
func (v *RangeValidator) Correct(value interface{}) interface{} {
	n, ok := asInt(value)
	if !ok {
		return value
	}
	if n < v.Min {
		n = v.Min
	}
	if n > v.Max {
		n = v.Max
	}
	return n
}

// OptionsValidator restricts the value to a fixed set of options.
type OptionsValidator struct {
	Options []interface{}
}

// NewOptionsValidator builds an OptionsValidator; empty options panic like the
// Python implementation raising ValueError.
func NewOptionsValidator(options ...interface{}) *OptionsValidator {
	if len(options) == 0 {
		panic("common: the `options` can't be empty")
	}
	return &OptionsValidator{Options: options}
}

// Validate reports whether value is one of the options.
func (v *OptionsValidator) Validate(value interface{}) bool {
	for _, o := range v.Options {
		if configEqual(o, value) {
			return true
		}
	}
	return false
}

// Correct returns value if valid, otherwise the first option.
func (v *OptionsValidator) Correct(value interface{}) interface{} {
	if v.Validate(value) {
		return value
	}
	return v.Options[0]
}

// BoolValidator restricts the value to true/false.
type BoolValidator struct {
	OptionsValidator
}

// NewBoolValidator builds a boolean validator.
func NewBoolValidator() *BoolValidator {
	return &BoolValidator{OptionsValidator: *NewOptionsValidator(true, false)}
}

// FolderValidator validates an existing directory path.
type FolderValidator struct{ DefaultValidator }

// Validate reports whether value is an existing directory.
func (FolderValidator) Validate(value interface{}) bool {
	s, ok := value.(string)
	return ok && isDir(s)
}

// Correct creates the directory (and parents) and returns its absolute path
// with forward slashes.
func (FolderValidator) Correct(value interface{}) interface{} {
	s, _ := value.(string)
	_ = os.MkdirAll(s, 0o755)
	abs, err := filepath.Abs(s)
	if err != nil {
		abs = s
	}
	return strings.ReplaceAll(abs, "\\", "/")
}

// FolderListValidator validates a list of existing directories.
type FolderListValidator struct{ DefaultValidator }

// Validate reports whether every entry is an existing directory.
func (FolderListValidator) Validate(value interface{}) bool {
	list, ok := value.([]string)
	if !ok {
		return false
	}
	for _, s := range list {
		if !isDir(s) {
			return false
		}
	}
	return true
}

// Correct filters the list down to the directories that exist.
func (FolderListValidator) Correct(value interface{}) interface{} {
	list, ok := value.([]string)
	if !ok {
		return []string{}
	}
	out := make([]string, 0, len(list))
	for _, s := range list {
		if isDir(s) {
			abs, err := filepath.Abs(s)
			if err != nil {
				abs = s
			}
			out = append(out, strings.ReplaceAll(abs, "\\", "/"))
		}
	}
	return out
}

// ColorValidator validates a QColor (or color name) value.
type ColorValidator struct {
	Default *qt.QColor
}

// NewColorValidator builds a color validator with the given default.
func NewColorValidator(def string) *ColorValidator {
	return &ColorValidator{Default: qt.NewQColor6(def)}
}

// Validate reports whether value can be parsed as a valid QColor.
func (v *ColorValidator) Validate(value interface{}) bool {
	if c, ok := value.(*qt.QColor); ok {
		return c.IsValid()
	}
	c := toQColor(value)
	if c == nil {
		return false
	}
	defer c.Delete()
	return c.IsValid()
}

// Correct returns the parsed color, or the default if invalid.
func (v *ColorValidator) Correct(value interface{}) interface{} {
	if c, ok := value.(*qt.QColor); ok && c.IsValid() {
		return c
	}
	if v.Validate(value) {
		return toQColor(value)
	}
	return v.Default
}

// ---------------------------------------------------------------------------
// Serializers

// ConfigSerializer serializes a config value to/from its on-disk JSON form.
type ConfigSerializer interface {
	Serialize(value interface{}) interface{}
	Deserialize(value interface{}) interface{}
}

// DefaultSerializer is the identity serializer.
type DefaultSerializer struct{}

// Serialize returns value unchanged.
func (DefaultSerializer) Serialize(value interface{}) interface{} { return value }

// Deserialize returns value unchanged.
func (DefaultSerializer) Deserialize(value interface{}) interface{} { return value }

// EnumSerializer serializes a Theme as its name string.
type EnumSerializer struct{ DefaultSerializer }

// Serialize returns the Theme name ("Light"/"Dark"/"Auto").
func (EnumSerializer) Serialize(value interface{}) interface{} {
	t, ok := value.(Theme)
	if !ok {
		return ThemeLight.String()
	}
	return t.String()
}

// Deserialize converts a name string back to a Theme.
func (EnumSerializer) Deserialize(value interface{}) interface{} {
	s, ok := value.(string)
	if !ok {
		return ThemeLight
	}
	return themeFromString(s)
}

// ColorSerializer serializes a *qt.QColor as its "#AARRGGBB" name.
type ColorSerializer struct{ DefaultSerializer }

// Serialize returns the HexArgb name of the color.
func (ColorSerializer) Serialize(value interface{}) interface{} {
	c := toQColor(value)
	if c == nil {
		return "#00000000"
	}
	return c.NameWithFormat(qt.QColor__HexArgb)
}

// Deserialize parses a color name (or []int) into a *qt.QColor.
func (ColorSerializer) Deserialize(value interface{}) interface{} {
	if c := toQColor(value); c != nil {
		return c
	}
	return qt.NewQColor6("#009faa")
}

// ---------------------------------------------------------------------------
// ConfigItem

// ConfigItem is a single typed configuration entry. It mirrors the Python
// ConfigItem but stores the value as interface{}; a valueChanged callback
// replaces the pyqtSignal.
type ConfigItem struct {
	Group        string
	Name         string
	Restart      bool
	DefaultValue interface{}

	validator  ConfigValidator
	serializer ConfigSerializer
	value      interface{}

	// OnValueChanged is emitted when the value changes.
	OnValueChanged func(v interface{})
}

// NewConfigItem builds a config item.
func NewConfigItem(group, name string, def interface{}, validator ConfigValidator, serializer ConfigSerializer, restart bool) *ConfigItem {
	if validator == nil {
		validator = DefaultValidator{}
	}
	if serializer == nil {
		serializer = DefaultSerializer{}
	}
	return &ConfigItem{
		Group:        group,
		Name:         name,
		Restart:      restart,
		DefaultValue: validator.Correct(def),
		validator:    validator,
		serializer:   serializer,
		value:        validator.Correct(def),
	}
}

// Value returns the current value.
func (c *ConfigItem) Value() interface{} { return c.value }

// SetValue corrects and stores the value, emitting OnValueChanged on change.
func (c *ConfigItem) SetValue(v interface{}) {
	corrected := c.validator.Correct(v)
	if configEqual(c.value, corrected) {
		return
	}
	c.value = corrected
	if c.OnValueChanged != nil {
		c.OnValueChanged(c.value)
	}
}

// Key returns the dotted config key ("group.name" or "group").
func (c *ConfigItem) Key() string {
	if c.Name == "" {
		return c.Group
	}
	return c.Group + "." + c.Name
}

// Range returns the [min, max] bounds when the item was built with a
// *RangeValidator (the Go equivalent of Python's RangeConfigItem.range). It
// returns 0, 0 for items without a range validator.
func (c *ConfigItem) Range() (int, int) {
	if v, ok := c.validator.(*RangeValidator); ok {
		return v.Min, v.Max
	}
	return 0, 0
}

// Options returns the option list when the item was built with an
// *OptionsValidator (the Go equivalent of Python's OptionsConfigItem.options).
// It returns nil for items without an options validator.
func (c *ConfigItem) Options() []interface{} {
	if v, ok := c.validator.(*OptionsValidator); ok {
		return v.Options
	}
	return nil
}

// Serialize returns the on-disk representation of the value.
func (c *ConfigItem) Serialize() interface{} { return c.serializer.Serialize(c.value) }

// DeserializeFrom parses an on-disk value into the item.
func (c *ConfigItem) DeserializeFrom(v interface{}) {
	c.SetValue(c.serializer.Deserialize(v))
}

// ---------------------------------------------------------------------------
// QConfig

// QConfig is the application configuration singleton. It replaces the Python
// QObject-based QConfig; pyqtSignal emissions are represented by callback
// registries.
type QConfig struct {
	mu sync.Mutex

	theme        Theme
	themeColor   *qt.QColor
	fontFamilies []string

	themeMode        *ConfigItem
	themeColorItem   *ConfigItem
	fontFamiliesItem *ConfigItem

	// extraItems are additional ConfigItems that examples register so they are
	// persisted/loaded alongside the three built-in items (the Go equivalent of
	// subclassing QConfig and passing the items to qconfig.load).
	extraItems []*ConfigItem

	file string

	themeListeners         []func(Theme)
	colorListeners         []func(*qt.QColor)
	restartListeners       []func()
	themeFinishedListeners []func()
}

// NewQConfig builds a QConfig with the same defaults as the Python package.
func NewQConfig() *QConfig {
	c := &QConfig{
		theme:        ThemeLight,
		themeColor:   qt.NewQColor6("#009faa"),
		fontFamilies: []string{"Segoe UI", "Microsoft YaHei", "PingFang SC"},
		file:         "config/config.json",
	}
	c.themeMode = NewConfigItem("QFluentWidgets", "ThemeMode", ThemeLight, NewOptionsValidator(ThemeLight, ThemeDark, ThemeAuto), EnumSerializer{}, false)
	c.themeColorItem = NewConfigItem("QFluentWidgets", "ThemeColor", c.themeColor, NewColorValidator("#009faa"), ColorSerializer{}, false)
	c.fontFamiliesItem = NewConfigItem("QFluentWidgets", "FontFamilies", c.fontFamilies, DefaultValidator{}, DefaultSerializer{}, false)
	return c
}

// ThemeMode returns the ThemeMode config item.
func (c *QConfig) ThemeMode() *ConfigItem { return c.themeMode }

// ThemeColorItem returns the ThemeColor config item.
func (c *QConfig) ThemeColorItem() *ConfigItem { return c.themeColorItem }

// FontFamiliesItem returns the FontFamilies config item.
func (c *QConfig) FontFamiliesItem() *ConfigItem { return c.fontFamiliesItem }

// RegisterItems adds extra ConfigItems to the shared persistence set. Examples
// call this once with their own items before Load so that Save/Load handle them
// together with the built-in theme items. Registering the same item twice is
// harmless (Load/Save operate on a set).
func (c *QConfig) RegisterItems(items ...*ConfigItem) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.extraItems = append(c.extraItems, items...)
}

// extraItemSnapshot returns a copy of the registered extra items.
func (c *QConfig) extraItemSnapshot() []*ConfigItem {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]*ConfigItem{}, c.extraItems...)
}

// Get returns the value of a config item.
func (c *QConfig) Get(item *ConfigItem) interface{} { return item.Value() }

// Set stores a value on a config item and, when save is true, persists the
// whole config to disk. Settings themeMode/themeColor also emits the matching
// signals.
func (c *QConfig) Set(item *ConfigItem, value interface{}, save bool) {
	if configEqual(item.Value(), value) {
		return
	}

	item.SetValue(value)

	if save {
		c.Save()
	}

	if item.Restart {
		c.emitRestart()
	}

	switch item {
	case c.themeMode:
		if t, ok := item.Value().(Theme); ok {
			c.SetTheme(t)
			c.emitTheme(t)
		}
	case c.themeColorItem:
		if col := toQColor(item.Value()); col != nil {
			c.mu.Lock()
			c.themeColor = col
			c.mu.Unlock()
		}
		// The QSS render cache keys on the raw QSS (which still contains the
		// --ThemeColor* placeholders), so a theme-color change must drop cached
		// renders before the OnThemeColorChanged listeners re-render. The gallery
		// theme-color card drives this path via QConfigInstance.Set directly.
		invalidateRenderCache()
		c.emitColor(item.Value())
	}
}

// ToDict returns a nested map of the config suitable for JSON marshalling.
func (c *QConfig) ToDict(serialize bool) map[string]interface{} {
	items := append([]*ConfigItem{c.themeMode, c.themeColorItem, c.fontFamiliesItem}, c.extraItemSnapshot()...)
	out := map[string]interface{}{}
	for _, item := range items {
		var v interface{} = item.Value()
		if serialize {
			v = item.Serialize()
		}
		group := out[item.Group]
		if item.Name == "" {
			out[item.Group] = v
			continue
		}
		m, ok := group.(map[string]interface{})
		if !ok {
			m = map[string]interface{}{}
			out[item.Group] = m
		}
		m[item.Name] = v
	}
	return out
}

// Save writes the config to the configured JSON file.
func (c *QConfig) Save() {
	dir := filepath.Dir(c.file)
	_ = os.MkdirAll(dir, 0o755)
	b, err := json.MarshalIndent(c.ToDict(true), "", "    ")
	if err != nil {
		return
	}
	_ = os.WriteFile(c.file, b, 0o644)
}

// Load reads the config JSON file (or the given path) and applies it.
func (c *QConfig) Load(file string) {
	if file != "" {
		c.file = file
	}

	cfg := map[string]interface{}{}
	if b, err := os.ReadFile(c.file); err == nil {
		_ = json.Unmarshal(b, &cfg)
	}

	index := map[string]*ConfigItem{
		c.themeMode.Key():        c.themeMode,
		c.themeColorItem.Key():   c.themeColorItem,
		c.fontFamiliesItem.Key(): c.fontFamiliesItem,
	}
	for _, item := range c.extraItemSnapshot() {
		index[item.Key()] = item
	}

	for k, v := range cfg {
		if m, ok := v.(map[string]interface{}); ok {
			for sub, subv := range m {
				key := k + "." + sub
				if item, found := index[key]; found {
					item.DeserializeFrom(subv)
				}
			}
			continue
		}
		if item, found := index[k]; found {
			item.DeserializeFrom(v)
		}
	}

	if t, ok := c.themeMode.Value().(Theme); ok {
		c.SetTheme(t)
	}

	if fams := toStringSlice(c.fontFamiliesItem.Value()); fams != nil {
		c.mu.Lock()
		c.fontFamilies = fams
		c.mu.Unlock()
	}
}

// SetTheme sets the theme without touching the config file. ThemeAuto resolves
// to the detected system theme.
func (c *QConfig) SetTheme(t Theme) {
	if t == ThemeAuto {
		t = detectSystemTheme()
	}
	c.mu.Lock()
	c.theme = t
	c.mu.Unlock()
}

// Theme returns the resolved theme (never ThemeAuto).
func (c *QConfig) Theme() Theme {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.theme
}

// ThemeColor returns the current theme color as a *qt.QColor.
func (c *QConfig) ThemeColor() *qt.QColor {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.themeColor
}

// FontFamilies returns a copy of the font family list.
func (c *QConfig) FontFamilies() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]string, len(c.fontFamilies))
	copy(out, c.fontFamilies)
	return out
}

// SetFontFamilies stores the font family list.
func (c *QConfig) SetFontFamilies(families []string, save bool) {
	c.Set(c.fontFamiliesItem, families, save)
	c.mu.Lock()
	c.fontFamilies = append([]string(nil), families...)
	c.mu.Unlock()
	invalidateRenderCache()
}

// ---------------------------------------------------------------------------
// Callback registries (pyqtSignal equivalents)

// OnThemeChanged registers a listener for theme changes.
func (c *QConfig) OnThemeChanged(l func(Theme)) {
	c.mu.Lock()
	c.themeListeners = append(c.themeListeners, l)
	c.mu.Unlock()
}

// OnThemeColorChanged registers a listener for theme color changes.
func (c *QConfig) OnThemeColorChanged(l func(*qt.QColor)) {
	c.mu.Lock()
	c.colorListeners = append(c.colorListeners, l)
	c.mu.Unlock()
}

// OnAppRestart registers a listener for the appRestart signal.
func (c *QConfig) OnAppRestart(l func()) {
	c.mu.Lock()
	c.restartListeners = append(c.restartListeners, l)
	c.mu.Unlock()
}

// OnThemeChangedFinished registers a listener for themeChangedFinished.
func (c *QConfig) OnThemeChangedFinished(l func()) {
	c.mu.Lock()
	c.themeFinishedListeners = append(c.themeFinishedListeners, l)
	c.mu.Unlock()
}

// EmitThemeChangedFinished emits the themeChangedFinished signal.
func (c *QConfig) EmitThemeChangedFinished() {
	for _, l := range c.snapshotFinished() {
		l()
	}
}

func (c *QConfig) emitTheme(v interface{}) {
	t, _ := v.(Theme)
	for _, l := range c.snapshotTheme() {
		l(t)
	}
}

func (c *QConfig) emitColor(v interface{}) {
	col := toQColor(v)
	for _, l := range c.snapshotColor() {
		l(col)
	}
}

func (c *QConfig) emitRestart() {
	for _, l := range c.snapshotRestart() {
		l()
	}
}

func (c *QConfig) snapshotTheme() []func(Theme) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]func(Theme){}, c.themeListeners...)
}

func (c *QConfig) snapshotColor() []func(*qt.QColor) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]func(*qt.QColor){}, c.colorListeners...)
}

func (c *QConfig) snapshotRestart() []func() {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]func(){}, c.restartListeners...)
}

func (c *QConfig) snapshotFinished() []func() {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]func(){}, c.themeFinishedListeners...)
}

// ---------------------------------------------------------------------------
// Package-level singleton and helpers

// QConfigInstance is the process-wide config instance (the Go equivalent of
// the Python module-level `qconfig`).
var QConfigInstance = NewQConfig()

// IsDarkTheme reports whether the current resolved theme is dark.
func IsDarkTheme() bool { return QConfigInstance.Theme() == ThemeDark }

// CurrentTheme returns the current resolved theme.
func CurrentTheme() Theme { return QConfigInstance.Theme() }

// IsDarkThemeMode reports whether the given theme (ThemeAuto falls back to the
// current theme) is dark.
func IsDarkThemeMode(t Theme) bool {
	if t == ThemeAuto {
		return IsDarkTheme()
	}
	return t == ThemeDark
}

// detectSystemTheme returns the OS theme; without a native dark-mode detector
// it degrades to ThemeLight (documented fallback in MIGRATION_GUIDE §8).
func detectSystemTheme() Theme {
	return ThemeLight
}

// ---------------------------------------------------------------------------
// helpers

func isDir(s string) bool {
	info, err := os.Stat(s)
	return err == nil && info.IsDir()
}

func toStringSlice(v interface{}) []string {
	switch x := v.(type) {
	case []string:
		return append([]string(nil), x...)
	case []interface{}:
		out := make([]string, 0, len(x))
		for _, e := range x {
			if s, ok := e.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

func asInt(v interface{}) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	case float32:
		return int(n), true
	default:
		return 0, false
	}
}

func configEqual(a, b interface{}) bool {
	if ca, ok := a.(*qt.QColor); ok {
		if cb, ok := b.(*qt.QColor); ok {
			return ca.Name() == cb.Name()
		}
	}
	return reflect.DeepEqual(a, b)
}

// toQColor coerces a value into a *qt.QColor. It accepts *qt.QColor, a hex
// string, or []int{R,G,B(,A)}.
func toQColor(v interface{}) *qt.QColor {
	switch x := v.(type) {
	case *qt.QColor:
		return x
	case string:
		return qt.NewQColor6(x)
	case []int:
		switch len(x) {
		case 3:
			return qt.NewQColor3(x[0], x[1], x[2])
		case 4:
			return qt.NewQColor11(x[0], x[1], x[2], x[3])
		}
	}
	return nil
}
